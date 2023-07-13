// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafkaconnect "github.com/hashicorp/terraform-provider-aws/internal/service/kafkaconnect"
)

func TestAccKafkaConnectWorkerConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_worker_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, kafkaconnect.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:             nil,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkerConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "properties_file_content", "key.converter=org.apache.kafka.connect.storage.StringConverter\nvalue.converter=org.apache.kafka.connect.storage.StringConverter\n"),
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

func TestAccKafkaConnectWorkerConfiguration_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_worker_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, kafkaconnect.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:             nil,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerConfigurationConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkerConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "testing"),
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

func testAccCheckWorkerConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Connect Worker Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectConn(ctx)

		_, err := tfkafkaconnect.FindWorkerConfigurationByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccWorkerConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_mskconnect_worker_configuration" "test" {
  name = %[1]q

  properties_file_content = <<EOF
key.converter=org.apache.kafka.connect.storage.StringConverter
value.converter=org.apache.kafka.connect.storage.StringConverter
EOF
}
`, rName)
}

func testAccWorkerConfigurationConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_mskconnect_worker_configuration" "test" {
  name        = %[1]q
  description = "testing"

  properties_file_content = <<EOF
key.converter=org.apache.kafka.connect.storage.StringConverter
value.converter=org.apache.kafka.connect.storage.StringConverter
EOF
}
`, rName)
}
