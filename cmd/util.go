package main

import (
	"os"
	"path/filepath"
	"strings"
)

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func exPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPathh := filepath.Dir(ex)
	return exPathh, nil
}

func compareHash(h1 string, h2 string) bool {
	var bigger string
	var smaller string
	if h1 > h2 {
		bigger = h1
		smaller = h2
	} else {
		bigger = h2
		smaller = h1
	}
	return strings.Contains(bigger, smaller)
}
