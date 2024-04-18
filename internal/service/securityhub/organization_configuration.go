// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"errors"
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

// @SDKResource("aws_securityhub_organization_configuration", name="Organization Configuration")
func resourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		ReadWithoutTimeout:   resourceOrganizationConfigurationRead,
		UpdateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		DeleteWithoutTimeout: resourceOrganizationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(180 * time.Second),
			Update: schema.DefaultTimeout(180 * time.Second),
			Delete: schema.DefaultTimeout(180 * time.Second),
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
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				DefaultFunc: func() (interface{}, error) { return defaultOrganizationConfiguration, nil },
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configuration_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.OrganizationConfigurationConfigurationType](),
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

	timeout := d.Timeout(schema.TimeoutUpdate)
	if d.IsNewResource() {
		timeout = d.Timeout(schema.TimeoutCreate)
	}

	// e.g. "DataUnavailableException: Central configuration couldn't be enabled because data from organization o-ira6i4k380 is still syncing. Retry later."
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, timeout, func() (interface{}, error) {
		return conn.UpdateOrganizationConfiguration(ctx, input)
	}, errCodeDataUnavailableException, "Retry later")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Organization Configuration (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID)
	}

	configurationType := types.OrganizationConfigurationConfigurationTypeLocal
	if input.OrganizationConfiguration != nil {
		configurationType = input.OrganizationConfiguration.ConfigurationType
	}

	if configurationType == types.OrganizationConfigurationConfigurationTypeCentral {
		if _, err := waitOrganizationConfigurationEnabled(ctx, conn, timeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Organization Configuration (%s) enable: %s", d.Id(), err)
		}
	}

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	output, err := findOrganizationConfiguration(ctx, conn)

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

// resourceOrganizationConfigurationDelete destroys the organizations configuration resource by updating it to a disabled configuration.
// If orgnanization configuration is of type central, then dependent resources (i.e finding_aggregator, delegated_admin) cannot be removed from AWS.
// Updating the organization configuration on destroy is necessary to allow dependent resources to be able to be cleaned up.
func resourceOrganizationConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.UpdateOrganizationConfigurationInput{
		AutoEnable:          aws.Bool(false),
		AutoEnableStandards: types.AutoEnableStandardsNone,
		OrganizationConfiguration: &types.OrganizationConfiguration{
			ConfigurationType: types.OrganizationConfigurationConfigurationTypeLocal,
		},
	}

	_, err := conn.UpdateOrganizationConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Organization Configuration (%s): %s", d.Id(), err)
	}

	configurationType := types.OrganizationConfigurationConfigurationTypeLocal
	if v, ok := d.GetOk("organization_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if organizationConfiguration := expandOrganizationConfiguration(v.([]interface{})[0].(map[string]interface{})); organizationConfiguration != nil {
			configurationType = organizationConfiguration.ConfigurationType
		}
	}

	if configurationType == types.OrganizationConfigurationConfigurationTypeCentral {
		if _, err := waitOrganizationConfigurationEnabled(ctx, conn, d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Organization Configuration (%s) delete: %s", d.Id(), err)
		}
	}

	return diags
}

func findOrganizationConfiguration(ctx context.Context, conn *securityhub.Client) (*securityhub.DescribeOrganizationConfigurationOutput, error) {
	input := &securityhub.DescribeOrganizationConfigurationInput{}

	output, err := conn.DescribeOrganizationConfiguration(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OrganizationConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusOrganizationConfiguration(ctx context.Context, conn *securityhub.Client) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findOrganizationConfiguration(ctx, conn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.OrganizationConfiguration.Status), nil
	}
}

func waitOrganizationConfigurationEnabled(ctx context.Context, conn *securityhub.Client, timeout time.Duration) (*securityhub.DescribeOrganizationConfigurationOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.OrganizationConfigurationStatusPending),
		Target:                    enum.Slice(types.OrganizationConfigurationStatusEnabled),
		Refresh:                   statusOrganizationConfiguration(ctx, conn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*securityhub.DescribeOrganizationConfigurationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.OrganizationConfiguration.StatusMessage)))

		return output, err
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
	}

	return tfMap
}

var defaultOrganizationConfiguration = []interface{}{
	map[string]interface{}{
		"configuration_type": types.OrganizationConfigurationConfigurationTypeLocal,
	},
}
