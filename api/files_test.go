package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	f1name = "file1.json"
	f2name = "file2.json"
)

var (
	defaultTimeoutSeconds = 4
	testInstance          *testing.T
	f1                    file1
	f2                    file2
)

type file1 struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

type file2 struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func setTestData() {
	data := `{
		"field1": "value1",
		"field2": "value2"
	}`

	ioutil.WriteFile(fmt.Sprintf("../testdata/%s", f1name), []byte(data), 0644)
	ioutil.WriteFile(fmt.Sprintf("../testdata/%s", f2name), []byte(data), 0644)
}

func TestWatchFilePath(t *testing.T) {
	testInstance = t

	done := make(chan bool)

	setTestData()

	defer func() {
		testInstance = nil
		close(done)

		os.Remove(fmt.Sprintf("../testdata/%s", f1name))
		os.Remove(fmt.Sprintf("../testdata/%s", f2name))
		time.Sleep(time.Second * time.Duration(defaultTimeoutSeconds))
	}()

	dfs := []DynamicFile{
		{
			File:       f1name,
			UpdateFunc: updateFile1,
		},
		{
			File:       f2name,
			UpdateFunc: updateFile2,
		},
	}

	fpw := NewFilePathWatcher("../testdata/%s", dfs)

	go fpw.Watch(done)
	// warmup
	time.Sleep(time.Second * time.Duration(defaultTimeoutSeconds))

	for _, df := range dfs {
		fpw.UpdateDynamicFile(fmt.Sprintf("../testdata/%s", df.File))
	}

	assert.Equal(t, "value1", f1.Field1)
	assert.Equal(t, "value2", f1.Field2)
	assert.Equal(t, "value1", f2.Field1)
	assert.Equal(t, "value2", f2.Field2)

	data := `{
		"field1": "value1",
		"field2": "changed"
	}`

	if err := ioutil.WriteFile(fmt.Sprintf("../testdata/%s", f1name), []byte(data), 0644); err != nil {
		assert.Fail(t, err.Error())
	}

	time.Sleep(time.Second * time.Duration(defaultTimeoutSeconds))

	b1, _ := ioutil.ReadFile(fmt.Sprintf("../testdata/%s", f1name))
	json.Unmarshal(b1, &f1)

	assert.Equal(t, "value1", f1.Field1)
	assert.Equal(t, "changed", f1.Field2)
	assert.Equal(t, "value1", f2.Field1)
	assert.Equal(t, "value2", f2.Field2)

	if err := ioutil.WriteFile(fmt.Sprintf("../testdata/%s", f2name), []byte(data), 0644); err != nil {
		assert.Fail(t, err.Error())
	}

	time.Sleep(time.Second * time.Duration(defaultTimeoutSeconds))

	b2, _ := ioutil.ReadFile(fmt.Sprintf("../testdata/%s", f2name))
	json.Unmarshal(b2, &f2)

	assert.Equal(t, "value1", f1.Field1)
	assert.Equal(t, "changed", f1.Field2)
	assert.Equal(t, "value1", f2.Field1)
	assert.Equal(t, "changed", f2.Field2)
}

func updateFile1(b []byte) {
	if err := json.Unmarshal(b, &f1); err != nil {
		assert.Fail(testInstance, err.Error())
	}
}

func updateFile2(b []byte) {
	if err := json.Unmarshal(b, &f2); err != nil {
		assert.Fail(testInstance, err.Error())
	}
}
