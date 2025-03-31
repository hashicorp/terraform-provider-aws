// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_template", name="Template")
func resourceTemplate() *schema.Resource {
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
			"html": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512000),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
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
func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	name := d.Get(names.AttrName).(string)
	template := &awstypes.Template{
		TemplateName: aws.String(name),
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

	input := &ses.CreateTemplateInput{
		Template: template,
	}

	_, err := conn.CreateTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Template (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	template, err := findTemplateByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Template (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("template/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("html", template.HtmlPart)
	d.Set(names.AttrName, template.TemplateName)
	d.Set("subject", template.SubjectPart)
	d.Set("text", template.TextPart)

	return diags
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	template := &awstypes.Template{
		TemplateName: aws.String(d.Id()),
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

	input := &ses.UpdateTemplateInput{
		Template: template,
	}

	_, err := conn.UpdateTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SES Template (%s): %s", d.Id(), err)
	}

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting SES Template: %s", d.Id())
	_, err := conn.DeleteTemplate(ctx, &ses.DeleteTemplateInput{
		TemplateName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Template (%s): %s", d.Id(), err)
	}

	return diags
}

func findTemplateByName(ctx context.Context, conn *ses.Client, name string) (*awstypes.Template, error) {
	input := &ses.GetTemplateInput{
		TemplateName: aws.String(name),
	}

	return findTemplate(ctx, conn, input)
}

func findTemplate(ctx context.Context, conn *ses.Client, input *ses.GetTemplateInput) (*awstypes.Template, error) {
	output, err := conn.GetTemplate(ctx, input)

	if errs.IsA[*awstypes.TemplateDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Template == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Template, nil
}
