// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_pinpoint_email_template", name="EmailTemplate")
func ResourceEmailTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEmailTemplateCreate,
		ReadWithoutTimeout:   resourceEmailTemplateRead,
		UpdateWithoutTimeout: resourceEmailTemplateUpdate,
		DeleteWithoutTimeout: resourceEmailTemplateDelete,

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
			"html": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"text": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subject": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_substitutions": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"recommender_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceEmailTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	templateName := d.Get("name").(string)

	templateRequest := &pinpoint.EmailTemplateRequest{}

	if v, ok := d.GetOk("html"); ok {
		templateRequest.HtmlPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("text"); ok {
		templateRequest.TextPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subject"); ok {
		templateRequest.Subject = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_substitutions"); ok {
		templateRequest.DefaultSubstitutions = aws.String(v.(string))
	}

	if v, ok := d.GetOk("recommender_id"); ok {
		templateRequest.RecommenderId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		templateRequest.TemplateDescription = aws.String(v.(string))
	}

	input := &pinpoint.CreateEmailTemplateInput{
		TemplateName:         &templateName,
		EmailTemplateRequest: templateRequest,
	}

	log.Printf("[DEBUG] Creating Pinpoint Email Template: %#v", input)
	_, err := conn.CreateEmailTemplateWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating Pinpoint template failed: %s", err.Error())
	}
	d.SetId(templateName)

	return append(diags, resourceEmailTemplateRead(ctx, d, meta)...)
}

func resourceEmailTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)
	input := &pinpoint.GetEmailTemplateInput{
		TemplateName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Pinpoint Email Template: %#v", input)
	gto, err := conn.GetEmailTemplateWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
			log.Printf("[WARN] Pinpoint template %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "Reading Pinpoint template '%s' failed: %s", aws.StringValue(input.TemplateName), err.Error())
	}

	d.Set("html", gto.EmailTemplateResponse.HtmlPart)
	d.Set("name", gto.EmailTemplateResponse.TemplateName)
	d.Set("subject", gto.EmailTemplateResponse.Subject)
	d.Set("text", gto.EmailTemplateResponse.TextPart)
	d.Set("default_substitutions", gto.EmailTemplateResponse.DefaultSubstitutions)
	d.Set("recommender_id", gto.EmailTemplateResponse.RecommenderId)
	d.Set("description", gto.EmailTemplateResponse.TemplateDescription)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "pinpoint",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("templates/%s/EMAIL", aws.StringValue(gto.EmailTemplateResponse.TemplateName)),
	}.String()
	d.Set("arn", arn)

	return diags
}

func resourceEmailTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	templateName := d.Id()

	templateRequest := &pinpoint.EmailTemplateRequest{}

	if v, ok := d.GetOk("html"); ok {
		templateRequest.HtmlPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("text"); ok {
		templateRequest.TextPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subject"); ok {
		templateRequest.Subject = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_substitutions"); ok {
		templateRequest.DefaultSubstitutions = aws.String(v.(string))
	}

	if v, ok := d.GetOk("recommender_id"); ok {
		templateRequest.RecommenderId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		templateRequest.TemplateDescription = aws.String(v.(string))
	}

	input := &pinpoint.UpdateEmailTemplateInput{
		TemplateName:         &templateName,
		EmailTemplateRequest: templateRequest,
	}

	log.Printf("[DEBUG] Updating Pinpoint Email Template: %#v", input)
	_, err := conn.UpdateEmailTemplateWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Updating Pinpoint template failed: %s", err.Error())
	}

	return append(diags, resourceEmailTemplateRead(ctx, d, meta)...)
}

func resourceEmailTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)
	input := &pinpoint.DeleteEmailTemplateInput{
		TemplateName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Pinpoint Email Template: %#v", input)
	_, err := conn.DeleteEmailTemplateWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Deleting Pinpoint template failed: %s", err.Error())
	}
	return diags
}
