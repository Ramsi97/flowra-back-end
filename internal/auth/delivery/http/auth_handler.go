package http

import (
	"net/http"

	"github.com/Ramsi97/flowra-back-end/internal/auth/domain"
	"github.com/Ramsi97/flowra-back-end/pkg/cloudinary"
	"github.com/gin-gonic/gin"
)

// AuthHandler holds the use-case dependency.
type AuthHandler struct {
	usecase domain.AuthUseCase
	cld     *cloudinary.Client
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(uc domain.AuthUseCase, cld *cloudinary.Client) *AuthHandler {
	return &AuthHandler{
		usecase: uc,
		cld:     cld,
	}
}

// Register godoc
// POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBind(&user); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle profile picture upload if file is present
	if user.ProfilePicture != nil {
		file, err := user.ProfilePicture.Open()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open profile picture"})
			return
		}
		defer file.Close()

		url, err := h.cld.UploadImage(c.Request.Context(), file, user.Email)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload profile picture"})
			return
		}
		user.ProfilePictureURL = url
	}

	if err := h.usecase.Register(&user); err != nil {
		c.Error(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "registered successfully"})
}

// Login godoc
// POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.usecase.Login(req.Email, req.Password)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// POST /auth/logout  (protected)
func (h *AuthHandler) Logout(c *gin.Context) {
	if err := h.usecase.Logout(); err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// UpdateProfile godoc
// PUT /auth/profile (protected)
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")
	var user domain.User
	if err := c.ShouldBind(&user); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle updated profile picture upload if file is present
	if user.ProfilePicture != nil {
		file, err := user.ProfilePicture.Open()
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open profile picture"})
			return
		}
		defer file.Close()

		url, err := h.cld.UploadImage(c.Request.Context(), file, user.Email)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload profile picture"})
			return
		}
		user.ProfilePictureURL = url
	}

	if err := h.usecase.UpdateProfile(c.Request.Context(), userID, &user); err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}
