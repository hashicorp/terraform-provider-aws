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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apprunner_auto_scaling_configuration_version", name="AutoScaling Configuration Version")
// @Tags(identifierAttribute="arn")
func resourceAutoScalingConfigurationVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAutoScalingConfigurationCreate,
		ReadWithoutTimeout:   resourceAutoScalingConfigurationRead,
		UpdateWithoutTimeout: resourceAutoScalingConfigurationUpdate,
		DeleteWithoutTimeout: resourceAutoScalingConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			"has_associated_service": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
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
			names.AttrStatus: {
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
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	name := d.Get("auto_scaling_configuration_name").(string)
	input := &apprunner.CreateAutoScalingConfigurationInput{
		AutoScalingConfigurationName: aws.String(name),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		input.MaxConcurrency = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("max_size"); ok {
		input.MaxSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("min_size"); ok {
		input.MinSize = aws.Int32(int32(v.(int)))
	}

	output, err := conn.CreateAutoScalingConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Runner AutoScaling Configuration Version (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.AutoScalingConfiguration.AutoScalingConfigurationArn))

	if _, err := waitAutoScalingConfigurationCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AutoScaling Configuration Version (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAutoScalingConfigurationRead(ctx, d, meta)...)
}

func resourceAutoScalingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	config, err := findAutoScalingConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Runner AutoScaling Configuration Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Runner AutoScaling Configuration Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, config.AutoScalingConfigurationArn)
	d.Set("auto_scaling_configuration_name", config.AutoScalingConfigurationName)
	d.Set("auto_scaling_configuration_revision", config.AutoScalingConfigurationRevision)
	d.Set("has_associated_service", config.HasAssociatedService)
	d.Set("is_default", config.IsDefault)
	d.Set("latest", config.Latest)
	d.Set("max_concurrency", config.MaxConcurrency)
	d.Set("max_size", config.MaxSize)
	d.Set("min_size", config.MinSize)
	d.Set(names.AttrStatus, config.Status)

	return diags
}

func resourceAutoScalingConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceAutoScalingConfigurationRead(ctx, d, meta)
}

func resourceAutoScalingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	log.Printf("[INFO] Deleting App Runner AutoScaling Configuration Version: %s", d.Id())
	_, err := conn.DeleteAutoScalingConfiguration(ctx, &apprunner.DeleteAutoScalingConfigurationInput{
		AutoScalingConfigurationArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Runner AutoScaling Configuration Version (%s): %s", d.Id(), err)
	}

	if _, err := waitAutoScalingConfigurationDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AutoScaling Configuration Version (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findAutoScalingConfigurationByARN(ctx context.Context, conn *apprunner.Client, arn string) (*types.AutoScalingConfiguration, error) {
	input := &apprunner.DescribeAutoScalingConfigurationInput{
		AutoScalingConfigurationArn: aws.String(arn),
	}

	output, err := conn.DescribeAutoScalingConfiguration(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AutoScalingConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := string(output.AutoScalingConfiguration.Status); status == autoScalingConfigurationStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output.AutoScalingConfiguration, nil
}

func findAutoScalingConfigurationSummary(ctx context.Context, conn *apprunner.Client, input *apprunner.ListAutoScalingConfigurationsInput, filter tfslices.Predicate[*types.AutoScalingConfigurationSummary]) (*types.AutoScalingConfigurationSummary, error) {
	output, err := findAutoScalingConfigurationSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findAutoScalingConfigurationSummaries(ctx context.Context, conn *apprunner.Client, input *apprunner.ListAutoScalingConfigurationsInput, filter tfslices.Predicate[*types.AutoScalingConfigurationSummary]) ([]*types.AutoScalingConfigurationSummary, error) {
	var output []*types.AutoScalingConfigurationSummary

	pages := apprunner.NewListAutoScalingConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AutoScalingConfigurationSummaryList {
			v := v
			if v := &v; filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

const (
	autoScalingConfigurationStatusActive   = "active"
	autoScalingConfigurationStatusInactive = "inactive"
)

func statusAutoScalingConfiguration(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAutoScalingConfigurationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitAutoScalingConfigurationCreated(ctx context.Context, conn *apprunner.Client, arn string) (*types.AutoScalingConfiguration, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{autoScalingConfigurationStatusActive},
		Refresh: statusAutoScalingConfiguration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AutoScalingConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitAutoScalingConfigurationDeleted(ctx context.Context, conn *apprunner.Client, arn string) (*types.AutoScalingConfiguration, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{autoScalingConfigurationStatusActive},
		Target:  []string{},
		Refresh: statusAutoScalingConfiguration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AutoScalingConfiguration); ok {
		return output, err
	}

	return nil, err
}
