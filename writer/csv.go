package writer

import (
	"compress/gzip"
	csv2 "encoding/csv"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type CsvWriter struct {
	path          string
	headerWritten bool
	writer        *csv2.Writer
	file          io.WriteCloser
	gzip          bool
}

func NewCswWriter(path string, entity interface{}) (*CsvWriter, error) {
	result := CsvWriter{}
	reposFile, err := os.Create(path)
	if err != nil {
		return &result, err
	}

	csv := csv2.NewWriter(reposFile)

	result.writer = csv
	result.file = reposFile
	return &result, nil
}


func NewCswGzWriter(path string, entity interface{}) (*CsvWriter, error) {
	result := CsvWriter{}
	reposFile, err := os.Create(path + ".gz")
	if err != nil {
		return &result, err
	}

	result.file = gzip.NewWriter(reposFile)

	csv := csv2.NewWriter(result.file)

	result.writer = csv
	return &result, nil
}


func (writer *CsvWriter) Write(data interface{}) error {
	t := reflect.TypeOf(data)
	if !writer.headerWritten {
		headers := make([]string, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			headers[i] = strings.ToLower(t.Field(i).Name)
		}
		writer.writer.Write(headers)
		writer.headerWritten = true
	}
	values := make([]string, t.NumField())
	reflected := reflect.ValueOf(data)
	for i := 0; i < t.NumField(); i++ {
		value := reflected.Field(i).Interface()
		switch v := value.(type) {
		case string:
			values[i] = v
		case *string:
			if v == nil {
				values[i] = ""
			} else {
				values[i] = *v
			}
		case int:
			values[i] = strconv.Itoa(v)
		case int32:
			values[i] = fmt.Sprintf("%d", v)
		case int64:
			values[i] = fmt.Sprintf("%d", v)
		case *int64:
			values[i] = fmt.Sprintf("%d", v)
		case time.Time:
			values[i] = v.Format(time.RFC3339)
		case bool:
			if v {
				values[i] = "true"
			} else {
				values[i] = "false"
			}
		default:
			log.Warn().Msg("Unknown field type: " + reflected.Field(i).String() + ": " + reflected.Field(i).Type().Name())
			values[i] = "???"

		}

	}
	return writer.writer.Write(values)
}

func (writer *CsvWriter) Close() error {
	writer.writer.Flush()
	return writer.file.Close()
}
