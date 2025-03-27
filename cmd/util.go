package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wlbyte/mydocker/consts"
	"github.com/wlbyte/mydocker/container"
)

func recordContainerInfo(ci *container.Container) error {
	errFormat := "recordContainerInfo: %w"
	curPath := consts.PATH_CONTAINER + "/" + ci.Id
	container.MkDir(curPath)
	bs, err := json.Marshal(ci)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := os.WriteFile(curPath+"/config.json", bs, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func GetContainerInfoAll(searchDir string) []*container.Container {
	fs := findJsonFilePathAll(searchDir)
	return getContainerInfoAll(fs)
}

func getContainerInfoAll(fs []string) []*container.Container {
	var cis []*container.Container

	for _, f := range fs {
		c := getContainerInfo(f)
		if c == nil {
			continue
		}
		cis = append(cis, c)
	}
	return cis
}

func findJsonFilePathAll(dir string) []string {
	var filePaths []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("[error] filepath.Walk:", path, err)
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			filePaths = append(filePaths, path)
		}
		return nil
	})
	return filePaths
}

func GetContainerInfo(containerID string) *container.Container {
	f := findJsonFilePath(containerID, consts.PATH_CONTAINER)
	return getContainerInfo(f)
}

func getContainerInfo(f string) *container.Container {
	var c *container.Container
	bs, err := os.ReadFile(f)
	if err != nil && err != io.EOF {
		log.Println("[info] getContainerInfo:", err)
		return nil
	}
	if err := json.Unmarshal(bs, &c); err != nil {
		log.Println("[info] getContainerInfo:", err)
		return nil
	}
	return c
}

func findJsonFilePath(subFilePath, searchDir string) string {
	var ret string
	filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("[error] findJsonFilePath:", path, err)
			return err
		}
		if info.Mode().IsRegular() && strings.Contains(path, subFilePath) && filepath.Ext(path) == ".json" {
			ret = path
		}
		return nil
	})
	return ret
}
