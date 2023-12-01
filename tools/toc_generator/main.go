package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	files, err := fineStepsFiles(".")
	if err != nil {
		log.Fatalf("error finding steps files: %v", err)
	}

	for _, f := range files {
		dir := filepath.Dir(f)
		parent := filepath.Base(dir)

		fmt.Printf("* [%s](%s)\n", parent, f)
	}
}

func fineStepsFiles(directory string) ([]string, error) {
	var stepsFiles []string

	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if info.Name() == "steps.md" {
			stepsFiles = append(stepsFiles, path)
		}

		return nil
	}

	if err := filepath.Walk(directory, visit); err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	return stepsFiles, nil
}
