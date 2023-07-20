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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	webACLCreateTimeout = 5 * time.Minute
	webACLUpdateTimeout = 5 * time.Minute
	webACLDeleteTimeout = 5 * time.Minute
)

// @SDKResource("aws_wafv2_web_acl", name="Web ACL")
// @Tags(identifierAttribute="arn")
func ResourceWebACL() *schema.Resource {
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
				"association_config": associationConfigSchema(),
				"capacity": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"captcha_config":       outerCaptchaConfigSchema(),
				"custom_response_body": customResponseBodySchema(),
				"default_action": {
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
							"name": {
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
							"priority": {
								Type:     schema.TypeInt,
								Required: true,
							},
							"rule_label":        ruleLabelsSchema(),
							"statement":         webACLRootStatementSchema(webACLRootStatementSchemaLevel),
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
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"token_domains": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 253),
							validation.StringMatch(regexp.MustCompile(`^[\w\.\-/]+$`), "must contain only alphanumeric, hyphen, dot, underscore and forward-slash characters"),
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
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	name := d.Get("name").(string)
	input := &wafv2.CreateWebACLInput{
		AssociationConfig: expandAssociationConfig(d.Get("association_config").([]interface{})),
		CaptchaConfig:     expandCaptchaConfig(d.Get("captcha_config").([]interface{})),
		DefaultAction:     expandDefaultAction(d.Get("default_action").([]interface{})),
		Name:              aws.String(name),
		Rules:             expandWebACLRules(d.Get("rule").(*schema.Set).List()),
		Scope:             aws.String(d.Get("scope").(string)),
		Tags:              getTagsIn(ctx),
		VisibilityConfig:  expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
		input.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("token_domains"); ok && v.(*schema.Set).Len() > 0 {
		input.TokenDomains = flex.ExpandStringSet(v.(*schema.Set))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, webACLCreateTimeout, func() (interface{}, error) {
		return conn.CreateWebACLWithContext(ctx, input)
	}, wafv2.ErrCodeWAFUnavailableEntityException)

	if err != nil {
		return diag.Errorf("creating WAFv2 WebACL (%s): %s", name, err)
	}

	output := outputRaw.(*wafv2.CreateWebACLOutput)

	d.SetId(aws.StringValue(output.Summary.Id))

	return resourceWebACLRead(ctx, d, meta)
}

func resourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	output, err := FindWebACLByThreePartKey(ctx, conn, d.Id(), d.Get("name").(string), d.Get("scope").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 WebACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading WAFv2 WebACL (%s): %s", d.Id(), err)
	}

	webACL := output.WebACL
	arn := aws.StringValue(webACL.ARN)
	d.Set("arn", arn)
	d.Set("capacity", webACL.Capacity)
	if err := d.Set("association_config", flattenAssociationConfig(webACL.AssociationConfig)); err != nil {
		return diag.Errorf("setting association_config: %s", err)
	}
	if err := d.Set("captcha_config", flattenCaptchaConfig(webACL.CaptchaConfig)); err != nil {
		return diag.Errorf("setting captcha_config: %s", err)
	}
	if err := d.Set("custom_response_body", flattenCustomResponseBodies(webACL.CustomResponseBodies)); err != nil {
		return diag.Errorf("setting custom_response_body: %s", err)
	}
	if err := d.Set("default_action", flattenDefaultAction(webACL.DefaultAction)); err != nil {
		return diag.Errorf("setting default_action: %s", err)
	}
	d.Set("description", webACL.Description)
	d.Set("lock_token", output.LockToken)
	d.Set("name", webACL.Name)
	rules := filterWebACLRules(webACL.Rules, expandWebACLRules(d.Get("rule").(*schema.Set).List()))
	if err := d.Set("rule", flattenWebACLRules(rules)); err != nil {
		return diag.Errorf("setting rule: %s", err)
	}
	d.Set("token_domains", aws.StringValueSlice(webACL.TokenDomains))
	if err := d.Set("visibility_config", flattenVisibilityConfig(webACL.VisibilityConfig)); err != nil {
		return diag.Errorf("setting visibility_config: %s", err)
	}

	return nil
}

func resourceWebACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &wafv2.UpdateWebACLInput{
			AssociationConfig: expandAssociationConfig(d.Get("association_config").([]interface{})),
			CaptchaConfig:     expandCaptchaConfig(d.Get("captcha_config").([]interface{})),
			DefaultAction:     expandDefaultAction(d.Get("default_action").([]interface{})),
			Id:                aws.String(d.Id()),
			LockToken:         aws.String(d.Get("lock_token").(string)),
			Name:              aws.String(d.Get("name").(string)),
			Rules:             expandWebACLRules(d.Get("rule").(*schema.Set).List()),
			Scope:             aws.String(d.Get("scope").(string)),
			VisibilityConfig:  expandVisibilityConfig(d.Get("visibility_config").([]interface{})),
		}

		if v, ok := d.GetOk("custom_response_body"); ok && v.(*schema.Set).Len() > 0 {
			input.CustomResponseBodies = expandCustomResponseBodies(v.(*schema.Set).List())
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("token_domains"); ok {
			input.TokenDomains = flex.ExpandStringSet(v.(*schema.Set))
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, webACLUpdateTimeout, func() (interface{}, error) {
			return conn.UpdateWebACLWithContext(ctx, input)
		}, wafv2.ErrCodeWAFUnavailableEntityException)

		if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFOptimisticLockException) {
			return diag.Errorf("updating WAFv2 WebACL (%s), resource has changed since last refresh please run a new plan before applying again: %s", d.Id(), err)
		}

		if err != nil {
			return diag.Errorf("updating WAFv2 WebACL (%s): %s", d.Id(), err)
		}
	}

	return resourceWebACLRead(ctx, d, meta)
}

func resourceWebACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	input := &wafv2.DeleteWebACLInput{
		Id:        aws.String(d.Id()),
		LockToken: aws.String(d.Get("lock_token").(string)),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
	}

	log.Printf("[INFO] Deleting WAFv2 WebACL: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, webACLDeleteTimeout, func() (interface{}, error) {
		return conn.DeleteWebACLWithContext(ctx, input)
	}, wafv2.ErrCodeWAFAssociatedItemException, wafv2.ErrCodeWAFUnavailableEntityException)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting WAFv2 WebACL (%s): %s", d.Id(), err)
	}

	return nil
}

func FindWebACLByThreePartKey(ctx context.Context, conn *wafv2.WAFV2, id, name, scope string) (*wafv2.GetWebACLOutput, error) {
	input := &wafv2.GetWebACLInput{
		Id:    aws.String(id),
		Name:  aws.String(name),
		Scope: aws.String(scope),
	}

	output, err := conn.GetWebACLWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
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
func filterWebACLRules(rules, configRules []*wafv2.Rule) []*wafv2.Rule {
	var fr []*wafv2.Rule
	pattern := `^ShieldMitigationRuleGroup_\d{12}_[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}_.*`
	for _, r := range rules {
		if regexp.MustCompile(pattern).MatchString(aws.StringValue(r.Name)) {
			filter := true
			for _, cr := range configRules {
				if aws.StringValue(cr.Name) == aws.StringValue(r.Name) {
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
