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
)

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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDeploymentStrategyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &appconfig.CreateDeploymentStrategyInput{
		DeploymentDurationInMinutes: aws.Int64(int64(d.Get("deployment_duration_in_minutes").(int))),
		GrowthFactor:                aws.Float64(d.Get("growth_factor").(float64)),
		GrowthType:                  aws.String(d.Get("growth_type").(string)),
		Name:                        aws.String(name),
		ReplicateTo:                 aws.String(d.Get("replicate_to").(string)),
		Tags:                        Tags(tags.IgnoreAWS()),
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
	conn := meta.(*conns.AWSClient).AppConfigConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for AppConfig Deployment Strategy (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceDeploymentStrategyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn()

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

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppConfig Deployment Strategy (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDeploymentStrategyRead(ctx, d, meta)...)
}

func resourceDeploymentStrategyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigConn()

	input := &appconfig.DeleteDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	}

	_, err := conn.DeleteDeploymentStrategyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appconfig Deployment Strategy (%s): %s", d.Id(), err)
	}

	return diags
}
