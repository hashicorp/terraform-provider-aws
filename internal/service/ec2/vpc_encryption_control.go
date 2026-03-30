// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_encryption_control", name="VPC Encryption Control")
// @Tags(identifierAttribute="id")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.VpcEncryptionControl")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importIgnore="egress_only_internet_gateway_exclusion;elastic_file_system_exclusion;internet_gateway_exclusion;lambda_exclusion;nat_gateway_exclusion;virtual_private_gateway_exclusion;vpc_lattice_exclusion;vpc_peering_exclusion")
// @Testing(generator=false)
func newResourceVPCEncryptionControl(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCEncryptionControl{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameVPCEncryptionControl = "VPC Encryption Control"
)

type resourceVPCEncryptionControl struct {
	framework.ResourceWithModel[resourceVPCEncryptionControlModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

var (
	vpcEncryptionControlExclusionStateInputType = fwtypes.StringEnumType[awstypes.VpcEncryptionControlExclusionStateInput]()

	vpcEncryptionControlExclusionStateInputAttribute = schema.StringAttribute{
		CustomType: vpcEncryptionControlExclusionStateInputType,
		Optional:   true,
		Computed:   true,
		Default:    vpcEncryptionControlExclusionStateInputType.AttributeDefault(awstypes.VpcEncryptionControlExclusionStateInputDisable),
	}
)

func (r *resourceVPCEncryptionControl) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"egress_only_internet_gateway_exclusion": vpcEncryptionControlExclusionStateInputAttribute,
			"elastic_file_system_exclusion":          vpcEncryptionControlExclusionStateInputAttribute,
			names.AttrID:                             framework.IDAttribute(),
			"internet_gateway_exclusion":             vpcEncryptionControlExclusionStateInputAttribute,
			"lambda_exclusion":                       vpcEncryptionControlExclusionStateInputAttribute,
			names.AttrMode: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VpcEncryptionControlMode](),
				Required:   true,
			},
			"nat_gateway_exclusion": vpcEncryptionControlExclusionStateInputAttribute,
			"resource_exclusions": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[vpcEncryptionControlExclusionsModel](ctx),
				Computed:   true,
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			"state_message": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:                      tftags.TagsAttribute(),
			names.AttrTagsAll:                   tftags.TagsAttributeComputedOnly(),
			"virtual_private_gateway_exclusion": vpcEncryptionControlExclusionStateInputAttribute,
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
			},
			"vpc_lattice_exclusion": vpcEncryptionControlExclusionStateInputAttribute,
			"vpc_peering_exclusion": vpcEncryptionControlExclusionStateInputAttribute,
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceVPCEncryptionControl) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCEncryptionControlModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.CreateVpcEncryptionControlInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpcEncryptionControl)

	out, err := conn.CreateVpcEncryptionControl(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrVPCID, plan.VPCID.String())
		return
	}
	if out == nil || out.VpcEncryptionControl == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), names.AttrVPCID, plan.VPCID.String())
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.VpcEncryptionControl.VpcEncryptionControlId)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrID), plan.ID))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	ec, err := waitVPCEncryptionControlAvailable(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrVPCID, plan.VPCID.String())
		return
	}

	if plan.Mode.ValueEnum() == awstypes.VpcEncryptionControlModeEnforce {
		ec = vpcEncryptionControlModify(ctx, conn, plan, createTimeout, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, ec, &plan, flex.WithFieldNamePrefix("VpcEncryptionControl")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceVPCEncryptionControl) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCEncryptionControlModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCEncryptionControlByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	flattenExclusionInputs(ctx, out, &state)

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceVPCEncryptionControl) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan, state resourceVPCEncryptionControlModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	var ec *awstypes.VpcEncryptionControl
	if diff.HasChanges() {
		updateTimeout := r.UpdateTimeout(ctx, state.Timeouts)
		ec = vpcEncryptionControlModify(ctx, conn, plan, updateTimeout, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		var err error
		ec, err = findVPCEncryptionControlByID(ctx, conn, state.ID.ValueString())
		if retry.NotFound(err) {
			resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			resp.State.RemoveResource(ctx)
			return
		}
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, ec, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceVPCEncryptionControl) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCEncryptionControlModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DeleteVpcEncryptionControlInput{
		VpcEncryptionControlId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteVpcEncryptionControl(ctx, &input)
	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEncryptionControlIdNotFound) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCEncryptionControlDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func waitVPCEncryptionControlAvailable(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcEncryptionControl, error) {
	stateConf := &retry.StateChangeConfOf[*awstypes.VpcEncryptionControl, awstypes.VpcEncryptionControlState]{
		Pending: enum.EnumSlice(awstypes.VpcEncryptionControlStateCreating, awstypes.VpcEncryptionControlStateEnforceInProgress, awstypes.VpcEncryptionControlStateMonitorInProgress),
		Target:  enum.EnumSlice(awstypes.VpcEncryptionControlStateAvailable),
		Refresh: statusVPCEncryptionControl(conn, id),
		Timeout: timeout,
	}

	return wrapError(stateConf.WaitForStateContext(ctx))
}

func waitVPCEncryptionControlDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcEncryptionControl, error) {
	stateConf := &retry.StateChangeConfOf[*awstypes.VpcEncryptionControl, awstypes.VpcEncryptionControlState]{
		Pending: enum.EnumSlice(awstypes.VpcEncryptionControlStateDeleting),
		Target:  []awstypes.VpcEncryptionControlState{},
		Refresh: statusVPCEncryptionControl(conn, id),
		Timeout: timeout,
	}

	return wrapError(stateConf.WaitForStateContext(ctx))
}

