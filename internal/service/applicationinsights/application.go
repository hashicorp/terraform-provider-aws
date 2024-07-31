// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationinsights

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationinsights"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationinsights/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_applicationinsights_application", name="Application")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/applicationinsights/types;types.ApplicationInfo")
func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationCreate,
		ReadWithoutTimeout:   resourceApplicationRead,
		UpdateWithoutTimeout: resourceApplicationUpdate,
		DeleteWithoutTimeout: resourceApplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"auto_config_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"auto_create": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cwe_monitor_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"grouping_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.GroupingType](),
			},
			"ops_center_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ops_item_sns_topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"resource_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ApplicationInsightsClient(ctx)

	input := &applicationinsights.CreateApplicationInput{
		AutoConfigEnabled: aws.Bool(d.Get("auto_config_enabled").(bool)),
		AutoCreate:        aws.Bool(d.Get("auto_create").(bool)),
		CWEMonitorEnabled: aws.Bool(d.Get("cwe_monitor_enabled").(bool)),
		OpsCenterEnabled:  aws.Bool(d.Get("ops_center_enabled").(bool)),
		ResourceGroupName: aws.String(d.Get("resource_group_name").(string)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("grouping_type"); ok {
		input.GroupingType = awstypes.GroupingType(v.(string))
	}

	if v, ok := d.GetOk("ops_item_sns_topic_arn"); ok {
		input.OpsItemSNSTopicArn = aws.String(v.(string))
	}

	output, err := conn.CreateApplication(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ApplicationInsights Application: %s", err)
	}

	d.SetId(aws.ToString(output.ApplicationInfo.ResourceGroupName))

	if _, err := waitApplicationCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ApplicationInsights Application (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ApplicationInsightsClient(ctx)

	application, err := findApplicationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ApplicationInsights Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ApplicationInsights Application (%s): %s", d.Id(), err)
	}

	rgName := aws.ToString(application.ResourceGroupName)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "applicationinsights",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "application/resource-group/" + rgName,
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("auto_config_enabled", application.AutoConfigEnabled)
	d.Set("cwe_monitor_enabled", application.CWEMonitorEnabled)
	d.Set("ops_center_enabled", application.OpsCenterEnabled)
	d.Set("ops_item_sns_topic_arn", application.OpsItemSNSTopicArn)
	d.Set("resource_group_name", rgName)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ApplicationInsightsClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &applicationinsights.UpdateApplicationInput{
			ResourceGroupName: aws.String(d.Id()),
		}

		if d.HasChange("auto_config_enabled") {
			input.AutoConfigEnabled = aws.Bool(d.Get("auto_config_enabled").(bool))
		}

		if d.HasChange("cwe_monitor_enabled") {
			input.CWEMonitorEnabled = aws.Bool(d.Get("cwe_monitor_enabled").(bool))
		}

		if d.HasChange("ops_center_enabled") {
			input.OpsCenterEnabled = aws.Bool(d.Get("ops_center_enabled").(bool))
		}

		if d.HasChange("ops_item_sns_topic_arn") {
			if _, n := d.GetChange("ops_item_sns_topic_arn"); n != nil {
				input.OpsItemSNSTopicArn = aws.String(n.(string))
			} else {
				input.RemoveSNSTopic = aws.Bool(true)
			}
		}

		_, err := conn.UpdateApplication(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ApplicationInsights Application (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ApplicationInsightsClient(ctx)

	log.Printf("[DEBUG] Deleting ApplicationInsights Application: %s", d.Id())
	_, err := conn.DeleteApplication(ctx, &applicationinsights.DeleteApplicationInput{
		ResourceGroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ApplicationInsights Application: %s", err)
	}

	if _, err := waitApplicationTerminated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ApplicationInsights Application (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findApplicationByName(ctx context.Context, conn *applicationinsights.Client, name string) (*awstypes.ApplicationInfo, error) {
	input := applicationinsights.DescribeApplicationInput{
		ResourceGroupName: aws.String(name),
	}

	output, err := conn.DescribeApplication(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApplicationInfo, nil
}

func statusApplication(ctx context.Context, conn *applicationinsights.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findApplicationByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.LifeCycle), nil
	}
}

func waitApplicationCreated(ctx context.Context, conn *applicationinsights.Client, name string) (*awstypes.ApplicationInfo, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{"CREATING"},
		Target:  []string{"NOT_CONFIGURED", "ACTIVE"},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApplicationInfo); ok {
		return output, err
	}

	return nil, err
}

func waitApplicationTerminated(ctx context.Context, conn *applicationinsights.Client, name string) (*awstypes.ApplicationInfo, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{"ACTIVE", "NOT_CONFIGURED", "DELETING"},
		Target:  []string{},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApplicationInfo); ok {
		return output, err
	}

	return nil, err
}
