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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafv2_web_acl", name="Web ACL")
// @Tags(identifierAttribute="arn")
func resourceWebACL() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebACLCreate,
		ReadWithoutTimeout:   resourceWebACLRead,
		UpdateWithoutTimeout: resourceWebACLUpdate,
		DeleteWithoutTimeout: resourceWebACLDelete,

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
				"application_integration_url": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"association_config": associationConfigSchema(),
				"capacity": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"captcha_config":       outerCaptchaConfigSchema(),
				"challenge_config":     outerChallengeConfigSchema(),
				"custom_response_body": customResponseBodySchema(),
				names.AttrDefaultAction: {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"allow": allowConfigSchema(),
							"block": blockConfigSchema(),
						},
					},
				},
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
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
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
								Optional: true,
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
							"override_action": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"count": emptySchema(),
										"none":  emptySchema(),
									},
								},
							},
							names.AttrPriority: {
								Type:     schema.TypeInt,
								Required: true,
							},
							"rule_label":        ruleLabelsSchema(),
							"statement":         webACLRootStatementSchema(webACLRootStatementSchemaLevel),
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
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"token_domains": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 253),
							validation.StringMatch(regexache.MustCompile(`^[\w\.\-/]+$`), "must contain only alphanumeric, hyphen, dot, underscore and forward-slash characters"),
						),
					},
				},
				"visibility_config": visibilityConfigSchema(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWebACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &wafv2.CreateWebACLInput{
		AssociationConfig: expandAssociationConfig(d.Get("association_config").([]interface{})),
		CaptchaConfig:     expandCaptchaConfig(d.Get("captcha_config").([]interface{})),
		ChallengeConfig:   expandChallengeConfig(d.Get("challenge_config").([]interface{})),
		DefaultAction:     expandDefaultAction(d.Get(names.AttrDefaultAction).([]interface{})),
		Name:              aws.String(name),
		Rules:             expandWebACLRules(d.Get(names.AttrRule).(*schema.Set).List()),
		Scope:             awstypes.Scope(d.Get(names.AttrScope).(string)),
		Tags:              getTagsIn(ctx),
		VisibilityConfig:  expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
		input.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("token_domains"); ok && v.(*schema.Set).Len() > 0 {
		input.TokenDomains = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	const (
		timeout = 5 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, timeout, func() (interface{}, error) {
		return conn.CreateWebACL(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAFv2 WebACL (%s): %s", name, err)
	}

	output := outputRaw.(*wafv2.CreateWebACLOutput)

	d.SetId(aws.ToString(output.Summary.Id))

	return append(diags, resourceWebACLRead(ctx, d, meta)...)
}

func resourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	output, err := findWebACLByThreePartKey(ctx, conn, d.Id(), d.Get(names.AttrName).(string), d.Get(names.AttrScope).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 WebACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL (%s): %s", d.Id(), err)
	}

	webACL := output.WebACL
	d.Set("application_integration_url", output.ApplicationIntegrationURL)
	d.Set(names.AttrARN, webACL.ARN)
	d.Set("capacity", webACL.Capacity)
	if err := d.Set("association_config", flattenAssociationConfig(webACL.AssociationConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting association_config: %s", err)
	}
	if err := d.Set("captcha_config", flattenCaptchaConfig(webACL.CaptchaConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting captcha_config: %s", err)
	}
	if err := d.Set("challenge_config", flattenChallengeConfig(webACL.ChallengeConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting challenge_config: %s", err)
	}
	if err := d.Set("custom_response_body", flattenCustomResponseBodies(webACL.CustomResponseBodies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting custom_response_body: %s", err)
	}
	if err := d.Set(names.AttrDefaultAction, flattenDefaultAction(webACL.DefaultAction)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_action: %s", err)
	}
	d.Set(names.AttrDescription, webACL.Description)
	d.Set("lock_token", output.LockToken)
	d.Set(names.AttrName, webACL.Name)
	rules := filterWebACLRules(webACL.Rules, expandWebACLRules(d.Get(names.AttrRule).(*schema.Set).List()))
	if err := d.Set(names.AttrRule, flattenWebACLRules(rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}
	d.Set("token_domains", aws.StringSlice(webACL.TokenDomains))
	if err := d.Set("visibility_config", flattenVisibilityConfig(webACL.VisibilityConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting visibility_config: %s", err)
	}

	return diags
}

func resourceWebACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		aclName := d.Get(names.AttrName).(string)
		aclScope := d.Get(names.AttrScope).(string)
		aclLockToken := d.Get("lock_token").(string)
		// Find the AWS managed ShieldMitigationRuleGroup group rule if existent and add it into the set of rules to update
		// so that the provider will not remove the Shield rule when changes are applied to the WebACL.
		rules := expandWebACLRules(d.Get(names.AttrRule).(*schema.Set).List())
		if sr := findShieldRule(rules); len(sr) == 0 {
			output, err := findWebACLByThreePartKey(ctx, conn, d.Id(), aclName, aclScope)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL (%s): %s", d.Id(), err)
			}

			rules = append(rules, findShieldRule(output.WebACL.Rules)...)
		}

		input := &wafv2.UpdateWebACLInput{
			AssociationConfig: expandAssociationConfig(d.Get("association_config").([]interface{})),
			CaptchaConfig:     expandCaptchaConfig(d.Get("captcha_config").([]interface{})),
			ChallengeConfig:   expandChallengeConfig(d.Get("challenge_config").([]interface{})),
			DefaultAction:     expandDefaultAction(d.Get(names.AttrDefaultAction).([]interface{})),
			Id:                aws.String(d.Id()),
			LockToken:         aws.String(aclLockToken),
			Name:              aws.String(aclName),
			Rules:             rules,
			Scope:             awstypes.Scope(aclScope),
			VisibilityConfig:  expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
		}

		if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
			input.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("token_domains"); ok {
			input.TokenDomains = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		const (
			timeout = 5 * time.Minute
		)
		_, err := tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, timeout, func() (interface{}, error) {
			return conn.UpdateWebACL(ctx, input)
		})

		if errs.IsA[*awstypes.WAFOptimisticLockException](err) {
			var output *wafv2.GetWebACLOutput
			output, err = findWebACLByThreePartKey(ctx, conn, d.Id(), aclName, aclScope)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL (%s): %s", d.Id(), err)
			}

			if newLockToken := aws.ToString(output.LockToken); newLockToken != aclLockToken {
				// Retrieved a new lock token, retry due to other processes modifying the web acl out of band (See: https://docs.aws.amazon.com/sdk-for-go/api/service/shield/#Shield.EnableApplicationLayerAutomaticResponse)
				input.LockToken = aws.String(newLockToken)
				_, err = tfresource.RetryWhenIsOneOf2[*awstypes.WAFAssociatedItemException, *awstypes.WAFUnavailableEntityException](ctx, timeout, func() (interface{}, error) {
					return conn.UpdateWebACL(ctx, input)
				})

				if errs.IsA[*awstypes.WAFOptimisticLockException](err) {
					return sdkdiag.AppendErrorf(diags, "updating WAFv2 WebACL (%s), resource has changed since last refresh please run a new plan before applying again: %s", d.Id(), err)
				}
			}
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAFv2 WebACL (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWebACLRead(ctx, d, meta)...)
}

func resourceWebACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	aclName := d.Get(names.AttrName).(string)
	aclScope := d.Get(names.AttrScope).(string)
	aclLockToken := d.Get("lock_token").(string)
	input := &wafv2.DeleteWebACLInput{
		Id:        aws.String(d.Id()),
		LockToken: aws.String(aclLockToken),
		Name:      aws.String(aclName),
		Scope:     awstypes.Scope(aclScope),
	}

	log.Printf("[INFO] Deleting WAFv2 WebACL: %s", d.Id())
	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenIsOneOf2[*awstypes.WAFAssociatedItemException, *awstypes.WAFUnavailableEntityException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteWebACL(ctx, input)
	})

	if errs.IsA[*awstypes.WAFOptimisticLockException](err) {
		var output *wafv2.GetWebACLOutput
		output, err = findWebACLByThreePartKey(ctx, conn, d.Id(), aclName, aclScope)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL (%s): %s", d.Id(), err)
		}

		if newLockToken := aws.ToString(output.LockToken); newLockToken != aclLockToken {
			// Retrieved a new lock token, retry due to other processes modifying the web acl out of band (See: https://docs.aws.amazon.com/sdk-for-go/api/service/shield/#Shield.EnableApplicationLayerAutomaticResponse)
			input.LockToken = aws.String(newLockToken)
			_, err = tfresource.RetryWhenIsOneOf2[*awstypes.WAFAssociatedItemException, *awstypes.WAFUnavailableEntityException](ctx, timeout, func() (interface{}, error) {
				return conn.DeleteWebACL(ctx, input)
			})

			if errs.IsA[*awstypes.WAFOptimisticLockException](err) {
				return sdkdiag.AppendErrorf(diags, "deleting WAFv2 WebACL (%s), resource has changed since last refresh please run a new plan before applying again: %s", d.Id(), err)
			}
		}
	}

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAFv2 WebACL (%s): %s", d.Id(), err)
	}

	return diags
}

func findWebACLByThreePartKey(ctx context.Context, conn *wafv2.Client, id, name, scope string) (*wafv2.GetWebACLOutput, error) {
	input := &wafv2.GetWebACLInput{
		Id:    aws.String(id),
		Name:  aws.String(name),
		Scope: awstypes.Scope(scope),
	}

	output, err := conn.GetWebACL(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.WebACL == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// filterWebACLRules removes the AWS-added Shield Advanced auto mitigation rule here
// so that the provider will not report diff and/or attempt to remove the rule as it is
// owned and managed by AWS.
// See https://github.com/hashicorp/terraform-provider-aws/issues/22869
// See https://docs.aws.amazon.com/waf/latest/developerguide/ddos-automatic-app-layer-response-rg.html
func filterWebACLRules(rules, configRules []awstypes.Rule) []awstypes.Rule {
	var fr []awstypes.Rule
	sr := findShieldRule(rules)

	if len(sr) == 0 {
		return rules
	}

	for _, r := range rules {
		if aws.ToString(r.Name) == aws.ToString(sr[0].Name) {
			filter := true
			for _, cr := range configRules {
				if aws.ToString(cr.Name) == aws.ToString(r.Name) {
					// exception to filtering -- it's in the config
					filter = false
				}
			}

			if filter {
				continue
			}
		}
		fr = append(fr, r)
	}
	return fr
}

func findShieldRule(rules []awstypes.Rule) []awstypes.Rule {
	pattern := `^ShieldMitigationRuleGroup_\d{12}_[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}_.*`
	var sr []awstypes.Rule
	for _, r := range rules {
		if regexache.MustCompile(pattern).MatchString(aws.ToString(r.Name)) {
			sr = append(sr, r)
		}
	}
	return sr
}
