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

// @SDKResource("aws_appintegrations_event_integration", name="Event Integration")
// @Tags(identifierAttribute="arn")
func ResourceEventIntegration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventIntegrationCreate,
		ReadWithoutTimeout:   resourceEventIntegrationRead,
		UpdateWithoutTimeout: resourceEventIntegrationUpdate,
		DeleteWithoutTimeout: resourceEventIntegrationDelete,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEventIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn(ctx)

	name := d.Get("name").(string)
	input := &appintegrationsservice.CreateEventIntegrationInput{
		ClientToken:    aws.String(id.UniqueId()),
		EventBridgeBus: aws.String(d.Get("eventbridge_bus").(string)),
		EventFilter:    expandEventFilter(d.Get("event_filter").([]interface{})),
		Name:           aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating AppIntegrations Event Integration %s", input)
	output, err := conn.CreateEventIntegrationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating AppIntegrations Event Integration (%s): %s", name, err)
	}

	if output == nil {
		return diag.Errorf("creating AppIntegrations Event Integration (%s): empty output", name)
	}

	// Name is unique
	d.SetId(name)

	return resourceEventIntegrationRead(ctx, d, meta)
}

func resourceEventIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn(ctx)

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
		return diag.Errorf("getting AppIntegrations Event Integration (%s): %s", d.Id(), err)
	}

	if resp == nil {
		return diag.Errorf("getting AppIntegrations Event Integration (%s): empty response", d.Id())
	}

	d.Set("arn", resp.EventIntegrationArn)
	d.Set("description", resp.Description)
	d.Set("eventbridge_bus", resp.EventBridgeBus)
	d.Set("name", resp.Name)

	if err := d.Set("event_filter", flattenEventFilter(resp.EventFilter)); err != nil {
		return diag.Errorf("setting event_filter: %s", err)
	}

	setTagsOut(ctx, resp.Tags)

	return nil
}

func resourceEventIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn(ctx)

	name := d.Id()

	if d.HasChange("description") {
		_, err := conn.UpdateEventIntegrationWithContext(ctx, &appintegrationsservice.UpdateEventIntegrationInput{
			Name:        aws.String(name),
			Description: aws.String(d.Get("description").(string)),
		})

		if err != nil {
			return diag.Errorf("updating EventIntegration (%s): %s", d.Id(), err)
		}
	}

	return resourceEventIntegrationRead(ctx, d, meta)
}

func resourceEventIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn(ctx)

	name := d.Id()

	_, err := conn.DeleteEventIntegrationWithContext(ctx, &appintegrationsservice.DeleteEventIntegrationInput{
		Name: aws.String(name),
	})

	if err != nil {
		return diag.Errorf("deleting EventIntegration (%s): %s", d.Id(), err)
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
