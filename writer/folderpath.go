package writer

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// getEffingoFolderPath returns the home directory for the
// effingo log and cache files
func getEffingoFolderPath() (string, error) {
	curUser, err := user.Current()
	if err != nil {
		return "", err
	}

	system := runtime.GOOS

	var effingoPath string
	switch system {
	case "windows":
		effingoPath = filepath.Join(curUser.HomeDir, ".effingo")
	default:
		effingoPath = filepath.Join(curUser.HomeDir, ".cache/effingo")
	}

	return effingoPath, nil
}

// CreateEffingoFolter creates the effingo folder that will hold the
// cache and log gile
func CreateEffingoFolter() error {
	effingoPath, err := getEffingoFolderPath()
	if err != nil {
		return err
	}

	if err := os.Mkdir(effingoPath, os.ModePerm); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}

	return nil
}
