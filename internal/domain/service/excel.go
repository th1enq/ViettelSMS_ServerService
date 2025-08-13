package srv

import "github.com/th1enq/ViettelSMS_ServerService/internal/domain/entity"

type XLSXService interface {
	GetRows(filePath string) ([][]string, error)
	Validate(row []string) error
	Parse(row []string) (*entity.Server, error)
}
