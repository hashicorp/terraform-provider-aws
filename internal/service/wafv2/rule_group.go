// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ruleGroupCreateTimeout = 5 * time.Minute
	ruleGroupUpdateTimeout = 5 * time.Minute
	ruleGroupDeleteTimeout = 5 * time.Minute
)

// @SDKResource("aws_wafv2_rule_group", name="Rule Group")
// @Tags(identifierAttribute="arn")
func ResourceRuleGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleGroupCreate,
		ReadWithoutTimeout:   resourceRuleGroupRead,
		UpdateWithoutTimeout: resourceRuleGroupUpdate,
		DeleteWithoutTimeout: resourceRuleGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"capacity": {
					Type:         schema.TypeInt,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.IntAtLeast(1),
				},
				"custom_response_body": customResponseBodySchema(),
				"description": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 256),
				},
				"lock_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), "must contain only alphanumeric hyphen and underscore characters"),
					),
				},
				"rule": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"action": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"allow":     allowConfigSchema(),
										"block":     blockConfigSchema(),
										"captcha":   captchaConfigSchema(),
										"challenge": challengeConfigSchema(),
										"count":     countConfigSchema(),
									},
								},
							},
							"captcha_config": outerCaptchaConfigSchema(),
							"name": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 128),
							},
							"priority": {
								Type:     schema.TypeInt,
								Required: true,
							},
							"rule_label":        ruleLabelsSchema(),
							"statement":         ruleGroupRootStatementSchema(ruleGroupRootStatementSchemaLevel),
							"visibility_config": visibilityConfigSchema(),
						},
					},
				},
				"scope": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice(wafv2.Scope_Values(), false),
				},
				names.AttrTags:      tftags.TagsSchema(),
				names.AttrTagsAll:   tftags.TagsSchemaComputed(),
				"visibility_config": visibilityConfigSchema(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRuleGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	name := d.Get("name").(string)
	input := &wafv2.CreateRuleGroupInput{
		Capacity:         aws.Int64(int64(d.Get("capacity").(int))),
		Name:             aws.String(name),
		Rules:            expandRules(d.Get("rule").(*schema.Set).List()),
		Scope:            aws.String(d.Get("scope").(string)),
		Tags:             getTagsIn(ctx),
		VisibilityConfig: expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
		input.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, ruleGroupCreateTimeout, func() (interface{}, error) {
		return conn.CreateRuleGroupWithContext(ctx, input)
	}, wafv2.ErrCodeWAFUnavailableEntityException)

	if err != nil {
		return diag.Errorf("creating WAFv2 RuleGroup (%s): %s", name, err)
	}

	output := outputRaw.(*wafv2.CreateRuleGroupOutput)

	d.SetId(aws.StringValue(output.Summary.Id))

	return resourceRuleGroupRead(ctx, d, meta)
}

func resourceRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	output, err := FindRuleGroupByThreePartKey(ctx, conn, d.Id(), d.Get("name").(string), d.Get("scope").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 RuleGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading WAFv2 RuleGroup (%s): %s", d.Id(), err)
	}

	ruleGroup := output.RuleGroup
	d.Set("arn", ruleGroup.ARN)
	d.Set("capacity", ruleGroup.Capacity)
	if err := d.Set("custom_response_body", flattenCustomResponseBodies(ruleGroup.CustomResponseBodies)); err != nil {
		return diag.Errorf("setting custom_response_body: %s", err)
	}
	d.Set("description", ruleGroup.Description)
	d.Set("lock_token", output.LockToken)
	d.Set("name", ruleGroup.Name)
	if err := d.Set("rule", flattenRules(ruleGroup.Rules)); err != nil {
		return diag.Errorf("setting rule: %s", err)
	}
	if err := d.Set("visibility_config", flattenVisibilityConfig(ruleGroup.VisibilityConfig)); err != nil {
		return diag.Errorf("setting visibility_config: %s", err)
	}

	return nil
}

func resourceRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &wafv2.UpdateRuleGroupInput{
			Id:               aws.String(d.Id()),
			LockToken:        aws.String(d.Get("lock_token").(string)),
			Name:             aws.String(d.Get("name").(string)),
			Rules:            expandRules(d.Get("rule").(*schema.Set).List()),
			Scope:            aws.String(d.Get("scope").(string)),
			VisibilityConfig: expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
		}

		if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
			input.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		log.Printf("[INFO] Updating WAFv2 RuleGroup: %s", input)
		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, ruleGroupUpdateTimeout, func() (interface{}, error) {
			return conn.UpdateRuleGroupWithContext(ctx, input)
		}, wafv2.ErrCodeWAFUnavailableEntityException)

		if err != nil {
			return diag.Errorf("updating WAFv2 RuleGroup (%s): %s", d.Id(), err)
		}
	}

	return resourceRuleGroupRead(ctx, d, meta)
}

func resourceRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	input := &wafv2.DeleteRuleGroupInput{
		Id:        aws.String(d.Id()),
		LockToken: aws.String(d.Get("lock_token").(string)),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
	}

	log.Printf("[INFO] Deleting WAFv2 RuleGroup: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, ruleGroupDeleteTimeout, func() (interface{}, error) {
		return conn.DeleteRuleGroupWithContext(ctx, input)
	}, wafv2.ErrCodeWAFAssociatedItemException, wafv2.ErrCodeWAFUnavailableEntityException)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting WAFv2 RuleGroup (%s): %s", d.Id(), err)
	}

	return nil
}

func FindRuleGroupByThreePartKey(ctx context.Context, conn *wafv2.WAFV2, id, name, scope string) (*wafv2.GetRuleGroupOutput, error) {
	input := &wafv2.GetRuleGroupInput{
		Id:    aws.String(id),
		Name:  aws.String(name),
		Scope: aws.String(scope),
	}

	output, err := conn.GetRuleGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RuleGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
