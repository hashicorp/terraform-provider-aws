// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDuration: {
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.InspectorEvent](),
						},
						names.AttrTopicARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrName: {
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
			names.AttrTargetARN: {
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
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &inspector.CreateAssessmentTemplateInput{
		AssessmentTargetArn:    aws.String(d.Get(names.AttrTargetARN).(string)),
		AssessmentTemplateName: aws.String(name),
		DurationInSeconds:      aws.Int32(int32(d.Get(names.AttrDuration).(int))),
		RulesPackageArns:       flex.ExpandStringValueSet(d.Get("rules_package_arns").(*schema.Set)),
	}

	output, err := conn.CreateAssessmentTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector Classic Assessment Template (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.AssessmentTemplateArn))

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
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	template, err := FindAssessmentTemplateByID(ctx, conn, d.Id())
	if errs.IsA[*retry.NotFoundError](err) {
		log.Printf("[WARN] Inspector Classic Assessment Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.Inspector, create.ErrActionReading, ResNameAssessmentTemplate, d.Id(), err)
	}

	arn := aws.ToString(template.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDuration, template.DurationInSeconds)
	d.Set(names.AttrName, template.Name)
	d.Set("rules_package_arns", template.RulesPackageArns)
	d.Set(names.AttrTargetARN, template.AssessmentTargetArn)

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
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

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
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	log.Printf("[INFO] Deleting Inspector Classic Assessment Template: %s", d.Id())
	_, err := conn.DeleteAssessmentTemplate(ctx, &inspector.DeleteAssessmentTemplateInput{
		AssessmentTemplateArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Inspector Classic Assessment Template (%s): %s", d.Id(), err)
	}

	return diags
}

func FindAssessmentTemplateByID(ctx context.Context, conn *inspector.Client, arn string) (*awstypes.AssessmentTemplate, error) {
	in := &inspector.DescribeAssessmentTargetsInput{
		AssessmentTargetArns: []string{arn},
	}

	out, err := conn.DescribeAssessmentTemplates(ctx, &inspector.DescribeAssessmentTemplatesInput{
		AssessmentTemplateArns: []string{arn},
	})

	if err != nil {
		return nil, err
	}

	if out.AssessmentTemplates == nil || len(out.AssessmentTemplates) == 0 {
		return nil, &retry.NotFoundError{
			LastRequest: in,
		}
	}

	if i := len(out.AssessmentTemplates); i > 1 {
		return nil, tfresource.NewTooManyResultsError(i, in)
	}

	return &out.AssessmentTemplates[0], nil
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
		Event:       awstypes.InspectorEvent(tfMap["event"].(string)),
		ResourceArn: templateArn,
		TopicArn:    aws.String(tfMap[names.AttrTopicARN].(string)),
	}

	return eventSubscription
}

func flattenSubscriptions(subscriptions []awstypes.Subscription) []interface{} {
	if len(subscriptions) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, subscription := range subscriptions {
		for _, eventSubscription := range subscription.EventSubscriptions {
			tfList = append(tfList, flattenEventSubscription(eventSubscription, subscription.TopicArn))
		}
	}

	return tfList
}

func flattenEventSubscription(eventSubscription awstypes.EventSubscription, topicArn *string) map[string]interface{} {
	tfMap := map[string]interface{}{}

	tfMap["event"] = eventSubscription.Event
	tfMap[names.AttrTopicARN] = topicArn

	return tfMap
}

func subscribeToEvents(ctx context.Context, conn *inspector.Client, eventSubscriptions []*inspector.SubscribeToEventInput) error {
	for _, eventSubscription := range eventSubscriptions {
		_, err := conn.SubscribeToEvent(ctx, eventSubscription)

		if err != nil {
			return create.Error(names.Inspector, create.ErrActionCreating, ResNameAssessmentTemplate, *eventSubscription.TopicArn, err)
		}
	}

	return nil
}

func unsubscribeFromEvents(ctx context.Context, conn *inspector.Client, eventSubscriptions []*inspector.SubscribeToEventInput) error {
	for _, eventSubscription := range eventSubscriptions {
		input := &inspector.UnsubscribeFromEventInput{
			Event:       eventSubscription.Event,
			ResourceArn: eventSubscription.ResourceArn,
			TopicArn:    eventSubscription.TopicArn,
		}

		_, err := conn.UnsubscribeFromEvent(ctx, input)

		if err != nil {
			return create.Error(names.Inspector, create.ErrActionDeleting, ResNameAssessmentTemplate, *eventSubscription.TopicArn, err)
		}
	}

	return nil
}
