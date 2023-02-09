package iam

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	roleNameMaxLen       = 64
	roleNamePrefixMaxLen = roleNameMaxLen - resource.UniqueIDSuffixLength
)

func ResourceRole() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRoleCreate,
		ReadWithoutTimeout:   resourceRoleRead,
		UpdateWithoutTimeout: resourceRoleUpdate,
		DeleteWithoutTimeout: resourceRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRoleImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assume_role_policy": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1000),
					validation.StringDoesNotMatch(regexp.MustCompile("[“‘]"), "cannot contain specially formatted single or double quotes: [“‘]"),
					validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), `must satisfy regular expression pattern: [\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*)`),
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
						"name": {
							Type:     schema.TypeString,
							Optional: true, // semantically required but syntactically optional to allow empty inline_policy
							ValidateFunc: validation.All(
								validation.StringIsNotEmpty,
								validRolePolicyName,
							),
						},
						"policy": {
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
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validResourceName(roleNameMaxLen),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validResourceName(roleNamePrefixMaxLen),
			},
			"path": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRoleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("force_detach_policies", false)
	return []*schema.ResourceData{d}, nil
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	assumeRolePolicy, err := structure.NormalizeJsonString(d.Get("assume_role_policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "assume_role_policy (%s) is invalid JSON: %s", assumeRolePolicy, err)
	}

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
		Path:                     aws.String(d.Get("path").(string)),
		RoleName:                 aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_session_duration"); ok {
		input.MaxSessionDuration = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("permissions_boundary"); ok {
		input.PermissionsBoundary = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := retryCreateRole(ctx, conn, input)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if input.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating IAM Role (%s) with tags: %s. Trying create without tags.", name, err)
		input.Tags = nil

		output, err = retryCreateRole(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Role (%s): %s", name, err)
	}

	roleName := aws.StringValue(output.Role.RoleName)

	if v, ok := d.GetOk("inline_policy"); ok && v.(*schema.Set).Len() > 0 {
		policies := expandRoleInlinePolicies(roleName, v.(*schema.Set).List())
		if err := addRoleInlinePolicies(ctx, policies, meta); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IAM Role (%s): %s", name, err)
		}
	}

	if v, ok := d.GetOk("managed_policy_arns"); ok && v.(*schema.Set).Len() > 0 {
		managedPolicies := flex.ExpandStringSet(v.(*schema.Set))
		if err := addRoleManagedPolicies(ctx, roleName, managedPolicies, meta); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IAM Role (%s): %s", name, err)
		}
	}

	d.SetId(roleName)

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if input.Tags == nil && len(tags) > 0 && meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID {
		err := roleUpdateTags(ctx, conn, d.Id(), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for IAM Role (%s): %s", d.Id(), err)
			return append(diags, resourceRoleRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding tags after create for IAM Role (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRoleRead(ctx, d, meta)...)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindRoleByName(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Role (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): %s", d.Id(), err)
	}

	role := outputRaw.(*iam.Role)

	// occasionally, immediately after a role is created, AWS will give an ARN like AROAQ7SSZBKHRKPWRZUN6 (unique ID)
	if role, err = waitRoleARNIsNotUniqueID(ctx, conn, d.Id(), role); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): waiting for valid ARN: %s", d.Id(), err)
	}

	d.Set("arn", role.Arn)
	d.Set("create_date", role.CreateDate.Format(time.RFC3339))
	d.Set("description", role.Description)
	d.Set("max_session_duration", role.MaxSessionDuration)
	d.Set("name", role.RoleName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(role.RoleName)))
	d.Set("path", role.Path)
	if role.PermissionsBoundary != nil {
		d.Set("permissions_boundary", role.PermissionsBoundary.PermissionsBoundaryArn)
	}
	d.Set("unique_id", role.RoleId)

	assumeRolePolicy, err := url.QueryUnescape(aws.StringValue(role.AssumeRolePolicyDocument))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("assume_role_policy").(string), assumeRolePolicy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role (%s): %s", d.Id(), err)
	}

	d.Set("assume_role_policy", policyToSet)

	inlinePolicies, err := readRoleInlinePolicies(ctx, aws.StringValue(role.RoleName), meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading inline policies for IAM role %s, error: %s", d.Id(), err)
	}

	var configPoliciesList []*iam.PutRolePolicyInput
	if v := d.Get("inline_policy").(*schema.Set); v.Len() > 0 {
		configPoliciesList = expandRoleInlinePolicies(aws.StringValue(role.RoleName), v.List())
	}

	if !inlinePoliciesEquivalent(inlinePolicies, configPoliciesList) {
		if err := d.Set("inline_policy", flattenRoleInlinePolicies(inlinePolicies)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting inline_policy: %s", err)
		}
	}

	managedPolicies, err := readRolePolicyAttachments(ctx, conn, aws.StringValue(role.RoleName))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading managed policies for IAM role %s, error: %s", d.Id(), err)
	}
	d.Set("managed_policy_arns", managedPolicies)

	tags := KeyValueTags(role.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

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
				return conn.UpdateAssumeRolePolicyWithContext(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, iam.ErrCodeMalformedPolicyDocumentException, "Invalid principal in policy") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s) assume role policy: %s", d.Id(), err)
		}
	}

	if d.HasChange("description") {
		input := &iam.UpdateRoleDescriptionInput{
			RoleName:    aws.String(d.Id()),
			Description: aws.String(d.Get("description").(string)),
		}

		_, err := conn.UpdateRoleDescriptionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s) description: %s", d.Id(), err)
		}
	}

	if d.HasChange("max_session_duration") {
		input := &iam.UpdateRoleInput{
			RoleName:           aws.String(d.Id()),
			MaxSessionDuration: aws.Int64(int64(d.Get("max_session_duration").(int))),
		}

		_, err := conn.UpdateRoleWithContext(ctx, input)

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

			_, err := conn.PutRolePermissionsBoundaryWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s) permissions boundary: %s", d.Id(), err)
			}
		} else {
			input := &iam.DeleteRolePermissionsBoundaryInput{
				RoleName: aws.String(d.Id()),
			}

			_, err := conn.DeleteRolePermissionsBoundaryWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting IAM Role (%s) permissions boundary: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("inline_policy") && inlinePoliciesActualDiff(d) {
		roleName := d.Get("name").(string)

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

		var policyNames []*string
		for _, policy := range remove {
			tfMap, ok := policy.(map[string]interface{})

			if !ok {
				continue
			}

			if v, ok := tfMap["name"].(string); ok && v != "" {
				policyNames = append(policyNames, aws.String(tfMap["name"].(string)))
			}
		}
		if err := deleteRoleInlinePolicies(ctx, conn, roleName, policyNames); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s): %s", d.Id(), err)
		}

		policies := expandRoleInlinePolicies(roleName, add)
		if err := addRoleInlinePolicies(ctx, policies, meta); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("managed_policy_arns") {
		roleName := d.Get("name").(string)

		o, n := d.GetChange("managed_policy_arns")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := flex.ExpandStringSet(os.Difference(ns))
		add := flex.ExpandStringSet(ns.Difference(os))

		if err := deleteRolePolicyAttachments(ctx, conn, roleName, remove); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s): %s", d.Id(), err)
		}

		if err := addRoleManagedPolicies(ctx, roleName, add, meta); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Role (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := roleUpdateTags(ctx, conn, d.Id(), o, n)

		// Some partitions may not support tagging, giving error
		if meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed updating tags for IAM Role %s: %s", d.Id(), err)
			return append(diags, resourceRoleRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for IAM Role (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRoleRead(ctx, d, meta)...)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	hasInline := false
	if v, ok := d.GetOk("inline_policy"); ok && v.(*schema.Set).Len() > 0 {
		hasInline = true
	}

	hasManaged := false
	if v, ok := d.GetOk("managed_policy_arns"); ok && v.(*schema.Set).Len() > 0 {
		hasManaged = true
	}

	err := DeleteRole(ctx, conn, d.Id(), d.Get("force_detach_policies").(bool), hasInline, hasManaged)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Role (%s): %s", d.Id(), err)
	}

	return diags
}

