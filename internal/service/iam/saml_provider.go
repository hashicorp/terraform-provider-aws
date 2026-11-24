// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_saml_provider", name="SAML Provider")
// @Tags(identifierAttribute="arn", resourceType="SAMLProvider")
// @Testing(tagsTest=false)
// @ArnIdentity
// @Testing(preIdentityVersion="v6.4.0")
func resourceSAMLProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSAMLProviderCreate,
		ReadWithoutTimeout:   resourceSAMLProviderRead,
		UpdateWithoutTimeout: resourceSAMLProviderUpdate,
		DeleteWithoutTimeout: resourceSAMLProviderDelete,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"context_provider": contextProviderSchema(),
				"identity_center_instance_arn": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidARN,
					RequiredWith: []string{"context_provider"},
				},
				"identity_store_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"saml_key_hashes": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"issuer_url": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"key_hashes": {
								Type:     schema.TypeList,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"saml_metadata_document": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1000, 10000000),
				},
				"saml_provider_uuid": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaOptionalComputedWarnShouldBeComputedOnly(),
				"valid_until": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func resourceSAMLProviderCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &iam.CreateSAMLProviderInput{
		Name:                 aws.String(name),
		SAMLMetadataDocument: aws.String(d.Get("saml_metadata_document").(string)),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("context_provider"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ContextProvider = expandContextProvider(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("identity_center_instance_arn"); ok {
		input.IdentityCenterInstanceArn = aws.String(v.(string))
	}

	output, err := conn.CreateSAMLProvider(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	partition := meta.(*conns.AWSClient).Partition(ctx)
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateSAMLProvider(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM SAML Provider (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.SAMLProviderArn))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := samlProviderCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]any)) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceSAMLProviderRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM SAML Provider (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSAMLProviderRead(ctx, d, meta)...)
}

func resourceSAMLProviderRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	output, err := findSAMLProviderByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IAM SAML Provider %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM SAML Provider (%s): %s", d.Id(), err)
	}

	name, err := nameFromSAMLProviderARN(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrARN, d.Id())
	d.Set(names.AttrName, name)
	if err := d.Set("context_provider", flattenContextProvider(output.ContextProvider)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting context_provider: %s", err)
	}
	d.Set("identity_center_instance_arn", output.IdentityCenterInstanceArn)
	d.Set("identity_store_arn", output.IdentityStoreArn)
	if err := d.Set("saml_key_hashes", flattenSAMLKeyHashEntries(output.SAMLKeyHashes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting saml_key_hashes: %s", err)
	}
	d.Set("saml_metadata_document", output.SAMLMetadataDocument)
	d.Set("saml_provider_uuid", output.SAMLProviderUUID)
	if output.ValidUntil != nil {
		d.Set("valid_until", aws.ToTime(output.ValidUntil).Format(time.RFC3339))
	} else {
		d.Set("valid_until", nil)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceSAMLProviderUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	// UpdateSAMLProvider can't change the metadata and the context provider in
	// one request, and the context provider can only be removed after the
	// IAM Identity Center instance has been disassociated.
	if d.HasChange("saml_metadata_document") {
		input := &iam.UpdateSAMLProviderInput{
			SAMLProviderArn:      aws.String(d.Id()),
			SAMLMetadataDocument: aws.String(d.Get("saml_metadata_document").(string)),
		}

		if _, err := conn.UpdateSAMLProvider(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM SAML Provider (%s): %s", d.Id(), err)
		}
	}

	contextProviderChanged := d.HasChange("context_provider")
	identityCenterChanged := d.HasChange("identity_center_instance_arn")
	removingIdentityCenter := identityCenterChanged && d.Get("identity_center_instance_arn").(string) == ""

	if contextProviderChanged && !removingIdentityCenter {
		if err := updateSAMLProviderContextProvider(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if identityCenterChanged {
		if err := updateSAMLProviderIdentityCenterInstance(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if contextProviderChanged && removingIdentityCenter {
		if err := updateSAMLProviderContextProvider(ctx, conn, d); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceSAMLProviderRead(ctx, d, meta)...)
}

func updateSAMLProviderContextProvider(ctx context.Context, conn *iam.Client, d *schema.ResourceData) error {
	input := &iam.UpdateSAMLProviderInput{
		SAMLProviderArn: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("context_provider"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ContextProvider = expandContextProvider(v.([]any)[0].(map[string]any))
	} else {
		input.RemoveContextProvider = true
	}

	if _, err := conn.UpdateSAMLProvider(ctx, input); err != nil {
		return fmt.Errorf("updating IAM SAML Provider (%s) context provider: %w", d.Id(), err)
	}

	return nil
}

func updateSAMLProviderIdentityCenterInstance(ctx context.Context, conn *iam.Client, d *schema.ResourceData) error {
	input := &iam.UpdateSAMLProviderInput{
		SAMLProviderArn: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("identity_center_instance_arn"); ok {
		input.IdentityCenterInstanceArn = aws.String(v.(string))
	} else {
		input.RemoveIdentityCenterInstance = true
	}

	if _, err := conn.UpdateSAMLProvider(ctx, input); err != nil {
		return fmt.Errorf("updating IAM SAML Provider (%s) IAM Identity Center instance: %w", d.Id(), err)
	}

	return nil
}

func resourceSAMLProviderDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	log.Printf("[DEBUG] Deleting IAM SAML Provider: %s", d.Id())
	_, err := conn.DeleteSAMLProvider(ctx, &iam.DeleteSAMLProviderInput{
		SAMLProviderArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM SAML Provider (%s): %s", d.Id(), err)
	}

	return diags
}

func findSAMLProviderByARN(ctx context.Context, conn *iam.Client, arn string) (*iam.GetSAMLProviderOutput, error) {
	input := &iam.GetSAMLProviderInput{
		SAMLProviderArn: aws.String(arn),
	}

	output, err := conn.GetSAMLProvider(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func samlProviderTags(ctx context.Context, conn *iam.Client, identifier string, optFns ...func(*iam.Options)) ([]awstypes.Tag, error) {
	input := iam.ListSAMLProviderTagsInput{
		SAMLProviderArn: aws.String(identifier),
	}
	var output []awstypes.Tag

	pages := iam.NewListSAMLProviderTagsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Tags...)
	}

	return output, nil
}

// contextProviderSchema returns the schema of the trusted identity propagation
// context provider block shared by SAML and OIDC providers.
func contextProviderSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key_definition": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 10,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"source_key": {
								Type:     schema.TypeString,
								Required: true,
							},
							"target_key": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"subject_source_key": {
					Type:     schema.TypeString,
					Required: true,
				},
				"subject_type_source_key": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"context_provider.0.subject_type_value"},
				},
				"subject_type_value": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"context_provider.0.subject_type_source_key"},
				},
			},
		},
	}
}

