// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package transfer

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_transfer_web_app", name="Web App")
// @Tags(identifierAttribute="arn")
func newWebAppResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &webAppResource{}

	return r, nil
}

type webAppResource struct {
	framework.ResourceWithModel[webAppResourceModel]
}

func (r *webAppResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_endpoint": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
					stringvalidator.ConflictsWith(path.MatchRoot("endpoint_details").AtListIndex(0).AtName("vpc")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"web_app_endpoint_policy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.WebAppEndpointPolicy](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"web_app_id":    framework.IDAttribute(),
			"web_app_units": framework.ResourceOptionalComputedListOfObjectsAttribute[webAppUnitsModel](ctx, 1, nil),
		},
		Blocks: map[string]schema.Block{
			"endpoint_details": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webAppEndpointDetailsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"vpc": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webAppEndpointDetailsVPCModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ConflictsWith(path.MatchRoot("access_endpoint")),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSecurityGroupIDs: schema.SetAttribute{
										ElementType: types.StringType,
										Optional:    true,
										Computed:    true,
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
											setplanmodifier.UseStateForUnknown(),
										},
									},
									names.AttrSubnetIDs: schema.SetAttribute{
										ElementType: types.StringType,
										Required:    true,
									},
									names.AttrVPCEndpointID: schema.StringAttribute{
										Computed: true,
									},
									names.AttrVPCID: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
			"identity_provider_details": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webAppIdentityProviderDetailsModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"identity_center_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[identityCenterConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"application_arn": schema.StringAttribute{
										Computed: true,
									},
									"instance_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrRole: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *webAppResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data webAppResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	var input transfer.CreateWebAppInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateWebApp(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating Transfer Web App", err.Error())

		return
	}

	webAppID := aws.ToString(output.WebAppId)
	webApp, err := findWebAppByID(ctx, conn, webAppID)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Transfer Web App (%s)", webAppID), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, webApp, &data, fwflex.WithFieldNamePrefix("Described"))...)
	if response.Diagnostics.HasError() {
		return
	}
	if webApp.DescribedEndpointDetails != nil {
		switch t := webApp.DescribedEndpointDetails.(type) {
		case *awstypes.DescribedWebAppEndpointDetailsMemberVpc:
			response.Diagnostics.Append(setSecurityGroupIDsFromVPCEndpointID(ctx, r.Meta().EC2Client(ctx), t.Value, &data)...)
			if response.Diagnostics.HasError() {
				return
			}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *webAppResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data webAppResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	webAppID := fwflex.StringValueFromFramework(ctx, data.WebAppID)
	out, err := findWebAppByID(ctx, conn, webAppID)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Transfer Web App (%s)", webAppID), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("Described"))...)
	if response.Diagnostics.HasError() {
		return
	}
	if out.DescribedEndpointDetails != nil {
		switch t := out.DescribedEndpointDetails.(type) {
		case *awstypes.DescribedWebAppEndpointDetailsMemberVpc:
			response.Diagnostics.Append(setSecurityGroupIDsFromVPCEndpointID(ctx, r.Meta().EC2Client(ctx), t.Value, &data)...)
			if response.Diagnostics.HasError() {
				return
			}
		}
	}

	setTagsOut(ctx, out.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *webAppResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old webAppResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		webAppID := fwflex.StringValueFromFramework(ctx, new.WebAppID)
		var input transfer.UpdateWebAppInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input, fwflex.WithIgnoredFieldNamesAppend("IdentityProviderDetails"), fwflex.WithIgnoredFieldNamesAppend("EndpointDetails"))...)
		if response.Diagnostics.HasError() {
			return
		}

		//
		if !new.IdentityProviderDetails.Equal(old.IdentityProviderDetails) {
			if v, diags := new.IdentityProviderDetails.ToPtr(ctx); v != nil && !diags.HasError() {
				if v, diags := v.IdentityCenterConfig.ToPtr(ctx); v != nil && !diags.HasError() {
					input.IdentityProviderDetails = &awstypes.UpdateWebAppIdentityProviderDetailsMemberIdentityCenterConfig{
						Value: awstypes.UpdateWebAppIdentityCenterConfig{
							Role: fwflex.StringFromFramework(ctx, v.Role),
						},
					}
				} else {
					response.Diagnostics.Append(diags...)
					return
				}
			} else {
				response.Diagnostics.Append(diags...)
				return
			}
		}
		if !new.EndpointDetails.Equal(old.EndpointDetails) {
			if v, diags := new.EndpointDetails.ToPtr(ctx); v != nil && !diags.HasError() {
				if v, diags := v.VPC.ToPtr(ctx); v != nil && !diags.HasError() {
					input.EndpointDetails = &awstypes.UpdateWebAppEndpointDetailsMemberVpc{
						Value: awstypes.UpdateWebAppVpcConfig{
							SubnetIds: fwflex.ExpandFrameworkStringValueSet(ctx, v.SubnetIDs),
						},
					}
					// Reset AccessEndpoint to null to avoid conflicts.
					// AccessEndpoint must not be specified when EndpointDetails is set.
					// Note:
					// AccessEndpoint is a computed attribute when endpoint_details.vpc is specified,
					// because endpoint_details.vpc and access_endpoint are defined as conflicting
					// attributes in the schema.
					input.AccessEndpoint = nil
				} else {
					response.Diagnostics.Append(diags...)
					return
				}
			} else {
				response.Diagnostics.Append(diags...)
				return
			}
		}

		_, err := conn.UpdateWebApp(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Transfer Web App (%s)", webAppID), err.Error())

			return
		}

		webApp, err := findWebAppByID(ctx, conn, webAppID)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading Transfer Web App (%s)", webAppID), err.Error())

			return
		}

		// Set values for unknowns.
		response.Diagnostics.Append(fwflex.Flatten(ctx, webApp, &new, fwflex.WithFieldNamePrefix("Described"))...)
		if response.Diagnostics.HasError() {
			return
		}
		if webApp.DescribedEndpointDetails != nil {
			switch t := webApp.DescribedEndpointDetails.(type) {
			case *awstypes.DescribedWebAppEndpointDetailsMemberVpc:
				response.Diagnostics.Append(setSecurityGroupIDsFromVPCEndpointID(ctx, r.Meta().EC2Client(ctx), t.Value, &new)...)
				if response.Diagnostics.HasError() {
					return
				}
			}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *webAppResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data webAppResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	webAppID := fwflex.StringValueFromFramework(ctx, data.WebAppID)
	input := transfer.DeleteWebAppInput{
		WebAppId: aws.String(webAppID),
	}
	_, err := conn.DeleteWebApp(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Transfer Web App (%s)", webAppID), err.Error())

		return
	}
}

