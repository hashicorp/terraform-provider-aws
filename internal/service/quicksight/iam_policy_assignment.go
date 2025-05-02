// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_iam_policy_assignment", name="IAM Policy Assignment")
func newIAMPolicyAssignmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &iamPolicyAssignmentResource{}, nil
}

const (
	resNameIAMPolicyAssignment = "IAM Policy Assignment"

	defaultIAMPolicyAssignmentNamespace = "default"
	identitiesUserKey                   = "user"
	identitiesGroupKey                  = "group"
)

type iamPolicyAssignmentResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *iamPolicyAssignmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"assignment_id": schema.StringAttribute{
				Computed: true,
			},
			"assignment_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assignment_status": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.AssignmentStatus](),
				},
			},
			names.AttrAWSAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrNamespace: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(defaultIAMPolicyAssignmentNamespace),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_arn": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"identities": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"user": schema.SetAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"group": schema.SetAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *iamPolicyAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan resourceIAMPolicyAssignmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID(ctx))
	}
	awsAccountID, namespace, assignmentName := flex.StringValueFromFramework(ctx, plan.AWSAccountID), flex.StringValueFromFramework(ctx, plan.Namespace), flex.StringValueFromFramework(ctx, plan.AssignmentName)
	in := quicksight.CreateIAMPolicyAssignmentInput{
		AssignmentName:   aws.String(assignmentName),
		AwsAccountId:     aws.String(awsAccountID),
		Namespace:        aws.String(namespace),
		AssignmentStatus: awstypes.AssignmentStatus(plan.AssignmentStatus.ValueString()),
	}

	if !plan.Identities.IsNull() {
		var identities []identitiesData
		resp.Diagnostics.Append(plan.Identities.ElementsAs(ctx, &identities, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Identities = expandIdentities(ctx, identities)
	}
	if !plan.PolicyARN.IsNull() {
		in.PolicyArn = plan.PolicyARN.ValueStringPointer()
	}

	out, err := conn.CreateIAMPolicyAssignment(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameIAMPolicyAssignment, plan.AssignmentName.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameIAMPolicyAssignment, plan.AssignmentName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, iamPolicyAssignmentCreateResourceID(awsAccountID, namespace, assignmentName))
	plan.AssignmentID = flex.StringToFramework(ctx, out.AssignmentId)

	// wait for IAM to propagate before returning
	_, err = tfresource.RetryWhenNotFound(ctx, iamPropagationTimeout, func() (any, error) {
		return findIAMPolicyAssignmentByThreePartKey(ctx, conn, awsAccountID, namespace, assignmentName)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameIAMPolicyAssignment, plan.AssignmentName.String(), nil),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *iamPolicyAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceIAMPolicyAssignmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, namespace, assignmentName, err := iamPolicyAssignmentParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameIAMPolicyAssignment, state.ID.String(), nil),
			err.Error(),
		)
	}

	out, err := findIAMPolicyAssignmentByThreePartKey(ctx, conn, awsAccountID, namespace, assignmentName)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, resNameIAMPolicyAssignment, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.AssignmentID = flex.StringToFramework(ctx, out.AssignmentId)
	state.AssignmentName = flex.StringToFramework(ctx, out.AssignmentName)
	state.AssignmentStatus = flex.StringValueToFramework(ctx, out.AssignmentStatus)
	state.AWSAccountID = flex.StringToFramework(ctx, out.AwsAccountId)
	identities, d := flattenIdentities(ctx, out.Identities)
	resp.Diagnostics.Append(d...)
	state.Identities = identities
	state.PolicyARN = flex.StringToFramework(ctx, out.PolicyArn)
	state.Namespace = flex.StringValueToFramework(ctx, namespace)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *iamPolicyAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan, state resourceIAMPolicyAssignmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AssignmentStatus.Equal(state.AssignmentStatus) ||
		!plan.Identities.Equal(state.Identities) ||
		!plan.PolicyARN.Equal(state.PolicyARN) {
		in := quicksight.UpdateIAMPolicyAssignmentInput{
			AwsAccountId:     plan.AWSAccountID.ValueStringPointer(),
			Namespace:        plan.Namespace.ValueStringPointer(),
			AssignmentName:   plan.AssignmentName.ValueStringPointer(),
			AssignmentStatus: awstypes.AssignmentStatus(plan.AssignmentStatus.ValueString()),
		}

		if !plan.Identities.IsNull() {
			var identities []identitiesData
			resp.Diagnostics.Append(plan.Identities.ElementsAs(ctx, &identities, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.Identities = expandIdentities(ctx, identities)
		}
		if !plan.PolicyARN.IsNull() {
			in.PolicyArn = plan.PolicyARN.ValueStringPointer()
		}

		out, err := conn.UpdateIAMPolicyAssignment(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameIAMPolicyAssignment, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameIAMPolicyAssignment, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		plan.AssignmentID = flex.StringToFramework(ctx, out.AssignmentId)

		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	}
}

func (r *iamPolicyAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceIAMPolicyAssignmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, namespace, assignmentName, err := iamPolicyAssignmentParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameIAMPolicyAssignment, state.ID.String(), nil),
			err.Error(),
		)
	}

	_, err = conn.DeleteIAMPolicyAssignment(ctx, &quicksight.DeleteIAMPolicyAssignmentInput{
		AssignmentName: aws.String(assignmentName),
		AwsAccountId:   aws.String(awsAccountID),
		Namespace:      aws.String(namespace),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameIAMPolicyAssignment, state.ID.String(), nil),
			err.Error(),
		)
	}

	// wait for IAM to propagate before returning
	_, err = tfresource.RetryUntilNotFound(ctx, iamPropagationTimeout, func() (any, error) {
		return findIAMPolicyAssignmentByThreePartKey(ctx, conn, awsAccountID, namespace, assignmentName)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameIAMPolicyAssignment, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
}

func findIAMPolicyAssignmentByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace, assignmentName string) (*awstypes.IAMPolicyAssignment, error) {
	input := &quicksight.DescribeIAMPolicyAssignmentInput{
		AssignmentName: aws.String(assignmentName),
		AwsAccountId:   aws.String(awsAccountID),
		Namespace:      aws.String(namespace),
	}

	return findIAMPolicyAssignment(ctx, conn, input)
}