func expandContextProvider(tfMap map[string]any) *awstypes.ContextProvider {
	apiObject := &awstypes.ContextProvider{
		SubjectSourceKey: aws.String(tfMap["subject_source_key"].(string)),
	}

	if v, ok := tfMap["key_definition"].([]any); ok {
		for _, v := range v {
			tfMap, ok := v.(map[string]any)
			if !ok {
				continue
			}

			keyDefinition := awstypes.ContextProviderKeyDefinition{
				SourceKey: aws.String(tfMap["source_key"].(string)),
			}

			if v, ok := tfMap["target_key"].(string); ok && v != "" {
				keyDefinition.TargetKey = aws.String(v)
			}

			apiObject.KeyDefinitionList = append(apiObject.KeyDefinitionList, keyDefinition)
		}
	}

	if v, ok := tfMap["subject_type_source_key"].(string); ok && v != "" {
		apiObject.SubjectTypeSourceKey = aws.String(v)
	}

	if v, ok := tfMap["subject_type_value"].(string); ok && v != "" {
		apiObject.SubjectTypeValue = aws.String(v)
	}

	return apiObject
}

func flattenContextProvider(apiObject *awstypes.ContextProvider) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"subject_source_key":      aws.ToString(apiObject.SubjectSourceKey),
		"subject_type_source_key": aws.ToString(apiObject.SubjectTypeSourceKey),
		"subject_type_value":      aws.ToString(apiObject.SubjectTypeValue),
	}

	var keyDefinitions []any
	for _, v := range apiObject.KeyDefinitionList {
		keyDefinitions = append(keyDefinitions, map[string]any{
			"source_key": aws.ToString(v.SourceKey),
			"target_key": aws.ToString(v.TargetKey),
		})
	}
	tfMap["key_definition"] = keyDefinitions

	return []any{tfMap}
}

func flattenSAMLKeyHashEntries(apiObjects []awstypes.SAMLKeyHashEntry) []any {
	var tfList []any

	for _, v := range apiObjects {
		tfList = append(tfList, map[string]any{
			"issuer_url": aws.ToString(v.IssuerURL),
			"key_hashes": flex.FlattenStringValueList(v.KeyHashes),
		})
	}

	return tfList
}

func nameFromSAMLProviderARN(v string) (string, error) {
	arn, err := arn.Parse(v)

	if err != nil {
		return "", fmt.Errorf("parsing IAM SAML Provider ARN (%s): %w", v, err)
	}

	return strings.TrimPrefix(arn.Resource, "saml-provider/"), nil
}
