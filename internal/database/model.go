package database

type ScanResult struct {
	ImageID string `gorm:"primarykey;type:TEXT"`
	Report  string `gorm:"not null;type:TEXT"`
}
