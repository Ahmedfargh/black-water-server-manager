package functionalscontrollers

import (
	"net/http"
	"strconv"
	"strings"

	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	sshservice "github.com/ahmedfargh/server-manager/Services/SSH"
	"github.com/gin-gonic/gin"
)

type GenerateSSHKeyRequest struct {
	Name                  string `json:"name" binding:"required"`
	KeyType               string `json:"key_type" binding:"required"` // ed25519 or rsa
	Comment               string `json:"comment"`
	AddToAuthorizedKeys   bool   `json:"add_to_authorized_keys"`
	RsaBits               int    `json:"rsa_bits"` // optional, default 4096
}

type ImportSSHKeyRequest struct {
	Name                string `json:"name" binding:"required"`
	PublicKey           string `json:"public_key" binding:"required"`
	Comment             string `json:"comment"`
	AddToAuthorizedKeys bool   `json:"add_to_authorized_keys"`
}

func GenerateSSHKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GenerateSSHKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		keyService := sshservice.NewSSHKeyService()
		generated, err := keyService.GenerateKeyPair(req.KeyType, req.Comment, req.RsaBits)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Optional: Add to authorized_keys
		if req.AddToAuthorizedKeys {
			if err := keyService.AddToAuthorizedKeys(generated.PublicKey); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Key generated but failed to add to authorized_keys: " + err.Error()})
				return
			}
		}

		var userID uint
		if uid, exists := c.Get("userID"); exists {
			if u, ok := uid.(uint); ok {
				userID = u
			}
		}

		sshModel := models.SSHKey{
			Name:                  strings.TrimSpace(req.Name),
			KeyType:               generated.KeyType,
			PublicKey:             generated.PublicKey,
			Fingerprint:           generated.Fingerprint,
			Comment:               generated.Comment,
			AddedToAuthorizedKeys: req.AddToAuthorizedKeys,
			UserID:                userID,
		}

		sshCrud := crud.NewSSHKeyCRUD(nil)
		if err := sshCrud.CreateSSHKey(&sshModel); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save key in database: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":          "SSH Key generated successfully",
			"key":              sshModel,
			"private_key_pem":  generated.PrivateKeyPEM,
		})
	}
}

func ImportSSHKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ImportSSHKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		keyService := sshservice.NewSSHKeyService()
		pubKey, comment, fingerprint, err := keyService.ParsePublicKey(req.PublicKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Comment != "" {
			comment = req.Comment
		}

		keyType := pubKey.Type()
		if req.AddToAuthorizedKeys {
			if err := keyService.AddToAuthorizedKeys(req.PublicKey); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to append to authorized_keys: " + err.Error()})
				return
			}
		}

		var userID uint
		if uid, exists := c.Get("userID"); exists {
			if u, ok := uid.(uint); ok {
				userID = u
			}
		}

		sshModel := models.SSHKey{
			Name:                  strings.TrimSpace(req.Name),
			KeyType:               keyType,
			PublicKey:             strings.TrimSpace(req.PublicKey),
			Fingerprint:           fingerprint,
			Comment:               comment,
			AddedToAuthorizedKeys: req.AddToAuthorizedKeys,
			UserID:                userID,
		}

		sshCrud := crud.NewSSHKeyCRUD(nil)
		if err := sshCrud.CreateSSHKey(&sshModel); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save key in database: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "SSH Public Key imported successfully",
			"key":     sshModel,
		})
	}
}

func GetSSHKeysHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 10
		}

		sshCrud := crud.NewSSHKeyCRUD(nil)
		keys, total, err := sshCrud.GetSSHKeys(uint(page), uint(limit))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		keyService := sshservice.NewSSHKeyService()
		for i := range keys {
			inAuth, _ := keyService.IsInAuthorizedKeys(keys[i].PublicKey)
			keys[i].AddedToAuthorizedKeys = inAuth
		}

		c.JSON(http.StatusOK, gin.H{
			"keys":  keys,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func DeleteSSHKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
			return
		}

		sshCrud := crud.NewSSHKeyCRUD(nil)
		key, err := sshCrud.GetSSHKeyByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
			return
		}

		// Remove from authorized_keys
		keyService := sshservice.NewSSHKeyService()
		_ = keyService.RemoveFromAuthorizedKeys(key.PublicKey)

		if err := sshCrud.DeleteSSHKey(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "SSH Key removed successfully"})
	}
}

func ToggleAuthorizedSSHKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
			return
		}

		sshCrud := crud.NewSSHKeyCRUD(nil)
		key, err := sshCrud.GetSSHKeyByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
			return
		}

		keyService := sshservice.NewSSHKeyService()
		isAuth, err := keyService.IsInAuthorizedKeys(key.PublicKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var newStatus bool
		if isAuth {
			if err := keyService.RemoveFromAuthorizedKeys(key.PublicKey); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove key from authorized_keys: " + err.Error()})
				return
			}
			newStatus = false
		} else {
			if err := keyService.AddToAuthorizedKeys(key.PublicKey); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add key to authorized_keys: " + err.Error()})
				return
			}
			newStatus = true
		}

		key.AddedToAuthorizedKeys = newStatus
		_ = sshCrud.UpdateSSHKey(key, uint(id))

		c.JSON(http.StatusOK, gin.H{
			"message":                  "SSH key authorization toggled successfully",
			"added_to_authorized_keys": newStatus,
		})
	}
}
