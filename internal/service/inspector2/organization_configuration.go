// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_inspector2_organization_configuration", name="Organization Configuration")
func resourceOrganizationConfiguration() *schema.Resource {
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
	orgConfigMutex = "f14b54d7-2b10-58c2-9c1b-c48260a4825d"
)

func resourceOrganizationConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(meta.(*conns.AWSClient).AccountID(ctx))

	return append(diags, resourceOrganizationConfigurationUpdate(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	output, err := findOrganizationConfiguration(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Organization Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector2 Organization Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("auto_enable", []any{flattenAutoEnable(output.AutoEnable)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting auto_enable: %s", err)
	}
	d.Set("max_account_limit_reached", output.MaxAccountLimitReached)

	return diags
}

func resourceOrganizationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	autoEnable := expandAutoEnable(d.Get("auto_enable").([]any)[0].(map[string]any))
	input := &inspector2.UpdateOrganizationConfigurationInput{
		AutoEnable: autoEnable,
	}

	conns.GlobalMutexKV.Lock(orgConfigMutex)
	defer conns.GlobalMutexKV.Unlock(orgConfigMutex)

	_, err := conn.UpdateOrganizationConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Inspector2 Organization Configuration (%s): %s", d.Id(), err)
	}

	timeout := d.Timeout(schema.TimeoutUpdate)
	if d.IsNewResource() {
		timeout = d.Timeout(schema.TimeoutCreate)
	}

	if _, err := waitOrganizationConfigurationUpdated(ctx, conn, autoEnable, timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Inspector2 Organization Configuration (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	conns.GlobalMutexKV.Lock(orgConfigMutex)
	defer conns.GlobalMutexKV.Unlock(orgConfigMutex)

	log.Printf("[DEBUG] Deleting Inspector2 Organization Configuration: %s", d.Id())
	autoEnable := &awstypes.AutoEnable{
		Ec2:        aws.Bool(false),
		Ecr:        aws.Bool(false),
		Lambda:     aws.Bool(false),
		LambdaCode: aws.Bool(false),
	}
	_, err := conn.UpdateOrganizationConfiguration(ctx, &inspector2.UpdateOrganizationConfigurationInput{
		AutoEnable: autoEnable,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Inspector2 Organization Configuration (%s): %s", d.Id(), err)
	}

	if _, err := waitOrganizationConfigurationUpdated(ctx, conn, autoEnable, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Inspector2 Organization Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findOrganizationConfiguration(ctx context.Context, conn *inspector2.Client) (*inspector2.DescribeOrganizationConfigurationOutput, error) {
	input := &inspector2.DescribeOrganizationConfigurationInput{}
	output, err := conn.DescribeOrganizationConfiguration(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Invoking account does not have access to describe the organization configuration") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AutoEnable == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func waitOrganizationConfigurationUpdated(ctx context.Context, conn *inspector2.Client, target *awstypes.AutoEnable, timeout time.Duration) (*inspector2.DescribeOrganizationConfigurationOutput, error) { //nolint:unparam
	var output *inspector2.DescribeOrganizationConfigurationOutput

	_, err := tfresource.RetryUntilEqual(ctx, timeout, true, func() (bool, error) {
		var err error
		output, err = findOrganizationConfiguration(ctx, conn)

		if err != nil {
			return false, err
		}

		equal := aws.ToBool(output.AutoEnable.Ec2) == aws.ToBool(target.Ec2)
		equal = equal && aws.ToBool(output.AutoEnable.Ecr) == aws.ToBool(target.Ecr)
		equal = equal && aws.ToBool(output.AutoEnable.Lambda) == aws.ToBool(target.Lambda)
		equal = equal && aws.ToBool(output.AutoEnable.LambdaCode) == aws.ToBool(target.LambdaCode)

		return equal, nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func flattenAutoEnable(apiObject *awstypes.AutoEnable) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Ec2; v != nil {
		tfMap["ec2"] = aws.ToBool(v)
	}

	if v := apiObject.Ecr; v != nil {
		tfMap["ecr"] = aws.ToBool(v)
	}

	if v := apiObject.Lambda; v != nil {
		tfMap["lambda"] = aws.ToBool(v)
	}

	if v := apiObject.LambdaCode; v != nil {
		tfMap["lambda_code"] = aws.ToBool(v)
	}

	return tfMap
}

func expandAutoEnable(tfMap map[string]any) *awstypes.AutoEnable {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AutoEnable{}

	if v, ok := tfMap["ec2"].(bool); ok {
		apiObject.Ec2 = aws.Bool(v)
	}

	if v, ok := tfMap["ecr"].(bool); ok {
		apiObject.Ecr = aws.Bool(v)
	}

	if v, ok := tfMap["lambda"].(bool); ok {
		apiObject.Lambda = aws.Bool(v)
	}

	if v, ok := tfMap["lambda_code"].(bool); ok {
		apiObject.LambdaCode = aws.Bool(v)
	}

	return apiObject
}
