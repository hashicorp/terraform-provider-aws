// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccIngestion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestion awstypes.Ingestion
	resourceName := "aws_appfabric_ingestion.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rName, tenantID, serviceAccountToken),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngestionExists(ctx, t, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN), // nosemgrep:ci.semgrep.acctest.checks.arn-resourceattrset // TODO: need TFC Org for testing
					resource.TestCheckResourceAttr(resourceName, "ingestion_type", "auditLog"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func testAccIngestion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestion awstypes.Ingestion
	resourceName := "aws_appfabric_ingestion.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rName, tenantID, serviceAccountToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, t, resourceName, &ingestion),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfappfabric.ResourceIngestion, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccIngestion_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestion awstypes.Ingestion
	resourceName := "aws_appfabric_ingestion.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_tags1(rName, tenantID, serviceAccountToken, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, t, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIngestionConfig_tags2(rName, tenantID, serviceAccountToken, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, t, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccIngestionConfig_tags1(rName, tenantID, serviceAccountToken, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, t, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckIngestionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_ingestion" {
				continue
			}

			_, err := tfappfabric.FindIngestionByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_bundle_arn"], rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppFabric Ingestion %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIngestionExists(ctx context.Context, t *testing.T, n string, v *awstypes.Ingestion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppFabricClient(ctx)

		output, err := tfappfabric.FindIngestionByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_bundle_arn"], rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccIngestionConfig_base(rName, tenantID, serviceAccountToken string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_appfabric_app_authorization" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  app            = "TERRAFORMCLOUD"
  auth_type      = "apiKey"

  credential {
    api_key_credential {
      api_key = %[3]q
    }
  }

  tenant {
    tenant_display_name = %[1]q
    tenant_identifier   = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_appfabric_app_authorization_connection" "test" {
  app_bundle_arn        = aws_appfabric_app_bundle.test.arn
  app_authorization_arn = aws_appfabric_app_authorization.test.arn
}
`, rName, tenantID, serviceAccountToken)
}

func testAccIngestionConfig_basic(rName, tenantID, serviceAccountToken string) string {
	return acctest.ConfigCompose(testAccIngestionConfig_base(rName, tenantID, serviceAccountToken), fmt.Sprintf(`
resource "aws_appfabric_ingestion" "test" {
  app            = aws_appfabric_app_authorization_connection.test.app
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  tenant_id      = %[1]q
  ingestion_type = "auditLog"
}
`, tenantID))
}

func testAccIngestionConfig_tags1(rName, tenantID, serviceAccountToken, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccIngestionConfig_base(rName, tenantID, serviceAccountToken), fmt.Sprintf(`
resource "aws_appfabric_ingestion" "test" {
  app            = aws_appfabric_app_authorization_connection.test.app
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  tenant_id      = %[1]q
  ingestion_type = "auditLog"

  tags = {
    %[2]q = %[3]q
  }
}
`, tenantID, tagKey1, tagValue1))
}

func testAccIngestionConfig_tags2(rName, tenantID, serviceAccountToken, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccIngestionConfig_base(rName, tenantID, serviceAccountToken), fmt.Sprintf(`
resource "aws_appfabric_ingestion" "test" {
  app            = aws_appfabric_app_authorization_connection.test.app
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  tenant_id      = %[1]q
  ingestion_type = "auditLog"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, tenantID, tagKey1, tagValue1, tagKey2, tagValue2))
}
