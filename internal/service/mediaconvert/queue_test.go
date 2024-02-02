// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmediaconvert "github.com/hashicorp/terraform-provider-aws/internal/service/mediaconvert"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccMediaConvertQueue_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediaconvert.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mediaconvert", regexache.MustCompile(`queues/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", mediaconvert.PricingPlanOnDemand),
					resource.TestCheckResourceAttr(resourceName, "status", mediaconvert.QueueStatusActive),
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

func TestAccMediaConvertQueue_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediaconvert.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmediaconvert.ResourceQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMediaConvertQueue_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediaconvert.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccQueueConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccMediaConvertQueue_reservationPlanSettings(t *testing.T) {
	acctest.Skip(t, "MediaConvert Reserved Queues are $400/month and cannot be deleted for 1 year.")

	ctx := acctest.Context(t)
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediaconvert.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_reserved(rName, mediaconvert.CommitmentOneYear, mediaconvert.RenewalTypeAutoRenew, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", mediaconvert.PricingPlanReserved),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.commitment", mediaconvert.CommitmentOneYear),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.renewal_type", mediaconvert.RenewalTypeAutoRenew),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.reserved_slots", "1"),
				),
			},
			{
				Config: testAccQueueConfig_reserved(rName, mediaconvert.CommitmentOneYear, mediaconvert.RenewalTypeExpire, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", mediaconvert.PricingPlanReserved),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.commitment", mediaconvert.CommitmentOneYear),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.renewal_type", mediaconvert.RenewalTypeExpire),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.reserved_slots", "1"),
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

func TestAccMediaConvertQueue_withStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediaconvert.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_status(rName, mediaconvert.QueueStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "status", mediaconvert.QueueStatusPaused),
				),
			},
			{
				Config: testAccQueueConfig_status(rName, mediaconvert.QueueStatusActive),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "status", mediaconvert.QueueStatusActive),
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

func TestAccMediaConvertQueue_withDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var queue mediaconvert.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description1 := sdkacctest.RandomWithPrefix("Description: ")
	description2 := sdkacctest.RandomWithPrefix("Description: ")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediaconvert.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_description(rName, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "description", description1),
				),
			},
			{
				Config: testAccQueueConfig_description(rName, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "description", description2),
				),
			},
		},
	})
}

func testAccCheckQueueDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_convert_queue" {
				continue
			}

			conn, err := tfmediaconvert.GetAccountClient(ctx, acctest.Provider.Meta().(*conns.AWSClient))
			if err != nil {
				return fmt.Errorf("Error getting Media Convert Account Client: %s", err)
			}

			_, err = tfmediaconvert.FindQueueByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Media Convert Queue %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckQueueExists(ctx context.Context, n string, v *mediaconvert.Queue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn, err := tfmediaconvert.GetAccountClient(ctx, acctest.Provider.Meta().(*conns.AWSClient))
		if err != nil {
			return fmt.Errorf("Error getting Media Convert Account Client: %s", err)
		}

		output, err := tfmediaconvert.FindQueueByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	_, err := tfmediaconvert.GetAccountClient(ctx, acctest.Provider.Meta().(*conns.AWSClient))

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccQueueConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q
}
`, rName)
}

func testAccQueueConfig_status(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name   = %[1]q
  status = %[2]q
}
`, rName, status)
}

func testAccQueueConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccQueueConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccQueueConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccQueueConfig_reserved(rName, commitment, renewalType string, reservedSlots int) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name         = %[1]q
  pricing_plan = %[2]q

  reservation_plan_settings {
    commitment     = %[3]q
    renewal_type   = %[4]q
    reserved_slots = %[5]d
  }
}
`, rName, mediaconvert.PricingPlanReserved, commitment, renewalType, reservedSlots)
}
