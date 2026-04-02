// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dataexchange_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdataexchange "github.com/hashicorp/terraform-provider-aws/internal/service/dataexchange"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataExchangeDataSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var proj dataexchange.GetDataSetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, "tf-acc-test-updated")
	resourceName := "aws_dataexchange_data_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`data-sets/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSetConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dataexchange", regexache.MustCompile(`data-sets/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeDataSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var proj dataexchange.GetDataSetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_data_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, t, resourceName, &proj),
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
				Config: testAccDataSetConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDataSetConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDataExchangeDataSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var proj dataexchange.GetDataSetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_data_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(ctx, t, resourceName, &proj),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdataexchange.ResourceDataSet(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdataexchange.ResourceDataSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataSetExists(ctx context.Context, t *testing.T, n string, v *dataexchange.GetDataSetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).DataExchangeClient(ctx)
		resp, err := tfdataexchange.FindDataSetByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("DataExchange DataSet not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckDataSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataExchangeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dataexchange_data_set" {
				continue
			}

			// Try to find the resource
			_, err := tfdataexchange.FindDataSetByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataExchange DataSet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}
`, rName)
}

func testAccDataSetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDataSetConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
