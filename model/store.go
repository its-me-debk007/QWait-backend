package model

import (
	"github.com/lib/pq"
)

type Store struct {
	ID               uint           `json:"id" gorm:"primaryKey;autoIncrement:true"`
	Name             string         `json:"name"`
	Location         string         `json:"location"`
	Category         string         `json:"category"`
	Latitude         float64        `json:"-"`
	Longitude        float64        `json:"-"`
	Counters         int            `json:"counters"`
	Customers        pq.StringArray `json:"customers" gorm:"type:text[]"`
	WaitingTime      int            `json:"waiting_time"`
	AvgTimePerPerson int            `json:"avg_time_per_person"`
	About            string         `json:"about"`
	Timings          string         `json:"timings"`
	ProfilePic       string         `json:"profile_pic" gorm:"default:https://res.cloudinary.com/debk007cloud/image/upload/v1668334132/low-resolution-splashes-wallpaper-preview_weaxun.jpg"`
	StoreStats       StoreStats     `json:"store_stats" gorm:"embedded"`
}

type StoreStats struct {
	CustomersThisMonth int     `json:"customers_this_month"`
	CustomersPrevMonth int     `json:"customers_prev_month"`
	CustomerIncrement  float32 `json:"customer_increment"`
}
