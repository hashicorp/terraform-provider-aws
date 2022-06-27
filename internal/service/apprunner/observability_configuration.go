package apprunner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceObservabilityConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("observability_configuration_name").(string)

	input := &apprunner.CreateObservabilityConfigurationInput{
		ObservabilityConfigurationName: aws.String(name),
	}

	if v, ok := d.GetOk("trace_configuration"); ok {
		input.TraceConfiguration = expandTraceConfigurations(v.(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateObservabilityConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner Observability Configuration (%s): %w", name, err))
	}

	if output == nil || output.ObservabilityConfiguration == nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner Observability Configuration (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.ObservabilityConfiguration.ObservabilityConfigurationArn))

	if err := WaitObservabilityConfigurationActive(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for App Runner Observability Configuration (%s) creation: %w", d.Id(), err))
	}

	return resourceObservabilityConfigurationRead(ctx, d, meta)
}

func resourceObservabilityConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &apprunner.DescribeObservabilityConfigurationInput{
		ObservabilityConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeObservabilityConfigurationWithContext(ctx, input)

	return nil
}

func resourceObservabilityConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn

	return resourceObservabilityConfigurationRead(ctx, d, meta)
}

func resourceObservabilityConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn

	input := &apprunner.DeleteObservabilityConfigurationInput{
		ObservabilityConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteObservabilityConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting App Runner Observability Configuration (%s): %w", d.Id(), err))
	}

	if err := WaitObservabilityConfigurationInactive(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error waiting for App Runner Observability Configuration (%s) deletion: %w", d.Id(), err))
	}

	return nil
}

func expandTraceConfigurations(tfMap map[string]interface{}) *apprunner.TraceConfiguration {
	if tfMap == nil {
		return nil
	}

	traceConfiguration := &apprunner.TraceConfiguration{}

	if v, ok := tfMap["vendor"].(string); ok {
		traceConfiguration.Vendor = aws.String(v)
	}

	return traceConfiguration
}
