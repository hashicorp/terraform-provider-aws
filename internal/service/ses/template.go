// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_template")
func ResourceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTemplateCreate,
		ReadWithoutTimeout:   resourceTemplateRead,
		UpdateWithoutTimeout: resourceTemplateUpdate,
		DeleteWithoutTimeout: resourceTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"html": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512000),
			},
			"subject": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512000),
			},
		},
	}
}
func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	templateName := d.Get(names.AttrName).(string)

	template := ses.Template{
		TemplateName: aws.String(templateName),
	}

	if v, ok := d.GetOk("html"); ok {
		template.HtmlPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subject"); ok {
		template.SubjectPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("text"); ok {
		template.TextPart = aws.String(v.(string))
	}

	input := ses.CreateTemplateInput{
		Template: &template,
	}

	log.Printf("[DEBUG] Creating SES template: %#v", input)
	_, err := conn.CreateTemplateWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating SES template failed: %s", err.Error())
	}
	d.SetId(templateName)

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)
	input := ses.GetTemplateInput{
		TemplateName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading SES template: %#v", input)
	gto, err := conn.GetTemplateWithContext(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ses.ErrCodeTemplateDoesNotExistException) {
			log.Printf("[WARN] SES template %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "Reading SES template '%s' failed: %s", aws.StringValue(input.TemplateName), err.Error())
	}

	d.Set("html", gto.Template.HtmlPart)
	d.Set(names.AttrName, gto.Template.TemplateName)
	d.Set("subject", gto.Template.SubjectPart)
	d.Set("text", gto.Template.TextPart)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("template/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)

	return diags
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	templateName := d.Id()

	template := ses.Template{
		TemplateName: aws.String(templateName),
	}

	if v, ok := d.GetOk("html"); ok {
		template.HtmlPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subject"); ok {
		template.SubjectPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("text"); ok {
		template.TextPart = aws.String(v.(string))
	}

	input := ses.UpdateTemplateInput{
		Template: &template,
	}

	_, err := conn.UpdateTemplateWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SES template (%s): %s", templateName, err)
	}

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)
	input := ses.DeleteTemplateInput{
		TemplateName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Delete SES template: %#v", input)
	_, err := conn.DeleteTemplateWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Deleting SES template '%s' failed: %s", *input.TemplateName, err.Error())
	}
	return diags
}
