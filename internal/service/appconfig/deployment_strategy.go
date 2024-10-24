// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_duration_in_minutes": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 1440),
			},
			names.AttrDescription: {
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.GrowthTypeLinear,
				ValidateDiagFunc: enum.Validate[awstypes.GrowthType](),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"replicate_to": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ReplicateTo](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDeploymentStrategyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appconfig.CreateDeploymentStrategyInput{
		DeploymentDurationInMinutes: aws.Int32(int32(d.Get("deployment_duration_in_minutes").(int))),
		GrowthFactor:                aws.Float32(float32(d.Get("growth_factor").(float64))),
		GrowthType:                  awstypes.GrowthType(d.Get("growth_type").(string)),
		Name:                        aws.String(name),
		ReplicateTo:                 awstypes.ReplicateTo(d.Get("replicate_to").(string)),
		Tags:                        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("final_bake_time_in_minutes"); ok {
		input.FinalBakeTimeInMinutes = int32(v.(int))
	}

	strategy, err := conn.CreateDeploymentStrategy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Deployment Strategy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(strategy.Id))

	return append(diags, resourceDeploymentStrategyRead(ctx, d, meta)...)
}

func resourceDeploymentStrategyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	input := &appconfig.GetDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	}

	output, err := conn.GetDeploymentStrategy(ctx, input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

	d.Set(names.AttrDescription, output.Description)
	d.Set("deployment_duration_in_minutes", output.DeploymentDurationInMinutes)
	d.Set("final_bake_time_in_minutes", output.FinalBakeTimeInMinutes)
	d.Set("growth_factor", output.GrowthFactor)
	d.Set("growth_type", output.GrowthType)
	d.Set(names.AttrName, output.Name)
	d.Set("replicate_to", output.ReplicateTo)

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("deploymentstrategy/%s", d.Id()),
		Service:   "appconfig",
	}.String()
	d.Set(names.AttrARN, arn)

	return diags
}

func resourceDeploymentStrategyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		updateInput := &appconfig.UpdateDeploymentStrategyInput{
			DeploymentStrategyId: aws.String(d.Id()),
		}

		if d.HasChange("deployment_duration_in_minutes") {
			updateInput.DeploymentDurationInMinutes = aws.Int32(int32(d.Get("deployment_duration_in_minutes").(int)))
		}

		if d.HasChange(names.AttrDescription) {
			updateInput.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("final_bake_time_in_minutes") {
			updateInput.FinalBakeTimeInMinutes = aws.Int32(int32(d.Get("final_bake_time_in_minutes").(int)))
		}

		if d.HasChange("growth_factor") {
			updateInput.GrowthFactor = aws.Float32(d.Get("growth_factor").(float32))
		}

		if d.HasChange("growth_type") {
			updateInput.GrowthType = awstypes.GrowthType(d.Get("growth_type").(string))
		}

		_, err := conn.UpdateDeploymentStrategy(ctx, updateInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Deployment Strategy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDeploymentStrategyRead(ctx, d, meta)...)
}

func resourceDeploymentStrategyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	log.Printf("[INFO] Deleting AppConfig Deployment Strategy: %s", d.Id())
	_, err := conn.DeleteDeploymentStrategy(ctx, &appconfig.DeleteDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appconfig Deployment Strategy (%s): %s", d.Id(), err)
	}

	return diags
}
