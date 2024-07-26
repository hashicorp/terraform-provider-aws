// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("firewall_status", func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("subnet_mapping")
			}),
			verify.SetTagsDiff,
		),

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
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
					Required: true,
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
				"update_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrVPCID: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
			}
		},
	}
}

func resourceFirewallCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &networkfirewall.CreateFirewallInput{
		FirewallName:      aws.String(name),
		FirewallPolicyArn: aws.String(d.Get("firewall_policy_arn").(string)),
		SubnetMappings:    expandSubnetMappings(d.Get("subnet_mapping").(*schema.Set).List()),
		Tags:              getTagsIn(ctx),
		VpcId:             aws.String(d.Get(names.AttrVPCID).(string)),
	}

	if v, ok := d.GetOk("delete_protection"); ok {
		input.DeleteProtection = v.(bool)
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEncryptionConfiguration); ok {
		input.EncryptionConfiguration = expandEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("firewall_policy_change_protection"); ok {
		input.FirewallPolicyChangeProtection = v.(bool)
	}

	if v, ok := d.GetOk("subnet_change_protection"); ok {
		input.SubnetChangeProtection = v.(bool)
	}

	output, err := conn.CreateFirewall(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating NetworkFirewall Firewall (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Firewall.FirewallArn))

	if _, err := waitFirewallCreated(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for NetworkFirewall Firewall (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFirewallRead(ctx, d, meta)...)
}

func resourceFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	output, err := findFirewallByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] NetworkFirewall Firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall (%s): %s", d.Id(), err)
	}

	firewall := output.Firewall
	d.Set(names.AttrARN, firewall.FirewallArn)
	d.Set("delete_protection", firewall.DeleteProtection)
	d.Set(names.AttrDescription, firewall.Description)
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
	d.Set("update_token", output.UpdateToken)
	d.Set(names.AttrVPCID, firewall.VpcId)

	setTagsOut(ctx, firewall.Tags)

	return diags
}

func resourceFirewallUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	updateToken := d.Get("update_token").(string)

	if d.HasChange("delete_protection") {
		input := &networkfirewall.UpdateFirewallDeleteProtectionInput{
			DeleteProtection: d.Get("delete_protection").(bool),
			FirewallArn:      aws.String(d.Id()),
			UpdateToken:      aws.String(updateToken),
		}

		output, err := conn.UpdateFirewallDeleteProtection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) delete protection: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange(names.AttrDescription) {
		input := &networkfirewall.UpdateFirewallDescriptionInput{
			Description: aws.String(d.Get(names.AttrDescription).(string)),
			FirewallArn: aws.String(d.Id()),
			UpdateToken: aws.String(updateToken),
		}

		output, err := conn.UpdateFirewallDescription(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) description: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange(names.AttrEncryptionConfiguration) {
		input := &networkfirewall.UpdateFirewallEncryptionConfigurationInput{
			EncryptionConfiguration: expandEncryptionConfiguration(d.Get(names.AttrEncryptionConfiguration).([]interface{})),
			FirewallArn:             aws.String(d.Id()),
			UpdateToken:             aws.String(updateToken),
		}

		output, err := conn.UpdateFirewallEncryptionConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) encryption configuration: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	// Note: The *_change_protection fields below are handled before their respective fields
	// to account for disabling and subsequent changes.

	if d.HasChange("firewall_policy_change_protection") {
		input := &networkfirewall.UpdateFirewallPolicyChangeProtectionInput{
			FirewallArn:                    aws.String(d.Id()),
			FirewallPolicyChangeProtection: d.Get("firewall_policy_change_protection").(bool),
			UpdateToken:                    aws.String(updateToken),
		}

		output, err := conn.UpdateFirewallPolicyChangeProtection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) policy change protection: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange("firewall_policy_arn") {
		input := &networkfirewall.AssociateFirewallPolicyInput{
			FirewallArn:       aws.String(d.Id()),
			FirewallPolicyArn: aws.String(d.Get("firewall_policy_arn").(string)),
			UpdateToken:       aws.String(updateToken),
		}

		output, err := conn.AssociateFirewallPolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) firewall policy ARN: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange("subnet_change_protection") {
		input := &networkfirewall.UpdateSubnetChangeProtectionInput{
			FirewallArn:            aws.String(d.Id()),
			SubnetChangeProtection: d.Get("subnet_change_protection").(bool),
			UpdateToken:            aws.String(updateToken),
		}

		output, err := conn.UpdateSubnetChangeProtection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating NetworkFirewall Firewall (%s) subnet change protection: %s", d.Id(), err)
		}

		updateToken = aws.ToString(output.UpdateToken)
	}

	if d.HasChange("subnet_mapping") {
		o, n := d.GetChange("subnet_mapping")
		subnetsToRemove, subnetsToAdd := subnetMappingsDiff(o.(*schema.Set), n.(*schema.Set))

		if len(subnetsToAdd) > 0 {
			input := &networkfirewall.AssociateSubnetsInput{
				FirewallArn:    aws.String(d.Id()),
				SubnetMappings: subnetsToAdd,
				UpdateToken:    aws.String(updateToken),
			}

			_, err := conn.AssociateSubnets(ctx, input)

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
			input := &networkfirewall.DisassociateSubnetsInput{
				FirewallArn: aws.String(d.Id()),
				SubnetIds:   subnetsToRemove,
				UpdateToken: aws.String(updateToken),
			}

			_, err := conn.DisassociateSubnets(ctx, input)

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

func resourceFirewallDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	log.Printf("[DEBUG] Deleting NetworkFirewall Firewall: %s", d.Id())
	_, err := conn.DeleteFirewall(ctx, &networkfirewall.DeleteFirewallInput{
		FirewallArn: aws.String(d.Id()),
	})

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
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Firewall == nil || output.FirewallStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findFirewallByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeFirewallOutput, error) {
	input := &networkfirewall.DescribeFirewallInput{
		FirewallArn: aws.String(arn),
	}

	return findFirewall(ctx, conn, input)
}

func statusFirewall(ctx context.Context, conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFirewallByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FirewallStatus.Status), nil
	}
}

func waitFirewallCreated(ctx context.Context, conn *networkfirewall.Client, timeout time.Duration, arn string) (*networkfirewall.DescribeFirewallOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallStatusValueProvisioning),
		Target:  enum.Slice(awstypes.FirewallStatusValueReady),
		Refresh: statusFirewall(ctx, conn, arn),
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
		Refresh: statusFirewall(ctx, conn, arn),
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
		Pending: enum.Slice(awstypes.FirewallStatusValueDeleting),
		Target:  []string{},
		Refresh: statusFirewall(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeFirewallOutput); ok {
		return output, err
	}

	return nil, err
}

func expandSubnetMappings(tfList []interface{}) []awstypes.SubnetMapping {
	apiObjects := make([]awstypes.SubnetMapping, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
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

func expandSubnetMappingIDs(tfList []interface{}) []string {
	var ids []string

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if id, ok := tfMap[names.AttrSubnetID].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}

	return ids
}

func flattenFirewallStatus(apiObject *awstypes.FirewallStatus) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"sync_states": flattenSyncStates(apiObject.SyncStates),
	}

	return []interface{}{tfMap}
}

func flattenSyncStates(apiObject map[string]awstypes.SyncState) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := make([]interface{}, 0, len(apiObject))

	for k, v := range apiObject {
		tfMap := map[string]interface{}{
			"attachment":               flattenAttachment(v.Attachment),
			names.AttrAvailabilityZone: k,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAttachment(apiObject *awstypes.Attachment) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"endpoint_id":      aws.ToString(apiObject.EndpointId),
		names.AttrSubnetID: aws.ToString(apiObject.SubnetId),
	}

	return []interface{}{tfMap}
}

func flattenSubnetMappings(apiObjects []awstypes.SubnetMapping) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
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
