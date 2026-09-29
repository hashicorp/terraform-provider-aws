// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Account data retention is an account- and Region-level singleton with no
// delete API, so these tests mutate a setting shared by everything in the
// account and leave the last applied mode in place on destroy.

func testAccAccountDataRetention_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_bedrock_account_data_retention.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataRetentionConfig_basic(string(awstypes.DataRetentionModeDefault)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountDataRetentionMode(ctx, awstypes.DataRetentionModeDefault),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), knownvalue.StringExact(string(awstypes.DataRetentionModeDefault))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("updated_at"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccAccountDataRetention_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_bedrock_account_data_retention.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataRetentionConfig_basic(string(awstypes.DataRetentionModeNone)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountDataRetentionMode(ctx, awstypes.DataRetentionModeNone),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), knownvalue.StringExact(string(awstypes.DataRetentionModeNone))),
				},
			},
			{
				Config: testAccAccountDataRetentionConfig_basic(string(awstypes.DataRetentionModeDefault)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountDataRetentionMode(ctx, awstypes.DataRetentionModeDefault),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrMode), knownvalue.StringExact(string(awstypes.DataRetentionModeDefault))),
				},
			},
		},
	})
}

func testAccCheckAccountDataRetentionExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[name]; !ok {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameAccountDataRetention, name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		if _, err := conn.GetAccountDataRetention(ctx, &bedrock.GetAccountDataRetentionInput{}); err != nil {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameAccountDataRetention, name, err)
		}

		return nil
	}
}

// testAccCheckAccountDataRetentionMode asserts the mode AWS actually reports,
// rather than only what is recorded in state.
func testAccCheckAccountDataRetentionMode(ctx context.Context, want awstypes.DataRetentionMode) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		out, err := conn.GetAccountDataRetention(ctx, &bedrock.GetAccountDataRetentionInput{})
		if err != nil {
			return err
		}

		if got := out.Mode; got != want {
			return fmt.Errorf("Bedrock Account Data Retention mode = %s, want %s", got, want)
		}

		return nil
	}
}

func testAccAccountDataRetentionConfig_basic(mode string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_account_data_retention" "test" {
  mode = %[1]q
}
`, mode)
}
