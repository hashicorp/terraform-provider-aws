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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	policyNameMaxLen       = 128
	policyNamePrefixMaxLen = policyNameMaxLen - resource.UniqueIDSuffixLength
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
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
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validResourceName(policyNamePrefixMaxLen),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	request := &iam.CreatePolicyInput{
		Description:    aws.String(d.Get("description").(string)),
		Path:           aws.String(d.Get("path").(string)),
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
	}

	if len(tags) > 0 {
		request.Tags = Tags(tags.IgnoreAWS())
	}

	response, err := conn.CreatePolicyWithContext(ctx, request)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if request.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating IAM Policy (%s) with tags: %s. Trying create without tags.", name, err)
		request.Tags = nil

		response, err = conn.CreatePolicyWithContext(ctx, request)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Policy %s: %s", name, err)
	}

	d.SetId(aws.StringValue(response.Policy.Arn))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if request.Tags == nil && len(tags) > 0 {
		err := policyUpdateTags(ctx, conn, d.Id(), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for IAM Policy (%s): %s", d.Id(), err)
			return append(diags, resourcePolicyRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding tags after create for IAM Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &iam.GetPolicyInput{
		PolicyArn: aws.String(d.Id()),
	}

	// Handle IAM eventual consistency
	var getPolicyResponse *iam.GetPolicyOutput
	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error
		getPolicyResponse, err = conn.GetPolicyWithContext(ctx, input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getPolicyResponse, err = conn.GetPolicyWithContext(ctx, input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM policy %s: %s", d.Id(), err)
	}

	if !d.IsNewResource() && (getPolicyResponse == nil || getPolicyResponse.Policy == nil) {
		log.Printf("[WARN] IAM Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	policy := getPolicyResponse.Policy

	d.Set("arn", policy.Arn)
	d.Set("description", policy.Description)
	d.Set("name", policy.PolicyName)
	d.Set("path", policy.Path)
	d.Set("policy_id", policy.PolicyId)

	tags := KeyValueTags(policy.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	// Retrieve policy

	getPolicyVersionRequest := &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(d.Id()),
		VersionId: policy.DefaultVersionId,
	}

	// Handle IAM eventual consistency
	var getPolicyVersionResponse *iam.GetPolicyVersionOutput
	err = resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error
		getPolicyVersionResponse, err = conn.GetPolicyVersionWithContext(ctx, getPolicyVersionRequest)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getPolicyVersionResponse, err = conn.GetPolicyVersionWithContext(ctx, getPolicyVersionRequest)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Policy (%s) version (%s) not found, removing from state", d.Id(), aws.StringValue(policy.DefaultVersionId))
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM policy version %s: %s", d.Id(), err)
	}

	var policyDocument string
	if getPolicyVersionResponse != nil && getPolicyVersionResponse.PolicyVersion != nil {
		var err error
		policyDocument, err = url.QueryUnescape(aws.StringValue(getPolicyVersionResponse.PolicyVersion.Document))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing IAM policy (%s) document: %s", d.Id(), err)
		}
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
	conn := meta.(*conns.AWSClient).IAMConn()

	if d.HasChangesExcept("tags", "tags_all") {
		if err := policyPruneVersions(ctx, d.Id(), conn); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM policy %s: pruning versions: %s", d.Id(), err)
		}

		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
		}

		request := &iam.CreatePolicyVersionInput{
			PolicyArn:      aws.String(d.Id()),
			PolicyDocument: aws.String(policy),
			SetAsDefault:   aws.Bool(true),
		}

		if _, err := conn.CreatePolicyVersionWithContext(ctx, request); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM policy %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := policyUpdateTags(ctx, conn, d.Id(), o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed updating tags for IAM Policy (%s): %s", d.Id(), err)
			return append(diags, resourcePolicyRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for IAM Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	if err := policyDeleteNonDefaultVersions(ctx, d.Id(), conn); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM policy (%s): deleting non-default versions: %s", d.Id(), err)
	}

	request := &iam.DeletePolicyInput{
		PolicyArn: aws.String(d.Id()),
	}

	_, err := conn.DeletePolicyWithContext(ctx, request)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM policy (%s): %s", d.Id(), err)
	}

	return diags
}

// policyPruneVersions deletes the oldest versions.
//
// Old versions are deleted until there are 4 or less remaining, which means at
// least one more can be created before hitting the maximum of 5.
//
// The default version is never deleted.

func policyPruneVersions(ctx context.Context, arn string, conn *iam.IAM) error {
	versions, err := policyListVersions(ctx, arn, conn)
	if err != nil {
		return err
	}
	if len(versions) < 5 {
		return nil
	}

	var oldestVersion *iam.PolicyVersion

	for _, version := range versions {
		if *version.IsDefaultVersion {
			continue
		}
		if oldestVersion == nil ||
			version.CreateDate.Before(*oldestVersion.CreateDate) {
			oldestVersion = version
		}
	}

	return policyDeleteVersion(ctx, arn, aws.StringValue(oldestVersion.VersionId), conn)
}

func policyDeleteNonDefaultVersions(ctx context.Context, arn string, conn *iam.IAM) error {
	versions, err := policyListVersions(ctx, arn, conn)
	if err != nil {
		return err
	}

	for _, version := range versions {
		if aws.BoolValue(version.IsDefaultVersion) {
			continue
		}
		if err := policyDeleteVersion(ctx, arn, aws.StringValue(version.VersionId), conn); err != nil {
			return err
		}
	}

	return nil
}

func policyDeleteVersion(ctx context.Context, arn, versionID string, conn *iam.IAM) error {
	request := &iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(versionID),
	}

	_, err := conn.DeletePolicyVersionWithContext(ctx, request)
	if err != nil {
		return fmt.Errorf("deleting policy version (%s): %w", versionID, err)
	}
	return nil
}

func policyListVersions(ctx context.Context, arn string, conn *iam.IAM) ([]*iam.PolicyVersion, error) {
	request := &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(arn),
	}

	response, err := conn.ListPolicyVersionsWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("Error listing versions for IAM policy %s: %w", arn, err)
	}
	return response.Versions, nil
}
