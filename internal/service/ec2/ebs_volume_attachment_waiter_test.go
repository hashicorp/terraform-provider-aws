// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// sequenceRefreshFunc returns a StateRefreshFunc that walks through states one
// by one on each call. The final call returns an empty state (not-found).
func sequenceRefreshFunc(states []awstypes.VolumeAttachmentState) retry.StateRefreshFunc {
	i := 0
	return func(_ context.Context) (any, string, error) {
		if i >= len(states) {
			return nil, "", nil // not found → target reached
		}
		s := states[i]
		i++
		return &awstypes.VolumeAttachment{State: s}, string(s), nil
	}
}

// TestVolumeAttachmentWaiterDeleted_attachedTransitionSucceeds verifies that the
// waiter handles the attached → detaching → gone sequence without error.
// This is the regression test for https://github.com/hashicorp/terraform-provider-aws/issues/47314:
// AWS does not immediately transition to detaching after DetachVolume is called,
// so attached must be a valid pending state.
func TestVolumeAttachmentWaiterDeleted_attachedTransitionSucceeds(t *testing.T) {
	t.Parallel()

	conf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.VolumeAttachmentStateAttached,
			awstypes.VolumeAttachmentStateDetaching,
		),
		Target: []string{},
		Refresh: sequenceRefreshFunc([]awstypes.VolumeAttachmentState{
			awstypes.VolumeAttachmentStateAttached,
			awstypes.VolumeAttachmentStateDetaching,
		}),
		Timeout:    5 * time.Second,
		Delay:      0,
		MinTimeout: 0,
	}

	_, err := conf.WaitForStateContext(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
}

// TestVolumeAttachmentWaiterDeleted_attachedOnlyPending_fails shows the pre-fix
// behaviour: with only detaching in Pending, seeing attached immediately fails.
func TestVolumeAttachmentWaiterDeleted_attachedOnlyPending_fails(t *testing.T) {
	t.Parallel()

	conf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VolumeAttachmentStateDetaching),
		Target:  []string{},
		Refresh: sequenceRefreshFunc([]awstypes.VolumeAttachmentState{
			awstypes.VolumeAttachmentStateAttached,
		}),
		Timeout:    5 * time.Second,
		Delay:      0,
		MinTimeout: 0,
	}

	_, err := conf.WaitForStateContext(context.Background())
	if err == nil {
		t.Fatal("expected an error for unexpected 'attached' state, got none")
	}
}
