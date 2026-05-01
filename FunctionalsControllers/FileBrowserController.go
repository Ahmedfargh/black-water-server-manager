package functionalscontrollers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/ahmedfargh/server-manager/Managers"
	"github.com/gin-gonic/gin"
)

func FileBrowserHandler(c *gin.Context) {
	dirPath := c.Query("path")
	if dirPath == "" {
		dirPath = "." // Default to current directory if no path is provided
	}
	// Clean the path to prevent directory traversal attacks
	cleanPath := filepath.Clean(dirPath)

	fileManager := Managers.FileManager{}
	files, err := fileManager.ListDirectory(cleanPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(files)
	c.JSON(http.StatusOK, gin.H{"files": files})
}
