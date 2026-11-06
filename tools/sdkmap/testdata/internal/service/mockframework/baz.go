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

// @FrameworkResource("aws_mockframework_baz", name="Baz")
func newBazResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &bazResource{}
	r.bazQux = r

	return r, nil
}

type bazResource struct {
	bazQuxResource
}

func (r *bazResource) create(ctx context.Context, data *bazQuxResourceModel) (string, error) {
	conn := r.Meta().MockFrameworkClient(ctx)

	input := mockframework.CreateBazInput{
		Name: fwflex.StringFromFramework(ctx, data.Name),
	}

	output, err := conn.CreateBaz(ctx, &input)
	if err != nil {
		return "", err
	}

	return aws.ToString(output.BazId), nil
}

func (r *bazResource) delete(ctx context.Context, data *bazQuxResourceModel) error {
	conn := r.Meta().MockFrameworkClient(ctx)

	input := mockframework.DeleteBazInput{
		BazId: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteBaz(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func (r *bazResource) findByID(ctx context.Context, id string) (any, error) {
	return findBazByID(ctx, r.Meta().MockFrameworkClient(ctx), id)
}
