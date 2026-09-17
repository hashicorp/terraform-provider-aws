// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// TestExpandLogsEncryptionConfigurationEncryptionScope validates that the new
// encryption_scope attribute is carried from the Terraform model into the SDK
// LogsEncryptionConfiguration input via AutoFlex. This runs fully offline (no AWS).
func TestExpandLogsEncryptionConfigurationEncryptionScope(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testCases := map[string]awstypes.EncryptionScope{
		"new_destination_log_groups": awstypes.EncryptionScopeNewDestinationLogGroups,
		"encrypted_source_only":      awstypes.EncryptionScopeEncryptedSourceOnly,
	}

	for name, scope := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := logsEncryptionConfigurationModel{
				EncryptionStrategy: fwtypes.StringEnumValue(awstypes.EncryptionStrategyCustomerManaged),
				EncryptionScope:    fwtypes.StringEnumValue(scope),
			}

			var out awstypes.LogsEncryptionConfiguration
			if diags := fwflex.Expand(ctx, model, &out); diags.HasError() {
				t.Fatalf("unexpected diags expanding model: %v", diags)
			}

			if got, want := out.EncryptionScope, scope; got != want {
				t.Errorf("EncryptionScope = %q, want %q", got, want)
			}
			if got, want := out.EncryptionStrategy, awstypes.EncryptionStrategyCustomerManaged; got != want {
				t.Errorf("EncryptionStrategy = %q, want %q", got, want)
			}
		})
	}
}
