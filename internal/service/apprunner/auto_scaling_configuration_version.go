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

// @SDKResource("aws_apprunner_auto_scaling_configuration_version", name="AutoScaling Configuration Version")
// @Tags(identifierAttribute="arn")
func ResourceAutoScalingConfigurationVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAutoScalingConfigurationCreate,
		ReadWithoutTimeout:   resourceAutoScalingConfigurationRead,
		UpdateWithoutTimeout: resourceAutoScalingConfigurationUpdate,
		DeleteWithoutTimeout: resourceAutoScalingConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_configuration_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"auto_scaling_configuration_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"latest": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"max_concurrency": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 200),
			},
			"max_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      25,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 25),
			},
			"min_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 25),
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

func resourceAutoScalingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	name := d.Get("auto_scaling_configuration_name").(string)
	input := &apprunner.CreateAutoScalingConfigurationInput{
		AutoScalingConfigurationName: aws.String(name),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		input.MaxConcurrency = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("max_size"); ok {
		input.MaxSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("min_size"); ok {
		input.MinSize = aws.Int64(int64(v.(int)))
	}

	output, err := conn.CreateAutoScalingConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating App Runner AutoScaling Configuration Version (%s): %s", name, err)
	}

	if output == nil || output.AutoScalingConfiguration == nil {
		return diag.Errorf("creating App Runner AutoScaling Configuration Version (%s): empty output", name)
	}

	d.SetId(aws.StringValue(output.AutoScalingConfiguration.AutoScalingConfigurationArn))

	if err := WaitAutoScalingConfigurationActive(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for AutoScaling Configuration Version (%s) creation: %s", d.Id(), err)
	}

	return resourceAutoScalingConfigurationRead(ctx, d, meta)
}

func resourceAutoScalingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	input := &apprunner.DescribeAutoScalingConfigurationInput{
		AutoScalingConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeAutoScalingConfigurationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] App Runner AutoScaling Configuration Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading App Runner AutoScaling Configuration Version (%s): %s", d.Id(), err)
	}

	if output == nil || output.AutoScalingConfiguration == nil {
		return diag.Errorf("reading App Runner AutoScaling Configuration Version (%s): empty output", d.Id())
	}

	if aws.StringValue(output.AutoScalingConfiguration.Status) == AutoScalingConfigurationStatusInactive {
		if d.IsNewResource() {
			return diag.Errorf("reading App Runner AutoScaling Configuration Version (%s): %s after creation", d.Id(), aws.StringValue(output.AutoScalingConfiguration.Status))
		}
		log.Printf("[WARN] App Runner AutoScaling Configuration Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	config := output.AutoScalingConfiguration
	arn := aws.StringValue(config.AutoScalingConfigurationArn)

	d.Set("arn", arn)
	d.Set("auto_scaling_configuration_name", config.AutoScalingConfigurationName)
	d.Set("auto_scaling_configuration_revision", config.AutoScalingConfigurationRevision)
	d.Set("latest", config.Latest)
	d.Set("max_concurrency", config.MaxConcurrency)
	d.Set("max_size", config.MaxSize)
	d.Set("min_size", config.MinSize)
	d.Set("status", config.Status)

	return nil
}

func resourceAutoScalingConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceAutoScalingConfigurationRead(ctx, d, meta)
}

func resourceAutoScalingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	input := &apprunner.DeleteAutoScalingConfigurationInput{
		AutoScalingConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteAutoScalingConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting App Runner AutoScaling Configuration Version (%s): %s", d.Id(), err)
	}

	if err := WaitAutoScalingConfigurationInactive(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("waiting for AutoScaling Configuration Version (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
