//go:build gofuzz
// +build gofuzz

package ini

import (
	"bytes"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func Fuzz(data []byte) int {
	b := bytes.NewReader(data)

	if _, err := Parse(b); err != nil {
		return 0
	}

	return 1
}
