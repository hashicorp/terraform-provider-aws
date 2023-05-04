package events

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
					json, _ := structure.NormalizeJsonString(v.(string))
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
	conn := meta.(*conns.AWSClient).EventsConn()

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := expandPutRuleInput(d, name)
	input.Tags = GetTagsIn(ctx)

	arn, err := retryPutRule(ctx, conn, input)

	// Some partitions may not support tag-on-create
	if input.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] EventBridge Rule (%s) create failed (%s) with tags. Trying create without tags.", name, err)
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

	// Post-create tagging supported in some partitions
	if tags := KeyValueTags(ctx, GetTagsIn(ctx)); input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(ctx, conn, arn, nil, tags)

		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] error adding tags after create for EventBridge Rule (%s): %s", d.Id(), err)
			return append(diags, resourceRuleRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EventBridge Rule (%s) tags: %s", name, err)
		}
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

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
		pattern, err := structure.NormalizeJsonString(aws.StringValue(output.EventPattern))
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
	conn := meta.(*conns.AWSClient).EventsConn()

	_, ruleName, err := RuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := expandPutRuleInput(d, ruleName)
	_, err = retryPutRule(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge Rule (%s): %s", d.Id(), err)
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

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
		json, _ := structure.NormalizeJsonString(v)
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
		json, err := structure.NormalizeJsonString(v)
		if err != nil {
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %w", k, err))

			// Invalid JSON? Return immediately,
			// there is no need to collect other
			// errors.
			return
		}

		// Check whether the normalized JSON is within the given length.
		const maxJSONLength = 2048
		if len(json) > maxJSONLength {
			errors = append(errors, fmt.Errorf("%q cannot be longer than %d characters: %q", k, maxJSONLength, json))
		}
		return
	}
}
