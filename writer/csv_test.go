package writer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type Test struct {
	Name  string
	Value int
}

func TestCsv(t *testing.T) {

	csv, err := NewCswWriter("/tmp/test.csv")
	assert.Nil(t, err)

	csv.Write(Test{"asd", 1})
	csv.Close()
}
