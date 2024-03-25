// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_saml_provider", name="SAML Provider")
// @Tags(identifierAttribute="id", resourceType="SAMLProvider")
// @Testing(tagsTest=false)
func resourceSAMLProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSAMLProviderCreate,
		ReadWithoutTimeout:   resourceSAMLProviderRead,
		UpdateWithoutTimeout: resourceSAMLProviderUpdate,
		DeleteWithoutTimeout: resourceSAMLProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"saml_metadata_document": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1000, 10000000),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"valid_until": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSAMLProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	name := d.Get("name").(string)
	input := &iam.CreateSAMLProviderInput{
		Name:                 aws.String(name),
		SAMLMetadataDocument: aws.String(d.Get("saml_metadata_document").(string)),
		Tags:                 getTagsIn(ctx),
	}

	output, err := conn.CreateSAMLProviderWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateSAMLProviderWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM SAML Provider (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.SAMLProviderArn))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := samlProviderCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceSAMLProviderRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM SAML Provider (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSAMLProviderRead(ctx, d, meta)...)
}

func resourceSAMLProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	output, err := findSAMLProviderByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
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

	d.Set("arn", d.Id())
	d.Set("name", name)
	d.Set("saml_metadata_document", output.SAMLMetadataDocument)
	if output.ValidUntil != nil {
		d.Set("valid_until", aws.TimeValue(output.ValidUntil).Format(time.RFC3339))
	} else {
		d.Set("valid_until", nil)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceSAMLProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &iam.UpdateSAMLProviderInput{
			SAMLProviderArn:      aws.String(d.Id()),
			SAMLMetadataDocument: aws.String(d.Get("saml_metadata_document").(string)),
		}

		_, err := conn.UpdateSAMLProviderWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM SAML Provider (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSAMLProviderRead(ctx, d, meta)...)
}

func resourceSAMLProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	log.Printf("[DEBUG] Deleting IAM SAML Provider: %s", d.Id())
	_, err := conn.DeleteSAMLProviderWithContext(ctx, &iam.DeleteSAMLProviderInput{
		SAMLProviderArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM SAML Provider (%s): %s", d.Id(), err)
	}

	return diags
}

func findSAMLProviderByARN(ctx context.Context, conn *iam.IAM, arn string) (*iam.GetSAMLProviderOutput, error) {
	input := &iam.GetSAMLProviderInput{
		SAMLProviderArn: aws.String(arn),
	}

	output, err := conn.GetSAMLProviderWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func nameFromSAMLProviderARN(v string) (string, error) {
	arn, err := arn.Parse(v)

	if err != nil {
		return "", fmt.Errorf("parsing IAM SAML Provider ARN (%s): %w", v, err)
	}

	return strings.TrimPrefix(arn.Resource, "saml-provider/"), nil
}
