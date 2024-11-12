package service

import (
	"errors"
	"fmt"
	"strings"

	cyclonedx "github.com/CycloneDX/cyclonedx-go"
	"github.com/kerezsiz42/scanner-operator2/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var InvalidCycloneDXBOM = errors.New("invalid CycloneDX BOM")

type ScanServiceInterface interface {
	GetScanResult(imageId string) (*database.ScanResult, error)
	ListScanResults() ([]*database.ScanResult, error)
	DeleteScanResult(imageId string) error
	UpsertScanResult(imageId string, report string) (*database.ScanResult, error)
}

type ScanService struct {
	db *gorm.DB
}

func NewScanService(db *gorm.DB) *ScanService {
	return &ScanService{
		db: db,
	}
}

func (s *ScanService) GetScanResult(imageId string) (*database.ScanResult, error) {
	scanResult := database.ScanResult{}
	res := s.db.First(&scanResult, "image_id = ?", imageId)
	if res.Error != nil {
		return nil, fmt.Errorf("error while getting ScanResult: %w", res.Error)
	}

	return &scanResult, nil
}

func (s *ScanService) ListScanResults() ([]*database.ScanResult, error) {
	scanResults := []*database.ScanResult{}
	res := s.db.Find(&scanResults)
	if res.Error != nil {
		return nil, fmt.Errorf("error while listing ScanResults: %w", res.Error)
	}

	return scanResults, nil
}

func (s *ScanService) DeleteScanResult(imageId string) error {
	res := s.db.Where("image_id = ?", imageId).Delete(&database.ScanResult{})
	if res.Error != nil {
		return fmt.Errorf("error while deleting ScanResult: %w", res.Error)
	}

	return nil
}

func (s *ScanService) UpsertScanResult(imageId string, report string) (*database.ScanResult, error) {
	bom := cyclonedx.BOM{}
	reader := strings.NewReader(report)
	decoder := cyclonedx.NewBOMDecoder(reader, cyclonedx.BOMFileFormatJSON)
	if err := decoder.Decode(&bom); err != nil {
		return nil, fmt.Errorf("%w: %w", InvalidCycloneDXBOM, err)
	}

	scanResult := database.ScanResult{
		ImageID: imageId,
		Report:  report,
	}

	res := s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&scanResult)
	if res.Error != nil {
		return nil, fmt.Errorf("error while inserting ScanResult: %w", res.Error)
	}

	return &scanResult, nil
}
