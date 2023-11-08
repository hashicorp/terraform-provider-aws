// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ruleCreateRetryTimeout = 2 * time.Minute
	ruleDeleteRetryTimeout = 5 * time.Minute
)

// @SDKResource("aws_cloudwatch_event_rule", name="Rule")
// @Tags(identifierAttribute="arn")
func ResourceRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleCreate,
		ReadWithoutTimeout:   resourceRuleRead,
		UpdateWithoutTimeout: resourceRuleUpdate,
		DeleteWithoutTimeout: resourceRuleDelete,

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
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"event_bus_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validBusNameOrARN,
				Default:      DefaultEventBusName,
			},
			"event_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateEventPatternValue(),
				AtLeastOneOf: []string{"schedule_expression", "event_pattern"},
				StateFunc: func(v interface{}) string {
					json, _ := RuleEventPatternJSONDecoder(v.(string))
					return json
				},
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateRuleName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateRuleName,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"schedule_expression": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
				AtLeastOneOf: []string{"schedule_expression", "event_pattern"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := expandPutRuleInput(d, name)
	input.Tags = getTagsIn(ctx)

	arn, err := retryPutRule(ctx, conn, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		arn, err = retryPutRule(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Rule (%s): %s", name, err)
	}

	eventBusName, ruleName := aws.StringValue(input.EventBusName), aws.StringValue(input.Name)
	d.SetId(RuleCreateResourceID(eventBusName, ruleName))

	_, err = tfresource.RetryWhenNotFound(ctx, ruleCreateRetryTimeout, func() (interface{}, error) {
		return FindRuleByTwoPartKey(ctx, conn, eventBusName, ruleName)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Rule (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, arn, tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceRuleRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting EventBridge Rule (%s) tags: %s", name, err)
		}
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	eventBusName, ruleName, err := RuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := FindRuleByTwoPartKey(ctx, conn, eventBusName, ruleName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Rule (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(output.Arn)
	d.Set("arn", arn)
	d.Set("description", output.Description)
	d.Set("event_bus_name", eventBusName) // Use event bus name from resource ID as API response may collapse any ARN.
	if output.EventPattern != nil {
		pattern, err := RuleEventPatternJSONDecoder(aws.StringValue(output.EventPattern))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "event pattern contains an invalid JSON: %s", err)
		}
		d.Set("event_pattern", pattern)
	}
	d.Set("is_enabled", aws.StringValue(output.State) == eventbridge.RuleStateEnabled)
	d.Set("name", output.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(output.Name)))
	d.Set("role_arn", output.RoleArn)
	d.Set("schedule_expression", output.ScheduleExpression)

	return diags
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		_, ruleName, err := RuleParseResourceID(d.Id())

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := expandPutRuleInput(d, ruleName)
		_, err = retryPutRule(ctx, conn, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Rule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	eventBusName, ruleName, err := RuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &eventbridge.DeleteRuleInput{
		Name: aws.String(ruleName),
	}
	if eventBusName != "" {
		input.EventBusName = aws.String(eventBusName)
	}

	log.Printf("[DEBUG] Deleting EventBridge Rule: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, ruleDeleteRetryTimeout, func() (interface{}, error) {
		return conn.DeleteRuleWithContext(ctx, input)
	}, "ValidationException", "Rule can't be deleted since it has targets")

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func retryPutRule(ctx context.Context, conn *eventbridge.EventBridge, input *eventbridge.PutRuleInput) (string, error) {
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.PutRuleWithContext(ctx, input)
	}, "ValidationException", "cannot be assumed by principal")

	if err != nil {
		return "", err
	}

	return aws.StringValue(outputRaw.(*eventbridge.PutRuleOutput).RuleArn), nil
}

func FindRuleByTwoPartKey(ctx context.Context, conn *eventbridge.EventBridge, eventBusName, ruleName string) (*eventbridge.DescribeRuleOutput, error) {
	input := eventbridge.DescribeRuleInput{
		Name: aws.String(ruleName),
	}
	if eventBusName != "" {
		input.EventBusName = aws.String(eventBusName)
	}

	output, err := conn.DescribeRuleWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// RuleEventPatternJSONDecoder decodes unicode translation of <,>,&
func RuleEventPatternJSONDecoder(jsonString interface{}) (string, error) {
	var j interface{}

	if jsonString == nil || jsonString.(string) == "" {
		return "", nil
	}

	s := jsonString.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return s, err
	}

	b, err := json.Marshal(j)
	if err != nil {
		return "", err
	}

	if bytes.Contains(b, []byte("\\u003c")) || bytes.Contains(b, []byte("\\u003e")) || bytes.Contains(b, []byte("\\u0026")) {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return string(b[:]), nil
}

func expandPutRuleInput(d *schema.ResourceData, name string) *eventbridge.PutRuleInput {
	apiObject := &eventbridge.PutRuleInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		apiObject.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_bus_name"); ok {
		apiObject.EventBusName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_pattern"); ok {
		json, _ := RuleEventPatternJSONDecoder(v.(string))
		apiObject.EventPattern = aws.String(json)
	}

	if v, ok := d.GetOk("role_arn"); ok {
		apiObject.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_expression"); ok {
		apiObject.ScheduleExpression = aws.String(v.(string))
	}

	state := eventbridge.RuleStateDisabled
	if d.Get("is_enabled").(bool) {
		state = eventbridge.RuleStateEnabled
	}
	apiObject.State = aws.String(state)

	return apiObject
}

func validateEventPatternValue() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		json, err := RuleEventPatternJSONDecoder(v.(string))
		if err != nil {
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %w", k, err))

			// Invalid JSON? Return immediately,
			// there is no need to collect other
			// errors.
			return
		}

		// Check whether the normalized JSON is within the given length.
		const maxJSONLength = 4096
		if len(json) > maxJSONLength {
			errors = append(errors, fmt.Errorf("%q cannot be longer than %d characters: %q", k, maxJSONLength, json))
		}
		return
	}
}
