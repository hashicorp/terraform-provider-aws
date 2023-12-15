// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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

const (
	serverlesClusterCreatedTimeout = 15 * time.Minute
)

// @SDKResource("aws_elasticache_serverless", name="ElastiCache Serverless")
// @Tags(identifierAttribute="arn")
func ResourceElasticacheServerless() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceElasticacheServerlessCreate,
		ReadWithoutTimeout:   resourceElasticacheServerlessRead,
		UpdateWithoutTimeout: resourceElasticacheServerlessUpdate,
		DeleteWithoutTimeout: resourceElasticacheServerlessDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Required: true,
			},
			"serverless_cache_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cache_usage_limits": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_storage": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maximum": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"unit": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"ecpu_per_second": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maximum": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"daily_snapshot_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				//MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"full_engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reader_endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				//MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"major_engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"snapshot_arns": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						verify.ValidARN,
						validation.StringDoesNotContainAny(","),
					),
				},
			},
			"snapshot_retention_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtMost(35),
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"user_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceElasticacheServerlessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	name := d.Get("serverless_cache_name").(string)
	engine := d.Get("engine").(string)

	input := &elasticache.CreateServerlessCacheInput{
		ServerlessCacheName: aws.String(name),
		Engine:              aws.String(engine),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk("cache_usage_limits"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CacheUsageLimits = expandCacheUsageLimits(v.([]interface{}))
	}

	if v, ok := d.GetOk("daily_snapshot_time"); ok {
		input.DailySnapshotTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("major_engine_version"); ok {
		input.MajorEngineVersion = aws.String(v.(string))
	}

	if SGIds := d.Get("security_group_ids").(*schema.Set); SGIds.Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringSet(SGIds)
	}

	snaps := d.Get("snapshot_arns").([]interface{})
	if len(snaps) > 0 {
		input.SnapshotArnsToRestore = flex.ExpandStringList(snaps)
		log.Printf("[DEBUG] Restoring Redis cluster from S3 snapshot: %#v", snaps)
	}

	if v, ok := d.GetOk("snapshot_retention_limit"); ok {
		input.SnapshotRetentionLimit = aws.Int64(int64(v.(int)))
	}

	if SubnetIds := d.Get("subnet_ids").(*schema.Set); SubnetIds.Len() > 0 {
		input.SubnetIds = flex.ExpandStringSet(SubnetIds)
	}

	if v, ok := d.GetOk("user_group_id"); ok {
		input.UserGroupId = aws.String(v.(string))
	}

	output, err := conn.CreateServerlessCacheWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Serverless ElastiCache  (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ServerlessCache.ServerlessCacheName))

	if _, err := waitServerlesssCacheAvailable(ctx, conn, d.Id(), serverlesClusterCreatedTimeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache Cache Cluster (%s) create: %s", d.Id(), err)
	}
	return append(diags, resourceElasticacheServerlessRead(ctx, d, meta)...)
}

func resourceElasticacheServerlessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	output, err := FindElasicCacheServerlessByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Serverless ElastiCache (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Serverless ElastiCache (%s): %s", d.Id(), err)
	}

	d.Set("engine", output.Engine)
	d.Set("serverless_cache_name", output.ServerlessCacheName)
	d.Set("arn", output.ARN)

	d.Set("description", output.Description)
	d.Set("subnet_ids", aws.StringValueSlice(output.SubnetIds))
	if err := d.Set("cache_usage_limits", flattenCacheUsageLimits(output.CacheUsageLimits)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cache_usage_limits: %s", err)
	}

	d.Set("daily_snapshot_time", output.DailySnapshotTime)

	d.Set("kms_key_id", output.KmsKeyId)
	d.Set("major_engine_version", output.MajorEngineVersion)

	d.Set("security_group_ids", flex.FlattenStringSet(output.SecurityGroupIds))
	d.Set("snapshot_retention_limit", output.SnapshotRetentionLimit)

	d.Set("user_group_id", output.UserGroupId)

	if output.CreateTime != nil {
		d.Set("create_time", aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	} else {
		d.Set("create_time", nil)
	}

	if err := d.Set("endpoint", flattenEndpoint(output.Endpoint)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}

	d.Set("full_engine_version", output.FullEngineVersion)
	if err := d.Set("reader_endpoint", flattenEndpoint(output.ReaderEndpoint)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting reader_endpoint: %s", err)
	}

	d.Set("status", output.Status)

	return diags

}

func resourceElasticacheServerlessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	if d.HasChanges("daily_snapshot_time") {
		input := &elasticache.ModifyServerlessCacheInput{

			DailySnapshotTime:   aws.String(d.Get("daily_snapshot_time").(string)),
			ServerlessCacheName: aws.String(d.Get("serverless_cache_name").(string)),
		}

		_, err := conn.ModifyServerlessCacheWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Serverless ElastiCache (%s): %s", d.Id(), err)
		}

		_, err = waitServerlesssCacheAvailable(ctx, conn, d.Id(), ServerlessCacheUpdatedTimeout)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Serverless ElastiCache Cluster (%s) to update: %s", d.Id(), err)
		}
	}

	if d.HasChanges("description") {
		input := &elasticache.ModifyServerlessCacheInput{

			Description:         aws.String(d.Get("description").(string)),
			ServerlessCacheName: aws.String(d.Get("serverless_cache_name").(string)),
		}

		_, err := conn.ModifyServerlessCacheWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Serverless ElastiCache (%s): %s", d.Id(), err)
		}

		_, err = waitServerlesssCacheAvailable(ctx, conn, d.Id(), ServerlessCacheUpdatedTimeout)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Serverless ElastiCache Cluster (%s) to update: %s", d.Id(), err)
		}
	}

	if d.HasChange("cache_usage_limits") {
		input := &elasticache.ModifyServerlessCacheInput{

			CacheUsageLimits:    expandCacheUsageLimits(d.Get("cache_usage_limits").([]interface{})),
			ServerlessCacheName: aws.String(d.Get("serverless_cache_name").(string)),
		}

		_, err := conn.ModifyServerlessCacheWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Serverless ElastiCache (%s): %s", d.Id(), err)
		}

	}

	if d.HasChanges("security_group_ids") {
		input := &elasticache.ModifyServerlessCacheInput{

			SecurityGroupIds:    flex.ExpandStringSet(d.Get("security_group_ids").(*schema.Set)),
			ServerlessCacheName: aws.String(d.Get("serverless_cache_name").(string)),
		}

		_, err := conn.ModifyServerlessCacheWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Serverless ElastiCache (%s): %s", d.Id(), err)
		}

		_, err = waitServerlesssCacheAvailable(ctx, conn, d.Id(), ServerlessCacheUpdatedTimeout)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Serverless ElastiCache Cluster (%s) to update: %s", d.Id(), err)
		}
	}

	if d.HasChanges("snapshot_retention_limit") {
		input := &elasticache.ModifyServerlessCacheInput{

			SnapshotRetentionLimit: aws.Int64(int64(d.Get("snapshot_retention_limit").(int))),
			ServerlessCacheName:    aws.String(d.Get("serverless_cache_name").(string)),
		}

		_, err := conn.ModifyServerlessCacheWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Serverless ElastiCache (%s): %s", d.Id(), err)
		}

		_, err = waitServerlesssCacheAvailable(ctx, conn, d.Id(), ServerlessCacheUpdatedTimeout)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Serverless ElastiCache Cluster (%s) to update: %s", d.Id(), err)
		}
	}

	if d.HasChanges("user_group_id") {
		input := &elasticache.ModifyServerlessCacheInput{

			UserGroupId:         aws.String(d.Get("user_group_id").(string)),
			ServerlessCacheName: aws.String(d.Get("serverless_cache_name").(string)),
		}

		_, err := conn.ModifyServerlessCacheWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Serverless ElastiCache (%s): %s", d.Id(), err)
		}

		_, err = waitServerlesssCacheAvailable(ctx, conn, d.Id(), ServerlessCacheUpdatedTimeout)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Serverless ElastiCache Cluster (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceElasticacheServerlessRead(ctx, d, meta)...)
}

func resourceElasticacheServerlessDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	log.Printf("[DEBUG] Deleting Sserverless ElastiCache: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute, func() (interface{}, error) {
		return conn.DeleteServerlessCacheWithContext(ctx, &elasticache.DeleteServerlessCacheInput{
			ServerlessCacheName: aws.String(d.Id()),
		})
	}, "DependencyViolation")

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeServerlessCacheNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Serverless ElastiCache (%s): %s", d.Id(), err)
	}

	_, err = waitServerlesssCacheDeleted(ctx, conn, d.Id(), ServerlessCacheDeletedTimeout)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Serverless Cache Cluster (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func expandCacheUsageLimits(tfMap []interface{}) *elasticache.CacheUsageLimits {
	if tfMap == nil {
		return nil
	}

	apiObject := &elasticache.CacheUsageLimits{}
	tfMapRaw := tfMap[0].(map[string]interface{})

	if v, ok := tfMapRaw["data_storage"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataStorage = expandDataStorage(v)
	}

	if v, ok := tfMapRaw["ecpu_per_second"].([]interface{}); ok && len(v) > 0 {
		apiObject.ECPUPerSecond = expandECPUPerSecond(v)
	}

	return apiObject
}

func expandDataStorage(tfMap []interface{}) *elasticache.DataStorage {
	if tfMap == nil {
		return nil
	}

	m := tfMap[0].(map[string]interface{})
	apiObject := &elasticache.DataStorage{
		Maximum: aws.Int64(int64(m["maximum"].(int))),
		Unit:    aws.String(m["unit"].(string)),
	}

	return apiObject
}

func expandECPUPerSecond(tfMap []interface{}) *elasticache.ECPUPerSecond {
	if tfMap == nil {
		return nil
	}

	m := tfMap[0].(map[string]interface{})
	apiObject := &elasticache.ECPUPerSecond{
		Maximum: aws.Int64(int64(m["maximum"].(int))),
	}

	return apiObject
}

func flattenCacheUsageLimits(apiObject *elasticache.CacheUsageLimits) []map[string]interface{} {

	if apiObject == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"data_storage":    flattenDataSorage(apiObject.DataStorage),
		"ecpu_per_second": flattenECPUSecond(apiObject.ECPUPerSecond),
	}

	return []map[string]interface{}{m}
}

func flattenDataSorage(apiObjects *elasticache.DataStorage) []map[string]interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []map[string]interface{}

	flattenedObject := map[string]interface{}{}
	if v := apiObjects.Maximum; v != nil {
		flattenedObject["maximum"] = aws.Int64Value(v)
	}

	if v := apiObjects.Unit; v != nil {
		flattenedObject["unit"] = aws.StringValue(v)
	}

	tfList = append(tfList, flattenedObject)

	return tfList
}

func flattenECPUSecond(apiObjects *elasticache.ECPUPerSecond) []map[string]interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []map[string]interface{}

	flattenedObject := map[string]interface{}{}
	if v := apiObjects.Maximum; v != nil {
		flattenedObject["maximum"] = aws.Int64Value(v)
	}

	tfList = append(tfList, flattenedObject)

	return tfList
}

func flattenEndpoint(apiObject *elasticache.Endpoint) []map[string]interface{} {

	if apiObject == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"address": aws.StringValue(apiObject.Address),
		"port":    aws.Int64Value(apiObject.Port),
	}

	return []map[string]interface{}{m}
}
