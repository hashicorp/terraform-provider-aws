// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_msk_serverless_cluster", name="Serverless Cluster")
// @Tags(identifierAttribute="id")
func resourceServerlessCluster() *schema.Resource {
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
			names.AttrARN: {
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
												names.AttrEnabled: {
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
			names.AttrClusterName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"cluster_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 5,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						names.AttrSubnetIDs: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	name := d.Get(names.AttrClusterName).(string)
	input := &kafka.CreateClusterV2Input{
		ClusterName: aws.String(name),
		Serverless: &types.ServerlessRequest{
			ClientAuthentication: expandServerlessClientAuthentication(d.Get("client_authentication").([]interface{})[0].(map[string]interface{})),
			VpcConfigs:           expandVpcConfigs(d.Get(names.AttrVPCConfig).([]interface{})),
		},
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateClusterV2(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Serverless Cluster (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ClusterArn))

	if _, err := waitClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Serverless Cluster (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceServerlessClusterRead(ctx, d, meta)...)
}

func resourceServerlessClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	cluster, err := findServerlessClusterByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK Serverless Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Serverless Cluster (%s): %s", d.Id(), err)
	}

	clusterARN := aws.ToString(cluster.ClusterArn)
	d.Set(names.AttrARN, clusterARN)
	if cluster.Serverless.ClientAuthentication != nil {
		if err := d.Set("client_authentication", []interface{}{flattenServerlessClientAuthentication(cluster.Serverless.ClientAuthentication)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting client_authentication: %s", err)
		}
	} else {
		d.Set("client_authentication", nil)
	}
	d.Set(names.AttrClusterName, cluster.ClusterName)
	clusterUUID, _ := clusterUUIDFromARN(clusterARN)
	d.Set("cluster_uuid", clusterUUID)
	if err := d.Set(names.AttrVPCConfig, flattenVpcConfigs(cluster.Serverless.VpcConfigs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	setTagsOut(ctx, cluster.Tags)

	return diags
}

func resourceServerlessClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceServerlessClusterRead(ctx, d, meta)...)
}

func findServerlessClusterByARN(ctx context.Context, conn *kafka.Client, arn string) (*types.Cluster, error) {
	output, err := findClusterV2ByARN(ctx, conn, arn)

	if err != nil {
		return nil, err
	}

	if output.Serverless == nil {
		return nil, tfresource.NewEmptyResultError(arn)
	}

	return output, nil
}

func expandServerlessClientAuthentication(tfMap map[string]interface{}) *types.ServerlessClientAuthentication {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ServerlessClientAuthentication{}

	if v, ok := tfMap["sasl"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Sasl = expandServerlessSasl(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandServerlessSasl(tfMap map[string]interface{}) *types.ServerlessSasl { // nosemgrep:ci.caps2-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ServerlessSasl{}

	if v, ok := tfMap["iam"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Iam = expandIam(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandIam(tfMap map[string]interface{}) *types.Iam { // nosemgrep:ci.caps4-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Iam{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func flattenServerlessClientAuthentication(apiObject *types.ServerlessClientAuthentication) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Sasl; v != nil {
		tfMap["sasl"] = []interface{}{flattenServerlessSasl(v)}
	}

	return tfMap
}

func flattenServerlessSasl(apiObject *types.ServerlessSasl) map[string]interface{} { // nosemgrep:ci.caps2-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Iam; v != nil {
		tfMap["iam"] = []interface{}{flattenIam(v)}
	}

	return tfMap
}

func flattenIam(apiObject *types.Iam) map[string]interface{} { // nosemgrep:ci.caps4-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	return tfMap
}

func expandVpcConfig(tfMap map[string]interface{}) *types.VpcConfig { // nosemgrep:ci.caps5-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VpcConfig{}

	if v, ok := tfMap[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandVpcConfigs(tfList []interface{}) []types.VpcConfig { // nosemgrep:ci.caps5-in-func-name
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.VpcConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandVpcConfig(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenVpcConfig(apiObject types.VpcConfig) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap[names.AttrSecurityGroupIDs] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = v
	}

	return tfMap
}

func flattenVpcConfigs(apiObjects []types.VpcConfig) []interface{} { // nosemgrep:ci.caps5-in-func-name
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenVpcConfig(apiObject))
	}

	return tfList
}
