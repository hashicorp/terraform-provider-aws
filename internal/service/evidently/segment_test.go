// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/evidently/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchevidently "github.com/hashicorp/terraform-provider-aws/internal/service/evidently"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEvidentlySegment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var segment awstypes.Segment

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_evidently_segment.test"
	pattern := "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSegmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentConfig_basic(rName, pattern),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSegmentExists(ctx, resourceName, &segment),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "evidently", fmt.Sprintf("segment/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrID, "evidently", fmt.Sprintf("segment/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedTime),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pattern", pattern),
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

func TestAccEvidentlySegment_description(t *testing.T) {
	ctx := acctest.Context(t)
	var segment awstypes.Segment

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"
	resourceName := "aws_evidently_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSegmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSegmentExists(ctx, resourceName, &segment),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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

func TestAccEvidentlySegment_patternJSON(t *testing.T) {
	ctx := acctest.Context(t)
	var segment awstypes.Segment

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_evidently_segment.test"
	pattern := "  {\n\t  \"Price\": [\n\t\t  {\n\t\t\t  \"numeric\": [\">\",10,\"<=\",20]\n\t\t  }\n\t  ]\n  }\n"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSegmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentConfig_patternJSON(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSegmentExists(ctx, resourceName, &segment),
					resource.TestCheckResourceAttr(resourceName, "pattern", pattern),
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

func TestAccEvidentlySegment_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var segment awstypes.Segment

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_evidently_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSegmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSegmentExists(ctx, resourceName, &segment),
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
				Config: testAccSegmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSegmentExists(ctx, resourceName, &segment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSegmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSegmentExists(ctx, resourceName, &segment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEvidentlySegment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var segment awstypes.Segment

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	pattern := "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"
	resourceName := "aws_evidently_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSegmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSegmentConfig_basic(rName, pattern),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSegmentExists(ctx, resourceName, &segment),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatchevidently.ResourceSegment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSegmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_evidently_segment" {
				continue
			}

			_, err := tfcloudwatchevidently.FindSegmentByNameOrARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Evidently Segment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSegmentExists(ctx context.Context, n string, v *awstypes.Segment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Evidently Segment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyClient(ctx)

		output, err := tfcloudwatchevidently.FindSegmentByNameOrARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSegmentConfig_basic(rName, pattern string) string {
	return fmt.Sprintf(`
resource "aws_evidently_segment" "test" {
  name    = %[1]q
  pattern = %[2]q
}
`, rName, pattern)
}

func testAccSegmentConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_evidently_segment" "test" {
  name        = %[1]q
  pattern     = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"
  description = %[2]q
}
`, rName, description)
}

func testAccSegmentConfig_patternJSON(rName string) string {
	return fmt.Sprintf(`
resource "aws_evidently_segment" "test" {
  name    = %[1]q
  pattern = <<JSON
  {
	  "Price": [
		  {
			  "numeric": [">",10,"<=",20]
		  }
	  ]
  }
  JSON
}
`, rName)
}

func testAccSegmentConfig_tags1(rName, tag, value string) string {
	return fmt.Sprintf(`
resource "aws_evidently_segment" "test" {
  name    = %[1]q
  pattern = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag, value)
}

func testAccSegmentConfig_tags2(rName, tag1, value1, tag2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_evidently_segment" "test" {
  name    = %[1]q
  pattern = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1, value1, tag2, value2)
}
