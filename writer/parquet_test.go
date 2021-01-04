package writer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type Test3 struct {
	Name string `parquet:"type=UTF8, encoding=PLAIN_DICTIONARY"`
	Age  int32  `parquet:"type=INT32, encoding=PLAIN"`
}

func TestParquet(t *testing.T) {

	pq, err := NewParquetWriter("/tmp/test.parquet", new(Test3))
	assert.Nil(t, err)

	for i := 0; i < 100; i++ {
		stu := Test3{
			"asd",
			int32(i),
		}
		err = pq.Write(stu)
		assert.Nil(t, err)
	}

	err = pq.Close()
	assert.Nil(t, err)
}
