// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_trigger", name="Trigger")
// @Tags(identifierAttribute="arn")
func ResourceTrigger() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTriggerCreate,
		ReadWithoutTimeout:   resourceTriggerRead,
		UpdateWithoutTimeout: resourceTriggerUpdate,
		DeleteWithoutTimeout: resourceTriggerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrActions: {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arguments": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"crawler_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"job_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrTimeout: {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"security_configuration": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"notification_property": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"notify_delay_after": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"event_batching_condition": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"batch_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 100),
						},
						"batch_window": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      900,
							ValidateFunc: validation.IntBetween(1, 900),
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"predicate": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"conditions": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"job_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"crawler_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"logical_operator": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          awstypes.LogicalOperatorEquals,
										ValidateDiagFunc: enum.Validate[awstypes.LogicalOperator](),
									},
									names.AttrState: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.JobRunState](),
									},
									"crawl_state": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.CrawlState](),
									},
								},
							},
						},
						"logical": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.LogicalAnd,
							ValidateDiagFunc: enum.Validate[awstypes.Logical](),
						},
					},
				},
			},
			names.AttrSchedule: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_on_creation": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TriggerType](),
			},
			"workflow_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTriggerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	name := d.Get(names.AttrName).(string)
	triggerType := d.Get(names.AttrType).(string)
	input := &glue.CreateTriggerInput{
		Actions:         expandActions(d.Get(names.AttrActions).([]interface{})),
		Name:            aws.String(name),
		Tags:            getTagsIn(ctx),
		Type:            awstypes.TriggerType(triggerType),
		StartOnCreation: d.Get("start_on_creation").(bool),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_batching_condition"); ok {
		input.EventBatchingCondition = expandEventBatchingCondition(v.([]interface{}))
	}

	if v, ok := d.GetOk("predicate"); ok {
		input.Predicate = expandPredicate(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrSchedule); ok {
		input.Schedule = aws.String(v.(string))
	}

	if v, ok := d.GetOk("workflow_name"); ok {
		input.WorkflowName = aws.String(v.(string))
	}

	if d.Get(names.AttrEnabled).(bool) && triggerType != string(awstypes.TriggerTypeOnDemand) {
		start := true

		if triggerType == string(awstypes.TriggerTypeEvent) {
			start = false
		}

		input.StartOnCreation = start
	}

	if v, ok := d.GetOk("workflow_name"); ok {
		input.WorkflowName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("start_on_creation"); ok {
		input.StartOnCreation = v.(bool)
	}

	log.Printf("[DEBUG] Creating Glue Trigger: %+v", input)
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.CreateTrigger(ctx, input)
		if err != nil {
			// Retry IAM propagation errors
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Service is unable to assume provided role") {
				return retry.RetryableError(err)
			}
			// Retry concurrent workflow modification errors
			if errs.IsAErrorMessageContains[*awstypes.ConcurrentModificationException](err, "was modified while adding trigger") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.CreateTrigger(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Trigger (%s): %s", name, err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] Waiting for Glue Trigger (%s) to create", d.Id())
	if _, err := waitTriggerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for Glue Trigger (%s) to be Created: %s", d.Id(), err)
	}

	if d.Get(names.AttrEnabled).(bool) && triggerType == string(awstypes.TriggerTypeOnDemand) {
		input := &glue.StartTriggerInput{
			Name: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Starting Glue Trigger: %+v", input)
		_, err := conn.StartTrigger(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "starting Glue Trigger (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTriggerRead(ctx, d, meta)...)
}

func resourceTriggerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	output, err := FindTriggerByName(ctx, conn, d.Id())
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			log.Printf("[WARN] Glue Trigger (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Glue Trigger (%s): %s", d.Id(), err)
	}

	trigger := output.Trigger
	if trigger == nil {
		log.Printf("[WARN] Glue Trigger (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err := d.Set(names.AttrActions, flattenActions(trigger.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting actions: %s", err)
	}

	triggerARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("trigger/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, triggerARN)

	d.Set(names.AttrDescription, trigger.Description)

	var enabled bool
	d.Set(names.AttrState, string(trigger.State))

	if trigger.Type == awstypes.TriggerTypeOnDemand || trigger.Type == awstypes.TriggerTypeEvent {
		enabled = (trigger.State == awstypes.TriggerStateCreated || trigger.State == awstypes.TriggerStateCreating) && d.Get(names.AttrEnabled).(bool)
	} else {
		enabled = (trigger.State == awstypes.TriggerStateActivated || trigger.State == awstypes.TriggerStateActivating)
	}
	d.Set(names.AttrEnabled, enabled)

	if err := d.Set("predicate", flattenPredicate(trigger.Predicate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting predicate: %s", err)
	}

	if err := d.Set("event_batching_condition", flattenEventBatchingCondition(trigger.EventBatchingCondition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting event_batching_condition: %s", err)
	}

	d.Set(names.AttrName, trigger.Name)
	d.Set(names.AttrSchedule, trigger.Schedule)
	d.Set(names.AttrType, trigger.Type)
	d.Set("workflow_name", trigger.WorkflowName)

	return diags
}

func resourceTriggerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChanges(names.AttrActions, names.AttrDescription, "predicate", names.AttrSchedule, "event_batching_condition") {
		triggerUpdate := &awstypes.TriggerUpdate{
			Actions: expandActions(d.Get(names.AttrActions).([]interface{})),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			triggerUpdate.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("predicate"); ok {
			triggerUpdate.Predicate = expandPredicate(v.([]interface{}))
		}

		if v, ok := d.GetOk(names.AttrSchedule); ok {
			triggerUpdate.Schedule = aws.String(v.(string))
		}

		if v, ok := d.GetOk("event_batching_condition"); ok {
			triggerUpdate.EventBatchingCondition = expandEventBatchingCondition(v.([]interface{}))
		}

		input := &glue.UpdateTriggerInput{
			Name:          aws.String(d.Id()),
			TriggerUpdate: triggerUpdate,
		}

		log.Printf("[DEBUG] Updating Glue Trigger: %+v", input)
		_, err := conn.UpdateTrigger(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Trigger (%s): %s", d.Id(), err)
		}

		if _, err := waitTriggerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Glue Trigger (%s) to be Update: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrEnabled) {
		if d.Get(names.AttrEnabled).(bool) {
			input := &glue.StartTriggerInput{
				Name: aws.String(d.Id()),
			}

			log.Printf("[DEBUG] Starting Glue Trigger: %+v", input)
			_, err := conn.StartTrigger(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "starting Glue Trigger (%s): %s", d.Id(), err)
			}
		} else {
			//Skip if Trigger is type is ON_DEMAND and is in CREATED state as this means the trigger is not running or has ran already.
			if !(d.Get(names.AttrType).(string) == string(awstypes.TriggerTypeOnDemand) && d.Get(names.AttrState).(string) == string(awstypes.TriggerStateCreated)) {
				input := &glue.StopTriggerInput{
					Name: aws.String(d.Id()),
				}

				log.Printf("[DEBUG] Stopping Glue Trigger: %+v", input)
				_, err := conn.StopTrigger(ctx, input)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "stopping Glue Trigger (%s): %s", d.Id(), err)
				}
			}
		}
	}

	return diags
}

func resourceTriggerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Deleting Glue Trigger: %s", d.Id())
	err := deleteTrigger(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Trigger (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for Glue Trigger (%s) to delete", d.Id())
	if _, err := waitTriggerDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for Glue Trigger (%s) to be Deleted: %s", d.Id(), err)
	}

	return diags
}

func deleteTrigger(ctx context.Context, conn *glue.Client, Name string) error {
	input := &glue.DeleteTriggerInput{
		Name: aws.String(Name),
	}

	_, err := conn.DeleteTrigger(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil
		}
		return err
	}

	return nil
}

