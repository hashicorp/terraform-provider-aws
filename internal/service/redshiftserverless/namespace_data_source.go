// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_redshiftserverless_namespace", name="Namespace")
func newDataSourceNamespace(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceNamespace{}, nil
}

const (
	DSNameNamespace = "Namespace Data Source"
)

type dataSourceNamespace struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceNamespace) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"admin_password_secret_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"admin_password_secret_kms_key_id": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"admin_username": schema.StringAttribute{
				Computed: true,
			},
			"db_name": schema.StringAttribute{
				Computed: true,
			},
			"default_iam_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"iam_roles": schema.SetAttribute{
				CustomType:  fwtypes.SetOfARNType,
				ElementType: fwtypes.ARNType,
				Computed:    true,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"log_exports": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"namespace_id": schema.StringAttribute{
				Computed: true,
			},
			"namespace_name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *dataSourceNamespace) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().RedshiftServerlessClient(ctx)

	var data dataSourceNamespaceData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNamespaceByName(ctx, conn, data.NamespaceName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionReading, DSNameNamespace, data.NamespaceName.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithIgnoredFieldNamesAppend("IamRoles"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = flex.StringToFramework(ctx, out.NamespaceName)
	data.ARN = flex.StringToFrameworkARN(ctx, out.NamespaceArn)
	data.IAMRoles = fwtypes.NewSetValueOfMust[fwtypes.ARN](ctx, flattenNamespaceIAMRoles(ctx, out.IamRoles))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceNamespaceData struct {
	AdminPasswordSecretArn      fwtypes.ARN                     `tfsdk:"admin_password_secret_arn"`
	AdminPasswordSecretKMSKeyID fwtypes.ARN                     `tfsdk:"admin_password_secret_kms_key_id"`
	AdminUsername               types.String                    `tfsdk:"admin_username"`
	ARN                         fwtypes.ARN                     `tfsdk:"arn"`
	DBName                      types.String                    `tfsdk:"db_name"`
	DefaultIAMRoleARN           fwtypes.ARN                     `tfsdk:"default_iam_role_arn"`
	IAMRoles                    fwtypes.SetValueOf[fwtypes.ARN] `tfsdk:"iam_roles"`
	ID                          types.String                    `tfsdk:"id"`
	KMSKeyID                    types.String                    `tfsdk:"kms_key_id"`
	LogExports                  types.Set                       `tfsdk:"log_exports"`
	NamespaceID                 types.String                    `tfsdk:"namespace_id"`
	NamespaceName               types.String                    `tfsdk:"namespace_name"`
}
