package database

type ApplicationEmoji struct {
	Id       string `gorm:"primaryKey"`
	Name     string
	Animated bool
}
