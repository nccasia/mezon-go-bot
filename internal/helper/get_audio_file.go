package helper

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func GetAudioFiles(directory, prefix string) ([]string, error) {
	var files []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if the file matches the prefix and extension .ogg
		if !info.IsDir() && strings.HasPrefix(info.Name(), prefix) && strings.HasSuffix(info.Name(), ".ogg") {
			files = append(files, info.Name())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort the file list to play in order
	sort.Strings(files)
	return files, nil
}

func FindFileByName(directory, fileName string) (string, error) {
	var result string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if the current file matches the target name
		if !info.IsDir() && info.Name() == fileName {
			result = info.Name()    // Store the full path of the found file
			return filepath.SkipDir // Stop searching further
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	// If result is empty, the file was not found
	if result == "" {
		return "", os.ErrNotExist
	}
	return result, nil
}
