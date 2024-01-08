// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmsap

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmsap"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssmsap/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findApplicationByID(ctx context.Context, conn *ssmsap.Client, id string) (*awstypes.Application, map[string]string, error) {
	in := &ssmsap.GetApplicationInput{
		ApplicationId: aws.String(id),
	}

	out, err := conn.GetApplication(ctx, in)
	if err != nil {

		//the API currently doesn't return a ResourceNotFoundException; instead it returns a ValidationException if the application is not found
		// var nfe *awstypes.ResourceNotFoundException
		var nfe *awstypes.ValidationException
		if errors.As(err, &nfe) {
			return nil, nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, nil, err
	}

	if out == nil || out.Application == nil {
		return nil, nil, tfresource.NewEmptyResultError(in)
	}

	return out.Application, out.Tags, nil
}

func findComponentByID(ctx context.Context, conn *ssmsap.Client, application_id string, id string) (*awstypes.Component, map[string]string, error) {
	in := &ssmsap.GetComponentInput{
		ApplicationId: aws.String(application_id),
		ComponentId:   aws.String(id),
	}

	out, err := conn.GetComponent(ctx, in)
	if err != nil {

		//the API currently doesn't return a ResourceNotFoundException; instead it returns a ValidationException if the component is not found
		// var nfe *awstypes.ResourceNotFoundException
		var nfe *awstypes.ValidationException
		if errors.As(err, &nfe) {
			return nil, nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, nil, err
	}

	if out == nil || out.Component == nil {
		return nil, nil, tfresource.NewEmptyResultError(in)
	}

	return out.Component, out.Tags, nil
}
