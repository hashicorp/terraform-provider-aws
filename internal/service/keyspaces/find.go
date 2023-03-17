package keyspaces

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/keyspaces"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindKeyspaceByName(ctx context.Context, conn *keyspaces.Keyspaces, name string) (*keyspaces.GetKeyspaceOutput, error) {
	input := keyspaces.GetKeyspaceInput{
		KeyspaceName: aws.String(name),
	}

	output, err := conn.GetKeyspaceWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, keyspaces.ErrCodeResourceNotFoundException) {
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

func FindTableByTwoPartKey(ctx context.Context, conn *keyspaces.Keyspaces, keyspaceName, tableName string) (*keyspaces.GetTableOutput, error) {
	input := keyspaces.GetTableInput{
		KeyspaceName: aws.String(keyspaceName),
		TableName:    aws.String(tableName),
	}

	output, err := conn.GetTableWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, keyspaces.ErrCodeResourceNotFoundException) {
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

	if status := aws.StringValue(output.Status); status == keyspaces.TableStatusDeleted {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}
