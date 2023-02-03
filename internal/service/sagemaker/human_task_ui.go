package sagemaker

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceHumanTaskUI() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHumanTaskUICreate,
		ReadWithoutTimeout:   resourceHumanTaskUIRead,
		UpdateWithoutTimeout: resourceHumanTaskUIUpdate,
		DeleteWithoutTimeout: resourceHumanTaskUIDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
						"content": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128000),
						},
						"content_sha256": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url": {
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
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9](-*[a-z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHumanTaskUICreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("human_task_ui_name").(string)
	input := &sagemaker.CreateHumanTaskUiInput{
		HumanTaskUiName: aws.String(name),
		UiTemplate:      expandHumanTaskUiUiTemplate(d.Get("ui_template").([]interface{})),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating SageMaker HumanTaskUi: %s", input)
	_, err := conn.CreateHumanTaskUiWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker HumanTaskUi (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceHumanTaskUIRead(ctx, d, meta)...)
}

func resourceHumanTaskUIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	humanTaskUi, err := FindHumanTaskUIByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker HumanTaskUi (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker HumanTaskUi (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(humanTaskUi.HumanTaskUiArn)
	d.Set("arn", arn)
	d.Set("human_task_ui_name", humanTaskUi.HumanTaskUiName)

	if err := d.Set("ui_template", flattenHumanTaskUiUiTemplate(humanTaskUi.UiTemplate, d.Get("ui_template.0.content").(string))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ui_template: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SageMaker HumanTaskUi (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceHumanTaskUIUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker HumanTaskUi (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceHumanTaskUIRead(ctx, d, meta)...)
}

func resourceHumanTaskUIDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	log.Printf("[DEBUG] Deleting SageMaker HumanTaskUi: %s", d.Id())
	_, err := conn.DeleteHumanTaskUiWithContext(ctx, &sagemaker.DeleteHumanTaskUiInput{
		HumanTaskUiName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker HumanTaskUi (%s): %s", d.Id(), err)
	}

	return diags
}

func expandHumanTaskUiUiTemplate(l []interface{}) *sagemaker.UiTemplate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.UiTemplate{
		Content: aws.String(m["content"].(string)),
	}

	return config
}

func flattenHumanTaskUiUiTemplate(config *sagemaker.UiTemplateInfo, content string) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"content_sha256": aws.StringValue(config.ContentSha256),
		"url":            aws.StringValue(config.Url),
		"content":        content,
	}

	return []map[string]interface{}{m}
}
