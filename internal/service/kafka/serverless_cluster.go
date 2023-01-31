package kafka

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceServerlessCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServerlessClusterCreate,
		ReadWithoutTimeout:   resourceServerlessClusterRead,
		UpdateWithoutTimeout: resourceServerlessClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_authentication": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sasl": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"iam": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							MaxItems: 5,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceServerlessClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("cluster_name").(string)
	input := &kafka.CreateClusterV2Input{
		ClusterName: aws.String(name),
		Serverless: &kafka.ServerlessRequest{
			ClientAuthentication: expandServerlessClientAuthentication(d.Get("client_authentication").([]interface{})[0].(map[string]interface{})),
			VpcConfigs:           expandVpcConfigs(d.Get("vpc_config").([]interface{})),
		},
		Tags: Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Creating MSK Serverless Cluster: %s", input)
	output, err := conn.CreateClusterV2WithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating MSK Serverless Cluster (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ClusterArn))

	_, err = waitClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("waiting for MSK Serverless Cluster (%s) create: %s", d.Id(), err)
	}

	return resourceServerlessClusterRead(ctx, d, meta)
}

func resourceServerlessClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	cluster, err := FindServerlessClusterByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK Serverless Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading MSK Serverless Cluster (%s): %s", d.Id(), err)
	}

	d.Set("arn", cluster.ClusterArn)
	if cluster.Serverless.ClientAuthentication != nil {
		if err := d.Set("client_authentication", []interface{}{flattenServerlessClientAuthentication(cluster.Serverless.ClientAuthentication)}); err != nil {
			return diag.Errorf("setting client_authentication: %s", err)
		}
	} else {
		d.Set("client_authentication", nil)
	}
	d.Set("cluster_name", cluster.ClusterName)
	if err := d.Set("vpc_config", flattenVpcConfigs(cluster.Serverless.VpcConfigs)); err != nil {
		return diag.Errorf("setting vpc_config: %s", err)
	}

	tags := KeyValueTags(cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceServerlessClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating MSK Serverless Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceServerlessClusterRead(ctx, d, meta)
}

func expandServerlessClientAuthentication(tfMap map[string]interface{}) *kafka.ServerlessClientAuthentication {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.ServerlessClientAuthentication{}

	if v, ok := tfMap["sasl"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Sasl = expandServerlessSasl(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandServerlessSasl(tfMap map[string]interface{}) *kafka.ServerlessSasl { // nosemgrep:ci.caps2-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.ServerlessSasl{}

	if v, ok := tfMap["iam"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Iam = expandIam(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandIam(tfMap map[string]interface{}) *kafka.Iam { // nosemgrep:ci.caps4-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.Iam{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func flattenServerlessClientAuthentication(apiObject *kafka.ServerlessClientAuthentication) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Sasl; v != nil {
		tfMap["sasl"] = []interface{}{flattenServerlessSasl(v)}
	}

	return tfMap
}

func flattenServerlessSasl(apiObject *kafka.ServerlessSasl) map[string]interface{} { // nosemgrep:ci.caps2-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Iam; v != nil {
		tfMap["iam"] = []interface{}{flattenIam(v)}
	}

	return tfMap
}

func flattenIam(apiObject *kafka.Iam) map[string]interface{} { // nosemgrep:ci.caps4-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	return tfMap
}

func expandVpcConfig(tfMap map[string]interface{}) *kafka.VpcConfig { // nosemgrep:ci.caps5-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.VpcConfig{}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	return apiObject
}

func expandVpcConfigs(tfList []interface{}) []*kafka.VpcConfig { // nosemgrep:ci.caps5-in-func-name
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*kafka.VpcConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandVpcConfig(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenVpcConfig(apiObject *kafka.VpcConfig) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap["security_group_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap["subnet_ids"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenVpcConfigs(apiObjects []*kafka.VpcConfig) []interface{} { // nosemgrep:ci.caps5-in-func-name
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenVpcConfig(apiObject))
	}

	return tfList
}