func expandActions(l []interface{}) []awstypes.Action {
	actions := []awstypes.Action{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})

		action := awstypes.Action{}

		if v, ok := m["crawler_name"].(string); ok && v != "" {
			action.CrawlerName = aws.String(v)
		}

		if v, ok := m["job_name"].(string); ok && v != "" {
			action.JobName = aws.String(v)
		}

		if v, ok := m["arguments"].(map[string]interface{}); ok && len(v) > 0 {
			action.Arguments = flex.ExpandStringValueMap(v)
		}

		if v, ok := m[names.AttrTimeout].(int); ok && v > 0 {
			action.Timeout = aws.Int32(int32(v))
		}

		if v, ok := m["security_configuration"].(string); ok && v != "" {
			action.SecurityConfiguration = aws.String(v)
		}

		if v, ok := m["notification_property"].([]interface{}); ok && len(v) > 0 {
			action.NotificationProperty = expandTriggerNotificationProperty(v)
		}

		actions = append(actions, action)
	}

	return actions
}

func expandTriggerNotificationProperty(l []interface{}) *awstypes.NotificationProperty {
	m := l[0].(map[string]interface{})

	property := &awstypes.NotificationProperty{}

	if v, ok := m["notify_delay_after"].(int); ok && v > 0 {
		property.NotifyDelayAfter = aws.Int32(int32(v))
	}

	return property
}

