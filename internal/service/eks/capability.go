// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package eks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	set "github.com/hashicorp/go-set/v3"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_eks_capability", name="Capability")
// @Tags(identifierAttribute="arn")
func newCapabilityResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &capabilityResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type capabilityResource struct {
	framework.ResourceWithModel[capabilityResourceModel]
	framework.WithTimeouts
}

func (r *capabilityResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"capability_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrClusterName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delete_propagation_policy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapabilityDeletePropagationPolicy](),
				Required:   true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapabilityType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[capabilityConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"argo_cd": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[argoCDConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrNamespace: schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplaceIfConfigured(),
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"server_url": schema.StringAttribute{
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"aws_idc": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[argoCDAWSIDCConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"idc_instance_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"idc_managed_application_arn": schema.StringAttribute{
													Computed: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(),
													},
												},
												"idc_region": schema.StringAttribute{
													Optional: true,
													Computed: true,
													Validators: []validator.String{
														fwvalidators.AWSRegion(),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplaceIfConfigured(),
														stringplanmodifier.UseStateForUnknown(),
													},
												},
											},
										},
									},
									"network_access": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[argoCDNetworkAccessConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"vpce_ids": schema.SetAttribute{
													CustomType:  fwtypes.SetOfStringType,
													ElementType: types.StringType,
													Optional:    true,
												},
											},
										},
									},
									"rbac_role_mapping": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[argoCDRoleMappingModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrRole: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.ArgoCdRole](),
													Required:   true,
												},
											},
											Blocks: map[string]schema.Block{
												"identity": schema.SetNestedBlock{
													CustomType: fwtypes.NewSetNestedObjectTypeOf[SSOIdentity](ctx),
													Validators: []validator.Set{
														setvalidator.IsRequired(),
														setvalidator.SizeAtLeast(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrID: schema.StringAttribute{
																Required: true,
															},
															names.AttrType: schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.SsoIdentityType](),
																Required:   true,
															},
														},
													},
												},
											},
										},
									},
								},
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

