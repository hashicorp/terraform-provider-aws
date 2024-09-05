// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsArchive_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeArchiveOutput
	archiveName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_archive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_basic(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, archiveName),
					resource.TestCheckResourceAttr(resourceName, "retention_days", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "events", fmt.Sprintf("archive/%s", archiveName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", ""),
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

func TestAccEventsArchive_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeArchiveOutput
	resourceName := "aws_cloudwatch_event_archive.test"
	archiveName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_basic(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, resourceName, &v1),
				),
			},
			{
				Config: testAccArchiveConfig_updateAttributes(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "retention_days", "7"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"company.team.service\"]}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
				),
			},
		},
	})
}

func TestAccEventsArchive_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeArchiveOutput
	archiveName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_archive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_basic(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfevents.ResourceArchive(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckArchiveDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_event_archive" {
				continue
			}

			_, err := tfevents.FindArchiveByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Archive %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckArchiveExists(ctx context.Context, n string, v *eventbridge.DescribeArchiveOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

		output, err := tfevents.FindArchiveByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func TestAccEventsArchive_retentionSetOnCreation(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeArchiveOutput
	archiveName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_archive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_retentionOnCreation(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, archiveName),
					resource.TestCheckResourceAttr(resourceName, "retention_days", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "events", fmt.Sprintf("archive/%s", archiveName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", ""),
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

func testAccArchiveConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_archive" "test" {
  name             = %[1]q
  event_source_arn = aws_cloudwatch_event_bus.test.arn
}
`, name)
}

func testAccArchiveConfig_updateAttributes(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_archive" "test" {
  name             = %[1]q
  event_source_arn = aws_cloudwatch_event_bus.test.arn
  retention_days   = 7
  description      = "test"
  event_pattern    = <<PATTERN
{
  "source": ["company.team.service"]
}
PATTERN
}
`, name)
}

func testAccArchiveConfig_retentionOnCreation(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_archive" "test" {
  name             = %[1]q
  event_source_arn = aws_cloudwatch_event_bus.test.arn
  retention_days   = 1
}
`, name)
}
