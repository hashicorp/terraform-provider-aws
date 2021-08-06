package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func resourceAwsAppstreamImageBuilder() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsAppstreamImageBuilderCreate,
		ReadWithoutTimeout:   resourceAwsAppstreamImageBuilderRead,
		UpdateWithoutTimeout: resourceAwsAppstreamImageBuilderUpdate,
		DeleteWithoutTimeout: resourceAwsAppstreamImageBuilderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"access_endpoints": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appstream.AccessEndpointType_Values(), false),
						},
						"vpce_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"appstream_agent_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"domain_join_info": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
				ForceNew: true,
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"image_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"image_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{appstream.ImageBuilderStateRunning, appstream.ImageBuilderStateStopped}, false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
			"tags":     tagsSchemaForceNew(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsAppstreamImageBuilderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn
	input := &appstream.CreateImageBuilderInput{
		Name: aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))),
	}

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	if v, ok := d.GetOk("access_endpoints"); ok {
		input.AccessEndpoints = expandAccessEndpoints(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("appstream_agent_version"); ok {
		input.AppstreamAgentVersion = aws.String(v.(string))
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

	if v, ok := d.GetOk("image_name"); ok {
		input.ImageName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_type"); ok {
		input.InstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		input.VpcConfig = expandVpcConfig(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().AppstreamTags()
	}

	resp, err := conn.CreateImageBuilderWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Appstream ImageBuilder (%s): %w", d.Id(), err))
	}

	if v, ok := d.GetOk("state"); ok {
		if v == "RUNNING" {
			desiredState := v
			_, err := conn.StartImageBuilderWithContext(ctx, &appstream.StartImageBuilderInput{
				Name: resp.ImageBuilder.Name,
			})

			if err != nil {
				return diag.FromErr(fmt.Errorf("error starting Appstream ImageBuilder (%s): %w", d.Id(), err))
			}
			for {
				resp, err := conn.DescribeImageBuildersWithContext(ctx, &appstream.DescribeImageBuildersInput{
					Names: aws.StringSlice([]string{*input.Name}),
				})
				if err != nil {
					return diag.FromErr(fmt.Errorf("error describing Appstream ImageBuilder (%s): %w", d.Id(), err))
				}

				currentState := resp.ImageBuilders[0].State
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

	d.SetId(aws.StringValue(resp.ImageBuilder.Name))

	return resourceAwsAppstreamImageBuilderRead(ctx, d, meta)
}

func resourceAwsAppstreamImageBuilderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeImageBuildersWithContext(ctx, &appstream.DescribeImageBuildersInput{Names: []*string{aws.String(d.Id())}})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream ImageBuilder (%s): %w", d.Id(), err))
	}
	for _, v := range resp.ImageBuilders {
		d.Set("name", v.Name)
		d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(v.Name)))

		if err = d.Set("access_endpoints", flattenAccessEndpoints(v.AccessEndpoints)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "access_endpoints", d.Id(), err))
		}
		if err = d.Set("domain_join_info", flattenDomainInfo(v.DomainJoinInfo)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "domain_join_info", d.Id(), err))
		}

		d.Set("appstream_agent_version", v.AppstreamAgentVersion)
		d.Set("description", v.Description)
		d.Set("display_name", v.DisplayName)
		d.Set("enable_default_internet_access", v.EnableDefaultInternetAccess)
		d.Set("image_arn", v.ImageArn)
		d.Set("iam_role_arn", v.IamRoleArn)
		d.Set("arn", v.Arn)

		d.Set("instance_type", v.InstanceType)
		if err = d.Set("vpc_config", flattenVpcConfig(v.VpcConfig)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "vpc_config", d.Id(), err))
		}

		d.Set("state", v.State)

		tg, err := conn.ListTagsForResource(&appstream.ListTagsForResourceInput{
			ResourceArn: v.Arn,
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing stack tags for AppStream ImageBuilder (%s): %w", d.Id(), err))
		}
		tags := keyvaluetags.AppstreamKeyValueTags(tg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

		if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "tags", d.Id(), err))
		}

		if err = d.Set("tags_all", tags.Map()); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "tags_all", d.Id(), err))
		}

		return nil
	}
	return nil
}

func resourceAwsAppstreamImageBuilderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	desiredState := d.Get("state")
	if d.HasChange("state") {
		if desiredState == "STOPPED" {
			_, err := conn.StopImageBuilderWithContext(ctx, &appstream.StopImageBuilderInput{
				Name: aws.String(d.Id()),
			})
			if err != nil {
				return diag.FromErr(err)
			}
			for {

				resp, err := conn.DescribeImageBuildersWithContext(ctx, &appstream.DescribeImageBuildersInput{
					Names: aws.StringSlice([]string{d.Id()}),
				})
				if err != nil {
					return diag.FromErr(fmt.Errorf("error describing Appstream ImageBuilder (%s): %w", d.Id(), err))
				}

				currentState := resp.ImageBuilders[0].State
				if aws.StringValue(currentState) == desiredState {
					break
				}
				if aws.StringValue(currentState) != desiredState {
					time.Sleep(20 * time.Second)
					continue
				}
			}
		} else if desiredState == "RUNNING" {
			_, err := conn.StartImageBuilderWithContext(ctx, &appstream.StartImageBuilderInput{
				Name: aws.String(d.Id()),
			})
			if err != nil {
				return diag.FromErr(err)
			}
			for {

				resp, err := conn.DescribeImageBuildersWithContext(ctx, &appstream.DescribeImageBuildersInput{
					Names: aws.StringSlice([]string{d.Id()}),
				})
				if err != nil {
					return diag.FromErr(fmt.Errorf("error describing Appstream ImageBuilder (%s): %w", d.Id(), err))
				}

				currentState := resp.ImageBuilders[0].State
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

	return resourceAwsAppstreamImageBuilderRead(ctx, d, meta)

}

func resourceAwsAppstreamImageBuilderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	_, err := conn.DeleteImageBuilderWithContext(ctx, &appstream.DeleteImageBuilderInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Appstream ImageBuilder (%s): %w", d.Id(), err))
	}
	return nil

}

func expandAccessEndpoints(accessEndpoints []interface{}) []*appstream.AccessEndpoint {
	if len(accessEndpoints) == 0 {
		return nil
	}

	var endpoints []*appstream.AccessEndpoint

	for _, v := range accessEndpoints {
		v1 := v.(map[string]interface{})

		endpoint := &appstream.AccessEndpoint{
			EndpointType: aws.String(v1["endpoint_type"].(string)),
		}
		if v2, ok := v1["vpce_id"]; ok {
			endpoint.VpceId = aws.String(v2.(string))
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

func flattenAccessEndpoints(accessEndpoints []*appstream.AccessEndpoint) []map[string]interface{} {
	if accessEndpoints == nil {
		return nil
	}

	var endpoints []map[string]interface{}

	for _, endpoint := range accessEndpoints {
		endpoints = append(endpoints, map[string]interface{}{
			"endpoint_type": aws.StringValue(endpoint.EndpointType),
			"vpce_id":       aws.StringValue(endpoint.VpceId),
		})
	}

	return endpoints
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
