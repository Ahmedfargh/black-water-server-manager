package Managers

import "os"

type FileInfo struct {
	Name  string      `json:"name"`
	IsDir bool        `json:"is_dir"`
	Size  int64       `json:"size"`
	Mode  os.FileMode `json:"mode"`
}
type FileManager struct {
}

func (fm *FileManager) ListDirectory(dirPath string) ([]FileInfo, error) {
	dirs, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	dir_list := make([]FileInfo, 0)
	for _, dir := range dirs {
		info, err := dir.Info()
		if err != nil {
			continue
		}
		dir_list = append(dir_list, FileInfo{
			Name:  dir.Name(),
			IsDir: dir.IsDir(),
			Size:  info.Size(),
			Mode:  info.Mode(),
		})

	}
	return dir_list, nil
}
