package cache

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

func computeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func IsCacheUpToDate(srcDir, cacheDir string) (bool, error) {
	srcFiles := map[string]string{}
	cacheFiles := map[string]string{}

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(srcDir, path)
			hash, err := computeFileHash(path)
			if err != nil {
				return err
			}
			srcFiles[relPath] = hash
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	err = filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(cacheDir, path)
			hash, err := computeFileHash(path)
			if err != nil {
				return err
			}
			cacheFiles[relPath] = hash
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	// Compare src and cache files
	if len(srcFiles) != len(cacheFiles) {
		return false, nil
	}

	for file, hash := range srcFiles {
		if cacheHash, exists := cacheFiles[file]; !exists || cacheHash != hash {
			return false, nil
		}
	}
	return true, nil
}

func CopySrcToCache(srcDir, cacheDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(srcDir, path)
			destPath := filepath.Join(cacheDir, relPath)
			destDir := filepath.Dir(destPath)
			if _, err := os.Stat(destDir); os.IsNotExist(err) {
				os.MkdirAll(destDir, os.ModePerm)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if err := os.WriteFile(destPath, data, info.Mode()); err != nil {
				return err
			}
		}
		return nil
	})
}
