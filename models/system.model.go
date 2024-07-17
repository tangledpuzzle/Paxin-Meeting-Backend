package models

type System struct {
	ID               uint `gorm:"primary_key"`
	LatestVersionIOS string
}
