// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_securityhub_organization_configuration")
func ResourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		ReadWithoutTimeout:   resourceOrganizationConfigurationRead,
		UpdateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		// on delete reset to a default configuration.
		// non-default CENTRAL configuration blocks dependent resources from being destroyed.
		DeleteWithoutTimeout: resourceOrganizationConfigurationUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"auto_enable_standards": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.AutoEnableStandards](),
			},
			"organization_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configuration_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.OrganizationConfigurationConfigurationType](),
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceOrganizationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(d.Get("auto_enable").(bool)),
	}

	if v, ok := d.GetOk("auto_enable_standards"); ok {
		input.AutoEnableStandards = types.AutoEnableStandards(v.(string))
	}

	if v, ok := d.GetOk("organization_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OrganizationConfiguration = expandOrganizationConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.UpdateOrganizationConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Organization Configuration (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID)
	}

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	output, err := WaitOrganizationConfigurationEnabled(ctx, conn, d.Timeout(schema.TimeoutDefault))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Organization Configuration %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Organization Configuration (%s): %s", d.Id(), err)
	}

	d.Set("auto_enable", output.AutoEnable)
	d.Set("auto_enable_standards", output.AutoEnableStandards)

	if err := d.Set("organization_configuration", []interface{}{flattenOrganizationConfiguration(output.OrganizationConfiguration)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting organization_configuration: %s", err)
	}

	return diags
}

func findOrganizationConfiguration(ctx context.Context, conn *securityhub.Client) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &securityhub.DescribeOrganizationConfigurationInput{}
		output, err := conn.DescribeOrganizationConfiguration(ctx, input)

		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, "", &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.OrganizationConfiguration == nil {
			return nil, "", tfresource.NewEmptyResultError(input)
		}

		switch output.OrganizationConfiguration.Status {
		case types.OrganizationConfigurationStatusPending:
			return nil, "", nil
		case types.OrganizationConfigurationStatusEnabled, "":
			return output, string(output.OrganizationConfiguration.Status), nil
		default:
			var statusErr error
			if msg := output.OrganizationConfiguration.StatusMessage; msg != nil && len(*msg) > 0 {
				statusErr = fmt.Errorf("StatusMessage: %s", *msg)
			}
			return nil, "", &retry.UnexpectedStateError{
				LastError: statusErr,
				State:     string(output.OrganizationConfiguration.Status),
				ExpectedState: []string{
					string(types.OrganizationConfigurationStatusEnabled),
					string(types.OrganizationConfigurationStatusPending),
				},
			}
		}
	}
}

func WaitOrganizationConfigurationEnabled(ctx context.Context, conn *securityhub.Client, timeout time.Duration) (*securityhub.DescribeOrganizationConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.OrganizationConfigurationStatusPending),
		Target:                    append(enum.Slice(types.OrganizationConfigurationStatusEnabled), ""),
		Refresh:                   findOrganizationConfiguration(ctx, conn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*securityhub.DescribeOrganizationConfigurationOutput); ok {
		return out, err
	}

	return nil, err
}

func expandOrganizationConfiguration(tfMap map[string]interface{}) *types.OrganizationConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.OrganizationConfiguration{}

	if v, ok := tfMap["configuration_type"].(string); ok && len(v) > 0 {
		apiObject.ConfigurationType = types.OrganizationConfigurationConfigurationType(v)
	}

	return apiObject
}

func flattenOrganizationConfiguration(apiObject *types.OrganizationConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"configuration_type": apiObject.ConfigurationType,
		"status":             apiObject.Status,
	}

	return tfMap
}
