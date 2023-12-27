// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestRandInt(t *testing.T) { //nolint:paralleltest
	t.Setenv("VCR_PATH", t.TempDir())

	t.Setenv("VCR_MODE", "RECORDING")
	rec1 := acctest.RandInt(t)
	rec2 := acctest.RandInt(t)
	acctest.CloseVCRRecorder(t)

	t.Setenv("VCR_MODE", "REPLAYING")
	rep1 := acctest.RandInt(t)
	rep2 := acctest.RandInt(t)

	if rep1 != rec1 {
		t.Errorf("REPLAYING: %d, RECORDING: %d", rep1, rec1)
	}
	if rep2 != rec2 {
		t.Errorf("REPLAYING: %d, RECORDING: %d", rep2, rec2)
	}
}

func TestRandomWithPrefix(t *testing.T) { //nolint:paralleltest
	t.Setenv("VCR_PATH", t.TempDir())

	t.Setenv("VCR_MODE", "RECORDING")
	rec1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rec2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	acctest.CloseVCRRecorder(t)

	t.Setenv("VCR_MODE", "REPLAYING")
	rep1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rep2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	if rep1 != rec1 {
		t.Errorf("REPLAYING: %s, RECORDING: %s", rep1, rec1)
	}
	if rep2 != rec2 {
		t.Errorf("REPLAYING: %s, RECORDING: %s", rep2, rec2)
	}
}
