package domain

import (
	"mime/multipart"
	"time"
)

type User struct{
	ID string `form:"id"`
	FullName string `form:"full_name"`
	Email string `form:"email"`
	Password string `form:"password"`
	Gender string `form:"gender"`	
	ProfilePicture *multipart.FileHeader `form:"profile_picture"`
	CreatedAt time.Time `form:"created_at"`
}

type UserResponse struct {
	Token string `json:"token"`
	User User `json:"user"`
} 

type AuthUseCase interface {
	Login(email, password string) (UserResponse, error)
	Logout() error
	Register(user *User) error
}