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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_ingestion.test"

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
				Config: testAccIngestionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					resource.TestCheckResourceAttr(resourceName, "app", "OKTA"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "ingestion_type", "auditLog"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", "test-tenant-id"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_ingestion.test"

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
				Config: testAccIngestionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceIngestion, resourceName),
				),
				ExpectNonEmptyPlan: true,
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

func testAccIngestionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_appfabric_ingestion" "test" {
  app            = "OKTA"
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  tenant_id      = "test-tenant-id"
  ingestion_type = "auditLog"
}
`, rName)
}
