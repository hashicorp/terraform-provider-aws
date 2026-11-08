// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kinesisanalyticsv2_application_maintenance_configuration", name="Application Maintenance Configuration")
func resourceApplicationMaintenanceConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationMaintenanceConfigurationPut,
		ReadWithoutTimeout:   resourceApplicationMaintenanceConfigurationRead,
		UpdateWithoutTimeout: resourceApplicationMaintenanceConfigurationPut,
		DeleteWithoutTimeout: resourceApplicationMaintenanceConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationMaintenanceConfigurationImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"application_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},
			"application_maintenance_window_start_time": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(5, 5),
					validation.StringMatch(regexache.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`), "must be in HH:MM format (00:00-23:59)"),
				),
			},
			"original_maintenance_window_start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceApplicationMaintenanceConfigurationPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	applicationName := d.Get("application_name").(string)

	// Read application details to verify state and store original maintenance window
	// From flink docs, this UpdateApplicationMaintenanceConfiguration can only be invoked when the state of the application is either on READY or RUNNING
	// See the docs here: https://docs.aws.amazon.com/managed-flink/latest/apiv2/API_UpdateApplicationMaintenanceConfiguration.html
	application, err := findApplicationDetailByName(ctx, conn, applicationName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics v2 Application (%s): %s", applicationName, err)
	}

	// Verify application is in READY or RUNNING state
	if status := application.ApplicationStatus; status != awstypes.ApplicationStatusReady && status != awstypes.ApplicationStatusRunning {
		return sdkdiag.AppendErrorf(diags, "updating Kinesis Analytics v2 Application Maintenance Configuration (%s): application must be in READY or RUNNING state, current state: %s", applicationName, status)
	}

	// Store original maintenance window on create
	// We are storing this because by default, AWS Managed Flink provides default value for maintenance window
	// As stated by AWS docs, we can't opt out for this maintenance window
	// We need to store the default value so that when we're deleting this resource, the maintenance window will be put back to its default value
	if d.IsNewResource() {
		if application.ApplicationMaintenanceConfigurationDescription != nil {
			d.Set("original_maintenance_window_start_time", application.ApplicationMaintenanceConfigurationDescription.ApplicationMaintenanceWindowStartTime)
		}
	}

	input := &kinesisanalyticsv2.UpdateApplicationMaintenanceConfigurationInput{
		ApplicationName: aws.String(applicationName),
		ApplicationMaintenanceConfigurationUpdate: &awstypes.ApplicationMaintenanceConfigurationUpdate{
			ApplicationMaintenanceWindowStartTimeUpdate: aws.String(d.Get("application_maintenance_window_start_time").(string)),
		},
	}

	_, err = conn.UpdateApplicationMaintenanceConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Kinesis Analytics v2 Application Maintenance Configuration (%s): %s", applicationName, err)
	}

	d.SetId(applicationName)

	return append(diags, resourceApplicationMaintenanceConfigurationRead(ctx, d, meta)...)
}

func resourceApplicationMaintenanceConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	applicationName := d.Id()

	application, err := findApplicationDetailByName(ctx, conn, applicationName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics v2 Application Maintenance Configuration (%s): %s", d.Id(), err)
	}

	d.Set("application_name", application.ApplicationName)

	if application.ApplicationMaintenanceConfigurationDescription != nil {
		d.Set("application_maintenance_window_start_time", application.ApplicationMaintenanceConfigurationDescription.ApplicationMaintenanceWindowStartTime)
	}

	return diags
}

func resourceApplicationMaintenanceConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	applicationName := d.Id()
	originalTime := d.Get("original_maintenance_window_start_time").(string)

	// Verify application is in READY or RUNNING state
	application, err := findApplicationDetailByName(ctx, conn, applicationName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics v2 Application (%s): %s", applicationName, err)
	}

	if status := application.ApplicationStatus; status != awstypes.ApplicationStatusReady && status != awstypes.ApplicationStatusRunning {
		return sdkdiag.AppendErrorf(diags, "restoring Kinesis Analytics v2 Application Maintenance Configuration (%s): application must be in READY or RUNNING state, current state: %s", applicationName, status)
	}

	input := &kinesisanalyticsv2.UpdateApplicationMaintenanceConfigurationInput{
		ApplicationName: aws.String(applicationName),
		ApplicationMaintenanceConfigurationUpdate: &awstypes.ApplicationMaintenanceConfigurationUpdate{
			ApplicationMaintenanceWindowStartTimeUpdate: aws.String(originalTime),
		},
	}

	_, err = conn.UpdateApplicationMaintenanceConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "restoring Kinesis Analytics v2 Application Maintenance Configuration (%s): %s", applicationName, err)
	}

	return diags
}

func resourceApplicationMaintenanceConfigurationImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	arn, err := arn.Parse(d.Id())
	if err != nil {
		return nil, fmt.Errorf("parsing ARN (%s): %w", d.Id(), err)
	}

	// application/<name>
	parts := strings.Split(arn.Resource, "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("unexpected ARN format: %q", d.Id())
	}

	applicationName := parts[1]

	application, err := findApplicationDetailByName(ctx, conn, applicationName)
	if err != nil {
		return nil, err
	}

	d.Set(names.AttrName, applicationName)
	d.Set("application_name", applicationName)

	if application.ApplicationMaintenanceConfigurationDescription != nil {
		// Store current maintenance window as original for future restore on delete
		d.Set("original_maintenance_window_start_time", application.ApplicationMaintenanceConfigurationDescription.ApplicationMaintenanceWindowStartTime)
		d.Set("application_maintenance_window_start_time", application.ApplicationMaintenanceConfigurationDescription.ApplicationMaintenanceWindowStartTime)
	}

	return []*schema.ResourceData{d}, nil
}
