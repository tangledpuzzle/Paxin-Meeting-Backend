package models

type Stations struct {
    ID         uint   `gorm:"primary_key"`
    Hex        string
    Line   	   string
    Name       string
}
