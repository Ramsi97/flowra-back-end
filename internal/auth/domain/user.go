package domain

import (
	"mime/multipart"
	"time"
)

type User struct {
	ID                string                `form:"id"                json:"id"`
	FullName          string                `form:"full_name"          json:"full_name"`
	Email             string                `form:"email"             json:"email"`
	Password          string                `form:"password"          json:"-"`
	Gender            string                `form:"gender"            json:"gender"`
	ProfilePicture    *multipart.FileHeader `form:"profile_picture"   json:"-"`
	ProfilePictureURL string                `form:"-"                 json:"profile_picture_url"`
	CreatedAt         time.Time             `form:"created_at"        json:"created_at"`
}

type UserResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type AuthUseCase interface {
	Login(email, password string) (UserResponse, error)
	Logout() error
	Register(user *User) error
}
