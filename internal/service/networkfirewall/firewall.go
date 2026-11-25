// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package networkfirewall

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkfirewall_firewall", name="Firewall")
// @Tags(identifierAttribute="id")
func resourceFirewall() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallCreate,
		ReadWithoutTimeout:   resourceFirewallRead,
		UpdateWithoutTimeout: resourceFirewallUpdate,
		DeleteWithoutTimeout: resourceFirewallDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("firewall_status", func(ctx context.Context, diff *schema.ResourceDiff, meta any) bool {
				return diff.HasChanges("subnet_mapping", "nat_gateway_create_mapping")
			}),
			// An existing firewall can't be converted to or from a proxy firewall.
			// Skipped on create: ForceNew re-writes the set into the diff, which
			// adds an empty element when the set has values unknown at plan time.
			customdiff.ForceNewIf("nat_gateway_create_mapping", func(ctx context.Context, d *schema.ResourceDiff, meta any) bool {
				if d.Id() == "" {
					return false
				}
				o, n := d.GetChange("nat_gateway_create_mapping")
				return (o.(*schema.Set).Len() == 0) != (n.(*schema.Set).Len() == 0)
			}),
			validateProxyFirewallDiff,
		),

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"availability_zone_change_protection": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"availability_zone_mapping": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_zone_id": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				// Not sent on create or read back. Used when the firewall is deleted and
				// when an update releases NAT gateways from the firewall.
				"delete_associated_nat_gateways": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"delete_protection": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				"enabled_analysis_types": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: enum.Validate[awstypes.EnabledAnalysisType](),
					},
				},
				names.AttrEncryptionConfiguration: encryptionConfigurationSchema(),
				"firewall_policy_arn": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"firewall_policy_change_protection": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"firewall_status": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"sync_states": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"attachment": {
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrDNSName: {
														Type:     schema.TypeString,
														Computed: true,
													},
													"endpoint_id": {
														Type:     schema.TypeString,
														Computed: true,
													},
													names.AttrStatus: {
														Type:     schema.TypeString,
														Computed: true,
													},
													names.AttrStatusMessage: {
														Type:     schema.TypeString,
														Computed: true,
													},
													names.AttrSubnetID: {
														Type:     schema.TypeString,
														Computed: true,
													},
												},
											},
										},
										names.AttrAvailabilityZone: {
											Type:     schema.TypeString,
											Computed: true,
										},
										"nat_gateway_attachment": {
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"nat_gateway_id": {
														Type:     schema.TypeString,
														Computed: true,
													},
													names.AttrStatus: {
														Type:     schema.TypeString,
														Computed: true,
													},
													names.AttrStatusMessage: {
														Type:     schema.TypeString,
														Computed: true,
													},
												},
											},
										},
									},
								},
							},
							"transit_gateway_attachment_sync_states": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"attachment_id": {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"nat_gateway_create_mapping": {
					Type:          schema.TypeSet,
					Optional:      true,
					ConflictsWith: []string{names.AttrTransitGatewayID},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							// AllocationIds and PrivateIps are paired by position.
							"allocation_ids": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 8,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"private_ips": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 8,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.IsIPv4Address,
								},
							},
							names.AttrSubnetID: {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"proxy_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrCertificateARN: {
								Type:         schema.TypeString,
								Optional:     true,
								ForceNew:     true,
								ValidateFunc: verify.ValidARN,
								RequiredWith: []string{"proxy_settings.0.domain_name"},
							},
							names.AttrDomainName: {
								Type:         schema.TypeString,
								Optional:     true,
								ForceNew:     true,
								ValidateFunc: validation.StringLenBetween(1, 253),
								RequiredWith: []string{"proxy_settings.0.certificate_arn"},
							},
							"listener": {
								Type:     schema.TypeSet,
								Required: true,
								MinItems: 1,
								MaxItems: 2,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IsPortNumber,
										},
										names.AttrType: {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.ListenerPropertyType](),
										},
									},
								},
							},
						},
					},
				},
				"subnet_change_protection": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"subnet_mapping": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrIPAddressType: {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[awstypes.IPAddressType](),
							},
							names.AttrSubnetID: {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaOptionalComputedWarnShouldBeComputedOnly(),
				names.AttrTransitGatewayID: {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ExactlyOneOf: []string{names.AttrTransitGatewayID, names.AttrVPCID},
				},
				"transit_gateway_owner_account_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"transparent_filtering_nat_scope": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.TransparentFilteringNatScope](),
				},
				"update_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrVPCID: {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ExactlyOneOf: []string{names.AttrTransitGatewayID, names.AttrVPCID},
				},
			}
		},
	}
}

func resourceFirewallCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := networkfirewall.CreateFirewallInput{
		FirewallName:      aws.String(name),
		FirewallPolicyArn: aws.String(d.Get("firewall_policy_arn").(string)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("availability_zone_change_protection"); ok {
		input.AvailabilityZoneChangeProtection = v.(bool)
	}

	if v := d.Get("availability_zone_mapping").(*schema.Set); v.Len() > 0 {
		input.AvailabilityZoneMappings = expandAvailabilityZoneMapping(v.List())
	}

	if v, ok := d.GetOk("delete_protection"); ok {
		input.DeleteProtection = v.(bool)
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v := d.Get("enabled_analysis_types").(*schema.Set); v.Len() > 0 {
		input.EnabledAnalysisTypes = flex.ExpandStringyValueSet[awstypes.EnabledAnalysisType](v)
	}

	if v, ok := d.GetOk(names.AttrEncryptionConfiguration); ok {
		input.EncryptionConfiguration = expandEncryptionConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("firewall_policy_change_protection"); ok {
		input.FirewallPolicyChangeProtection = v.(bool)
	}

	if v, ok := d.GetOk("subnet_change_protection"); ok {
		input.SubnetChangeProtection = v.(bool)
	}

	if v := d.Get("subnet_mapping").(*schema.Set); v.Len() > 0 {
		input.SubnetMappings = expandSubnetMappings(v.List())
	}

	if v, ok := d.GetOk(names.AttrTransitGatewayID); ok {
		input.TransitGatewayId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVPCID); ok {
		input.VpcId = aws.String(v.(string))
	}

	// A proxy firewall takes its VPC and endpoint subnets through VpcEndpoint
	// and rejects the top-level VpcId and SubnetMappings.
	if v := d.Get("nat_gateway_create_mapping").(*schema.Set); v.Len() > 0 {
		input.NatGatewayCreateMappings = expandNatGatewayCreateMappings(v.List())
		input.VpcEndpoint = &awstypes.VpcEndpoint{
			SubnetMappings: input.SubnetMappings,
			VpcId:          input.VpcId,
		}
		input.SubnetMappings = nil
		input.VpcId = nil
	}

	if v, ok := d.GetOk("proxy_settings"); ok {
		input.ProxySettings = expandProxySettings(v.([]any))
	}

	if v, ok := d.GetOk("transparent_filtering_nat_scope"); ok {
		input.TransparentFilteringNatScope = awstypes.TransparentFilteringNatScope(v.(string))
	}

	output, err := conn.CreateFirewall(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating NetworkFirewall Firewall (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Firewall.FirewallArn))

	if output.Firewall.TransitGatewayId != nil {
		if _, err := waitFirewallTransitGatewayAttachmentCreated(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Firewall Transit Gateway Attachment (%s) create: %s", d.Id(), err)
		}
	} else {
		if _, err := waitFirewallCreated(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Firewall (%s) create: %s", d.Id(), err)
		}
	}
	return append(diags, resourceFirewallRead(ctx, d, meta)...)
}

func resourceFirewallRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	output, err := findFirewallByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] NetworkFirewall Firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall (%s): %s", d.Id(), err)
	}

	firewall := output.Firewall
	d.Set(names.AttrARN, firewall.FirewallArn)
	d.Set("availability_zone_change_protection", firewall.AvailabilityZoneChangeProtection)
	d.Set("availability_zone_mapping", flattenAvailabilityZoneMapping(firewall.AvailabilityZoneMappings))
	// Not returned by the API. Default it for imported state and for state
	// written before the argument existed, so that neither shows a diff.
	if !d.IsNewResource() {
		if v := d.GetRawState(); v.IsNull() || !v.Type().HasAttribute("delete_associated_nat_gateways") || v.GetAttr("delete_associated_nat_gateways").IsNull() {
			d.Set("delete_associated_nat_gateways", true)
		}
	}
	d.Set("delete_protection", firewall.DeleteProtection)
	d.Set(names.AttrDescription, firewall.Description)
	d.Set("enabled_analysis_types", firewall.EnabledAnalysisTypes)
	if err := d.Set(names.AttrEncryptionConfiguration, flattenEncryptionConfiguration(firewall.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	d.Set("firewall_policy_arn", firewall.FirewallPolicyArn)
	d.Set("firewall_policy_change_protection", firewall.FirewallPolicyChangeProtection)
	if err := d.Set("firewall_status", flattenFirewallStatus(output.FirewallStatus)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting firewall_status: %s", err)
	}
	d.Set(names.AttrName, firewall.FirewallName)
	// DescribeFirewall doesn't return NatGatewayCreateMappings, only the IDs of
	// the NAT gateways the firewall created. Rebuild the mappings from those NAT
	// gateways for import and drift detection. Until any exist (while the
	// firewall is provisioning), keep the configured value.
	priorNATGatewayCreateMappings := expandNatGatewayCreateMappings(d.Get("nat_gateway_create_mapping").(*schema.Set).List())
	natGatewayCreateMappings := firewall.NatGatewayCreateMappings
	if len(natGatewayCreateMappings) == 0 {
		natGatewayCreateMappings, err = findFirewallNATGatewayCreateMappings(ctx, meta.(*conns.AWSClient), output.FirewallStatus, priorNATGatewayCreateMappings)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall (%s) NAT gateways: %s", d.Id(), err)
		}
	}
	if len(natGatewayCreateMappings) == 0 {
		natGatewayCreateMappings = priorNATGatewayCreateMappings
	}
	if err := d.Set("nat_gateway_create_mapping", flattenNatGatewayCreateMappings(natGatewayCreateMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting nat_gateway_create_mapping: %s", err)
	}
	if err := d.Set("proxy_settings", flattenProxySettings(firewall.ProxySettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting proxy_settings: %s", err)
	}
	d.Set("subnet_change_protection", firewall.SubnetChangeProtection)
	subnetMappings, vpcID := firewall.SubnetMappings, firewall.VpcId
	if firewall.VpcEndpoint != nil && len(subnetMappings) == 0 {
		subnetMappings, vpcID = firewall.VpcEndpoint.SubnetMappings, firewall.VpcEndpoint.VpcId
	}
	if err := d.Set("subnet_mapping", flattenSubnetMappings(subnetMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_mapping: %s", err)
	}
	d.Set(names.AttrTransitGatewayID, firewall.TransitGatewayId)
	d.Set("transit_gateway_owner_account_id", firewall.TransitGatewayOwnerAccountId)
	d.Set("transparent_filtering_nat_scope", firewall.TransparentFilteringNatScope)
	d.Set("update_token", output.UpdateToken)
	d.Set(names.AttrVPCID, vpcID)

	setTagsOut(ctx, firewall.Tags)

	return diags
}

func resourceFirewallUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	updateToken := d.Get("update_token").(string)

	if d.HasChange("delete_protection") {
		input := networkfirewall.UpdateFirewallDeleteProtectionInput{
			DeleteProtection: d.Get("delete_protection").(bool),
			FirewallArn:      aws.String(d.Id()),
			UpdateToken:      aws.String(updateToken),
		}

		output, err := conn.UpdateFirewallDeleteProtection(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) delete protection: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange(names.AttrDescription) {
		input := networkfirewall.UpdateFirewallDescriptionInput{
			Description: aws.String(d.Get(names.AttrDescription).(string)),
			FirewallArn: aws.String(d.Id()),
			UpdateToken: aws.String(updateToken),
		}

		output, err := conn.UpdateFirewallDescription(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) description: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange("enabled_analysis_types") {
		input := networkfirewall.UpdateFirewallAnalysisSettingsInput{
			EnabledAnalysisTypes: flex.ExpandStringyValueSet[awstypes.EnabledAnalysisType](d.Get("enabled_analysis_types").(*schema.Set)),
			FirewallArn:          aws.String(d.Id()),
			UpdateToken:          aws.String(updateToken),
		}

		output, err := conn.UpdateFirewallAnalysisSettings(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) enabled analysis types: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange(names.AttrEncryptionConfiguration) {
		input := networkfirewall.UpdateFirewallEncryptionConfigurationInput{
			EncryptionConfiguration: expandEncryptionConfiguration(d.Get(names.AttrEncryptionConfiguration).([]any)),
			FirewallArn:             aws.String(d.Id()),
			UpdateToken:             aws.String(updateToken),
		}

		output, err := conn.UpdateFirewallEncryptionConfiguration(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) encryption configuration: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	// Note: The *_change_protection fields below are handled before their respective fields
	// to account for disabling and subsequent changes.

	if d.HasChange("availability_zone_change_protection") {
		input := networkfirewall.UpdateAvailabilityZoneChangeProtectionInput{
			AvailabilityZoneChangeProtection: d.Get("availability_zone_change_protection").(bool),
			FirewallArn:                      aws.String(d.Id()),
			UpdateToken:                      aws.String(updateToken),
		}
		output, err := conn.UpdateAvailabilityZoneChangeProtection(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) availability zone change protection: %s", d.Id(), err)
		}
		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange("availability_zone_mapping") {
		o, n := d.GetChange("availability_zone_mapping")
		availabilityZoneToRemove, availabilityZoneToAdd := availabilityZoneMappingsDiff(o.(*schema.Set), n.(*schema.Set))

		if len(availabilityZoneToAdd) > 0 {
			input := networkfirewall.AssociateAvailabilityZonesInput{
				FirewallArn:              aws.String(d.Id()),
				AvailabilityZoneMappings: availabilityZoneToAdd,
				UpdateToken:              aws.String(updateToken),
			}

			_, err := conn.AssociateAvailabilityZones(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating NetworkFirewall Firewall (%s) availability zones: %s", d.Id(), err)
			}

			output, err := waitFirewallUpdated(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id())

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Firewall (%s) update: %s", d.Id(), err)
			}

			updateToken = aws.ToString(output.UpdateToken)
		}

		if len(availabilityZoneToRemove) > 0 {
			input := networkfirewall.DisassociateAvailabilityZonesInput{
				FirewallArn:              aws.String(d.Id()),
				AvailabilityZoneMappings: availabilityZoneToRemove,
				UpdateToken:              aws.String(updateToken),
			}

			_, err := conn.DisassociateAvailabilityZones(ctx, &input)

			if err == nil {
				output, err := waitFirewallUpdated(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id())

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Firewall (%s) update: %s", d.Id(), err)
				}

				updateToken = aws.ToString(output.UpdateToken)
			} else if !errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "inaccessible") {
				return sdkdiag.AppendErrorf(diags, "disassociating NetworkFirewall Firewall (%s) availability zones: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("firewall_policy_change_protection") {
		input := networkfirewall.UpdateFirewallPolicyChangeProtectionInput{
			FirewallArn:                    aws.String(d.Id()),
			FirewallPolicyChangeProtection: d.Get("firewall_policy_change_protection").(bool),
			UpdateToken:                    aws.String(updateToken),
		}

		output, err := conn.UpdateFirewallPolicyChangeProtection(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) policy change protection: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange("firewall_policy_arn") {
		input := networkfirewall.AssociateFirewallPolicyInput{
			FirewallArn:       aws.String(d.Id()),
			FirewallPolicyArn: aws.String(d.Get("firewall_policy_arn").(string)),
			UpdateToken:       aws.String(updateToken),
		}

		output, err := conn.AssociateFirewallPolicy(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) firewall policy ARN: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange("subnet_change_protection") {
		input := networkfirewall.UpdateSubnetChangeProtectionInput{
			FirewallArn:            aws.String(d.Id()),
			SubnetChangeProtection: d.Get("subnet_change_protection").(bool),
			UpdateToken:            aws.String(updateToken),
		}

		output, err := conn.UpdateSubnetChangeProtection(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) subnet change protection: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange("proxy_settings.0.listener") {
		// ProxySettings is replaced as a whole, so the unchanged certificate and
		// domain name must be sent again.
		input := networkfirewall.UpdateProxySettingsInput{
			FirewallArn:   aws.String(d.Id()),
			ProxySettings: expandProxySettings(d.Get("proxy_settings").([]any)),
			UpdateToken:   aws.String(updateToken),
		}

		output, err := conn.UpdateProxySettings(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) proxy settings: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChanges("subnet_mapping", "nat_gateway_create_mapping") {
		var err error
		if d.Get("nat_gateway_create_mapping").(*schema.Set).Len() > 0 {
			err = updateProxyFirewallAvailabilityZones(ctx, d, meta, updateToken)
		} else {
			err = updateFirewallSubnets(ctx, conn, d, updateToken)
		}
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceFirewallRead(ctx, d, meta)...)
}

func updateFirewallSubnets(ctx context.Context, conn *networkfirewall.Client, d *schema.ResourceData, updateToken string) error {
	o, n := d.GetChange("subnet_mapping")
	subnetsToRemove, subnetsToAdd := subnetMappingsDiff(o.(*schema.Set), n.(*schema.Set))

	updateToken, err := associateFirewallSubnets(ctx, conn, d, updateToken, subnetsToAdd, nil)
	if err != nil {
		return err
	}

	return disassociateFirewallSubnets(ctx, conn, d, updateToken, subnetsToRemove)
}

// updateProxyFirewallAvailabilityZones applies endpoint subnet and NAT gateway
// mapping changes to a proxy firewall. A proxy firewall has one endpoint
// subnet and one NAT gateway per AZ, and rejects associating an AZ it's already
// in, so a changed AZ is removed and then added again:
//  1. associate AZs the firewall isn't in yet
//  2. disassociate removed and changed AZs, and wait until they're gone
//  3. if requested, delete the NAT gateways that step 2 released, which the
//     service keeps; this also frees their Elastic IPs for step 4
//  4. associate changed AZs again
func updateProxyFirewallAvailabilityZones(ctx context.Context, d *schema.ResourceData, meta any, updateToken string) error {
	client := meta.(*conns.AWSClient)
	conn := client.NetworkFirewallClient(ctx)

	o, n := d.GetChange("subnet_mapping")
	oldSubnets, newSubnets := o.(*schema.Set).List(), n.(*schema.Set).List()
	o, n = d.GetChange("nat_gateway_create_mapping")
	oldNATGateways, newNATGateways := o.(*schema.Set).List(), n.(*schema.Set).List()

	changes, err := planProxyFirewallAvailabilityZones(ctx, client, oldSubnets, oldNATGateways, newSubnets, newNATGateways)
	if err != nil {
		return err
	}

	if updateToken, err = associateProxyAvailabilityZones(ctx, conn, d, updateToken, changes.added); err != nil {
		return err
	}

	if len(changes.removed) > 0 {
		firewall, err := findFirewallByARN(ctx, conn, d.Id())
		if err != nil {
			return fmt.Errorf("reading NetworkFirewall Firewall (%s): %w", d.Id(), err)
		}

		var azNames, subnetIDs, releasedNATGatewayIDs []string
		for _, az := range changes.removed {
			azNames = append(azNames, az.name)
			subnetIDs = append(subnetIDs, aws.ToString(az.subnet.SubnetId))
			for _, v := range firewall.FirewallStatus.SyncStates[az.name].NatGatewayAttachments {
				releasedNATGatewayIDs = append(releasedNATGatewayIDs, aws.ToString(v.NatGatewayId))
			}
		}

		if err := disassociateFirewallSubnets(ctx, conn, d, updateToken, subnetIDs); err != nil {
			return err
		}

		output, err := waitFirewallAvailabilityZonesRemoved(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id(), azNames)
		if err != nil {
			return fmt.Errorf("waiting for NetworkFirewall Firewall (%s) Availability Zones removal: %w", d.Id(), err)
		}
		updateToken = aws.ToString(output.UpdateToken)

		if d.Get("delete_associated_nat_gateways").(bool) {
			for _, id := range releasedNATGatewayIDs {
				// Once released, it's an ordinary NAT gateway.
				if err := tfec2.DeleteNATGateway(ctx, client.EC2Client(ctx), id, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return err
				}
			}
		}
	}

	_, err = associateProxyAvailabilityZones(ctx, conn, d, updateToken, changes.readded)

	return err
}

func associateProxyAvailabilityZones(ctx context.Context, conn *networkfirewall.Client, d *schema.ResourceData, updateToken string, azs []proxyAvailabilityZone) (string, error) {
	var subnets []awstypes.SubnetMapping
	var natGateways []awstypes.NatGatewayCreateMapping
	for _, az := range azs {
		subnets = append(subnets, az.subnet)
		natGateways = append(natGateways, az.natGateway)
	}

	return associateFirewallSubnets(ctx, conn, d, updateToken, subnets, natGateways)
}

func associateFirewallSubnets(ctx context.Context, conn *networkfirewall.Client, d *schema.ResourceData, updateToken string, subnets []awstypes.SubnetMapping, natGateways []awstypes.NatGatewayCreateMapping) (string, error) {
	if len(subnets) == 0 && len(natGateways) == 0 {
		return updateToken, nil
	}

	input := networkfirewall.AssociateSubnetsInput{
		FirewallArn:    aws.String(d.Id()),
		SubnetMappings: subnets,
		UpdateToken:    aws.String(updateToken),
	}
	if len(natGateways) > 0 {
		input.NatGatewayCreateMappings = natGateways
	}

	if _, err := conn.AssociateSubnets(ctx, &input); err != nil {
		return "", fmt.Errorf("associating NetworkFirewall Firewall (%s) subnets: %w", d.Id(), err)
	}

	output, err := waitFirewallUpdated(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id())

	if err != nil {
		return "", fmt.Errorf("waiting for NetworkFirewall Firewall (%s) update: %w", d.Id(), err)
	}

	// A failed new AZ doesn't fail the firewall, so check each AZ.
	if err := failedSyncStatesError(output.FirewallStatus); err != nil {
		return "", fmt.Errorf("associating NetworkFirewall Firewall (%s) subnets: %w", d.Id(), err)
	}

	return aws.ToString(output.UpdateToken), nil
}

// disassociateFirewallSubnets disassociates the subnets and waits for the
// update. Callers read the next update token from DescribeFirewall.
func disassociateFirewallSubnets(ctx context.Context, conn *networkfirewall.Client, d *schema.ResourceData, updateToken string, subnetIDs []string) error {
	if len(subnetIDs) == 0 {
		return nil
	}

	input := networkfirewall.DisassociateSubnetsInput{
		FirewallArn: aws.String(d.Id()),
		SubnetIds:   subnetIDs,
		UpdateToken: aws.String(updateToken),
	}

	_, err := conn.DisassociateSubnets(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "inaccessible") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("disassociating NetworkFirewall Firewall (%s) subnets: %w", d.Id(), err)
	}

	if _, err := waitFirewallUpdated(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
		return fmt.Errorf("waiting for NetworkFirewall Firewall (%s) update: %w", d.Id(), err)
	}

	return nil
}

// findFirewallNATGatewayCreateMappings rebuilds a proxy firewall's NAT gateway
// create mappings from the NAT gateways it created. Addresses that are being
// added or removed are skipped. When a NAT gateway has the same Elastic IP and
// private IP pairs as the prior mapping for its subnet, the prior order is kept,
// because EC2 doesn't guarantee the order of secondary addresses.
func findFirewallNATGatewayCreateMappings(ctx context.Context, client *conns.AWSClient, status *awstypes.FirewallStatus, prior []awstypes.NatGatewayCreateMapping) ([]awstypes.NatGatewayCreateMapping, error) {
	if status == nil {
		return nil, nil
	}

	priorBySubnet := make(map[string]awstypes.NatGatewayCreateMapping)
	for _, v := range prior {
		priorBySubnet[aws.ToString(v.SubnetId)] = v
	}

	var mappings []awstypes.NatGatewayCreateMapping
	for _, syncState := range status.SyncStates {
		for _, attachment := range syncState.NatGatewayAttachments {
			if attachment.Status == awstypes.NatGatewayAttachmentStatusDeleting {
				continue
			}

			id := aws.ToString(attachment.NatGatewayId)
			natGateway, err := tfec2.FindNATGatewayByID(ctx, client.EC2Client(ctx), id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return nil, fmt.Errorf("reading EC2 NAT Gateway (%s): %w", id, err)
			}

			// Primary address first; it's the first private IP of a private NAT gateway.
			addresses := slices.Clone(natGateway.NatGatewayAddresses)
			slices.SortStableFunc(addresses, func(a, b ec2types.NatGatewayAddress) int {
				switch aPrimary, bPrimary := aws.ToBool(a.IsPrimary), aws.ToBool(b.IsPrimary); {
				case aPrimary && !bPrimary:
					return -1
				case !aPrimary && bPrimary:
					return 1
				}
				return 0
			})

			mapping := awstypes.NatGatewayCreateMapping{
				SubnetId: natGateway.SubnetId,
			}
			for _, address := range addresses {
				if address.Status != ec2types.NatGatewayAddressStatusSucceeded {
					continue
				}
				if natGateway.ConnectivityType == ec2types.ConnectivityTypePublic {
					mapping.AllocationIds = append(mapping.AllocationIds, aws.ToString(address.AllocationId))
				}
				mapping.PrivateIps = append(mapping.PrivateIps, aws.ToString(address.PrivateIp))
			}

			if v, ok := priorBySubnet[aws.ToString(mapping.SubnetId)]; ok && sameNATGatewayAddresses(v, mapping) {
				mapping = v
			}

			mappings = append(mappings, mapping)
		}
	}

	return mappings, nil
}

// sameNATGatewayAddresses reports whether two mappings have the same pairs of
// Elastic IP and private IP, or the same private IPs, in any order.
func sameNATGatewayAddresses(a, b awstypes.NatGatewayCreateMapping) bool {
	pairs := func(m awstypes.NatGatewayCreateMapping) []string {
		var out []string
		for i, ip := range m.PrivateIps {
			var allocationID string
			if i < len(m.AllocationIds) {
				allocationID = m.AllocationIds[i]
			}
			out = append(out, allocationID+"/"+ip)
		}
		slices.Sort(out)
		return out
	}

	return len(a.AllocationIds) == len(b.AllocationIds) && slices.Equal(pairs(a), pairs(b))
}

// proxyAvailabilityZone is one AZ of a proxy firewall: its endpoint subnet and
// the NAT gateway that Network Firewall creates for it.
type proxyAvailabilityZone struct {
	name       string
	subnet     awstypes.SubnetMapping
	natGateway awstypes.NatGatewayCreateMapping
}

func (az proxyAvailabilityZone) equal(other proxyAvailabilityZone) bool {
	// ip_address_type is computed: during planning it can be unset, or unknown
	// (a placeholder value). Only two actual IP address types can differ.
	isSet := func(v awstypes.IPAddressType) bool {
		return slices.Contains(enum.EnumValues[awstypes.IPAddressType](), v)
	}
	ipAddressTypeEqual := !isSet(az.subnet.IPAddressType) || !isSet(other.subnet.IPAddressType) || az.subnet.IPAddressType == other.subnet.IPAddressType

	return aws.ToString(az.subnet.SubnetId) == aws.ToString(other.subnet.SubnetId) &&
		ipAddressTypeEqual &&
		aws.ToString(az.natGateway.SubnetId) == aws.ToString(other.natGateway.SubnetId) &&
		slices.Equal(az.natGateway.AllocationIds, other.natGateway.AllocationIds) &&
		slices.Equal(az.natGateway.PrivateIps, other.natGateway.PrivateIps)
}

type proxyAvailabilityZoneChanges struct {
	added   []proxyAvailabilityZone // AZs the firewall isn't in yet
	removed []proxyAvailabilityZone // old values of removed and changed AZs
	readded []proxyAvailabilityZone // new values of changed AZs
}

// errProxyFirewallReplacementRequired reports an Availability Zone change that
// can't be applied in place: the firewall must keep at least one AZ while a
// changed AZ is removed and added again.
var errProxyFirewallReplacementRequired = errors.New(`can't change the only Availability Zone of a proxy firewall in place: add another Availability Zone first, or replace the firewall`)

// planProxyFirewallAvailabilityZones groups endpoint subnets and NAT gateway
// mappings by AZ and compares the old and new AZs. It returns
// errProxyFirewallReplacementRequired when the change can't be applied in place.
func planProxyFirewallAvailabilityZones(ctx context.Context, client *conns.AWSClient, oldSubnets, oldNATGateways, newSubnets, newNATGateways []any) (*proxyAvailabilityZoneChanges, error) {
	oldAZs, err := proxyFirewallAvailabilityZones(ctx, client, oldSubnets, oldNATGateways)
	if err != nil {
		return nil, err
	}
	newAZs, err := proxyFirewallAvailabilityZones(ctx, client, newSubnets, newNATGateways)
	if err != nil {
		return nil, err
	}

	var changes proxyAvailabilityZoneChanges
	for name, newAZ := range newAZs {
		oldAZ, ok := oldAZs[name]
		switch {
		case !ok:
			changes.added = append(changes.added, newAZ)
		case !oldAZ.equal(newAZ):
			changes.removed = append(changes.removed, oldAZ)
			changes.readded = append(changes.readded, newAZ)
		}
	}
	for name, oldAZ := range oldAZs {
		if _, ok := newAZs[name]; !ok {
			changes.removed = append(changes.removed, oldAZ)
		}
	}

	// The firewall must keep at least one AZ between removing changed AZs and
	// adding them again.
	if len(newAZs) == 0 || len(oldAZs)+len(changes.added)-len(changes.removed) < 1 {
		return nil, errProxyFirewallReplacementRequired
	}

	return &changes, nil
}

// proxyFirewallAvailabilityZones looks up the AZ of each endpoint subnet and
// NAT gateway mapping and pairs them. Each AZ needs exactly one of each.
func proxyFirewallAvailabilityZones(ctx context.Context, client *conns.AWSClient, tfSubnets, tfNATGateways []any) (map[string]proxyAvailabilityZone, error) {
	subnetAZ := func(id string) (string, error) {
		subnet, err := tfec2.FindSubnetByID(ctx, client.EC2Client(ctx), id)
		if err != nil {
			return "", fmt.Errorf("reading EC2 Subnet (%s): %w", id, err)
		}
		return aws.ToString(subnet.AvailabilityZone), nil
	}

	azs := make(map[string]proxyAvailabilityZone)

	for _, subnet := range expandSubnetMappings(tfSubnets) {
		name, err := subnetAZ(aws.ToString(subnet.SubnetId))
		if err != nil {
			return nil, err
		}
		if _, ok := azs[name]; ok {
			return nil, fmt.Errorf(`proxy firewall can have only one "subnet_mapping" per Availability Zone, got several in %s`, name)
		}
		azs[name] = proxyAvailabilityZone{name: name, subnet: subnet}
	}

	paired := make(map[string]bool)
	for _, natGateway := range expandNatGatewayCreateMappings(tfNATGateways) {
		name, err := subnetAZ(aws.ToString(natGateway.SubnetId))
		if err != nil {
			return nil, err
		}
		az, ok := azs[name]
		if !ok {
			return nil, fmt.Errorf(`"nat_gateway_create_mapping" in %s has no "subnet_mapping" in the same Availability Zone`, name)
		}
		if paired[name] {
			return nil, fmt.Errorf(`proxy firewall can have only one "nat_gateway_create_mapping" per Availability Zone, got several in %s`, name)
		}
		az.natGateway = natGateway
		azs[name] = az
		paired[name] = true
	}

	for name := range azs {
		if !paired[name] {
			return nil, fmt.Errorf(`"subnet_mapping" in %s has no "nat_gateway_create_mapping" in the same Availability Zone`, name)
		}
	}

	return azs, nil
}

func resourceFirewallDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	log.Printf("[DEBUG] Deleting NetworkFirewall Firewall: %s", d.Id())
	input := networkfirewall.DeleteFirewallInput{
		FirewallArn: aws.String(d.Id()),
	}
	// The service rejects DeleteAssociatedNatGateways on firewalls that aren't proxy firewalls.
	if d.Get("nat_gateway_create_mapping").(*schema.Set).Len() > 0 {
		input.DeleteAssociatedNatGateways = d.Get("delete_associated_nat_gateways").(bool)
	}
	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.InvalidOperationException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.DeleteFirewall(ctx, &input)
	}, "still in use")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting NetworkFirewall Firewall (%s): %s", d.Id(), err)
	}

	if _, err := waitFirewallDeleted(ctx, conn, d.Timeout(schema.TimeoutDelete), d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Firewall (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFirewall(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.DescribeFirewallInput) (*networkfirewall.DescribeFirewallOutput, error) {
	output, err := conn.DescribeFirewall(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Firewall == nil || output.FirewallStatus == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findFirewallByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeFirewallOutput, error) {
	input := networkfirewall.DescribeFirewallInput{
		FirewallArn: aws.String(arn),
	}

	return findFirewall(ctx, conn, &input)
}

func statusFirewall(conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFirewallByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FirewallStatus.Status), nil
	}
}

func statusFirewallTransitGatewayAttachment(conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFirewallByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.FirewallStatus.TransitGatewayAttachmentSyncState == nil {
			return nil, "", nil
		}

		return output, string(output.FirewallStatus.TransitGatewayAttachmentSyncState.TransitGatewayAttachmentStatus), nil
	}
}

func waitFirewallCreated(ctx context.Context, conn *networkfirewall.Client, timeout time.Duration, arn string) (*networkfirewall.DescribeFirewallOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallStatusValueProvisioning),
		Target:  enum.Slice(awstypes.FirewallStatusValueReady),
		Refresh: statusFirewall(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok {
		retry.SetLastError(err, failedSyncStatesError(output.FirewallStatus))

		return output, err
	}

	return nil, err
}

func waitFirewallTransitGatewayAttachmentCreated(ctx context.Context, conn *networkfirewall.Client, timeout time.Duration, arn string) (*networkfirewall.DescribeFirewallOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayAttachmentStatusCreating),
		Target:  enum.Slice(awstypes.TransitGatewayAttachmentStatusPendingAcceptance, awstypes.TransitGatewayAttachmentStatusReady),
		Refresh: statusFirewallTransitGatewayAttachment(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok {
		return output, err
	}

	return nil, err
}

func waitFirewallUpdated(ctx context.Context, conn *networkfirewall.Client, timeout time.Duration, arn string) (*networkfirewall.DescribeFirewallOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallStatusValueProvisioning),
		Target:  enum.Slice(awstypes.FirewallStatusValueReady),
		Refresh: statusFirewall(conn, arn),
		Timeout: timeout,
		// Delay added to account for Associate/DisassociateSubnet calls that return
		// a READY status immediately after the method is called instead of immediately
		// returning PROVISIONING.
		Delay: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok {
		retry.SetLastError(err, failedSyncStatesError(output.FirewallStatus))

		return output, err
	}

	return nil, err
}

// waitFirewallAvailabilityZonesRemoved waits until the firewall is READY and
// none of the AZs are left in its sync states. The service rejects associating
// an AZ that's still being removed.
func waitFirewallAvailabilityZonesRemoved(ctx context.Context, conn *networkfirewall.Client, timeout time.Duration, arn string, azNames []string) (*networkfirewall.DescribeFirewallOutput, error) {
	const (
		removed = "REMOVED"
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallStatusValueProvisioning, awstypes.FirewallStatusValueReady),
		Target:  []string{removed},
		Refresh: func(ctx context.Context) (any, string, error) {
			output, err := findFirewallByARN(ctx, conn, arn)

			if err != nil {
				return nil, "", err
			}

			if output.FirewallStatus.Status != awstypes.FirewallStatusValueReady {
				return output, string(output.FirewallStatus.Status), nil
			}

			for _, name := range azNames {
				if _, ok := output.FirewallStatus.SyncStates[name]; ok {
					return output, string(output.FirewallStatus.Status), nil
				}
			}

			return output, removed, nil
		},
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok {
		return output, err
	}

	return nil, err
}

// failedSyncStatesError returns the status messages of failed endpoint and NAT
// gateway attachments, or nil if none failed.
func failedSyncStatesError(apiObject *awstypes.FirewallStatus) error {
	if apiObject == nil {
		return nil
	}

	var failures []error

	for az, syncState := range apiObject.SyncStates {
		if v := syncState.Attachment; v != nil && v.Status == awstypes.AttachmentStatusFailed {
			failures = append(failures, fmt.Errorf("%s endpoint (%s): %s", az, aws.ToString(v.SubnetId), aws.ToString(v.StatusMessage)))
		}

		for _, v := range syncState.NatGatewayAttachments {
			if v.Status == awstypes.NatGatewayAttachmentStatusFailed {
				failures = append(failures, fmt.Errorf("%s NAT gateway (%s): %s", az, aws.ToString(v.NatGatewayId), aws.ToString(v.StatusMessage)))
			}
		}
	}

	return errors.Join(failures...)
}

func waitFirewallDeleted(ctx context.Context, conn *networkfirewall.Client, timeout time.Duration, arn string) (*networkfirewall.DescribeFirewallOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallStatusValueDeleting, awstypes.FirewallStatusValueProvisioning),
		Target:  []string{},
		Refresh: statusFirewall(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok {
		return output, err
	}

	return nil, err
}

func expandSubnetMappings(tfList []any) []awstypes.SubnetMapping {
	apiObjects := make([]awstypes.SubnetMapping, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.SubnetMapping{
			SubnetId: aws.String(tfMap[names.AttrSubnetID].(string)),
		}

		if v, ok := tfMap[names.AttrIPAddressType].(string); ok && v != "" {
			apiObject.IPAddressType = awstypes.IPAddressType(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandSubnetMappingIDs(tfList []any) []string {
	var ids []string

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		if id, ok := tfMap[names.AttrSubnetID].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}

	return ids
}

func flattenFirewallStatus(apiObject *awstypes.FirewallStatus) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"sync_states":                            flattenSyncStates(apiObject.SyncStates),
		"transit_gateway_attachment_sync_states": flattenTransitGatewayAttachmentSyncState(apiObject.TransitGatewayAttachmentSyncState),
	}

	return []any{tfMap}
}

func flattenSyncStates(apiObject map[string]awstypes.SyncState) []any {
	if apiObject == nil {
		return nil
	}

	tfList := make([]any, 0, len(apiObject))

	for k, v := range apiObject {
		tfMap := map[string]any{
			"attachment":               flattenAttachment(v.Attachment),
			names.AttrAvailabilityZone: k,
			"nat_gateway_attachment":   flattenNatGatewayAttachments(v.NatGatewayAttachments),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAttachment(apiObject *awstypes.Attachment) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrDNSName:       aws.ToString(apiObject.DnsName),
		"endpoint_id":           aws.ToString(apiObject.EndpointId),
		names.AttrStatus:        apiObject.Status,
		names.AttrStatusMessage: aws.ToString(apiObject.StatusMessage),
		names.AttrSubnetID:      aws.ToString(apiObject.SubnetId),
	}

	return []any{tfMap}
}

func flattenNatGatewayAttachments(apiObjects []awstypes.NatGatewayAttachment) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"nat_gateway_id":        aws.ToString(apiObject.NatGatewayId),
			names.AttrStatus:        apiObject.Status,
			names.AttrStatusMessage: aws.ToString(apiObject.StatusMessage),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandNatGatewayCreateMappings(tfList []any) []awstypes.NatGatewayCreateMapping {
	apiObjects := make([]awstypes.NatGatewayCreateMapping, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		// Skip the empty element that Plugin SDK v2 can leave in a set whose
		// elements had unknown values at plan time.
		subnetID, _ := tfMap[names.AttrSubnetID].(string)
		if subnetID == "" {
			continue
		}

		apiObject := awstypes.NatGatewayCreateMapping{
			SubnetId: aws.String(subnetID),
		}

		if v, ok := tfMap["allocation_ids"].([]any); ok && len(v) > 0 {
			apiObject.AllocationIds = flex.ExpandStringValueList(v)
		}

		if v, ok := tfMap["private_ips"].([]any); ok && len(v) > 0 {
			apiObject.PrivateIps = flex.ExpandStringValueList(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenNatGatewayCreateMappings(apiObjects []awstypes.NatGatewayCreateMapping) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"allocation_ids":   apiObject.AllocationIds,
			"private_ips":      apiObject.PrivateIps,
			names.AttrSubnetID: aws.ToString(apiObject.SubnetId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandProxySettings(tfList []any) *awstypes.ProxySettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ProxySettings{}

	if v, ok := tfMap["listener"].(*schema.Set); ok {
		for _, tfMapRaw := range v.List() {
			tfMap, ok := tfMapRaw.(map[string]any)
			if !ok || tfMap[names.AttrType].(string) == "" {
				continue
			}

			apiObject.ListenerProperties = append(apiObject.ListenerProperties, awstypes.ListenerProperty{
				Port: aws.Int32(int32(tfMap[names.AttrPort].(int))),
				Type: awstypes.ListenerPropertyType(tfMap[names.AttrType].(string)),
			})
		}
	}

	if v, ok := tfMap[names.AttrCertificateARN].(string); ok && v != "" {
		apiObject.CertificateArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDomainName].(string); ok && v != "" {
		apiObject.DomainName = aws.String(v)
	}

	return apiObject
}

func flattenProxySettings(apiObject *awstypes.ProxySettings) []any {
	if apiObject == nil {
		return nil
	}

	listeners := make([]any, 0, len(apiObject.ListenerProperties))
	for _, v := range apiObject.ListenerProperties {
		listeners = append(listeners, map[string]any{
			names.AttrPort: aws.ToInt32(v.Port),
			names.AttrType: v.Type,
		})
	}

	tfMap := map[string]any{
		names.AttrCertificateARN: aws.ToString(apiObject.CertificateArn),
		names.AttrDomainName:     aws.ToString(apiObject.DomainName),
		"listener":               listeners,
	}

	return []any{tfMap}
}

// validateProxyFirewallDiff enforces the proxy firewall rules that the schema
// can't express. A firewall is a proxy firewall when nat_gateway_create_mapping is set.
func validateProxyFirewallDiff(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	rawConfig := d.GetRawConfig()
	if !rawConfig.IsKnown() || rawConfig.IsNull() {
		return nil
	}

	natGateways := d.Get("nat_gateway_create_mapping").(*schema.Set)
	proxySettings := d.Get("proxy_settings").([]any)

	if natGateways.Len() == 0 {
		if len(proxySettings) > 0 {
			return errors.New(`"proxy_settings" requires "nat_gateway_create_mapping"`)
		}
		if v := rawConfig.GetAttr("transparent_filtering_nat_scope"); !v.IsNull() {
			return errors.New(`"transparent_filtering_nat_scope" requires "nat_gateway_create_mapping"`)
		}
		return nil
	}

	if len(proxySettings) == 0 {
		return errors.New(`"proxy_settings" is required when "nat_gateway_create_mapping" is set`)
	}
	if tfMap, ok := proxySettings[0].(map[string]any); ok {
		// At most one HTTP and one HTTPS listener, on different ports.
		listenerTypes, listenerPorts := make(map[string]bool), make(map[int]bool)
		for _, tfMapRaw := range tfMap["listener"].(*schema.Set).List() {
			listener := tfMapRaw.(map[string]any)
			typ, port := listener[names.AttrType].(string), listener[names.AttrPort].(int)
			if listenerTypes[typ] {
				return fmt.Errorf(`"proxy_settings" can have only one %s "listener"`, typ)
			}
			if listenerPorts[port] {
				return fmt.Errorf(`"proxy_settings" "listener" ports must be different, got %d twice`, port)
			}
			listenerTypes[typ], listenerPorts[port] = true, true
		}
	}

	// Use the raw config: list lengths may be unknown until apply.
	if v := rawConfig.GetAttr("nat_gateway_create_mapping"); v.IsKnown() && !v.IsNull() {
		for it := v.ElementIterator(); it.Next(); {
			_, tfObject := it.Element()
			if !tfObject.IsKnown() || tfObject.IsNull() {
				continue
			}

			allocationIDs, privateIPs := tfObject.GetAttr("allocation_ids"), tfObject.GetAttr("private_ips")
			if !allocationIDs.IsKnown() || !privateIPs.IsKnown() {
				continue
			}

			var allocationIDsLen, privateIPsLen int
			if !allocationIDs.IsNull() {
				allocationIDsLen = allocationIDs.LengthInt()
			}
			if !privateIPs.IsNull() {
				privateIPsLen = privateIPs.LengthInt()
			}

			if allocationIDsLen == 0 && privateIPsLen == 0 {
				return errors.New(`each "nat_gateway_create_mapping" requires "allocation_ids" (public NAT gateway) or "private_ips" (private NAT gateway)`)
			}
			// A public NAT gateway needs exactly one private IP per Elastic IP.
			if allocationIDsLen > 0 && allocationIDsLen != privateIPsLen {
				return fmt.Errorf(`"nat_gateway_create_mapping" with %d "allocation_ids" requires %d "private_ips", got %d`, allocationIDsLen, allocationIDsLen, privateIPsLen)
			}
		}
	}

	// Subnet IDs of resources created in the same apply aren't known yet. The
	// service checks the pairing on create, and the update checks it again.
	if !rawConfig.GetAttr("subnet_mapping").IsWhollyKnown() || !rawConfig.GetAttr("nat_gateway_create_mapping").IsWhollyKnown() {
		return nil
	}

	// ResourceDiff returns nested lists inside changed set elements as unknown,
	// so take the new values from the configuration, which is wholly known here.
	client := meta.(*conns.AWSClient)
	newSubnets := tfMapsFromCtySet(rawConfig.GetAttr("subnet_mapping"))
	newNATGateways := tfMapsFromCtySet(rawConfig.GetAttr("nat_gateway_create_mapping"))

	// Converting to or from a proxy firewall replaces it, so a new or converted
	// firewall only needs its endpoint subnets and NAT gateways paired by AZ.
	if o, _ := d.GetChange("nat_gateway_create_mapping"); d.Id() == "" || o.(*schema.Set).Len() == 0 {
		_, err := proxyFirewallAvailabilityZones(ctx, client, newSubnets, newNATGateways)
		return err
	}

	if !d.HasChanges("subnet_mapping", "nat_gateway_create_mapping") {
		return nil
	}

	o, _ := d.GetChange("subnet_mapping")
	oldSubnets := o.(*schema.Set).List()
	o, _ = d.GetChange("nat_gateway_create_mapping")
	oldNATGateways := o.(*schema.Set).List()

	_, err := planProxyFirewallAvailabilityZones(ctx, client, oldSubnets, oldNATGateways, newSubnets, newNATGateways)

	if !errors.Is(err, errProxyFirewallReplacementRequired) {
		return err
	}

	// Replacing the firewall deletes it first. Unless the NAT gateways it created
	// are deleted with it, they keep their Elastic IPs, and a new firewall that
	// reuses one of them would fail to create its NAT gateway.
	if !d.Get("delete_associated_nat_gateways").(bool) {
		oldAllocationIDs := make(map[string]bool)
		for _, v := range expandNatGatewayCreateMappings(oldNATGateways) {
			for _, id := range v.AllocationIds {
				oldAllocationIDs[id] = true
			}
		}
		for _, v := range expandNatGatewayCreateMappings(newNATGateways) {
			for _, id := range v.AllocationIds {
				if oldAllocationIDs[id] {
					return fmt.Errorf(`changing the only Availability Zone of a proxy firewall replaces the firewall, but Elastic IP %s stays with the NAT gateway of the old firewall because "delete_associated_nat_gateways" is false: set it to true, or use a different Elastic IP`, id)
				}
			}
		}
	}

	// ForceNew on a set only takes effect when its element count changes, and
	// here it doesn't. Force replacement through firewall_status instead, which
	// the ComputedIf above marks as changing whenever the AZs change.
	if err := d.SetNewComputed("firewall_status"); err != nil {
		return err
	}

	return d.ForceNew("firewall_status")
}

// tfMapsFromCtySet converts a wholly known set of objects from the raw
// configuration into the []any of map[string]any that the expand functions
// take. Null attributes become zero values; lists become []any.
func tfMapsFromCtySet(v cty.Value) []any {
	if v.IsNull() || !v.IsKnown() {
		return nil
	}

	var tfList []any
	for it := v.ElementIterator(); it.Next(); {
		_, tfObject := it.Element()
		tfMap := make(map[string]any)
		for name, attr := range tfObject.AsValueMap() {
			switch {
			case attr.IsNull():
				if attr.Type().IsListType() {
					tfMap[name] = []any{}
				} else {
					tfMap[name] = ""
				}
			case attr.Type().IsListType():
				var elems []any
				for _, e := range attr.AsValueSlice() {
					elems = append(elems, e.AsString())
				}
				tfMap[name] = elems
			default:
				tfMap[name] = attr.AsString()
			}
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSubnetMappings(apiObjects []awstypes.SubnetMapping) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrIPAddressType: apiObject.IPAddressType,
			names.AttrSubnetID:      aws.ToString(apiObject.SubnetId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func subnetMappingsDiff(old, new *schema.Set) ([]string, []awstypes.SubnetMapping) {
	if old.Len() == 0 {
		return nil, expandSubnetMappings(new.List())
	}
	if new.Len() == 0 {
		return expandSubnetMappingIDs(old.List()), nil
	}

	subnetMappingsHash := sdkv2.SimpleSchemaSetFunc(names.AttrIPAddressType, names.AttrSubnetID)
	oldHashedSet := schema.NewSet(subnetMappingsHash, old.List())
	newHashedSet := schema.NewSet(subnetMappingsHash, new.List())

	toRemove := oldHashedSet.Difference(newHashedSet)
	toAdd := new.Difference(old)

	subnetsToRemove := expandSubnetMappingIDs(toRemove.List())
	subnetsToAdd := expandSubnetMappings(toAdd.List())

	return subnetsToRemove, subnetsToAdd
}

func expandAvailabilityZoneMapping(tfList []any) []awstypes.AvailabilityZoneMapping {
	apiObjects := make([]awstypes.AvailabilityZoneMapping, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.AvailabilityZoneMapping{
			AvailabilityZone: aws.String(tfMap["availability_zone_id"].(string)),
		}

		if v, ok := tfMap["availability_zone_id"].(string); ok && v != "" {
			apiObject.AvailabilityZone = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenAvailabilityZoneMapping(apiObjects []awstypes.AvailabilityZoneMapping) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"availability_zone_id": aws.ToString(apiObject.AvailabilityZone),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTransitGatewayAttachmentSyncState(apiObject *awstypes.TransitGatewayAttachmentSyncState) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"attachment_id": apiObject.AttachmentId,
	}

	return []any{tfMap}
}

func availabilityZoneMappingsDiff(old, new *schema.Set) ([]awstypes.AvailabilityZoneMapping, []awstypes.AvailabilityZoneMapping) {
	if old.Len() == 0 {
		return nil, expandAvailabilityZoneMapping(new.List())
	}
	if new.Len() == 0 {
		return expandAvailabilityZoneMapping(old.List()), nil
	}

	toRemove := old.Difference(new)
	toAdd := new.Difference(old)

	availabilityZonesToRemove := expandAvailabilityZoneMapping(toRemove.List())
	availabilityZonesToAdd := expandAvailabilityZoneMapping(toAdd.List())

	return availabilityZonesToRemove, availabilityZonesToAdd
}
