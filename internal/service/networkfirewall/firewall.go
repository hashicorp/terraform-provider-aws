// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package networkfirewall

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
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
				return diff.HasChange("subnet_mapping")
			}),
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
	d.Set("subnet_change_protection", firewall.SubnetChangeProtection)
	if err := d.Set("subnet_mapping", flattenSubnetMappings(firewall.SubnetMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_mapping: %s", err)
	}
	d.Set(names.AttrTransitGatewayID, firewall.TransitGatewayId)
	d.Set("transit_gateway_owner_account_id", firewall.TransitGatewayOwnerAccountId)
	d.Set("update_token", output.UpdateToken)
	d.Set(names.AttrVPCID, firewall.VpcId)

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

	if d.HasChange("subnet_mapping") {
		o, n := d.GetChange("subnet_mapping")
		subnetsToRemove, subnetsToAdd := subnetMappingsDiff(o.(*schema.Set), n.(*schema.Set))

		if len(subnetsToAdd) > 0 {
			input := networkfirewall.AssociateSubnetsInput{
				FirewallArn:    aws.String(d.Id()),
				SubnetMappings: subnetsToAdd,
				UpdateToken:    aws.String(updateToken),
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
		"endpoint_id":      aws.ToString(apiObject.EndpointId),
		names.AttrSubnetID: aws.ToString(apiObject.SubnetId),
	}

	return []any{tfMap}
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
