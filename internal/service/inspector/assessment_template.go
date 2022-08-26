package inspector

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameAssessmentTemplate = "Assessment Template"
)

func ResourceAssessmentTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAssessmentTemplateCreate,
		Read:   resourceAssessmentTemplateRead,
		Update: resourceAssessmentTemplateUpdate,
		Delete: resourceAssessmentTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"duration": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"rules_package_arns": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAssessmentTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).InspectorConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &inspector.CreateAssessmentTemplateInput{
		AssessmentTargetArn:    aws.String(d.Get("target_arn").(string)),
		AssessmentTemplateName: aws.String(d.Get("name").(string)),
		DurationInSeconds:      aws.Int64(int64(d.Get("duration").(int))),
		RulesPackageArns:       flex.ExpandStringSet(d.Get("rules_package_arns").(*schema.Set)),
	}

	log.Printf("[DEBUG] Creating Inspector assessment template: %s", req)
	resp, err := conn.CreateAssessmentTemplate(req)
	if err != nil {
		return fmt.Errorf("error creating Inspector assessment template: %s", err)
	}

	d.SetId(aws.StringValue(resp.AssessmentTemplateArn))

	if len(tags) > 0 {
		if err := updateTags(conn, d.Id(), nil, tags); err != nil {
			return fmt.Errorf("error adding Inspector assessment template (%s) tags: %s", d.Id(), err)
		}
	}

	input := []*inspector.SubscribeToEventInput{}

	if v, ok := d.GetOk("event_subscription"); ok && v.(*schema.Set).Len() > 0 {
		input = expandEventSubscriptions(v.(*schema.Set).List(), resp.AssessmentTemplateArn)
	}

	if err := subscribeToEvents(conn, input); err != nil {
		return create.Error(names.Inspector, create.ErrActionCreating, ResNameAssessmentTemplate, d.Id(), err)
	}

	return resourceAssessmentTemplateRead(d, meta)
}

func resourceAssessmentTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).InspectorConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeAssessmentTemplates(&inspector.DescribeAssessmentTemplatesInput{
		AssessmentTemplateArns: aws.StringSlice([]string{d.Id()}),
	})
	if err != nil {
		return fmt.Errorf("error reading Inspector assessment template (%s): %s", d.Id(), err)
	}

	if resp.AssessmentTemplates == nil || len(resp.AssessmentTemplates) == 0 {
		log.Printf("[WARN] Inspector assessment template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	template := resp.AssessmentTemplates[0]

	arn := aws.StringValue(template.Arn)
	d.Set("arn", arn)
	d.Set("duration", template.DurationInSeconds)
	d.Set("name", template.Name)
	d.Set("target_arn", template.AssessmentTargetArn)

	if err := d.Set("rules_package_arns", flex.FlattenStringSet(template.RulesPackageArns)); err != nil {
		return fmt.Errorf("error setting rules_package_arns: %s", err)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Inspector assessment template (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	output, err := findSubscriptionsByAssessmentTemplateARN(conn, arn)

	if err != nil {
		return create.Error(names.Inspector, create.ErrActionReading, ResNameAssessmentTemplate, d.Id(), err)
	}

	if err := d.Set("event_subscription", flattenSubscriptions(output)); err != nil {
		return create.Error(names.Inspector, create.ErrActionSetting, ResNameAssessmentTemplate, d.Id(), err)
	}

	return nil
}

func resourceAssessmentTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).InspectorConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := updateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Inspector assessment template (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("event_subscription") {
		old, new := d.GetChange("event_subscription")
		oldSet := old.(*schema.Set)
		newSet := new.(*schema.Set)

		eventSubscriptionsToAdd := newSet.Difference(oldSet)
		eventSubscriptionsToRemove := oldSet.Difference(newSet)

		templateId := aws.String(d.Id())

		addEventSubscriptionsInput := expandEventSubscriptions(eventSubscriptionsToAdd.List(), templateId)
		removeEventSubscriptionsInput := expandEventSubscriptions(eventSubscriptionsToRemove.List(), templateId)

		if err := subscribeToEvents(conn, addEventSubscriptionsInput); err != nil {
			return create.Error(names.Inspector, create.ErrActionUpdating, ResNameAssessmentTemplate, d.Id(), err)
		}

		if err := unsubscribeFromEvents(conn, removeEventSubscriptionsInput); err != nil {
			return create.Error(names.Inspector, create.ErrActionUpdating, ResNameAssessmentTemplate, d.Id(), err)
		}
	}

	return resourceAssessmentTemplateRead(d, meta)
}

func resourceAssessmentTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).InspectorConn

	_, err := conn.DeleteAssessmentTemplate(&inspector.DeleteAssessmentTemplateInput{
		AssessmentTemplateArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleting Inspector assessment template (%s): %s", d.Id(), err)
	}

	return nil
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

func subscribeToEvents(conn *inspector.Inspector, eventSubscriptions []*inspector.SubscribeToEventInput) error {
	for _, eventSubscription := range eventSubscriptions {
		_, err := conn.SubscribeToEvent(eventSubscription)

		if err != nil {
			return create.Error(names.Inspector, create.ErrActionCreating, ResNameAssessmentTemplate, *eventSubscription.TopicArn, err)
		}
	}

	return nil
}

func unsubscribeFromEvents(conn *inspector.Inspector, eventSubscriptions []*inspector.SubscribeToEventInput) error {
	for _, eventSubscription := range eventSubscriptions {
		input := &inspector.UnsubscribeFromEventInput{
			Event:       eventSubscription.Event,
			ResourceArn: eventSubscription.ResourceArn,
			TopicArn:    eventSubscription.TopicArn,
		}

		_, err := conn.UnsubscribeFromEvent(input)

		if err != nil {
			return create.Error(names.Inspector, create.ErrActionDeleting, ResNameAssessmentTemplate, *eventSubscription.TopicArn, err)
		}
	}

	return nil
}