func (r *webAppResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("web_app_id"), request, response)
}

func findWebAppByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedWebApp, error) {
	input := transfer.DescribeWebAppInput{
		WebAppId: aws.String(id),
	}

	return findWebApp(ctx, conn, &input)
}

func findWebApp(ctx context.Context, conn *transfer.Client, input *transfer.DescribeWebAppInput) (*awstypes.DescribedWebApp, error) {
	out, err := conn.DescribeWebApp(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.WebApp == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.WebApp, nil
}

type webAppResourceModel struct {
	framework.WithRegionModel
	AccessEndpoint          types.String                                                        `tfsdk:"access_endpoint"`
	ARN                     types.String                                                        `tfsdk:"arn"`
	EndpointDetails         fwtypes.ListNestedObjectValueOf[webAppEndpointDetailsModel]         `tfsdk:"endpoint_details"`
	IdentityProviderDetails fwtypes.ListNestedObjectValueOf[webAppIdentityProviderDetailsModel] `tfsdk:"identity_provider_details"`
	Tags                    tftags.Map                                                          `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                          `tfsdk:"tags_all"`
	WebAppEndpointPolicy    fwtypes.StringEnum[awstypes.WebAppEndpointPolicy]                   `tfsdk:"web_app_endpoint_policy"`
	WebAppID                types.String                                                        `tfsdk:"web_app_id"`
	WebAppUnits             fwtypes.ListNestedObjectValueOf[webAppUnitsModel]                   `tfsdk:"web_app_units"`
}

type webAppEndpointDetailsModel struct {
	VPC fwtypes.ListNestedObjectValueOf[webAppEndpointDetailsVPCModel] `tfsdk:"vpc"`
}

type webAppEndpointDetailsVPCModel struct {
	SecurityGroupIDs fwtypes.SetValueOf[types.String] `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
	VPCEndpointID    types.String                     `tfsdk:"vpc_endpoint_id"`
	VPCID            types.String                     `tfsdk:"vpc_id"`
}

var (
	_ fwflex.Expander  = webAppEndpointDetailsModel{}
	_ fwflex.Flattener = &webAppEndpointDetailsModel{}
)

func (m webAppEndpointDetailsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.VPC.IsNull():
		data, d := m.VPC.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.WebAppEndpointDetailsMemberVpc
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

func (m *webAppEndpointDetailsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.DescribedWebAppEndpointDetailsMemberVpc:
		var data webAppEndpointDetailsVPCModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		if data.SecurityGroupIDs.IsNull() || data.SecurityGroupIDs.IsUnknown() {
			data.SecurityGroupIDs = fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{})
		}
		m.VPC = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("artifact flatten: %s", reflect.TypeOf(v).String()),
		)
	}
	return diags
}

