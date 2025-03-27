package consts

import (
	"fmt"
)

const (
	PATH_HOME          = "/var/lib/mydocker"
	PATH_CONTAINER     = PATH_HOME + "/containers"
	PATH_IMAGE         = PATH_HOME + "/image"
	PATH_FS_ROOT       = PATH_HOME + "/overlay2"
	PATH_LOWER_FORMAT  = PATH_FS_ROOT + "/%s/lower"
	PATH_UPPER_FORMAT  = PATH_FS_ROOT + "/%s/upper"
	PATH_MERGED_FORMAT = PATH_FS_ROOT + "/%s/merged"
	PATH_WORK_FORMAT   = PATH_FS_ROOT + "/%s/work"
	MOUNT_PATH_FORMAT  = "lowerdir=%s,upperdir=%s,workdir=%s"
)

const (
	MODE_0755      = 0755
	STATUS_RUNNING = "running"
	STATUS_STOPPED = "stopped"
)

func GetPathLower(containerID string) string {
	return fmt.Sprintf(PATH_LOWER_FORMAT, containerID)
}

func GetPathUpper(containerID string) string {
	return fmt.Sprintf(PATH_UPPER_FORMAT, containerID)
}

func GetPathWork(containerID string) string {
	return fmt.Sprintf(PATH_WORK_FORMAT, containerID)
}

func GetPathMerged(containerID string) string {
	return fmt.Sprintf(PATH_MERGED_FORMAT, containerID)
}

func GetMountSrcDir(containerID string) string {
	return fmt.Sprintf(MOUNT_PATH_FORMAT, GetPathLower(containerID), GetPathUpper(containerID), GetPathWork(containerID))
}
