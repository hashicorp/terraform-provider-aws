// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_deployment_strategy", name="Deployment Strategy")
// @Tags(identifierAttribute="arn")
func resourceDeploymentStrategy() *schema.Resource {
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
	}
}

func resourceDeploymentStrategyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := appconfig.CreateDeploymentStrategyInput{
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

	strategy, err := conn.CreateDeploymentStrategy(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppConfig Deployment Strategy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(strategy.Id))

	return append(diags, resourceDeploymentStrategyRead(ctx, d, meta)...)
}

func resourceDeploymentStrategyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	output, err := findDeploymentStrategyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppConfig Deployment Strategy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Deployment Strategy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, deploymentStrategyARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrDescription, output.Description)
	d.Set("deployment_duration_in_minutes", output.DeploymentDurationInMinutes)
	d.Set("final_bake_time_in_minutes", output.FinalBakeTimeInMinutes)
	d.Set("growth_factor", output.GrowthFactor)
	d.Set("growth_type", output.GrowthType)
	d.Set(names.AttrName, output.Name)
	d.Set("replicate_to", output.ReplicateTo)

	return diags
}

func resourceDeploymentStrategyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := appconfig.UpdateDeploymentStrategyInput{
			DeploymentStrategyId: aws.String(d.Id()),
		}

		if d.HasChange("deployment_duration_in_minutes") {
			input.DeploymentDurationInMinutes = aws.Int32(int32(d.Get("deployment_duration_in_minutes").(int)))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("final_bake_time_in_minutes") {
			input.FinalBakeTimeInMinutes = aws.Int32(int32(d.Get("final_bake_time_in_minutes").(int)))
		}

		if d.HasChange("growth_factor") {
			input.GrowthFactor = aws.Float32(d.Get("growth_factor").(float32))
		}

		if d.HasChange("growth_type") {
			input.GrowthType = awstypes.GrowthType(d.Get("growth_type").(string))
		}

		_, err := conn.UpdateDeploymentStrategy(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Deployment Strategy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDeploymentStrategyRead(ctx, d, meta)...)
}

func resourceDeploymentStrategyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	log.Printf("[INFO] Deleting AppConfig Deployment Strategy: %s", d.Id())
	input := appconfig.DeleteDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	}
	_, err := conn.DeleteDeploymentStrategy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppConfig Deployment Strategy (%s): %s", d.Id(), err)
	}

	return diags
}

func findDeploymentStrategyByID(ctx context.Context, conn *appconfig.Client, id string) (*appconfig.GetDeploymentStrategyOutput, error) {
	input := appconfig.GetDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(id),
	}

	return findDeploymentStrategy(ctx, conn, &input)
}

func findDeploymentStrategy(ctx context.Context, conn *appconfig.Client, input *appconfig.GetDeploymentStrategyInput) (*appconfig.GetDeploymentStrategyOutput, error) {
	output, err := conn.GetDeploymentStrategy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func deploymentStrategyARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, "appconfig", "deploymentstrategy/"+id)
}
