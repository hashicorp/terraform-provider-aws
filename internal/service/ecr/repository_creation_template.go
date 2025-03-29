// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecr_repository_creation_template", name="Repository Creation Template")
func resourceRepositoryCreationTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryCreationTemplateCreate,
		ReadWithoutTimeout:   resourceRepositoryCreationTemplateRead,
		UpdateWithoutTimeout: resourceRepositoryCreationTemplateUpdate,
		DeleteWithoutTimeout: resourceRepositoryCreationTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"applied_for": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.RCTAppliedFor](),
				},
			},
			"custom_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			names.AttrEncryptionConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.EncryptionTypeAes256,
							ValidateDiagFunc: enum.Validate[types.EncryptionType](),
						},
						names.AttrKMSKey: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"image_tag_mutability": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.ImageTagMutabilityMutable,
				ValidateDiagFunc: enum.Validate[types.ImageTagMutability](),
			},
			"lifecycle_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalentLifecyclePolicyJSON(old, new)
					return equal
				},
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrPrefix: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 30),
					validation.StringMatch(
						regexache.MustCompile(`(?:ROOT|(?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)*[a-z0-9]+(?:[._-][a-z0-9]+)*)`),
						"must only include alphanumeric, underscore, period, hyphen, or slash characters, or be the string `ROOT`"),
				),
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_policy": {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrResourceTags: tftags.TagsSchema(),
		},
	}
}

func resourceRepositoryCreationTemplateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	prefix := d.Get(names.AttrPrefix).(string)
	input := &ecr.CreateRepositoryCreationTemplateInput{
		EncryptionConfiguration: expandRepositoryEncryptionConfigurationForRepositoryCreationTemplate(d.Get(names.AttrEncryptionConfiguration).([]any)),
		ImageTagMutability:      types.ImageTagMutability((d.Get("image_tag_mutability").(string))),
		Prefix:                  aws.String(prefix),
	}

	if v, ok := d.GetOk("applied_for"); ok && v.(*schema.Set).Len() > 0 {
		input.AppliedFor = flex.ExpandStringyValueSet[types.RCTAppliedFor](v.(*schema.Set))
	}

	if v, ok := d.GetOk("custom_role_arn"); ok {
		input.CustomRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("lifecycle_policy"); ok {
		policy, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.LifecyclePolicy = aws.String(policy)
	}

	if v, ok := d.GetOk("repository_policy"); ok {
		policy, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.RepositoryPolicy = aws.String(policy)
	}

	if v, ok := d.GetOk(names.AttrResourceTags); ok && len(v.(map[string]any)) > 0 {
		input.ResourceTags = svcTags(tftags.New(ctx, v.(map[string]any)))
	}

	output, err := conn.CreateRepositoryCreationTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Repository Creation Template (%s): %s", prefix, err)
	}

	d.SetId(aws.ToString(output.RepositoryCreationTemplate.Prefix))

	return append(diags, resourceRepositoryCreationTemplateRead(ctx, d, meta)...)
}

func resourceRepositoryCreationTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	rct, registryID, err := findRepositoryCreationTemplateByRepositoryPrefix(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Repository Creation Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Repository Creation Template (%s): %s", d.Id(), err)
	}

	d.Set("applied_for", rct.AppliedFor)
	d.Set("custom_role_arn", rct.CustomRoleArn)
	d.Set(names.AttrDescription, rct.Description)
	if err := d.Set(names.AttrEncryptionConfiguration, flattenRepositoryEncryptionConfigurationForRepositoryCreationTemplate(rct.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	d.Set("image_tag_mutability", rct.ImageTagMutability)

	if _, err := equivalentLifecyclePolicyJSON(d.Get("lifecycle_policy").(string), aws.ToString(rct.LifecyclePolicy)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err := structure.NormalizeJsonString(aws.ToString(rct.LifecyclePolicy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("lifecycle_policy", policyToSet)
	d.Set(names.AttrPrefix, rct.Prefix)
	d.Set("registry_id", registryID)

	policyToSet, err = verify.SecondJSONUnlessEquivalent(d.Get("repository_policy").(string), aws.ToString(rct.RepositoryPolicy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("repository_policy", policyToSet)
	d.Set(names.AttrResourceTags, keyValueTags(ctx, rct.ResourceTags).Map())

	return diags
}

func resourceRepositoryCreationTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	prefix := d.Get(names.AttrPrefix).(string)
	input := &ecr.UpdateRepositoryCreationTemplateInput{
		Prefix: aws.String(prefix),
	}

	if d.HasChange("applied_for") {
		if v, ok := d.GetOk("applied_for"); ok && v.(*schema.Set).Len() > 0 {
			input.AppliedFor = flex.ExpandStringyValueSet[types.RCTAppliedFor](v.(*schema.Set))
		}
	}

	if d.HasChange("custom_role_arn") {
		input.CustomRoleArn = aws.String(d.Get("custom_role_arn").(string))
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange(names.AttrEncryptionConfiguration) {
		input.EncryptionConfiguration = expandRepositoryEncryptionConfigurationForRepositoryCreationTemplate(d.Get(names.AttrEncryptionConfiguration).([]any))
	}

	if d.HasChange("image_tag_mutability") {
		input.ImageTagMutability = types.ImageTagMutability((d.Get("image_tag_mutability").(string)))
	}

	if d.HasChange("lifecycle_policy") {
		policy, err := structure.NormalizeJsonString(d.Get("lifecycle_policy").(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.LifecyclePolicy = aws.String(policy)
	}

	if d.HasChange("repository_policy") {
		policy, err := structure.NormalizeJsonString(d.Get("repository_policy").(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.RepositoryPolicy = aws.String(policy)
	}

	if d.HasChange(names.AttrResourceTags) {
		input.ResourceTags = svcTags(tftags.New(ctx, d.Get(names.AttrResourceTags).(map[string]any)))
	}

	_, err := conn.UpdateRepositoryCreationTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating ECR Repository Creation Template (%s): %s", prefix, err)
	}

	return append(diags, resourceRepositoryCreationTemplateRead(ctx, d, meta)...)
}

func resourceRepositoryCreationTemplateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	log.Printf("[DEBUG] Deleting ECR Repository Creation Template: %s", d.Id())
	_, err := conn.DeleteRepositoryCreationTemplate(ctx, &ecr.DeleteRepositoryCreationTemplateInput{
		Prefix: aws.String(d.Id()),
	})

	if errs.IsA[*types.TemplateNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Repository Creation Template (%s): %s", d.Id(), err)
	}

	return diags
}

func findRepositoryCreationTemplateByRepositoryPrefix(ctx context.Context, conn *ecr.Client, repositoryPrefix string) (*types.RepositoryCreationTemplate, *string, error) {
	input := &ecr.DescribeRepositoryCreationTemplatesInput{
		Prefixes: []string{repositoryPrefix},
	}

	return findRepositoryCreationTemplate(ctx, conn, input)
}

func findRepositoryCreationTemplate(ctx context.Context, conn *ecr.Client, input *ecr.DescribeRepositoryCreationTemplatesInput) (*types.RepositoryCreationTemplate, *string, error) {
	output, registryID, err := findRepositoryCreationTemplates(ctx, conn, input)

	if err != nil {
		return nil, nil, err
	}

	rct, err := tfresource.AssertSingleValueResult(output)

	return rct, registryID, err
}

func findRepositoryCreationTemplates(ctx context.Context, conn *ecr.Client, input *ecr.DescribeRepositoryCreationTemplatesInput) ([]types.RepositoryCreationTemplate, *string, error) {
	var (
		output     []types.RepositoryCreationTemplate
		registryID *string
	)

	pages := ecr.NewDescribeRepositoryCreationTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.TemplateNotFoundException](err) {
			return nil, nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, nil, err
		}

		output = append(output, page.RepositoryCreationTemplates...)
		registryID = page.RegistryId
	}

	return output, registryID, nil
}

func expandRepositoryEncryptionConfigurationForRepositoryCreationTemplate(data []any) *types.EncryptionConfigurationForRepositoryCreationTemplate {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	ec := data[0].(map[string]any)
	config := &types.EncryptionConfigurationForRepositoryCreationTemplate{
		EncryptionType: types.EncryptionType((ec["encryption_type"].(string))),
	}
	if v, ok := ec[names.AttrKMSKey]; ok {
		if s := v.(string); s != "" {
			config.KmsKey = aws.String(v.(string))
		}
	}
	return config
}

func flattenRepositoryEncryptionConfigurationForRepositoryCreationTemplate(ec *types.EncryptionConfigurationForRepositoryCreationTemplate) []map[string]any {
	if ec == nil {
		return nil
	}

	config := map[string]any{
		"encryption_type": ec.EncryptionType,
		names.AttrKMSKey:  aws.ToString(ec.KmsKey),
	}

	return []map[string]any{
		config,
	}
}
