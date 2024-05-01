// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	roleNameMaxLen        = 64
	roleNamePrefixMaxLen  = roleNameMaxLen - id.UniqueIDSuffixLength
	ResNameRole           = "IAM Role"
	MaxSessionDurationMin = 3600
	MaxSessionDurationMax = 43200
)

// @FrameworkResource(name="Role")
// @Tags(identifierAttribute="id", resourceType="Role")
func NewResourceRole(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceIamRole{}
	r.SetMigratedFromPluginSDK(true)
	return r, nil
}

type resourceIamRole struct {
	framework.ResourceWithConfigure
}

func (r *resourceIamRole) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_iam_role"
}

// TODO: should stringvalidator have something like `RegexNotMatches`? Making custom one for now that's opposite of given one
// From terraform-plugin-framework, just the opposite implementation
// https://github.com/hashicorp/terraform-plugin-framework-validators/blob/main/stringvalidator/regex_matches.go
func RegexNotMatches(regexp *regexp.Regexp, message string) validator.String {
	return regexNotMatchesValidator{
		regexp:  regexp,
		message: message,
	}
}

type regexNotMatchesValidator struct {
	regexp  *regexp.Regexp
	message string
}

// Description describes the validation in plain text formatting.
func (v regexNotMatchesValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must not match regular expression '%s'", v.regexp)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v regexNotMatchesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v regexNotMatchesValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if v.regexp.MatchString(value) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}

// As this is map and logic a little more complex, felt appropriate to make it's own planmodifier
func EditPlanForSameReorderedPolicies() planmodifier.Map {
	return editPlanForSameReorderedPolicies{}
}

type editPlanForSameReorderedPolicies struct{}

func (m editPlanForSameReorderedPolicies) Description(_ context.Context) string {
	return "If plan and state of inline policy is the same equivalent policy, do not include in plan"
}

func (m editPlanForSameReorderedPolicies) MarkdownDescription(_ context.Context) string {
	return "If plan and state of inline policy is the same equivalent policy, do not include in plan"
}

func (m editPlanForSameReorderedPolicies) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}

	if req.StateValue.IsUnknown() || req.StateValue.IsNull() {
		return
	}

	planInlinePoliciesMap := flex.ExpandFrameworkStringValueMap(ctx, req.PlanValue)

	if len(planInlinePoliciesMap) == 0 {
		return
	}
	stateInlinePoliciesMap := flex.ExpandFrameworkStringValueMap(ctx, req.StateValue)

	// If policies match, set plan for policy to use state version so that we don't see if diff bc ordering does not matter
	for name, planPolicyDoc := range planInlinePoliciesMap {
		if statePolicyDoc, ok := stateInlinePoliciesMap[name]; ok {
			if verify.PolicyStringsEquivalent(planPolicyDoc, statePolicyDoc) {
				planInlinePoliciesMap[name] = statePolicyDoc
			}
		}
	}

	resp.PlanValue = flex.FlattenFrameworkStringValueMap(ctx, planInlinePoliciesMap)
}

func (r *resourceIamRole) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": framework.IDAttribute(),
			"assume_role_policy": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.IAMPolicyType,
			},
			"create_date": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), `must satisfy regular expression pattern: [\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*)`),
					RegexNotMatches(regexache.MustCompile("[“‘]"), "cannot contain specially formatted single or double quotes: [“‘]"),
				},
			},
			"force_detach_policies": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"inline_policies": schema.MapAttribute{
				ElementType: fwtypes.IAMPolicyType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					EditPlanForSameReorderedPolicies(),
				},
				Validators: []validator.Map{
					mapvalidator.KeysAre(stringvalidator.LengthBetween(1, rolePolicyNameMaxLen)),
					mapvalidator.KeysAre(stringvalidator.RegexMatches(regexache.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]")),
				},
			},
			"managed_policy_arns": schema.SetAttribute{
				Optional:    true,
				ElementType: fwtypes.ARNType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"max_session_duration": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(MaxSessionDurationMin),
				Validators: []validator.Int64{
					int64validator.Between(MaxSessionDurationMin, MaxSessionDurationMax),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
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
				Computed: true,
				Default:  stringdefault.StaticString("/"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 512),
				},
			},
			"permissions_boundary": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"unique_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

