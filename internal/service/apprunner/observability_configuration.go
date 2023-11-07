// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apprunner_observability_configuration", name="Observability Configuration")
// @Tags(identifierAttribute="arn")
func ResourceObservabilityConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceObservabilityConfigurationCreate,
		ReadWithoutTimeout:   resourceObservabilityConfigurationRead,
		UpdateWithoutTimeout: resourceObservabilityConfigurationUpdate,
		DeleteWithoutTimeout: resourceObservabilityConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
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
			"latest": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"trace_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vendor": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(flattenTracingVendorValues(types.TracingVendor("").Values()), false),
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceObservabilityConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("creating App Runner Observability Configuration (%s): %s", name, err)
	}

	if output == nil || output.ObservabilityConfiguration == nil {
		return diag.Errorf("creating App Runner Observability Configuration (%s): empty output", name)
	}

	d.SetId(aws.ToString(output.ObservabilityConfiguration.ObservabilityConfigurationArn))

	if err := WaitObservabilityConfigurationActive(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for App Runner Observability Configuration (%s) creation: %s", d.Id(), err)
	}

	return resourceObservabilityConfigurationRead(ctx, d, meta)
}

func resourceObservabilityConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	input := &apprunner.DescribeObservabilityConfigurationInput{
		ObservabilityConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeObservabilityConfiguration(ctx, input)

	if !d.IsNewResource() && errs.IsA[*types.ResourceNotFoundException](err) {
		log.Printf("[WARN] App Runner Observability Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading App Runner Observability Configuration (%s): %s", d.Id(), err)
	}

	if output == nil || output.ObservabilityConfiguration == nil {
		return diag.Errorf("reading App Runner Observability Configuration (%s): empty output", d.Id())
	}

	if string(output.ObservabilityConfiguration.Status) == ObservabilityConfigurationStatusInactive {
		if d.IsNewResource() {
			return diag.Errorf("reading App Runner Observability Configuration (%s): %s after creation", d.Id(), string(output.ObservabilityConfiguration.Status))
		}
		log.Printf("[WARN] App Runner Observability Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	config := output.ObservabilityConfiguration
	arn := aws.ToString(config.ObservabilityConfigurationArn)

	d.Set("arn", arn)
	d.Set("observability_configuration_name", config.ObservabilityConfigurationName)
	d.Set("observability_configuration_revision", config.ObservabilityConfigurationRevision)
	d.Set("latest", config.Latest)
	d.Set("status", config.Status)

	if err := d.Set("trace_configuration", flattenTraceConfiguration(config.TraceConfiguration)); err != nil {
		return diag.Errorf("setting trace_configuration: %s", err)
	}

	return nil
}

func resourceObservabilityConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceObservabilityConfigurationRead(ctx, d, meta)
}

func resourceObservabilityConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	input := &apprunner.DeleteObservabilityConfigurationInput{
		ObservabilityConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteObservabilityConfiguration(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting App Runner Observability Configuration (%s): %s", d.Id(), err)
	}

	if err := WaitObservabilityConfigurationInactive(ctx, conn, d.Id()); err != nil {
		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil
		}
		return diag.Errorf("waiting for App Runner Observability Configuration (%s) deletion: %s", d.Id(), err)
	}

	return nil
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

func flattenTracingVendorValues(t []types.TracingVendor) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}
