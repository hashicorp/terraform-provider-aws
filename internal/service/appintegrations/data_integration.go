package appintegrations

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDataIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataIntegrationCreate,
		ReadContext:   resourceDataIntegrationRead,
		UpdateContext: resourceDataIntegrationUpdate,
		DeleteContext: resourceDataIntegrationDelete,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDataIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &appintegrationsservice.CreateDataIntegrationInput{
		ClientToken:    aws.String(resource.UniqueId()),
		KmsKey:         aws.String(d.Get("kms_key").(string)),
		Name:           aws.String(name),
		ScheduleConfig: expandScheduleConfig(d.Get("schedule_config").([]interface{})),
		SourceURI:      aws.String(d.Get("source_uri").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating AppIntegrations Data Integration %s", input)

	output, err := conn.CreateDataIntegrationWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating AppIntegrations Data Integration (%s): %w", name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating AppIntegrations Data Integration (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.Id))

	return resourceDataIntegrationRead(ctx, d, meta)
}

func resourceDataIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Id()

	resp, err := conn.GetDataIntegrationWithContext(ctx, &appintegrationsservice.GetDataIntegrationInput{
		Identifier: aws.String(id),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appintegrationsservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] AppIntegrations Data Integration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting AppIntegrations Data Integration (%s): %w", d.Id(), err))
	}

	if resp == nil {
		return diag.FromErr(fmt.Errorf("error getting AppIntegrations Data Integration (%s): empty response", d.Id()))
	}

	d.Set("arn", resp.Arn)
	d.Set("description", resp.Description)
	d.Set("kms_key", resp.KmsKey)
	d.Set("name", resp.Name)
	d.Set("source_uri", resp.SourceURI)

	if err := d.Set("schedule_config", flattenScheduleConfig(resp.ScheduleConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting schedule_config: %w", err))
	}

	tags := KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceDataIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn

	id := d.Id()

	if d.HasChanges("description", "name") {
		_, err := conn.UpdateDataIntegrationWithContext(ctx, &appintegrationsservice.UpdateDataIntegrationInput{
			Description: aws.String(d.Get("description").(string)),
			Identifier:  aws.String(id),
			Name:        aws.String(d.Get("name").(string)),
		})

		if err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error updating DataIntegration (%s): %w", d.Id(), err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceDataIntegrationRead(ctx, d, meta)
}

func resourceDataIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn

	id := d.Id()

	_, err := conn.DeleteDataIntegrationWithContext(ctx, &appintegrationsservice.DeleteDataIntegrationInput{
		DataIntegrationIdentifier: aws.String(id),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting DataIntegration (%s): %w", d.Id(), err))
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
