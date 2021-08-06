package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func resourceAwsAppStreamFleet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsAppStreamFleetCreate,
		ReadWithoutTimeout:   resourceAwsAppStreamFleetRead,
		UpdateWithoutTimeout: resourceAwsAppStreamFleetUpdate,
		DeleteWithoutTimeout: resourceAwsAppStreamFleetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"compute_capacity": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_instances": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"disconnect_timeout_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(60, 360000),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"domain_join_info": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"organizational_unit_distinguished_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"enable_default_internet_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"fleet_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(appstream.FleetType_Values(), false),
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},
			"idle_disconnect_timeout_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(60, 3600),
			},
			"image_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"image_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"max_user_duration_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(600, 360000),
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"stream_view": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(appstream.StreamView_Values(), false),
			},
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{appstream.FleetStateRunning, appstream.FleetStateStopped}, false),
			},
			"vpc_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsAppStreamFleetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn
	input := &appstream.CreateFleetInput{
		Name: aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))),
	}

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	if v, ok := d.GetOk("compute_capacity"); ok {
		input.ComputeCapacity = expandComputeCapacity(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disconnect_timeout_in_seconds"); ok {
		input.DisconnectTimeoutInSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_join_info"); ok {
		input.DomainJoinInfo = expandDomainJoinInfo(v.([]interface{}))
	}

	if v, ok := d.GetOk("enable_default_internet_access"); ok {
		input.EnableDefaultInternetAccess = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("fleet_type"); ok {
		input.FleetType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_name"); ok {
		input.ImageName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_type"); ok {
		input.InstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_user_duration_in_seconds"); ok {
		input.MaxUserDurationInSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		input.VpcConfig = expandVpcConfig(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().AppstreamTags()
	}

	var err error
	var output *appstream.CreateFleetOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		output, err = conn.CreateFleetWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateFleetWithContext(ctx, input)
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Appstream Fleet (%s): %w", d.Id(), err))
	}

	if v, ok := d.GetOk("state"); ok {
		if v == "RUNNING" {
			desiredState := v
			_, err := conn.StartFleetWithContext(ctx, &appstream.StartFleetInput{
				Name: output.Fleet.Name,
			})

			if err != nil {
				return diag.FromErr(fmt.Errorf("error starting Appstream Fleet (%s): %w", d.Id(), err))
			}
			for {
				resp, err := conn.DescribeFleetsWithContext(ctx, &appstream.DescribeFleetsInput{
					Names: aws.StringSlice([]string{*input.Name}),
				})
				if err != nil {
					return diag.FromErr(fmt.Errorf("error describing Appstream Fleet (%s): %w", d.Id(), err))
				}

				currentState := resp.Fleets[0].State
				if aws.StringValue(currentState) == desiredState {
					break
				}
				if aws.StringValue(currentState) != desiredState {
					time.Sleep(20 * time.Second)
					continue
				}

			}
		}
	}

	d.SetId(aws.StringValue(output.Fleet.Name))

	return resourceAwsAppStreamFleetRead(ctx, d, meta)
}

func resourceAwsAppStreamFleetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeFleetsWithContext(ctx, &appstream.DescribeFleetsInput{Names: []*string{aws.String(d.Id())}})
	if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appstream Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream Fleet (%s): %w", d.Id(), err))
	}

	for _, v := range resp.Fleets {
		d.Set("name", v.Name)
		d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(v.Name)))

		if err = d.Set("compute_capacity", flattenComputeCapacity(v.ComputeCapacityStatus)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Fleet (%s): %w", "compute_capacity", d.Id(), err))
		}
		if err = d.Set("domain_join_info", flattenDomainInfo(v.DomainJoinInfo)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Fleet (%s): %w", "domain_join_info", d.Id(), err))
		}

		d.Set("description", v.Description)
		d.Set("display_name", v.DisplayName)
		d.Set("disconnect_timeout_in_seconds", v.DisconnectTimeoutInSeconds)
		d.Set("idle_disconnect_timeout_in_seconds", v.IdleDisconnectTimeoutInSeconds)
		d.Set("enable_default_internet_access", v.EnableDefaultInternetAccess)
		d.Set("fleet_type", v.FleetType)
		d.Set("image_name", v.ImageName)
		d.Set("image_arn", v.ImageArn)
		d.Set("iam_role_arn", v.IamRoleArn)
		d.Set("stream_view", v.StreamView)
		d.Set("arn", v.Arn)

		d.Set("instance_type", v.InstanceType)
		d.Set("max_user_duration_in_seconds", v.MaxUserDurationInSeconds)
		if err = d.Set("vpc_config", flattenVpcConfig(v.VpcConfig)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Fleet (%s): %w", "vpc_config", d.Id(), err))
		}

		d.Set("state", v.State)

		tg, err := conn.ListTagsForResource(&appstream.ListTagsForResourceInput{
			ResourceArn: v.Arn,
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing stack tags for AppStream Stack (%s): %w", d.Id(), err))
		}
		if tg.Tags == nil {
			log.Printf("[DEBUG] Apsstream Stack tags (%s) not found", d.Id())
			return nil
		}
		tags := keyvaluetags.AppstreamKeyValueTags(tg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

		if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "tags", d.Id(), err))
		}

		if err = d.Set("tags_all", tags.Map()); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream Stack (%s): %w", "tags_all", d.Id(), err))
		}

		return nil
	}
	return nil
}

func resourceAwsAppStreamFleetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*AWSClient).appstreamconn
	input := &appstream.UpdateFleetInput{
		Name: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("compute_capacity") {
		input.ComputeCapacity = expandComputeCapacity(d.Get("compute_capacity").([]interface{}))
	}

	if d.HasChange("domain_join_info") {
		input.DomainJoinInfo = expandDomainJoinInfo(d.Get("domain_join_info").([]interface{}))
	}

	if d.HasChange("disconnect_timeout_in_seconds") {
		input.DisconnectTimeoutInSeconds = aws.Int64(int64(d.Get("disconnect_timeout_in_seconds").(int)))
	}

	if d.HasChange("idle_disconnect_timeout_in_seconds") {
		input.IdleDisconnectTimeoutInSeconds = aws.Int64(int64(d.Get("idle_disconnect_timeout_in_seconds").(int)))
	}

	if d.HasChange("display_name") {
		input.DisplayName = aws.String(d.Get("display_name").(string))
	}

	if d.HasChange("image_name") {
		input.ImageName = aws.String(d.Get("image_name").(string))
	}

	if d.HasChange("image_arn") {
		input.ImageArn = aws.String(d.Get("image_arn").(string))
	}

	if d.HasChange("iam_role_arn") {
		input.IamRoleArn = aws.String(d.Get("iam_role_arn").(string))
	}

	if d.HasChange("stream_view") {
		input.StreamView = aws.String(d.Get("stream_view").(string))
	}

	if d.HasChange("instance_type") {
		input.InstanceType = aws.String(d.Get("instance_type").(string))
	}

	if d.HasChange("max_user_duration_in_seconds") {
		input.MaxUserDurationInSeconds = aws.Int64(int64(d.Get("max_user_duration_in_seconds").(int)))
	}

	if d.HasChange("vpc_config") {
		input.VpcConfig = expandVpcConfig(d.Get("vpc_config").([]interface{}))
	}

	resp, err := conn.UpdateFleetWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Appstream Fleet (%s): %w", d.Id(), err))
	}

	if d.HasChange("tags") {
		arn := aws.StringValue(resp.Fleet.Arn)

		o, n := d.GetChange("tags")
		if err := keyvaluetags.AppstreamUpdateTags(conn, arn, o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating Appstream Fleet tags (%s): %w", d.Id(), err))
		}
	}

	desiredState := d.Get("state")
	if d.HasChange("state") {
		if desiredState == "STOPPED" {
			_, err := conn.StopFleetWithContext(ctx, &appstream.StopFleetInput{
				Name: aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))),
			})
			if err != nil {
				return diag.FromErr(err)
			}
			for {
				resp, err := conn.DescribeFleetsWithContext(ctx, &appstream.DescribeFleetsInput{
					Names: aws.StringSlice([]string{*input.Name}),
				})
				if err != nil {
					return diag.FromErr(fmt.Errorf("error describing Appstream Fleet (%s): %w", d.Id(), err))
				}

				currentState := resp.Fleets[0].State
				if aws.StringValue(currentState) == desiredState {
					break
				}
				if aws.StringValue(currentState) != desiredState {
					time.Sleep(20 * time.Second)
					continue
				}
			}
		} else if desiredState == "RUNNING" {
			_, err = conn.StartFleetWithContext(ctx, &appstream.StartFleetInput{
				Name: aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))),
			})
			if err != nil {
				return diag.FromErr(err)
			}
			for {
				resp, err := conn.DescribeFleetsWithContext(ctx, &appstream.DescribeFleetsInput{
					Names: aws.StringSlice([]string{*input.Name}),
				})
				if err != nil {
					return diag.FromErr(fmt.Errorf("error describing Appstream Fleet (%s): %w", d.Id(), err))
				}

				currentState := resp.Fleets[0].State
				if aws.StringValue(currentState) == desiredState {
					break
				}
				if aws.StringValue(currentState) != desiredState {
					time.Sleep(20 * time.Second)
					continue
				}
			}
		}
	}
	return resourceAwsAppStreamFleetRead(ctx, d, meta)

}

func resourceAwsAppStreamFleetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	resp, err := conn.DescribeFleetsWithContext(ctx, &appstream.DescribeFleetsInput{
		Names: aws.StringSlice([]string{*aws.String(d.Id())}),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream Fleet (%s): %w", d.Id(), err))
	}

	currentState := aws.StringValue(resp.Fleets[0].State)

	if currentState == "RUNNING" {
		desiredState := "STOPPED"
		_, err = conn.StopFleet(&appstream.StopFleetInput{
			Name: aws.String(d.Id()),
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error stopping Appstream Fleet (%s): %w", d.Id(), err))
		}
		for {
			resp, err = conn.DescribeFleetsWithContext(ctx, &appstream.DescribeFleetsInput{
				Names: aws.StringSlice([]string{*aws.String(d.Id())}),
			})
			if err != nil {
				return diag.FromErr(fmt.Errorf("error describing Appstream Fleet (%s): %w", d.Id(), err))
			}

			cState := resp.Fleets[0].State
			if aws.StringValue(cState) == desiredState {
				break
			}
			if aws.StringValue(cState) != desiredState {
				time.Sleep(20 * time.Second)
				continue
			}
		}
	}

	_, err = conn.DeleteFleetWithContext(ctx, &appstream.DeleteFleetInput{
		Name: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Appstream Fleet (%s): %w", d.Id(), err))
	}
	return nil

}

func expandComputeCapacity(computeCapacity []interface{}) *appstream.ComputeCapacity {
	if len(computeCapacity) == 0 {
		return nil
	}

	computeConf := &appstream.ComputeCapacity{}

	attr := computeCapacity[0].(map[string]interface{})
	if v, ok := attr["desired_instances"]; ok {
		computeConf.DesiredInstances = aws.Int64(int64(v.(int)))
	}

	return computeConf
}

func flattenComputeCapacity(computeCapacity *appstream.ComputeCapacityStatus) []interface{} {
	if computeCapacity == nil {
		return nil
	}

	compAttr := map[string]interface{}{}
	compAttr["desired_instances"] = aws.Int64Value(computeCapacity.Desired)

	return []interface{}{compAttr}
}

func expandDomainJoinInfo(domainInfo []interface{}) *appstream.DomainJoinInfo {
	if len(domainInfo) == 0 {
		return nil
	}

	infoConfig := &appstream.DomainJoinInfo{}

	attr := domainInfo[0].(map[string]interface{})
	if v, ok := attr["directory_name"]; ok {
		infoConfig.DirectoryName = aws.String(v.(string))
	}
	if v, ok := attr["organizational_unit_distinguished_name"]; ok {
		infoConfig.OrganizationalUnitDistinguishedName = aws.String(v.(string))
	}

	return infoConfig
}

func flattenDomainInfo(domainInfo *appstream.DomainJoinInfo) []interface{} {
	if domainInfo == nil {
		return nil
	}

	compAttr := map[string]interface{}{}
	compAttr["directory_name"] = aws.StringValue(domainInfo.DirectoryName)
	compAttr["organizational_unit_distinguished_name"] = aws.StringValue(domainInfo.OrganizationalUnitDistinguishedName)

	return []interface{}{compAttr}
}

func expandVpcConfig(vpcConfig []interface{}) *appstream.VpcConfig {
	if len(vpcConfig) == 0 {
		return nil
	}

	infoConfig := &appstream.VpcConfig{}

	attr := vpcConfig[0].(map[string]interface{})
	if v, ok := attr["security_group_ids"]; ok {
		infoConfig.SecurityGroupIds = expandStringList(v.([]interface{}))
	}
	if v, ok := attr["subnet_ids"]; ok {
		infoConfig.SubnetIds = expandStringList(v.([]interface{}))
	}

	return infoConfig
}

func flattenVpcConfig(vpcConfig *appstream.VpcConfig) []interface{} {
	if vpcConfig == nil {
		return nil
	}

	compAttr := map[string]interface{}{}
	compAttr["security_group_ids"] = aws.StringValueSlice(vpcConfig.SecurityGroupIds)
	compAttr["subnet_ids"] = aws.StringValueSlice(vpcConfig.SubnetIds)

	return []interface{}{compAttr}
}
