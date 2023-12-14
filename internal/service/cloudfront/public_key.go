// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_cloudfront_public_key")
func ResourcePublicKey() *schema.Resource {
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
			"comment": {
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
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validPublicKeyName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validPublicKeyNamePrefix,
			},
		},
	}
}

func resourcePublicKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get("name").(string)),
		create.WithConfiguredPrefix(d.Get("name_prefix").(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()
	input := &cloudfront.CreatePublicKeyInput{
		PublicKeyConfig: &cloudfront.PublicKeyConfig{
			EncodedKey: aws.String(d.Get("encoded_key").(string)),
			Name:       aws.String(name),
		},
	}

	if v, ok := d.GetOk("caller_reference"); ok {
		input.PublicKeyConfig.CallerReference = aws.String(v.(string))
	} else {
		input.PublicKeyConfig.CallerReference = aws.String(id.UniqueId())
	}

	if v, ok := d.GetOk("comment"); ok {
		input.PublicKeyConfig.Comment = aws.String(v.(string))
	}

	output, err := conn.CreatePublicKeyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Public Key (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.PublicKey.Id))

	return append(diags, resourcePublicKeyRead(ctx, d, meta)...)
}

func resourcePublicKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

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
	d.Set("comment", publicKeyConfig.Comment)
	d.Set("encoded_key", publicKeyConfig.EncodedKey)
	d.Set("etag", output.ETag)
	d.Set("name", publicKeyConfig.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(publicKeyConfig.Name)))

	return diags
}

func resourcePublicKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	input := &cloudfront.UpdatePublicKeyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
		PublicKeyConfig: &cloudfront.PublicKeyConfig{
			EncodedKey: aws.String(d.Get("encoded_key").(string)),
			Name:       aws.String(d.Get("name").(string)),
		},
	}

	if v, ok := d.GetOk("caller_reference"); ok {
		input.PublicKeyConfig.CallerReference = aws.String(v.(string))
	} else {
		input.PublicKeyConfig.CallerReference = aws.String(id.UniqueId())
	}

	if v, ok := d.GetOk("comment"); ok {
		input.PublicKeyConfig.Comment = aws.String(v.(string))
	}

	_, err := conn.UpdatePublicKeyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Public Key (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePublicKeyRead(ctx, d, meta)...)
}

func resourcePublicKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	log.Printf("[DEBUG] Deleting CloudFront Public Key: %s", d.Id())
	_, err := conn.DeletePublicKeyWithContext(ctx, &cloudfront.DeletePublicKeyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchPublicKey) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Public Key (%s): %s", d.Id(), err)
	}

	return diags
}

func findPublicKeyByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetPublicKeyOutput, error) {
	input := &cloudfront.GetPublicKeyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetPublicKeyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchPublicKey) {
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
