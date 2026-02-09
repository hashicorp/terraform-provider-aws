// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dataexchange_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdataexchange "github.com/hashicorp/terraform-provider-aws/internal/service/dataexchange"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataExchangeRevision_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var proj dataexchange.GetRevisionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRevisionExists(ctx, t, resourceName, &proj),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dataexchange", "data-sets/{data_set_id}/revisions/{revision_id}"),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", "aws_dataexchange_data_set.test", names.AttrID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDataExchangeRevision_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var proj dataexchange.GetRevisionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRevisionExists(ctx, t, resourceName, &proj),
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
				Config: testAccRevisionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRevisionExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRevisionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRevisionExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDataExchangeRevision_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var proj dataexchange.GetRevisionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRevisionExists(ctx, t, resourceName, &proj),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdataexchange.ResourceRevision(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdataexchange.ResourceRevision(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataExchangeRevision_disappears_dataSet(t *testing.T) {
	ctx := acctest.Context(t)
	var proj dataexchange.GetRevisionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.DataExchangeEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataExchangeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRevisionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRevisionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRevisionExists(ctx, t, resourceName, &proj),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdataexchange.ResourceDataSet(), "aws_dataexchange_data_set.test"),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdataexchange.ResourceRevision(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRevisionExists(ctx context.Context, t *testing.T, n string, v *dataexchange.GetRevisionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).DataExchangeClient(ctx)

		dataSetId, revisionId, err := tfdataexchange.RevisionParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := tfdataexchange.FindRevisionByID(ctx, conn, dataSetId, revisionId)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("DataExchange Revision not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckRevisionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataExchangeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dataexchange_revision" {
				continue
			}

			dataSetId, revisionId, err := tfdataexchange.RevisionParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			// Try to find the resource
			_, err = tfdataexchange.FindRevisionByID(ctx, conn, dataSetId, revisionId)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataExchange Revision %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRevisionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}

resource "aws_dataexchange_revision" "test" {
  data_set_id = aws_dataexchange_data_set.test.id
}
`, rName)
}

func testAccRevisionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}

resource "aws_dataexchange_revision" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRevisionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}

resource "aws_dataexchange_revision" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
