// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfshield "github.com/hashicorp/terraform-provider-aws/internal/service/shield"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccShieldSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t) //nolint:staticcheck // will be used when hardcoded skip is commented
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Due to the high cost of this subscription, we hardcode this test to
	// skip rather than gating behind an environment variable.
	// Run this test be removing the line below.
	t.Skipf("running this test requires a yearly commitment to AWS Shield Advanced with a $3000 monthly fee in the associated account")

	var subscription shield.DescribeSubscriptionOutput
	resourceName := "aws_shield_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionConfig_basic(string(awstypes.AutoRenewEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "auto_renew", string(awstypes.AutoRenewEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckSubscriptionExists(ctx context.Context, name string, subscription *shield.DescribeSubscriptionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameSubscription, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameSubscription, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)
		resp, err := conn.DescribeSubscription(ctx, &shield.DescribeSubscriptionInput{})

		if err != nil {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameSubscription, rs.Primary.ID, err)
		}

		*subscription = *resp

		return nil
	}
}

func testAccSubscriptionConfig_basic(autoRenew string) string {
	return fmt.Sprintf(`
resource "aws_shield_subscription" "test" {
  auto_renew   = %[1]q
  skip_destroy = true
}
`, autoRenew)
}
