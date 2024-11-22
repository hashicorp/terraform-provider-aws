// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package worklink

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/worklink"
	awstypes "github.com/aws/aws-sdk-go-v2/service/worklink/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_worklink_fleet", name="Fleet")
func resourceFleet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFleetCreate,
		ReadWithoutTimeout:   resourceFleetRead,
		UpdateWithoutTimeout: resourceFleetUpdate,
		DeleteWithoutTimeout: resourceFleetDelete,

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
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z](?:[0-9a-z\-]{0,46}[0-9a-z])?$`), "must contain only alphanumeric characters"),
					validation.StringLenBetween(1, 48),
				),
			},
			names.AttrDisplayName: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"audit_stream_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"network": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},
			"device_ca_certificate": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(v interface{}) string {
					s, ok := v.(string)
					if !ok {
						return ""
					}
					return strings.TrimSpace(s)
				},
			},
			"identity_provider": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
						},
						"saml_metadata": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 204800),
						},
					},
				},
			},
			"company_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLastUpdatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"optimize_for_end_user_location": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},

		DeprecationMessage: `The aws_worklink_fleet resource has been deprecated and will be removed in a future version. Use Amazon WorkSpaces Secure Browser instead`,
	}
}

func resourceFleetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkClient(ctx)

	input := &worklink.CreateFleetInput{
		FleetName:                  aws.String(d.Get(names.AttrName).(string)),
		OptimizeForEndUserLocation: aws.Bool(d.Get("optimize_for_end_user_location").(bool)),
	}

	if v, ok := d.GetOk(names.AttrDisplayName); ok {
		input.DisplayName = aws.String(v.(string))
	}

	resp, err := conn.CreateFleet(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkLink Fleet: %s", err)
	}

	d.SetId(aws.ToString(resp.FleetArn))

	if err := updateAuditStreamConfiguration(ctx, conn, d); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkLink Fleet: %s", err)
	}

	if err := updateCompanyNetworkConfiguration(ctx, conn, d); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkLink Fleet: %s", err)
	}

	if err := updateDevicePolicyConfiguration(ctx, conn, d); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkLink Fleet: %s", err)
	}

	if err := updateIdentityProviderConfiguration(ctx, conn, d); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkLink Fleet: %s", err)
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkClient(ctx)

	output, err := findFleetByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkLink Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing WorkLink Fleet (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, d.Id())
	d.Set(names.AttrName, output.FleetName)
	d.Set(names.AttrDisplayName, output.DisplayName)
	d.Set("optimize_for_end_user_location", output.OptimizeForEndUserLocation)
	d.Set("company_code", output.CompanyCode)
	d.Set(names.AttrCreatedTime, output.CreatedTime.Format(time.RFC3339))
	if output.LastUpdatedTime != nil {
		d.Set(names.AttrLastUpdatedTime, output.LastUpdatedTime.Format(time.RFC3339))
	}
	auditStreamConfigurationResp, err := conn.DescribeAuditStreamConfiguration(ctx, &worklink.DescribeAuditStreamConfigurationInput{
		FleetArn: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing WorkLink Fleet (%s) audit stream configuration: %s", d.Id(), err)
	}
	d.Set("audit_stream_arn", auditStreamConfigurationResp.AuditStreamArn)

	companyNetworkConfigurationResp, err := conn.DescribeCompanyNetworkConfiguration(ctx, &worklink.DescribeCompanyNetworkConfigurationInput{
		FleetArn: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing WorkLink Fleet (%s) company network configuration: %s", d.Id(), err)
	}
	if err := d.Set("network", flattenNetworkConfigResponse(companyNetworkConfigurationResp)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network: %s", err)
	}

	identityProviderConfigurationResp, err := conn.DescribeIdentityProviderConfiguration(ctx, &worklink.DescribeIdentityProviderConfigurationInput{
		FleetArn: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing WorkLink Fleet (%s) identity provider configuration: %s", d.Id(), err)
	}
	if err := d.Set("identity_provider", flattenIdentityProviderConfigResponse(identityProviderConfigurationResp)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting identity_provider: %s", err)
	}

	devicePolicyConfigurationResp, err := conn.DescribeDevicePolicyConfiguration(ctx, &worklink.DescribeDevicePolicyConfigurationInput{
		FleetArn: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing WorkLink Fleet (%s) device policy configuration: %s", d.Id(), err)
	}
	d.Set("device_ca_certificate", strings.TrimSpace(aws.ToString(devicePolicyConfigurationResp.DeviceCaCertificate)))

	return diags
}

func resourceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkClient(ctx)

	input := &worklink.UpdateFleetMetadataInput{
		FleetArn:                   aws.String(d.Id()),
		OptimizeForEndUserLocation: aws.Bool(d.Get("optimize_for_end_user_location").(bool)),
	}

	if v, ok := d.GetOk(names.AttrDisplayName); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if d.HasChanges(names.AttrDisplayName, "optimize_for_end_user_location") {
		_, err := conn.UpdateFleetMetadata(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkLink Fleet (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("audit_stream_arn") {
		if err := updateAuditStreamConfiguration(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkLink Fleet (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("network") {
		if err := updateCompanyNetworkConfiguration(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkLink Fleet (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("device_ca_certificate") {
		if err := updateDevicePolicyConfiguration(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkLink Fleet (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("identity_provider") {
		if err := updateIdentityProviderConfiguration(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkLink Fleet (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkLinkClient(ctx)

	input := &worklink.DeleteFleetInput{
		FleetArn: aws.String(d.Id()),
	}

	if _, err := conn.DeleteFleet(ctx, input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting WorkLink Fleet resource share (%s): %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    fleetStateRefresh(ctx, conn, d.Id()),
		Timeout:    15 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for WorkLink Fleet (%s) to become deleted: %s", d.Id(), err)
	}

	return diags
}

func findFleetByARN(ctx context.Context, conn *worklink.Client, arn string) (*worklink.DescribeFleetMetadataOutput, error) {
	input := &worklink.DescribeFleetMetadataInput{
		FleetArn: aws.String(arn),
	}

	output, err := conn.DescribeFleetMetadata(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func fleetStateRefresh(ctx context.Context, conn *worklink.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &worklink.DescribeFleetMetadataOutput{}

		resp, err := conn.DescribeFleetMetadata(ctx, &worklink.DescribeFleetMetadataInput{
			FleetArn: aws.String(arn),
		})
		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return emptyResp, "DELETED", nil
			}
		}

		return resp, string(resp.FleetStatus), nil
	}
}

func updateAuditStreamConfiguration(ctx context.Context, conn *worklink.Client, d *schema.ResourceData) error {
	input := &worklink.UpdateAuditStreamConfigurationInput{
		FleetArn: aws.String(d.Id()),
	}
	if v, ok := d.GetOk("audit_stream_arn"); ok {
		input.AuditStreamArn = aws.String(v.(string))
	} else if d.IsNewResource() {
		return nil
	}

	if _, err := conn.UpdateAuditStreamConfiguration(ctx, input); err != nil {
		return fmt.Errorf("updating Audit Stream Configuration: %w", err)
	}

	return nil
}

func updateCompanyNetworkConfiguration(ctx context.Context, conn *worklink.Client, d *schema.ResourceData) error {
	oldNetwork, newNetwork := d.GetChange("network")
	if len(oldNetwork.([]interface{})) > 0 && len(newNetwork.([]interface{})) == 0 {
		return fmt.Errorf("Company Network Configuration cannot be removed from WorkLink Fleet(%s),"+
			" use 'terraform taint' to recreate the resource if you wish.", d.Id())
	}

	if v, ok := d.GetOk("network"); ok && len(v.([]interface{})) > 0 {
		config := v.([]interface{})[0].(map[string]interface{})
		input := &worklink.UpdateCompanyNetworkConfigurationInput{
			FleetArn:         aws.String(d.Id()),
			SecurityGroupIds: flex.ExpandStringValueSet(config[names.AttrSecurityGroupIDs].(*schema.Set)),
			SubnetIds:        flex.ExpandStringValueSet(config[names.AttrSubnetIDs].(*schema.Set)),
			VpcId:            aws.String(config[names.AttrVPCID].(string)),
		}
		if _, err := conn.UpdateCompanyNetworkConfiguration(ctx, input); err != nil {
			return fmt.Errorf("updating Company Network Configuration: %w", err)
		}
	}
	return nil
}

func updateDevicePolicyConfiguration(ctx context.Context, conn *worklink.Client, d *schema.ResourceData) error {
	input := &worklink.UpdateDevicePolicyConfigurationInput{
		FleetArn: aws.String(d.Id()),
	}
	if v, ok := d.GetOk("device_ca_certificate"); ok {
		input.DeviceCaCertificate = aws.String(v.(string))
	} else if d.IsNewResource() {
		return nil
	}

	if _, err := conn.UpdateDevicePolicyConfiguration(ctx, input); err != nil {
		return fmt.Errorf("updating Device Policy Configuration: %w", err)
	}
	return nil
}

func updateIdentityProviderConfiguration(ctx context.Context, conn *worklink.Client, d *schema.ResourceData) error {
	oldIdentityProvider, newIdentityProvider := d.GetChange("identity_provider")

	if len(oldIdentityProvider.([]interface{})) > 0 && len(newIdentityProvider.([]interface{})) == 0 {
		return fmt.Errorf("Identity Provider Configuration cannot be removed from WorkLink Fleet(%s),"+
			" use 'terraform taint' to recreate the resource if you wish.", d.Id())
	}

	if v, ok := d.GetOk("identity_provider"); ok && len(v.([]interface{})) > 0 {
		config := v.([]interface{})[0].(map[string]interface{})
		input := &worklink.UpdateIdentityProviderConfigurationInput{
			FleetArn:                     aws.String(d.Id()),
			IdentityProviderType:         awstypes.IdentityProviderType(config[names.AttrType].(string)),
			IdentityProviderSamlMetadata: aws.String(config["saml_metadata"].(string)),
		}
		if _, err := conn.UpdateIdentityProviderConfiguration(ctx, input); err != nil {
			return fmt.Errorf("updating Identity Provider Configuration: %w", err)
		}
	}

	return nil
}
