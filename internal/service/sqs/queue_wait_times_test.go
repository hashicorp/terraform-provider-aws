// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSQSQueue_waitTimesPerformance(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}

	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_waitTimesPerformance(rName, 1, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccSQSQueue_waitTimesConfigurationEdgeCases(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SQSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_waitTimesEdgeCases(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestSQSQueue_waitTimesConfigurationIntegration(t *testing.T) {
	ctx := context.Background()

	// Test with custom wait times
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

func TestSQSQueue_waitTimesConfigurationDefaults(t *testing.T) {
	ctx := context.Background()

	config := &conns.Config{
		SQSWaitTimes: nil,
	}

	client := &conns.AWSClient{}
	configuredClient, diags := config.ConfigureProvider(ctx, client)

	if diags.HasError() {
		t.Fatalf("Expected no errors, got: %v", diags)
	}

	if configuredClient.SQSWaitTimes != nil {
		t.Fatal("Expected SQSWaitTimes to be nil when not configured")
	}
}

func testAccQueueConfig_waitTimesPerformance(rName string, createOccurrence, deleteOccurrence int) string {
	return fmt.Sprintf(`
provider "aws" {
  sqs_wait_times {
    create_continuous_target_occurrence = %[2]d
    delete_continuous_target_occurrence = %[3]d
  }
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}
`, rName, createOccurrence, deleteOccurrence)
}

func testAccQueueConfig_waitTimesEdgeCases(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
  sqs_wait_times {
    create_continuous_target_occurrence = 1
    delete_continuous_target_occurrence = 1
  }
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}
`, rName)
}