type resourceIamRoleData struct {
	ARN                 fwtypes.ARN       `tfsdk:"arn"`
	AssumeRolePolicy    fwtypes.IAMPolicy `tfsdk:"assume_role_policy"`
	CreateDate          types.String      `tfsdk:"create_date"`
	ID                  types.String      `tfsdk:"id"`
	Description         types.String      `tfsdk:"description"`
	ForceDetachPolicies types.Bool        `tfsdk:"force_detach_policies"`
	MaxSessionDuration  types.Int64       `tfsdk:"max_session_duration"`
	Name                types.String      `tfsdk:"name"`
	NamePrefix          types.String      `tfsdk:"name_prefix"`
	Path                types.String      `tfsdk:"path"`
	PermissionsBoundary fwtypes.ARN       `tfsdk:"permissions_boundary"`
	InlinePolicies      types.Map         `tfsdk:"inline_policies"`
	UniqueID            types.String      `tfsdk:"unique_id"`
	ManagedPolicyArns   types.Set         `tfsdk:"managed_policy_arns"`
	Tags                types.Map         `tfsdk:"tags"`
	TagsAll             types.Map         `tfsdk:"tags_all"`
}

func oldSDKRoleSchema() schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
			},
			"assume_role_policy": schema.StringAttribute{
				Required: true,
			},
			"create_date": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Default: stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
				},
				Optional: true,
				Computed: true,
			},
			"force_detach_policies": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"id": framework.IDAttribute(),
			"managed_policy_arns": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"max_session_duration": schema.Int64Attribute{
				Optional: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(roleNameMaxLen),
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
				Computed: true,
				Default:  stringdefault.StaticString("/"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permissions_boundary": schema.StringAttribute{
				Optional: true,
			},
			"tags":     tftags.TagsAttribute(),
			"tags_all": tftags.TagsAttributeComputedOnly(),
			"unique_id": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"inline_policy": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true,
						},
						"policy": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceIamRole) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := oldSDKRoleSchema()

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeRoleResourceStateV0toV1,
		},
	}
}