func setSecurityGroupIDsFromVPCEndpointID(ctx context.Context, conn *ec2.Client, vpcConfig awstypes.DescribedWebAppVpcConfig, new *webAppResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	newEndpointDetailsData, d := new.EndpointDetails.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	vpcData, d := newEndpointDetailsData.VPC.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vpcEndpointID := aws.ToString(vpcConfig.VpcEndpointId)
	sgIDs, err := findSecurityGroupIDsFromVPCEndpointID(ctx, conn, vpcEndpointID)
	if err != nil {
		diags.AddError("reading VPC Endpoint Security Groups", fmt.Sprintf("unable to read security groups for VPC Endpoint (%s): %s", vpcEndpointID, err.Error()))
		return diags
	}
	vpcData.SecurityGroupIDs = fwflex.FlattenFrameworkStringValueSetOfString(ctx, sgIDs)
	newEndpointDetailsData.VPC = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, vpcData)
	new.EndpointDetails = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, newEndpointDetailsData)
	return diags
}

func findSecurityGroupIDsFromVPCEndpointID(ctx context.Context, conn *ec2.Client, vpcEndpointID string) ([]string, error) {
	vpcEndpoint, err := tfec2.FindVPCEndpointByID(ctx, conn, vpcEndpointID)
	if err != nil {
		return nil, fmt.Errorf("could not read security groups for VPC Endpoint (%s): %w", vpcEndpointID, err)
	}
	securityGroupIDs := make([]string, 0, len(vpcEndpoint.Groups))
	for _, group := range vpcEndpoint.Groups {
		if group.GroupId == nil {
			continue
		}
		securityGroupIDs = append(securityGroupIDs, aws.ToString(group.GroupId))
	}
	return securityGroupIDs, nil
}

type webAppIdentityProviderDetailsModel struct {
	IdentityCenterConfig fwtypes.ListNestedObjectValueOf[identityCenterConfigModel] `tfsdk:"identity_center_config"`
}

type identityCenterConfigModel struct {
	ApplicationARN types.String `tfsdk:"application_arn"`
	InstanceARN    fwtypes.ARN  `tfsdk:"instance_arn"`
	Role           fwtypes.ARN  `tfsdk:"role"`
}

type webAppUnitsModel struct {
	Provisioned types.Int64 `tfsdk:"provisioned"`
}

var (
	_ fwflex.Expander  = webAppUnitsModel{}
	_ fwflex.Flattener = &webAppUnitsModel{}
)

func (m webAppUnitsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.Provisioned.IsNull():
		var r awstypes.WebAppUnitsMemberProvisioned
		r.Value = aws.ToInt32(fwflex.Int32FromFrameworkInt64(ctx, &m.Provisioned))
		return &r, diags
	}
	return nil, diags
}

func (m *webAppUnitsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.WebAppUnitsMemberProvisioned:
		m.Provisioned = fwflex.Int32ToFrameworkInt64(ctx, &t.Value)
	}
	return diags
}

var (
	_ fwflex.Expander  = webAppIdentityProviderDetailsModel{}
	_ fwflex.Flattener = &webAppIdentityProviderDetailsModel{}
)

func (m webAppIdentityProviderDetailsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.IdentityCenterConfig.IsNull():
		data, d := m.IdentityCenterConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.WebAppIdentityProviderDetailsMemberIdentityCenterConfig
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

func (m *webAppIdentityProviderDetailsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.DescribedWebAppIdentityProviderDetailsMemberIdentityCenterConfig:
		var data identityCenterConfigModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.IdentityCenterConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("artifact flatten: %s", reflect.TypeOf(v).String()),
		)
	}
	return diags
}
