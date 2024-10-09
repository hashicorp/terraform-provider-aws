// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/google/go-cmp/cmp"
)

func TestValues(t *testing.T) {
	t.Parallel()

	want := []string{
		"READ",
		"WRITE",
		"READ_ACP",
		"WRITE_ACP",
		"FULL_CONTROL",
	}
	got := Values[types.AclPermission]()

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}
