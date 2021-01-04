package writer

import (
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/source"
	"github.com/xitongsys/parquet-go/writer"
)

type ParquetWriter struct {
	parquetFile source.ParquetFile
	writer      *writer.ParquetWriter
}

type Test2 struct {
	Name  string `parquet:"name=name, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Value int32  `parquet:"name=age, type=INT32, encoding=PLAIN"`
}

func NewParquetWriter(path string, obj interface{}) (*ParquetWriter, error) {
	fw, err := local.NewLocalFileWriter(path)
	if err != nil {
		return &ParquetWriter{}, err
	}
	w, err := writer.NewParquetWriter(fw, obj, 4)
	if err != nil {
		return &ParquetWriter{}, err
	}

	return &ParquetWriter{
		parquetFile: fw,
		writer:      w,
	}, nil
}
func (pw *ParquetWriter) Write(data interface{}) error {
	return pw.writer.Write(data)
}

func (pw *ParquetWriter) Close() error {

	if pw.writer != nil {
		err := pw.writer.WriteStop()
		if err != nil {
			return err
		}

	}
	return pw.parquetFile.Close()
}
