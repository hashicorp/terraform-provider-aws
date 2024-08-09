// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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

// @SDKResource("aws_wafv2_rule_group", name="Rule Group")
// @Tags(identifierAttribute="arn")
func resourceRuleGroup() *schema.Resource {
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
				d.Set(names.AttrName, name)
				d.Set(names.AttrScope, scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
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
				names.AttrDescription: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 256),
				},
				"lock_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					ForceNew:      true,
					ConflictsWith: []string{names.AttrNamePrefix},
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric hyphen and underscore characters"),
					),
				},
				names.AttrNamePrefix: {
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					ForceNew:      true,
					ConflictsWith: []string{names.AttrName},
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128-id.UniqueIDSuffixLength),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric hyphen and underscore characters"),
					),
				},
				names.AttrRule: {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrAction: {
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
							names.AttrName: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 128),
							},
							names.AttrPriority: {
								Type:     schema.TypeInt,
								Required: true,
							},
							"rule_label":        ruleLabelsSchema(),
							"statement":         ruleGroupRootStatementSchema(ruleGroupRootStatementSchemaLevel),
							"visibility_config": visibilityConfigSchema(),
						},
					},
				},
				names.AttrScope: {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.Scope](),
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &wafv2.CreateRuleGroupInput{
		Capacity:         aws.Int64(int64(d.Get("capacity").(int))),
		Name:             aws.String(name),
		Rules:            expandRules(d.Get(names.AttrRule).(*schema.Set).List()),
		Scope:            awstypes.Scope(d.Get(names.AttrScope).(string)),
		Tags:             getTagsIn(ctx),
		VisibilityConfig: expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
		input.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	const (
		timeout = 5 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, timeout, func() (interface{}, error) {
		return conn.CreateRuleGroup(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAFv2 RuleGroup (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*wafv2.CreateRuleGroupOutput).Summary.Id))
	d.Set(names.AttrName, name) // Required in Read.

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	output, err := findRuleGroupByThreePartKey(ctx, conn, d.Id(), d.Get(names.AttrName).(string), d.Get(names.AttrScope).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 RuleGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 RuleGroup (%s): %s", d.Id(), err)
	}

	ruleGroup := output.RuleGroup
	d.Set(names.AttrARN, ruleGroup.ARN)
	d.Set("capacity", ruleGroup.Capacity)
	if err := d.Set("custom_response_body", flattenCustomResponseBodies(ruleGroup.CustomResponseBodies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting custom_response_body: %s", err)
	}
	d.Set(names.AttrDescription, ruleGroup.Description)
	d.Set("lock_token", output.LockToken)
	d.Set(names.AttrName, ruleGroup.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(ruleGroup.Name)))
	if err := d.Set(names.AttrRule, flattenRules(ruleGroup.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}
	if err := d.Set("visibility_config", flattenVisibilityConfig(ruleGroup.VisibilityConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting visibility_config: %s", err)
	}

	return diags
}

func resourceRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &wafv2.UpdateRuleGroupInput{
			Id:               aws.String(d.Id()),
			LockToken:        aws.String(d.Get("lock_token").(string)),
			Name:             aws.String(d.Get(names.AttrName).(string)),
			Rules:            expandRules(d.Get(names.AttrRule).(*schema.Set).List()),
			Scope:            awstypes.Scope(d.Get(names.AttrScope).(string)),
			VisibilityConfig: expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
		}

		if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
			input.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		const (
			timeout = 5 * time.Minute
		)
		_, err := tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, timeout, func() (interface{}, error) {
			return conn.UpdateRuleGroup(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAFv2 RuleGroup (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRuleGroupRead(ctx, d, meta)...)
}

func resourceRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	input := &wafv2.DeleteRuleGroupInput{
		Id:        aws.String(d.Id()),
		LockToken: aws.String(d.Get("lock_token").(string)),
		Name:      aws.String(d.Get(names.AttrName).(string)),
		Scope:     awstypes.Scope(d.Get(names.AttrScope).(string)),
	}

	log.Printf("[INFO] Deleting WAFv2 RuleGroup: %s", d.Id())
	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenIsOneOf2[*awstypes.WAFAssociatedItemException, *awstypes.WAFUnavailableEntityException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteRuleGroup(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAFv2 RuleGroup (%s): %s", d.Id(), err)
	}

	return diags
}

func findRuleGroupByThreePartKey(ctx context.Context, conn *wafv2.Client, id, name, scope string) (*wafv2.GetRuleGroupOutput, error) {
	input := &wafv2.GetRuleGroupInput{
		Id:    aws.String(id),
		Name:  aws.String(name),
		Scope: awstypes.Scope(scope),
	}

	output, err := conn.GetRuleGroup(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
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