func (r *capabilityResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data capabilityResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	clusterName, capabilityName := fwflex.StringValueFromFramework(ctx, data.ClusterName), fwflex.StringValueFromFramework(ctx, data.CapabilityName)
	var input eks.CreateCapabilityInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientRequestToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateCapability(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating EKS Capability (%s,%s)", clusterName, capabilityName), err.Error())

		return
	}

	capability, err := waitCapabilityCreated(ctx, conn, clusterName, capabilityName, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EKS Capability (%s,%s) create", clusterName, capabilityName), err.Error())

		return
	}

	// Normalize.
	if capability.Configuration != nil && capability.Configuration.ArgoCd == nil {
		capability.Configuration = nil
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, capability, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *capabilityResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data capabilityResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	clusterName, capabilityName := fwflex.StringValueFromFramework(ctx, data.ClusterName), fwflex.StringValueFromFramework(ctx, data.CapabilityName)
	output, err := findCapabilityByTwoPartKey(ctx, conn, clusterName, capabilityName)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EKS Capability (%s,%s)", clusterName, capabilityName), err.Error())

		return
	}

	// Normalize.
	if output.Configuration != nil && output.Configuration.ArgoCd == nil {
		output.Configuration = nil
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *capabilityResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new capabilityResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		clusterName, capabilityName := fwflex.StringValueFromFramework(ctx, new.ClusterName), fwflex.StringValueFromFramework(ctx, new.CapabilityName)
		var input eks.UpdateCapabilityInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input, fwflex.WithIgnoredFieldNamesAppend("RbacRoleMappings"))...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientRequestToken = aws.String(sdkid.UniqueId())

		// argo_cd block can only be modified in-place (not added or removed).
		var oldConfiguration, newConfiguration awstypes.CapabilityConfigurationRequest
		response.Diagnostics.Append(fwflex.Expand(ctx, old.Configuration, &oldConfiguration)...)
		if response.Diagnostics.HasError() {
			return
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, new.Configuration, &newConfiguration)...)
		if response.Diagnostics.HasError() {
			return
		}

		if oldArgoCD, newArgoCD := oldConfiguration.ArgoCd, newConfiguration.ArgoCd; oldArgoCD != nil && newArgoCD != nil {
			add, remove, update, _ := intflex.DiffSlicesWithModify(oldArgoCD.RbacRoleMappings, newArgoCD.RbacRoleMappings,
				func(a, b awstypes.ArgoCdRoleMapping) bool {
					hashIdentity := func(v awstypes.SsoIdentity) string {
						return string(v.Type) + ":" + aws.ToString(v.Id)
					}
					return a.Role == b.Role && set.HashSetFromFunc(a.Identities, hashIdentity).Equal(set.HashSetFromFunc(b.Identities, hashIdentity))
				}, func(a, b awstypes.ArgoCdRoleMapping) bool {
					return a.Role == b.Role
				})

			input.Configuration.ArgoCd.RbacRoleMappings = &awstypes.UpdateRoleMappings{}
			if addOrUpdate := append(add, update...); len(addOrUpdate) > 0 { //nolint:gocritic // append re-assign is intentional
				input.Configuration.ArgoCd.RbacRoleMappings.AddOrUpdateRoleMappings = addOrUpdate
			}
			if len(remove) > 0 {
				input.Configuration.ArgoCd.RbacRoleMappings.RemoveRoleMappings = remove
			}
		}

		output, err := conn.UpdateCapability(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating EKS Capability (%s,%s)", clusterName, capabilityName), err.Error())

			return
		}

		updateID := aws.ToString(output.Update.Id)
		if _, err := waitCapabilityUpdateSuccessful(ctx, conn, clusterName, capabilityName, updateID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for EKS Capability (%s,%s) update (%s)", clusterName, capabilityName, updateID), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *capabilityResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data capabilityResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EKSClient(ctx)

	clusterName, capabilityName := fwflex.StringValueFromFramework(ctx, data.ClusterName), fwflex.StringValueFromFramework(ctx, data.CapabilityName)
	input := eks.DeleteCapabilityInput{
		CapabilityName: aws.String(capabilityName),
		ClusterName:    aws.String(clusterName),
	}
	_, err := conn.DeleteCapability(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EKS Capability (%s,%s)", clusterName, capabilityName), err.Error())

		return
	}

	if _, err := waitCapabilityDeleted(ctx, conn, clusterName, capabilityName, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EKS Capability (%s,%s) delete", clusterName, capabilityName), err.Error())

		return
	}
}

func (r *capabilityResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		capability = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, capability, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.State.SetAttribute(ctx, path.Root(names.AttrClusterName), parts[0])
	response.State.SetAttribute(ctx, path.Root("capability_name"), parts[1])
}

func findCapabilityByTwoPartKey(ctx context.Context, conn *eks.Client, clusterName, capabilityName string) (*awstypes.Capability, error) {
	input := eks.DescribeCapabilityInput{
		CapabilityName: aws.String(capabilityName),
		ClusterName:    aws.String(clusterName),
	}

	return findCapability(ctx, conn, &input)
}

