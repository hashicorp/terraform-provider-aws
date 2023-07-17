// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appintegrations

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appintegrations_data_integration", name="Data Integration")
// @Tags(identifierAttribute="arn")
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"kms_key": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\/\._\-]+$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
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
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\/\._\-]+$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
							),
						},
						"schedule_expression": {
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
					validation.StringMatch(regexp.MustCompile(`^\w+\:\/\/\w+\/[\w/!@#+=.-]+$`), "should be a valid source uri"),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDataIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn(ctx)

	name := d.Get("name").(string)
	input := &appintegrationsservice.CreateDataIntegrationInput{
		ClientToken:    aws.String(id.UniqueId()),
		KmsKey:         aws.String(d.Get("kms_key").(string)),
		Name:           aws.String(name),
		ScheduleConfig: expandScheduleConfig(d.Get("schedule_config").([]interface{})),
		SourceURI:      aws.String(d.Get("source_uri").(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateDataIntegrationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating AppIntegrations Data Integration (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Id))

	return resourceDataIntegrationRead(ctx, d, meta)
}

func resourceDataIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn(ctx)

	output, err := conn.GetDataIntegrationWithContext(ctx, &appintegrationsservice.GetDataIntegrationInput{
		Identifier: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appintegrationsservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] AppIntegrations Data Integration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading AppIntegrations Data Integration (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("description", output.Description)
	d.Set("kms_key", output.KmsKey)
	d.Set("name", output.Name)
	if err := d.Set("schedule_config", flattenScheduleConfig(output.ScheduleConfiguration)); err != nil {
		return diag.Errorf("schedule_config tags: %s", err)
	}
	d.Set("source_uri", output.SourceURI)

	setTagsOut(ctx, output.Tags)

	return nil
}

func resourceDataIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn(ctx)

	if d.HasChanges("description", "name") {
		_, err := conn.UpdateDataIntegrationWithContext(ctx, &appintegrationsservice.UpdateDataIntegrationInput{
			Description: aws.String(d.Get("description").(string)),
			Identifier:  aws.String(d.Id()),
			Name:        aws.String(d.Get("name").(string)),
		})

		if err != nil {
			return diag.Errorf("updating AppIntegrations Data Integration (%s): %s", d.Id(), err)
		}
	}

	return resourceDataIntegrationRead(ctx, d, meta)
}

func resourceDataIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn(ctx)

	_, err := conn.DeleteDataIntegrationWithContext(ctx, &appintegrationsservice.DeleteDataIntegrationInput{
		DataIntegrationIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting AppIntegrations Data Integration (%s): %s", d.Id(), err)
	}

	return nil
}

func expandScheduleConfig(scheduleConfig []interface{}) *appintegrationsservice.ScheduleConfiguration {
	if len(scheduleConfig) == 0 || scheduleConfig[0] == nil {
		return nil
	}

	tfMap, ok := scheduleConfig[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &appintegrationsservice.ScheduleConfiguration{
		FirstExecutionFrom: aws.String(tfMap["first_execution_from"].(string)),
		Object:             aws.String(tfMap["object"].(string)),
		ScheduleExpression: aws.String(tfMap["schedule_expression"].(string)),
	}

	return result
}

func flattenScheduleConfig(scheduleConfig *appintegrationsservice.ScheduleConfiguration) []interface{} {
	if scheduleConfig == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"first_execution_from": aws.StringValue(scheduleConfig.FirstExecutionFrom),
		"object":               aws.StringValue(scheduleConfig.Object),
		"schedule_expression":  aws.StringValue(scheduleConfig.ScheduleExpression),
	}

	return []interface{}{values}
}
