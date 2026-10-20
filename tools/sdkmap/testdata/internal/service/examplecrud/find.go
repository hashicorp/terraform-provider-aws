// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplecrud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplecrud"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplecrud/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findWidgetByID(ctx context.Context, conn *examplecrud.Client, id string) (*awstypes.Widget, error) {
	input := examplecrud.DescribeWidgetInput{
		WidgetId: aws.String(id),
	}

	output, err := conn.DescribeWidget(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Widget == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Widget, nil
}

func findGadgetByID(ctx context.Context, conn *examplecrud.Client, id string) (*awstypes.Gadget, error) {
	input := examplecrud.DescribeGadgetInput{
		GadgetId: aws.String(id),
	}

	output, err := conn.DescribeGadget(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Gadget == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Gadget, nil
}
