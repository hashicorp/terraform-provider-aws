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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_customer_managed_policy_attachments_exclusive", name="Customer Managed Policy Attachments Exclusive")
// @IdentityAttribute("instance_arn")
// @IdentityAttribute("permission_set_arn")
// @ArnFormat(global=true)
// @ImportIDHandler("customerManagedPolicyAttachmentsExclusiveImportID")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstances")
// @Testing(hasNoPreExistingResource=true)
// @Testing(checkDestroyNoop=true)
// @Testing(importStateIdAttributes="instance_arn;permission_set_arn", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newCustomerManagedPolicyAttachmentsExclusiveResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &customerManagedPolicyAttachmentsExclusiveResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)

	return r, nil
}

type customerManagedPolicyAttachmentsExclusiveResource struct {
	framework.ResourceWithModel[customerManagedPolicyAttachmentsExclusiveResourceModel]
	framework.WithNoOpDelete
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *customerManagedPolicyAttachmentsExclusiveResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
		},
		Blocks: map[string]schema.Block{
			"customer_managed_policy_reference": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customerManagedPolicyReferenceModel](ctx),
				Validators: []validator.List{
					listvalidator.NoNullValues(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrPath: schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("/"),
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *customerManagedPolicyAttachmentsExclusiveResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customerManagedPolicyAttachmentsExclusiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var policies []customerManagedPolicyReferenceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, plan.CustomerManagedPolicyReferences.ElementsAs(ctx, &policies, false))
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.syncAttachments(ctx, plan.InstanceARN.ValueString(), plan.PermissionSetARN.ValueString(), policies, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PermissionSetARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *customerManagedPolicyAttachmentsExclusiveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state customerManagedPolicyAttachmentsExclusiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCustomerManagedPolicyAttachmentsByTwoPartKey(ctx, conn, state.PermissionSetARN.ValueString(), state.InstanceARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.PermissionSetARN.ValueString())
		return
	}

	state.CustomerManagedPolicyReferences = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, flattenCustomerManagedPolicyReferences(out))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *customerManagedPolicyAttachmentsExclusiveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state customerManagedPolicyAttachmentsExclusiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.CustomerManagedPolicyReferences.Equal(state.CustomerManagedPolicyReferences) {
		var policies []customerManagedPolicyReferenceModel
		smerr.AddEnrich(ctx, &resp.Diagnostics, plan.CustomerManagedPolicyReferences.ElementsAs(ctx, &policies, false))
		if resp.Diagnostics.HasError() {
			return
		}

		err := r.syncAttachments(ctx, plan.InstanceARN.ValueString(), plan.PermissionSetARN.ValueString(), policies, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PermissionSetARN.ValueString())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

var _ inttypes.ImportIDParser = customerManagedPolicyAttachmentsExclusiveImportID{}

type customerManagedPolicyAttachmentsExclusiveImportID struct{}

func (customerManagedPolicyAttachmentsExclusiveImportID) Parse(id string) (string, map[string]any, error) {
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

func (r *customerManagedPolicyAttachmentsExclusiveResource) syncAttachments(ctx context.Context, instanceARN, permissionSetARN string, want []customerManagedPolicyReferenceModel, timeout time.Duration) error {
	conn := r.Meta().SSOAdminClient(ctx)

	have, err := findCustomerManagedPolicyAttachmentsByTwoPartKey(ctx, conn, permissionSetARN, instanceARN)
	if err != nil {
		return smarterr.NewError(err)
	}

	wantAPI := make([]awstypes.CustomerManagedPolicyReference, len(want))
	for i, w := range want {
		wantAPI[i] = awstypes.CustomerManagedPolicyReference{
			Name: w.Name.ValueStringPointer(),
			Path: w.Path.ValueStringPointer(),
		}
	}

	create, remove, _ := intflex.DiffSlices(have, wantAPI, func(h, w awstypes.CustomerManagedPolicyReference) bool {
		return aws.ToString(h.Name) == aws.ToString(w.Name) && aws.ToString(h.Path) == aws.ToString(w.Path)
	})

	for _, policy := range create {
		input := &ssoadmin.AttachCustomerManagedPolicyReferenceToPermissionSetInput{
			InstanceArn:      aws.String(instanceARN),
			PermissionSetArn: aws.String(permissionSetARN),
			CustomerManagedPolicyReference: &awstypes.CustomerManagedPolicyReference{
				Name: policy.Name,
				Path: policy.Path,
			},
		}
		_, err := conn.AttachCustomerManagedPolicyReferenceToPermissionSet(ctx, input)
		if err != nil {
			return smarterr.NewError(err)
		}
	}

	for _, policy := range remove {
		input := &ssoadmin.DetachCustomerManagedPolicyReferenceFromPermissionSetInput{
			InstanceArn:      aws.String(instanceARN),
			PermissionSetArn: aws.String(permissionSetARN),
			CustomerManagedPolicyReference: &awstypes.CustomerManagedPolicyReference{
				Name: policy.Name,
				Path: policy.Path,
			},
		}
		_, err := conn.DetachCustomerManagedPolicyReferenceFromPermissionSet(ctx, input)
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

type customerManagedPolicyAttachmentsExclusiveResourceModel struct {
	framework.WithRegionModel
	InstanceARN                     types.String                                                         `tfsdk:"instance_arn"`
	PermissionSetARN                types.String                                                         `tfsdk:"permission_set_arn"`
	CustomerManagedPolicyReferences fwtypes.ListNestedObjectValueOf[customerManagedPolicyReferenceModel] `tfsdk:"customer_managed_policy_reference"`
	Timeouts                        timeouts.Value                                                       `tfsdk:"timeouts"`
}

type customerManagedPolicyReferenceModel struct {
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`
}

func findCustomerManagedPolicyAttachmentsByTwoPartKey(ctx context.Context, conn *ssoadmin.Client, permissionSetARN, instanceARN string) ([]awstypes.CustomerManagedPolicyReference, error) {
	input := &ssoadmin.ListCustomerManagedPolicyReferencesInPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}

	var policies []awstypes.CustomerManagedPolicyReference
	paginator := ssoadmin.NewListCustomerManagedPolicyReferencesInPermissionSetPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil, smarterr.NewError(&retry.NotFoundError{
					LastError: err,
				})
			}
			return nil, smarterr.NewError(err)
		}

		policies = append(policies, page.CustomerManagedPolicyReferences...)
	}

	return policies, nil
}

func flattenCustomerManagedPolicyReferences(apiObjects []awstypes.CustomerManagedPolicyReference) []customerManagedPolicyReferenceModel {
	var models []customerManagedPolicyReferenceModel
	for _, apiObject := range apiObjects {
		models = append(models, customerManagedPolicyReferenceModel{
			Name: types.StringValue(aws.ToString(apiObject.Name)),
			Path: types.StringValue(aws.ToString(apiObject.Path)),
		})
	}
	return models
}