func findCapability(ctx context.Context, conn *eks.Client, input *eks.DescribeCapabilityInput) (*awstypes.Capability, error) {
	output, err := conn.DescribeCapability(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Capability == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Capability, nil
}

func findCapabilityUpdateByThreePartKey(ctx context.Context, conn *eks.Client, clusterName, capabilityName, id string) (*awstypes.Update, error) {
	input := eks.DescribeUpdateInput{
		CapabilityName: aws.String(capabilityName),
		Name:           aws.String(clusterName),
		UpdateId:       aws.String(id),
	}

	return findUpdate(ctx, conn, &input)
}

func statusCapability(conn *eks.Client, clusterName, capabilityName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCapabilityByTwoPartKey(ctx, conn, clusterName, capabilityName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusCapabilityUpdate(conn *eks.Client, clusterName, capabilityName, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCapabilityUpdateByThreePartKey(ctx, conn, clusterName, capabilityName, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitCapabilityCreated(ctx context.Context, conn *eks.Client, clusterName, capabilityName string, timeout time.Duration) (*awstypes.Capability, error) {
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapabilityStatusCreating),
		Target:  enum.Slice(awstypes.CapabilityStatusActive),
		Refresh: statusCapability(conn, clusterName, capabilityName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Capability); ok {
		if status, health := output.Status, output.Health; status == awstypes.CapabilityStatusCreateFailed && health != nil {
			retry.SetLastError(err, capabilityIssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitCapabilityDeleted(ctx context.Context, conn *eks.Client, clusterName, capabilityName string, timeout time.Duration) (*awstypes.Capability, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapabilityStatusActive, awstypes.CapabilityStatusDeleting),
		Target:  []string{},
		Refresh: statusCapability(conn, clusterName, capabilityName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Capability); ok {
		if status, health := output.Status, output.Health; status == awstypes.CapabilityStatusDeleteFailed && health != nil {
			retry.SetLastError(err, capabilityIssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitCapabilityUpdateSuccessful(ctx context.Context, conn *eks.Client, clusterName, capabilityName, id string, timeout time.Duration) (*awstypes.Update, error) {
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(awstypes.UpdateStatusInProgress),
		Target:  enum.Slice(awstypes.UpdateStatusSuccessful),
		Refresh: statusCapabilityUpdate(conn, clusterName, capabilityName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Update); ok {
		if status := output.Status; status == awstypes.UpdateStatusCancelled || status == awstypes.UpdateStatusFailed {
			retry.SetLastError(err, errorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func capabilityIssueError(apiObject awstypes.CapabilityIssue) error {
	return fmt.Errorf("%s: %s", apiObject.Code, aws.ToString(apiObject.Message))
}

func capabilityIssuesError(apiObjects []awstypes.CapabilityIssue) error {
	var errs []error

	for _, apiObject := range apiObjects {
		errs = append(errs, capabilityIssueError(apiObject))
	}

	return errors.Join(errs...)
}

type capabilityResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                                   `tfsdk:"arn"`
	CapabilityName          types.String                                                   `tfsdk:"capability_name"`
	ClusterName             types.String                                                   `tfsdk:"cluster_name"`
	Configuration           fwtypes.ListNestedObjectValueOf[capabilityConfigurationModel]  `tfsdk:"configuration"`
	DeletePropagationPolicy fwtypes.StringEnum[awstypes.CapabilityDeletePropagationPolicy] `tfsdk:"delete_propagation_policy"`
	RoleARN                 fwtypes.ARN                                                    `tfsdk:"role_arn"`
	Tags                    tftags.Map                                                     `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                     `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                 `tfsdk:"timeouts"`
	Type                    fwtypes.StringEnum[awstypes.CapabilityType]                    `tfsdk:"type"`
	Version                 types.String                                                   `tfsdk:"version"`
}

type capabilityConfigurationModel struct {
	ArgoCD fwtypes.ListNestedObjectValueOf[argoCDConfigModel] `tfsdk:"argo_cd"`
}

type argoCDConfigModel struct {
	AWSIDC           fwtypes.ListNestedObjectValueOf[argoCDAWSIDCConfigModel]        `tfsdk:"aws_idc"`
	Namespace        types.String                                                    `tfsdk:"namespace"`
	NetworkAccess    fwtypes.ListNestedObjectValueOf[argoCDNetworkAccessConfigModel] `tfsdk:"network_access"`
	RBACRoleMappings fwtypes.SetNestedObjectValueOf[argoCDRoleMappingModel]          `tfsdk:"rbac_role_mapping"`
	ServerURL        types.String                                                    `tfsdk:"server_url"`
}

type argoCDAWSIDCConfigModel struct {
	IDCInstanceARN           fwtypes.ARN  `tfsdk:"idc_instance_arn"`
	IDCManagedApplicationARN fwtypes.ARN  `tfsdk:"idc_managed_application_arn"`
	IDCRegion                types.String `tfsdk:"idc_region"`
}

type argoCDNetworkAccessConfigModel struct {
	VPCEIDs fwtypes.SetOfString `tfsdk:"vpce_ids"`
}

type argoCDRoleMappingModel struct {
	Identities fwtypes.SetNestedObjectValueOf[SSOIdentity] `tfsdk:"identity"`
	Role       fwtypes.StringEnum[awstypes.ArgoCdRole]     `tfsdk:"role"`
}

type SSOIdentity struct {
	ID   types.String                                 `tfsdk:"id"`
	Type fwtypes.StringEnum[awstypes.SsoIdentityType] `tfsdk:"type"`
}