func upgradeRoleResourceStateV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	type resourceIamRoleDataV0 struct {
		ARN                 types.String `tfsdk:"arn"`
		AssumeRolePolicy    types.String `tfsdk:"assume_role_policy"`
		CreateDate          types.String `tfsdk:"create_date"`
		Description         types.String `tfsdk:"description"`
		ForceDetachPolicies types.Bool   `tfsdk:"force_detach_policies"`
		ID                  types.String `tfsdk:"id"`
		ManagedPolicyArns   types.Set    `tfsdk:"managed_policy_arns"`
		MaxSessionDuration  types.Int64  `tfsdk:"max_session_duration"`
		Name                types.String `tfsdk:"name"`
		NamePrefix          types.String `tfsdk:"name_prefix"`
		Path                types.String `tfsdk:"path"`
		PermissionsBoundary types.String `tfsdk:"permissions_boundary"`
		Tags                types.Map    `tfsdk:"tags"`
		TagsAll             types.Map    `tfsdk:"tags_all"`
		UniqueID            types.String `tfsdk:"unique_id"`
		InlinePolicy        types.Set    `tfsdk:"inline_policy"`
	}

	var roleDataV0 resourceIamRoleDataV0

	resp.Diagnostics.Append(req.State.Get(ctx, &roleDataV0)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleDataCurrent := resourceIamRoleData{
		ARN:                 fwtypes.ARNValueMust(roleDataV0.ARN.ValueString()),
		AssumeRolePolicy:    fwtypes.IAMPolicyValue(roleDataV0.AssumeRolePolicy.ValueString()),
		CreateDate:          roleDataV0.CreateDate,
		Description:         roleDataV0.Description,
		ForceDetachPolicies: roleDataV0.ForceDetachPolicies,
		ID:                  roleDataV0.ID,
		MaxSessionDuration:  roleDataV0.MaxSessionDuration,
		Name:                roleDataV0.Name,
		NamePrefix:          roleDataV0.NamePrefix,
		Path:                roleDataV0.Path,
		UniqueID:            roleDataV0.UniqueID,
		ManagedPolicyArns:   types.SetNull(fwtypes.ARNType),
		Tags:                roleDataV0.Tags,
		TagsAll:             roleDataV0.TagsAll,
	}

	if roleDataV0.PermissionsBoundary.ValueString() == "" {
		roleDataCurrent.PermissionsBoundary = fwtypes.ARNNull()
	} else {
		roleDataCurrent.PermissionsBoundary = fwtypes.ARNValueMust(roleDataV0.PermissionsBoundary.ValueString())
	}

	type inlinePolicyData struct {
		Name   types.String `tfsdk:"name"`
		Policy types.String `tfsdk:"policy"`
	}

	var inlinePolicies []inlinePolicyData
	resp.Diagnostics.Append(roleDataV0.InlinePolicy.ElementsAs(ctx, &inlinePolicies, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inlinePoliciesMap := make(map[string]string)
	for _, inlinePolicy := range inlinePolicies {
		inlinePoliciesMap[inlinePolicy.Name.ValueString()] = inlinePolicy.Policy.ValueString()
	}
	roleDataCurrent.InlinePolicies = flex.FlattenFrameworkStringValueMap(ctx, inlinePoliciesMap)
	var policyARNs []string

	resp.Diagnostics.Append(roleDataV0.ManagedPolicyArns.ElementsAs(ctx, &policyARNs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	roleDataCurrent.ManagedPolicyArns = flex.FlattenFrameworkStringValueSet(ctx, policyARNs)

	diags := resp.State.Set(ctx, roleDataCurrent)
	resp.Diagnostics.Append(diags...)
}

func (r resourceIamRole) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().IAMClient(ctx)

	var plan resourceIamRoleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assumeRolePolicy, err := structure.NormalizeJsonString(plan.AssumeRolePolicy.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameRole, plan.AssumeRolePolicy.String(), nil),
			fmt.Errorf("assume_role_policy (%s) is invalid JSON: %s", assumeRolePolicy, err).Error(),
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
		input.MaxSessionDuration = aws.Int32(int32(plan.MaxSessionDuration.ValueInt64()))
	}

	if !plan.PermissionsBoundary.IsNull() {
		input.PermissionsBoundary = aws.String(plan.PermissionsBoundary.ValueString())
	}

	output, err := retryCreateRole(ctx, conn, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameRole, name, nil),
			err.Error(),
		)
		return
	}

	roleName := aws.ToString(output.Role.RoleName)

	if !plan.InlinePolicies.IsNull() && !plan.InlinePolicies.IsUnknown() {
		inlinePoliciesMap := flex.ExpandFrameworkStringValueMap(ctx, plan.InlinePolicies)

		policies := expandRoleInlinePolicies(roleName, inlinePoliciesMap)
		if err := r.addRoleInlinePolicies(ctx, policies); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameRole, name, nil),
				err.Error(),
			)
			return
		}
	}

	if !plan.ManagedPolicyArns.IsNull() && !plan.ManagedPolicyArns.IsUnknown() {
		var managedPolicies []string
		resp.Diagnostics.Append(plan.ManagedPolicyArns.ElementsAs(ctx, &managedPolicies, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := r.addRoleManagedPolicies(ctx, roleName, managedPolicies); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameRole, name, nil),
				err.Error(),
			)
			return
		}
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := roleCreateTags(ctx, conn, name, tags)

		// TODO: not sure how to convert this to framework
		// If default tags only, continue. Otherwise, error.
		// if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		// return append(diags, resourceRoleRead(ctx, d, meta)...)
		// }

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, fmt.Sprintf("%s tags", ResNameRole), name, nil),
				err.Error(),
			)
			return
		}
	}

	plan.ARN = fwtypes.ARNValueMust(*output.Role.Arn)
	plan.CreateDate = flex.StringValueToFramework(ctx, output.Role.CreateDate.Format(time.RFC3339))
	plan.ID = flex.StringToFramework(ctx, output.Role.RoleName)
	plan.Name = flex.StringToFramework(ctx, output.Role.RoleName)
	plan.NamePrefix = flex.StringToFramework(ctx, create.NamePrefixFromName(aws.ToString(output.Role.RoleName)))
	plan.UniqueID = flex.StringToFramework(ctx, output.Role.RoleId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r resourceIamRole) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state resourceIamRoleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasInline := false
	if !state.InlinePolicies.IsNull() && !state.InlinePolicies.IsUnknown() {
		hasInline = true
	}

	hasManaged := false
	if !state.ManagedPolicyArns.IsNull() && !state.ManagedPolicyArns.IsUnknown() {
		hasManaged = true
	}

	err := DeleteRole(ctx, conn, state.Name.ValueString(), state.ForceDetachPolicies.ValueBool(), hasInline, hasManaged)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionDeleting, state.Name.String(), state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceIamRole) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *resourceIamRole) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if !request.Plan.Raw.IsNull() && !request.State.Raw.IsNull() {
		var state, plan resourceIamRoleData

		response.Diagnostics.Append(request.State.Get(ctx, &state)...)

		if response.Diagnostics.HasError() {
			return
		}

		response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

		if response.Diagnostics.HasError() {
			return
		}

		if state.Description.ValueString() == plan.Description.ValueString() {
			response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root("description"), state.Description)...)
		}

		if state.AssumeRolePolicy.ValueString() == plan.AssumeRolePolicy.ValueString() {
			response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root("assume_role_policy"), state.AssumeRolePolicy)...)
		}

		if state.NamePrefix.ValueString() == plan.NamePrefix.ValueString() {
			response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root("name_prefix"), state.NamePrefix)...)
		}
	}
	r.SetTagsAll(ctx, request, response)
}

