// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package networkfirewall

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
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
// @ArnIdentity
// @CustomImport
// @Testing(preIdentityVersion="v6.72.0")
// @Testing(importIgnore="delete_associated_nat_gateways;update_token")
func resourceFirewall() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallCreate,
		ReadWithoutTimeout:   resourceFirewallRead,
		UpdateWithoutTimeout: resourceFirewallUpdate,
		DeleteWithoutTimeout: resourceFirewallDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.Import(ctx, rd, meta); err != nil {
					return nil, err
				}

				rd.Set("delete_associated_nat_gateways", false)

				return []*schema.ResourceData{rd}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("firewall_status", func(ctx context.Context, diff *schema.ResourceDiff, meta any) bool {
				return diff.HasChanges("nat_gateway_create_mapping", "subnet_mapping")
			}),
			// A public NAT gateway needs one private IP address for each Elastic IP address.
			func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
				for i, v := range diff.Get("nat_gateway_create_mapping").([]any) {
					tfMap, ok := v.(map[string]any)
					if !ok {
						continue
					}

					if !diff.NewValueKnown(fmt.Sprintf("nat_gateway_create_mapping.%d.allocation_ids", i)) || !diff.NewValueKnown(fmt.Sprintf("nat_gateway_create_mapping.%d.private_ips", i)) {
						continue
					}

					if a, p := len(tfMap["allocation_ids"].([]any)), len(tfMap["private_ips"].([]any)); a > 0 && a != p {
						return fmt.Errorf("nat_gateway_create_mapping.%d: specify one private_ips address for each of the %d allocation_ids", i, a)
					}
				}

				return nil
			},
			// A firewall's capability (routed or NAT/Proxy) and its NAT gateways' settings can't be changed after creation.
			// A NAT gateway mapping can only be removed along with its Availability Zone's subnet_mapping.
			func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
				if diff.Id() == "" {
					return nil
				}

				// NAT gateway mappings can only be added together with the endpoint subnet for a new Availability Zone.
				o, n := diff.GetChange("nat_gateway_create_mapping")
				os, ns := natGatewayCreateMappingsSet(o), natGatewayCreateMappingsSet(n)
				if os.Len() > 0 && os.Difference(ns).Len() == 0 && ns.Difference(os).Len() > 0 && !diff.HasChange("subnet_mapping") {
					return errors.New("nat_gateway_create_mapping can only be added along with the subnet_mapping for a new Availability Zone")
				}

				return nil
			},
			func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
				if diff.Id() == "" {
					return nil
				}

				o, n := diff.GetChange("nat_gateway_create_mapping")
				ol, nl := o.([]any), n.([]any)
				if (len(ol) == 0) != (len(nl) == 0) {
					return diff.ForceNew("nat_gateway_create_mapping")
				}

				os, ns := natGatewayCreateMappingsSet(o), natGatewayCreateMappingsSet(n)
				if os.Difference(ns).Len() == 0 {
					return nil
				}
				if o, n := diff.GetChange("subnet_mapping"); o.(*schema.Set).Difference(n.(*schema.Set)).Len() > 0 {
					return nil
				}

				// ForceNew on a list of blocks only applies to a change in the number of blocks,
				// so force a new resource on the changed attributes of each changed mapping.
				for i, v := range nl {
					if os.Contains(v) {
						continue
					}
					for _, k := range []string{"allocation_ids", "private_ips", names.AttrSubnetID} {
						if key := fmt.Sprintf("nat_gateway_create_mapping.%d.%s", i, k); diff.HasChange(key) {
							if err := diff.ForceNew(key); err != nil {
								return err
							}
						}
					}
				}

				return nil
			},
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
				"delete_associated_nat_gateways": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
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
					Type:          schema.TypeList,
					Optional:      true,
					ConflictsWith: []string{"availability_zone_mapping", names.AttrTransitGatewayID},
					Elem:          natGatewayCreateMappingResource(),
				},
				"proxy_settings": {
					Type:         schema.TypeList,
					Optional:     true,
					Computed:     true,
					MaxItems:     1,
					RequiredWith: []string{"nat_gateway_create_mapping"},
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
								RequiredWith: []string{"proxy_settings.0.certificate_arn"},
							},
							"listener_property": {
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
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
					RequiredWith:     []string{"nat_gateway_create_mapping"},
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

	// A NAT/Proxy firewall takes its VPC and endpoint subnets in VpcEndpoint, not at the top level.
	if v := d.Get("nat_gateway_create_mapping").([]any); len(v) > 0 {
		input.NatGatewayCreateMappings = expandNatGatewayCreateMappings(v)
		input.VpcEndpoint = &awstypes.VpcEndpoint{
			SubnetMappings: input.SubnetMappings,
			VpcId:          input.VpcId,
		}
		input.SubnetMappings = nil
		input.VpcId = nil

		if v, ok := d.GetOk("proxy_settings"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ProxySettings = expandProxySettings(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk("transparent_filtering_nat_scope"); ok {
			input.TransparentFilteringNatScope = awstypes.TransparentFilteringNatScope(v.(string))
		}
	}

	// The name of a just-deleted firewall can stay reserved for a short time, e.g. on replacement.
	const (
		timeout = 5 * time.Minute
	)
	output, err := tfresource.RetryWhenIsAErrorMessageContains[*networkfirewall.CreateFirewallOutput, *awstypes.InvalidRequestException](ctx, timeout, func(ctx context.Context) (*networkfirewall.CreateFirewallOutput, error) {
		return conn.CreateFirewall(ctx, &input)
	}, "already exists")

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

	return resourceFirewallFlatten(ctx, d, meta.(*conns.AWSClient).EC2Client(ctx), output)
}

func resourceFirewallFlatten(ctx context.Context, d *schema.ResourceData, ec2Conn *ec2.Client, output *networkfirewall.DescribeFirewallOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	firewall := output.Firewall
	d.Set(names.AttrARN, firewall.FirewallArn)
	d.Set("availability_zone_change_protection", firewall.AvailabilityZoneChangeProtection)
	d.Set("availability_zone_mapping", flattenAvailabilityZoneMapping(firewall.AvailabilityZoneMappings))
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
	// DescribeFirewall doesn't return the mappings, so rebuild them from the NAT gateways attached to the firewall.
	natGatewayCreateMappings := firewall.NatGatewayCreateMappings
	if natGatewayCreateMappings == nil {
		var err error
		natGatewayCreateMappings, err = findNATGatewayCreateMappings(ctx, ec2Conn, output.FirewallStatus)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall (%s) NAT gateways: %s", d.Id(), err)
		}
	}
	natGatewayCreateMappings = orderNATGatewayCreateMappings(natGatewayCreateMappings, d.Get("nat_gateway_create_mapping").([]any))
	if err := d.Set("nat_gateway_create_mapping", flattenNatGatewayCreateMappings(natGatewayCreateMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting nat_gateway_create_mapping: %s", err)
	}
	if err := d.Set("proxy_settings", flattenProxySettings(firewall.ProxySettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting proxy_settings: %s", err)
	}
	d.Set("subnet_change_protection", firewall.SubnetChangeProtection)
	// A NAT/Proxy firewall's VpcEndpoint keeps the subnets that it was created with,
	// while the top-level SubnetMappings also has the ones associated later.
	subnetMappings, vpcID := firewall.SubnetMappings, firewall.VpcId
	if v := firewall.VpcEndpoint; v != nil {
		if len(subnetMappings) == 0 {
			subnetMappings = v.SubnetMappings
		}
		if vpcID == nil {
			vpcID = v.VpcId
		}
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

	var natGatewayMappingsToAdd []awstypes.NatGatewayCreateMapping
	if d.HasChange("nat_gateway_create_mapping") {
		o, n := d.GetChange("nat_gateway_create_mapping")
		natGatewayMappingsToAdd = expandNatGatewayCreateMappings(natGatewayCreateMappingsSet(n).Difference(natGatewayCreateMappingsSet(o)).List())
	}

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

	if d.HasChange("proxy_settings") {
		if v, ok := d.GetOk("proxy_settings"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input := networkfirewall.UpdateProxySettingsInput{
				FirewallArn:   aws.String(d.Id()),
				ProxySettings: expandProxySettings(v.([]any)[0].(map[string]any)),
				UpdateToken:   aws.String(updateToken),
			}

			output, err := conn.UpdateProxySettings(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) proxy settings: %s", d.Id(), err)
			}

			updateToken = aws.ToString(output.UpdateToken)
		}
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

	if d.HasChange("subnet_mapping") {
		o, n := d.GetChange("subnet_mapping")
		subnetsToRemove, subnetsToAdd := subnetMappingsDiff(o.(*schema.Set), n.(*schema.Set))

		if len(subnetsToAdd) > 0 {
			input := networkfirewall.AssociateSubnetsInput{
				FirewallArn:              aws.String(d.Id()),
				NatGatewayCreateMappings: natGatewayMappingsToAdd,
				SubnetMappings:           subnetsToAdd,
				UpdateToken:              aws.String(updateToken),
			}

			_, err := conn.AssociateSubnets(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating NetworkFirewall Firewall (%s) subnets: %s", d.Id(), err)
			}

			output, err := waitFirewallUpdated(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id())

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Firewall (%s) update: %s", d.Id(), err)
			}

			updateToken = aws.ToString(output.UpdateToken)
		}

		if len(subnetsToRemove) > 0 {
			input := networkfirewall.DisassociateSubnetsInput{
				FirewallArn: aws.String(d.Id()),
				SubnetIds:   subnetsToRemove,
				UpdateToken: aws.String(updateToken),
			}

			_, err := conn.DisassociateSubnets(ctx, &input)

			if err == nil {
				/*output*/ _, err := waitFirewallUpdated(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id())

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Firewall (%s) update: %s", d.Id(), err)
				}

				// updateToken = aws.ToString(output.UpdateToken)
			} else if !errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "inaccessible") {
				return sdkdiag.AppendErrorf(diags, "disassociating NetworkFirewall Firewall (%s) subnets: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceFirewallRead(ctx, d, meta)...)
}

func resourceFirewallDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	log.Printf("[DEBUG] Deleting NetworkFirewall Firewall: %s", d.Id())
	input := networkfirewall.DeleteFirewallInput{
		FirewallArn: aws.String(d.Id()),
	}
	// DeleteAssociatedNatGateways can't be set for a routed firewall, which has no proxy settings.
	if len(d.Get("proxy_settings").([]any)) > 0 {
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

// findNATGatewayCreateMappings describes the NAT gateways that are attached to a NAT/Proxy firewall.
// NAT gateways detached from removed Availability Zones aren't in the firewall status, so they're left out.
func findNATGatewayCreateMappings(ctx context.Context, conn *ec2.Client, status *awstypes.FirewallStatus) ([]awstypes.NatGatewayCreateMapping, error) {
	if status == nil {
		return nil, nil
	}

	var natGatewayIDs []string
	for _, syncState := range status.SyncStates {
		for _, v := range syncState.NatGatewayAttachments {
			natGatewayIDs = append(natGatewayIDs, aws.ToString(v.NatGatewayId))
		}
	}

	if len(natGatewayIDs) == 0 {
		return nil, nil
	}

	input := ec2.DescribeNatGatewaysInput{
		NatGatewayIds: natGatewayIDs,
	}
	natGateways, err := tfec2.FindNATGateways(ctx, conn, &input)

	if retry.NotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	apiObjects := make([]awstypes.NatGatewayCreateMapping, 0, len(natGateways))
	for _, natGateway := range natGateways {
		apiObject := awstypes.NatGatewayCreateMapping{
			SubnetId: natGateway.SubnetId,
		}

		// The primary address comes first.
		addresses := slices.Clone(natGateway.NatGatewayAddresses)
		slices.SortStableFunc(addresses, func(a, b ec2types.NatGatewayAddress) int {
			switch ap, bp := aws.ToBool(a.IsPrimary), aws.ToBool(b.IsPrimary); {
			case ap == bp:
				return 0
			case ap:
				return -1
			default:
				return 1
			}
		})
		for _, v := range addresses {
			switch v.Status {
			case ec2types.NatGatewayAddressStatusFailed, ec2types.NatGatewayAddressStatusDisassociating, ec2types.NatGatewayAddressStatusUnassigning:
				continue
			}

			if v.AllocationId != nil {
				apiObject.AllocationIds = append(apiObject.AllocationIds, aws.ToString(v.AllocationId))
			}
			apiObject.PrivateIps = append(apiObject.PrivateIps, aws.ToString(v.PrivateIp))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

// orderNATGatewayCreateMappings orders mappings as in tfList, by subnet ID, so that a list read back from
// the NAT gateways doesn't differ from configuration only in order. Other mappings follow by subnet ID.
func orderNATGatewayCreateMappings(apiObjects []awstypes.NatGatewayCreateMapping, tfList []any) []awstypes.NatGatewayCreateMapping {
	index := make(map[string]int)
	for i, v := range tfList {
		if tfMap, ok := v.(map[string]any); ok {
			index[tfMap[names.AttrSubnetID].(string)] = i
		}
	}

	key := func(apiObject awstypes.NatGatewayCreateMapping) (int, string) {
		subnetID := aws.ToString(apiObject.SubnetId)
		if i, ok := index[subnetID]; ok {
			return i, subnetID
		}
		return len(tfList), subnetID
	}

	apiObjects = slices.Clone(apiObjects)
	slices.SortStableFunc(apiObjects, func(a, b awstypes.NatGatewayCreateMapping) int {
		ai, as := key(a)
		bi, bs := key(b)
		return cmp.Or(cmp.Compare(ai, bi), cmp.Compare(as, bs))
	})

	return apiObjects
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

// statusFirewallNATGatewayAttachments wraps statusFirewall and fails as soon as Network Firewall can't create
// or attach a NAT/Proxy firewall's NAT gateway, since the firewall then never becomes READY.
func statusFirewallNATGatewayAttachments(conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	refresh := statusFirewall(conn, arn)

	return func(ctx context.Context) (any, string, error) {
		outputRaw, status, err := refresh(ctx)

		if output, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok && err == nil {
			for az, syncState := range output.FirewallStatus.SyncStates {
				for _, v := range syncState.NatGatewayAttachments {
					if v.Status == awstypes.NatGatewayAttachmentStatusFailed {
						return output, status, fmt.Errorf("NAT gateway (%s) in Availability Zone (%s): %s", aws.ToString(v.NatGatewayId), az, aws.ToString(v.StatusMessage))
					}
				}
			}
		}

		return outputRaw, status, err
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
		Refresh: statusFirewallNATGatewayAttachments(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok {
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
		Refresh: statusFirewallNATGatewayAttachments(conn, arn),
		Timeout: timeout,
		// Delay added to account for Associate/DisassociateSubnet calls that return
		// a READY status immediately after the method is called instead of immediately
		// returning PROVISIONING.
		Delay: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok {
		return output, err
	}

	return nil, err
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
		names.AttrDNSName:  aws.ToString(apiObject.DnsName),
		"endpoint_id":      aws.ToString(apiObject.EndpointId),
		names.AttrSubnetID: aws.ToString(apiObject.SubnetId),
	}

	return []any{tfMap}
}

func flattenNatGatewayAttachments(apiObjects []awstypes.NatGatewayAttachment) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"nat_gateway_id": aws.ToString(apiObject.NatGatewayId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

// natGatewayCreateMappingResource is a list element, not a set element, as a set holds an extra
// empty element while its nested values are unknown during planning.
func natGatewayCreateMappingResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"allocation_ids": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 8,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_ips": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
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
	}
}

// natGatewayCreateMappingsSet compares NAT gateway mappings regardless of their order.
func natGatewayCreateMappingsSet(v any) *schema.Set {
	return schema.NewSet(schema.HashResource(natGatewayCreateMappingResource()), v.([]any))
}

func expandNatGatewayCreateMappings(tfList []any) []awstypes.NatGatewayCreateMapping {
	apiObjects := make([]awstypes.NatGatewayCreateMapping, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		subnetID, ok := tfMap[names.AttrSubnetID].(string)
		if !ok || subnetID == "" {
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

func expandProxySettings(tfMap map[string]any) *awstypes.ProxySettings {
	apiObject := &awstypes.ProxySettings{}

	if v, ok := tfMap[names.AttrCertificateARN].(string); ok && v != "" {
		apiObject.CertificateArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDomainName].(string); ok && v != "" {
		apiObject.DomainName = aws.String(v)
	}

	if v, ok := tfMap["listener_property"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ListenerProperties = expandListenerProperties(v.List())
	}

	return apiObject
}

func expandListenerProperties(tfList []any) []awstypes.ListenerProperty {
	apiObjects := make([]awstypes.ListenerProperty, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, awstypes.ListenerProperty{
			Port: aws.Int32(int32(tfMap[names.AttrPort].(int))),
			Type: awstypes.ListenerPropertyType(tfMap[names.AttrType].(string)),
		})
	}

	return apiObjects
}

func flattenProxySettings(apiObject *awstypes.ProxySettings) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrCertificateARN: aws.ToString(apiObject.CertificateArn),
		names.AttrDomainName:     aws.ToString(apiObject.DomainName),
		"listener_property":      flattenListenerProperties(apiObject.ListenerProperties),
	}

	return []any{tfMap}
}

func flattenListenerProperties(apiObjects []awstypes.ListenerProperty) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrPort: aws.ToInt32(apiObject.Port),
			names.AttrType: apiObject.Type,
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
