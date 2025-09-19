package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetsPath(fileExtension, assetsPath string) (string, string, error) {
	sRdmNum, err := getRdmNum()
	if err != nil {
		return "", "", err
	}
	return filepath.Join(assetsPath, fmt.Sprintf("%v%v", sRdmNum, fileExtension)), sRdmNum, nil
}

func getFileExtension(contentType string) string {
	fileExtension := strings.Split(contentType, "/")
	if len(fileExtension) != 2 {
		return ".bin"
	}
	return "." + fileExtension[1]
}

func getRdmNum() (string, error) {
	rdmNum := make([]byte, 32)
	_, err := rand.Read(rdmNum)
	if err != nil {
		return "", err
	}
	b64Enc := base64.RawURLEncoding
	return b64Enc.EncodeToString(rdmNum), nil
}
