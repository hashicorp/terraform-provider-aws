// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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
	roleNameMaxLen       = 64
	roleNamePrefixMaxLen = roleNameMaxLen - id.UniqueIDSuffixLength
	ResNameIamRole       = "IAM Role"
)

// @FrameworkResource(name="Role")
// @Tags(identifierAttribute="arn")
func newResourceRole(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceIamRole{}, nil
}

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

func (r resourceIamRole) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().IAMConn(ctx)

	var state resourceIamRoleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasInline := false
	if !state.InlinePolicy.IsNull() && !state.InlinePolicy.IsUnknown() {
		hasInline = true
	}

	hasManaged := false
	if !state.ManagedPolicyArns.IsNull() && !state.ManagedPolicyArns.IsUnknown() {
		hasManaged = true
	}

	err := DeleteRole(ctx, conn, state.Name.ValueString(), state.ForceDetachPolicies.ValueBool(), hasInline, hasManaged)

	if err != nil {
		// TODO: do something like this to skip deletes on roles that are gone?
		// if err.IsA[*awstypes.ResourceNotFoundException](err) {
		// return
		// }
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionDeleting, state.Name.String(), state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r resourceIamRole) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().IAMConn(ctx)

	var state resourceIamRoleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//NOTE: Have to always set this to true? Else not sure what to do
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindRoleByName(ctx, conn, state.Name.ValueString())
	}, true)

	// NOTE: Same issue here, I left old conditional here as example, not sure what else can/should be done
	// if !d.IsNewResource() && tfresource.NotFound(err) {
	if tfresource.NotFound(err) {
		// log.Printf("[WARN] IAM Role (%s) not found, removing from state", d.Id())
		// d.SetId("")
		// return diags
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

	role := outputRaw.(*iam.Role)

	// occasionally, immediately after a role is created, AWS will give an ARN like AROAQ7SSZBKHREXAMPLE (unique ID)
	if role, err = waitRoleARNIsNotUniqueID(ctx, conn, state.ARN.ValueString(), role); err != nil {
		// TODO: have to update this error
		// return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): waiting for valid ARN: %s", d.Id(), err)
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionSetting, state.Name.String(), state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	// TODO: remove example section later
	// state.ApplicationAccount = flex.StringToFramework(ctx, out.ApplicationAccount)
	// state.ApplicationARN = flex.StringToFrameworkARN(ctx, out.ApplicationArn)
	// state.ApplicationProviderARN = flex.StringToFrameworkARN(ctx, out.ApplicationProviderArn)
	// state.Description = flex.StringToFramework(ctx, out.Description)
	// state.ID = flex.StringToFramework(ctx, out.ApplicationArn)
	// state.InstanceARN = flex.StringToFrameworkARN(ctx, out.InstanceArn)
	// state.Name = flex.StringToFramework(ctx, out.Name)
	// state.Status = flex.StringValueToFramework(ctx, out.Status)

	state.ARN = flex.StringToFramework(ctx, role.Arn)
	state.CreateDate = flex.StringValueToFramework(ctx, role.CreateDate.Format(time.RFC3339))
	state.Path = flex.StringToFramework(ctx, role.Path)
	// TODO: add more of these when ready to actually test

	// d.Set("description", role.Description)
	// d.Set("max_session_duration", role.MaxSessionDuration)
	// d.Set("name", role.RoleName)
	// d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(role.RoleName)))

	// if role.PermissionsBoundary != nil {
	// d.Set("permissions_boundary", role.PermissionsBoundary.PermissionsBoundaryArn)
	// } else {
	// d.Set("permissions_boundary", nil)
	// }
	// d.Set("unique_id", role.RoleId)

	// assumeRolePolicy, err := url.QueryUnescape(aws.StringValue(role.AssumeRolePolicyDocument))
	// if err != nil {
	// return sdkdiag.AppendFromErr(diags, err)
	// }

	// policyToSet, err := verify.PolicyToSet(d.Get("assume_role_policy").(string), assumeRolePolicy)
	// if err != nil {
	// return sdkdiag.AppendFromErr(diags, err)
	// }

	// d.Set("assume_role_policy", policyToSet)

	// inlinePolicies, err := readRoleInlinePolicies(ctx, aws.StringValue(role.RoleName), meta)
	// if err != nil {
	// return sdkdiag.AppendErrorf(diags, "reading inline policies for IAM role %s, error: %s", d.Id(), err)
	// }

	// var configPoliciesList []*iam.PutRolePolicyInput
	// if v := d.Get("inline_policy").(*schema.Set); v.Len() > 0 {
	// configPoliciesList = expandRoleInlinePolicies(aws.StringValue(role.RoleName), v.List())
	// }

	// if !inlinePoliciesEquivalent(inlinePolicies, configPoliciesList) {
	// if err := d.Set("inline_policy", flattenRoleInlinePolicies(inlinePolicies)); err != nil {
	// return sdkdiag.AppendErrorf(diags, "setting inline_policy: %s", err)
	// }
	// }

	// policyARNs, err := findRoleAttachedPolicies(ctx, conn, d.Id())
	// if err != nil {
	// return sdkdiag.AppendErrorf(diags, "reading IAM Policies attached to Role (%s): %s", d.Id(), err)
	// }
	// d.Set("managed_policy_arns", policyARNs)

	setTagsOut(ctx, role.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r resourceIamRole) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: finish this in later test
}

// TODO: import state?
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import

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

func DeleteRole(ctx context.Context, conn *iam.IAM, roleName string, forceDetach, hasInline, hasManaged bool) error {
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
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.DeleteRoleWithContext(ctx, input)
	}, iam.ErrCodeDeleteConflictException)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	return err
}

