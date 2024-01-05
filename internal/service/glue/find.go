// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDevEndpointByName(ctx context.Context, conn *glue.Glue, name string) (*glue.DevEndpoint, error) {
	input := &glue.GetDevEndpointInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.GetDevEndpointWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DevEndpoint == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DevEndpoint, nil
}

func FindJobByName(ctx context.Context, conn *glue.Glue, name string) (*glue.Job, error) {
	input := &glue.GetJobInput{
		JobName: aws.String(name),
	}

	output, err := conn.GetJobWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Job == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Job, nil
}

func FindDatabaseByName(ctx context.Context, conn *glue.Glue, catalogID, name string) (*glue.GetDatabaseOutput, error) {
	input := &glue.GetDatabaseInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	output, err := conn.GetDatabaseWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindDataQualityRulesetByName(ctx context.Context, conn *glue.Glue, name string) (*glue.GetDataQualityRulesetOutput, error) {
	input := &glue.GetDataQualityRulesetInput{
		Name: aws.String(name),
	}

	output, err := conn.GetDataQualityRulesetWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// FindTriggerByName returns the Trigger corresponding to the specified name.
func FindTriggerByName(ctx context.Context, conn *glue.Glue, name string) (*glue.GetTriggerOutput, error) {
	input := &glue.GetTriggerInput{
		Name: aws.String(name),
	}

	output, err := conn.GetTriggerWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindRegistryByID returns the Registry corresponding to the specified ID.
func FindRegistryByID(ctx context.Context, conn *glue.Glue, id string) (*glue.GetRegistryOutput, error) {
	input := &glue.GetRegistryInput{
		RegistryId: createRegistryID(id),
	}

	output, err := conn.GetRegistryWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindSchemaByID returns the Schema corresponding to the specified ID.
func FindSchemaByID(ctx context.Context, conn *glue.Glue, id string) (*glue.GetSchemaOutput, error) {
	input := &glue.GetSchemaInput{
		SchemaId: createSchemaID(id),
	}

	output, err := conn.GetSchemaWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindSchemaVersionByID returns the latest available Schema corresponding to the specified ID.
func FindSchemaVersionByID(ctx context.Context, conn *glue.Glue, id string) (*glue.GetSchemaVersionOutput, error) {
	input := &glue.GetSchemaVersionInput{
		SchemaId: createSchemaID(id),
		SchemaVersionNumber: &glue.SchemaVersionNumber{
			LatestVersion: aws.Bool(true),
		},
	}

	output, err := conn.GetSchemaVersionWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindSchemaVersionByIDAndVersion returns the Schema corresponding to the specified ID and version number.
func FindSchemaVersionByIDAndVersion(ctx context.Context, conn *glue.Glue, id string, versionNumber *int64) (*glue.GetSchemaVersionOutput, error) {
	input := &glue.GetSchemaVersionInput{
		SchemaId: createSchemaID(id),
		SchemaVersionNumber: &glue.SchemaVersionNumber{
			VersionNumber: versionNumber,
		},
	}

	output, err := conn.GetSchemaVersionWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindPartitionByValues returns the Partition corresponding to the specified Partition Values.
func FindPartitionByValues(ctx context.Context, conn *glue.Glue, id string) (*glue.Partition, error) {
	catalogID, dbName, tableName, values, err := readPartitionID(id)
	if err != nil {
		return nil, err
	}

	input := &glue.GetPartitionInput{
		CatalogId:       aws.String(catalogID),
		DatabaseName:    aws.String(dbName),
		TableName:       aws.String(tableName),
		PartitionValues: aws.StringSlice(values),
	}

	output, err := conn.GetPartitionWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.Partition == nil {
		return nil, nil
	}

	return output.Partition, nil
}

// FindConnectionByName returns the Connection corresponding to the specified Name and CatalogId.
func FindConnectionByName(ctx context.Context, conn *glue.Glue, name, catalogID string) (*glue.Connection, error) {
	input := &glue.GetConnectionInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	output, err := conn.GetConnectionWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Connection == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Connection, nil
}

// FindPartitionIndexByName returns the Partition Index corresponding to the specified Partition Index Name.
func FindPartitionIndexByName(ctx context.Context, conn *glue.Glue, id string) (*glue.PartitionIndexDescriptor, error) {
	catalogID, dbName, tableName, partIndex, err := readPartitionIndexID(id)
	if err != nil {
		return nil, err
	}

	input := &glue.GetPartitionIndexesInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
	}

	var result *glue.PartitionIndexDescriptor

	output, err := conn.GetPartitionIndexesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, partInd := range output.PartitionIndexDescriptorList {
		if partInd == nil {
			continue
		}

		if aws.StringValue(partInd.IndexName) == partIndex {
			result = partInd
			break
		}
	}

	if result == nil {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	return result, nil
}

func FindClassifierByName(ctx context.Context, conn *glue.Glue, name string) (*glue.Classifier, error) {
	input := &glue.GetClassifierInput{
		Name: aws.String(name),
	}

	output, err := conn.GetClassifierWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Classifier == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Classifier, nil
}

func FindCrawlerByName(ctx context.Context, conn *glue.Glue, name string) (*glue.Crawler, error) {
	input := &glue.GetCrawlerInput{
		Name: aws.String(name),
	}

	output, err := conn.GetCrawlerWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Crawler == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Crawler, nil
}
