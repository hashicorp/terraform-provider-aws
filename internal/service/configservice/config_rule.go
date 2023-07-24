// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_config_rule", name="Config Rule")
// @Tags(identifierAttribute="arn")
func ResourceConfigRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRulePutConfig,
		ReadWithoutTimeout:   resourceConfigRuleRead,
		UpdateWithoutTimeout: resourceRulePutConfig,
		DeleteWithoutTimeout: resourceConfigRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameConfigRule = "Config Rule"
)

func resourceRulePutConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

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
		Tags:       getTagsIn(ctx),
	}
	log.Printf("[DEBUG] Creating AWSConfig config rule: %s", input)
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.PutConfigRuleWithContext(ctx, &input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, configservice.ErrCodeInsufficientPermissionsException) {
				// IAM is eventually consistent
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(fmt.Errorf("Failed to create AWSConfig rule: %w", err))
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.PutConfigRuleWithContext(ctx, &input)
	}
	if err != nil {
		return append(diags, create.DiagError(names.ConfigService, create.ErrActionUpdating, ResNameConfigRule, name, err)...)
	}

	d.SetId(name)

	return append(diags, resourceConfigRuleRead(ctx, d, meta)...)
}

func resourceConfigRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	rule, err := FindConfigRule(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Config Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return append(diags, create.DiagError(names.ConfigService, create.ErrActionReading, ResNameConfigRule, d.Id(), err)...)
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

	return diags
}

func resourceConfigRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	name := d.Get("name").(string)

	log.Printf("[DEBUG] Deleting AWS Config config rule %q", name)
	input := &configservice.DeleteConfigRuleInput{
		ConfigRuleName: aws.String(name),
	}
	err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteConfigRuleWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, configservice.ErrCodeResourceInUseException) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteConfigRuleWithContext(ctx, input)
	}
	if err != nil {
		return append(diags, create.DiagError(names.ConfigService, create.ErrActionDeleting, ResNameConfigRule, d.Id(), err)...)
	}

	if _, err := waitRuleDeleted(ctx, conn, d.Id()); err != nil {
		return append(diags, create.DiagError(names.ConfigService, create.ErrActionWaitingForDeletion, ResNameConfigRule, d.Id(), err)...)
	}

	return diags
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
