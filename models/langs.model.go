package models

type MultilangTitle struct {
	En string `gorm:"null"`
	Ru string `gorm:"null"`
	Ka string `gorm:"null"`
	Es string `gorm:"null"`

	// Add more language fields as needed
}
