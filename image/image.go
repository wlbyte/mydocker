package image

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/wlbyte/mydocker/constants"
)

func BuildImage(imageName string) error {
	srcDir := constants.MNT_PATH
	// ctx := context.TODO()

	// files, err := archives.FilesFromDisk(ctx, nil, map[string]string{srcDir: ""})
	// if err != nil {
	// 	return fmt.Errorf("BuildImage: %w", err)
	// }
	imageTar := imageName + ".tar"
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
	log.Println("[debug] build image:", imageTar)
	return nil
}
