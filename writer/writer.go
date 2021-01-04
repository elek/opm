package writer

import (
	"github.com/pkg/errors"
)

type Writer interface {
	Write(data interface{}) error
	Close() error
}

func NewWriter(name string, format string, entitiy interface{}) (Writer, error) {
	if format == "parquet" {
		return NewParquetWriter(name+".parquet", entitiy)
	} else if format == "csv" {
		return NewCswWriter(name+".csv", entitiy)
	} else {
		return nil, errors.New("No such writer implementation: " + format + ". Use parquet or csv.")
	}
}