func expandConditions(l []interface{}) []awstypes.Condition {
	conditions := []awstypes.Condition{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})

		condition := awstypes.Condition{
			LogicalOperator: awstypes.LogicalOperator(m["logical_operator"].(string)),
		}

		if v, ok := m["crawler_name"].(string); ok && v != "" {
			condition.CrawlerName = aws.String(v)
		}

		if v, ok := m["crawl_state"].(string); ok && v != "" {
			condition.CrawlState = awstypes.CrawlState(v)
		}

		if v, ok := m["job_name"].(string); ok && v != "" {
			condition.JobName = aws.String(v)
		}

		if v, ok := m[names.AttrState].(string); ok && v != "" {
			condition.State = awstypes.JobRunState(v)
		}

		conditions = append(conditions, condition)
	}

	return conditions
}

func expandPredicate(l []interface{}) *awstypes.Predicate {
	m := l[0].(map[string]interface{})

	predicate := &awstypes.Predicate{
		Conditions: expandConditions(m["conditions"].([]interface{})),
	}

	if v, ok := m["logical"].(string); ok && v != "" {
		predicate.Logical = awstypes.Logical(v)
	}

	return predicate
}

func flattenActions(actions []awstypes.Action) []interface{} {
	l := []interface{}{}

	for _, action := range actions {
		m := map[string]interface{}{
			"arguments":       action.Arguments,
			names.AttrTimeout: int(aws.ToInt32(action.Timeout)),
		}

		if v := aws.ToString(action.CrawlerName); v != "" {
			m["crawler_name"] = v
		}

		if v := aws.ToString(action.JobName); v != "" {
			m["job_name"] = v
		}

		if v := aws.ToString(action.SecurityConfiguration); v != "" {
			m["security_configuration"] = v
		}

		if v := action.NotificationProperty; v != nil {
			m["notification_property"] = flattenTriggerNotificationProperty(v)
		}

		l = append(l, m)
	}

	return l
}

func flattenConditions(conditions []awstypes.Condition) []interface{} {
	l := []interface{}{}

	for _, condition := range conditions {
		m := map[string]interface{}{
			"logical_operator": string(condition.LogicalOperator),
		}

		if v := aws.ToString(condition.CrawlerName); v != "" {
			m["crawler_name"] = v
		}

		if v := string(condition.CrawlState); v != "" {
			m["crawl_state"] = v
		}

		if v := aws.ToString(condition.JobName); v != "" {
			m["job_name"] = v
		}

		if v := string(condition.State); v != "" {
			m[names.AttrState] = v
		}

		l = append(l, m)
	}

	return l
}

func flattenPredicate(predicate *awstypes.Predicate) []map[string]interface{} {
	if predicate == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"conditions": flattenConditions(predicate.Conditions),
		"logical":    string(predicate.Logical),
	}

	return []map[string]interface{}{m}
}

func flattenTriggerNotificationProperty(property *awstypes.NotificationProperty) []map[string]interface{} {
	if property == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"notify_delay_after": aws.ToInt32(property.NotifyDelayAfter),
	}

	return []map[string]interface{}{m}
}

func expandEventBatchingCondition(l []interface{}) *awstypes.EventBatchingCondition {
	m := l[0].(map[string]interface{})

	ebc := &awstypes.EventBatchingCondition{
		BatchSize: aws.Int32(int32(m["batch_size"].(int))),
	}

	if v, ok := m["batch_window"].(int); ok && v > 0 {
		ebc.BatchWindow = aws.Int32(int32(v))
	}

	return ebc
}

func flattenEventBatchingCondition(ebc *awstypes.EventBatchingCondition) []map[string]interface{} {
	if ebc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"batch_size":   aws.ToInt32(ebc.BatchSize),
		"batch_window": aws.ToInt32(ebc.BatchWindow),
	}

	return []map[string]interface{}{m}
}
