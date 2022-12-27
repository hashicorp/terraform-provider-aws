package appconfig

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindExtensionById(ctx context.Context, conn *appconfig.AppConfig, id string) (*appconfig.GetExtensionOutput, error) {
	in := &appconfig.GetExtensionInput{ExtensionIdentifier: aws.String(id)}
	out, err := conn.GetExtensionWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func FindExtensionAssociationById(ctx context.Context, conn *appconfig.AppConfig, id string) (*appconfig.GetExtensionAssociationOutput, error) {
	in := &appconfig.GetExtensionAssociationInput{ExtensionAssociationId: aws.String(id)}
	out, err := conn.GetExtensionAssociationWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
