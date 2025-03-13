package container

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func NewParentProcess(tty bool) (*exec.Cmd, *os.File, error) {
	// 创建目录和镜像环境
	mntPath := "/root/merged"
	rootPath := "/root"
	NewWorkspace(rootPath, mntPath)
	// 创建匿名管道用于传递参数，将readPipe作为子进程的ExtraFiles，子进程从readPipe中读取参数
	// 父进程中则通过writePipe将参数写入管道
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Printf("[error] new pipe %v", err)
		return nil, nil, err
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
	cmd.Dir = mntPath
	return cmd, writePipe, nil
}

func NewWorkspace(rootPath, mntPath string) {
	createLower(rootPath)
	createUpperWorker(rootPath)
	mountOverlayFS(rootPath, mntPath)
}

func createLower(rootURL string) {
	busyboxURL := rootURL + "/busybox"
	busyboxTarURL := rootURL + "/busybox.tar"
	MkDirErrorExit(busyboxURL, 0777)
	if pathNotExist(busyboxTarURL) {
		log.Println("[debug] busybox image is not exist")
		os.Exit(1)
	}
	if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
		log.Println("[error] untar busybox:", err)
		os.Exit(1)
	}
}

func createUpperWorker(rootPath string) {
	upperPath := rootPath + "/upper"
	MkDirErrorExit(upperPath, 0755)
	workPath := rootPath + "/work"
	MkDirErrorExit(workPath, 0755)
}

func mountOverlayFS(rootPath, mntPath string) {
	// mount -t overlay overlay -o lowerdir=/lower,upperdir=/upper,workdir=/work /merged
	MkDirErrorExit(mntPath, 0755)
	dirs := "lowerdir=" + rootPath + "/busybox" + ",upperdir=" + rootPath + "/upper" + ",workdir=" + rootPath + "/work"
	if output, err := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntPath).CombinedOutput(); err != nil {
		log.Println("[error] mount overlayfs:" + err.Error() + "," + string(output))
	}
}

func DelWorkspace(rootPath, mntPath string) {
	umountOverlayFS(mntPath)
	delDirs(rootPath)
}

func umountOverlayFS(mntPath string) {
	if output, err := exec.Command("umount", mntPath).CombinedOutput(); err != nil {
		fmt.Println("[error] umount overlayFS " + mntPath + ":" + err.Error() + ", " + string(output))
	}
	log.Println("[debug] umount", mntPath)
	RmDir(mntPath)
}

func delDirs(rootPath string) {
	RmDir(rootPath + "/upper")
	RmDir(rootPath + "/work")
}

func pathNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func mkDir(path string, perm os.FileMode) error {
	if pathNotExist(path) {
		if err := os.Mkdir(path, perm); err != nil {
			return fmt.Errorf("[error] mkDir: %w", err)
		}
	}
	return nil
}

func MkDirErrorExit(path string, perm os.FileMode) {
	if err := mkDir(path, perm); err != nil {
		fmt.Println("[error] create dir " + path + ":" + err.Error())
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
		fmt.Println("[error] " + err.Error())
		return
	}
	log.Println("[debug] rm dir", path)
}
