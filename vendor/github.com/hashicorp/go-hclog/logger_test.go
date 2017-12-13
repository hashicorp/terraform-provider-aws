package hclog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("formats log entries", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing")

		str := buf.String()

		dataIdx := strings.IndexByte(str, ' ')

		// ts := str[:dataIdx]
		rest := str[dataIdx+1:]

		assert.Equal(t, "[INFO ] test: this is test: who=programmer why=testing\n", rest)
	})

	t.Run("quotes values with spaces", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing is fun")

		str := buf.String()

		dataIdx := strings.IndexByte(str, ' ')

		// ts := str[:dataIdx]
		rest := str[dataIdx+1:]

		assert.Equal(t, "[INFO ] test: this is test: who=programmer why=\"testing is fun\"\n", rest)
	})

	t.Run("outputs stack traces", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Info("who", "programmer", "why", "testing", Stacktrace())

		lines := strings.Split(buf.String(), "\n")

		require.True(t, len(lines) > 1)

		assert.Equal(t, "github.com/hashicorp/go-hclog.Stacktrace", lines[1])
	})

	t.Run("outputs stack traces with it's given a name", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Info("who", "programmer", "why", "testing", "foo", Stacktrace())

		lines := strings.Split(buf.String(), "\n")

		require.True(t, len(lines) > 1)

		assert.Equal(t, "github.com/hashicorp/go-hclog.Stacktrace", lines[1])
	})

	t.Run("includes the caller location", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:            "test",
			Output:          &buf,
			IncludeLocation: true,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing is fun")

		str := buf.String()

		dataIdx := strings.IndexByte(str, ' ')

		// ts := str[:dataIdx]
		rest := str[dataIdx+1:]

		// This test will break if you move this around, it's line dependent, just fyi
		assert.Equal(t, "[INFO ] go-hclog/logger_test.go:100: test: this is test: who=programmer why=\"testing is fun\"\n", rest)
	})

	t.Run("prefixes the name", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			// No name!
			Output: &buf,
		})

		logger.Info("this is test")
		str := buf.String()
		dataIdx := strings.IndexByte(str, ' ')
		rest := str[dataIdx+1:]
		assert.Equal(t, "[INFO ] this is test\n", rest)

		buf.Reset()

		another := logger.Named("sublogger")
		another.Info("this is test")
		str = buf.String()
		dataIdx = strings.IndexByte(str, ' ')
		rest = str[dataIdx+1:]
		assert.Equal(t, "[INFO ] sublogger: this is test\n", rest)
	})
}

func TestLogger_JSON(t *testing.T) {
	t.Run("json formatting", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:       "test",
			Output:     &buf,
			JSONFormat: true,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing is fun")

		b := buf.Bytes()

		var raw map[string]interface{}
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "this is test", raw["@message"])
		assert.Equal(t, "programmer", raw["who"])
		assert.Equal(t, "testing is fun", raw["why"])
	})
	t.Run("json formatting error type", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:       "test",
			Output:     &buf,
			JSONFormat: true,
		})

		errMsg := errors.New("this is an error")
		logger.Info("this is test", "who", "programmer", "err", errMsg)

		b := buf.Bytes()

		var raw map[string]interface{}
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "this is test", raw["@message"])
		assert.Equal(t, "programmer", raw["who"])
		assert.Equal(t, errMsg.Error(), raw["err"])
	})
	t.Run("json formatting custom error type json marshaler", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:       "test",
			Output:     &buf,
			JSONFormat: true,
		})

		errMsg := &customErrJSON{"this is an error"}
		rawMsg, err := errMsg.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		expectedMsg, err := strconv.Unquote(string(rawMsg))
		if err != nil {
			t.Fatal(err)
		}

		logger.Info("this is test", "who", "programmer", "err", errMsg)

		b := buf.Bytes()

		var raw map[string]interface{}
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "this is test", raw["@message"])
		assert.Equal(t, "programmer", raw["who"])
		assert.Equal(t, expectedMsg, raw["err"])
	})
	t.Run("json formatting custom error type text marshaler", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:       "test",
			Output:     &buf,
			JSONFormat: true,
		})

		errMsg := &customErrText{"this is an error"}
		rawMsg, err := errMsg.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		expectedMsg := string(rawMsg)

		logger.Info("this is test", "who", "programmer", "err", errMsg)

		b := buf.Bytes()

		var raw map[string]interface{}
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "this is test", raw["@message"])
		assert.Equal(t, "programmer", raw["who"])
		assert.Equal(t, expectedMsg, raw["err"])
	})
}

type customErrJSON struct {
	Message string
}

// error impl.
func (c *customErrJSON) Error() string {
	return c.Message
}

// json.Marshaler impl.
func (c customErrJSON) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(fmt.Sprintf("json-marshaler: %s", c.Message))), nil
}

type customErrText struct {
	Message string
}

// error impl.
func (c *customErrText) Error() string {
	return c.Message
}

// text.Marshaler impl.
func (c customErrText) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("text-marshaler: %s", c.Message)), nil
}

func BenchmarkLogger(b *testing.B) {
	b.Run("info with 10 pairs", func(b *testing.B) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:            "test",
			Output:          &buf,
			IncludeLocation: true,
		})

		for i := 0; i < b.N; i++ {
			logger.Info("this is some message",
				"name", "foo",
				"what", "benchmarking yourself",
				"why", "to see what's slow",
				"k4", "value",
				"k5", "value",
				"k6", "value",
				"k7", "value",
				"k8", "value",
				"k9", "value",
				"k10", "value",
			)
		}
	})
}
