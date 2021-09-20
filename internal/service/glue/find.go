package glue

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.DevEndpoint, nil
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
		RegistryId: createAwsRegistryID(id),
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
		SchemaId: createAwsSchemaID(id),
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
		SchemaId: createAwsSchemaID(id),
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

	catalogID, dbName, tableName, values, err := readAwsPartitionID(id)
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

	if output == nil && output.Partition == nil {
		return nil, nil
	}

	return output.Partition, nil
}
