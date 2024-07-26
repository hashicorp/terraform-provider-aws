// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	roleNameMaxLen       = 64
	roleNamePrefixMaxLen = roleNameMaxLen - id.UniqueIDSuffixLength
)

// @SDKResource("aws_iam_role", name="Role")
// @Tags(identifierAttribute="id", resourceType="Role")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/iam/types;types.Role")
func resourceRole() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRoleCreate,
		ReadWithoutTimeout:   resourceRoleRead,
		UpdateWithoutTimeout: resourceRoleUpdate,
		DeleteWithoutTimeout: resourceRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRoleImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assume_role_policy": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          verify.ValidIAMPolicyJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1000),
					validation.StringDoesNotMatch(regexache.MustCompile("[“‘]"), "cannot contain specially formatted single or double quotes: [“‘]"),
					validation.StringMatch(regexache.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), `must satisfy regular expression pattern: [\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*)`),
				),
			},
			"force_detach_policies": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"inline_policy": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true, // semantically required but syntactically optional to allow empty inline_policy
							ValidateFunc: validation.All(
								validation.StringIsNotEmpty,
								validRolePolicyName,
							),
						},
						names.AttrPolicy: {
							Type:                  schema.TypeString,
							Optional:              true, // semantically required but syntactically optional to allow empty inline_policy
							ValidateFunc:          verify.ValidIAMPolicyJSON,
							DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
							DiffSuppressOnRefresh: true,
							StateFunc: func(v interface{}) string {
								json, _ := verify.LegacyPolicyNormalize(v)
								return json
							},
						},
					},
				},
				DiffSuppressFunc: func(k, _, _ string, d *schema.ResourceData) bool {
					if d.Id() == "" {
						return false
					}

					return !inlinePoliciesActualDiff(d)
				},
			},
			"managed_policy_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"max_session_duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(3600, 43200),
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validResourceName(roleNameMaxLen),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validResourceName(roleNamePrefixMaxLen),
			},
			names.AttrPath: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "/",
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"permissions_boundary": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	assumeRolePolicy, err := structure.NormalizeJsonString(d.Get("assume_role_policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "assume_role_policy (%s) is invalid JSON: %s", assumeRolePolicy, err)
	}

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
		Path:                     aws.String(d.Get(names.AttrPath).(string)),
		RoleName:                 aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_session_duration"); ok {
		input.MaxSessionDuration = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("permissions_boundary"); ok {
		input.PermissionsBoundary = aws.String(v.(string))
	}

	output, err := retryCreateRole(ctx, conn, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	partition := meta.(*conns.AWSClient).Partition
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = retryCreateRole(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Role (%s): %s", name, err)
	}

	roleName := aws.ToString(output.Role.RoleName)

	if v, ok := d.GetOk("inline_policy"); ok && v.(*schema.Set).Len() > 0 {
		policies := expandRoleInlinePolicies(roleName, v.(*schema.Set).List())
		if err := addRoleInlinePolicies(ctx, conn, policies); err != nil {
			derr := deleteRole(ctx, conn, roleName, true, true, false)
			if derr != nil {
				return sdkdiag.AppendErrorf(diags, "creating IAM role (%s), inline policy failed (%s), deleting role: %s", d.Id(), err, derr)
			}

			return sdkdiag.AppendErrorf(diags, "creating IAM Role (%s): %s", name, err)
		}
	}

	if v, ok := d.GetOk("managed_policy_arns"); ok && v.(*schema.Set).Len() > 0 {
		managedPolicies := flex.ExpandStringSet(v.(*schema.Set))
		if err := addRoleManagedPolicies(ctx, conn, roleName, managedPolicies); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IAM Role (%s): %s", name, err)
		}
	}

	d.SetId(roleName)

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := roleCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceRoleRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM Role (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRoleRead(ctx, d, meta)...)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findRoleByName(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Role (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): %s", d.Id(), err)
	}

	role := outputRaw.(*awstypes.Role)

	// occasionally, immediately after a role is created, AWS will give an ARN like AROAQ7SSZBKHREXAMPLE (unique ID)
	if role, err = waitRoleARNIsNotUniqueID(ctx, conn, d.Id(), role); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): waiting for valid ARN: %s", d.Id(), err)
	}

	d.Set(names.AttrARN, role.Arn)
	d.Set("create_date", role.CreateDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, role.Description)
	d.Set("max_session_duration", role.MaxSessionDuration)
	d.Set(names.AttrName, role.RoleName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(role.RoleName)))
	d.Set(names.AttrPath, role.Path)
	if role.PermissionsBoundary != nil {
		d.Set("permissions_boundary", role.PermissionsBoundary.PermissionsBoundaryArn)
	} else {
		d.Set("permissions_boundary", nil)
	}
	d.Set("unique_id", role.RoleId)

	assumeRolePolicy, err := url.QueryUnescape(aws.ToString(role.AssumeRolePolicyDocument))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("assume_role_policy").(string), assumeRolePolicy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("assume_role_policy", policyToSet)

	inlinePolicies, err := readRoleInlinePolicies(ctx, conn, aws.ToString(role.RoleName))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading inline policies for IAM role %s, error: %s", d.Id(), err)
	}

	var configPoliciesList []*iam.PutRolePolicyInput
	if v := d.Get("inline_policy").(*schema.Set); v.Len() > 0 {
		configPoliciesList = expandRoleInlinePolicies(aws.ToString(role.RoleName), v.List())
	}

	if !inlinePoliciesEquivalent(inlinePolicies, configPoliciesList) {
		if err := d.Set("inline_policy", flattenRoleInlinePolicies(inlinePolicies)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting inline_policy: %s", err)
		}
	}

	policyARNs, err := findRoleAttachedPolicies(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policies attached to Role (%s): %s", d.Id(), err)
	}
	d.Set("managed_policy_arns", policyARNs)

	setTagsOut(ctx, role.Tags)

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if d.HasChange("assume_role_policy") {
		assumeRolePolicy, err := structure.NormalizeJsonString(d.Get("assume_role_policy").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "assume_role_policy (%s) is invalid JSON: %s", assumeRolePolicy, err)
		}

		input := &iam.UpdateAssumeRolePolicyInput{
			RoleName:       aws.String(d.Id()),
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
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s) assume role policy: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrDescription) {
		input := &iam.UpdateRoleDescriptionInput{
			RoleName:    aws.String(d.Id()),
			Description: aws.String(d.Get(names.AttrDescription).(string)),
		}

		_, err := conn.UpdateRoleDescription(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s) description: %s", d.Id(), err)
		}
	}

	if d.HasChange("max_session_duration") {
		input := &iam.UpdateRoleInput{
			RoleName:           aws.String(d.Id()),
			MaxSessionDuration: aws.Int32(int32(d.Get("max_session_duration").(int))),
		}

		_, err := conn.UpdateRole(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s) MaxSessionDuration: %s", d.Id(), err)
		}
	}

	if d.HasChange("permissions_boundary") {
		permissionsBoundary := d.Get("permissions_boundary").(string)
		if permissionsBoundary != "" {
			input := &iam.PutRolePermissionsBoundaryInput{
				PermissionsBoundary: aws.String(permissionsBoundary),
				RoleName:            aws.String(d.Id()),
			}

			_, err := conn.PutRolePermissionsBoundary(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s) permissions boundary: %s", d.Id(), err)
			}
		} else {
			input := &iam.DeleteRolePermissionsBoundaryInput{
				RoleName: aws.String(d.Id()),
			}

			_, err := conn.DeleteRolePermissionsBoundary(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting IAM Role (%s) permissions boundary: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("inline_policy") && inlinePoliciesActualDiff(d) {
		roleName := d.Get(names.AttrName).(string)

		o, n := d.GetChange("inline_policy")

		if o == nil {
			o = new(schema.Set)
		}

		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		remove := os.Difference(ns).List()
		add := ns.Difference(os).List()

		var policyNames []string
		for _, policy := range remove {
			tfMap, ok := policy.(map[string]interface{})

			if !ok {
				continue
			}

			if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
				policyNames = append(policyNames, tfMap[names.AttrName].(string))
			}
		}
		if err := deleteRoleInlinePolicies(ctx, conn, roleName, policyNames); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s): %s", d.Id(), err)
		}

		policies := expandRoleInlinePolicies(roleName, add)
		if err := addRoleInlinePolicies(ctx, conn, policies); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("managed_policy_arns") {
		o, n := d.GetChange("managed_policy_arns")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := flex.ExpandStringSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

		if err := deleteRolePolicyAttachments(ctx, conn, d.Id(), del); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s): %s", d.Id(), err)
		}

		if err := addRoleManagedPolicies(ctx, conn, d.Id(), add); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRoleRead(ctx, d, meta)...)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	hasInline := false
	if v, ok := d.GetOk("inline_policy"); ok && v.(*schema.Set).Len() > 0 {
		hasInline = true
	}

	hasManaged := false
	if v, ok := d.GetOk("managed_policy_arns"); ok && v.(*schema.Set).Len() > 0 {
		hasManaged = true
	}

	err := deleteRole(ctx, conn, d.Id(), d.Get("force_detach_policies").(bool), hasInline, hasManaged)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Role (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceRoleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("force_detach_policies", false)
	return []*schema.ResourceData{d}, nil
}