func DeleteRole(ctx context.Context, conn *iam.IAM, roleName string, forceDetach, hasInline, hasManaged bool) error {
	if err := deleteRoleInstanceProfiles(ctx, conn, roleName); err != nil {
		return fmt.Errorf("unable to detach instance profiles: %w", err)
	}

	if forceDetach || hasManaged {
		managedPolicies, err := readRolePolicyAttachments(ctx, conn, roleName)
		if err != nil {
			return err
		}

		if err := deleteRolePolicyAttachments(ctx, conn, roleName, managedPolicies); err != nil {
			return fmt.Errorf("unable to detach policies: %w", err)
		}
	}

	if forceDetach || hasInline {
		inlinePolicies, err := readRolePolicyNames(ctx, conn, roleName)
		if err != nil {
			return err
		}

		if err := deleteRoleInlinePolicies(ctx, conn, roleName, inlinePolicies); err != nil {
			return fmt.Errorf("unable to delete inline policies: %w", err)
		}
	}

	deleteRoleInput := &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	}
	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteRoleWithContext(ctx, deleteRoleInput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeDeleteConflictException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteRoleWithContext(ctx, deleteRoleInput)
	}

	return err
}

func deleteRoleInstanceProfiles(ctx context.Context, conn *iam.IAM, roleName string) error {
	resp, err := conn.ListInstanceProfilesForRoleWithContext(ctx, &iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	})
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}
	if err != nil {
		return err
	}

	// Loop and remove this Role from any Profiles
	for _, i := range resp.InstanceProfiles {
		input := &iam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: i.InstanceProfileName,
			RoleName:            aws.String(roleName),
		}

		_, err := conn.RemoveRoleFromInstanceProfileWithContext(ctx, input)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			return err
		}
	}

	return nil
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