func wrapError[T any](v T, err error) (T, error) {
	return v, smarterr.NewError(err)
}

func waitVPCEncryptionControlExclusionsApplied(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcEncryptionControl, error) {
	var output *awstypes.VpcEncryptionControl
	err := tfresource.WaitUntil(ctx, timeout, func(ctx context.Context) (bool, error) {
		var err error
		output, err = findVPCEncryptionControlByID(ctx, conn, id)
		if retry.NotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		exclusions := output.ResourceExclusions
		if exclusions == nil {
			return true, nil
		}

		if isExclusionPending(exclusions.EgressOnlyInternetGateway) {
			return false, nil
		}
		if isExclusionPending(exclusions.ElasticFileSystem) {
			return false, nil
		}
		if isExclusionPending(exclusions.InternetGateway) {
			return false, nil
		}
		if isExclusionPending(exclusions.Lambda) {
			return false, nil
		}
		if isExclusionPending(exclusions.NatGateway) {
			return false, nil
		}
		if isExclusionPending(exclusions.VirtualPrivateGateway) {
			return false, nil
		}
		if isExclusionPending(exclusions.VpcLattice) {
			return false, nil
		}
		if isExclusionPending(exclusions.VpcPeering) {
			return false, nil
		}

		return true, nil
	}, tfresource.WaitOpts{})

	if err != nil {
		return nil, smarterr.NewError(err)
	}
	return output, nil
}

// Returns `true` if still in progress
func isExclusionPending(apiObject *awstypes.VpcEncryptionControlExclusion) bool {
	if apiObject == nil {
		return false
	}
	return apiObject.State == awstypes.VpcEncryptionControlExclusionStateEnabling || apiObject.State == awstypes.VpcEncryptionControlExclusionStateDisabling
}

func waitVPCEncryptionControlMigratableResourcesMigrated(ctx context.Context, conn *ec2.Client, vpcID string, timeout time.Duration) error {
	err := tfresource.WaitUntil(ctx, timeout, func(context.Context) (bool, error) {
		blockers, err := encryptionEnforcementBlockers(ctx, conn, vpcID)
		if err != nil {
			return false, err
		}

		if slices.ContainsFunc(blockers, isResourceMigratable) {
			return false, nil
		}

		return true, nil
	}, tfresource.WaitOpts{})

	if err != nil {
		return smarterr.NewError(err)
	}
	return nil
}

