// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// NewStateChangeConf is a wrapper around the retry.StateChangeConf data structure
// to enable compatibility with VCR testing
//
// When VCR testing is enabled in replay mode, the Delay and PollInterval fields are
// overridden to allow interactions to be replayed with no observable delay between
// state change refreshes.
func NewStateChangeConf(ctx context.Context, stateConf retry.StateChangeConf) *retry.StateChangeConf {
	if inContext, ok := conns.FromContext(ctx); ok && inContext.VCREnabled() {
		if mode, _ := vcr.Mode(); mode == recorder.ModeReplayOnly {
			stateConf.Delay = time.Millisecond * 1
			stateConf.PollInterval = time.Millisecond * 1
		}
	}

	return &stateConf
}
