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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_account", name="Account")
func resourceAccount() *schema.Resource {
	resourceV0 := &schema.Resource{Schema: map[string]*schema.Schema{}}

	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountCreate,
		ReadWithoutTimeout:   resourceAccountRead,
		UpdateWithoutTimeout: resourceAccountUpdate,
		DeleteWithoutTimeout: resourceAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type: resourceV0.CoreConfigSchema().ImpliedType(),
				Upgrade: func(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
					if v, ok := rawState["enable_default_standards"]; !ok || v == nil {
						rawState["enable_default_standards"] = "true"
					}

					return rawState, nil
				},
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_enable_controls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"control_finding_generator": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.ControlFindingGenerator](),
			},
			"enable_default_standards": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
		},
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	inputC := &securityhub.EnableSecurityHubInput{
		EnableDefaultStandards: aws.Bool(d.Get("enable_default_standards").(bool)),
	}

	if v, ok := d.GetOk("control_finding_generator"); ok {
		inputC.ControlFindingGenerator = types.ControlFindingGenerator(v.(string))
	}

	_, err := conn.EnableSecurityHub(ctx, inputC)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Account: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	autoEnableControls := d.Get("auto_enable_controls").(bool)
	inputU := &securityhub.UpdateSecurityHubConfigurationInput{
		AutoEnableControls: aws.Bool(autoEnableControls),
	}

	_, err = conn.UpdateSecurityHubConfiguration(ctx, inputU)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Account (%s): %s", d.Id(), err)
	}

	arn := accountHubARN(meta.(*conns.AWSClient))
	const (
		timeout = 1 * time.Minute
	)
	_, err = tfresource.RetryUntilEqual(ctx, timeout, autoEnableControls, func() (bool, error) {
		output, err := findHubByARN(ctx, conn, arn)

		if err != nil {
			return false, err
		}

		return aws.ToBool(output.AutoEnableControls), nil
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Account (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	arn := accountHubARN(meta.(*conns.AWSClient))
	output, err := findHubByARN(ctx, conn, arn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Account %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Account (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.HubArn)
	d.Set("auto_enable_controls", output.AutoEnableControls)
	d.Set("control_finding_generator", output.ControlFindingGenerator)
	// enable_default_standards is never returned
	d.Set("enable_default_standards", d.Get("enable_default_standards"))

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.UpdateSecurityHubConfigurationInput{
		AutoEnableControls: aws.Bool(d.Get("auto_enable_controls").(bool)),
	}

	if d.HasChange("control_finding_generator") {
		input.ControlFindingGenerator = types.ControlFindingGenerator(d.Get("control_finding_generator").(string))
	}

	_, err := conn.UpdateSecurityHubConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Account (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Account: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, adminAccountDeletedTimeout, func() (interface{}, error) {
		return conn.DisableSecurityHub(ctx, &securityhub.DisableSecurityHubInput{})
	}, errCodeInvalidInputException, "Cannot disable Security Hub on the Security Hub administrator")

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Account (%s): %s", d.Id(), err)
	}

	return diags
}

func findHubByARN(ctx context.Context, conn *securityhub.Client, arn string) (*securityhub.DescribeHubOutput, error) {
	input := &securityhub.DescribeHubInput{
		HubArn: aws.String(arn),
	}

	return findHub(ctx, conn, input)
}

func findHub(ctx context.Context, conn *securityhub.Client, input *securityhub.DescribeHubInput) (*securityhub.DescribeHubOutput, error) {
	output, err := conn.DescribeHub(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
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

// Security Hub ARN: https://docs.aws.amazon.com/service-authorization/latest/reference/list_awssecurityhub.html#awssecurityhub-resources-for-iam-policies
func accountHubARN(meta *conns.AWSClient) string {
	return fmt.Sprintf("arn:%s:securityhub:%s:%s:hub/default", meta.Partition, meta.Region, meta.AccountID)
}