func (r resourceIamRole) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state resourceIamRoleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//NOTE: Have to always set this to true? Else not sure what to do
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindRoleByName(ctx, conn, state.ID.ValueString())
	}, true)

	// NOTE: Same issue here, left old conditional here as example, not sure what else can/should be done
	// if !d.IsNewResource() && tfresource.NotFound(err) {
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionSetting, state.Name.String(), state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	role := outputRaw.(*awstypes.Role)

	// occasionally, immediately after a role is created, AWS will give an ARN like AROAQ7SSZBKHREXAMPLE (unique ID)
	if role, err = waitRoleARNIsNotUniqueID(ctx, conn, state.ARN.ValueString(), role); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionSetting, state.Name.String(), state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = fwtypes.ARNValueMust(*role.Arn)
	state.CreateDate = flex.StringValueToFramework(ctx, role.CreateDate.Format(time.RFC3339))
	state.Path = flex.StringToFramework(ctx, role.Path)
	state.Name = flex.StringToFramework(ctx, role.RoleName)
	state.ID = flex.StringToFramework(ctx, role.RoleName)
	state.Description = flex.StringToFramework(ctx, role.Description)
	state.NamePrefix = flex.StringToFramework(ctx, create.NamePrefixFromName(aws.ToString(role.RoleName)))
	state.MaxSessionDuration = flex.Int32ValueToFramework(ctx, int32(*role.MaxSessionDuration))
	state.UniqueID = flex.StringToFramework(ctx, role.RoleId)

	if state.ForceDetachPolicies.IsNull() {
		state.ForceDetachPolicies = types.BoolValue(false)
	}

	if role.PermissionsBoundary != nil {
		state.PermissionsBoundary = fwtypes.ARNValueMust(*role.PermissionsBoundary.PermissionsBoundaryArn)
	} else {
		state.PermissionsBoundary = fwtypes.ARNNull()
	}

	assumeRolePolicy, err := url.QueryUnescape(aws.ToString(role.AssumeRolePolicyDocument))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionReading, state.ID.String(), state.AssumeRolePolicy.String(), err),
			err.Error(),
		)
		return
	}

	policyToSet, err := verify.PolicyToSet(state.AssumeRolePolicy.ValueString(), assumeRolePolicy)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionReading, state.ID.String(), state.AssumeRolePolicy.String(), err),
			err.Error(),
		)
		return
	}
	state.AssumeRolePolicy = fwtypes.IAMPolicyValue(policyToSet)

	// Unforunately because of `aws_iam_role_policy` and those like it, we have to ignore unless
	// added via create
	if !state.InlinePolicies.IsNull() && !state.InlinePolicies.IsUnknown() {
		inlinePolicies, err := r.readRoleInlinePolicies(ctx, aws.ToString(role.RoleName))
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionReading, state.InlinePolicies.String(), state.ID.String(), err),
				err.Error(),
			)
			return
		}

		var configPoliciesList []*iam.PutRolePolicyInput
		inlinePoliciesMap := flex.ExpandFrameworkStringValueMap(ctx, state.InlinePolicies)
		configPoliciesList = expandRoleInlinePolicies(aws.ToString(role.RoleName), inlinePoliciesMap)

		if !inlinePoliciesEquivalent(inlinePolicies, configPoliciesList) {
			state.InlinePolicies = flex.FlattenFrameworkStringValueMap(ctx, flattenRoleInlinePolicies(inlinePolicies))
		}
	}

	// like Inline policies, only reading if set in state already via updates, create, etc
	if !state.ManagedPolicyArns.IsNull() && !state.ManagedPolicyArns.IsUnknown() {
		policyARNs, err := findRoleAttachedPolicies(ctx, conn, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionReading, state.ManagedPolicyArns.String(), state.ID.String(), err),
				err.Error(),
			)
			return
		}
		if len(policyARNs) == 0 {
			state.ManagedPolicyArns = types.SetValueMust(fwtypes.ARNType, []attr.Value{})
		} else {
			state.ManagedPolicyArns = flex.FlattenFrameworkStringValueSet(ctx, policyARNs)
		}
	}
	setTagsOut(ctx, role.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r resourceIamRole) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().IAMClient(ctx)

	var plan, state resourceIamRoleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AssumeRolePolicy.Equal(state.AssumeRolePolicy) {
		assumeRolePolicy, err := structure.NormalizeJsonString(plan.AssumeRolePolicy.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.AssumeRolePolicy.String(), plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		input := &iam.UpdateAssumeRolePolicyInput{
			RoleName:       aws.String(state.ID.ValueString()),
			PolicyDocument: aws.String(assumeRolePolicy),
		}

		_, err = tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.UpdateAssumeRolePolicy(ctx, input)
			},
			func(err error) (bool, error) {
				if errs.IsAErrorMessageContains[*awstypes.MalformedPolicyDocumentException](err, "Invalid principal in policy") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.AssumeRolePolicy.String(), state.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	if !plan.Description.Equal(state.Description) {
		input := &iam.UpdateRoleDescriptionInput{
			RoleName:    aws.String(state.ID.ValueString()),
			Description: aws.String(plan.Description.ValueString()),
		}

		_, err := conn.UpdateRoleDescription(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionReading, state.ID.String(), plan.Description.String(), err),
				err.Error(),
			)
			return
		}

		state.Description = plan.Description
	}

	if !plan.MaxSessionDuration.Equal(state.MaxSessionDuration) {
		input := &iam.UpdateRoleInput{
			RoleName:           aws.String(state.ID.ValueString()),
			MaxSessionDuration: aws.Int32(int32(plan.MaxSessionDuration.ValueInt64())),
		}

		_, err := conn.UpdateRole(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.ID.String(), plan.MaxSessionDuration.String(), err),
				err.Error(),
			)
			return
		}
		state.MaxSessionDuration = plan.MaxSessionDuration
	}

	if !plan.PermissionsBoundary.Equal(state.PermissionsBoundary) {
		if !plan.PermissionsBoundary.IsNull() {
			input := &iam.PutRolePermissionsBoundaryInput{
				PermissionsBoundary: aws.String(plan.PermissionsBoundary.ValueString()),
				RoleName:            aws.String(state.ID.ValueString()),
			}

			_, err := conn.PutRolePermissionsBoundary(ctx, input)

			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.ID.String(), plan.PermissionsBoundary.String(), err),
					err.Error(),
				)
				return
			}
		} else {
			input := &iam.DeleteRolePermissionsBoundaryInput{
				RoleName: aws.String(state.ID.ValueString()),
			}

			_, err := conn.DeleteRolePermissionsBoundary(ctx, input)

			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.IAM, create.ErrActionDeleting, state.ID.String(), plan.PermissionsBoundary.String(), err),
					err.Error(),
				)
				return
			}
		}

		state.PermissionsBoundary = plan.PermissionsBoundary
	}

	if !plan.InlinePolicies.Equal(state.InlinePolicies) && inlinePoliciesActualDiff(ctx, &plan, &state) {
		oldInlinePoliciesMap := flex.ExpandFrameworkStringValueMap(ctx, state.InlinePolicies)
		newInlinePoliciesMap := flex.ExpandFrameworkStringValueMap(ctx, plan.InlinePolicies)

		var removePolicyNames []string
		for k := range oldInlinePoliciesMap {
			if _, ok := newInlinePoliciesMap[k]; !ok {
				removePolicyNames = append(removePolicyNames, k)
			}
		}

		// need set like object to store policy names we want to add
		addPolicyNames := make(map[string]int64)
		for k, v := range newInlinePoliciesMap {
			val, ok := oldInlinePoliciesMap[k]
			// If the key exists
			if !ok {
				addPolicyNames[k] = 0
				continue
			}

			if !verify.PolicyStringsEquivalent(v, val) {
				addPolicyNames[k] = 0
			}
		}

		roleName := state.Name.ValueString()
		nsPolicies := expandRoleInlinePolicies(roleName, newInlinePoliciesMap)

		// getting policy objects we want to add based on add_policy_names map
		var addPolicies []*iam.PutRolePolicyInput
		for _, val := range nsPolicies {
			if _, ok := addPolicyNames[*val.PolicyName]; ok {
				addPolicies = append(addPolicies, val)
			}
		}

		// Always add before delete
		if err := r.addRoleInlinePolicies(ctx, addPolicies); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.ID.String(), plan.InlinePolicies.String(), err),
				err.Error(),
			)
			return
		}

		if err := deleteRoleInlinePolicies(ctx, conn, roleName, removePolicyNames); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.ID.String(), plan.InlinePolicies.String(), err),
				err.Error(),
			)
			return
		}
	}

	if !plan.ManagedPolicyArns.Equal(state.ManagedPolicyArns) {
		var oldManagedARNs, newManagedARNs []string
		resp.Diagnostics.Append(state.ManagedPolicyArns.ElementsAs(ctx, &oldManagedARNs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(plan.ManagedPolicyArns.ElementsAs(ctx, &newManagedARNs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var add, del []string

		oldPolicyArnMap := make(map[string]int64)
		for _, v := range oldManagedARNs {
			oldPolicyArnMap[v] = 0
		}

		for _, v := range newManagedARNs {
			if _, ok := oldPolicyArnMap[v]; !ok {
				add = append(add, v)
			}
		}

		newPolicyArnMap := make(map[string]int64)
		for _, v := range newManagedARNs {
			newPolicyArnMap[v] = 0
		}

		for _, v := range oldManagedARNs {
			if _, ok := newPolicyArnMap[v]; !ok {
				del = append(del, v)
			}
		}

		if err := r.addRoleManagedPolicies(ctx, state.ID.ValueString(), add); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.ID.String(), plan.ManagedPolicyArns.String(), err),
				err.Error(),
			)
			return
		}

		if err := deleteRolePolicyAttachments(ctx, conn, state.ID.ValueString(), del); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.ID.String(), plan.ManagedPolicyArns.String(), err),
				err.Error(),
			)
			return
		}
	}

	if !plan.TagsAll.Equal(state.TagsAll) {
		err := roleUpdateTags(ctx, conn, plan.ID.ValueString(), state.TagsAll, plan.TagsAll)

		// Some partitions (e.g. ISO) may not support tagging.
		parition := r.Meta().Partition
		if errs.IsUnsupportedOperationInPartitionError(parition, err) {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.ID.String(), plan.TagsAll.String(), err),
				err.Error(),
			)
			return
		}

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, state.ID.String(), plan.TagsAll.String(), err),
				err.Error(),
			)
			return
		}
	}
	plan.NamePrefix = flex.StringToFramework(ctx, create.NamePrefixFromName(plan.Name.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func findRoleByName(ctx context.Context, conn *iam.Client, name string) (*awstypes.Role, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}

	return findRole(ctx, conn, input)
}

func findRole(ctx context.Context, conn *iam.Client, input *iam.GetRoleInput) (*awstypes.Role, error) {
	output, err := conn.GetRole(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
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

func retryCreateRole(ctx context.Context, conn *iam.Client, input *iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateRole(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.MalformedPolicyDocumentException](err, "Invalid principal in policy") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return nil, err
	}

	output, ok := outputRaw.(*iam.CreateRoleOutput)
	if !ok || output == nil || aws.ToString(output.Role.RoleName) == "" {
		return nil, fmt.Errorf("create IAM role (%s) returned an empty result", aws.ToString(input.RoleName))
	}

	return output, err
}

func (r resourceIamRole) addRoleManagedPolicies(ctx context.Context, roleName string, policies []string) error {
	conn := r.Meta().IAMClient(ctx)
	var errs []error

	for _, arn := range policies {
		if err := attachPolicyToRole(ctx, conn, roleName, arn); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func DeleteRole(ctx context.Context, conn *iam.Client, roleName string, forceDetach, hasInline, hasManaged bool) error {
	if err := deleteRoleInstanceProfiles(ctx, conn, roleName); err != nil {
		return err
	}

	if forceDetach || hasManaged {
		policyARNs, err := findRoleAttachedPolicies(ctx, conn, roleName)

		if err != nil {
			return fmt.Errorf("reading IAM Policies attached to Role (%s): %w", roleName, err)
		}

		if err := deleteRolePolicyAttachments(ctx, conn, roleName, policyARNs); err != nil {
			return err
		}
	}

	if forceDetach || hasInline {
		inlinePolicies, err := findRolePolicyNames(ctx, conn, roleName)

		if err != nil {
			return fmt.Errorf("reading IAM Role (%s) inline policies: %w", roleName, err)
		}

		if err := deleteRoleInlinePolicies(ctx, conn, roleName, inlinePolicies); err != nil {
			return err
		}
	}

	input := &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.DeleteConflictException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.DeleteRole(ctx, input)
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil
	}

	return err
}

func deleteRoleInstanceProfiles(ctx context.Context, conn *iam.Client, roleName string) error {
	instanceProfiles, err := findInstanceProfilesForRole(ctx, conn, roleName)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading IAM Instance Profiles for Role (%s): %w", roleName, err)
	}

	var errsList []error

	for _, instanceProfile := range instanceProfiles {
		instanceProfileName := aws.ToString(instanceProfile.InstanceProfileName)
		input := &iam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: aws.String(instanceProfileName),
			RoleName:            aws.String(roleName),
		}

		_, err := conn.RemoveRoleFromInstanceProfile(ctx, input)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			continue
		}

		if err != nil {
			errsList = append(errsList, fmt.Errorf("removing IAM Role (%s) from Instance Profile (%s): %w", roleName, instanceProfileName, err))
		}
	}

	return errors.Join(errsList...)
}

