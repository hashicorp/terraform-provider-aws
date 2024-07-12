// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSQSQueueDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_sqs_queue.test"
	datasourceName := "data.aws_sqs_queue.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccQueueCheckDataSource(datasourceName, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(datasourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccQueueCheckDataSource(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		sqsQueueRs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			names.AttrARN,
			names.AttrName,
		}

		for _, attrName := range attrNames {
			if rs.Primary.Attributes[attrName] != sqsQueueRs.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					rs.Primary.Attributes[attrName],
					sqsQueueRs.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

func testAccQueueDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "wrong" {
  name = "%[1]s_wrong"
}

resource "aws_sqs_queue" "test" {
  name = "%[1]s"
}

data "aws_sqs_queue" "by_name" {
  name = aws_sqs_queue.test.name
}
`, rName)
}
