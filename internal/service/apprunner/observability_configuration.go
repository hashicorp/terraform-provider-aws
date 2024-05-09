// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
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

// @SDKResource("aws_apprunner_observability_configuration", name="Observability Configuration")
// @Tags(identifierAttribute="arn")
func resourceObservabilityConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceObservabilityConfigurationCreate,
		ReadWithoutTimeout:   resourceObservabilityConfigurationRead,
		UpdateWithoutTimeout: resourceObservabilityConfigurationUpdate,
		DeleteWithoutTimeout: resourceObservabilityConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"observability_configuration_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"observability_configuration_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"trace_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vendor": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.TracingVendor](),
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceObservabilityConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	name := d.Get("observability_configuration_name").(string)
	input := &apprunner.CreateObservabilityConfigurationInput{
		ObservabilityConfigurationName: aws.String(name),
		Tags:                           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("trace_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TraceConfiguration = expandTraceConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateObservabilityConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Runner Observability Configuration (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ObservabilityConfiguration.ObservabilityConfigurationArn))

	if _, err := waitObservabilityConfigurationCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner Observability Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceObservabilityConfigurationRead(ctx, d, meta)...)
}

func resourceObservabilityConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	config, err := findObservabilityConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Runner Observability Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Runner Observability Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, config.ObservabilityConfigurationArn)
	d.Set("latest", config.Latest)
	d.Set("observability_configuration_name", config.ObservabilityConfigurationName)
	d.Set("observability_configuration_revision", config.ObservabilityConfigurationRevision)
	d.Set(names.AttrStatus, config.Status)
	if err := d.Set("trace_configuration", flattenTraceConfiguration(config.TraceConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting trace_configuration: %s", err)
	}

	return diags
}

func resourceObservabilityConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceObservabilityConfigurationRead(ctx, d, meta)
}

func resourceObservabilityConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	log.Printf("[INFO] Deleting App Runner Observability Configuration: %s", d.Id())
	_, err := conn.DeleteObservabilityConfiguration(ctx, &apprunner.DeleteObservabilityConfigurationInput{
		ObservabilityConfigurationArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Runner Observability Configuration (%s): %s", d.Id(), err)
	}

	if _, err := waitObservabilityConfigurationDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner Observability Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findObservabilityConfigurationByARN(ctx context.Context, conn *apprunner.Client, arn string) (*types.ObservabilityConfiguration, error) {
	input := &apprunner.DescribeObservabilityConfigurationInput{
		ObservabilityConfigurationArn: aws.String(arn),
	}

	output, err := conn.DescribeObservabilityConfiguration(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ObservabilityConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.ObservabilityConfiguration.Status; status == types.ObservabilityConfigurationStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output.ObservabilityConfiguration, nil
}

func statusObservabilityConfiguration(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findObservabilityConfigurationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitObservabilityConfigurationCreated(ctx context.Context, conn *apprunner.Client, arn string) (*types.ObservabilityConfiguration, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(types.ObservabilityConfigurationStatusActive),
		Refresh: statusObservabilityConfiguration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ObservabilityConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitObservabilityConfigurationDeleted(ctx context.Context, conn *apprunner.Client, arn string) (*types.ObservabilityConfiguration, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ObservabilityConfigurationStatusActive),
		Target:  []string{},
		Refresh: statusObservabilityConfiguration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ObservabilityConfiguration); ok {
		return output, err
	}

	return nil, err
}

func expandTraceConfiguration(l []interface{}) *types.TraceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &types.TraceConfiguration{}

	if v, ok := m["vendor"].(string); ok && v != "" {
		configuration.Vendor = types.TracingVendor(v)
	}

	return configuration
}

func flattenTraceConfiguration(traceConfiguration *types.TraceConfiguration) []interface{} {
	if traceConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"vendor": string(traceConfiguration.Vendor),
	}

	return []interface{}{m}
}
