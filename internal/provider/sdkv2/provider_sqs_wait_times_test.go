// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestProvider_sqsWaitTimesConfigurationParsing(t *testing.T) {
	config := &conns.SQSWaitTimesConfig{
		CreateContinuousTargetOccurrence: 1,
		DeleteContinuousTargetOccurrence: 2,
	}

	if config.CreateContinuousTargetOccurrence != 1 {
		t.Errorf("Expected CreateContinuousTargetOccurrence to be 1, got %d", config.CreateContinuousTargetOccurrence)
	}

	if config.DeleteContinuousTargetOccurrence != 2 {
		t.Errorf("Expected DeleteContinuousTargetOccurrence to be 2, got %d", config.DeleteContinuousTargetOccurrence)
	}

	client := &conns.AWSClient{
		sqsWaitTimes: config,
	}

	if client.SQSWaitTimes() == nil {
		t.Fatal("Expected SQSWaitTimes to be set")
	}

	if client.SQSWaitTimes().CreateContinuousTargetOccurrence != 1 {
		t.Errorf("Expected CreateContinuousTargetOccurrence to be 1, got %d", client.SQSWaitTimes().CreateContinuousTargetOccurrence)
	}

	if client.SQSWaitTimes().DeleteContinuousTargetOccurrence != 2 {
		t.Errorf("Expected DeleteContinuousTargetOccurrence to be 2, got %d", client.SQSWaitTimes().DeleteContinuousTargetOccurrence)
	}
}
