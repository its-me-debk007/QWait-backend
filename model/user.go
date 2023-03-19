package model

import "time"

type User struct {
	CreatedAt  time.Time `json:"created_at"`
	Name       string    `json:"name"`
	Email      string    `json:"email"    gorm:"unique"`
	PhoneNo    string    `json:"phone_no"    binding:"required"    gorm:"primary_key"`
	VerCode    int64     `json:"-"`
	ProfilePic string    `json:"profile_pic"    gorm:"default:https://res.cloudinary.com/debk007cloud/image/upload/v1668334132/low-resolution-splashes-wallpaper-preview_weaxun.jpg"`
}
