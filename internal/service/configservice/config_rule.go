package configservice

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceConfigRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceRulePutConfig,
		Read:   resourceConfigRuleRead,
		Update: resourceRulePutConfig,
		Delete: resourceConfigRuleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"rule_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"input_parameters": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"maximum_execution_frequency": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(configservice.MaximumExecutionFrequency_Values(), false),
			},
			"scope": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compliance_resource_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"compliance_resource_types": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 100,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 256),
							},
							Set: schema.HashString,
						},
						"tag_key": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 128),
						},
						"tag_value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
					},
				},
			},
			"source": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_policy_details": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable_debug_log_delivery": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"policy_runtime": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 64),
											validation.StringMatch(regexp.MustCompile(`^guard\-2\.x\.x$`), "Must match cloudformation-guard version"),
										),
									},
									"policy_text": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 10000),
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { // policy_text always returns empty
											if d.Id() != "" && old == "" {
												return true
											}
											return false
										},
									},
								},
							},
						},
						"owner": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(configservice.Owner_Values(), false),
						},
						"source_detail": {
							Type:     schema.TypeSet,
							Set:      ruleSourceDetailsHash,
							Optional: true,
							MaxItems: 25,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_source": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "aws.config",
										ValidateFunc: validation.StringInSlice(configservice.EventSource_Values(), false),
									},
									"maximum_execution_frequency": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(configservice.MaximumExecutionFrequency_Values(), false),
									},
									"message_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(configservice.MessageType_Values(), false),
									},
								},
							},
						},
						"source_identifier": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
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

func resourceRulePutConfig(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	ruleInput := configservice.ConfigRule{
		ConfigRuleName: aws.String(name),
		Scope:          expandRuleScope(d.Get("scope").([]interface{})),
		Source:         expandRuleSource(d.Get("source").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		ruleInput.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("input_parameters"); ok {
		ruleInput.InputParameters = aws.String(v.(string))
	}
	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		ruleInput.MaximumExecutionFrequency = aws.String(v.(string))
	}

	input := configservice.PutConfigRuleInput{
		ConfigRule: &ruleInput,
		Tags:       Tags(tags.IgnoreAWS()),
	}
	log.Printf("[DEBUG] Creating AWSConfig config rule: %s", input)
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.PutConfigRule(&input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, configservice.ErrCodeInsufficientPermissionsException) {
				// IAM is eventually consistent
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(fmt.Errorf("Failed to create AWSConfig rule: %w", err))
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.PutConfigRule(&input)
	}
	if err != nil {
		return fmt.Errorf("Error creating AWSConfig rule: %w", err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] AWSConfig config rule %q created", name)

	if !d.IsNewResource() && d.HasChange("tags_all") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Config Config Rule (%s) tags: %w", arn, err)
		}
	}

	return resourceConfigRuleRead(d, meta)
}

func resourceConfigRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	rule, err := FindConfigRule(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Config Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(rule.ConfigRuleArn)
	d.Set("arn", arn)
	d.Set("rule_id", rule.ConfigRuleId)
	d.Set("name", rule.ConfigRuleName)
	d.Set("description", rule.Description)
	d.Set("input_parameters", rule.InputParameters)
	d.Set("maximum_execution_frequency", rule.MaximumExecutionFrequency)

	if rule.Scope != nil {
		d.Set("scope", flattenRuleScope(rule.Scope))
	}

	d.Set("source", flattenRuleSource(rule.Source))

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Config Config Rule (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceConfigRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	name := d.Get("name").(string)

	log.Printf("[DEBUG] Deleting AWS Config config rule %q", name)
	input := &configservice.DeleteConfigRuleInput{
		ConfigRuleName: aws.String(name),
	}
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteConfigRule(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, configservice.ErrCodeResourceInUseException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteConfigRule(input)
	}
	if err != nil {
		return fmt.Errorf("error Deleting Config Rule failed: %w", err)
	}

	if _, err := waitRuleDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Config Service Rule (%s) to deleted: %w", d.Id(), err)
	}

	return nil
}

func ruleSourceDetailsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["message_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["event_source"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["maximum_execution_frequency"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return create.StringHashcode(buf.String())
}
