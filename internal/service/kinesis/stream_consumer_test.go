// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkinesis "github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisStreamConsumer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_stream_consumer.test"
	streamName := "aws_kinesis_stream.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamConsumerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kinesis", regexache.MustCompile(fmt.Sprintf("stream/%[1]s/consumer/%[1]s", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStreamARN, streamName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "creation_timestamp"),
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

func TestAccKinesisStreamConsumer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkinesis.ResourceStreamConsumer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKinesisStreamConsumer_maxConcurrentConsumers(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamConsumerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Test creation of max number (5 according to AWS API docs) of concurrent consumers for a single stream
				Config: testAccStreamConsumerConfig_multiple(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.0", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.1", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.2", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.3", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.4", resourceName)),
				),
			},
		},
	})
}

func TestAccKinesisStreamConsumer_exceedMaxConcurrentConsumers(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamConsumerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Test creation of more than the max number (5 according to AWS API docs) of concurrent consumers for a single stream
				Config: testAccStreamConsumerConfig_multiple(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.0", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.1", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.2", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.3", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.4", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.5", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.6", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.7", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.8", resourceName)),
					testAccStreamConsumerExists(ctx, t, fmt.Sprintf("%s.9", resourceName)),
				),
			},
		},
	})
}

func TestAccKinesisStreamConsumer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamConsumerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, t, resourceName),
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
				Config: testAccStreamConsumerConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStreamConsumerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckStreamConsumerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KinesisClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_stream_consumer" {
				continue
			}

			_, err := tfkinesis.FindStreamConsumerByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Kinesis Stream Consumer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStreamConsumerExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KinesisClient(ctx)

		_, err := tfkinesis.FindStreamConsumerByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccStreamConsumerConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2
}
`, rName)
}

func testAccStreamConsumerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStreamConsumerConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test" {
  name       = %[1]q
  stream_arn = aws_kinesis_stream.test.arn
}
`, rName))
}

func testAccStreamConsumerConfig_multiple(rName string, count int) string {
	return acctest.ConfigCompose(testAccStreamConsumerConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test" {
  count      = %[2]d
  name       = "%[1]s-${count.index}"
  stream_arn = aws_kinesis_stream.test.arn
}
`, rName, count))
}

func testAccStreamConsumerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccStreamConsumerConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test" {
  name       = %[1]q
  stream_arn = aws_kinesis_stream.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccStreamConsumerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccStreamConsumerConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test" {
  name       = %[1]q
  stream_arn = aws_kinesis_stream.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
