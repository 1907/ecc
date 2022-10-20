package files

import (
	"fmt"
	"io/ioutil"
	"os"
)

func checkNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

func checkPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

func isNotExistDir(src string) bool {
	return checkNotExist(src)
}

func ReadDir(dir string) ([]os.FileInfo, error) {
	perm := checkPermission(dir)
	if perm == true {
		return nil, fmt.Errorf("permission denied dir: %s", dir)
	}

	if isNotExistDir(dir) {
		return nil, fmt.Errorf("does not exist dir: %s", dir)
	}

	files, err := ioutil.ReadDir(dir)
	if err == nil {
		return files, err
	}

	return nil, fmt.Errorf("ReadDir: %s, err: %v", dir, err)
}
