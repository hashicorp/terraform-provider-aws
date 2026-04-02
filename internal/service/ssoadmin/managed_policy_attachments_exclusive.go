// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_managed_policy_attachments_exclusive", name="Managed Policy Attachments Exclusive")
// @IdentityAttribute("instance_arn")
// @IdentityAttribute("permission_set_arn")
// @ArnFormat(global=true)
// @ImportIDHandler("managedPolicyAttachmentsExclusiveImportID")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstances")
// @Testing(hasNoPreExistingResource=true)
// @Testing(checkDestroyNoop=true)
// @Testing(importStateIdAttributes="instance_arn;permission_set_arn", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newManagedPolicyAttachmentsExclusiveResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &managedPolicyAttachmentsExclusiveResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)

	return r, nil
}

type managedPolicyAttachmentsExclusiveResource struct {
	framework.ResourceWithModel[managedPolicyAttachmentsExclusiveResourceModel]
	framework.WithNoOpDelete
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *managedPolicyAttachmentsExclusiveResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permission_set_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"managed_policy_arns": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.NoNullValues(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *managedPolicyAttachmentsExclusiveResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan managedPolicyAttachmentsExclusiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var policyARNs []string
	smerr.AddEnrich(ctx, &resp.Diagnostics, plan.ManagedPolicyARNs.ElementsAs(ctx, &policyARNs, false))
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.syncAttachments(ctx, plan.InstanceARN.ValueString(), plan.PermissionSetARN.ValueString(), policyARNs, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PermissionSetARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *managedPolicyAttachmentsExclusiveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state managedPolicyAttachmentsExclusiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findManagedPolicyAttachmentsByTwoPartKey(ctx, conn, state.PermissionSetARN.ValueString(), state.InstanceARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.PermissionSetARN.ValueString())
		return
	}

	state.ManagedPolicyARNs = flex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, out)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *managedPolicyAttachmentsExclusiveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state managedPolicyAttachmentsExclusiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.ManagedPolicyARNs.Equal(state.ManagedPolicyARNs) {
		var policyARNs []string
		smerr.AddEnrich(ctx, &resp.Diagnostics, plan.ManagedPolicyARNs.ElementsAs(ctx, &policyARNs, false))
		if resp.Diagnostics.HasError() {
			return
		}

		err := r.syncAttachments(ctx, plan.InstanceARN.ValueString(), plan.PermissionSetARN.ValueString(), policyARNs, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PermissionSetARN.ValueString())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

var _ inttypes.ImportIDParser = managedPolicyAttachmentsExclusiveImportID{}

type managedPolicyAttachmentsExclusiveImportID struct{}

func (managedPolicyAttachmentsExclusiveImportID) Parse(id string) (string, map[string]any, error) {
	instanceARN, permissionSetARN, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, smarterr.NewError(fmt.Errorf("id \"%s\" should be in the format <instance-arn>"+intflex.ResourceIdSeparator+"<permission-set-arn>", id))
	}

	result := map[string]any{
		"instance_arn":       instanceARN,
		"permission_set_arn": permissionSetARN,
	}

	return id, result, nil
}

func (r *managedPolicyAttachmentsExclusiveResource) syncAttachments(ctx context.Context, instanceARN, permissionSetARN string, want []string, timeout time.Duration) error {
	conn := r.Meta().SSOAdminClient(ctx)

	have, err := findManagedPolicyAttachmentsByTwoPartKey(ctx, conn, permissionSetARN, instanceARN)
	if err != nil {
		return smarterr.NewError(err)
	}

	create, remove, _ := intflex.DiffSlices(have, want, func(s1, s2 string) bool { return s1 == s2 })

	for _, arn := range create {
		input := &ssoadmin.AttachManagedPolicyToPermissionSetInput{
			InstanceArn:      aws.String(instanceARN),
			ManagedPolicyArn: aws.String(arn),
			PermissionSetArn: aws.String(permissionSetARN),
		}
		_, err := conn.AttachManagedPolicyToPermissionSet(ctx, input)
		if err != nil {
			return smarterr.NewError(err)
		}
	}

	for _, arn := range remove {
		input := &ssoadmin.DetachManagedPolicyFromPermissionSetInput{
			InstanceArn:      aws.String(instanceARN),
			ManagedPolicyArn: aws.String(arn),
			PermissionSetArn: aws.String(permissionSetARN),
		}
		_, err := conn.DetachManagedPolicyFromPermissionSet(ctx, input)
		if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return smarterr.NewError(err)
		}
	}

	if len(create) > 0 || len(remove) > 0 {
		if err := provisionPermissionSet(ctx, conn, permissionSetARN, instanceARN, timeout); err != nil {
			return smarterr.NewError(err)
		}
	}

	return nil
}

type managedPolicyAttachmentsExclusiveResourceModel struct {
	framework.WithRegionModel
	InstanceARN       types.String        `tfsdk:"instance_arn"`
	PermissionSetARN  types.String        `tfsdk:"permission_set_arn"`
	ManagedPolicyARNs fwtypes.SetOfString `tfsdk:"managed_policy_arns"`
	Timeouts          timeouts.Value      `tfsdk:"timeouts"`
}

func findManagedPolicyAttachmentsByTwoPartKey(ctx context.Context, conn *ssoadmin.Client, permissionSetARN, instanceARN string) ([]string, error) {
	input := &ssoadmin.ListManagedPoliciesInPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}

	var policyARNs []string
	paginator := ssoadmin.NewListManagedPoliciesInPermissionSetPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil, smarterr.NewError(&retry.NotFoundError{
					LastError: err,
				})
			}
			return policyARNs, smarterr.NewError(err)
		}

		for _, p := range page.AttachedManagedPolicies {
			if p.Arn != nil {
				policyARNs = append(policyARNs, aws.ToString(p.Arn))
			}
		}
	}

	return policyARNs, nil
}
