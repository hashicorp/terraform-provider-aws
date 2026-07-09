// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// TestHarnessMemoryConfigurationRoundTrip proves the memory configuration union
// round-trips SDK -> model -> SDK for the configurable members and that the
// default managed member (returned by the service when no memory block is
// configured) expands back to no memory configuration. autoflex passes the
// dereferenced union member value to Flatten, which is what the fixtures below
// feed in.
func TestHarnessMemoryConfigurationRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	agentCore := &awstypes.HarnessMemoryConfigurationMemberAgentCoreMemoryConfiguration{
		Value: awstypes.HarnessAgentCoreMemoryConfiguration{
			Arn:           aws.String("arn:aws:bedrock-agentcore:us-west-2:123456789012:memory/mem-abc123"),
			ActorId:       aws.String("actor-1"),
			MessagesCount: aws.Int32(10),
			RetrievalConfig: map[string]awstypes.HarnessAgentCoreMemoryRetrievalConfig{
				"strategy-key": {
					RelevanceScore: aws.Float32(0.5),
					StrategyId:     aws.String("strategy-1"),
					TopK:           aws.Int32(3),
				},
			},
		},
	}

	testCases := map[string]struct {
		flattenInput   any                                 // union member value, as autoflex passes it
		expectedExpand awstypes.HarnessMemoryConfiguration // union member pointer, as Expand returns it
	}{
		"agentcore": {
			flattenInput:   *agentCore,
			expectedExpand: agentCore,
		},
		"disabled": {
			flattenInput:   awstypes.HarnessMemoryConfigurationMemberDisabled{},
			expectedExpand: &awstypes.HarnessMemoryConfigurationMemberDisabled{},
		},
		// The service assigns a managed memory configuration by default; it is not
		// a configurable block, so it expands back to no memory configuration.
		"managed_default": {
			flattenInput:   awstypes.HarnessMemoryConfigurationMemberManagedMemoryConfiguration{Value: awstypes.HarnessManagedMemoryConfiguration{}},
			expectedExpand: nil,
		},
	}

	opts := cmp.Options{
		cmpopts.IgnoreUnexported(
			awstypes.HarnessMemoryConfigurationMemberAgentCoreMemoryConfiguration{},
			awstypes.HarnessMemoryConfigurationMemberDisabled{},
			awstypes.HarnessAgentCoreMemoryConfiguration{},
			awstypes.HarnessAgentCoreMemoryRetrievalConfig{},
			awstypes.HarnessDisabledMemoryConfiguration{},
		),
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var model harnessMemoryConfigurationModel
			if diags := model.Flatten(ctx, tc.flattenInput); diags.HasError() {
				t.Fatalf("Flatten: %v", diags)
			}

			out, diags := model.expandToHarnessMemoryConfiguration(ctx)
			if diags.HasError() {
				t.Fatalf("Expand: %v", diags)
			}

			if diff := cmp.Diff(tc.expectedExpand, out, opts...); diff != "" {
				t.Errorf("SDK -> model -> SDK round-trip mismatch (-in +out):\n%s", diff)
			}
		})
	}
}

// TestCollapseDefaultManagedMemory verifies that the server-assigned managed
// default is collapsed to a null memory configuration (so an omitted memory block
// does not produce a perpetual diff), while an explicit disabled configuration is
// preserved.
func TestCollapseDefaultManagedMemory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("managed default collapses to null", func(t *testing.T) {
		t.Parallel()

		// Mirror what autoflex produces for the managed default: a populated
		// (all-null) memory element. Collapse must reset it to null.
		data := &harnessResourceModel{
			Memory: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, newTestMemoryModel(ctx, false)),
		}
		harness := &awstypes.Harness{
			Memory: &awstypes.HarnessMemoryConfigurationMemberManagedMemoryConfiguration{},
		}

		collapseDefaultManagedMemory(ctx, harness, data)

		if !data.Memory.IsNull() {
			t.Fatalf("expected memory to be null for the default managed configuration")
		}
	})

	t.Run("explicit disabled is preserved", func(t *testing.T) {
		t.Parallel()

		data := &harnessResourceModel{
			Memory: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, newTestMemoryModel(ctx, true)),
		}
		harness := &awstypes.Harness{
			Memory: &awstypes.HarnessMemoryConfigurationMemberDisabled{},
		}

		collapseDefaultManagedMemory(ctx, harness, data)

		if data.Memory.IsNull() {
			t.Fatalf("expected an explicit disabled memory configuration to be preserved")
		}
	})
}

// newTestMemoryModel builds a memory configuration model with its framework-typed
// fields initialized as proper nulls (the framework/autoflex does this in the real
// Create/Read path; a bare struct literal would leave them as invalid zero values).
func newTestMemoryModel(ctx context.Context, disabled bool) *harnessMemoryConfigurationModel {
	m := &harnessMemoryConfigurationModel{
		AgentCoreMemoryConfiguration: fwtypes.NewListNestedObjectValueOfNull[harnessAgentCoreMemoryConfigurationModel](ctx),
		Disabled:                     types.BoolNull(),
	}
	if disabled {
		m.Disabled = types.BoolValue(true)
	}
	return m
}
