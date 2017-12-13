package hclog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdlogAdapter(t *testing.T) {
	t.Run("picks debug level", func(t *testing.T) {
		var s stdlogAdapter

		level, rest := s.pickLevel("[DEBUG] coffee?")

		assert.Equal(t, Debug, level)
		assert.Equal(t, "coffee?", rest)
	})

	t.Run("picks trace level", func(t *testing.T) {
		var s stdlogAdapter

		level, rest := s.pickLevel("[TRACE] coffee?")

		assert.Equal(t, Trace, level)
		assert.Equal(t, "coffee?", rest)
	})

	t.Run("picks info level", func(t *testing.T) {
		var s stdlogAdapter

		level, rest := s.pickLevel("[INFO] coffee?")

		assert.Equal(t, Info, level)
		assert.Equal(t, "coffee?", rest)
	})

	t.Run("picks warn level", func(t *testing.T) {
		var s stdlogAdapter

		level, rest := s.pickLevel("[WARN] coffee?")

		assert.Equal(t, Warn, level)
		assert.Equal(t, "coffee?", rest)
	})

	t.Run("picks error level", func(t *testing.T) {
		var s stdlogAdapter

		level, rest := s.pickLevel("[ERROR] coffee?")

		assert.Equal(t, Error, level)
		assert.Equal(t, "coffee?", rest)
	})

	t.Run("picks error as err level", func(t *testing.T) {
		var s stdlogAdapter

		level, rest := s.pickLevel("[ERR] coffee?")

		assert.Equal(t, Error, level)
		assert.Equal(t, "coffee?", rest)
	})
}