func isResourceMigratable(v awstypes.VpcEncryptionNonCompliantResource) bool {
	if aws.ToString(v.Type) == "interface" {
		if strings.HasPrefix(aws.ToString(v.Description), "ELB app/") {
			return true
		}
		if strings.HasPrefix(aws.ToString(v.Description), "ELB net/") {
			return true
		}
	}
	return false
}

func statusVPCEncryptionControl(conn *ec2.Client, id string) retry.StateRefreshFuncOf[*awstypes.VpcEncryptionControl, awstypes.VpcEncryptionControlState] {
	return func(ctx context.Context) (*awstypes.VpcEncryptionControl, awstypes.VpcEncryptionControlState, error) {
		encryptionControl, err := findVPCEncryptionControlByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return encryptionControl, encryptionControl.State, nil
	}
}

func findVPCEncryptionControlByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcEncryptionControl, error) {
	output, err := findVPCEncryptionControlsByIDs(ctx, conn, []string{id})
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEncryptionControlsByIDs(ctx context.Context, conn *ec2.Client, ids []string) ([]awstypes.VpcEncryptionControl, error) {
	input := ec2.DescribeVpcEncryptionControlsInput{
		VpcEncryptionControlIds: ids,
	}

	var output []awstypes.VpcEncryptionControl

	pages := newDescribeVpcEncryptionControlsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEncryptionControlIdNotFound) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.VpcEncryptionControls...)
	}

	return output, nil
}

func vpcEncryptionControlModify(ctx context.Context, conn *ec2.Client, plan resourceVPCEncryptionControlModel, timeout time.Duration, diags *diag.Diagnostics) *awstypes.VpcEncryptionControl {
	err := waitVPCEncryptionControlMigratableResourcesMigrated(ctx, conn, plan.VPCID.ValueString(), timeout)
	if err != nil {
		smerr.AddError(ctx, diags, err, names.AttrVPCID, plan.VPCID.String())
		return nil
	}

	var modifyInput ec2.ModifyVpcEncryptionControlInput
	expandForModify(ctx, plan, &modifyInput)

	out, err := conn.ModifyVpcEncryptionControl(ctx, &modifyInput)
	if tfawserr.ErrCodeEquals(err, "VpcEncryptionControlStateTransitionFailed") {
		blockers, _ := encryptionEnforcementBlockers(ctx, conn, plan.VPCID.ValueString())
		if len(blockers) > 0 {
			var buf strings.Builder
			for _, ncr := range blockers {
				fmt.Fprintf(&buf, "* Type: %q, Id: %q, Description: %q, IsExcludable: %t\n",
					aws.ToString(ncr.Type),
					aws.ToString(ncr.Id),
					aws.ToString(ncr.Description),
					aws.ToBool(ncr.IsExcludable),
				)
			}
			smerr.AddError(ctx, diags, fmt.Errorf("The following resources prevented enforcement:\n\n%s", buf.String()), names.AttrVPCID, plan.VPCID.String())
			return nil
		}
	}
	if err != nil {
		smerr.AddError(ctx, diags, err, names.AttrVPCID, plan.VPCID.String())
		return nil
	}
	if out == nil || out.VpcEncryptionControl == nil {
		smerr.AddError(ctx, diags, errors.New("empty output"), names.AttrVPCID, plan.VPCID.String())
		return nil
	}

	ec, err := waitVPCEncryptionControlAvailable(ctx, conn, plan.ID.ValueString(), timeout)
	if use, ok := errs.As[*retry.UnexpectedStateError](err); ok && use.State == string(awstypes.VpcEncryptionControlStateEnforceFailed) {
		// Ignore errors here and fall through to the outer error
		blockers, _ := encryptionEnforcementBlockers(ctx, conn, plan.VPCID.ValueString())
		if len(blockers) > 0 {
			var buf strings.Builder
			for _, ncr := range blockers {
				fmt.Fprintf(&buf, "* Type: %q, Id: %q, Description: %q, IsExcludable: %t\n",
					aws.ToString(ncr.Type),
					aws.ToString(ncr.Id),
					aws.ToString(ncr.Description),
					aws.ToBool(ncr.IsExcludable),
				)
			}
			smerr.AddError(ctx, diags, fmt.Errorf("The following resources prevented enforcement:\n\n%s", buf.String()), names.AttrVPCID, plan.VPCID.String())
			return nil
		}
	}
	if err != nil {
		smerr.AddError(ctx, diags, err, names.AttrVPCID, plan.VPCID.String())
		return nil
	}

	if ec.ResourceExclusions != nil {
		ec, err = waitVPCEncryptionControlExclusionsApplied(ctx, conn, plan.ID.ValueString(), timeout)
		if err != nil {
			smerr.AddError(ctx, diags, err, names.AttrVPCID, plan.VPCID.String())
			return nil
		}
	}

	return ec
}

