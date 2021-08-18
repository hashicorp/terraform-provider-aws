package aws

import (
	"bytes"
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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/appstream/waiter"
)

func resourceAwsAppStreamImageBuilder() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsAppStreamImageBuilderCreate,
		ReadWithoutTimeout:   resourceAwsAppStreamImageBuilderRead,
		DeleteWithoutTimeout: resourceAwsAppStreamImageBuilderDelete,
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
				Set: accessEndpointsHash,
			},
			"appstream_agent_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
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
				Computed: true,
				ForceNew: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
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

func resourceAwsAppStreamImageBuilderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn
	input := &appstream.CreateImageBuilderInput{
		Name:         aws.String(d.Get("name").(string)),
		InstanceType: aws.String(d.Get("instance_type").(string)),
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

	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		input.VpcConfig = expandVpcConfig(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().AppstreamTags()
	}

	var err error
	var output *appstream.CreateImageBuilderOutput
	err = resource.RetryContext(ctx, waiter.ImageBuilderOperationTimeout, func() *resource.RetryError {
		output, err = conn.CreateImageBuilderWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateImageBuilderWithContext(ctx, input)
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Appstream ImageBuilder (%s): %w", d.Id(), err))
	}

	// Start Imagebuilder workflow
	_, err = conn.StartImageBuilderWithContext(ctx, &appstream.StartImageBuilderInput{
		Name: output.ImageBuilder.Name,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error starting Appstream ImageBuilder (%s): %w", d.Id(), err))
	}

	if _, err = waiter.ImageBuilderStateRunning(ctx, conn, aws.StringValue(output.ImageBuilder.Name)); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Appstream ImageBuilder (%s) to be running: %w", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.ImageBuilder.Name))

	return resourceAwsAppStreamImageBuilderRead(ctx, d, meta)
}

func resourceAwsAppStreamImageBuilderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeImageBuildersWithContext(ctx, &appstream.DescribeImageBuildersInput{Names: []*string{aws.String(d.Id())}})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appstream ImageBuilder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Appstream ImageBuilder (%s): %w", d.Id(), err))
	}
	for _, v := range resp.ImageBuilders {

		if err = d.Set("access_endpoints", flattenAccessEndpoints(v.AccessEndpoints)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "access_endpoints", d.Id(), err))
		}
		if err = d.Set("domain_join_info", flattenDomainInfo(v.DomainJoinInfo)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "domain_join_info", d.Id(), err))
		}

		d.Set("appstream_agent_version", v.AppstreamAgentVersion)
		d.Set("arn", v.Arn)
		d.Set("created_time", aws.TimeValue(v.CreatedTime).Format(time.RFC3339))
		d.Set("description", v.Description)
		d.Set("display_name", v.DisplayName)
		d.Set("enable_default_internet_access", v.EnableDefaultInternetAccess)
		d.Set("image_arn", v.ImageArn)
		d.Set("iam_role_arn", v.IamRoleArn)

		d.Set("instance_type", v.InstanceType)
		if err = d.Set("vpc_config", flattenVpcConfig(v.VpcConfig)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting `%s` for AppStream ImageBuilder (%s): %w", "vpc_config", d.Id(), err))
		}

		d.Set("name", v.Name)
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
	}
	return nil
}

func resourceAwsAppStreamImageBuilderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).appstreamconn

	// Stop Imagebuilder workflow
	_, err := conn.StopImageBuilderWithContext(ctx, &appstream.StopImageBuilderInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error stopping Appstream ImageBuilder (%s): %w", d.Id(), err))
	}

	if _, err = waiter.ImageBuilderStateStopped(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Appstream ImageBuilder (%s) to be stopped: %w", d.Id(), err))
	}

	_, err = conn.DeleteImageBuilderWithContext(ctx, &appstream.DeleteImageBuilderInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Appstream ImageBuilder (%s): %w", d.Id(), err))
	}

	if _, err = waiter.ImageBuilderStateDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error waiting for Appstream ImageBuilder (%s) to be deleted: %w", d.Id(), err))
	}

	return nil

}

func expandAccessEndpoint(tfMap map[string]interface{}) *appstream.AccessEndpoint {
	if tfMap == nil {
		return nil
	}

	apiObject := &appstream.AccessEndpoint{
		EndpointType: aws.String(tfMap["endpoint_type"].(string)),
	}
	if v, ok := tfMap["vpce_id"]; ok {
		apiObject.VpceId = aws.String(v.(string))
	}

	return apiObject
}

func expandAccessEndpoints(tfList []interface{}) []*appstream.AccessEndpoint {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*appstream.AccessEndpoint

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandAccessEndpoint(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenAccessEndpoint(apiObject *appstream.AccessEndpoint) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["endpoint_type"] = aws.StringValue(apiObject.EndpointType)
	tfMap["vpce_id"] = aws.StringValue(apiObject.VpceId)

	return tfMap
}

func flattenAccessEndpoints(apiObjects []*appstream.AccessEndpoint) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenAccessEndpoint(apiObject))
	}

	return tfList
}

func expandDomainJoinInfo(tfList []interface{}) *appstream.DomainJoinInfo {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &appstream.DomainJoinInfo{}

	attr := tfList[0].(map[string]interface{})
	if v, ok := attr["directory_name"]; ok {
		apiObject.DirectoryName = aws.String(v.(string))
	}
	if v, ok := attr["organizational_unit_distinguished_name"]; ok {
		apiObject.OrganizationalUnitDistinguishedName = aws.String(v.(string))
	}

	return apiObject
}

func flattenDomainInfo(apiObject *appstream.DomainJoinInfo) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := map[string]interface{}{}
	tfList["directory_name"] = aws.StringValue(apiObject.DirectoryName)
	tfList["organizational_unit_distinguished_name"] = aws.StringValue(apiObject.OrganizationalUnitDistinguishedName)

	return []interface{}{tfList}
}

func expandVpcConfig(tfList []interface{}) *appstream.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &appstream.VpcConfig{}

	attr := tfList[0].(map[string]interface{})
	if v, ok := attr["security_group_ids"]; ok {
		apiObject.SecurityGroupIds = expandStringList(v.([]interface{}))
	}
	if v, ok := attr["subnet_ids"]; ok {
		apiObject.SubnetIds = expandStringList(v.([]interface{}))
	}

	return apiObject
}

func flattenVpcConfig(apiObject *appstream.VpcConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := map[string]interface{}{}
	tfList["security_group_ids"] = aws.StringValueSlice(apiObject.SecurityGroupIds)
	tfList["subnet_ids"] = aws.StringValueSlice(apiObject.SubnetIds)

	return []interface{}{tfList}
}

func accessEndpointsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["endpoint_type"].(string))
	buf.WriteString(m["vpce_id"].(string))
	return hashcode.String(buf.String())
}
