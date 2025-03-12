package cgroups

import "github.com/wlbyte/mydocker/cgroups/subsystems"

type CgroupManager struct {
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, sub := range subsystems.SubsystemsIns {
		if err := sub.Set(c.Path, res); err != nil {
			return err
		}
	}
	return nil
}

func (c *CgroupManager) Apply(pid int, res *subsystems.ResourceConfig) error {
	for _, sub := range subsystems.SubsystemsIns {
		if err := sub.Apply(c.Path, pid, res); err != nil {
			return err
		}
	}
	return nil
}

func (c *CgroupManager) Destroy() error {
	for _, sub := range subsystems.SubsystemsIns {
		return sub.Remove(c.Path)
	}
	return nil
}
