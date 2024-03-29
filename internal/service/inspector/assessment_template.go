// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameAssessmentTemplate = "Assessment Template"
)

// @SDKResource("aws_inspector_assessment_template", name="Assessment Template")
// @Tags(identifierAttribute="id")
func ResourceAssessmentTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAssessmentTemplateCreate,
		ReadWithoutTimeout:   resourceAssessmentTemplateRead,
		UpdateWithoutTimeout: resourceAssessmentTemplateUpdate,
		DeleteWithoutTimeout: resourceAssessmentTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"duration": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"event_subscription": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(inspector.Event_Values(), false),
						},
						"topic_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rules_package_arns": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAssessmentTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn(ctx)

	name := d.Get("name").(string)
	input := &inspector.CreateAssessmentTemplateInput{
		AssessmentTargetArn:    aws.String(d.Get("target_arn").(string)),
		AssessmentTemplateName: aws.String(name),
		DurationInSeconds:      aws.Int64(int64(d.Get("duration").(int))),
		RulesPackageArns:       flex.ExpandStringSet(d.Get("rules_package_arns").(*schema.Set)),
	}

	output, err := conn.CreateAssessmentTemplateWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector Classic Assessment Template (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.AssessmentTemplateArn))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Inspector Classic Assessment Template (%s) tags: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("event_subscription"); ok && v.(*schema.Set).Len() > 0 {
		input := expandEventSubscriptions(v.(*schema.Set).List(), output.AssessmentTemplateArn)

		if err := subscribeToEvents(ctx, conn, input); err != nil {
			return create.AppendDiagError(diags, names.Inspector, create.ErrActionCreating, ResNameAssessmentTemplate, d.Id(), err)
		}
	}

	return append(diags, resourceAssessmentTemplateRead(ctx, d, meta)...)
}

func resourceAssessmentTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn(ctx)

	resp, err := conn.DescribeAssessmentTemplatesWithContext(ctx, &inspector.DescribeAssessmentTemplatesInput{
		AssessmentTemplateArns: aws.StringSlice([]string{d.Id()}),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector Classic Assessment Template (%s): %s", d.Id(), err)
	}

	if resp.AssessmentTemplates == nil || len(resp.AssessmentTemplates) == 0 {
		log.Printf("[WARN] Inspector Classic Assessment Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	template := resp.AssessmentTemplates[0]

	arn := aws.StringValue(template.Arn)
	d.Set("arn", arn)
	d.Set("duration", template.DurationInSeconds)
	d.Set("name", template.Name)
	d.Set("rules_package_arns", aws.StringValueSlice(template.RulesPackageArns))
	d.Set("target_arn", template.AssessmentTargetArn)

	output, err := findSubscriptionsByAssessmentTemplateARN(ctx, conn, arn)

	if err != nil {
		return create.AppendDiagError(diags, names.Inspector, create.ErrActionReading, ResNameAssessmentTemplate, d.Id(), err)
	}

	if err := d.Set("event_subscription", flattenSubscriptions(output)); err != nil {
		return create.AppendDiagError(diags, names.Inspector, create.ErrActionSetting, ResNameAssessmentTemplate, d.Id(), err)
	}

	return diags
}

func resourceAssessmentTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn(ctx)

	if d.HasChange("event_subscription") {
		old, new := d.GetChange("event_subscription")
		oldSet := old.(*schema.Set)
		newSet := new.(*schema.Set)

		eventSubscriptionsToAdd := newSet.Difference(oldSet)
		eventSubscriptionsToRemove := oldSet.Difference(newSet)

		templateId := aws.String(d.Id())

		addEventSubscriptionsInput := expandEventSubscriptions(eventSubscriptionsToAdd.List(), templateId)
		removeEventSubscriptionsInput := expandEventSubscriptions(eventSubscriptionsToRemove.List(), templateId)

		if err := subscribeToEvents(ctx, conn, addEventSubscriptionsInput); err != nil {
			return create.AppendDiagError(diags, names.Inspector, create.ErrActionUpdating, ResNameAssessmentTemplate, d.Id(), err)
		}

		if err := unsubscribeFromEvents(ctx, conn, removeEventSubscriptionsInput); err != nil {
			return create.AppendDiagError(diags, names.Inspector, create.ErrActionUpdating, ResNameAssessmentTemplate, d.Id(), err)
		}
	}

	return append(diags, resourceAssessmentTemplateRead(ctx, d, meta)...)
}

func resourceAssessmentTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn(ctx)

	log.Printf("[INFO] Deleting Inspector Classic Assessment Template: %s", d.Id())
	_, err := conn.DeleteAssessmentTemplateWithContext(ctx, &inspector.DeleteAssessmentTemplateInput{
		AssessmentTemplateArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Inspector Classic Assessment Template (%s): %s", d.Id(), err)
	}

	return diags
}

func expandEventSubscriptions(tfList []interface{}, templateArn *string) []*inspector.SubscribeToEventInput {
	if len(tfList) == 0 {
		return nil
	}

	var eventSubscriptions []*inspector.SubscribeToEventInput

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		eventSubscription := expandEventSubscription(tfMap, templateArn)

		eventSubscriptions = append(eventSubscriptions, eventSubscription)
	}

	return eventSubscriptions
}

func expandEventSubscription(tfMap map[string]interface{}, templateArn *string) *inspector.SubscribeToEventInput {
	if tfMap == nil {
		return nil
	}

	eventSubscription := &inspector.SubscribeToEventInput{
		Event:       aws.String(tfMap["event"].(string)),
		ResourceArn: templateArn,
		TopicArn:    aws.String(tfMap["topic_arn"].(string)),
	}

	return eventSubscription
}

func flattenSubscriptions(subscriptions []*inspector.Subscription) []interface{} {
	if len(subscriptions) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, subscription := range subscriptions {
		if subscription == nil {
			continue
		}

		for _, eventSubscription := range subscription.EventSubscriptions {
			if eventSubscription == nil {
				continue
			}

			tfList = append(tfList, flattenEventSubscription(eventSubscription, subscription.TopicArn))
		}
	}

	return tfList
}

func flattenEventSubscription(eventSubscription *inspector.EventSubscription, topicArn *string) map[string]interface{} {
	if eventSubscription == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["event"] = eventSubscription.Event
	tfMap["topic_arn"] = topicArn

	return tfMap
}

func subscribeToEvents(ctx context.Context, conn *inspector.Inspector, eventSubscriptions []*inspector.SubscribeToEventInput) error {
	for _, eventSubscription := range eventSubscriptions {
		_, err := conn.SubscribeToEventWithContext(ctx, eventSubscription)

		if err != nil {
			return create.Error(names.Inspector, create.ErrActionCreating, ResNameAssessmentTemplate, *eventSubscription.TopicArn, err)
		}
	}

	return nil
}

func unsubscribeFromEvents(ctx context.Context, conn *inspector.Inspector, eventSubscriptions []*inspector.SubscribeToEventInput) error {
	for _, eventSubscription := range eventSubscriptions {
		input := &inspector.UnsubscribeFromEventInput{
			Event:       eventSubscription.Event,
			ResourceArn: eventSubscription.ResourceArn,
			TopicArn:    eventSubscription.TopicArn,
		}

		_, err := conn.UnsubscribeFromEventWithContext(ctx, input)

		if err != nil {
			return create.Error(names.Inspector, create.ErrActionDeleting, ResNameAssessmentTemplate, *eventSubscription.TopicArn, err)
		}
	}

	return nil
}
