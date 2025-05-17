// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"reflect"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_iam_service_linked_role",name="Service Linked Role")
func newDataSourceServiceLinkedRole(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServiceLinkedRole{}, nil
}

type dataSourceServiceLinkedRole struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceServiceLinkedRole) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aws_service_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`\.`),
						"must be a full service hostname e.g. elasticbeanstalk.amazonaws.com",
					),
				},
			},
			"custom_suffix": schema.StringAttribute{
				Optional: true,
			},
			"create_if_missing": schema.BoolAttribute{
				Optional: true,
			},
			"create_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrPath: schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"unique_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceServiceLinkedRole) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data ServiceLinkedRoleDataSourceModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var role *awstypes.Role
	conn := d.Meta().IAMClient(ctx)

	//AWS API does not provide a Get/List method for Service Linked Roles.
	//Matching the role path prefix and role name using regex is the only option to find Service Linked roles
	var nameRegex string
	pathPrefix := fmt.Sprintf("/aws-service-role/%s", data.AWSServiceName.ValueString())
	customSuffix := data.CustomSuffix.ValueString()
	awsServiceName := data.AWSServiceName.ValueString()
	if customSuffix == "" {
		//regex to match AWSServiceRole prefix and 1 or more characters exluding _ (underscore)
		nameRegex = `AWSServiceRole[^_]+$`
	} else {
		//regex to match AWSServiceRole prefix and any role name, _ (underscore) and the provided suffix
		nameRegex = fmt.Sprintf(`AWSServiceRole[0-9A-Za-z]+_%s$`, customSuffix)
	}
	roles, err := findRoles(ctx, conn, pathPrefix, nameRegex)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading IAM Service Linked Role (%s)", awsServiceName), err.Error())
		return
	}
	switch len(roles) {
	case 0:
		if data.CreateIfMissing.ValueBool() {
			input := &iam.CreateServiceLinkedRoleInput{
				AWSServiceName: &awsServiceName,
			}
			if customSuffix != "" {
				input.CustomSuffix = &customSuffix
			}

			output, err := conn.CreateServiceLinkedRole(ctx, input)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("creating IAM Service Linked Role (%s)", awsServiceName), err.Error())
				return
			}
			role = output.Role // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
		} else {
			response.Diagnostics.AddError(fmt.Sprintf("reading IAM Service Linked Role (%s)", awsServiceName), "Role was not found.")
			return
		}
	case 1:
		role = &roles[0]
	default:
		response.Diagnostics.AddError(fmt.Sprintf("reading IAM Service Linked Role (%s)", awsServiceName), "More than one role was returned.")
	}
	response.Diagnostics.Append(fwflex.Flatten(ctx, role, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type ServiceLinkedRoleDataSourceModel struct {
	ARN             types.String      `tfsdk:"arn"`
	AWSServiceName  types.String      `tfsdk:"aws_service_name"`
	CreateDate      timetypes.RFC3339 `tfsdk:"create_date"`
	CreateIfMissing types.Bool        `tfsdk:"create_if_missing"`
	CustomSuffix    types.String      `tfsdk:"custom_suffix"`
	Description     types.String      `tfsdk:"description"`
	Path            types.String      `tfsdk:"path"`
	RoleId          types.String      `tfsdk:"unique_id"`
	RoleName        types.String      `tfsdk:"name"`
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
