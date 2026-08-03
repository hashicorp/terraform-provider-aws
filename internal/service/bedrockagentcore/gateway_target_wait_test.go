// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

func TestGatewayTargetWaiterStates(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		states      func() (pending, target []string)
		wantPending []string
		wantTarget  []string
	}{
		"create": {
			states:      gatewayTargetCreatedWaiterStates,
			wantPending: enum.Slice(awstypes.TargetStatusCreating),
			wantTarget:  enum.Slice(awstypes.TargetStatusReady, awstypes.TargetStatusCreatePendingAuth),
		},
		"update": {
			states:      gatewayTargetUpdatedWaiterStates,
			wantPending: enum.Slice(awstypes.TargetStatusUpdating),
			wantTarget:  enum.Slice(awstypes.TargetStatusReady, awstypes.TargetStatusUpdatePendingAuth),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotPending, gotTarget := test.states()
			if diff := cmp.Diff(test.wantPending, gotPending); diff != "" {
				t.Errorf("unexpected pending states (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantTarget, gotTarget); diff != "" {
				t.Errorf("unexpected target states (-want, +got):\n%s", diff)
			}
		})
	}
}
