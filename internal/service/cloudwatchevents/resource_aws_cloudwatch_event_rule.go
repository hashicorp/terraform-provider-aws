package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfevents "github.com/hashicorp/terraform-provider-aws/aws/internal/service/cloudwatchevents"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/cloudwatchevents/finder"
	iamwaiter "github.com/hashicorp/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	cloudWatchEventRuleDeleteRetryTimeout = 5 * time.Minute
)

func ResourceRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceRuleCreate,
		Read:   resourceRuleRead,
		Update: resourceRuleUpdate,
		Delete: resourceRuleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateCloudWatchEventRuleName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateCloudWatchEventRuleName,
			},
			"schedule_expression": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
				AtLeastOneOf: []string{"schedule_expression", "event_pattern"},
			},
			"event_bus_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validBusNameOrARN,
				Default:      tfevents.DefaultEventBusName,
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
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchEventsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

	input, err := buildPutRuleInputStruct(d, name)

	if err != nil {
		return err
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().CloudwatcheventsTags()
	}

	log.Printf("[DEBUG] Creating CloudWatch Events Rule: %s", input)
	// IAM Roles take some time to propagate
	err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		_, err = conn.PutRule(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "cannot be assumed by principal") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutRule(input)
	}

	if err != nil {
		return fmt.Errorf("error creating CloudWatch Events Rule (%s): %w", name, err)
	}

	d.SetId(tfevents.RuleCreateResourceID(aws.StringValue(input.EventBusName), aws.StringValue(input.Name)))

	return resourceRuleRead(d, meta)
}

func resourceRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchEventsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	eventBusName, ruleName, err := tfevents.RuleParseResourceID(d.Id())

	if err != nil {
		return err
	}

	output, err := finder.RuleByEventBusAndRuleNames(conn, eventBusName, ruleName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Events Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudWatch Events Rule (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(output.Arn)
	d.Set("arn", arn)
	d.Set("description", output.Description)
	if output.EventPattern != nil {
		pattern, err := structure.NormalizeJsonString(aws.StringValue(output.EventPattern))
		if err != nil {
			return fmt.Errorf("event pattern contains an invalid JSON: %w", err)
		}
		d.Set("event_pattern", pattern)
	}
	d.Set("name", output.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(output.Name)))
	d.Set("role_arn", output.RoleArn)
	d.Set("schedule_expression", output.ScheduleExpression)
	d.Set("event_bus_name", eventBusName) // Use event bus name from resource ID as API response may collapse any ARN.

	enabled, err := tfevents.RuleEnabledFromState(aws.StringValue(output.State))

	if err != nil {
		return err
	}

	d.Set("is_enabled", enabled)

	tags, err := tftags.CloudwatcheventsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for CloudWatch Events Rule (%s): %w", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchEventsConn

	_, ruleName, err := tfevents.RuleParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input, err := buildPutRuleInputStruct(d, ruleName)

	if err != nil {
		return err
	}

	// IAM Roles take some time to propagate
	err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.PutRule(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "cannot be assumed by principal") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutRule(input)
	}

	if err != nil {
		return fmt.Errorf("error updating CloudWatch Events Rule (%s): %w", d.Id(), err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.CloudwatcheventsUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating CloudwWatch Event Rule (%s) tags: %w", arn, err)
		}
	}

	return resourceRuleRead(d, meta)
}

func resourceRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchEventsConn

	eventBusName, ruleName, err := tfevents.RuleParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &events.DeleteRuleInput{
		Name: aws.String(ruleName),
	}
	if eventBusName != "" {
		input.EventBusName = aws.String(eventBusName)
	}

	log.Printf("[DEBUG] Deleting CloudWatch Events Rule: %s", d.Id())
	err = resource.Retry(cloudWatchEventRuleDeleteRetryTimeout, func() *resource.RetryError {
		_, err := conn.DeleteRule(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "Rule can't be deleted since it has targets") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteRule(input)
	}

	if tfawserr.ErrCodeEquals(err, events.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudWatch Events Rule (%s): %w", d.Id(), err)
	}

	return nil
}

func buildPutRuleInputStruct(d *schema.ResourceData, name string) (*events.PutRuleInput, error) {
	input := events.PutRuleInput{
		Name: aws.String(name),
	}
	var eventBusName string
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("event_bus_name"); ok {
		eventBusName = v.(string)
		input.EventBusName = aws.String(eventBusName)
	}
	if v, ok := d.GetOk("event_pattern"); ok {
		pattern, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("event pattern contains an invalid JSON: %w", err)
		}
		input.EventPattern = aws.String(pattern)
	}
	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("schedule_expression"); ok {
		input.ScheduleExpression = aws.String(v.(string))
	}

	input.State = aws.String(tfevents.RuleStateFromEnabled(d.Get("is_enabled").(bool)))

	return &input, nil
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
		if len(json) > 2048 {
			errors = append(errors, fmt.Errorf("%q cannot be longer than %d characters: %q", k, 2048, json))
		}
		return
	}
}
