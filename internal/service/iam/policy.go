// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	policyNameMaxLen       = 128
	policyNamePrefixMaxLen = policyNameMaxLen - id.UniqueIDSuffixLength
)

// @SDKResource("aws_iam_policy", name="Policy")
// @Tags(identifierAttribute="id", resourceType="Policy")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/iam.Policy")
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validResourceName(policyNameMaxLen),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validResourceName(policyNamePrefixMaxLen),
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
				ForceNew: true,
			},
			"policy": {
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
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &iam.CreatePolicyInput{
		Description:    aws.String(d.Get("description").(string)),
		Path:           aws.String(d.Get("path").(string)),
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	output, err := conn.CreatePolicyWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreatePolicyWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Policy.Arn))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := policyCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourcePolicyRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM Policy (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	type policyWithVersion struct {
		policy        *iam.Policy
		policyVersion *iam.PolicyVersion
	}
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		iamPolicy := &policyWithVersion{}

		if v, err := findPolicyByARN(ctx, conn, d.Id()); err == nil {
			iamPolicy.policy = v
		} else {
			return nil, err
		}

		if v, err := findPolicyVersion(ctx, conn, d.Id(), aws.StringValue(iamPolicy.policy.DefaultVersionId)); err == nil {
			iamPolicy.policyVersion = v
		} else {
			return nil, err
		}

		return iamPolicy, nil
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s): %s", d.Id(), err)
	}

	output := outputRaw.(*policyWithVersion)
	policy := output.policy

	d.Set("arn", policy.Arn)
	d.Set("description", policy.Description)
	d.Set("name", policy.PolicyName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(policy.PolicyName)))
	d.Set("path", policy.Path)
	d.Set("policy_id", policy.PolicyId)

	setTagsOut(ctx, policy.Tags)

	policyDocument, err := url.QueryUnescape(aws.StringValue(output.policyVersion.Document))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing IAM Policy (%s) document: %s", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), policyDocument)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		if err := policyPruneVersions(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
		}

		input := &iam.CreatePolicyVersionInput{
			PolicyArn:      aws.String(d.Id()),
			PolicyDocument: aws.String(policy),
			SetAsDefault:   aws.Bool(true),
		}

		_, err = conn.CreatePolicyVersionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	// Delete non-default policy versions.
	versions, err := findPolicyVersionsByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s) versions: %s", d.Id(), err)
	}

	for _, version := range versions {
		if aws.BoolValue(version.IsDefaultVersion) {
			continue
		}

		if err := policyDeleteVersion(ctx, conn, d.Id(), aws.StringValue(version.VersionId)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting IAM Policy: %s", d.Id())
	_, err = conn.DeletePolicyWithContext(ctx, &iam.DeletePolicyInput{
		PolicyArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Policy (%s): %s", d.Id(), err)
	}

	return diags
}

// policyPruneVersions deletes the oldest version.
//
// Old versions are deleted until there are 4 or less remaining, which means at
// least one more can be created before hitting the maximum of 5.
//
// The default version is never deleted.
func policyPruneVersions(ctx context.Context, conn *iam.IAM, arn string) error {
	versions, err := findPolicyVersionsByARN(ctx, conn, arn)

	if err != nil {
		return err
	}

	if len(versions) < 5 {
		return nil
	}

	var oldestVersion *iam.PolicyVersion

	for _, version := range versions {
		if aws.BoolValue(version.IsDefaultVersion) {
			continue
		}

		if oldestVersion == nil || version.CreateDate.Before(aws.TimeValue(oldestVersion.CreateDate)) {
			oldestVersion = version
		}
	}

	if oldestVersion == nil {
		return nil
	}

	return policyDeleteVersion(ctx, conn, arn, aws.StringValue(oldestVersion.VersionId))
}

func policyDeleteVersion(ctx context.Context, conn *iam.IAM, arn, versionID string) error {
	input := &iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(versionID),
	}

	_, err := conn.DeletePolicyVersionWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("deleting IAM Policy (%s) version (%s): %w", arn, versionID, err)
	}

	return nil
}

func findPolicyByARN(ctx context.Context, conn *iam.IAM, arn string) (*iam.Policy, error) {
	input := &iam.GetPolicyInput{
		PolicyArn: aws.String(arn),
	}

	output, err := conn.GetPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policy, nil
}

func findPolicyByTwoPartKey(ctx context.Context, conn *iam.IAM, name, pathPrefix string) (*iam.Policy, error) {
	input := &iam.ListPoliciesInput{}
	if pathPrefix != "" {
		input.PathPrefix = aws.String(pathPrefix)
	}

	output, err := findPolicies(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if name != "" {
		output = slices.Filter(output, func(v *iam.Policy) bool {
			return aws.StringValue(v.PolicyName) == name
		})
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findPolicies(ctx context.Context, conn *iam.IAM, input *iam.ListPoliciesInput) ([]*iam.Policy, error) {
	var output []*iam.Policy

	err := conn.ListPoliciesPagesWithContext(ctx, input, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Policies {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findPolicyVersion(ctx context.Context, conn *iam.IAM, arn, versionID string) (*iam.PolicyVersion, error) {
	input := &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(versionID),
	}

	output, err := conn.GetPolicyVersionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PolicyVersion == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PolicyVersion, nil
}

func findPolicyVersionsByARN(ctx context.Context, conn *iam.IAM, arn string) ([]*iam.PolicyVersion, error) {
	input := &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(arn),
	}
	var output []*iam.PolicyVersion

	err := conn.ListPolicyVersionsPagesWithContext(ctx, input, func(page *iam.ListPolicyVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Versions {
			if v != nil {
				output = append(output, v)
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

func policyTags(ctx context.Context, conn *iam.IAM, identifier string) ([]*iam.Tag, error) {
	output, err := conn.ListPolicyTagsWithContext(ctx, &iam.ListPolicyTagsInput{
		PolicyArn: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}

	return output.Tags, nil
}
