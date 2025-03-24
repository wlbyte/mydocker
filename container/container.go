package container

import "github.com/wlbyte/mydocker/cgroups/subsystems"

type Container struct {
	Id             string                     `json:"id"`
	Name           string                     `json:"name"`
	Pid            int                        `json:"pid"`
	Cmds           []string                   `json:"cmds"`
	Status         string                     `json:"status"`
	TTY            bool                       `json:"tty"`
	Detach         bool                       `json:"detach"`
	Volume         string                     `json:"volume"`
	ResourceConfig *subsystems.ResourceConfig `json:"resourceConfig"`
	CreateAt       string                     `json:"createAt"`
}
