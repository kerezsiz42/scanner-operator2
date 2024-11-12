package database

type ScanResult struct {
	ImageID string `gorm:"primarykey;type:VARCHAR"`
	Report  string `gorm:"not null;type:TEXT"`
}