func findRoleAttachedPolicies(ctx context.Context, conn *iam.Client, roleName string) ([]string, error) {
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}
	var output []string

	pages := iam.NewListAttachedRolePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AttachedPolicies {
			if !reflect.ValueOf(v).IsZero() {
				output = append(output, aws.ToString(v.PolicyArn))
			}
		}
	}

	return output, nil
}

func deleteRolePolicyAttachments(ctx context.Context, conn *iam.Client, roleName string, policyARNs []string) error {
	var errsList []error

	for _, policyARN := range policyARNs {
		input := &iam.DetachRolePolicyInput{
			PolicyArn: aws.String(policyARN),
			RoleName:  aws.String(roleName),
		}

		_, err := conn.DetachRolePolicy(ctx, input)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			continue
		}

		if err != nil {
			errsList = append(errsList, fmt.Errorf("detaching IAM Policy (%s) from Role (%s): %w", policyARN, roleName, err))
		}
	}

	return errors.Join(errsList...)
}

func findRolePolicyNames(ctx context.Context, conn *iam.Client, roleName string) ([]string, error) {
	input := &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	}
	var output []string

	pages := iam.NewListRolePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.PolicyNames {
			if v != "" {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func deleteRoleInlinePolicies(ctx context.Context, conn *iam.Client, roleName string, policyNames []string) error {
	var errsList []error

	for _, policyName := range policyNames {
		if len(policyName) == 0 {
			continue
		}

		input := &iam.DeleteRolePolicyInput{
			PolicyName: aws.String(policyName),
			RoleName:   aws.String(roleName),
		}

		_, err := conn.DeleteRolePolicy(ctx, input)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			continue
		}

		if err != nil {
			errsList = append(errsList, fmt.Errorf("deleting IAM Role (%s) policy (%s): %w", roleName, policyName, err))
		}
	}

	return errors.Join(errsList...)
}

func expandRoleInlinePolicies(roleName string, tfPoliciesMap map[string]string) []*iam.PutRolePolicyInput {
	if len(tfPoliciesMap) == 0 {
		return nil
	}

	var apiObjects []*iam.PutRolePolicyInput

	for policyName, policyDocument := range tfPoliciesMap {
		apiObject := expandRoleInlinePolicy(roleName, policyName, policyDocument)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandRoleInlinePolicy(roleName string, policyName string, policyDocument string) *iam.PutRolePolicyInput {
	apiObject := &iam.PutRolePolicyInput{}

	apiObject.PolicyName = aws.String(policyName)
	apiObject.PolicyDocument = aws.String(policyDocument)
	apiObject.RoleName = aws.String(roleName)

	return apiObject
}

func (r resourceIamRole) addRoleInlinePolicies(ctx context.Context, policies []*iam.PutRolePolicyInput) error {
	conn := r.Meta().IAMClient(ctx)
	var errs []error

	for _, policy := range policies {
		if len(aws.ToString(policy.PolicyName)) == 0 || len(aws.ToString(policy.PolicyDocument)) == 0 {
			continue
		}

		if _, err := conn.PutRolePolicy(ctx, policy); err != nil {
			errs = append(errs, fmt.Errorf("adding inline policy (%s): %w", aws.ToString(policy.PolicyName), err))
		}
	}

	return errors.Join(errs...)
}

func inlinePoliciesActualDiff(ctx context.Context, plan *resourceIamRoleData, state *resourceIamRoleData) bool {
	roleName := state.Name.ValueString()

	oldInlinePoliciesMap := flex.ExpandFrameworkStringValueMap(ctx, state.InlinePolicies)
	newInlinePoliciesMap := flex.ExpandFrameworkStringValueMap(ctx, plan.InlinePolicies)

	osPolicies := expandRoleInlinePolicies(roleName, oldInlinePoliciesMap)
	nsPolicies := expandRoleInlinePolicies(roleName, newInlinePoliciesMap)

	return !inlinePoliciesEquivalent(nsPolicies, osPolicies)
}

func inlinePoliciesEquivalent(readPolicies, configPolicies []*iam.PutRolePolicyInput) bool {
	if readPolicies == nil && configPolicies == nil {
		return true
	}

	if len(readPolicies) == 0 && len(configPolicies) == 1 {
		if equivalent, err := awspolicy.PoliciesAreEquivalent(`{}`, aws.ToString(configPolicies[0].PolicyDocument)); err == nil && equivalent {
			return true
		}
	}

	if len(readPolicies) != len(configPolicies) {
		return false
	}

	matches := 0

	for _, policyOne := range readPolicies {
		for _, policyTwo := range configPolicies {
			if aws.ToString(policyOne.PolicyName) == aws.ToString(policyTwo.PolicyName) {
				matches++
				if equivalent, err := awspolicy.PoliciesAreEquivalent(aws.ToString(policyOne.PolicyDocument), aws.ToString(policyTwo.PolicyDocument)); err != nil || !equivalent {
					return false
				}
				break
			}
		}
	}

	return matches == len(readPolicies)
}

func (r resourceIamRole) readRoleInlinePolicies(ctx context.Context, roleName string) ([]*iam.PutRolePolicyInput, error) {
	conn := r.Meta().IAMClient(ctx)

	policyNames, err := findRolePolicyNames(ctx, conn, roleName)

	if err != nil {
		return nil, err
	}

	var apiObjects []*iam.PutRolePolicyInput

	for _, policyName := range policyNames {
		output, err := conn.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: aws.String(policyName),
		})

		if err != nil {
			return nil, err
		}

		policy, err := url.QueryUnescape(aws.ToString(output.PolicyDocument))
		if err != nil {
			return nil, err
		}

		p, err := verify.LegacyPolicyNormalize(policy)
		if err != nil {
			return nil, fmt.Errorf("policy (%s) is invalid JSON: %w", p, err)
		}

		apiObject := &iam.PutRolePolicyInput{
			RoleName:       aws.String(roleName),
			PolicyDocument: aws.String(p),
			PolicyName:     aws.String(policyName),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func flattenRoleInlinePolicies(apiObjects []*iam.PutRolePolicyInput) map[string]string {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := make(map[string]string)

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfMap[aws.ToString(apiObject.PolicyName)] = aws.ToString(apiObject.PolicyDocument)
	}

	return tfMap
}

func roleTags(ctx context.Context, conn *iam.Client, identifier string) ([]awstypes.Tag, error) {
	output, err := conn.ListRoleTags(ctx, &iam.ListRoleTagsInput{
		RoleName: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}

	return output.Tags, nil
}
