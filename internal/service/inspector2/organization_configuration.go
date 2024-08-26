// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_inspector2_organization_configuration")
func ResourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConfigurationCreate,
		ReadWithoutTimeout:   resourceOrganizationConfigurationRead,
		UpdateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		DeleteWithoutTimeout: resourceOrganizationConfigurationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ec2": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"ecr": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"lambda": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"lambda_code": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"max_account_limit_reached": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

const (
	ResNameOrganizationConfiguration = "Organization Configuration"
	orgConfigMutex                   = "f14b54d7-2b10-58c2-9c1b-c48260a4825d"
)

func resourceOrganizationConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceOrganizationConfigurationUpdate(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	out, err := conn.DescribeOrganizationConfiguration(ctx, &inspector2.DescribeOrganizationConfigurationInput{})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 OrganizationConfiguration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionReading, ResNameOrganizationConfiguration, d.Id(), err)
	}

	if err := d.Set("auto_enable", []interface{}{flattenAutoEnable(out.AutoEnable)}); err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionSetting, ResNameOrganizationConfiguration, d.Id(), err)
	}

	d.Set("max_account_limit_reached", out.MaxAccountLimitReached)

	return diags
}

func resourceOrganizationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	update := false

	in := &inspector2.UpdateOrganizationConfigurationInput{}

	if d.HasChanges("auto_enable") {
		in.AutoEnable = expandAutoEnable(d.Get("auto_enable").([]interface{})[0].(map[string]interface{}))
		update = true
	}

	if !update {
		return diags
	}

	conns.GlobalMutexKV.Lock(orgConfigMutex)
	defer conns.GlobalMutexKV.Unlock(orgConfigMutex)

	log.Printf("[DEBUG] Updating Inspector2 Organization Configuration (%s): %#v", d.Id(), in)
	_, err := conn.UpdateOrganizationConfiguration(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionUpdating, ResNameOrganizationConfiguration, d.Id(), err)
	}

	if err := waitOrganizationConfigurationUpdated(ctx, conn, d.Get("auto_enable.0.ec2").(bool), d.Get("auto_enable.0.ecr").(bool), d.Get("auto_enable.0.lambda").(bool), d.Get("auto_enable.0.lambda_code").(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionWaitingForUpdate, ResNameOrganizationConfiguration, d.Id(), err)
	}

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	conns.GlobalMutexKV.Lock(orgConfigMutex)
	defer conns.GlobalMutexKV.Unlock(orgConfigMutex)

	in := &inspector2.UpdateOrganizationConfigurationInput{
		AutoEnable: &types.AutoEnable{
			Ec2:        aws.Bool(false),
			Ecr:        aws.Bool(false),
			Lambda:     aws.Bool(false),
			LambdaCode: aws.Bool(false),
		},
	}

	log.Printf("[DEBUG] Setting Inspector2 Organization Configuration (%s): %#v", d.Id(), in)
	_, err := conn.UpdateOrganizationConfiguration(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionUpdating, ResNameOrganizationConfiguration, d.Id(), err)
	}

	if err := waitOrganizationConfigurationUpdated(ctx, conn, false, false, false, false, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionWaitingForUpdate, ResNameOrganizationConfiguration, d.Id(), err)
	}

	return diags
}

func waitOrganizationConfigurationUpdated(ctx context.Context, conn *inspector2.Client, ec2, ecr, lambda, lambda_code bool, timeout time.Duration) error {
	needle := fmt.Sprintf("%t:%t:%t:%t", ec2, ecr, lambda, lambda_code)

	all := []string{
		fmt.Sprintf("%t:%t:%t:%t", false, false, false, false),
		fmt.Sprintf("%t:%t:%t:%t", false, false, false, true),
		fmt.Sprintf("%t:%t:%t:%t", false, true, false, false),
		fmt.Sprintf("%t:%t:%t:%t", false, true, false, true),
		fmt.Sprintf("%t:%t:%t:%t", false, false, true, false),
		fmt.Sprintf("%t:%t:%t:%t", false, false, true, true),
		fmt.Sprintf("%t:%t:%t:%t", false, true, true, false),
		fmt.Sprintf("%t:%t:%t:%t", false, true, true, true),
		fmt.Sprintf("%t:%t:%t:%t", true, false, false, false),
		fmt.Sprintf("%t:%t:%t:%t", true, false, false, true),
		fmt.Sprintf("%t:%t:%t:%t", true, false, true, false),
		fmt.Sprintf("%t:%t:%t:%t", true, false, true, true),
		fmt.Sprintf("%t:%t:%t:%t", true, true, false, false),
		fmt.Sprintf("%t:%t:%t:%t", true, true, false, true),
		fmt.Sprintf("%t:%t:%t:%t", true, true, true, false),
		fmt.Sprintf("%t:%t:%t:%t", true, true, true, true),
	}

	for i, v := range all {
		if v == needle {
			all = append(all[:i], all[i+1:]...)
			break
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:                   all,
		Target:                    []string{needle},
		Refresh:                   statusOrganizationConfiguration(ctx, conn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
		MinTimeout:                time.Second * 5,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func statusOrganizationConfiguration(ctx context.Context, conn *inspector2.Client) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.DescribeOrganizationConfiguration(ctx, &inspector2.DescribeOrganizationConfigurationInput{})
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, fmt.Sprintf("%t:%t:%t:%t", aws.ToBool(out.AutoEnable.Ec2), aws.ToBool(out.AutoEnable.Ecr), aws.ToBool(out.AutoEnable.Lambda), aws.ToBool(out.AutoEnable.LambdaCode)), nil
	}
}

func flattenAutoEnable(apiObject *types.AutoEnable) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Ec2; v != nil {
		m["ec2"] = aws.ToBool(v)
	}

	if v := apiObject.Ecr; v != nil {
		m["ecr"] = aws.ToBool(v)
	}

	if v := apiObject.Lambda; v != nil {
		m["lambda"] = aws.ToBool(v)
	}

	if v := apiObject.LambdaCode; v != nil {
		m["lambda_code"] = aws.ToBool(v)
	}

	return m
}

func expandAutoEnable(tfMap map[string]interface{}) *types.AutoEnable {
	if tfMap == nil {
		return nil
	}

	a := &types.AutoEnable{}

	if v, ok := tfMap["ec2"].(bool); ok {
		a.Ec2 = aws.Bool(v)
	}

	if v, ok := tfMap["ecr"].(bool); ok {
		a.Ecr = aws.Bool(v)
	}

	if v, ok := tfMap["lambda"].(bool); ok {
		a.Lambda = aws.Bool(v)
	}

	if v, ok := tfMap["lambda_code"].(bool); ok {
		a.LambdaCode = aws.Bool(v)
	}

	return a
}
