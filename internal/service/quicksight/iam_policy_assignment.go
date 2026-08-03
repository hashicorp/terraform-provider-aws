// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package quicksight

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_iam_policy_assignment", name="IAM Policy Assignment")
func newIAMPolicyAssignmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &iamPolicyAssignmentResource{}, nil
}

type iamPolicyAssignmentResource struct {
	framework.ResourceWithModel[iamPolicyAssignmentResourceModel]
	framework.WithImportByID
}

func (r *iamPolicyAssignmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
				CustomType: fwtypes.StringEnumType[awstypes.AssignmentStatus](),
				Required:   true,
			},
			names.AttrAWSAccountID: quicksightschema.AWSAccountIDAttribute(),
			names.AttrID:           framework.IDAttribute(),
			names.AttrNamespace:    quicksightschema.NamespaceAttribute(),
			"policy_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"identities": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[identitiesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							ElementType: types.StringType,
						},
						"user": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *iamPolicyAssignmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data iamPolicyAssignmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.CreateIAMPolicyAssignmentInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	identities, diags := data.Identities.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	input.Identities = expandIdentities(ctx, identities)

	awsAccountID, namespace, assignmentName := aws.ToString(input.AwsAccountId), aws.ToString(input.Namespace), aws.ToString(input.AssignmentName)
	id := iamPolicyAssignmentCreateResourceID(awsAccountID, namespace, assignmentName)
	output, err := conn.CreateIAMPolicyAssignment(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating QuickSight IAM Policy Assignment (%s)", id), err.Error())

		return
	}

	// Set values for unknowns.
	data.AssignmentID = fwflex.StringToFramework(ctx, output.AssignmentId)
	data.ID = fwflex.StringValueToFramework(ctx, id)

	// wait for IAM to propagate before returning
	_, err = tfresource.RetryWhenNotFound(ctx, iamPropagationTimeout, func(ctx context.Context) (any, error) {
		return findIAMPolicyAssignmentByThreePartKey(ctx, conn, awsAccountID, namespace, assignmentName)
	})

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for QuickSight IAM Policy Assignment (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *iamPolicyAssignmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data iamPolicyAssignmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccountID, namespace, assignmentName, err := iamPolicyAssignmentParseResourceID(data.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	output, err := findIAMPolicyAssignmentByThreePartKey(ctx, conn, awsAccountID, namespace, assignmentName)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading QuickSight IAM Policy Assignment (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	if len(output.Identities) == 0 {
		data.Identities = fwtypes.NewListNestedObjectValueOfNull[identitiesModel](ctx)
	} else {
		identities, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenIdentities(ctx, output.Identities))
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		data.Identities = identities
	}
	data.Namespace = fwflex.StringValueToFramework(ctx, namespace)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *iamPolicyAssignmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old iamPolicyAssignmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	if !new.AssignmentStatus.Equal(old.AssignmentStatus) ||
		!new.Identities.Equal(old.Identities) ||
		!new.PolicyARN.Equal(old.PolicyARN) {
		var input quicksight.UpdateIAMPolicyAssignmentInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		identities, diags := new.Identities.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		input.Identities = expandIdentities(ctx, identities)

		output, err := conn.UpdateIAMPolicyAssignment(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating QuickSight IAM Policy Assignment (%s)", new.ID.ValueString()), err.Error())

			return
		}

		// Set values for unknowns.
		new.AssignmentID = fwflex.StringToFramework(ctx, output.AssignmentId)

		response.Diagnostics.Append(response.State.Set(ctx, new)...)
	}
}

func (r *iamPolicyAssignmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data iamPolicyAssignmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccountID, namespace, assignmentName, err := iamPolicyAssignmentParseResourceID(data.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	input := quicksight.DeleteIAMPolicyAssignmentInput{
		AssignmentName: aws.String(assignmentName),
		AwsAccountId:   aws.String(awsAccountID),
		Namespace:      aws.String(namespace),
	}
	_, err = conn.DeleteIAMPolicyAssignment(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting QuickSight IAM Policy Assignment (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// wait for IAM to propagate before returning
	_, err = tfresource.RetryUntilNotFound(ctx, iamPropagationTimeout, func(ctx context.Context) (any, error) {
		return findIAMPolicyAssignmentByThreePartKey(ctx, conn, awsAccountID, namespace, assignmentName)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for QuickSight IAM Policy Assignment (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findIAMPolicyAssignmentByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace, assignmentName string) (*awstypes.IAMPolicyAssignment, error) {
	input := quicksight.DescribeIAMPolicyAssignmentInput{
		AssignmentName: aws.String(assignmentName),
		AwsAccountId:   aws.String(awsAccountID),
		Namespace:      aws.String(namespace),
	}

	return findIAMPolicyAssignment(ctx, conn, &input)
}

func findIAMPolicyAssignment(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeIAMPolicyAssignmentInput) (*awstypes.IAMPolicyAssignment, error) {
	output, err := conn.DescribeIAMPolicyAssignment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.IAMPolicyAssignment == nil {
		return nil, tfresource.NewEmptyResultError()
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

type iamPolicyAssignmentResourceModel struct {
	framework.WithRegionModel
	AssignmentID     types.String                                     `tfsdk:"assignment_id"`
	AssignmentName   types.String                                     `tfsdk:"assignment_name"`
	AssignmentStatus fwtypes.StringEnum[awstypes.AssignmentStatus]    `tfsdk:"assignment_status"`
	AWSAccountID     types.String                                     `tfsdk:"aws_account_id"`
	ID               types.String                                     `tfsdk:"id"`
	Identities       fwtypes.ListNestedObjectValueOf[identitiesModel] `tfsdk:"identities" autoflex:"-"`
	Namespace        types.String                                     `tfsdk:"namespace"`
	PolicyARN        fwtypes.ARN                                      `tfsdk:"policy_arn"`
}

type identitiesModel struct {
	Group fwtypes.SetOfString `tfsdk:"group"`
	User  fwtypes.SetOfString `tfsdk:"user"`
}

const (
	identitiesUserKey  = "user"
	identitiesGroupKey = "group"
)

func expandIdentities(ctx context.Context, tfObject *identitiesModel) map[string][]string {
	if tfObject == nil {
		return nil
	}

	apiObject := map[string][]string{}

	if !tfObject.Group.IsNull() {
		apiObject[identitiesGroupKey] = fwflex.ExpandFrameworkStringValueSet(ctx, tfObject.Group)
	}
	if !tfObject.User.IsNull() {
		apiObject[identitiesUserKey] = fwflex.ExpandFrameworkStringValueSet(ctx, tfObject.User)
	}

	return apiObject
}

func flattenIdentities(ctx context.Context, apiObject map[string][]string) *identitiesModel {
	if len(apiObject) == 0 {
		return nil
	}

	tfObject := &identitiesModel{
		Group: fwflex.FlattenFrameworkStringValueSetOfString(ctx, apiObject[identitiesGroupKey]),
		User:  fwflex.FlattenFrameworkStringValueSetOfString(ctx, apiObject[identitiesUserKey]),
	}

	return tfObject
}
