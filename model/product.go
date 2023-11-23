package model

import "time"

type Product struct {
	ID           int       `gorm:"id,primaryKey,autoIncrement"`
	ProductID    string    `gorm:"product_id,unique,not null"`
	Name         string    `gorm:"name,not null"`
	Description  string    `gorm:"description,not null"`
	ImageUrl     string    `gorm:"image_url,not null"`
	Price        string    `gorm:"price,not null"`
	Rating       string    `gorm:"rating,not null"`
	MerchantName string    `gorm:"merchant_name,not null"`
	CreatedDate  time.Time `gorm:"created_date,not null"`
	UpdatedDate *time.Time `gorm:"updated_date,not null"`
}
