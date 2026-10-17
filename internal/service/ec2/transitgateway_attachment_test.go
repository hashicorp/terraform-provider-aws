// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestTransitGatewayAttachmentStateIsInactive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		state        ec2types.TransitGatewayAttachmentState
		wantInactive bool
	}{
		"initiating": {
			state: ec2types.TransitGatewayAttachmentStateInitiating,
		},
		"initiating request": {
			state: ec2types.TransitGatewayAttachmentStateInitiatingRequest,
		},
		"pending acceptance": {
			state: ec2types.TransitGatewayAttachmentStatePendingAcceptance,
		},
		"pending": {
			state: ec2types.TransitGatewayAttachmentStatePending,
		},
		"available": {
			state: ec2types.TransitGatewayAttachmentStateAvailable,
		},
		"modifying": {
			state: ec2types.TransitGatewayAttachmentStateModifying,
		},
		"rolling back": {
			state:        ec2types.TransitGatewayAttachmentStateRollingBack,
			wantInactive: true,
		},
		"deleting": {
			state:        ec2types.TransitGatewayAttachmentStateDeleting,
			wantInactive: true,
		},
		"deleted": {
			state:        ec2types.TransitGatewayAttachmentStateDeleted,
			wantInactive: true,
		},
		"failing": {
			state:        ec2types.TransitGatewayAttachmentStateFailing,
			wantInactive: true,
		},
		"failed": {
			state:        ec2types.TransitGatewayAttachmentStateFailed,
			wantInactive: true,
		},
		"rejecting": {
			state:        ec2types.TransitGatewayAttachmentStateRejecting,
			wantInactive: true,
		},
		"rejected": {
			state:        ec2types.TransitGatewayAttachmentStateRejected,
			wantInactive: true,
		},
		"unknown": {
			state: ec2types.TransitGatewayAttachmentState("unknown"),
		},
		"empty": {},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := transitGatewayAttachmentStateIsInactive(tc.state); got != tc.wantInactive {
				t.Errorf("transitGatewayAttachmentStateIsInactive(%q) = %t, want %t", tc.state, got, tc.wantInactive)
			}
		})
	}
}

func TestFindActiveTransitGatewayAttachment(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attachments []ec2types.TransitGatewayAttachment
		wantID      string
		wantError   error
	}{
		"selects transitional attachment despite stale teardown attachment": {
			attachments: []ec2types.TransitGatewayAttachment{
				{
					State:                      ec2types.TransitGatewayAttachmentStateDeleting,
					TransitGatewayAttachmentId: aws.String("tgw-attach-stale"),
				},
				{
					State:                      ec2types.TransitGatewayAttachmentStatePending,
					TransitGatewayAttachmentId: aws.String("tgw-attach-active"),
				},
			},
			wantID: "tgw-attach-active",
		},
		"selects unknown state": {
			attachments: []ec2types.TransitGatewayAttachment{
				{
					State:                      ec2types.TransitGatewayAttachmentState("unknown"),
					TransitGatewayAttachmentId: aws.String("tgw-attach-unknown"),
				},
			},
			wantID: "tgw-attach-unknown",
		},
		"empty": {
			wantError: tfresource.ErrEmptyResult,
		},
		"only inactive attachments": {
			attachments: []ec2types.TransitGatewayAttachment{
				{
					State:                      ec2types.TransitGatewayAttachmentStateDeleted,
					TransitGatewayAttachmentId: aws.String("tgw-attach-deleted"),
				},
				{
					State:                      ec2types.TransitGatewayAttachmentStateFailed,
					TransitGatewayAttachmentId: aws.String("tgw-attach-failed"),
				},
			},
			wantError: tfresource.ErrEmptyResult,
		},
		"multiple active attachments": {
			attachments: []ec2types.TransitGatewayAttachment{
				{
					State:                      ec2types.TransitGatewayAttachmentStateInitiating,
					TransitGatewayAttachmentId: aws.String("tgw-attach-initiating"),
				},
				{
					State:                      ec2types.TransitGatewayAttachmentStateAvailable,
					TransitGatewayAttachmentId: aws.String("tgw-attach-available"),
				},
			},
			wantError: tfresource.ErrTooManyResults,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attachment, err := findActiveTransitGatewayAttachment(tc.attachments)
			if tc.wantError != nil {
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("findActiveTransitGatewayAttachment() error = %v, want %v", err, tc.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("findActiveTransitGatewayAttachment() error = %v", err)
			}
			if got := aws.ToString(attachment.TransitGatewayAttachmentId); got != tc.wantID {
				t.Errorf("attachment ID = %q, want %q", got, tc.wantID)
			}
		})
	}
}
