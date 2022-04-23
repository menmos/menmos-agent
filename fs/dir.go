package fs

import (
	"fmt"
	"os"
)

// EnsureDirExists checks if a given directory exists, creating it if necessary.
// Returns an error if the given path exists but is not a directory.
func EnsureDirExists(path string) error {
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	if err != nil {
		return err
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("agent path '%v' is not a directory", path)
	}

	return nil
}
