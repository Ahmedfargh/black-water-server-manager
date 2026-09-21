package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	controller "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func SSHRoutes(router *gin.Engine) {
	sshGroup := router.Group("/ssh")
	{
		// Keys management
		sshGroup.POST("/keys/generate", authentication.AuthMiddleware(), authentication.CheckRole("manage_ssh_keys"), controller.GenerateSSHKeyHandler())
		sshGroup.POST("/keys/import", authentication.AuthMiddleware(), authentication.CheckRole("manage_ssh_keys"), controller.ImportSSHKeyHandler())
		sshGroup.GET("/keys/list", authentication.AuthMiddleware(), authentication.CheckRole("read_ssh_keys"), controller.GetSSHKeysHandler())
		sshGroup.DELETE("/keys/:id", authentication.AuthMiddleware(), authentication.CheckRole("manage_ssh_keys"), controller.DeleteSSHKeyHandler())
		sshGroup.POST("/keys/:id/toggle-auth", authentication.AuthMiddleware(), authentication.CheckRole("manage_ssh_keys"), controller.ToggleAuthorizedSSHKeyHandler())
	}
}
