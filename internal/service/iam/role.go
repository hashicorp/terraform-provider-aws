// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"errors"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	roleNameMaxLen = 64
	// TODO: what?
	roleNamePrefixMaxLen = roleNameMaxLen - id.UniqueIDSuffixLength
	ResNameIamRole       = "IAM Role"
)

// TODO: finish this how does this work?

// @SDKResource("aws_iam_role", name="Role")
// @Tags
// func newIamRole(context.Context) (resource.ResourceWithConfigure, error) {
// r := &resourceIamRole{}
// // TODO
// // r.create = r.createSecurityGroupRule
// // r.delete = r.deleteSecurityGroupRule
// // r.findByID = r.findSecurityGroupRuleByID

// return r, nil
// }

type resourceIamRole struct {
	framework.ResourceWithConfigure
}

func (r *resourceIamRole) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_iam_role"
}

// TODO: Update this
func (r *resourceIamRole) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
			},
			"assume_role_policy": schema.StringAttribute{
				Required: true,
				// Validators: []validator.String{
				// // TODO: json validator
				// },
				// TODO: finish this, it get complicated
			},
			"create_date": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
					// TODO: regex does not match?
					stringvalidator.RegexMatches(
						regexache.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`),
						`must satisfy regular expression pattern: [\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*)`,
					),
				},
			},
			"force_detach_policies": schema.BoolAttribute{
				Optional: true,
				Default:  booldefault.StaticBool(false),
			},
			// TODO: inline policy goes crazy, have to figure what this type should look like
			// also read article again
			"inline_policy": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				// TODO: maybe some validation?
			},
			"managed_policy_arns": schema.SetAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				// TODO: set validator for arn
				// TODO: validate all elements of set are valid arns
				// how to do this with helper lib terraform-plugin-framework-validators
			},
			"max_session_duration": schema.Int64Attribute{
				Optional: true,
				Default:  int64default.StaticInt64(3600),
				Validators: []validator.Int64{
					int64validator.Between(3600, 43200),
				},
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(roleNameMaxLen),
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("name_prefix"),
					),
				},
			},
			"name_prefix": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(roleNamePrefixMaxLen),
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("name"),
					),
				},
			},
			"path": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Default: stringdefault.StaticString("/"),
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 512),
				},
			},
			"permissions_boundary": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			"unique_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

type resourceIamRoleData struct {
	ARN                 types.String `tfsdk:"arn"`
	AssumeRolePolicy    types.String `tfsdk:"assume_role_policy"`
	CreateDate          types.String `tfsdk:"create_date"`
	Description         types.String `tfsdk:"description"`
	ForceDetachPolicies types.Bool   `tfsdk:"force_detach_policies"`
	// TODO: still have to think this one out
	InlinePolicy        types.Map    `tfsdk:"inline_policy"`
	ManagedPolicyArns   types.Set    `tfsdk:"managed_policy_arns"`
	MaxSessionDuration  types.Int64  `tfsdk:"max_session_duration"`
	Name                types.String `tfsdk:"name"`
	NamePrefix          types.String `tfsdk:"name_prefix"`
	Path                types.String `tfsdk:"path"`
	PermissionsBoundary types.String `tfsdk:"permissions_boundary"`
	UniqueId            types.String `tfsdk:"unique_id"`
	Tags                types.Map    `tfsdk:"tags"`
	TagsAll             types.Map    `tfsdk:"tags_all"`
}

func (r resourceIamRole) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().IAMConn(ctx)

	var plan resourceIamRoleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assumeRolePolicy, err := structure.NormalizeJsonString(plan.AssumeRolePolicy.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameIamRole, plan.AssumeRolePolicy.String(), nil),
			errors.New(fmt.Sprintf("assume_role_policy (%s) is invalid JSON: %s", assumeRolePolicy, err)).Error(),
		)
		return
	}

	name := create.Name(plan.Name.ValueString(), plan.NamePrefix.ValueString())
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
		Path:                     aws.String(plan.Path.ValueString()),
		RoleName:                 aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	if !plan.Description.IsNull() {
		input.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.MaxSessionDuration.IsNull() {
		input.MaxSessionDuration = aws.Int64(plan.MaxSessionDuration.ValueInt64())
	}

	if !plan.PermissionsBoundary.IsNull() {
		input.PermissionsBoundary = aws.String(plan.PermissionsBoundary.ValueString())
	}

	output, err := retryCreateRole(ctx, conn, input)

	// TODO: So this needs tags... do we need on resourceIamRoleData?
	// if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
	// input.Tags = nil

	// output, err = retryCreateRole(ctx, conn, input)
	// }

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameIamRole, name, nil),
			err.Error(),
		)
		return
	}

	roleName := aws.StringValue(output.Role.RoleName)

	// TODO: has to figure this out because typing of inline policies
	// if !plan.InlinePolicy.IsNull() && !plan.InlinePolicy.IsUnknown() {
	// inline_policies_map := make(map[string]string)
	// plan.InlinePolicy.ElementsAs(ctx, inline_policies_map, false)
	// policies := expandRoleInlinePolicies(roleName, inline_policies_map)
	// if err := addRoleInlinePolicies(ctx, policies, meta); err != nil {
	// resp.Diagnostics.AddError(
	// create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameIamRole, name, nil),
	// err.Error(),
	// )
	// return
	// }
	// }

	if !plan.ManagedPolicyArns.IsNull() && !plan.ManagedPolicyArns.IsUnknown() {
		managedPolicies := flex.ExpandFrameworkStringSet(ctx, plan.ManagedPolicyArns)
		if err := r.addRoleManagedPolicies(ctx, roleName, managedPolicies); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameIamRole, name, nil),
				err.Error(),
			)
			return
		}
	}

	// TODO: do something with this?
	// some resources have been created but not all attributes
	// d.SetId(roleName)
	// state := plan
	// // TODO: do we need this?
	// // state.refreshFromOutput(ctx, out)
	// resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := roleCreateTags(ctx, conn, name, tags)

		// TODO: read errors or something
		// If default tags only, continue. Otherwise, error.
		// if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		// return append(diags, resourceRoleRead(ctx, d, meta)...)
		// }

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, fmt.Sprintf("%s tags", ResNameIamRole), name, nil),
				err.Error(),
			)
			return
		}
	}

	// last steps?
	state := plan
	// TODO: do we need something?this?
	// state.refreshFromOutput(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// TODO: delete/update/import
func FindRoleByName(ctx context.Context, conn *iam.IAM, name string) (*iam.Role, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}

	return findRole(ctx, conn, input)
}

func findRole(ctx context.Context, conn *iam.IAM, input *iam.GetRoleInput) (*iam.Role, error) {
	output, err := conn.GetRoleWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Role == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Role, nil
}

func retryCreateRole(ctx context.Context, conn *iam.IAM, input *iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateRoleWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, iam.ErrCodeMalformedPolicyDocumentException, "Invalid principal in policy") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return nil, err
	}

	output, ok := outputRaw.(*iam.CreateRoleOutput)
	if !ok || output == nil || aws.StringValue(output.Role.RoleName) == "" {
		return nil, fmt.Errorf("create IAM role (%s) returned an empty result", aws.StringValue(input.RoleName))
	}

	return output, err
}

func (r resourceIamRole) addRoleManagedPolicies(ctx context.Context, roleName string, policies []*string) error {
	conn := r.Meta().IAMConn(ctx)
	var errs []error

	for _, arn := range policies {
		if err := attachPolicyToRole(ctx, conn, roleName, aws.StringValue(arn)); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
