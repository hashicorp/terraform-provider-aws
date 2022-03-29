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

func ResourceEventIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEventIntegrationCreate,
		ReadContext:   resourceEventIntegrationRead,
		UpdateContext: resourceEventIntegrationUpdate,
		DeleteContext: resourceEventIntegrationDelete,
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
			"eventbridge_bus": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\/\._\-]{1,255}$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
			},
			"event_filter": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^aws\.partner\/.*$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\/\._\-]{1,255}$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEventIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &appintegrationsservice.CreateEventIntegrationInput{
		ClientToken:    aws.String(resource.UniqueId()),
		EventBridgeBus: aws.String(d.Get("eventbridge_bus").(string)),
		EventFilter:    expandEventFilter(d.Get("event_filter").([]interface{})),
		Name:           aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating AppIntegrations Event Integration %s", input)
	output, err := conn.CreateEventIntegrationWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating AppIntegrations Event Integration (%s): %w", name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating AppIntegrations Event Integration (%s): empty output", name))
	}

	// Name is unique
	d.SetId(name)

	return resourceEventIntegrationRead(ctx, d, meta)
}

func resourceEventIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Id()

	resp, err := conn.GetEventIntegrationWithContext(ctx, &appintegrationsservice.GetEventIntegrationInput{
		Name: aws.String(name),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appintegrationsservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] AppIntegrations Event Integration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting AppIntegrations Event Integration (%s): %w", d.Id(), err))
	}

	if resp == nil {
		return diag.FromErr(fmt.Errorf("error getting AppIntegrations Event Integration (%s): empty response", d.Id()))
	}

	d.Set("arn", resp.EventIntegrationArn)
	d.Set("description", resp.Description)
	d.Set("eventbridge_bus", resp.EventBridgeBus)
	d.Set("name", resp.Name)

	if err := d.Set("event_filter", flattenEventFilter(resp.EventFilter)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting event_filter: %w", err))
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

func resourceEventIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn

	name := d.Id()

	if d.HasChange("description") {
		_, err := conn.UpdateEventIntegrationWithContext(ctx, &appintegrationsservice.UpdateEventIntegrationInput{
			Name:        aws.String(name),
			Description: aws.String(d.Get("description").(string)),
		})

		if err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error updating EventIntegration (%s): %w", d.Id(), err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceEventIntegrationRead(ctx, d, meta)
}

func resourceEventIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn

	name := d.Id()

	_, err := conn.DeleteEventIntegrationWithContext(ctx, &appintegrationsservice.DeleteEventIntegrationInput{
		Name: aws.String(name),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting EventIntegration (%s): %w", d.Id(), err))
	}

	return nil
}

func expandEventFilter(eventFilter []interface{}) *appintegrationsservice.EventFilter {
	if len(eventFilter) == 0 || eventFilter[0] == nil {
		return nil
	}

	tfMap, ok := eventFilter[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &appintegrationsservice.EventFilter{
		Source: aws.String(tfMap["source"].(string)),
	}

	return result
}

func flattenEventFilter(eventFilter *appintegrationsservice.EventFilter) []interface{} {
	if eventFilter == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"source": aws.StringValue(eventFilter.Source),
	}

	return []interface{}{values}
}
