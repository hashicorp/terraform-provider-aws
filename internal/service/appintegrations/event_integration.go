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

// @SDKResource("aws_appintegrations_event_integration", name="Event Integration")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/appintegrations;appintegrations.GetEventIntegrationOutput")
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"eventbridge_bus": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z\/\._\-]{1,255}$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
			},
			"event_filter": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSource: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`^aws\.partner\/.*$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z\/\._\-]{1,255}$`), "should be not be more than 255 alphanumeric, forward slashes, dots, underscores, or hyphen characters"),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEventIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppIntegrationsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appintegrations.CreateEventIntegrationInput{
		ClientToken:    aws.String(id.UniqueId()),
		EventBridgeBus: aws.String(d.Get("eventbridge_bus").(string)),
		EventFilter:    expandEventFilter(d.Get("event_filter").([]interface{})),
		Name:           aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating AppIntegrations Event Integration %+v", input)
	output, err := conn.CreateEventIntegration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppIntegrations Event Integration (%s): %s", name, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating AppIntegrations Event Integration (%s): empty output", name)
	}

	// Name is unique
	d.SetId(name)

	return append(diags, resourceEventIntegrationRead(ctx, d, meta)...)
}

func resourceEventIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppIntegrationsClient(ctx)

	name := d.Id()

	resp, err := conn.GetEventIntegration(ctx, &appintegrations.GetEventIntegrationInput{
		Name: aws.String(name),
	})

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] AppIntegrations Event Integration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting AppIntegrations Event Integration (%s): %s", d.Id(), err)
	}

	if resp == nil {
		return sdkdiag.AppendErrorf(diags, "getting AppIntegrations Event Integration (%s): empty response", d.Id())
	}

	d.Set(names.AttrARN, resp.EventIntegrationArn)
	d.Set(names.AttrDescription, resp.Description)
	d.Set("eventbridge_bus", resp.EventBridgeBus)
	d.Set(names.AttrName, resp.Name)

	if err := d.Set("event_filter", flattenEventFilter(resp.EventFilter)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting event_filter: %s", err)
	}

	setTagsOut(ctx, resp.Tags)

	return diags
}

func resourceEventIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppIntegrationsClient(ctx)

	name := d.Id()

	if d.HasChange(names.AttrDescription) {
		_, err := conn.UpdateEventIntegration(ctx, &appintegrations.UpdateEventIntegrationInput{
			Name:        aws.String(name),
			Description: aws.String(d.Get(names.AttrDescription).(string)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventIntegration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceEventIntegrationRead(ctx, d, meta)...)
}

func resourceEventIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppIntegrationsClient(ctx)

	name := d.Id()

	_, err := conn.DeleteEventIntegration(ctx, &appintegrations.DeleteEventIntegrationInput{
		Name: aws.String(name),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventIntegration (%s): %s", d.Id(), err)
	}

	return diags
}

func expandEventFilter(eventFilter []interface{}) *awstypes.EventFilter {
	if len(eventFilter) == 0 || eventFilter[0] == nil {
		return nil
	}

	tfMap, ok := eventFilter[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.EventFilter{
		Source: aws.String(tfMap[names.AttrSource].(string)),
	}

	return result
}

func flattenEventFilter(eventFilter *awstypes.EventFilter) []interface{} {
	if eventFilter == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		names.AttrSource: aws.ToString(eventFilter.Source),
	}

	return []interface{}{values}
}
