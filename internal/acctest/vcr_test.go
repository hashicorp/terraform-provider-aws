// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestRandInt(t *testing.T) {
	ctx := acctest.Context(t)

	t.Setenv("VCR_PATH", t.TempDir())

	t.Setenv("VCR_MODE", "RECORD_ONLY")
	rec1 := acctest.RandInt(t)
	rec2 := acctest.RandInt(t)
	acctest.CloseVCRRecorder(ctx, t)

	t.Setenv("VCR_MODE", "REPLAY_ONLY")
	rep1 := acctest.RandInt(t)
	rep2 := acctest.RandInt(t)

	if rep1 != rec1 {
		t.Errorf("REPLAY_ONLY: %d, RECORD_ONLY: %d", rep1, rec1)
	}
	if rep2 != rec2 {
		t.Errorf("REPLAY_ONLY: %d, RECORD_ONLY: %d", rep2, rec2)
	}
}

func TestRandomWithPrefix(t *testing.T) {
	ctx := acctest.Context(t)

	t.Setenv("VCR_PATH", t.TempDir())

	t.Setenv("VCR_MODE", "RECORD_ONLY")
	rec1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rec2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	acctest.CloseVCRRecorder(ctx, t)

	t.Setenv("VCR_MODE", "REPLAY_ONLY")
	rep1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rep2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	if rep1 != rec1 {
		t.Errorf("REPLAY_ONLY: %s, RECORD_ONLY: %s", rep1, rec1)
	}
	if rep2 != rec2 {
		t.Errorf("REPLAY_ONLY: %s, RECORD_ONLY: %s", rep2, rec2)
	}
}
