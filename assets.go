package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetsPath(videoID uuid.UUID, fileExtension, assetsPath string) string {
	return filepath.Join(assetsPath, fmt.Sprintf("%v%v", videoID, fileExtension))
}

func getFileExtension(contentType string) string {
	fileExtension := strings.Split(contentType, "/")
	if len(fileExtension) != 2 {
		return ".bin"
	}
	return "." + fileExtension[1]
}
