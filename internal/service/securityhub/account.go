// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_securityhub_account")
func ResourceAccount() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_enable_controls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"control_finding_generator": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      securityhub.ControlFindingGeneratorSecurityControl,
				ValidateFunc: validation.StringInSlice(securityhub.ControlFindingGenerator_Values(), false),
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
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	input := &securityhub.EnableSecurityHubInput{
		EnableDefaultStandards: aws.Bool(d.Get("enable_default_standards").(bool)),
	}

	if v, ok := d.GetOk("control_finding_generator"); ok {
		input.ControlFindingGenerator = aws.String(v.(string))
	}

	_, err := conn.EnableSecurityHubWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Account: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	// auto_enable_controls has to be done from the update API
	return append(diags, resourceAccountUpdate(ctx, d, meta)...)
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	_, err := FindStandardsSubscriptions(ctx, conn, &securityhub.GetEnabledStandardsInput{})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Standards Subscriptions %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Standards Subscriptions: %s", err)
	}

	hub, err := conn.DescribeHubWithContext(ctx, &securityhub.DescribeHubInput{
		HubArn: aws.String(accountHubARN(meta.(*conns.AWSClient))),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Account (%s): %s", d.Id(), err)
	}

	d.Set("arn", hub.HubArn)
	d.Set("auto_enable_controls", hub.AutoEnableControls)
	d.Set("control_finding_generator", hub.ControlFindingGenerator)
	// enable_default_standards is never returned
	d.Set("enable_default_standards", d.Get("enable_default_standards"))

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	input := &securityhub.UpdateSecurityHubConfigurationInput{
		ControlFindingGenerator: aws.String(d.Get("control_finding_generator").(string)),
		AutoEnableControls:      aws.Bool(d.Get("auto_enable_controls").(bool)),
	}

	_, err := conn.UpdateSecurityHubConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Account (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Account: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, adminAccountNotFoundTimeout, func() (interface{}, error) {
		return conn.DisableSecurityHubWithContext(ctx, &securityhub.DisableSecurityHubInput{})
	}, securityhub.ErrCodeInvalidInputException, "Cannot disable Security Hub on the Security Hub administrator")

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Account (%s): %s", d.Id(), err)
	}

	return diags
}

// Security Hub ARN: https://docs.aws.amazon.com/service-authorization/latest/reference/list_awssecurityhub.html#awssecurityhub-resources-for-iam-policies
func accountHubARN(conn *conns.AWSClient) string {
	return fmt.Sprintf("arn:%s:securityhub:%s:%s:hub/default", conn.Partition, conn.Region, conn.AccountID)
}
