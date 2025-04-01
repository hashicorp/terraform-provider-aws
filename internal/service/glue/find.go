// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
)

// FindTriggerByName returns the Trigger corresponding to the specified name.
func FindTriggerByName(ctx context.Context, conn *glue.Client, name string) (*glue.GetTriggerOutput, error) {
	input := &glue.GetTriggerInput{
		Name: aws.String(name),
	}

	output, err := conn.GetTrigger(ctx, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindSchemaVersionByID returns the Schema corresponding to the specified ID.
func FindSchemaVersionByID(ctx context.Context, conn *glue.Client, id string) (*glue.GetSchemaVersionOutput, error) {
	input := &glue.GetSchemaVersionInput{
		SchemaId: createSchemaID(id),
		SchemaVersionNumber: &awstypes.SchemaVersionNumber{
			LatestVersion: true,
		},
	}

	output, err := conn.GetSchemaVersion(ctx, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}