func encryptionEnforcementBlockers(ctx context.Context, conn *ec2.Client, vpcId string) ([]awstypes.VpcEncryptionNonCompliantResource, error) {
	var blockers []awstypes.VpcEncryptionNonCompliantResource

	blockersInput := ec2.GetVpcResourcesBlockingEncryptionEnforcementInput{
		VpcId: aws.String(vpcId),
	}
	pages := newGetVpcResourcesBlockingEncryptionEnforcementPaginator(conn, &blockersInput)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		blockers = append(blockers, page.NonCompliantResources...)
	}

	return blockers, nil
}

func expandForModify(_ context.Context, plan resourceVPCEncryptionControlModel, apiObject *ec2.ModifyVpcEncryptionControlInput) {
	apiObject.VpcEncryptionControlId = plan.ID.ValueStringPointer()
	apiObject.Mode = plan.Mode.ValueEnum()

	if plan.Mode.ValueEnum() == awstypes.VpcEncryptionControlModeEnforce {
		apiObject.EgressOnlyInternetGatewayExclusion = plan.EgressOnlyInternetGatewayExclusion.ValueEnum()
		apiObject.ElasticFileSystemExclusion = plan.ElasticFileSystemExclusion.ValueEnum()
		apiObject.InternetGatewayExclusion = plan.InternetGatewayExclusion.ValueEnum()
		apiObject.LambdaExclusion = plan.LambdaExclusion.ValueEnum()
		apiObject.NatGatewayExclusion = plan.NatGatewayExclusion.ValueEnum()
		apiObject.VirtualPrivateGatewayExclusion = plan.VirtualPrivateGatewayExclusion.ValueEnum()
		apiObject.VpcLatticeExclusion = plan.VpcLatticeExclusion.ValueEnum()
		apiObject.VpcPeeringExclusion = plan.VpcPeeringExclusion.ValueEnum()
	}
}

// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenExclusionInputs(ctx context.Context, apiObject *awstypes.VpcEncryptionControl, tfObject *resourceVPCEncryptionControlModel) {
	exclusions := apiObject.ResourceExclusions
	if exclusions == nil {
		return
	}

	flattenExclusionInput(ctx, exclusions.EgressOnlyInternetGateway, &tfObject.EgressOnlyInternetGatewayExclusion)

	flattenExclusionInput(ctx, exclusions.ElasticFileSystem, &tfObject.ElasticFileSystemExclusion)

	flattenExclusionInput(ctx, exclusions.InternetGateway, &tfObject.InternetGatewayExclusion)

	flattenExclusionInput(ctx, exclusions.Lambda, &tfObject.LambdaExclusion)

	flattenExclusionInput(ctx, exclusions.NatGateway, &tfObject.NatGatewayExclusion)

	flattenExclusionInput(ctx, exclusions.VirtualPrivateGateway, &tfObject.VirtualPrivateGatewayExclusion)

	flattenExclusionInput(ctx, exclusions.VpcLattice, &tfObject.VpcLatticeExclusion)

	flattenExclusionInput(ctx, exclusions.VpcPeering, &tfObject.VpcPeeringExclusion)
}

func flattenExclusionInput(_ context.Context, apiObject *awstypes.VpcEncryptionControlExclusion, tfObject *fwtypes.StringEnum[awstypes.VpcEncryptionControlExclusionStateInput]) {
	if apiObject == nil {
		return
	}

	switch apiObject.State {
	case awstypes.VpcEncryptionControlExclusionStateEnabled:
		*tfObject = fwtypes.StringEnumValue(awstypes.VpcEncryptionControlExclusionStateInputEnable)

	case awstypes.VpcEncryptionControlExclusionStateDisabled:
		*tfObject = fwtypes.StringEnumValue(awstypes.VpcEncryptionControlExclusionStateInputDisable)
	}
}

