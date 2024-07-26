// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesis "github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisStreamConsumer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_stream_consumer.test"
	streamName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamConsumerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kinesis", regexache.MustCompile(fmt.Sprintf("stream/%[1]s/consumer/%[1]s", rName))),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkinesis.ResourceStreamConsumer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKinesisStreamConsumer_maxConcurrentConsumers(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamConsumerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Test creation of max number (5 according to AWS API docs) of concurrent consumers for a single stream
				Config: testAccStreamConsumerConfig_multiple(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.0", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.1", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.2", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.3", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.4", resourceName)),
				),
			},
		},
	})
}

func TestAccKinesisStreamConsumer_exceedMaxConcurrentConsumers(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamConsumerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Test creation of more than the max number (5 according to AWS API docs) of concurrent consumers for a single stream
				Config: testAccStreamConsumerConfig_multiple(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.0", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.1", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.2", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.3", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.4", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.5", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.6", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.7", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.8", resourceName)),
					testAccStreamConsumerExists(ctx, fmt.Sprintf("%s.9", resourceName)),
				),
			},
		},
	})
}

func testAccCheckStreamConsumerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_stream_consumer" {
				continue
			}

			_, err := tfkinesis.FindStreamConsumerByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccStreamConsumerExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisClient(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

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
