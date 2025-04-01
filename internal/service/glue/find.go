// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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

// FindRegistryByName returns the Registry corresponding to the specified name.
func FindRegistryByName(ctx context.Context, conn *glue.Client, name string) (*glue.GetRegistryOutput, error) {
	input := &glue.GetRegistryInput{
		RegistryId: &awstypes.RegistryId{
			RegistryName: aws.String(name),
		},
	}

	output, err := conn.GetRegistry(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

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
