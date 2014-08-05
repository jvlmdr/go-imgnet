package imgnet

import (
	"os"
	"path"
)

func Images(imgnetDir, wnid string) ([]string, error) {
	dir, err := os.Open(path.Join(imgnetDir, wnid))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	infos, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		files = append(files, info.Name())
	}
	return files, nil
}
