// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ERNameClusterAuth = "Ephemeral Resource Cluster Auth"
)

// @EphemeralResource(aws_eks_cluster_auth, name="ClusterAuth")
func newEphemeralClusterAuth(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralClusterAuth{}, nil
}

type ephemeralClusterAuth struct {
	framework.EphemeralResourceWithConfigure
}

func (e *ephemeralClusterAuth) Schema(ctx context.Context, _ ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (e *ephemeralClusterAuth) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	conn := e.Meta().STSClient(ctx)
	data := epClusterAuthData{}

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	generator, err := NewGenerator(false, false)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionReading, ERNameClusterAuth, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	token, err := generator.GetWithSTS(ctx, data.Name.ValueString(), conn)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EKS, create.ErrActionReading, ERNameClusterAuth, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, token, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.Result.Set(ctx, &data)...)
}

type epClusterAuthData struct {
	Name  types.String `tfsdk:"name"`
	Token types.String `tfsdk:"token"`
}
