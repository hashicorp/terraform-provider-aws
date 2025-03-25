// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpclattice_resource_configuration", name="Resource Configuration")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/vpclattice;vpclattice.GetResourceConfigurationOutput")
func newResourceConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceConfigurationResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type resourceConfigurationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	typeType := fwtypes.StringEnumType[awstypes.ResourceConfigurationType]()

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allow_association_to_shareable_service_network": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port_ranges": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexache.MustCompile("^((\\d{1,5}\\-\\d{1,5})|(\\d+))$"), "must contain one port number between 1 and 65535 or two separated by hyphen."),
					),
				},
			},
			names.AttrProtocol: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProtocolType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_configuration_group_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("resource_gateway_identifier")),
					stringvalidator.ConflictsWith(path.MatchRoot(names.AttrProtocol)),
					stringvalidator.AtLeastOneOf(path.MatchRoot(names.AttrProtocol), path.MatchRoot("resource_configuration_definition").AtListIndex(0).AtName("arn_resource")),
				},
			},
			"resource_gateway_identifier": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				CustomType: typeType,
				Optional:   true,
				Computed:   true,
				Default:    typeType.AttributeDefault(awstypes.ResourceConfigurationTypeSingle),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"resource_configuration_definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceConfigurationDefinitionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"arn_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[arnResourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("arn_resource"),
									path.MatchRelative().AtParent().AtName("dns_resource"),
									path.MatchRelative().AtParent().AtName("ip_resource"),
								),
								listvalidator.ConflictsWith(path.MatchRoot("port_ranges"), path.MatchRoot(names.AttrProtocol)),
							},
						},
						"dns_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dnsResourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDomainName: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrIPAddressType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.IpAddressType](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.AlsoRequires(path.MatchRoot("port_ranges")),
							},
						},
						"ip_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ipResourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrIPAddress: schema.StringAttribute{
										Required: true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.AlsoRequires(path.MatchRoot("port_ranges")),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	var input vpclattice.CreateResourceConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.ResourceConfigurationGroupIdentifier = fwflex.StringFromFramework(ctx, data.ResourceConfigurationGroupID)
	input.ResourceGatewayIdentifier = fwflex.StringFromFramework(ctx, data.ResourceGatewayID)
	input.Tags = getTagsIn(ctx)

	outputCRC, err := conn.CreateResourceConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating VPCLattice Resource Configuration (%s)", data.Name.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, outputCRC.Id)

	outputGRC, err := waitResourceConfigurationCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPCLattice Resource Configuration (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGRC, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	output, err := findResourceConfigurationByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPCLattice Resource Configuration (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	if !new.AllowAssociationToShareableServiceNetwork.Equal(old.AllowAssociationToShareableServiceNetwork) ||
		!new.PortRanges.Equal(old.PortRanges) ||
		!new.ResourceConfigurationDefinition.Equal(old.ResourceConfigurationDefinition) {
		var input vpclattice.UpdateResourceConfigurationInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ResourceConfigurationIdentifier = fwflex.StringFromFramework(ctx, new.ID)

		// "ValidationException: cannot modify resource configuration DNS or ARN".
		if input.ResourceConfigurationDefinition != nil {
			resourceConfigurationDefinition, diags := new.ResourceConfigurationDefinition.ToPtr(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			if resourceConfigurationDefinition.IPResource.IsNull() {
				input.ResourceConfigurationDefinition = nil
			}
		}

		_, err := conn.UpdateResourceConfiguration(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating VPCLattice Resource Configuration (%s)", new.ID.ValueString()), err.Error())

			return
		}

		if _, err := waitResourceConfigurationUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for VPCLattice Resource Configuration (%s) update", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	// Handle EventBridge-managed resource association deletion.
	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ValidationException](ctx, timeout, func() (any, error) {
		return conn.DeleteResourceConfiguration(ctx, &vpclattice.DeleteResourceConfigurationInput{
			ResourceConfigurationIdentifier: fwflex.StringFromFramework(ctx, data.ID),
		})
	}, "has existing association with service networks")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPCLattice Resource Configuration (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitResourceConfigurationDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPCLattice Resource Configuration (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findResourceConfigurationByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetResourceConfigurationOutput, error) {
	input := vpclattice.GetResourceConfigurationInput{
		ResourceConfigurationIdentifier: aws.String(id),
	}

	output, err := conn.GetResourceConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusResourceConfiguration(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findResourceConfigurationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitResourceConfigurationCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetResourceConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceConfigurationStatusCreateInProgress),
		Target:                    enum.Slice(awstypes.ResourceConfigurationStatusActive),
		Refresh:                   statusResourceConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetResourceConfigurationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))

		return output, err
	}

	return nil, err
}

func waitResourceConfigurationUpdated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetResourceConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceConfigurationStatusUpdateInProgress),
		Target:                    enum.Slice(awstypes.ResourceConfigurationStatusActive),
		Refresh:                   statusResourceConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetResourceConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitResourceConfigurationDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetResourceConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceConfigurationStatusActive, awstypes.ResourceConfigurationStatusDeleteInProgress),
		Target:  []string{},
		Refresh: statusResourceConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetResourceConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}

