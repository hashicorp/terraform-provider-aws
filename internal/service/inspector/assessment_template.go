// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package inspector

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_inspector_assessment_template", name="Assessment Template")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/inspector/types;types.AssessmentTemplate")
// @Testing(preIdentityVersion="v6.4.0")
// @Testing(preCheck="testAccPreCheck")
// @Tags(identifierAttribute="id")
func resourceAssessmentTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAssessmentTemplateCreate,
		ReadWithoutTimeout:   resourceAssessmentTemplateRead,
		UpdateWithoutTimeout: resourceAssessmentTemplateUpdate,
		DeleteWithoutTimeout: resourceAssessmentTemplateDelete,

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
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTargetARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceAssessmentTemplateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := inspector.CreateAssessmentTemplateInput{
		AssessmentTargetArn:    aws.String(d.Get(names.AttrTargetARN).(string)),
		AssessmentTemplateName: aws.String(name),
		DurationInSeconds:      aws.Int32(int32(d.Get(names.AttrDuration).(int))),
		RulesPackageArns:       flex.ExpandStringValueSet(d.Get("rules_package_arns").(*schema.Set)),
	}

	output, err := conn.CreateAssessmentTemplate(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector Classic Assessment Template (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.AssessmentTemplateArn))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Inspector Classic Assessment Template (%s) tags: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("event_subscription"); ok && v.(*schema.Set).Len() > 0 {
		if err := subscribeToEvents(ctx, conn, expandEventSubscriptions(v.(*schema.Set).List(), d.Id())); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceAssessmentTemplateRead(ctx, d, meta)...)
}

func resourceAssessmentTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	template, err := findAssessmentTemplateByARN(ctx, conn, d.Id())

	if retry.NotFound(err) {
		log.Printf("[WARN] Inspector Classic Assessment Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector Classic Assessment Template (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(template.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDuration, template.DurationInSeconds)
	d.Set(names.AttrName, template.Name)
	d.Set("rules_package_arns", template.RulesPackageArns)
	d.Set(names.AttrTargetARN, template.AssessmentTargetArn)

	output, err := findSubscriptionsByAssessmentTemplateARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector Classic Assessment Template (%s) subscriptions: %s", d.Id(), err)
	}

	if err := d.Set("event_subscription", flattenSubscriptions(output)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting event_subscription: %s", err)
	}

	return diags
}

func resourceAssessmentTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	if d.HasChange("event_subscription") {
		o, n := d.GetChange("event_subscription")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := expandEventSubscriptions(ns.Difference(os).List(), d.Id()), expandEventSubscriptions(os.Difference(ns).List(), d.Id())

		if err := subscribeToEvents(ctx, conn, add); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if err := unsubscribeFromEvents(ctx, conn, del); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceAssessmentTemplateRead(ctx, d, meta)...)
}

func resourceAssessmentTemplateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorClient(ctx)

	log.Printf("[INFO] Deleting Inspector Classic Assessment Template: %s", d.Id())
	input := inspector.DeleteAssessmentTemplateInput{
		AssessmentTemplateArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteAssessmentTemplate(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Inspector Classic Assessment Template (%s): %s", d.Id(), err)
	}

	return diags
}

type eventSubscription struct {
	event       awstypes.InspectorEvent
	resourceARN string
	topicARN    string
}

func subscribeToEvents(ctx context.Context, conn *inspector.Client, es []eventSubscription) error {
	for _, e := range es {
		if err := subscribeToEvent(ctx, conn, e); err != nil {
			return err
		}
	}

	return nil
}

func subscribeToEvent(ctx context.Context, conn *inspector.Client, e eventSubscription) error {
	input := inspector.SubscribeToEventInput{
		Event:       e.event,
		ResourceArn: aws.String(e.resourceARN),
		TopicArn:    aws.String(e.topicARN),
	}
	_, err := conn.SubscribeToEvent(ctx, &input)

	if err != nil {
		return fmt.Errorf("subscribing Inspector Classic Assessment Template (%s) to event (%s): %w", e.resourceARN, e.event, err)
	}

	return nil
}

func unsubscribeFromEvents(ctx context.Context, conn *inspector.Client, es []eventSubscription) error {
	for _, e := range es {
		if err := unsubscribeFromEvent(ctx, conn, e); err != nil {
			return err
		}
	}

	return nil
}

func unsubscribeFromEvent(ctx context.Context, conn *inspector.Client, e eventSubscription) error {
	input := inspector.UnsubscribeFromEventInput{
		Event:       e.event,
		ResourceArn: aws.String(e.resourceARN),
		TopicArn:    aws.String(e.topicARN),
	}
	_, err := conn.UnsubscribeFromEvent(ctx, &input)

	if err != nil {
		return fmt.Errorf("unsubscribing Inspector Classic Assessment Template (%s) from event (%s): %w", e.resourceARN, e.event, err)
	}

	return nil
}

func findAssessmentTemplates(ctx context.Context, conn *inspector.Client, input *inspector.DescribeAssessmentTemplatesInput) ([]awstypes.AssessmentTemplate, error) {
	output, err := conn.DescribeAssessmentTemplates(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if err := failedItemsError(output.FailedItems); err != nil {
		return nil, err
	}

	return output.AssessmentTemplates, nil
}

func findAssessmentTemplate(ctx context.Context, conn *inspector.Client, input *inspector.DescribeAssessmentTemplatesInput) (*awstypes.AssessmentTemplate, error) {
	output, err := findAssessmentTemplates(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAssessmentTemplateByARN(ctx context.Context, conn *inspector.Client, arn string) (*awstypes.AssessmentTemplate, error) {
	input := inspector.DescribeAssessmentTemplatesInput{
		AssessmentTemplateArns: []string{arn},
	}

	output, err := findAssessmentTemplate(ctx, conn, &input)

	if tfawserr.ErrMessageContains(err, string(awstypes.FailedItemErrorCodeItemDoesNotExist), arn) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findSubscriptions(ctx context.Context, conn *inspector.Client, input *inspector.ListEventSubscriptionsInput) ([]awstypes.Subscription, error) {
	var output []awstypes.Subscription

	pages := inspector.NewListEventSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Subscriptions...)
	}

	return output, nil
}

func findSubscriptionsByAssessmentTemplateARN(ctx context.Context, conn *inspector.Client, arn string) ([]awstypes.Subscription, error) {
	input := inspector.ListEventSubscriptionsInput{
		ResourceArn: aws.String(arn),
	}

	return findSubscriptions(ctx, conn, &input)
}

func expandEventSubscriptions(tfList []any, templateARN string) []eventSubscription {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []eventSubscription

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, eventSubscription{
			event:       awstypes.InspectorEvent(tfMap["event"].(string)),
			resourceARN: templateARN,
			topicARN:    tfMap[names.AttrTopicARN].(string),
		})
	}

	return apiObjects
}

func flattenSubscriptions(subscriptions []awstypes.Subscription) []any {
	if len(subscriptions) == 0 {
		return nil
	}

	var tfList []any

	for _, subscription := range subscriptions {
		for _, eventSubscription := range subscription.EventSubscriptions {
			tfList = append(tfList, flattenEventSubscription(eventSubscription, subscription.TopicArn))
		}
	}

	return tfList
}

func flattenEventSubscription(eventSubscription awstypes.EventSubscription, topicArn *string) map[string]any {
	tfMap := map[string]any{}

	tfMap["event"] = eventSubscription.Event
	tfMap[names.AttrTopicARN] = topicArn

	return tfMap
}