func readRolePolicyAttachments(ctx context.Context, conn *iam.IAM, roleName string) ([]*string, error) {
	managedPolicies := make([]*string, 0)
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	err := conn.ListAttachedRolePoliciesPagesWithContext(ctx, input, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		for _, v := range page.AttachedPolicies {
			managedPolicies = append(managedPolicies, v.PolicyArn)
		}
		return !lastPage
	})
	if err != nil && !tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, err
	}

	return managedPolicies, nil
}

func deleteRolePolicyAttachments(ctx context.Context, conn *iam.IAM, roleName string, managedPolicies []*string) error {
	for _, arn := range managedPolicies {
		input := &iam.DetachRolePolicyInput{
			PolicyArn: arn,
			RoleName:  aws.String(roleName),
		}

		_, err := conn.DetachRolePolicyWithContext(ctx, input)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			return fmt.Errorf("detaching managed policy (%s): %w", aws.StringValue(arn), err)
		}
	}

	return nil
}

func readRolePolicyNames(ctx context.Context, conn *iam.IAM, roleName string) ([]*string, error) {
	inlinePolicies := make([]*string, 0)
	input := &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	err := conn.ListRolePoliciesPagesWithContext(ctx, input, func(page *iam.ListRolePoliciesOutput, lastPage bool) bool {
		inlinePolicies = append(inlinePolicies, page.PolicyNames...)
		return !lastPage
	})

	if err != nil && !tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, err
	}

	return inlinePolicies, nil
}

func deleteRoleInlinePolicies(ctx context.Context, conn *iam.IAM, roleName string, policyNames []*string) error {
	for _, name := range policyNames {
		if len(aws.StringValue(name)) == 0 {
			continue
		}

		input := &iam.DeleteRolePolicyInput{
			PolicyName: name,
			RoleName:   aws.String(roleName),
		}

		_, err := conn.DeleteRolePolicyWithContext(ctx, input)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("removing inline policy (%s): %w", aws.StringValue(name), err)
		}
	}

	return nil
}

func flattenRoleInlinePolicy(apiObject *iam.PutRolePolicyInput) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["name"] = aws.StringValue(apiObject.PolicyName)
	tfMap["policy"] = aws.StringValue(apiObject.PolicyDocument)

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

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.PolicyName = aws.String(v)
		namePolicy = true
	}

	if v, ok := tfMap["policy"].(string); ok && v != "" {
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

func addRoleInlinePolicies(ctx context.Context, policies []*iam.PutRolePolicyInput, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn()

	var errs *multierror.Error
	for _, policy := range policies {
		if len(aws.StringValue(policy.PolicyName)) == 0 || len(aws.StringValue(policy.PolicyDocument)) == 0 {
			continue
		}

		if _, err := conn.PutRolePolicyWithContext(ctx, policy); err != nil {
			newErr := fmt.Errorf("adding inline policy (%s): %w", aws.StringValue(policy.PolicyName), err)
			errs = multierror.Append(errs, newErr)
		}
	}

	return errs.ErrorOrNil()
}

func addRoleManagedPolicies(ctx context.Context, roleName string, policies []*string, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn()

	var errs *multierror.Error
	for _, arn := range policies {
		if err := attachPolicyToRole(ctx, conn, roleName, aws.StringValue(arn)); err != nil {
			newErr := fmt.Errorf("attaching managed policy (%s): %w", aws.StringValue(arn), err)
			errs = multierror.Append(errs, newErr)
		}
	}

	return errs.ErrorOrNil()
}

func readRoleInlinePolicies(ctx context.Context, roleName string, meta interface{}) ([]*iam.PutRolePolicyInput, error) {
	conn := meta.(*conns.AWSClient).IAMConn()

	policyNames, err := readRolePolicyNames(ctx, conn, roleName)
	if err != nil {
		return nil, err
	}

	var apiObjects []*iam.PutRolePolicyInput
	for _, policyName := range policyNames {
		policyResp, err := conn.GetRolePolicyWithContext(ctx, &iam.GetRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: policyName,
		})
		if err != nil {
			return nil, err
		}

		policy, err := url.QueryUnescape(*policyResp.PolicyDocument)
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
			PolicyName:     policyName,
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func inlinePoliciesActualDiff(d *schema.ResourceData) bool {
	roleName := d.Get("name").(string)
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
		if equivalent, err := awspolicy.PoliciesAreEquivalent(`{}`, aws.StringValue(configPolicies[0].PolicyDocument)); err == nil && equivalent {
			return true
		}
	}

	if len(readPolicies) != len(configPolicies) {
		return false
	}

	matches := 0

	for _, policyOne := range readPolicies {
		for _, policyTwo := range configPolicies {
			if aws.StringValue(policyOne.PolicyName) == aws.StringValue(policyTwo.PolicyName) {
				matches++
				if equivalent, err := awspolicy.PoliciesAreEquivalent(aws.StringValue(policyOne.PolicyDocument), aws.StringValue(policyTwo.PolicyDocument)); err != nil || !equivalent {
					return false
				}
				break
			}
		}
	}

	return matches == len(readPolicies)
}
