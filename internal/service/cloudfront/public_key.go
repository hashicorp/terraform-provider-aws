// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_public_key", name="Public Key")
func resourcePublicKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePublicKeyCreate,
		ReadWithoutTimeout:   resourcePublicKeyRead,
		UpdateWithoutTimeout: resourcePublicKeyUpdate,
		DeleteWithoutTimeout: resourcePublicKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrComment: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encoded_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validPublicKeyName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validPublicKeyNamePrefix,
			},
		},
	}
}

func resourcePublicKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrName).(string)),
		create.WithConfiguredPrefix(d.Get(names.AttrNamePrefix).(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()
	input := &cloudfront.CreatePublicKeyInput{
		PublicKeyConfig: &awstypes.PublicKeyConfig{
			EncodedKey: aws.String(d.Get("encoded_key").(string)),
			Name:       aws.String(name),
		},
	}

	if v, ok := d.GetOk("caller_reference"); ok {
		input.PublicKeyConfig.CallerReference = aws.String(v.(string))
	} else {
		input.PublicKeyConfig.CallerReference = aws.String(id.UniqueId())
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		input.PublicKeyConfig.Comment = aws.String(v.(string))
	}

	output, err := conn.CreatePublicKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Public Key (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.PublicKey.Id))

	return append(diags, resourcePublicKeyRead(ctx, d, meta)...)
}

func resourcePublicKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findPublicKeyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Public Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Public Key (%s): %s", d.Id(), err)
	}

	publicKeyConfig := output.PublicKey.PublicKeyConfig
	d.Set("caller_reference", publicKeyConfig.CallerReference)
	d.Set(names.AttrComment, publicKeyConfig.Comment)
	d.Set("encoded_key", publicKeyConfig.EncodedKey)
	d.Set("etag", output.ETag)
	d.Set(names.AttrName, publicKeyConfig.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(publicKeyConfig.Name)))

	return diags
}

func resourcePublicKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.UpdatePublicKeyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
		PublicKeyConfig: &awstypes.PublicKeyConfig{
			EncodedKey: aws.String(d.Get("encoded_key").(string)),
			Name:       aws.String(d.Get(names.AttrName).(string)),
		},
	}

	if v, ok := d.GetOk("caller_reference"); ok {
		input.PublicKeyConfig.CallerReference = aws.String(v.(string))
	} else {
		input.PublicKeyConfig.CallerReference = aws.String(id.UniqueId())
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		input.PublicKeyConfig.Comment = aws.String(v.(string))
	}

	_, err := conn.UpdatePublicKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Public Key (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePublicKeyRead(ctx, d, meta)...)
}

func resourcePublicKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	log.Printf("[DEBUG] Deleting CloudFront Public Key: %s", d.Id())
	_, err := conn.DeletePublicKey(ctx, &cloudfront.DeletePublicKeyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if errs.IsA[*awstypes.NoSuchPublicKey](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Public Key (%s): %s", d.Id(), err)
	}

	return diags
}

func findPublicKeyByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetPublicKeyOutput, error) {
	input := &cloudfront.GetPublicKeyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetPublicKey(ctx, input)

	if errs.IsA[*awstypes.NoSuchPublicKey](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PublicKey == nil || output.PublicKey.PublicKeyConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func validPublicKeyName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z_-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, underscores and hyphens allowed in %q", k))
	}
	if len(value) > 128 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 128 characters", k))
	}
	return
}

func validPublicKeyNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z_-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, underscores and hyphens allowed in %q", k))
	}
	prefixMaxLength := 128 - id.UniqueIDSuffixLength
	if len(value) > prefixMaxLength {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than %d characters", k, prefixMaxLength))
	}
	return
}
