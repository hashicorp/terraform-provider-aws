package schemas

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDiscovererByID(ctx context.Context, conn *schemas.Schemas, id string) (*schemas.DescribeDiscovererOutput, error) {
	input := &schemas.DescribeDiscovererInput{
		DiscovererId: aws.String(id),
	}

	output, err := conn.DescribeDiscovererWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
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

	return output, nil
}

func FindRegistryByName(ctx context.Context, conn *schemas.Schemas, name string) (*schemas.DescribeRegistryOutput, error) {
	input := &schemas.DescribeRegistryInput{
		RegistryName: aws.String(name),
	}

	output, err := conn.DescribeRegistryWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
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

	return output, nil
}

func FindSchemaByNameAndRegistryName(ctx context.Context, conn *schemas.Schemas, name, registryName string) (*schemas.DescribeSchemaOutput, error) {
	input := &schemas.DescribeSchemaInput{
		RegistryName: aws.String(registryName),
		SchemaName:   aws.String(name),
	}

	output, err := conn.DescribeSchemaWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
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

	return output, nil
}

func FindRegistryPolicyByName(ctx context.Context, conn *schemas.Schemas, name string) (*schemas.GetResourcePolicyOutput, error) {
	input := &schemas.GetResourcePolicyInput{
		RegistryName: aws.String(name),
	}

	output, err := conn.GetResourcePolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
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

	return output, nil
}
