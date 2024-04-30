// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devicefarm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_devicefarm_upload")
func ResourceUpload() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUploadCreate,
		ReadWithoutTimeout:   resourceUploadRead,
		UpdateWithoutTimeout: resourceUploadUpdate,
		DeleteWithoutTimeout: resourceUploadDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"metadata": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"project_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.UploadType](),
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUploadCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	input := &devicefarm.CreateUploadInput{
		Name:       aws.String(d.Get("name").(string)),
		ProjectArn: aws.String(d.Get("project_arn").(string)),
		Type:       awstypes.UploadType(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("content_type"); ok {
		input.ContentType = aws.String(v.(string))
	}

	out, err := conn.CreateUpload(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DeviceFarm Upload: %s", err)
	}

	arn := aws.ToString(out.Upload.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm Upload: %s", arn)
	d.SetId(arn)

	return append(diags, resourceUploadRead(ctx, d, meta)...)
}

func resourceUploadRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	upload, err := FindUploadByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Upload (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm Upload (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(upload.Arn)
	d.Set("name", upload.Name)
	d.Set("type", upload.Type)
	d.Set("content_type", upload.ContentType)
	d.Set("url", upload.Url)
	d.Set("category", upload.Category)
	d.Set("metadata", upload.Metadata)
	d.Set("arn", arn)

	projectArn, err := decodeProjectARN(arn, "upload", meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding project_arn (%s): %s", arn, err)
	}

	d.Set("project_arn", projectArn)

	return diags
}

func resourceUploadUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	input := &devicefarm.UpdateUploadInput{
		Arn: aws.String(d.Id()),
	}

	if d.HasChange("name") {
		input.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("content_type") {
		input.ContentType = aws.String(d.Get("content_type").(string))
	}

	log.Printf("[DEBUG] Updating DeviceFarm Upload: %s", d.Id())
	_, err := conn.UpdateUpload(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Upload (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUploadRead(ctx, d, meta)...)
}

func resourceUploadDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	input := &devicefarm.DeleteUploadInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm Upload: %s", d.Id())
	_, err := conn.DeleteUpload(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DeviceFarm Upload: %s", err)
	}

	return diags
}
