// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccConfigServiceAggregateAuthorization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var aa types.AggregationAuthorization
	accountID := sdkacctest.RandStringFromCharSet(12, "0123456789")
	resourceName := "aws_config_aggregate_authorization.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAggregateAuthorizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAggregateAuthorizationConfig_basic(accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregateAuthorizationExists(ctx, t, resourceName, &aa),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "config", "aggregation-authorization/{account_id}/{authorized_aws_region}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, accountID),
					resource.TestCheckResourceAttr(resourceName, "authorized_aws_region", acctest.Region()),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrRegion),
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

func testAccConfigServiceAggregateAuthorization_deprecatedRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var aa types.AggregationAuthorization
	accountID := sdkacctest.RandStringFromCharSet(12, "0123456789")
	resourceName := "aws_config_aggregate_authorization.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAggregateAuthorizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAggregateAuthorizationConfig_deprecatedRegion(accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregateAuthorizationExists(ctx, t, resourceName, &aa),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "config", "aggregation-authorization/{account_id}/{region}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, accountID),
					resource.TestCheckNoResourceAttr(resourceName, "authorized_aws_region"),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
				),
			},
		},
	})
}

func testAccConfigServiceAggregateAuthorization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var aa types.AggregationAuthorization
	accountID := sdkacctest.RandStringFromCharSet(12, "0123456789")
	resourceName := "aws_config_aggregate_authorization.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAggregateAuthorizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAggregateAuthorizationConfig_basic(accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregateAuthorizationExists(ctx, t, resourceName, &aa),
					acctest.CheckSDKResourceDisappears(ctx, t, tfconfig.ResourceAggregateAuthorization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAggregateAuthorizationExists(ctx context.Context, t *testing.T, n string, v *types.AggregationAuthorization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ConfigServiceClient(ctx)

		authorizedRegion := rs.Primary.Attributes["authorized_aws_region"]
		if authorizedRegion == "" {
			authorizedRegion = rs.Primary.Attributes[names.AttrRegion]
		}
		output, err := tfconfig.FindAggregateAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], authorizedRegion)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAggregateAuthorizationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_aggregate_authorization" {
				continue
			}

			authorizedRegion := rs.Primary.Attributes["authorized_aws_region"]
			if authorizedRegion == "" {
				authorizedRegion = rs.Primary.Attributes[names.AttrRegion]
			}
			_, err := tfconfig.FindAggregateAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], authorizedRegion)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Aggregate Authorization %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAggregateAuthorizationConfig_basic(accountID string) string {
	return fmt.Sprintf(`
resource "aws_config_aggregate_authorization" "test" {
  account_id            = %[1]q
  authorized_aws_region = %[2]q
}
`, accountID, acctest.Region())
}

func testAccAggregateAuthorizationConfig_deprecatedRegion(accountID string) string {
	return fmt.Sprintf(`
resource "aws_config_aggregate_authorization" "test" {
  account_id = %[1]q
  region     = %[2]q
}
`, accountID, acctest.Region())
}
