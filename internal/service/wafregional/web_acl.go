// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_web_acl", name="Web ACL")
// @Tags(identifierAttribute="arn")
func resourceWebACL() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebACLCreate,
		ReadWithoutTimeout:   resourceWebACLRead,
		UpdateWithoutTimeout: resourceWebACLUpdate,
		DeleteWithoutTimeout: resourceWebACLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
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
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.WafActionType](),
						},
					},
				},
			},
			"logging_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_destination": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
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
												names.AttrType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.MatchFieldType](),
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
			"metric_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rule": {
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
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.WafActionType](),
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
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.WafOverrideActionType](),
									},
								},
							},
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.WafRuleTypeRegular,
							ValidateDiagFunc: enum.Validate[awstypes.WafRuleType](),
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
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.CreateWebACLInput{
			ChangeToken:   token,
			DefaultAction: ExpandAction(d.Get("default_action").([]interface{})),
			MetricName:    aws.String(d.Get("metric_name").(string)),
			Name:          aws.String(d.Get(names.AttrName).(string)),
			Tags:          getTagsIn(ctx),
		}

		return conn.CreateWebACL(ctx, input)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Web ACL (%s): %s", d.Get(names.AttrName).(string), err)
	}
	resp := out.(*wafregional.CreateWebACLOutput)
	d.SetId(aws.ToString(resp.WebACL.WebACLId))

	// The WAF API currently omits this, but use it when it becomes available
	webACLARN := aws.ToString(resp.WebACL.WebACLArn)
	if webACLARN == "" {
		webACLARN = arn.ARN{
			AccountID: meta.(*conns.AWSClient).AccountID,
			Partition: meta.(*conns.AWSClient).Partition,
			Region:    meta.(*conns.AWSClient).Region,
			Resource:  fmt.Sprintf("webacl/%s", d.Id()),
			Service:   "waf-regional",
		}.String()
	}

	loggingConfiguration := d.Get("logging_configuration").([]interface{})

	if len(loggingConfiguration) == 1 {
		input := &wafregional.PutLoggingConfigurationInput{
			LoggingConfiguration: expandLoggingConfiguration(loggingConfiguration, webACLARN),
		}

		log.Printf("[DEBUG] Updating WAF Regional Web ACL (%s)", d.Id())
		if _, err := conn.PutLoggingConfiguration(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), err)
		}
	}

	rules := d.Get("rule").(*schema.Set).List()
	if len(rules) > 0 {
		wr := NewRetryer(conn, region)
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			req := &wafregional.UpdateWebACLInput{
				ChangeToken:   token,
				DefaultAction: ExpandAction(d.Get("default_action").([]interface{})),
				Updates:       diffWebACLRules([]interface{}{}, rules),
				WebACLId:      aws.String(d.Id()),
			}
			return conn.UpdateWebACL(ctx, req)
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Web ACL (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWebACLRead(ctx, d, meta)...)
}

func resourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	params := &wafregional.GetWebACLInput{
		WebACLId: aws.String(d.Id()),
	}

	resp, err := conn.GetWebACL(ctx, params)
	if err != nil {
		if !d.IsNewResource() && errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] WAF Regional ACL (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "unable to read WAF Regional ACL (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() && (resp == nil || resp.WebACL == nil) {
		log.Printf("[WARN] WAF Regional ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// The WAF API currently omits this, but use it when it becomes available
	webACLARN := aws.ToString(resp.WebACL.WebACLArn)
	if webACLARN == "" {
		webACLARN = arn.ARN{
			AccountID: meta.(*conns.AWSClient).AccountID,
			Partition: meta.(*conns.AWSClient).Partition,
			Region:    meta.(*conns.AWSClient).Region,
			Resource:  fmt.Sprintf("webacl/%s", d.Id()),
			Service:   "waf-regional",
		}.String()
	}
	d.Set(names.AttrARN, webACLARN)

	if err := d.Set("default_action", FlattenAction(resp.WebACL.DefaultAction)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_action: %s", err)
	}
	d.Set(names.AttrName, resp.WebACL.Name)
	d.Set("metric_name", resp.WebACL.MetricName)
	if err := d.Set("rule", FlattenWebACLRules(resp.WebACL.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	getLoggingConfigurationInput := &wafregional.GetLoggingConfigurationInput{
		ResourceArn: aws.String(d.Get(names.AttrARN).(string)),
	}
	loggingConfiguration := []interface{}{}

	log.Printf("[DEBUG] Getting WAF Regional Web ACL (%s)", d.Id())
	getLoggingConfigurationOutput, err := conn.GetLoggingConfiguration(ctx, getLoggingConfigurationInput)

	if err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return sdkdiag.AppendErrorf(diags, "getting WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), err)
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
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChanges("default_action", "rule") {
		o, n := d.GetChange("rule")
		oldR, newR := o.(*schema.Set).List(), n.(*schema.Set).List()

		wr := NewRetryer(conn, region)
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			req := &wafregional.UpdateWebACLInput{
				ChangeToken:   token,
				DefaultAction: ExpandAction(d.Get("default_action").([]interface{})),
				Updates:       diffWebACLRules(oldR, newR),
				WebACLId:      aws.String(d.Id()),
			}
			return conn.UpdateWebACL(ctx, req)
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Web ACL (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("logging_configuration") {
		loggingConfiguration := d.Get("logging_configuration").([]interface{})

		if len(loggingConfiguration) == 1 {
			input := &wafregional.PutLoggingConfigurationInput{
				LoggingConfiguration: expandLoggingConfiguration(loggingConfiguration, d.Get(names.AttrARN).(string)),
			}

			log.Printf("[DEBUG] Updating WAF Regional Web ACL (%s)", d.Id())
			if _, err := conn.PutLoggingConfiguration(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating WAF Regional Web ACL (%s)", d.Id())
			}
		} else {
			input := &wafregional.DeleteLoggingConfigurationInput{
				ResourceArn: aws.String(d.Get(names.AttrARN).(string)),
			}

			log.Printf("[DEBUG] Deleting WAF Regional Web ACL (%s)", d.Id())
			if _, err := conn.DeleteLoggingConfiguration(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceWebACLRead(ctx, d, meta)...)
}

func resourceWebACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region
	wr := NewRetryer(conn, region)

	if rules := d.Get("rule").(*schema.Set).List(); len(rules) > 0 {
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			input := &wafregional.UpdateWebACLInput{
				ChangeToken:   token,
				DefaultAction: ExpandAction(d.Get("default_action").([]interface{})),
				Updates:       diffWebACLRules(rules, []interface{}{}),
				WebACLId:      aws.String(d.Id()),
			}

			return conn.UpdateWebACL(ctx, input)
		})

		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF Regional Web ACL (%s): %s", d.Id(), err)
			}
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Web ACL: %s", d.Id())
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &wafregional.DeleteWebACLInput{
			ChangeToken: token,
			WebACLId:    aws.String(d.Id()),
		}

		return conn.DeleteWebACL(ctx, req)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Web ACL (%s): %s", d.Id(), err)
	}

	return diags
}

func expandLoggingConfiguration(l []interface{}, resourceARN string) *awstypes.LoggingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	loggingConfiguration := &awstypes.LoggingConfiguration{
		LogDestinationConfigs: []string{
			m["log_destination"].(string),
		},
		RedactedFields: expandRedactedFields(m["redacted_fields"].([]interface{})),
		ResourceArn:    aws.String(resourceARN),
	}

	return loggingConfiguration
}

func expandRedactedFields(l []interface{}) []awstypes.FieldToMatch {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	if m["field_to_match"] == nil {
		return nil
	}

	redactedFields := make([]awstypes.FieldToMatch, 0)

	for _, fieldToMatch := range m["field_to_match"].(*schema.Set).List() {
		if fieldToMatch == nil {
			continue
		}

		redactedFields = append(redactedFields, *ExpandFieldToMatch(fieldToMatch.(map[string]interface{})))
	}

	return redactedFields
}

func flattenLoggingConfiguration(loggingConfiguration *awstypes.LoggingConfiguration) []interface{} {
	if loggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"log_destination": "",
		"redacted_fields": flattenRedactedFields(loggingConfiguration.RedactedFields),
	}

	if len(loggingConfiguration.LogDestinationConfigs) > 0 {
		m["log_destination"] = loggingConfiguration.LogDestinationConfigs[0]
	}

	return []interface{}{m}
}

func flattenRedactedFields(fieldToMatches []awstypes.FieldToMatch) []interface{} {
	if len(fieldToMatches) == 0 {
		return []interface{}{}
	}

	fieldToMatchResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
	l := make([]interface{}, len(fieldToMatches))

	for i, fieldToMatch := range fieldToMatches {
		l[i] = FlattenFieldToMatch(&fieldToMatch)[0]
	}

	m := map[string]interface{}{
		"field_to_match": schema.NewSet(schema.HashResource(fieldToMatchResource), l),
	}

	return []interface{}{m}
}

func diffWebACLRules(oldR, newR []interface{}) []awstypes.WebACLUpdate {
	updates := make([]awstypes.WebACLUpdate, 0)

	for _, or := range oldR {
		aclRule := or.(map[string]interface{})

		if idx, contains := sliceContainsMap(newR, aclRule); contains {
			newR = append(newR[:idx], newR[idx+1:]...)
			continue
		}
		updates = append(updates, ExpandWebACLUpdate(string(awstypes.ChangeActionDelete), aclRule))
	}

	for _, nr := range newR {
		aclRule := nr.(map[string]interface{})
		updates = append(updates, ExpandWebACLUpdate(string(awstypes.ChangeActionInsert), aclRule))
	}
	return updates
}
