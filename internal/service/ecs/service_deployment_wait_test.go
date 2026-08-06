// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func TestServiceDeploymentWaitRefreshStatus(t *testing.T) {
	t.Parallel()

	t.Run("successful and in-progress statuses pass through", func(t *testing.T) {
		t.Parallel()
		for _, status := range []string{
			string(awstypes.ServiceDeploymentStatusSuccessful),
			string(awstypes.ServiceDeploymentStatusRollbackSuccessful),
			string(awstypes.ServiceDeploymentStatusInProgress),
			string(awstypes.ServiceDeploymentStatusPending),
			string(awstypes.ServiceDeploymentStatusRollbackRequested),
			string(awstypes.ServiceDeploymentStatusRollbackInProgress),
		} {
			got, err := serviceDeploymentWaitRefreshStatus(status, nil)
			if err != nil {
				t.Fatalf("status %s: unexpected err: %v", status, err)
			}
			if got != status {
				t.Fatalf("status %s: got %q", status, got)
			}
		}
	})

	t.Run("failed terminals return errors", func(t *testing.T) {
		t.Parallel()
		for _, status := range []string{
			string(awstypes.ServiceDeploymentStatusStopped),
			string(awstypes.ServiceDeploymentStatusRollbackFailed),
		} {
			got, err := serviceDeploymentWaitRefreshStatus(status, aws.String("boom"))
			if err == nil {
				t.Fatalf("status %s: expected error", status)
			}
			if err.Error() != "boom" {
				t.Fatalf("status %s: got err %v", status, err)
			}
			if got != status {
				t.Fatalf("status %s: got %q", status, got)
			}
		}
	})
}
