package native

import (
	"fmt"
	"os"
)

func ensureDirExists(path string) error {
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
