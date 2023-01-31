package apprunner

import (
	"context"
	"fmt"
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceObservabilityConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("observability_configuration_name").(string)

	input := &apprunner.CreateObservabilityConfigurationInput{
		ObservabilityConfigurationName: aws.String(name),
	}

	if v, ok := d.GetOk("trace_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TraceConfiguration = expandTraceConfiguration(v.([]interface{}))
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
	conn := meta.(*conns.AWSClient).AppRunnerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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
		return diag.FromErr(fmt.Errorf("error reading App Runner Observability Configuration (%s): %w", d.Id(), err))
	}

	if output == nil || output.ObservabilityConfiguration == nil {
		return diag.FromErr(fmt.Errorf("error reading App Runner Observability Configuration (%s): empty output", d.Id()))
	}

	if aws.StringValue(output.ObservabilityConfiguration.Status) == ObservabilityConfigurationStatusInactive {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading App Runner Observability Configuration (%s): %s after creation", d.Id(), aws.StringValue(output.ObservabilityConfiguration.Status)))
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
		return diag.Errorf("error setting trace_configuration: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for App Runner Observability Configuration (%s): %s", arn, err))
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceObservabilityConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating App Runner Observability Configuration (%s) tags: %s", d.Get("arn").(string), err))
		}
	}

	return resourceObservabilityConfigurationRead(ctx, d, meta)
}

func resourceObservabilityConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()

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
