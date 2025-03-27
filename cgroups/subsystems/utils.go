package subsystems

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/wlbyte/mydocker/consts"
)

func FindCgroupMountpoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// txt 大概是这样的：104 85 0:20 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime - cgroup cgroup rw,memory
		txt := scanner.Text()
		// fields := strings.Fields(txt)
		fields := strings.Fields(txt)
		index := len(fields) - 1
		for _, opt := range strings.Split(fields[index], ",") {
			if opt == subsystem {
				if len(fields) > 4 {
					return fields[4]
				}
				return ""
			}
		}
	}
	if err = scanner.Err(); err != nil {
		log.Println("[error] scanner:", err)
	}
	return ""
}

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountpoint(subsystem)
	absPath := path.Join(cgroupRoot, cgroupPath)
	_, err := os.Stat(absPath)
	if err == nil {
		return absPath, nil
	}
	if autoCreate && os.IsNotExist(err) {
		if err := os.MkdirAll(absPath, consts.MODE_0755); err != nil {
			return "", fmt.Errorf("[error] create cgroup: %w", err)
		}
		return absPath, nil
	}
	return "", fmt.Errorf("create cgroup: %w", err)
}