func deleteRoleInstanceProfiles(ctx context.Context, conn *iam.IAM, roleName string) error {
	instanceProfiles, err := findInstanceProfilesForRole(ctx, conn, roleName)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading IAM Instance Profiles for Role (%s): %w", roleName, err)
	}

	var errs []error

	for _, instanceProfile := range instanceProfiles {
		instanceProfileName := aws.StringValue(instanceProfile.InstanceProfileName)
		input := &iam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: aws.String(instanceProfileName),
			RoleName:            aws.String(roleName),
		}

		_, err := conn.RemoveRoleFromInstanceProfileWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("removing IAM Role (%s) from Instance Profile (%s): %w", roleName, instanceProfileName, err))
		}
	}

	return errors.Join(errs...)
}

func findRoleAttachedPolicies(ctx context.Context, conn *iam.IAM, roleName string) ([]string, error) {
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}
	var output []string

	err := conn.ListAttachedRolePoliciesPagesWithContext(ctx, input, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AttachedPolicies {
			if v != nil {
				output = append(output, aws.StringValue(v.PolicyArn))
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func deleteRolePolicyAttachments(ctx context.Context, conn *iam.IAM, roleName string, policyARNs []string) error {
	var errs []error

	for _, policyARN := range policyARNs {
		input := &iam.DetachRolePolicyInput{
			PolicyArn: aws.String(policyARN),
			RoleName:  aws.String(roleName),
		}

		_, err := conn.DetachRolePolicyWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("detaching IAM Policy (%s) from Role (%s): %w", policyARN, roleName, err))
		}
	}

	return errors.Join(errs...)
}

func findRolePolicyNames(ctx context.Context, conn *iam.IAM, roleName string) ([]string, error) {
	input := &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	}
	var output []string

	err := conn.ListRolePoliciesPagesWithContext(ctx, input, func(page *iam.ListRolePoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.PolicyNames {
			if v != nil {
				output = append(output, aws.StringValue(v))
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func deleteRoleInlinePolicies(ctx context.Context, conn *iam.IAM, roleName string, policyNames []string) error {
	var errs []error

	for _, policyName := range policyNames {
		if len(policyName) == 0 {
			continue
		}

		input := &iam.DeleteRolePolicyInput{
			PolicyName: aws.String(policyName),
			RoleName:   aws.String(roleName),
		}

		_, err := conn.DeleteRolePolicyWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("deleting IAM Role (%s) policy (%s): %w", roleName, policyName, err))
		}
	}

	return errors.Join(errs...)
}
