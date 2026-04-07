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
	RestDays          []int                 `form:"rest_days"         json:"rest_days"` // 0=Sun, 6=Sat
	WorkDayStart      string                `form:"work_day_start"    json:"work_day_start"`
	WorkDayEnd        string                `form:"work_day_end"      json:"work_day_end"`
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
