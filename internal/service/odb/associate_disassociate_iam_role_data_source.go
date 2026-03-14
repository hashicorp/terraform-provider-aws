// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var AssociateDisassociateIAMRoleDataSource = newDataSourceAssociateDisassociateIAMRole

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_odb_associate_disassociate_iam_role", name="Associate Disassociate IAM Role")
func newDataSourceAssociateDisassociateIAMRole(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAssociateDisassociateIAMRole{}, nil
}

const (
	DSNameAssociateDisassociateIAMRole = "Associate Disassociate IAM Role Data Source"
)

type dataSourceAssociateDisassociateIAMRole struct {
	framework.DataSourceWithModel[dataSourceAssociateDisassociateIAMRoleModel]
}

func (d *dataSourceAssociateDisassociateIAMRole) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aws_integration": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[odbtypes.IamRoleStatus](),
				Computed:   true,
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"composite_arn": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceCompositeARNModel](ctx),
				Validators: []validator.List{
					// Only one combination of resource ARN and IAM role ARN is mandatory
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
					listvalidator.IsRequired(),
				},
				Description: "Combination of resource ARN and IAM role ARN is mandatory",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrIAMRoleARN: schema.StringAttribute{
							Required: true,
						},
						names.AttrResourceARN: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceAssociateDisassociateIAMRole) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data dataSourceAssociateDisassociateIAMRoleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var combinedARNs []dataSourceCompositeARNModel
	data.CombinedARN.ElementsAs(ctx, &combinedARNs, false)

	out, err := FindAssociatedDisassociatedIAMRoleOracleDBDataSource(ctx, conn, combinedARNs[0].ResourceARN.ValueString(), combinedARNs[0].IAMRoleARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameAssociateDisassociateIAMRole, data.CombinedARN.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("AssociateDisassociateIAMRole"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FindAssociatedDisassociatedIAMRoleOracleDBDataSource(ctx context.Context, conn *odb.Client, resourceARN string, roleARN string) (*odbtypes.IamRole, error) {
	parsedResourceARN, err := arn.Parse(resourceARN)
	if err != nil {
		return nil, err
	}
	resourceType := strings.Split(parsedResourceARN.Resource, "/")[0]
	resourceId := strings.Split(parsedResourceARN.Resource, "/")[1]
	switch resourceType {
	case "cloud-vm-cluster":
		input := odb.GetCloudVmClusterInput{
			CloudVmClusterId: &resourceId,
		}
		out, err := conn.GetCloudVmCluster(ctx, &input)
		if err != nil {
			return nil, err
		}
		iamRolesList := out.CloudVmCluster.IamRoles

		for _, element := range iamRolesList {
			if *element.IamRoleArn == roleARN {
				//we found the correct role
				return &element, nil
			}
		}
		err = errors.New("no IAM role found for the vm cluster : " + resourceARN)
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}

	case "cloud-autonomous-vm-cluster":
		input := odb.GetCloudAutonomousVmClusterInput{
			CloudAutonomousVmClusterId: &resourceId,
		}
		out, err := conn.GetCloudAutonomousVmCluster(ctx, &input)
		if err != nil {
			return nil, err
		}
		for _, element := range out.CloudAutonomousVmCluster.IamRoles {
			if *element.IamRoleArn == roleARN {
				//We found a match
				return &element, nil
			}
		}
		err = errors.New("no IAM role found for the cloud autonomous vm cluster : " + resourceARN)
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}
	}
	return nil, errors.New("IAM role association / disassociation not supported : " + resourceARN)
}

type dataSourceAssociateDisassociateIAMRoleModel struct {
	framework.WithRegionModel
	CombinedARN    fwtypes.ListNestedObjectValueOf[dataSourceCompositeARNModel] `tfsdk:"composite_arn"`
	AWSIntegration types.String                                                 `tfsdk:"aws_integration"`
	Status         fwtypes.StringEnum[odbtypes.IamRoleStatus]                   `tfsdk:"status"`
	StatusReason   types.String                                                 `tfsdk:"status_reason"`
}

type dataSourceCompositeARNModel struct {
	IAMRoleARN  types.String `tfsdk:"iam_role_arn"`
	ResourceARN types.String `tfsdk:"resource_arn"`
}
