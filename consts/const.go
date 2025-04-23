package consts

import (
	"fmt"
)

const PATH_HOME = "/var/lib/mydocker"

const (
	MODE_0755 = 0755
)

// container
const (
	STATUS_RUNNING     = "running"
	STATUS_STOPPED     = "stopped"
	STATUS_EXITED      = "exited"
	PATH_CONTAINER     = PATH_HOME + "/containers"
	PATH_FS_ROOT       = PATH_HOME + "/overlay2"
	PATH_LOWER_FORMAT  = PATH_FS_ROOT + "/%s/lower"
	PATH_UPPER_FORMAT  = PATH_FS_ROOT + "/%s/upper"
	PATH_MERGED_FORMAT = PATH_FS_ROOT + "/%s/merged"
	PATH_WORK_FORMAT   = PATH_FS_ROOT + "/%s/work"
	MOUNT_PATH_FORMAT  = "lowerdir=%s,upperdir=%s,workdir=%s"
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

// image
const (
	PATH_IMAGE = PATH_HOME + "/image"
)

// network
const (
	DEFAULT_NETWORK      = "default"
	DEFAULT_DRIVER       = "bridge"
	PATH_NETWORK         = PATH_HOME + "/network"
	PATH_IPAM            = PATH_NETWORK + "/ipam"
	PATH_NETWORK_NETWORK = PATH_NETWORK + "/network"
	PATH_IPAM_JSON       = PATH_IPAM + "/subnet.json"
)
