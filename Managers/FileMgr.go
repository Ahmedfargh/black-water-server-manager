package Managers

import (
	"bufio"
	"fmt"
	"os"
)

type FileInfo struct {
	Name  string      `json:"name"`
	IsDir bool        `json:"is_dir"`
	Size  int64       `json:"size"`
	Mode  os.FileMode `json:"mode"`
}
type FileSystemEvent struct {
	EventType string `json:"event_type"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Message   string `json:"message"`
}
type FileSystemEventMessage struct {
	Event    FileSystemEvent `json:"event"`
	Progress chan FileSystemEvent
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
func (fm *FileManager) CopyDirectory(srcPath string, dstPath string, progress chan FileSystemEvent) error {
	state, err := os.Stat(srcPath)
	fmt.Println("src", srcPath, "\n dst:-", dstPath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if state.IsDir() {
		err := os.MkdirAll(dstPath, state.Mode())
		if err != nil {
			fmt.Println(err)
			return err
		}
		dirs, err := os.ReadDir(srcPath)
		if err != nil {
			fmt.Println(err)

			return err
		}
		for _, dir := range dirs {
			err := fm.CopyDirectory(srcPath+"/"+dir.Name(), dstPath+"/"+dir.Name(), progress)
			if err != nil {
				fmt.Println(err)

				return err
			}
		}
	} else {
		return fm.copyFile(srcPath, dstPath, progress)
	}
	return nil
}
func (fm *FileManager) copyFile(source string, destination string, progress chan FileSystemEvent) error {
	srcFile, err := os.Open(source)
	srcFileInfo, err := srcFile.Stat()
	src_size := srcFileInfo.Size()
	if err != nil {
		fmt.Println(err)
		return err
	}
	progress <- FileSystemEvent{
		EventType: "copying_file",
		Path:      source,
		Size:      src_size,
		Message:   "Starting to copy file",
	}

	defer srcFile.Close()

	dstFile, err := os.Create(destination)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer dstFile.Close()

	scanner := bufio.NewReader(srcFile)
	buffer := make([]byte, 1024)
	for {
		n, err := scanner.Read(buffer)
		if err != nil {
			fmt.Println(err)
			break
		}
		_, err = dstFile.Write(buffer[:n])
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
