package glue

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	mlTransformStatusUnknown   = "Unknown"
	registryStatusUnknown      = "Unknown"
	schemaStatusUnknown        = "Unknown"
	schemaVersionStatusUnknown = "Unknown"
	triggerStatusUnknown       = "Unknown"
)

// statusMLTransform fetches the MLTransform and its Status
func statusMLTransform(ctx context.Context, conn *glue.Glue, transformId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &glue.GetMLTransformInput{
			TransformId: aws.String(transformId),
		}

		output, err := conn.GetMLTransformWithContext(ctx, input)

		if err != nil {
			return nil, mlTransformStatusUnknown, err
		}

		if output == nil {
			return output, mlTransformStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// statusRegistry fetches the Registry and its Status
func statusRegistry(ctx context.Context, conn *glue.Glue, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRegistryByID(ctx, conn, id)
		if err != nil {
			return nil, registryStatusUnknown, err
		}

		if output == nil {
			return output, registryStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// statusSchema fetches the Schema and its Status
func statusSchema(ctx context.Context, conn *glue.Glue, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSchemaByID(ctx, conn, id)
		if err != nil {
			return nil, schemaStatusUnknown, err
		}

		if output == nil {
			return output, schemaStatusUnknown, nil
		}

		return output, aws.StringValue(output.SchemaStatus), nil
	}
}

// statusSchemaVersion fetches the Schema Version and its Status
func statusSchemaVersion(ctx context.Context, conn *glue.Glue, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSchemaVersionByID(ctx, conn, id)
		if err != nil {
			return nil, schemaVersionStatusUnknown, err
		}

		if output == nil {
			return output, schemaVersionStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// statusTrigger fetches the Trigger and its Status
func statusTrigger(ctx context.Context, conn *glue.Glue, triggerName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &glue.GetTriggerInput{
			Name: aws.String(triggerName),
		}

		output, err := conn.GetTriggerWithContext(ctx, input)

		if err != nil {
			return nil, triggerStatusUnknown, err
		}

		if output == nil {
			return output, triggerStatusUnknown, nil
		}

		return output, aws.StringValue(output.Trigger.State), nil
	}
}

func statusDevEndpoint(ctx context.Context, conn *glue.Glue, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDevEndpointByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusPartitionIndex(ctx context.Context, conn *glue.Glue, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindPartitionIndexByName(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.IndexStatus), nil
	}
}
