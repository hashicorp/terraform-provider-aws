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
		states func() gatewayTargetWaiterStates
		want   gatewayTargetWaiterStates
	}{
		"create": {
			states: gatewayTargetCreatedWaiterStates,
			want: gatewayTargetWaiterStates{
				pending: enum.Slice(awstypes.TargetStatusCreating),
				target:  enum.Slice(awstypes.TargetStatusReady, awstypes.TargetStatusCreatePendingAuth),
			},
		},
		"update": {
			states: gatewayTargetUpdatedWaiterStates,
			want: gatewayTargetWaiterStates{
				pending: enum.Slice(awstypes.TargetStatusUpdating),
				target:  enum.Slice(awstypes.TargetStatusReady, awstypes.TargetStatusUpdatePendingAuth),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(test.want, test.states(), cmp.AllowUnexported(gatewayTargetWaiterStates{})); diff != "" {
				t.Errorf("unexpected waiter states (-want, +got):\n%s", diff)
			}
		})
	}
}
