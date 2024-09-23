// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"reflect"
	"regexp"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name=Roles)
func newDataSourceRoles(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceRoles{}

	return d, nil
}

const (
	DSNameRoles = "Roles Data Source"
)

type dataSourceRoles struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceRoles) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_iam_roles"
}

func (d *dataSourceRoles) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARNs: schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"name_regex": schema.StringAttribute{
				Optional: true,
			},
			names.AttrNames: schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"path_prefix": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *dataSourceRoles) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data RolesDataSourceModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().IAMClient(ctx)

	pathPrefix := data.PathPrefix.ValueString()
	nameRegex := data.NameRegex.ValueString()

	if _, err := regexp.Compile(nameRegex); err != nil {
		response.Diagnostics.AddError("Invalid name_regex", err.Error())
		return
	}
	results, err := findRoles(ctx, conn, pathPrefix, nameRegex)
	if err != nil {
		response.Diagnostics.AddError("find roles", err.Error())
		return
	}

	var roleResuls RoleResultsModel
	roleResuls.ARNs = make([]*string, len(results))
	roleResuls.Names = make([]*string, len(results))
	for i := 0; i < len(results); i++ {
		roleResuls.ARNs[i] = results[i].Arn
		roleResuls.Names[i] = results[i].RoleName
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, &roleResuls, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type RoleResultsModel struct {
	ARNs  []*string
	Names []*string
}

type RolesDataSourceModel struct {
	ARNs       types.List   `tfsdk:"arns"`
	Names      types.List   `tfsdk:"names"`
	PathPrefix types.String `tfsdk:"path_prefix"`
	NameRegex  types.String `tfsdk:"name_regex"`
}

func findRoles(ctx context.Context, conn *iam.Client, pathPrefix string, nameRegex string) ([]awstypes.Role, error) {
	var results []awstypes.Role

	input := &iam.ListRolesInput{}
	if pathPrefix != "" {
		input.PathPrefix = &pathPrefix
	}

	pages := iam.NewListRolesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("reading IAM roles: %s", err)
		}

		for _, role := range page.Roles {
			if reflect.ValueOf(role).IsZero() {
				continue
			}

			if nameRegex != "" && !regexache.MustCompile(nameRegex).MatchString(aws.ToString(role.RoleName)) {
				continue
			}

			results = append((results), role)
		}
	}
	return results, nil
}