func deleteRole(ctx context.Context, conn *iam.Client, roleName string, forceDetach, hasInline, hasManaged bool) error {
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

func flattenRoleInlinePolicy(apiObject *iam.PutRolePolicyInput) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrName] = aws.ToString(apiObject.PolicyName)
	tfMap[names.AttrPolicy] = aws.ToString(apiObject.PolicyDocument)

	return tfMap
}

func flattenRoleInlinePolicies(apiObjects []*iam.PutRolePolicyInput) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenRoleInlinePolicy(apiObject))
	}

	return tfList
}

func expandRoleInlinePolicy(roleName string, tfMap map[string]interface{}) *iam.PutRolePolicyInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &iam.PutRolePolicyInput{}

	namePolicy := false

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.PolicyName = aws.String(v)
		namePolicy = true
	}

	if v, ok := tfMap[names.AttrPolicy].(string); ok && v != "" {
		apiObject.PolicyDocument = aws.String(v)
		namePolicy = true
	}

	if namePolicy {
		apiObject.RoleName = aws.String(roleName)
	}

	return apiObject
}

func expandRoleInlinePolicies(roleName string, tfList []interface{}) []*iam.PutRolePolicyInput {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*iam.PutRolePolicyInput

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandRoleInlinePolicy(roleName, tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func addRoleInlinePolicies(ctx context.Context, conn *iam.Client, policies []*iam.PutRolePolicyInput) error {
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

func addRoleManagedPolicies(ctx context.Context, conn *iam.Client, roleName string, policies []*string) error {
	var errsList []error

	for _, arn := range policies {
		if err := attachPolicyToRole(ctx, conn, roleName, aws.ToString(arn)); err != nil {
			errsList = append(errsList, err)
		}
	}

	return errors.Join(errsList...)
}

func readRoleInlinePolicies(ctx context.Context, conn *iam.Client, roleName string) ([]*iam.PutRolePolicyInput, error) {
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

func inlinePoliciesActualDiff(d *schema.ResourceData) bool {
	roleName := d.Get(names.AttrName).(string)
	o, n := d.GetChange("inline_policy")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}

	os := o.(*schema.Set)
	ns := n.(*schema.Set)

	osPolicies := expandRoleInlinePolicies(roleName, os.List())
	nsPolicies := expandRoleInlinePolicies(roleName, ns.List())

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

func roleTags(ctx context.Context, conn *iam.Client, identifier string) ([]awstypes.Tag, error) {
	output, err := conn.ListRoleTags(ctx, &iam.ListRoleTagsInput{
		RoleName: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}

	return output.Tags, nil
}
