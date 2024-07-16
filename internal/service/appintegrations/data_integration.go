// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appintegrations

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appintegrations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appintegrations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appintegrations_data_integration", name="Data Integration")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/appintegrations;appintegrations.GetDataIntegrationOutput")
func ResourceDataIntegration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataIntegrationCreate,
		ReadWithoutTimeout:   resourceDataIntegrationRead,
		UpdateWithoutTimeout: resourceDataIntegrationUpdate,
		DeleteWithoutTimeout: resourceDataIntegrationDelete,

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
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			names.AttrKMSKey: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z\/\._\-]+$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
				),
			},
			"schedule_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"first_execution_from": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"object": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 255),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z\/\._\-]+$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
							),
						},
						names.AttrScheduleExpression: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"source_uri": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1000),
					validation.StringMatch(regexache.MustCompile(`^\w+\:\/\/\w+\/[\w/!@#+=.-]+$`), "should be a valid source uri"),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDataIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppIntegrationsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appintegrations.CreateDataIntegrationInput{
		ClientToken:    aws.String(id.UniqueId()),
		KmsKey:         aws.String(d.Get(names.AttrKMSKey).(string)),
		Name:           aws.String(name),
		ScheduleConfig: expandScheduleConfig(d.Get("schedule_config").([]interface{})),
		SourceURI:      aws.String(d.Get("source_uri").(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateDataIntegration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppIntegrations Data Integration (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	return append(diags, resourceDataIntegrationRead(ctx, d, meta)...)
}

func resourceDataIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppIntegrationsClient(ctx)

	output, err := conn.GetDataIntegration(ctx, &appintegrations.GetDataIntegrationInput{
		Identifier: aws.String(d.Id()),
	})

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] AppIntegrations Data Integration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppIntegrations Data Integration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrKMSKey, output.KmsKey)
	d.Set(names.AttrName, output.Name)
	if err := d.Set("schedule_config", flattenScheduleConfig(output.ScheduleConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "schedule_config tags: %s", err)
	}
	d.Set("source_uri", output.SourceURI)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceDataIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppIntegrationsClient(ctx)

	if d.HasChanges(names.AttrDescription, names.AttrName) {
		_, err := conn.UpdateDataIntegration(ctx, &appintegrations.UpdateDataIntegrationInput{
			Description: aws.String(d.Get(names.AttrDescription).(string)),
			Identifier:  aws.String(d.Id()),
			Name:        aws.String(d.Get(names.AttrName).(string)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppIntegrations Data Integration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataIntegrationRead(ctx, d, meta)...)
}

func resourceDataIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppIntegrationsClient(ctx)

	_, err := conn.DeleteDataIntegration(ctx, &appintegrations.DeleteDataIntegrationInput{
		DataIntegrationIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppIntegrations Data Integration (%s): %s", d.Id(), err)
	}

	return diags
}

func expandScheduleConfig(scheduleConfig []interface{}) *awstypes.ScheduleConfiguration {
	if len(scheduleConfig) == 0 || scheduleConfig[0] == nil {
		return nil
	}

	tfMap, ok := scheduleConfig[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.ScheduleConfiguration{
		FirstExecutionFrom: aws.String(tfMap["first_execution_from"].(string)),
		Object:             aws.String(tfMap["object"].(string)),
		ScheduleExpression: aws.String(tfMap[names.AttrScheduleExpression].(string)),
	}

	return result
}

func flattenScheduleConfig(scheduleConfig *awstypes.ScheduleConfiguration) []interface{} {
	if scheduleConfig == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"first_execution_from":       aws.ToString(scheduleConfig.FirstExecutionFrom),
		"object":                     aws.ToString(scheduleConfig.Object),
		names.AttrScheduleExpression: aws.ToString(scheduleConfig.ScheduleExpression),
	}

	return []interface{}{values}
}
