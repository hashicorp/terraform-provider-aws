// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
)

func BenchmarkValidate(b *testing.B) {
	for b.Loop() {
		Validate[types.AclPermission]()
	}
}
