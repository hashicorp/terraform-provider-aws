// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_rule", name="Rule")
// @Tags(identifierAttribute="arn")
func resourceRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleCreate,
		ReadWithoutTimeout:   resourceRuleRead,
		UpdateWithoutTimeout: resourceRuleUpdate,
		DeleteWithoutTimeout: resourceRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
				StateFunc:    trimTrailingPeriod,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validResolverName,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resolver_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rule_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.RuleTypeOption](),
			},
			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_ip": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPv4Address,
						},
						"ipv6": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPv6Address,
						},
						names.AttrPort: {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      53,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						names.AttrProtocol: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.ProtocolDo53,
							ValidateDiagFunc: enum.Validate[awstypes.Protocol](),
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			resourceRuleCustomizeDiff,
			verify.SetTagsDiff,
		),
	}
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	input := &route53resolver.CreateResolverRuleInput{
		CreatorRequestId: aws.String(id.PrefixedUniqueId("tf-r53-resolver-rule-")),
		DomainName:       aws.String(d.Get(names.AttrDomainName).(string)),
		RuleType:         awstypes.RuleTypeOption(d.Get("rule_type").(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resolver_endpoint_id"); ok {
		input.ResolverEndpointId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("target_ip"); ok {
		input.TargetIps = expandRuleTargetIPs(v.(*schema.Set))
	}

	output, err := conn.CreateResolverRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Rule: %s", err)
	}

	d.SetId(aws.ToString(output.ResolverRule.Id))

	if _, err := waitRuleCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Rule (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	rule, err := findResolverRuleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Rule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, rule.Arn)
	// To be consistent with other AWS services that do not accept a trailing period,
	// we remove the suffix from the Domain Name returned from the API
	d.Set(names.AttrDomainName, trimTrailingPeriod(aws.ToString(rule.DomainName)))
	d.Set(names.AttrName, rule.Name)
	d.Set(names.AttrOwnerID, rule.OwnerId)
	d.Set("resolver_endpoint_id", rule.ResolverEndpointId)
	d.Set("rule_type", rule.RuleType)
	d.Set("share_status", rule.ShareStatus)
	if err := d.Set("target_ip", flattenRuleTargetIPs(rule.TargetIps)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_ip: %s", err)
	}

	return diags
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	if d.HasChanges(names.AttrName, "resolver_endpoint_id", "target_ip") {
		input := &route53resolver.UpdateResolverRuleInput{
			Config:         &awstypes.ResolverRuleConfig{},
			ResolverRuleId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrName); ok {
			input.Config.Name = aws.String(v.(string))
		}

		if v, ok := d.GetOk("resolver_endpoint_id"); ok {
			input.Config.ResolverEndpointId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("target_ip"); ok {
			input.Config.TargetIps = expandRuleTargetIPs(v.(*schema.Set))
		}

		_, err := conn.UpdateResolverRule(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Rule (%s): %s", d.Id(), err)
		}

		if _, err := waitRuleUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Rule (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Rule: %s", d.Id())
	_, err := conn.DeleteResolverRule(ctx, &route53resolver.DeleteResolverRuleInput{
		ResolverRuleId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitRuleDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Rule (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceRuleCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() != "" {
		if diff.HasChange("resolver_endpoint_id") {
			if _, n := diff.GetChange("resolver_endpoint_id"); n.(string) == "" {
				if err := diff.ForceNew("resolver_endpoint_id"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func findResolverRuleByID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverRule, error) {
	input := &route53resolver.GetResolverRuleInput{
		ResolverRuleId: aws.String(id),
	}

	output, err := conn.GetResolverRule(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResolverRule == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResolverRule, nil
}

func statusRule(ctx context.Context, conn *route53resolver.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findResolverRuleByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitRuleCreated(ctx context.Context, conn *route53resolver.Client, id string, timeout time.Duration) (*awstypes.ResolverRule, error) {
	stateConf := &retry.StateChangeConf{
		Target:  enum.Slice(awstypes.ResolverRuleStatusComplete),
		Refresh: statusRule(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverRule); ok {
		if output.Status == awstypes.ResolverRuleStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitRuleUpdated(ctx context.Context, conn *route53resolver.Client, id string, timeout time.Duration) (*awstypes.ResolverRule, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResolverRuleStatusUpdating),
		Target:  enum.Slice(awstypes.ResolverRuleStatusComplete),
		Refresh: statusRule(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverRule); ok {
		if output.Status == awstypes.ResolverRuleStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitRuleDeleted(ctx context.Context, conn *route53resolver.Client, id string, timeout time.Duration) (*awstypes.ResolverRule, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResolverRuleStatusDeleting),
		Target:  []string{},
		Refresh: statusRule(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverRule); ok {
		if output.Status == awstypes.ResolverRuleStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func expandRuleTargetIPs(vTargetIps *schema.Set) []awstypes.TargetAddress {
	targetAddresses := []awstypes.TargetAddress{}

	for _, vTargetIp := range vTargetIps.List() {
		targetAddress := awstypes.TargetAddress{}

		mTargetIp := vTargetIp.(map[string]interface{})

		if vIp, ok := mTargetIp["ip"].(string); ok && vIp != "" {
			targetAddress.Ip = aws.String(vIp)
		}
		if vIpv6, ok := mTargetIp["ipv6"].(string); ok && vIpv6 != "" {
			targetAddress.Ipv6 = aws.String(vIpv6)
		}
		if vPort, ok := mTargetIp[names.AttrPort].(int); ok {
			targetAddress.Port = aws.Int32(int32(vPort))
		}
		if vProtocol, ok := mTargetIp[names.AttrProtocol].(string); ok && vProtocol != "" {
			targetAddress.Protocol = awstypes.Protocol(vProtocol)
		}

		targetAddresses = append(targetAddresses, targetAddress)
	}

	return targetAddresses
}

func flattenRuleTargetIPs(targetAddresses []awstypes.TargetAddress) []interface{} {
	if targetAddresses == nil {
		return []interface{}{}
	}

	vTargetIps := []interface{}{}

	for _, targetAddress := range targetAddresses {
		mTargetIp := map[string]interface{}{
			"ip":               aws.ToString(targetAddress.Ip),
			"ipv6":             aws.ToString(targetAddress.Ipv6),
			names.AttrPort:     int(aws.ToInt32(targetAddress.Port)),
			names.AttrProtocol: targetAddress.Protocol,
		}

		vTargetIps = append(vTargetIps, mTargetIp)
	}

	return vTargetIps
}

// trimTrailingPeriod is used to remove the trailing period
// of "name" or "domain name" attributes often returned from
// the Route53 API or provided as user input.
// The single dot (".") domain name is returned as-is.
func trimTrailingPeriod(v interface{}) string {
	var str string
	switch value := v.(type) {
	case *string:
		str = aws.ToString(value)
	case string:
		str = value
	default:
		return ""
	}

	if str == "." {
		return str
	}

	return strings.TrimSuffix(str, ".")
}
