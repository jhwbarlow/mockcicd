package prepare

import (
	"fmt"
	"log"
	"os"
)

type Preparer interface {
	Prepare(path string) error
}

type FilesystemPreparer struct{}

func NewFilesystemPreparer() *FilesystemPreparer {
	return new(FilesystemPreparer)
}

func (p *FilesystemPreparer) Prepare(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		log.Printf("creating source directory %q", path)
		if err := createDir(path); err != nil {
			return fmt.Errorf("creating source directory %q: %w", path, err)
		}

		return nil // Successfully created src dir
	} else if err != nil {
		return fmt.Errorf("checking if source directory %q exists: %w", path, err)
	}

	// Src dir already exists, so empty it
	log.Printf("emptying source directory %q", path)
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("emptying source directory %q: %w", path, err)
	}

	return nil
}

func createDir(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("creating directory %q: %w", path, err)
	}

	return nil
}
