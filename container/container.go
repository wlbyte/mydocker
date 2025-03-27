package container

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wlbyte/mydocker/cgroups/subsystems"
	"github.com/wlbyte/mydocker/consts"
	"github.com/wlbyte/mydocker/utils"
	"golang.org/x/sys/unix"
)

var ErrContainerNotExist = errors.New("container not exist")

type Container struct {
	Id             string                     `json:"id"`
	Name           string                     `json:"name"`
	ImageName      string                     `json:"imageName"`
	Pid            int                        `json:"pid"`
	Cmds           []string                   `json:"cmds"`
	Status         string                     `json:"status"`
	TTY            bool                       `json:"tty"`
	Detach         bool                       `json:"detach"`
	Volume         string                     `json:"volume"`
	Environment    []string                   `json:"environment"`
	ResourceConfig *subsystems.ResourceConfig `json:"resourceConfig"`
	CreateAt       string                     `json:"createAt"`
}

func NewParentProcess(c *Container) (*exec.Cmd, *os.File, error) {
	errFormat := "newPararentProcess: %w"
	// 创建目录和镜像环境
	if err := NewWorkspace(c); err != nil {
		return nil, nil, fmt.Errorf(errFormat, err)
	}
	// 创建匿名管道用于传递参数，将readPipe作为子进程的ExtraFiles，子进程从readPipe中读取参数
	// 父进程中则通过writePipe将参数写入管道
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		// log.Printf("[error] new pipe %v", err)
		return nil, nil, fmt.Errorf(errFormat, err)
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &unix.SysProcAttr{
		Cloneflags: unix.CLONE_NEWPID | unix.CLONE_NEWIPC |
			unix.CLONE_NEWNS | unix.CLONE_NEWNET |
			unix.CLONE_NEWUTS, // | unix.CLONE_NEWUSER,
		// UidMappings: []syscall.SysProcIDMap{
		// 	{ContainerID: 0, HostID: os.Getegid(), Size: 1},
		// },
		// GidMappings: []syscall.SysProcIDMap{
		// 	{ContainerID: 0, HostID: os.Getegid(), Size: 1},
		// },
	}
	cmd.Env = append(os.Environ(), c.Environment...)
	if c.TTY {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		logPath := fmt.Sprintf("%s/%s", consts.PATH_CONTAINER, c.Id)
		if err := MkDir(logPath); err != nil {
			return nil, nil, fmt.Errorf(errFormat, err)
		}
		logFile := fmt.Sprintf("%s/%s/%s.log", consts.PATH_CONTAINER, c.Id, c.Id)
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR, consts.MODE_0755)
		if err != nil {
			return nil, nil, fmt.Errorf(errFormat, err)
		}
		cmd.Stdout = f
		cmd.Stderr = f
	}

	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Dir = consts.GetPathMerged(c.Id)
	return cmd, writePipe, nil
}

func NewWorkspace(c *Container) error {
	errFormat := "initContainerDir %s: %w"
	// 创建数据目录
	containerPath := consts.PATH_CONTAINER
	rootFSPath := consts.PATH_FS_ROOT
	imagePath := consts.PATH_IMAGE
	if err := os.MkdirAll(containerPath, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, consts.PATH_CONTAINER, err)
	}
	if err := os.MkdirAll(rootFSPath, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, consts.PATH_FS_ROOT, err)
	}
	if err := os.MkdirAll(imagePath, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, consts.PATH_IMAGE, err)
	}
	// 创建当前容器目录
	lower := consts.GetPathLower(c.Id)
	upper := consts.GetPathUpper(c.Id)
	merged := consts.GetPathMerged(c.Id)
	work := consts.GetPathWork(c.Id)
	if err := os.MkdirAll(lower, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, lower, err)
	}
	if err := os.MkdirAll(upper, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, upper, err)
	}
	if err := os.MkdirAll(merged, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, merged, err)
	}
	if err := os.MkdirAll(work, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, work, err)
	}
	//初始化rootfs
	if err := initRootFS(c.Id, c.ImageName); err != nil {
		return fmt.Errorf(errFormat, "", err)
	}
	if err := mountPath(c.Id, c.Volume); err != nil {
		return fmt.Errorf(errFormat, "", err)
	}

	return nil
}

func initRootFS(containerID, imageName string) error {
	errFormat := "iniRootFS: %w"
	imagePath := filepath.Join(consts.PATH_IMAGE, imageName+".tar")
	if utils.PathNotExist(imagePath) {
		return fmt.Errorf(errFormat, errors.New("image not exist"))
	}
	if _, err := exec.Command("tar", "-xvf", imagePath, "-C", consts.GetPathLower(containerID)).CombinedOutput(); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func mountPath(containerID, volumePath string) error {
	errFormat := "mountPath %s: %w"
	// mount -t overlay overlay -o lowerdir=/lower,upperdir=/upper,workdir=/work /merged
	dstDir := consts.GetPathMerged(containerID)
	srcDir := consts.GetMountSrcDir(containerID)
	if output, err := exec.Command("mount", "-t", "overlay", "overlay", "-o", srcDir, dstDir).CombinedOutput(); err != nil {
		return fmt.Errorf(errFormat, output, err)
	}
	// 挂载-v 指定的 volume
	if volumePath != "" {
		volumes, err := parseVolumePath(volumePath)
		if err != nil {
			return fmt.Errorf(errFormat, "", err)
		}
		hostPath := volumes[0]
		containerPath := dstDir + volumes[1]
		if err := MkDir(hostPath); err != nil {
			return fmt.Errorf(errFormat, "", err)
		}
		if err := MkDir(containerPath); err != nil {
			return fmt.Errorf(errFormat, "", err)
		}
		if err := unix.Mount(hostPath, containerPath, "", unix.MS_BIND, ""); err != nil {
			return fmt.Errorf(errFormat, "unix.Mount", err)
		}
	}
	return nil
}

func DelWorkspace(c *Container) {
	if err := umountPath(c.Id, c.Volume); err != nil {
		log.Println("[error] DelWorkspace:", err)
	}
	if err := RmDir(filepath.Join(consts.PATH_CONTAINER, c.Id)); err != nil {
		log.Println("[error] DelWorkspace:", err)
	}
	if err := RmDir(filepath.Join(consts.PATH_FS_ROOT, c.Id)); err != nil {
		log.Println("[error] DelWorkspace:", err)
	}
}

func umountPath(containerID, volumePath string) error {
	errFormat := "unmountPath: %w"
	dstDir := consts.GetPathMerged(containerID)

	if volumePath != "" {
		volumes, err := parseVolumePath(volumePath)
		if err != nil {
			return fmt.Errorf(errFormat, err)
		}
		containerPath := dstDir + volumes[1]
		if err := unix.Unmount(containerPath, 0); err != nil {
			return fmt.Errorf(errFormat, err)
		}
	}
	if err := unix.Unmount(dstDir, 0); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func MkDir(path string) error {
	if err := os.MkdirAll(path, consts.MODE_0755); err != nil {
		return fmt.Errorf("MkDir: %w", err)
	}
	return nil
}

func RmDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("RemoveDir: %w", err)
	}
	return nil
}

func parseVolumePath(volumePath string) ([]string, error) {
	errFormat := "parseVolumePath: %w"
	sSlice := strings.Split(volumePath, ":")
	if len(sSlice) != 2 || sSlice[0] == "" || sSlice[1] == "" {
		return nil, fmt.Errorf(errFormat, errors.New("volume path must be split by ':'"))
	}
	return sSlice, nil
}
