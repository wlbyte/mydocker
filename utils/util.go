package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

func HashStr(v any) (string, error) {
	hash := sha256.New()
	if _, err := hash.Write(fmt.Appendln(nil, v)); err != nil {
		return "", fmt.Errorf("hashStr: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func PathNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}
