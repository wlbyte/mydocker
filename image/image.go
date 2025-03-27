package image

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/wlbyte/mydocker/consts"
)

type Image struct{}

func BuildImage(containerID, imageName string) error {
	srcDir := consts.GetPathMerged(containerID)
	// ctx := context.TODO()

	// files, err := archives.FilesFromDisk(ctx, nil, map[string]string{srcDir: ""})
	// if err != nil {
	// 	return fmt.Errorf("BuildImage: %w", err)
	// }
	imageTar := filepath.Join(consts.PATH_IMAGE, imageName+".tar")
	// outFile, err := os.Create(imageTar)
	// if err != nil {
	// 	return fmt.Errorf("BuildImage: %w", err)
	// }
	// defer outFile.Close()
	// outTarGz := archives.CompressedArchive{
	// 	Compression: archives.Gz{},
	// 	Archival:    archives.Tar{},
	// }
	// err = outTarGz.Archive(ctx, outFile, files)
	// if err != nil {
	// 	return fmt.Errorf("BuildImage: %w", err)
	// }
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", srcDir, ".").CombinedOutput(); err != nil {
		return fmt.Errorf("buildImage: %w", err)
	}
	return nil
}
