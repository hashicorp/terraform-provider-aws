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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDevEndpointByName(ctx context.Context, conn *glue.Client, name string) (*awstypes.DevEndpoint, error) {
	input := &glue.GetDevEndpointInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.GetDevEndpoint(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
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

func FindJobByName(ctx context.Context, conn *glue.Client, name string) (*awstypes.Job, error) {
	input := &glue.GetJobInput{
		JobName: aws.String(name),
	}

	output, err := conn.GetJob(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
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

func FindDatabaseByName(ctx context.Context, conn *glue.Client, catalogID, name string) (*glue.GetDatabaseOutput, error) {
	input := &glue.GetDatabaseInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	output, err := conn.GetDatabase(ctx, input)
	if errs.IsA[*awstypes.EntityNotFoundException](err) {
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

func FindDataQualityRulesetByName(ctx context.Context, conn *glue.Client, name string) (*glue.GetDataQualityRulesetOutput, error) {
	input := &glue.GetDataQualityRulesetInput{
		Name: aws.String(name),
	}

	output, err := conn.GetDataQualityRuleset(ctx, input)
	if errs.IsA[*awstypes.EntityNotFoundException](err) {
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

// FindRegistryByID returns the Registry corresponding to the specified ID.
func FindRegistryByID(ctx context.Context, conn *glue.Client, id string) (*glue.GetRegistryOutput, error) {
	input := &glue.GetRegistryInput{
		RegistryId: createRegistryID(id),
	}

	output, err := conn.GetRegistry(ctx, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindSchemaByID returns the Schema corresponding to the specified ID.
func FindSchemaByID(ctx context.Context, conn *glue.Client, id string) (*glue.GetSchemaOutput, error) {
	input := &glue.GetSchemaInput{
		SchemaId: createSchemaID(id),
	}

	output, err := conn.GetSchema(ctx, input)
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

// FindPartitionByValues returns the Partition corresponding to the specified Partition Values.
func FindPartitionByValues(ctx context.Context, conn *glue.Client, id string) (*awstypes.Partition, error) {
	catalogID, dbName, tableName, values, err := readPartitionID(id)
	if err != nil {
		return nil, err
	}

	input := &glue.GetPartitionInput{
		CatalogId:       aws.String(catalogID),
		DatabaseName:    aws.String(dbName),
		TableName:       aws.String(tableName),
		PartitionValues: values,
	}

	output, err := conn.GetPartition(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.Partition == nil {
		return nil, nil
	}

	return output.Partition, nil
}

// FindConnectionByName returns the Connection corresponding to the specified Name and CatalogId.
func FindConnectionByName(ctx context.Context, conn *glue.Client, name, catalogID string) (*awstypes.Connection, error) {
	input := &glue.GetConnectionInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	output, err := conn.GetConnection(ctx, input)
	if errs.IsA[*awstypes.EntityNotFoundException](err) {
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
func FindPartitionIndexByName(ctx context.Context, conn *glue.Client, id string) (*awstypes.PartitionIndexDescriptor, error) {
	catalogID, dbName, tableName, partIndex, err := readPartitionIndexID(id)
	if err != nil {
		return nil, err
	}

	input := &glue.GetPartitionIndexesInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
	}

	var result *awstypes.PartitionIndexDescriptor

	output, err := conn.GetPartitionIndexes(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
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
		index := partInd
		if aws.ToString(partInd.IndexName) == partIndex {
			result = &index
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

func FindClassifierByName(ctx context.Context, conn *glue.Client, name string) (*awstypes.Classifier, error) {
	input := &glue.GetClassifierInput{
		Name: aws.String(name),
	}

	output, err := conn.GetClassifier(ctx, input)
	if errs.IsA[*awstypes.EntityNotFoundException](err) {
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

func FindCrawlerByName(ctx context.Context, conn *glue.Client, name string) (*awstypes.Crawler, error) {
	input := &glue.GetCrawlerInput{
		Name: aws.String(name),
	}

	output, err := conn.GetCrawler(ctx, input)
	if errs.IsA[*awstypes.EntityNotFoundException](err) {
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
