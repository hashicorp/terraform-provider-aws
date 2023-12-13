// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_deployment_strategy", name="Deployment Strategy")
// @Tags(identifierAttribute="arn")
func ResourceDeploymentStrategy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeploymentStrategyCreate,
		ReadWithoutTimeout:   resourceDeploymentStrategyRead,
		UpdateWithoutTimeout: resourceDeploymentStrategyUpdate,
		DeleteWithoutTimeout: resourceDeploymentStrategyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_duration_in_minutes": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 1440),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"final_bake_time_in_minutes": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 1440),
			},
			"growth_factor": {
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatBetween(1.0, 100.0),
			},
			"growth_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      appconfig.GrowthTypeLinear,
				ValidateFunc: validation.StringInSlice(appconfig.GrowthType_Values(), false),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"replicate_to": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(appconfig.ReplicateTo_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDeploymentStrategyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	name := d.Get("name").(string)
	input := &appconfig.CreateDeploymentStrategyInput{
		DeploymentDurationInMinutes: aws.Int64(int64(d.Get("deployment_duration_in_minutes").(int))),
		GrowthFactor:                aws.Float64(d.Get("growth_factor").(float64)),
		GrowthType:                  aws.String(d.Get("growth_type").(string)),
		Name:                        aws.String(name),
		ReplicateTo:                 aws.String(d.Get("replicate_to").(string)),
		Tags:                        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("final_bake_time_in_minutes"); ok {
		input.FinalBakeTimeInMinutes = aws.Int64(int64(v.(int)))
	}

	strategy, err := conn.CreateDeploymentStrategyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Deployment Strategy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(strategy.Id))

	return append(diags, resourceDeploymentStrategyRead(ctx, d, meta)...)
}

func resourceDeploymentStrategyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	input := &appconfig.GetDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	}

	output, err := conn.GetDeploymentStrategyWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appconfig Deployment Strategy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting AppConfig Deployment Strategy (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting AppConfig Deployment Strategy (%s): empty response", d.Id())
	}

	d.Set("description", output.Description)
	d.Set("deployment_duration_in_minutes", output.DeploymentDurationInMinutes)
	d.Set("final_bake_time_in_minutes", output.FinalBakeTimeInMinutes)
	d.Set("growth_factor", output.GrowthFactor)
	d.Set("growth_type", output.GrowthType)
	d.Set("name", output.Name)
	d.Set("replicate_to", output.ReplicateTo)

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("deploymentstrategy/%s", d.Id()),
		Service:   "appconfig",
	}.String()
	d.Set("arn", arn)

	return diags
}

func resourceDeploymentStrategyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		updateInput := &appconfig.UpdateDeploymentStrategyInput{
			DeploymentStrategyId: aws.String(d.Id()),
		}

		if d.HasChange("deployment_duration_in_minutes") {
			updateInput.DeploymentDurationInMinutes = aws.Int64(int64(d.Get("deployment_duration_in_minutes").(int)))
		}

		if d.HasChange("description") {
			updateInput.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("final_bake_time_in_minutes") {
			updateInput.FinalBakeTimeInMinutes = aws.Int64(int64(d.Get("final_bake_time_in_minutes").(int)))
		}

		if d.HasChange("growth_factor") {
			updateInput.GrowthFactor = aws.Float64(d.Get("growth_factor").(float64))
		}

		if d.HasChange("growth_type") {
			updateInput.GrowthType = aws.String(d.Get("growth_type").(string))
		}

		_, err := conn.UpdateDeploymentStrategyWithContext(ctx, updateInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Deployment Strategy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDeploymentStrategyRead(ctx, d, meta)...)
}

func resourceDeploymentStrategyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn(ctx)

	log.Printf("[INFO] Deleting AppConfig Deployment Strategy: %s", d.Id())
	_, err := conn.DeleteDeploymentStrategyWithContext(ctx, &appconfig.DeleteDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appconfig Deployment Strategy (%s): %s", d.Id(), err)
	}

	return diags
}
