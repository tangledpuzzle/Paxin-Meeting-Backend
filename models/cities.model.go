package models

type City struct {
    ID         uint   `gorm:"primary_key"`
    Name       string
    District   string
    Population int
    Subject    string
    Lat        float64
    Lon        float64
}