package glue

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDevEndpointByName(conn *glue.Glue, name string) (*glue.DevEndpoint, error) {
	input := &glue.GetDevEndpointInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.GetDevEndpoint(input)

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindJobByName(conn *glue.Glue, name string) (*glue.Job, error) {
	input := &glue.GetJobInput{
		JobName: aws.String(name),
	}

	output, err := conn.GetJob(input)

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &resource.NotFoundError{
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

// FindTableByName returns the Table corresponding to the specified name.
func FindTableByName(conn *glue.Glue, catalogID, dbName, name string) (*glue.GetTableOutput, error) {
	input := &glue.GetTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		Name:         aws.String(name),
	}

	output, err := conn.GetTable(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindTriggerByName returns the Trigger corresponding to the specified name.
func FindTriggerByName(conn *glue.Glue, name string) (*glue.GetTriggerOutput, error) {
	input := &glue.GetTriggerInput{
		Name: aws.String(name),
	}

	output, err := conn.GetTrigger(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindRegistryByID returns the Registry corresponding to the specified ID.
func FindRegistryByID(conn *glue.Glue, id string) (*glue.GetRegistryOutput, error) {
	input := &glue.GetRegistryInput{
		RegistryId: createRegistryID(id),
	}

	output, err := conn.GetRegistry(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindSchemaByID returns the Schema corresponding to the specified ID.
func FindSchemaByID(conn *glue.Glue, id string) (*glue.GetSchemaOutput, error) {
	input := &glue.GetSchemaInput{
		SchemaId: createSchemaID(id),
	}

	output, err := conn.GetSchema(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindSchemaVersionByID returns the Schema corresponding to the specified ID.
func FindSchemaVersionByID(conn *glue.Glue, id string) (*glue.GetSchemaVersionOutput, error) {
	input := &glue.GetSchemaVersionInput{
		SchemaId: createSchemaID(id),
		SchemaVersionNumber: &glue.SchemaVersionNumber{
			LatestVersion: aws.Bool(true),
		},
	}

	output, err := conn.GetSchemaVersion(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindPartitionByValues returns the Partition corresponding to the specified Partition Values.
func FindPartitionByValues(conn *glue.Glue, id string) (*glue.Partition, error) {
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

	output, err := conn.GetPartition(input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.Partition == nil {
		return nil, nil
	}

	return output.Partition, nil
}

// FindConnectionByName returns the Connection corresponding to the specified Name and CatalogId.
func FindConnectionByName(conn *glue.Glue, name, catalogID string) (*glue.Connection, error) {
	input := &glue.GetConnectionInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	output, err := conn.GetConnection(input)
	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &resource.NotFoundError{
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
func FindPartitionIndexByName(conn *glue.Glue, id string) (*glue.PartitionIndexDescriptor, error) {
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

	output, err := conn.GetPartitionIndexes(input)

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	return result, nil
}

func FindClassifierByName(conn *glue.Glue, name string) (*glue.Classifier, error) {
	input := &glue.GetClassifierInput{
		Name: aws.String(name),
	}

	output, err := conn.GetClassifier(input)
	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return nil, &resource.NotFoundError{
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
