// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	policyNameMaxLen       = 128
	policyNamePrefixMaxLen = policyNameMaxLen - id.UniqueIDSuffixLength
)

// @SDKResource("aws_iam_policy", name="Policy")
// @Tags(identifierAttribute="arn", resourceType="Policy")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/iam/types;types.Policy")
// @ArnIdentity
// @Testing(preIdentityVersion="v6.4.0")
// @Testing(existsTakesT=false, destroyTakesT=false)
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachment_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"delay_after_policy_creation_in_ms": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validResourceName(policyNameMaxLen),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validResourceName(policyNamePrefixMaxLen),
			},
			names.AttrPath: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "/",
				ForceNew:         true,
				ValidateDiagFunc: validPolicyPath,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          verify.ValidIAMPolicyJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	name := create.Name(ctx, d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := iam.CreatePolicyInput{
		Description:    aws.String(d.Get(names.AttrDescription).(string)),
		Path:           aws.String(d.Get(names.AttrPath).(string)),
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	output, err := conn.CreatePolicy(ctx, &input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	partition := meta.(*conns.AWSClient).Partition(ctx)
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreatePolicy(ctx, &input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Policy.Arn))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := policyCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]any)) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourcePolicyRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM Policy (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	type policyWithVersion struct {
		policy        *awstypes.Policy
		policyVersion *awstypes.PolicyVersion
	}
	output, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func(ctx context.Context) (*policyWithVersion, error) {
		iamPolicy := &policyWithVersion{}

		if v, err := findPolicyByARN(ctx, conn, d.Id()); err == nil {
			iamPolicy.policy = v
		} else {
			return nil, err
		}

		if v, err := findPolicyVersionByTwoPartKey(ctx, conn, d.Id(), aws.ToString(iamPolicy.policy.DefaultVersionId)); err == nil {
			iamPolicy.policyVersion = v
		} else {
			return nil, err
		}

		return iamPolicy, nil
	}, d.IsNewResource())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IAM Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s): %s", d.Id(), err)
	}

	resourcePolicyFlatten(ctx, output.policy, d)

	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.policy.PolicyName)))

	if err := resourcePolicyFlattenPolicyDocument(aws.ToString(output.policyVersion.Document), d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if d.HasChangesExcept("delay_after_policy_creation_in_ms", names.AttrTags, names.AttrTagsAll) {
		if err := policyPruneVersions(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
		}

		if v, ok := d.GetOk("delay_after_policy_creation_in_ms"); ok {
			// Creating a policy and setting its version as default in a single operation can expose a brief interval where
			// valid STS tokens with attached Session Policies are rejected by AWS authorization servers that have
			// not received the new default policy version. Separating this into two distinct actions of creating a policy version,
			// pausing briefly, and then setting that to the default version can avoid this issue, and may be required
			// in environments with very high S3 IO loads.
			inputCPV := iam.CreatePolicyVersionInput{
				PolicyArn:      aws.String(d.Id()),
				PolicyDocument: aws.String(policy),
				SetAsDefault:   false,
			}

			output, err := conn.CreatePolicyVersion(ctx, &inputCPV)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "creating IAM Policy (%s) version: %s", d.Id(), err)
			}

			versionID := aws.ToString(output.PolicyVersion.VersionId)

			time.Sleep(time.Duration(v.(int)) * time.Millisecond) //nolint:durationcheck // OK

			inputSDPV := iam.SetDefaultPolicyVersionInput{
				PolicyArn: aws.String(d.Id()),
				VersionId: aws.String(versionID),
			}

			_, err = conn.SetDefaultPolicyVersion(ctx, &inputSDPV)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting IAM Policy (%s) default version (%s): %s", d.Id(), versionID, err)
			}
		} else {
			input := iam.CreatePolicyVersionInput{
				PolicyArn:      aws.String(d.Id()),
				PolicyDocument: aws.String(policy),
				SetAsDefault:   true,
			}
			_, err = conn.CreatePolicyVersion(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating IAM Policy (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	// Delete non-default policy versions.
	versions, err := findPolicyVersionsByARN(ctx, conn, d.Id())

	if retry.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s) versions: %s", d.Id(), err)
	}

	for _, version := range versions {
		if version.IsDefaultVersion {
			continue
		}

		if err := policyDeleteVersion(ctx, conn, d.Id(), aws.ToString(version.VersionId)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting IAM Policy: %s", d.Id())
	input := iam.DeletePolicyInput{
		PolicyArn: aws.String(d.Id()),
	}
	_, err = conn.DeletePolicy(ctx, &input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func resourcePolicyFlatten(ctx context.Context, policy *awstypes.Policy, d *schema.ResourceData) {
	d.Set(names.AttrARN, policy.Arn)
	d.Set("attachment_count", policy.AttachmentCount)
	d.Set(names.AttrDescription, policy.Description)
	d.Set(names.AttrName, policy.PolicyName)
	d.Set(names.AttrPath, policy.Path)
	d.Set("policy_id", policy.PolicyId)

	setTagsOut(ctx, policy.Tags)
}

func resourcePolicyFlattenPolicyDocument(policyDocument string, d *schema.ResourceData) error {
	policyDocument, err := url.QueryUnescape(policyDocument)
	if err != nil {
		return fmt.Errorf("parsing IAM Policy (%s) document: %w", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), policyDocument)
	if err != nil {
		return fmt.Errorf("parsing IAM Policy (%s) document: %w", d.Id(), err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return nil
}

// policyPruneVersions deletes the oldest version.
//
// Old versions are deleted until there are 4 or less remaining, which means at
// least one more can be created before hitting the maximum of 5.
//
// The default version is never deleted.
func policyPruneVersions(ctx context.Context, conn *iam.Client, arn string) error {
	versions, err := findPolicyVersionsByARN(ctx, conn, arn)

	if err != nil {
		return err
	}

	if len(versions) < 5 {
		return nil
	}

	var oldestVersion awstypes.PolicyVersion

	for _, version := range versions {
		if version.IsDefaultVersion {
			continue
		}

		if oldestVersion == (awstypes.PolicyVersion{}) || version.CreateDate.Before(aws.ToTime(oldestVersion.CreateDate)) {
			oldestVersion = version
		}
	}

	if oldestVersion == (awstypes.PolicyVersion{}) {
		return nil
	}

	return policyDeleteVersion(ctx, conn, arn, aws.ToString(oldestVersion.VersionId))
}

func policyDeleteVersion(ctx context.Context, conn *iam.Client, arn, versionID string) error {
	input := iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(versionID),
	}

	_, err := conn.DeletePolicyVersion(ctx, &input)

	if err != nil {
		return fmt.Errorf("deleting IAM Policy (%s) version (%s): %w", arn, versionID, err)
	}

	return nil
}

func findPolicyByARN(ctx context.Context, conn *iam.Client, arn string) (*awstypes.Policy, error) {
	input := iam.GetPolicyInput{
		PolicyArn: aws.String(arn),
	}

	return findPolicy(ctx, conn, &input)
}

func findPolicy(ctx context.Context, conn *iam.Client, input *iam.GetPolicyInput) (*awstypes.Policy, error) {
	output, err := conn.GetPolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Policy, nil
}

func findPolicyByTwoPartKey(ctx context.Context, conn *iam.Client, name, pathPrefix string) (*awstypes.Policy, error) {
	var input iam.ListPoliciesInput
	if pathPrefix != "" {
		input.PathPrefix = aws.String(pathPrefix)
	}

	output, err := findPolicies(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if name != "" {
		output = slices.Filter(output, func(v awstypes.Policy) bool {
			return aws.ToString(v.PolicyName) == name
		})
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPolicies(ctx context.Context, conn *iam.Client, input *iam.ListPoliciesInput) ([]awstypes.Policy, error) {
	var output []awstypes.Policy

	pages := iam.NewListPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Policies {
			if p := &v; !inttypes.IsZero(p) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findPolicyVersionByTwoPartKey(ctx context.Context, conn *iam.Client, arn, versionID string) (*awstypes.PolicyVersion, error) {
	input := iam.GetPolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(versionID),
	}

	return findPolicyVersion(ctx, conn, &input)
}

func findPolicyVersion(ctx context.Context, conn *iam.Client, input *iam.GetPolicyVersionInput) (*awstypes.PolicyVersion, error) {
	output, err := conn.GetPolicyVersion(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PolicyVersion == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.PolicyVersion, nil
}

func findPolicyVersionsByARN(ctx context.Context, conn *iam.Client, arn string) ([]awstypes.PolicyVersion, error) {
	input := iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(arn),
	}

	return findPolicyVersions(ctx, conn, &input)
}

func findPolicyVersions(ctx context.Context, conn *iam.Client, input *iam.ListPolicyVersionsInput) ([]awstypes.PolicyVersion, error) {
	var output []awstypes.PolicyVersion

	pages := iam.NewListPolicyVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Versions {
			if p := &v; !inttypes.IsZero(p) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func policyTags(ctx context.Context, conn *iam.Client, identifier string, optFns ...func(*iam.Options)) ([]awstypes.Tag, error) {
	input := iam.ListPolicyTagsInput{
		PolicyArn: aws.String(identifier),
	}
	var output []awstypes.Tag

	pages := iam.NewListPolicyTagsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Tags...)
	}

	return output, nil
}
