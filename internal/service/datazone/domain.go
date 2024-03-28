// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Domain")
// @Tags(identifierAttribute="arn")
func newResourceDomain(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDomain{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameDomain           = "Domain"
	CreateNumRetries        = 3
	CreateRetryDelaySeconds = 2
)

type resourceDomain struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDomain) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_datazone_domain"
}

func (r *resourceDomain) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"portal_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_execution_role": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"kms_key_identifier": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"single_sign_on": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.AuthType](),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Default: stringdefault.StaticString("DISABLED"),
						},
						"user_assignment": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.UserAssignment](),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDomain) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan resourceDomainData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.CreateDomainInput{
		Name:                aws.String(plan.Name.ValueString()),
		DomainExecutionRole: aws.String(plan.DomainExecutionRole.ValueString()),
		ClientToken:         aws.String(sdkid.UniqueId()),
		Tags:                getTagsIn(ctx),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.KmsKeyIdentifier.IsNull() {
		in.KmsKeyIdentifier = aws.String(plan.KmsKeyIdentifier.ValueString())
	}

	if !plan.SingleSignOn.IsNull() {
		var tfList []singleSignOnData
		resp.Diagnostics.Append(plan.SingleSignOn.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.SingleSignOn = expandSingleSignOn(tfList)
	}

	out, err := retryCreateDomain(ctx, conn, in, CreateNumRetries)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameDomain, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameDomain, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.Arn)
	plan.ID = flex.StringToFramework(ctx, out.Id)
	plan.PortalUrl = flex.StringToFramework(ctx, out.PortalUrl)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDomainCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForCreation, ResNameDomain, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDomain) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceDomainData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDomainByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameDomain, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.ID = flex.StringToFramework(ctx, out.Id)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.PortalUrl = flex.StringToFramework(ctx, out.PortalUrl)
	state.KmsKeyIdentifier = flex.StringToFrameworkARN(ctx, out.KmsKeyIdentifier)
	state.DomainExecutionRole = flex.StringToFrameworkARN(ctx, out.DomainExecutionRole)

	if out.SingleSignOn.Type == awstypes.AuthType("DISABLED") && state.SingleSignOn.IsNull() {
		// Do not set single sign on in state if it was null and response is DISABLED as this is equivalent
		elemType := fwtypes.NewObjectTypeOf[singleSignOnData](ctx).ObjectType
		state.SingleSignOn = types.ListNull(elemType)
	} else {
		singleSignOn, d := flattenSingleSignOn(ctx, out.SingleSignOn)
		resp.Diagnostics.Append(d...)
		state.SingleSignOn = singleSignOn
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDomain) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state resourceDomainData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.DomainExecutionRole.Equal(state.DomainExecutionRole) ||
		!plan.Name.Equal(state.Name) ||
		!plan.SingleSignOn.Equal(state.SingleSignOn) {
		in := &datazone.UpdateDomainInput{
			Identifier:  aws.String(plan.ID.ValueString()),
			ClientToken: aws.String(sdkid.UniqueId()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}

		if !plan.Name.IsNull() {
			in.Name = aws.String(plan.Name.ValueString())
		}

		if !plan.DomainExecutionRole.IsNull() {
			in.DomainExecutionRole = aws.String(plan.DomainExecutionRole.ValueString())
		}

		if !plan.SingleSignOn.IsNull() {
			var tfList []singleSignOnData
			resp.Diagnostics.Append(plan.SingleSignOn.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.SingleSignOn = expandSingleSignOn(tfList)
		}

		out, err := conn.UpdateDomain(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameDomain, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameDomain, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ID = flex.StringToFramework(ctx, out.Id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDomain) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state resourceDomainData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.DeleteDomainInput{
		Identifier:  aws.String(state.ID.ValueString()),
		ClientToken: aws.String(sdkid.UniqueId()),
	}

	_, err := conn.DeleteDomain(ctx, in)
	if err != nil {
		if isResourceMissing(err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameDomain, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDomainDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForDeletion, ResNameDomain, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDomain) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceDomain) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func retryCreateDomain(ctx context.Context, conn *datazone.Client, in *datazone.CreateDomainInput, retries int) (*datazone.CreateDomainOutput, error) {
	out, err := conn.CreateDomain(ctx, in)

	// Race condition can occur domain execution IAM role is not yet ready to be assumed by datazone
	if retries > 0 && errs.IsA[*awstypes.AccessDeniedException](err) {
		time.Sleep(CreateRetryDelaySeconds * time.Second)
		return retryCreateDomain(ctx, conn, in, retries-1)
	}

	return out, err
}

func waitDomainCreated(ctx context.Context, conn *datazone.Client, id string, timeout time.Duration) (*datazone.GetDomainOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusCreating),
		Target:  enum.Slice(awstypes.DomainStatusAvailable),
		Refresh: statusDomain(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetDomainOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDomainDeleted(ctx context.Context, conn *datazone.Client, id string, timeout time.Duration) (*datazone.GetDomainOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusAvailable, awstypes.DomainStatusDeleting),
		Target:  []string{},
		Refresh: statusDomain(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetDomainOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDomain(ctx context.Context, conn *datazone.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDomainByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findDomainByID(ctx context.Context, conn *datazone.Client, id string) (*datazone.GetDomainOutput, error) {
	in := &datazone.GetDomainInput{
		Identifier: aws.String(id),
	}

	out, err := conn.GetDomain(ctx, in)
	if err != nil {
		if isResourceMissing(err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func isResourceMissing(err error) bool {
	// DataZone returns a 403 when the domain does not exist
	// AccessDeniedException: User is not permitted to perform operation: GetDomain
	return errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not permitted to perform")
}

func flattenSingleSignOn(ctx context.Context, apiObject *awstypes.SingleSignOn) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: singleSignOnAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"type":            flex.StringValueToFramework(ctx, apiObject.Type),
		"user_assignment": flex.StringValueToFramework(ctx, apiObject.UserAssignment),
	}
	objVal, d := types.ObjectValue(singleSignOnAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func expandSingleSignOn(tfList []singleSignOnData) *awstypes.SingleSignOn {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.SingleSignOn{}

	if !tfObj.Type.IsNull() {
		apiObject.Type = awstypes.AuthType(tfObj.Type.ValueString())
	}

	if !tfObj.UserAssignment.IsNull() {
		apiObject.UserAssignment = awstypes.UserAssignment(tfObj.UserAssignment.ValueString())
	}

	return apiObject
}

type resourceDomainData struct {
	ARN                 types.String   `tfsdk:"arn"`
	SingleSignOn        types.List     `tfsdk:"single_sign_on"`
	DomainExecutionRole fwtypes.ARN    `tfsdk:"domain_execution_role"`
	KmsKeyIdentifier    fwtypes.ARN    `tfsdk:"kms_key_identifier"`
	Description         types.String   `tfsdk:"description"`
	PortalUrl           types.String   `tfsdk:"portal_url"`
	ID                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
	Tags                types.Map      `tfsdk:"tags"`
	TagsAll             types.Map      `tfsdk:"tags_all"`
}

type singleSignOnData struct {
	Type           types.String `tfsdk:"type"`
	UserAssignment types.String `tfsdk:"user_assignment"`
}

var singleSignOnAttrTypes = map[string]attr.Type{
	"type":            types.StringType,
	"user_assignment": types.StringType,
}
