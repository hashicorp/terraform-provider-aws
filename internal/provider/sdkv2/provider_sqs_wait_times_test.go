// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccProvider_sqsWaitTimesConfiguration(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_sqsWaitTimes(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_sqs_queue.test", "name", "test-queue"),
				),
			},
		},
	})
}

// Test that the provider configuration is properly parsed and stored
func TestProvider_sqsWaitTimesConfigurationParsing(t *testing.T) {
	ctx := context.Background()

	config := &conns.Config{
		SQSWaitTimes: &conns.SQSWaitTimesConfig{
			CreateContinuousTargetOccurrence: 1,
			DeleteContinuousTargetOccurrence: 2,
		},
	}

	client := &conns.AWSClient{}
	configuredClient, diags := config.ConfigureProvider(ctx, client)

	if diags.HasError() {
		t.Fatalf("Expected no errors, got: %v", diags)
	}

	if configuredClient.SQSWaitTimes == nil {
		t.Fatal("Expected SQSWaitTimes to be set")
	}

	if configuredClient.SQSWaitTimes.CreateContinuousTargetOccurrence != 1 {
		t.Errorf("Expected CreateContinuousTargetOccurrence to be 1, got %d", configuredClient.SQSWaitTimes.CreateContinuousTargetOccurrence)
	}

	if configuredClient.SQSWaitTimes.DeleteContinuousTargetOccurrence != 2 {
		t.Errorf("Expected DeleteContinuousTargetOccurrence to be 2, got %d", configuredClient.SQSWaitTimes.DeleteContinuousTargetOccurrence)
	}
}

func testAccProviderConfig_sqsWaitTimes() string {
	return `
provider "aws" {
  sqs_wait_times {
    create_continuous_target_occurrence = 1
    delete_continuous_target_occurrence = 2
  }
}

resource "aws_sqs_queue" "test" {
  name = "test-queue"
}
`
}
