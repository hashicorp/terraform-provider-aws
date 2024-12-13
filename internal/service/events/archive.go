// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_event_archive", name="Archive")
func resourceArchive() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceArchiveCreate,
		ReadWithoutTimeout:   resourceArchiveRead,
		UpdateWithoutTimeout: resourceArchiveUpdate,
		DeleteWithoutTimeout: resourceArchiveDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"event_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateEventPatternValue(),
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v.(string))
					return json
				},
			},
			"event_source_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validArchiveName,
			},
			"retention_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceArchiveCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &eventbridge.CreateArchiveInput{
		ArchiveName:    aws.String(name),
		EventSourceArn: aws.String(d.Get("event_source_arn").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_pattern"); ok {
		v, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.EventPattern = aws.String(v)
	}

	if v, ok := d.GetOk("retention_days"); ok {
		input.RetentionDays = aws.Int32(int32(v.(int)))
	}

	_, err := conn.CreateArchive(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Archive (%s)): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceArchiveRead(ctx, d, meta)...)
}

func resourceArchiveRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	output, err := findArchiveByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Archive (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Archive (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ArchiveArn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("event_pattern", output.EventPattern)
	d.Set("event_source_arn", output.EventSourceArn)
	d.Set(names.AttrName, output.ArchiveName)
	d.Set("retention_days", output.RetentionDays)

	return diags
}

func resourceArchiveUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	input := &eventbridge.UpdateArchiveInput{
		ArchiveName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_pattern"); ok {
		v, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.EventPattern = aws.String(v)
	}

	if v, ok := d.GetOk("retention_days"); ok {
		input.RetentionDays = aws.Int32(int32(v.(int)))
	}

	_, err := conn.UpdateArchive(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge Archive (%s): %s", d.Id(), err)
	}

	return append(diags, resourceArchiveRead(ctx, d, meta)...)
}

func resourceArchiveDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Archive: %s", d.Id())
	_, err := conn.DeleteArchive(ctx, &eventbridge.DeleteArchiveInput{
		ArchiveName: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Archive (%s): %s", d.Id(), err)
	}

	return diags
}

func findArchiveByName(ctx context.Context, conn *eventbridge.Client, name string) (*eventbridge.DescribeArchiveOutput, error) {
	input := &eventbridge.DescribeArchiveInput{
		ArchiveName: aws.String(name),
	}

	output, err := conn.DescribeArchive(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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
