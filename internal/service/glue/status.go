// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	schemaVersionStatusUnknown = "Unknown"
	triggerStatusUnknown       = "Unknown"
)

// statusSchemaVersion fetches the Schema Version and its Status
func statusSchemaVersion(ctx context.Context, conn *glue.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := FindSchemaVersionByID(ctx, conn, id)
		if err != nil {
			return nil, schemaVersionStatusUnknown, err
		}

		if output == nil {
			return output, schemaVersionStatusUnknown, nil
		}

		return output, string(output.Status), nil
	}
}

// statusTrigger fetches the Trigger and its Status
func statusTrigger(ctx context.Context, conn *glue.Client, triggerName string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &glue.GetTriggerInput{
			Name: aws.String(triggerName),
		}

		output, err := conn.GetTrigger(ctx, input)

		if err != nil {
			return nil, triggerStatusUnknown, err
		}

		if output == nil {
			return output, triggerStatusUnknown, nil
		}

		return output, string(output.Trigger.State), nil
	}
}
