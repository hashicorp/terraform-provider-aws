// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_event_rule", name="Rule")
// @Tags(identifierAttribute="arn")
func resourceRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleCreate,
		ReadWithoutTimeout:   resourceRuleRead,
		UpdateWithoutTimeout: resourceRuleUpdate,
		DeleteWithoutTimeout: resourceRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceRuleV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceRuleUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
				AtLeastOneOf: []string{names.AttrScheduleExpression, "event_pattern"},
				StateFunc: func(v interface{}) string {
					json, _ := ruleEventPatternJSONDecoder(v.(string))
					return json
				},
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_enabled": {
				Type:       schema.TypeBool,
				Optional:   true,
				Deprecated: `Use "state" instead`,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					rawPlan := d.GetRawPlan()
					rawIsEnabled := rawPlan.GetAttr("is_enabled")
					return rawIsEnabled.IsKnown() && rawIsEnabled.IsNull()
				},
				ConflictsWith: []string{
					names.AttrState,
				},
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validateRuleName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validateRuleName,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrScheduleExpression: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
				AtLeastOneOf: []string{names.AttrScheduleExpression, "event_pattern"},
			},
			names.AttrState: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.RuleState](),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					if oldValue != "" && newValue == "" {
						return true
					}
					return false
				},
				ConflictsWith: []string{
					"is_enabled",
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := expandPutRuleInput(d, name)
	input.Tags = getTagsIn(ctx)

	arn, err := retryPutRule(ctx, conn, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
		input.Tags = nil

		arn, err = retryPutRule(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Rule (%s): %s", name, err)
	}

	eventBusName, ruleName := aws.ToString(input.EventBusName), aws.ToString(input.Name)
	d.SetId(ruleCreateResourceID(eventBusName, ruleName))

	const (
		timeout = 2 * time.Minute
	)
	_, err = tfresource.RetryWhenNotFound(ctx, timeout, func() (interface{}, error) {
		return findRuleByTwoPartKey(ctx, conn, eventBusName, ruleName)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Rule (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, arn, tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
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
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	eventBusName, ruleName, err := ruleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findRuleByTwoPartKey(ctx, conn, eventBusName, ruleName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Rule (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(output.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("event_bus_name", eventBusName) // Use event bus name from resource ID as API response may collapse any ARN.
	if output.EventPattern != nil {
		pattern, err := ruleEventPatternJSONDecoder(aws.ToString(output.EventPattern))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		d.Set("event_pattern", pattern)
	}
	d.Set(names.AttrForceDestroy, d.Get(names.AttrForceDestroy).(bool))
	switch output.State {
	case types.RuleStateEnabled, types.RuleStateEnabledWithAllCloudtrailManagementEvents:
		d.Set("is_enabled", true)
	default:
		d.Set("is_enabled", false)
	}
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.Name)))
	d.Set(names.AttrRoleARN, output.RoleArn)
	d.Set(names.AttrScheduleExpression, output.ScheduleExpression)
	d.Set(names.AttrState, output.State)

	return diags
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, names.AttrForceDestroy) {
		_, ruleName, err := ruleParseResourceID(d.Id())
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
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	eventBusName, ruleName, err := ruleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &eventbridge.DeleteRuleInput{
		Name: aws.String(ruleName),
	}

	if eventBusName != "" {
		input.EventBusName = aws.String(eventBusName)
	}

	if v, ok := d.GetOk(names.AttrForceDestroy); ok {
		input.Force = v.(bool)
	}

	const (
		timeout = 5 * time.Minute
	)
	log.Printf("[DEBUG] Deleting EventBridge Rule: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, timeout, func() (interface{}, error) {
		return conn.DeleteRule(ctx, input)
	}, errCodeValidationException, "Rule can't be deleted since it has targets")

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func retryPutRule(ctx context.Context, conn *eventbridge.Client, input *eventbridge.PutRuleInput) (string, error) {
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.PutRule(ctx, input)
	}, errCodeValidationException, "cannot be assumed by principal")

	if err != nil {
		return "", err
	}

	return aws.ToString(outputRaw.(*eventbridge.PutRuleOutput).RuleArn), nil
}

func findRuleByTwoPartKey(ctx context.Context, conn *eventbridge.Client, eventBusName, ruleName string) (*eventbridge.DescribeRuleOutput, error) {
	input := &eventbridge.DescribeRuleInput{
		Name: aws.String(ruleName),
	}
	if eventBusName != "" {
		input.EventBusName = aws.String(eventBusName)
	}

	output, err := conn.DescribeRule(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

var (
	eventBusARNPattern     = regexache.MustCompile(`^arn:aws[\w-]*:events:[a-z]{2}-[a-z]+-[\w-]+:[0-9]{12}:event-bus\/[0-9A-Za-z_.-]+$`)
	partnerEventBusPattern = regexache.MustCompile(`^(?:arn:aws[\w-]*:events:[a-z]{2}-[a-z]+-[\w-]+:[0-9]{12}:event-bus\/)?aws\.partner(/[0-9A-Za-z_.-]+){2,}$`)
)

const ruleResourceIDSeparator = "/"

func ruleCreateResourceID(eventBusName, ruleName string) string {
	if eventBusName == "" || eventBusName == DefaultEventBusName {
		return ruleName
	}

	parts := []string{eventBusName, ruleName}
	id := strings.Join(parts, ruleResourceIDSeparator)

	return id
}

func ruleParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ruleResourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		return DefaultEventBusName, parts[0], nil
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}
	if len(parts) > 2 {
		i := strings.LastIndex(id, ruleResourceIDSeparator)
		eventBusName := id[:i]
		ruleName := id[i+1:]
		if eventBusARNPattern.MatchString(eventBusName) && ruleName != "" {
			return eventBusName, ruleName, nil
		}
		if partnerEventBusPattern.MatchString(eventBusName) && ruleName != "" {
			return eventBusName, ruleName, nil
		}
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EVENTBUSNAME%[2]sRULENAME or RULENAME", id, ruleResourceIDSeparator)
}

// ruleEventPatternJSONDecoder decodes unicode translation of <,>,&
func ruleEventPatternJSONDecoder(jsonString interface{}) (string, error) {
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

	if v, ok := d.GetOk(names.AttrDescription); ok {
		apiObject.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_bus_name"); ok {
		apiObject.EventBusName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_pattern"); ok {
		json, _ := ruleEventPatternJSONDecoder(v.(string))
		apiObject.EventPattern = aws.String(json)
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		apiObject.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrScheduleExpression); ok {
		apiObject.ScheduleExpression = aws.String(v.(string))
	}

	rawConfig := d.GetRawConfig()
	rawState := rawConfig.GetAttr(names.AttrState)
	if rawState.IsKnown() && !rawState.IsNull() {
		apiObject.State = types.RuleState(rawState.AsString())
	} else {
		rawIsEnabled := rawConfig.GetAttr("is_enabled")
		if rawIsEnabled.IsKnown() && !rawIsEnabled.IsNull() {
			if rawIsEnabled.True() {
				apiObject.State = types.RuleStateEnabled
			} else {
				apiObject.State = types.RuleStateDisabled
			}
		}
	}

	return apiObject
}

func validateEventPatternValue() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		json, err := ruleEventPatternJSONDecoder(v.(string))
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
