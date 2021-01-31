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

	csv, err := NewCswWriter("/tmp/test.csv", new(Test))
	assert.Nil(t, err)

	err = csv.Write(Test{"asd", 1})
	assert.Nil(t, err)

	err = csv.Close()
	assert.Nil(t, err)

}
