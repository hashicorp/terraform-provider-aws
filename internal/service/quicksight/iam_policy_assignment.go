// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="IAM Policy Assignment")
func newResourceIAMPolicyAssignment(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceIAMPolicyAssignment{}, nil
}

const (
	ResNameIAMPolicyAssignment = "IAM Policy Assignment"

	DefaultIAMPolicyAssignmentNamespace = "default"
	identitiesUserKey                   = "user"
	identitiesGroupKey                  = "group"
)

type resourceIAMPolicyAssignment struct {
	framework.ResourceWithConfigure
}

func (r *resourceIAMPolicyAssignment) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_quicksight_iam_policy_assignment"
}

func (r *resourceIAMPolicyAssignment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					stringvalidator.OneOf(quicksight.AssignmentStatus_Values()...),
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
				Default:  stringdefault.StaticString(DefaultIAMPolicyAssignmentNamespace),
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

func (r *resourceIAMPolicyAssignment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var plan resourceIAMPolicyAssignmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID)
	}
	plan.ID = types.StringValue(createIAMPolicyAssignmentID(plan.AWSAccountID.ValueString(), plan.Namespace.ValueString(), plan.AssignmentName.ValueString()))

	in := quicksight.CreateIAMPolicyAssignmentInput{
		AwsAccountId:     aws.String(plan.AWSAccountID.ValueString()),
		Namespace:        aws.String(plan.Namespace.ValueString()),
		AssignmentName:   aws.String(plan.AssignmentName.ValueString()),
		AssignmentStatus: aws.String(plan.AssignmentStatus.ValueString()),
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
		in.PolicyArn = aws.String(plan.PolicyARN.ValueString())
	}

	out, err := conn.CreateIAMPolicyAssignmentWithContext(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameIAMPolicyAssignment, plan.AssignmentName.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameIAMPolicyAssignment, plan.AssignmentName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	plan.AssignmentID = flex.StringToFramework(ctx, out.AssignmentId)

	// wait for IAM to propagate before returning
	_, err = tfresource.RetryWhenNotFound(ctx, iamPropagationTimeout, func() (interface{}, error) {
		return FindIAMPolicyAssignmentByID(ctx, conn, plan.ID.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameIAMPolicyAssignment, plan.AssignmentName.String(), nil),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceIAMPolicyAssignment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceIAMPolicyAssignmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindIAMPolicyAssignmentByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameIAMPolicyAssignment, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.AssignmentID = flex.StringToFramework(ctx, out.AssignmentId)
	state.AssignmentName = flex.StringToFramework(ctx, out.AssignmentName)
	state.AssignmentStatus = flex.StringToFramework(ctx, out.AssignmentStatus)
	state.AWSAccountID = flex.StringToFramework(ctx, out.AwsAccountId)
	identities, d := flattenIdentities(ctx, out.Identities)
	resp.Diagnostics.Append(d...)
	state.Identities = identities
	state.PolicyARN = flex.StringToFramework(ctx, out.PolicyArn)

	// To support import, parse the ID for the component keys and set
	// individual values in state
	_, namespace, _, err := ParseIAMPolicyAssignmentID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameIAMPolicyAssignment, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
	state.Namespace = flex.StringValueToFramework(ctx, namespace)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceIAMPolicyAssignment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightConn(ctx)

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
			AwsAccountId:     aws.String(plan.AWSAccountID.ValueString()),
			Namespace:        aws.String(plan.Namespace.ValueString()),
			AssignmentName:   aws.String(plan.AssignmentName.ValueString()),
			AssignmentStatus: aws.String(plan.AssignmentStatus.ValueString()),
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
			in.PolicyArn = aws.String(plan.PolicyARN.ValueString())
		}

		out, err := conn.UpdateIAMPolicyAssignmentWithContext(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameIAMPolicyAssignment, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameIAMPolicyAssignment, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		plan.AssignmentID = flex.StringToFramework(ctx, out.AssignmentId)

		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	}
}

func (r *resourceIAMPolicyAssignment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceIAMPolicyAssignmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteIAMPolicyAssignmentWithContext(ctx, &quicksight.DeleteIAMPolicyAssignmentInput{
		AwsAccountId:   aws.String(state.AWSAccountID.ValueString()),
		Namespace:      aws.String(state.Namespace.ValueString()),
		AssignmentName: aws.String(state.AssignmentName.ValueString()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameIAMPolicyAssignment, state.ID.String(), nil),
			err.Error(),
		)
	}

	// wait for IAM to propagate before returning
	_, err = tfresource.RetryUntilNotFound(ctx, iamPropagationTimeout, func() (interface{}, error) {
		return FindIAMPolicyAssignmentByID(ctx, conn, state.ID.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameIAMPolicyAssignment, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
}

func (r *resourceIAMPolicyAssignment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindIAMPolicyAssignmentByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.IAMPolicyAssignment, error) {
	awsAccountID, namespace, assignmentName, err := ParseIAMPolicyAssignmentID(id)
	if err != nil {
		return nil, err
	}

	in := &quicksight.DescribeIAMPolicyAssignmentInput{
		AwsAccountId:   aws.String(awsAccountID),
		Namespace:      aws.String(namespace),
		AssignmentName: aws.String(assignmentName),
	}

	out, err := conn.DescribeIAMPolicyAssignmentWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.IAMPolicyAssignment == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.IAMPolicyAssignment, nil
}

func ParseIAMPolicyAssignmentID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ",", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,NAMESPACE,ASSIGNMENT_NAME", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func createIAMPolicyAssignmentID(awsAccountID, namespace, assignmentName string) string {
	return fmt.Sprintf("%s,%s,%s", awsAccountID, namespace, assignmentName)
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

func expandIdentities(ctx context.Context, tfList []identitiesData) map[string][]*string {
	if len(tfList) == 0 {
		return nil
	}
	tfObj := tfList[0]

	apiObject := map[string][]*string{}
	if !tfObj.User.IsNull() {
		apiObject[identitiesUserKey] = flex.ExpandFrameworkStringSet(ctx, tfObj.User)
	}
	if !tfObj.Group.IsNull() {
		apiObject[identitiesGroupKey] = flex.ExpandFrameworkStringSet(ctx, tfObj.Group)
	}
	return apiObject
}

func flattenIdentities(ctx context.Context, apiObject map[string][]*string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: identitiesAttrTypes}

	if len(apiObject) == 0 {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"user":  flex.FlattenFrameworkStringSet(ctx, apiObject[identitiesUserKey]),
		"group": flex.FlattenFrameworkStringSet(ctx, apiObject[identitiesGroupKey]),
	}

	objVal, d := types.ObjectValue(identitiesAttrTypes, obj)
	diags.Append(d...)
	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}
