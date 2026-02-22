package authentication

import (
	"fmt"
	"strconv"

	service "github.com/ahmedfargh/server-manager/Authentication/Service"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Login(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		userResponse, err := authService.Login(loginRequest.Email, loginRequest.Password)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(200, userResponse)
	}
}

func Register(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := authService.Register(&user); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, user)
	}
}

func GetUsers(userCRUD *crud.UserCRUD) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}

		users, total, err := userCRUD.GetPaginatedUsers(page, limit)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		userMaps := make([]map[string]interface{}, len(users))
		for i, user := range users {
			userMaps[i] = user.ToMap()
		}

		c.JSON(200, gin.H{
			"data":  userMaps,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func GetUser(userCRUD *crud.UserCRUD) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid id"})
			return
		}
		user, err := userCRUD.GetUserByID(uint(id))
		if err != nil {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}
		c.JSON(200, user)
	}
}

func CreateUser(userCRUD *crud.UserCRUD) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth_user, err := userCRUD.GetUserByID(c.GetUint("userID"))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if !auth_user.HasPermission("create_users") {
			c.JSON(403, gin.H{"error": "permission denied"})
			return
		}
		type CreateUserRequest struct {
			User struct {
				Username string `json:"username" form:"username" binding:"required"`
				Email    string `json:"email" form:"email" binding:"required,email"`
				Password string `json:"password" form:"password" binding:"required"`
				RoleID   uint   `json:"role_id" form:"role_id" binding:"required"`
				Status   bool   `json:"status" form:"status"`
			} `json:"user" form:"user" binding:"required"`
		}

		var input CreateUserRequest
		if err := c.ShouldBind(&input); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.User.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}

		user := models.User{
			Username: input.User.Username,
			Email:    input.User.Email,
			Password: string(hashedPassword),
			RoleID:   input.User.RoleID,
			Status:   input.User.Status,
		}

		if err := userCRUD.CreateUser(&user); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		file, err := c.FormFile("image_path")
		if err == nil {
			filename := fmt.Sprintf("uploads/%s/%d_%s", user.Username, user.ID, file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.JSON(500, gin.H{"error": "Failed to save image"})
				return
			}
			user.ImagePath = filename
			if err := userCRUD.UpdateUser(&user, user.ID); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(201, user)
	}
}

func UpdateUser(userCRUD *crud.UserCRUD) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			fmt.Println(err)
			c.JSON(400, gin.H{"error": "invalid id"})
			return
		}
		user, err := userCRUD.GetUserByID(uint(id))
		if err != nil {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}
		if err := c.ShouldBind(user); err != nil {
			fmt.Println("error", err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if password := c.Request.FormValue("password"); password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err == nil {
				user.Password = string(hashedPassword)
			}
		}
		file, err := c.FormFile("image_path")
		if err == nil {
			filename := fmt.Sprintf("uploads/%d_%s", user.ID, file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.JSON(500, gin.H{"error": "Failed to save image"})
				return
			}
			user.ImagePath = filename
		}

		if err := userCRUD.UpdateUser(user, user.ID); err != nil {
			fmt.Println(err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, user.ToMap())
	}
}

func DeleteUser(userCRUD *crud.UserCRUD) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid id"})
			return
		}
		current_user := c.GetUint("userID")
		if current_user == uint(id) {
			fmt.Print(current_user)
			c.JSON(400, gin.H{"error": "invalid id"})
			return
		}

		user, err := userCRUD.GetUserByID(uint(id))
		if err != nil {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}
		if err := userCRUD.DeleteUser(user); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(204, nil)
	}
}

func GetRoles(roleCRUD *crud.RoleCRUD) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, err := roleCRUD.GetRoles()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, roles)
	}
}
func GetProfile(userCRUD *crud.UserCRUD) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(401, gin.H{"error": "User not authenticated"})
			return
		}

		userIDUint, ok := userID.(uint)
		if !ok {
			c.JSON(500, gin.H{"error": "Invalid user ID type"})
			return
		}

		user, err := userCRUD.GetUserByID(userIDUint)
		if err != nil {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}

		c.JSON(200, gin.H{
			"user_id": userIDUint,
			"user":    user.ToMap(),
			"message": "Profile retrieved successfully",
		})
	}
}
func UpdateProfile(userCRUD *crud.UserCRUD) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(400, gin.H{"error": "user not authenticated"})
			return
		}
		userIDUint, ok := userID.(uint)
		if !ok {
			c.JSON(500, gin.H{"error": "invalid user id type"})
			return
		}
		user, err := userCRUD.GetUserByID(userIDUint)
		if err != nil {
			c.JSON(404, gin.H{"error": "user not found"})
			return
		}
		if err := c.ShouldBind(user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		file, err := c.FormFile("image_path")
		if err == nil {
			filename := fmt.Sprintf("uploads/%d_%s", userIDUint, file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.JSON(500, gin.H{"error": "Failed to save image"})
				return
			}
			user.ImagePath = filename
		}
		if password := c.Request.FormValue("password"); password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err == nil {
				user.Password = string(hashedPassword)
			}
		}
		user.ID = userIDUint
		err = userCRUD.UpdateUser(user, userIDUint)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, user)
	}
}
func ListAll(userCRUD *crud.UserCRUD) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}

		users, total, err := userCRUD.GetPaginatedUsers(page, limit)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		userMaps := make([]map[string]interface{}, len(users))
		for i, user := range users {
			userMaps[i] = user.ToMap()
		}

		c.JSON(200, gin.H{
			"data":  userMaps,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}
