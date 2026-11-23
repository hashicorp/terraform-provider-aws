// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockframework

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockframework"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockframework/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkResource("aws_mockframework_qux", name="Qux")
func newQuxResource(context.Context) (resource.ResourceWithConfigure, error) {
	var r quxResource
	r.bazQux = &r

	return &r, nil
}

type quxResource struct {
	bazQuxResource
}

func (r *quxResource) create(ctx context.Context, data *bazQuxResourceModel) (string, error) {
	conn := r.Meta().MockFrameworkClient(ctx)

	input := mockframework.CreateQuxInput{
		Name: fwflex.StringFromFramework(ctx, data.Name),
	}

	output, err := conn.CreateQux(ctx, &input)
	if err != nil {
		return "", err
	}

	return aws.ToString(output.QuxId), nil
}

func (r *quxResource) delete(ctx context.Context, data *bazQuxResourceModel) error {
	conn := r.Meta().MockFrameworkClient(ctx)

	input := mockframework.DeleteQuxInput{
		QuxId: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteQux(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func (r *quxResource) findByID(ctx context.Context, id string) (any, error) {
	return findQuxByID(ctx, r.Meta().MockFrameworkClient(ctx), id)
}
