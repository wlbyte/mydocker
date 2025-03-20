package container

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/wlbyte/mydocker/constants"
	"golang.org/x/sys/unix"
)

func NewParentProcess(tty bool, volumePath string) (*exec.Cmd, *os.File, error) {
	errFormat := "newPararentProcess: %w"
	// 创建目录和镜像环境
	NewWorkspace(constants.ROOT_PATH, constants.MNT_PATH, volumePath)
	// 创建匿名管道用于传递参数，将readPipe作为子进程的ExtraFiles，子进程从readPipe中读取参数
	// 父进程中则通过writePipe将参数写入管道
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Printf("[error] new pipe %v", err)
		return nil, nil, fmt.Errorf(errFormat, err)
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Dir = constants.MNT_PATH
	return cmd, writePipe, nil
}

func NewWorkspace(rootPath, mntPath, volumePath string) {
	createDir(rootPath, mntPath, volumePath)
	initLower(rootPath)
	mountPath(rootPath, mntPath, volumePath)
	// 创建容器根配置目录
	curPath := constants.CONTAINER_ROOT_PATH
	MkDirErrorExit(curPath, 0755)
}

func initLower(rootURL string) {
	busyboxURL := rootURL + "/busybox"
	busyboxTarURL := rootURL + "/busybox.tar"
	if pathNotExist(busyboxTarURL) {
		log.Println("[debug] busybox image is not exist")
		os.Exit(1)
	}
	if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
		log.Println("[error] untar busybox:", err)
		os.Exit(1)
	}
}

func createDir(rootPath, mntPath, volumePath string) {
	// 创建 upper 目录
	upperPath := rootPath + "/upper"
	MkDirErrorExit(upperPath, 0755)
	// 创建work 目录
	workPath := rootPath + "/work"
	MkDirErrorExit(workPath, 0755)
	// 创建容器根目录
	MkDirErrorExit(mntPath, 0755)
}

func mountPath(rootPath, mntPath, volumePath string) {
	// mount -t overlay overlay -o lowerdir=/lower,upperdir=/upper,workdir=/work /merged
	dirs := "lowerdir=" + rootPath + "/busybox" + ",upperdir=" + rootPath + "/upper" + ",workdir=" + rootPath + "/work"
	if output, err := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntPath).CombinedOutput(); err != nil {
		log.Println("[error] mount overlayfs:" + err.Error() + "," + string(output))
		os.Exit(1)
	}
	log.Println("[debug] mountPath:", dirs, "->", mntPath)
	// 挂载-v 指定的 volume
	if volumePath != "" {
		volumes, err := parseVolumePath(volumePath)
		if err != nil {
			log.Println("[error] mountPath:", err)
			os.Exit(1)
		}
		hostPath := volumes[0]
		containerPath := mntPath + volumes[1]
		MkDirErrorExit(hostPath, 0755)
		MkDirErrorExit(containerPath, 0755)
		if err := unix.Mount(hostPath, containerPath, "", unix.MS_BIND, ""); err != nil {
			log.Println("[error] mountPath:", err)
			os.Exit(1)
		}
		log.Println("[debug] mountPath:", hostPath, "->", containerPath)
	}
}

func DelWorkspace(rootPath, mntPath, volumePath, containerDir string) {
	umountPath(mntPath, volumePath)
	dirs := []string{
		rootPath + "/upper",
		rootPath + "/work",
		mntPath,
		constants.CONTAINER_ROOT_PATH + "/" + containerDir,
	}
	delDirs(dirs)
}

func umountPath(mntPath, volumePath string) {
	if volumePath != "" {
		volumes, err := parseVolumePath(volumePath)
		if err != nil {
			log.Println("[error] umountPath:", err)
		}
		containerPath := mntPath + volumes[1]
		if err := unix.Unmount(containerPath, 0); err != nil {
			log.Println("[error] umountPath:", containerPath, err)
		} else {
			log.Println("[debug] umount", containerPath)
		}

	}
	if output, err := exec.Command("umount", mntPath).CombinedOutput(); err != nil {
		log.Println("[error] umountPath " + mntPath + ":" + string(output))
	} else {
		log.Println("[debug] umount", mntPath)
	}

}

func delDirs(dirs []string) {
	for _, d := range dirs {
		RmDir(d)
	}
}

func pathNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func mkDir(path string, perm os.FileMode) error {
	if pathNotExist(path) {
		if err := os.MkdirAll(path, perm); err != nil {
			return fmt.Errorf("[error] mkDir: %w", err)
		}
	}
	return nil
}

func MkDirErrorExit(path string, perm os.FileMode) {
	if err := mkDir(path, perm); err != nil {
		log.Println("[error] create dir " + path + ":" + err.Error())
		os.Exit(1)
	}
}

func rmDir(path string) error {
	if pathNotExist(path) {
		return nil
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove dir %s: %w", path, err)
	}
	return nil
}

func RmDir(path string) {
	if err := rmDir(path); err != nil {
		log.Println("[error] " + err.Error())
	} else {
		log.Println("[debug] rm dir", path)
	}
}

func parseVolumePath(volumePath string) ([]string, error) {
	errFormat := "parseVolumePath: %w"
	sSlice := strings.Split(volumePath, ":")
	if len(sSlice) != 2 || sSlice[0] == "" || sSlice[1] == "" {
		return nil, fmt.Errorf(errFormat, errors.New("volume path must be split by ':'"))
	}
	return sSlice, nil
}

// func mountVolume(rootPath, volumePath string) error {
// 	errFormat := "mountVolume: %w"
// 	volumes, err := parseVolumePath(volumePath)
// 	if err != nil {
// 		return fmt.Errorf(errFormat, err)
// 	}
// 	containerPath := rootPath + volumes[0]
// 	hostPath := volumes[1]
// 	MkDirErrorExit(containerPath, 0755)
// 	if err := unix.Mount(containerPath, hostPath, "", unix.MS_BIND, ""); err != nil {
// 		return fmt.Errorf(errFormat, err)
// 	}
// 	return nil
// }
