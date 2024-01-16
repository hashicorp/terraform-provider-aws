// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationinsights

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/applicationinsights"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_applicationinsights_application", name="Application")
// @Tags(identifierAttribute="arn")
func ResourceApplication() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cwe_monitor_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"grouping_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(applicationinsights.GroupingType_Values(), false),
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
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ApplicationInsightsConn(ctx)

	input := &applicationinsights.CreateApplicationInput{
		AutoConfigEnabled: aws.Bool(d.Get("auto_config_enabled").(bool)),
		AutoCreate:        aws.Bool(d.Get("auto_create").(bool)),
		CWEMonitorEnabled: aws.Bool(d.Get("cwe_monitor_enabled").(bool)),
		OpsCenterEnabled:  aws.Bool(d.Get("ops_center_enabled").(bool)),
		ResourceGroupName: aws.String(d.Get("resource_group_name").(string)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("grouping_type"); ok {
		input.GroupingType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ops_item_sns_topic_arn"); ok {
		input.OpsItemSNSTopicArn = aws.String(v.(string))
	}

	out, err := conn.CreateApplicationWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ApplicationInsights Application: %s", err)
	}

	d.SetId(aws.StringValue(out.ApplicationInfo.ResourceGroupName))

	if _, err := waitApplicationCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ApplicationInsights Application (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ApplicationInsightsConn(ctx)

	application, err := FindApplicationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ApplicationInsights Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ApplicationInsights Application (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/resource-group/%s", aws.StringValue(application.ResourceGroupName)),
		Service:   "applicationinsights",
	}.String()

	d.Set("arn", arn)
	d.Set("resource_group_name", application.ResourceGroupName)
	d.Set("auto_config_enabled", application.AutoConfigEnabled)
	d.Set("cwe_monitor_enabled", application.CWEMonitorEnabled)
	d.Set("ops_center_enabled", application.OpsCenterEnabled)
	d.Set("ops_item_sns_topic_arn", application.OpsItemSNSTopicArn)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ApplicationInsightsConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
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
			_, n := d.GetChange("ops_item_sns_topic_arn")
			if n != nil {
				input.OpsItemSNSTopicArn = aws.String(n.(string))
			} else {
				input.RemoveSNSTopic = aws.Bool(true)
			}
		}

		log.Printf("[DEBUG] Updating ApplicationInsights Application: %s", d.Id())
		_, err := conn.UpdateApplicationWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ApplicationInsights Application (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ApplicationInsightsConn(ctx)

	input := &applicationinsights.DeleteApplicationInput{
		ResourceGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting ApplicationInsights Application: %s", d.Id())
	_, err := conn.DeleteApplicationWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, applicationinsights.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting ApplicationInsights Application: %s", err)
	}

	if _, err := waitApplicationTerminated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ApplicationInsights Application (%s) delete: %s", d.Id(), err)
	}

	return diags
}