type resourceVPCEncryptionControlModel struct {
	framework.WithRegionModel
	EgressOnlyInternetGatewayExclusion fwtypes.StringEnum[awstypes.VpcEncryptionControlExclusionStateInput] `tfsdk:"egress_only_internet_gateway_exclusion"`
	ElasticFileSystemExclusion         fwtypes.StringEnum[awstypes.VpcEncryptionControlExclusionStateInput] `tfsdk:"elastic_file_system_exclusion"`
	ID                                 types.String                                                         `tfsdk:"id"`
	InternetGatewayExclusion           fwtypes.StringEnum[awstypes.VpcEncryptionControlExclusionStateInput] `tfsdk:"internet_gateway_exclusion"`
	LambdaExclusion                    fwtypes.StringEnum[awstypes.VpcEncryptionControlExclusionStateInput] `tfsdk:"lambda_exclusion"`
	Mode                               fwtypes.StringEnum[awstypes.VpcEncryptionControlMode]                `tfsdk:"mode"`
	NatGatewayExclusion                fwtypes.StringEnum[awstypes.VpcEncryptionControlExclusionStateInput] `tfsdk:"nat_gateway_exclusion"`
	ResourceExclusions                 fwtypes.ObjectValueOf[vpcEncryptionControlExclusionsModel]           `tfsdk:"resource_exclusions"`
	State                              types.String                                                         `tfsdk:"state"`
	StateMessage                       types.String                                                         `tfsdk:"state_message"`
	Tags                               tftags.Map                                                           `tfsdk:"tags"`
	TagsAll                            tftags.Map                                                           `tfsdk:"tags_all"`
	Timeouts                           timeouts.Value                                                       `tfsdk:"timeouts"`
	VirtualPrivateGatewayExclusion     fwtypes.StringEnum[awstypes.VpcEncryptionControlExclusionStateInput] `tfsdk:"virtual_private_gateway_exclusion"`
	VPCID                              types.String                                                         `tfsdk:"vpc_id"`
	VpcLatticeExclusion                fwtypes.StringEnum[awstypes.VpcEncryptionControlExclusionStateInput] `tfsdk:"vpc_lattice_exclusion"`
	VpcPeeringExclusion                fwtypes.StringEnum[awstypes.VpcEncryptionControlExclusionStateInput] `tfsdk:"vpc_peering_exclusion"`
}

type vpcEncryptionControlExclusionsModel struct {
	EgressOnlyInternetGateway fwtypes.ObjectValueOf[vpcEncryptionControlExclusionModel] `tfsdk:"egress_only_internet_gateway"`
	ElasticFileSystem         fwtypes.ObjectValueOf[vpcEncryptionControlExclusionModel] `tfsdk:"elastic_file_system"`
	InternetGateway           fwtypes.ObjectValueOf[vpcEncryptionControlExclusionModel] `tfsdk:"internet_gateway"`
	Lambda                    fwtypes.ObjectValueOf[vpcEncryptionControlExclusionModel] `tfsdk:"lambda"`
	NatGateway                fwtypes.ObjectValueOf[vpcEncryptionControlExclusionModel] `tfsdk:"nat_gateway"`
	VirtualPrivateGateway     fwtypes.ObjectValueOf[vpcEncryptionControlExclusionModel] `tfsdk:"virtual_private_gateway"`
	VpcLattice                fwtypes.ObjectValueOf[vpcEncryptionControlExclusionModel] `tfsdk:"vpc_lattice"`
	VpcPeering                fwtypes.ObjectValueOf[vpcEncryptionControlExclusionModel] `tfsdk:"vpc_peering"`
}

type vpcEncryptionControlExclusionModel struct {
	State        types.String `tfsdk:"state"`
	StateMessage types.String `tfsdk:"state_message"`
}
