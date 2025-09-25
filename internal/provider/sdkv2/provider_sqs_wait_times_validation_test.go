// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccProvider_sqsWaitTimesConfigurationValidation(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_sqsWaitTimesMinValues(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_sqs_queue.test", "name", "test-queue"),
				),
			},
		},
	})
}

func TestAccProvider_sqsWaitTimesConfigurationDefaults(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_sqsWaitTimesDefaults(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_sqs_queue.test", "name", "test-queue"),
				),
			},
		},
	})
}

func testAccProviderConfig_sqsWaitTimesMinValues() string {
	return `
provider "aws" {
  sqs_wait_times {
    create_continuous_target_occurrence = 1
    delete_continuous_target_occurrence = 1
  }
}

resource "aws_sqs_queue" "test" {
  name = "test-queue"
}
`
}

func testAccProviderConfig_sqsWaitTimesDefaults() string {
	return `
provider "aws" {
  sqs_wait_times {
    # Using default values
  }
}

resource "aws_sqs_queue" "test" {
  name = "test-queue"
}
`
}
