// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
							ValidateFunc: validation.StringInSlice(apprunner.TracingVendor_Values(), false),
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
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	name := d.Get("observability_configuration_name").(string)
	input := &apprunner.CreateObservabilityConfigurationInput{
		ObservabilityConfigurationName: aws.String(name),
		Tags:                           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("trace_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TraceConfiguration = expandTraceConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateObservabilityConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating App Runner Observability Configuration (%s): %s", name, err)
	}

	if output == nil || output.ObservabilityConfiguration == nil {
		return diag.Errorf("creating App Runner Observability Configuration (%s): empty output", name)
	}

	d.SetId(aws.StringValue(output.ObservabilityConfiguration.ObservabilityConfigurationArn))

	if err := WaitObservabilityConfigurationActive(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for App Runner Observability Configuration (%s) creation: %s", d.Id(), err)
	}

	return resourceObservabilityConfigurationRead(ctx, d, meta)
}

func resourceObservabilityConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	input := &apprunner.DescribeObservabilityConfigurationInput{
		ObservabilityConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeObservabilityConfigurationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
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

	if aws.StringValue(output.ObservabilityConfiguration.Status) == ObservabilityConfigurationStatusInactive {
		if d.IsNewResource() {
			return diag.Errorf("reading App Runner Observability Configuration (%s): %s after creation", d.Id(), aws.StringValue(output.ObservabilityConfiguration.Status))
		}
		log.Printf("[WARN] App Runner Observability Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	config := output.ObservabilityConfiguration
	arn := aws.StringValue(config.ObservabilityConfigurationArn)

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
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	input := &apprunner.DeleteObservabilityConfigurationInput{
		ObservabilityConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteObservabilityConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting App Runner Observability Configuration (%s): %s", d.Id(), err)
	}

	if err := WaitObservabilityConfigurationInactive(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("waiting for App Runner Observability Configuration (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func expandTraceConfiguration(l []interface{}) *apprunner.TraceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &apprunner.TraceConfiguration{}

	if v, ok := m["vendor"].(string); ok && v != "" {
		configuration.Vendor = aws.String(v)
	}

	return configuration
}

func flattenTraceConfiguration(traceConfiguration *apprunner.TraceConfiguration) []interface{} {
	if traceConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"vendor": aws.StringValue(traceConfiguration.Vendor),
	}

	return []interface{}{m}
}