type resourceConfigurationResourceModel struct {
	AllowAssociationToShareableServiceNetwork types.Bool                                                            `tfsdk:"allow_association_to_shareable_service_network"`
	ARN                                       types.String                                                          `tfsdk:"arn"`
	ID                                        types.String                                                          `tfsdk:"id"`
	Name                                      types.String                                                          `tfsdk:"name"`
	PortRanges                                fwtypes.SetOfString                                                   `tfsdk:"port_ranges"`
	Protocol                                  fwtypes.StringEnum[awstypes.ProtocolType]                             `tfsdk:"protocol"`
	ResourceConfigurationDefinition           fwtypes.ListNestedObjectValueOf[resourceConfigurationDefinitionModel] `tfsdk:"resource_configuration_definition"`
	ResourceConfigurationGroupID              types.String                                                          `tfsdk:"resource_configuration_group_id"`
	ResourceGatewayID                         types.String                                                          `tfsdk:"resource_gateway_identifier"`
	Tags                                      tftags.Map                                                            `tfsdk:"tags"`
	TagsAll                                   tftags.Map                                                            `tfsdk:"tags_all"`
	Timeouts                                  timeouts.Value                                                        `tfsdk:"timeouts"`
	Type                                      fwtypes.StringEnum[awstypes.ResourceConfigurationType]                `tfsdk:"type"`
}

type resourceConfigurationDefinitionModel struct {
	ARNResource fwtypes.ListNestedObjectValueOf[arnResourceModel] `tfsdk:"arn_resource"`
	DNSResource fwtypes.ListNestedObjectValueOf[dnsResourceModel] `tfsdk:"dns_resource"`
	IPResource  fwtypes.ListNestedObjectValueOf[ipResourceModel]  `tfsdk:"ip_resource"`
}

func (r *resourceConfigurationDefinitionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ResourceConfigurationDefinitionMemberArnResource:
		var data arnResourceModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		r.ARNResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.ResourceConfigurationDefinitionMemberDnsResource:
		var data dnsResourceModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		r.DNSResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.ResourceConfigurationDefinitionMemberIpResource:
		var data ipResourceModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		r.IPResource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	}

	return diags
}

func (r resourceConfigurationDefinitionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var v any

	switch {
	case !r.ARNResource.IsNull():
		data, d := r.ARNResource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var apiObject awstypes.ResourceConfigurationDefinitionMemberArnResource
		diags.Append(fwflex.Expand(ctx, data, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		v = &apiObject
	case !r.DNSResource.IsNull():
		data, d := r.DNSResource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var apiObject awstypes.ResourceConfigurationDefinitionMemberDnsResource
		diags.Append(fwflex.Expand(ctx, data, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		v = &apiObject
	case !r.IPResource.IsNull():
		data, d := r.IPResource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var apiObject awstypes.ResourceConfigurationDefinitionMemberIpResource
		diags.Append(fwflex.Expand(ctx, data, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		v = &apiObject
	}

	return v, diags
}

var (
	_ fwflex.Expander  = resourceConfigurationDefinitionModel{}
	_ fwflex.Flattener = &resourceConfigurationDefinitionModel{}
)

type arnResourceModel struct {
	ARN fwtypes.ARN `tfsdk:"arn"`
}

type dnsResourceModel struct {
	DomainName    types.String                                                    `tfsdk:"domain_name"`
	IPAddressType fwtypes.StringEnum[awstypes.ResourceConfigurationIpAddressType] `tfsdk:"ip_address_type"`
}

type ipResourceModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
}