func findIAMPolicyAssignment(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeIAMPolicyAssignmentInput) (*awstypes.IAMPolicyAssignment, error) {
	output, err := conn.DescribeIAMPolicyAssignment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.IAMPolicyAssignment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.IAMPolicyAssignment, nil
}

const iamPolicyAssignmentResourceIDSeparator = ","

func iamPolicyAssignmentCreateResourceID(awsAccountID, namespace, assignmentName string) string {
	parts := []string{awsAccountID, namespace, assignmentName}
	id := strings.Join(parts, iamPolicyAssignmentResourceIDSeparator)

	return id
}

func iamPolicyAssignmentParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, iamPolicyAssignmentResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sNAMESPACE%[2]sASSIGNMENT_NAME", id, iamPolicyAssignmentResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

var (
	identitiesAttrTypes = map[string]attr.Type{
		"user":  types.SetType{ElemType: types.StringType},
		"group": types.SetType{ElemType: types.StringType},
	}
)

type resourceIAMPolicyAssignmentData struct {
	AssignmentID     types.String `tfsdk:"assignment_id"`
	AssignmentName   types.String `tfsdk:"assignment_name"`
	AssignmentStatus types.String `tfsdk:"assignment_status"`
	AWSAccountID     types.String `tfsdk:"aws_account_id"`
	ID               types.String `tfsdk:"id"`
	Identities       types.List   `tfsdk:"identities"`
	Namespace        types.String `tfsdk:"namespace"`
	PolicyARN        types.String `tfsdk:"policy_arn"`
}

type identitiesData struct {
	User  types.Set `tfsdk:"user"`
	Group types.Set `tfsdk:"group"`
}

func expandIdentities(ctx context.Context, tfList []identitiesData) map[string][]string {
	if len(tfList) == 0 {
		return nil
	}
	tfObj := tfList[0]

	apiObject := map[string][]string{}
	if !tfObj.User.IsNull() {
		apiObject[identitiesUserKey] = flex.ExpandFrameworkStringValueSet(ctx, tfObj.User)
	}
	if !tfObj.Group.IsNull() {
		apiObject[identitiesGroupKey] = flex.ExpandFrameworkStringValueSet(ctx, tfObj.Group)
	}
	return apiObject
}

func flattenIdentities(ctx context.Context, apiObject map[string][]string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: identitiesAttrTypes}

	if len(apiObject) == 0 {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"user":  flex.FlattenFrameworkStringValueSet(ctx, apiObject[identitiesUserKey]),
		"group": flex.FlattenFrameworkStringValueSet(ctx, apiObject[identitiesGroupKey]),
	}

	objVal, d := types.ObjectValue(identitiesAttrTypes, obj)
	diags.Append(d...)
	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}
