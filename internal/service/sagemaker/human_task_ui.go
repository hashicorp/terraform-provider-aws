// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_human_task_ui", name="Human Task UI")
// @Tags(identifierAttribute="arn")
func resourceHumanTaskUI() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHumanTaskUICreate,
		ReadWithoutTimeout:   resourceHumanTaskUIRead,
		UpdateWithoutTimeout: resourceHumanTaskUIUpdate,
		DeleteWithoutTimeout: resourceHumanTaskUIDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ui_template": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrContent: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128000),
						},
						"content_sha256": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrURL: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"human_task_ui_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z](-*[0-9a-z])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceHumanTaskUICreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("human_task_ui_name").(string)
	input := &sagemaker.CreateHumanTaskUiInput{
		HumanTaskUiName: aws.String(name),
		Tags:            getTagsIn(ctx),
		UiTemplate:      expandHumanTaskUiUiTemplate(d.Get("ui_template").([]any)),
	}

	log.Printf("[DEBUG] Creating SageMaker AI HumanTaskUi: %#v", input)
	_, err := conn.CreateHumanTaskUi(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI HumanTaskUi (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceHumanTaskUIRead(ctx, d, meta)...)
}

func resourceHumanTaskUIRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	humanTaskUi, err := findHumanTaskUIByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker AI HumanTaskUi (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI HumanTaskUi (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, humanTaskUi.HumanTaskUiArn)
	d.Set("human_task_ui_name", humanTaskUi.HumanTaskUiName)

	if err := d.Set("ui_template", flattenHumanTaskUiUiTemplate(humanTaskUi.UiTemplate, d.Get("ui_template.0.content").(string))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ui_template: %s", err)
	}

	return diags
}

func resourceHumanTaskUIUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceHumanTaskUIRead(ctx, d, meta)...)
}

func resourceHumanTaskUIDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[DEBUG] Deleting SageMaker AI HumanTaskUi: %s", d.Id())
	_, err := conn.DeleteHumanTaskUi(ctx, &sagemaker.DeleteHumanTaskUiInput{
		HumanTaskUiName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI HumanTaskUi (%s): %s", d.Id(), err)
	}

	return diags
}

func findHumanTaskUIByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeHumanTaskUiOutput, error) {
	input := &sagemaker.DescribeHumanTaskUiInput{
		HumanTaskUiName: aws.String(name),
	}

	output, err := conn.DescribeHumanTaskUi(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
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

func expandHumanTaskUiUiTemplate(l []any) *awstypes.UiTemplate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.UiTemplate{
		Content: aws.String(m[names.AttrContent].(string)),
	}

	return config
}

func flattenHumanTaskUiUiTemplate(config *awstypes.UiTemplateInfo, content string) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"content_sha256":  aws.ToString(config.ContentSha256),
		names.AttrURL:     aws.ToString(config.Url),
		names.AttrContent: content,
	}

	return []map[string]any{m}
}
