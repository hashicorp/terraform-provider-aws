// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccIngestion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ingestion awstypes.Ingestion
	resourceName := "aws_appfabric_ingestion.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rName, tenantID, serviceAccountToken),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ingestion_type", "auditLog"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rName, tenantID, serviceAccountToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceIngestion, resourceName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_tags1(rName, tenantID, serviceAccountToken, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccIngestionConfig_tags1(rName, tenantID, serviceAccountToken, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckIngestionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_ingestion" {
				continue
			}

			_, err := tfappfabric.FindIngestionByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_bundle_arn"], rs.Primary.Attributes[names.AttrARN])

			if tfresource.NotFound(err) {
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

func testAccCheckIngestionExists(ctx context.Context, n string, v *awstypes.Ingestion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

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
