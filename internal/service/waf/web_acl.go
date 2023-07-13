// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_web_acl", name="Web ACL")
// @Tags(identifierAttribute="arn")
func ResourceWebACL() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebACLCreate,
		ReadWithoutTimeout:   resourceWebACLRead,
		UpdateWithoutTimeout: resourceWebACLUpdate,
		DeleteWithoutTimeout: resourceWebACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"metric_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z]+$`), "must contain only alphanumeric characters"),
			},
			"logging_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_destination": {
							Type:     schema.TypeString,
							Required: true,
						},
						"redacted_fields": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field_to_match": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"data": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"type": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"rules": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"override_action": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      waf.WafRuleTypeRegular,
							ValidateFunc: validation.StringInSlice(waf.WafRuleType_Values(), false),
						},
						"rule_id": {
							Type:     schema.TypeString,
							Required: true,
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

func resourceWebACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn(ctx)

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateWebACLInput{
			ChangeToken:   token,
			DefaultAction: ExpandAction(d.Get("default_action").([]interface{})),
			MetricName:    aws.String(d.Get("metric_name").(string)),
			Name:          aws.String(d.Get("name").(string)),
			Tags:          getTagsIn(ctx),
		}

		return conn.CreateWebACLWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Web ACL (%s): %s", d.Get("name").(string), err)
	}

	resp := out.(*waf.CreateWebACLOutput)
	d.SetId(aws.StringValue(resp.WebACL.WebACLId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("webacl/%s", d.Id()),
	}.String()

	loggingConfiguration := d.Get("logging_configuration").([]interface{})
	if len(loggingConfiguration) == 1 {
		input := &waf.PutLoggingConfigurationInput{
			LoggingConfiguration: expandLoggingConfiguration(loggingConfiguration, arn),
		}

		if _, err := conn.PutLoggingConfigurationWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "putting WAF Web ACL (%s) Logging Configuration: %s", d.Id(), err)
		}
	}

	rules := d.Get("rules").(*schema.Set).List()
	if len(rules) > 0 {
		wr := NewRetryer(conn)
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			req := &waf.UpdateWebACLInput{
				ChangeToken:   token,
				DefaultAction: ExpandAction(d.Get("default_action").([]interface{})),
				Updates:       diffWebACLRules([]interface{}{}, rules),
				WebACLId:      aws.String(d.Id()),
			}
			return conn.UpdateWebACLWithContext(ctx, req)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Web ACL (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWebACLRead(ctx, d, meta)...)
}

func resourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn(ctx)

	params := &waf.GetWebACLInput{
		WebACLId: aws.String(d.Id()),
	}

	resp, err := conn.GetWebACLWithContext(ctx, params)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
		log.Printf("[WARN] WAF Web ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Web ACL (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.WebACL == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading WAF Web ACL (%s): not found", d.Id())
		}

		log.Printf("[WARN] WAF Web ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", resp.WebACL.WebACLArn)
	arn := aws.StringValue(resp.WebACL.WebACLArn)

	if err := d.Set("default_action", FlattenAction(resp.WebACL.DefaultAction)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_action: %s", err)
	}
	d.Set("name", resp.WebACL.Name)
	d.Set("metric_name", resp.WebACL.MetricName)
	if err := d.Set("rules", FlattenWebACLRules(resp.WebACL.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rules: %s", err)
	}

	getLoggingConfigurationInput := &waf.GetLoggingConfigurationInput{
		ResourceArn: aws.String(arn),
	}
	loggingConfiguration := []interface{}{}

	getLoggingConfigurationOutput, err := conn.GetLoggingConfigurationWithContext(ctx, getLoggingConfigurationInput)

	if err != nil && !tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
		return sdkdiag.AppendErrorf(diags, "reading WAF Web ACL (%s) Logging Configuration: %s", d.Id(), err)
	}

	if getLoggingConfigurationOutput != nil {
		loggingConfiguration = flattenLoggingConfiguration(getLoggingConfigurationOutput.LoggingConfiguration)
	}

	if err := d.Set("logging_configuration", loggingConfiguration); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logging_configuration: %s", err)
	}

	return diags
}

func resourceWebACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn(ctx)

	if d.HasChanges("default_action", "rules") {
		o, n := d.GetChange("rules")
		oldR, newR := o.(*schema.Set).List(), n.(*schema.Set).List()

		wr := NewRetryer(conn)
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			req := &waf.UpdateWebACLInput{
				ChangeToken:   token,
				DefaultAction: ExpandAction(d.Get("default_action").([]interface{})),
				Updates:       diffWebACLRules(oldR, newR),
				WebACLId:      aws.String(d.Id()),
			}
			return conn.UpdateWebACLWithContext(ctx, req)
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Web ACL (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("logging_configuration") {
		loggingConfiguration := d.Get("logging_configuration").([]interface{})

		if len(loggingConfiguration) == 1 {
			input := &waf.PutLoggingConfigurationInput{
				LoggingConfiguration: expandLoggingConfiguration(loggingConfiguration, d.Get("arn").(string)),
			}

			if _, err := conn.PutLoggingConfigurationWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating WAF Web ACL (%s) Logging Configuration: %s", d.Id(), err)
			}
		} else {
			input := &waf.DeleteLoggingConfigurationInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
			}

			if _, err := conn.DeleteLoggingConfigurationWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting WAF Web ACL (%s) Logging Configuration: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceWebACLRead(ctx, d, meta)...)
}

func resourceWebACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn(ctx)

	// First, need to delete all rules
	rules := d.Get("rules").(*schema.Set).List()
	if len(rules) > 0 {
		wr := NewRetryer(conn)
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			req := &waf.UpdateWebACLInput{
				ChangeToken:   token,
				DefaultAction: ExpandAction(d.Get("default_action").([]interface{})),
				Updates:       diffWebACLRules(rules, []interface{}{}),
				WebACLId:      aws.String(d.Id()),
			}
			return conn.UpdateWebACLWithContext(ctx, req)
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "removing WAF Web ACL (%s) rules: %s", d.Id(), err)
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteWebACLInput{
			ChangeToken: token,
			WebACLId:    aws.String(d.Id()),
		}

		return conn.DeleteWebACLWithContext(ctx, req)
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting WAF Web ACL (%s): %s", d.Id(), err)
	}

	return diags
}

func expandLoggingConfiguration(l []interface{}, resourceARN string) *waf.LoggingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	loggingConfiguration := &waf.LoggingConfiguration{
		LogDestinationConfigs: []*string{
			aws.String(m["log_destination"].(string)),
		},
		RedactedFields: expandRedactedFields(m["redacted_fields"].([]interface{})),
		ResourceArn:    aws.String(resourceARN),
	}

	return loggingConfiguration
}

func expandRedactedFields(l []interface{}) []*waf.FieldToMatch {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	if m["field_to_match"] == nil {
		return nil
	}

	redactedFields := make([]*waf.FieldToMatch, 0)

	for _, fieldToMatch := range m["field_to_match"].(*schema.Set).List() {
		if fieldToMatch == nil {
			continue
		}

		redactedFields = append(redactedFields, ExpandFieldToMatch(fieldToMatch.(map[string]interface{})))
	}

	return redactedFields
}

func flattenLoggingConfiguration(loggingConfiguration *waf.LoggingConfiguration) []interface{} {
	if loggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"log_destination": "",
		"redacted_fields": flattenRedactedFields(loggingConfiguration.RedactedFields),
	}

	if len(loggingConfiguration.LogDestinationConfigs) > 0 {
		m["log_destination"] = aws.StringValue(loggingConfiguration.LogDestinationConfigs[0])
	}

	return []interface{}{m}
}

func flattenRedactedFields(fieldToMatches []*waf.FieldToMatch) []interface{} {
	if len(fieldToMatches) == 0 {
		return []interface{}{}
	}

	fieldToMatchResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
	l := make([]interface{}, len(fieldToMatches))

	for i, fieldToMatch := range fieldToMatches {
		l[i] = FlattenFieldToMatch(fieldToMatch)[0]
	}

	m := map[string]interface{}{
		"field_to_match": schema.NewSet(schema.HashResource(fieldToMatchResource), l),
	}

	return []interface{}{m}
}

func diffWebACLRules(oldR, newR []interface{}) []*waf.WebACLUpdate {
	updates := make([]*waf.WebACLUpdate, 0)

	for _, or := range oldR {
		aclRule := or.(map[string]interface{})

		if idx, contains := sliceContainsMap(newR, aclRule); contains {
			newR = append(newR[:idx], newR[idx+1:]...)
			continue
		}
		updates = append(updates, ExpandWebACLUpdate(waf.ChangeActionDelete, aclRule))
	}

	for _, nr := range newR {
		aclRule := nr.(map[string]interface{})
		updates = append(updates, ExpandWebACLUpdate(waf.ChangeActionInsert, aclRule))
	}
	return updates
}
