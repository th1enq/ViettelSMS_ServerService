package domain

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/wire"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type XLSXService interface {
	GetRows(filePath string) ([][]string, error)
	Validate(row []string) error
	Parse(row []string) (*Server, error)
}

var ExcelizeServiceSet = wire.NewSet(NewExcelizeService)

type excelizeService struct {
	logger *zap.Logger
}

func NewExcelizeService(logger *zap.Logger) XLSXService {
	return &excelizeService{
		logger: logger,
	}
}

func (e *excelizeService) GetRows(filePath string) ([][]string, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		e.logger.Error("failed to open Excel file", zap.String("filePath", filePath), zap.Error(err))
		return nil, err
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		e.logger.Warn("no sheets found in Excel file", zap.String("filePath", filePath))
		return nil, ErrInvalidFile
	}
	rows, err := file.GetRows(sheets[0])
	if err != nil {
		e.logger.Error("failed to get rows from Excel file", zap.String("filePath", filePath), zap.Error(err))
		return nil, ErrInvalidFile
	}
	return rows, nil
}

func (e *excelizeService) Validate(row []string) error {
	expectedHeaders := []string{
		"server_id",
		"server_name",
		"ipv4",
		"location",
		"os",
		"interval_time",
	}

	if len(row) < len(expectedHeaders) {
		return fmt.Errorf("invalid header: expected at least %d columns, got %d", len(expectedHeaders), len(row))
	}

	for i, expected := range expectedHeaders {
		if i >= len(row) || strings.TrimSpace(strings.ToLower(row[i])) != expected {
			return ErrInvalidFile
		}
	}

	return nil
}

func (e *excelizeService) Parse(row []string) (*Server, error) {
	if len(row) < 6 {
		return nil, fmt.Errorf("invalid row: expected at least 7 columns, got %d", len(row))
	}

	server := &Server{
		ServerID:     strings.TrimSpace(row[0]),
		ServerName:   strings.TrimSpace(row[1]),
		IPv4:         strings.TrimSpace(row[2]),
		Location:     strings.TrimSpace(row[3]),
		OS:           strings.TrimSpace(row[4]),
		IntervalTime: uint64(0),
	}
	if server.ServerID == "" {
		return nil, fmt.Errorf("invalid row: server_id is required")
	}
	if server.ServerName == "" {
		return nil, fmt.Errorf("invalid row: server_name is required")
	}
	if server.IPv4 == "" {
		return nil, fmt.Errorf("invalid row: ipv4 is required")
	}
	intervalTime := strings.TrimSpace(row[5])
	if intervalTime == "" {
		return nil, fmt.Errorf("invalid row: interval_time is required")
	}
	parsedIntervalTime, err := strconv.ParseUint(intervalTime, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid row: interval_time must be a valid number")
	}
	server.IntervalTime = parsedIntervalTime
	return server, nil
}
