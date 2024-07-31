// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_metric", name="Bucket Metric")
func resourceBucketMetric() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketMetricPut,
		ReadWithoutTimeout:   resourceBucketMetricRead,
		UpdateWithoutTimeout: resourceBucketMetricPut,
		DeleteWithoutTimeout: resourceBucketMetricDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrFilter: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_point": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
							AtLeastOneOf: []string{"filter.0.access_point", "filter.0.prefix", "filter.0.tags"},
						},
						names.AttrPrefix: {
							Type:         schema.TypeString,
							Optional:     true,
							AtLeastOneOf: []string{"filter.0.access_point", "filter.0.prefix", "filter.0.tags"},
						},
						names.AttrTags: {
							Type:         schema.TypeMap,
							Optional:     true,
							Elem:         &schema.Schema{Type: schema.TypeString},
							AtLeastOneOf: []string{"filter.0.access_point", "filter.0.prefix", "filter.0.tags"},
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceBucketMetricPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	name := d.Get(names.AttrName).(string)
	metricsConfiguration := &types.MetricsConfiguration{
		Id: aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		if tfMap, ok := v.([]interface{})[0].(map[string]interface{}); ok {
			metricsConfiguration.Filter = expandMetricsFilter(ctx, tfMap)
		}
	}

	bucket := d.Get(names.AttrBucket).(string)
	input := &s3.PutBucketMetricsConfigurationInput{
		Bucket:               aws.String(bucket),
		Id:                   aws.String(name),
		MetricsConfiguration: metricsConfiguration,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketMetricsConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "MetricsConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Metric: %s", bucket, err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("%s:%s", bucket, name))

		_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
			return findMetricsConfiguration(ctx, conn, bucket, name)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Metric (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBucketMetricRead(ctx, d, meta)...)
}

func resourceBucketMetricRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, name, err := BucketMetricParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	mc, err := findMetricsConfiguration(ctx, conn, bucket, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Metric (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Metric (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	if mc.Filter != nil {
		if err := d.Set(names.AttrFilter, []interface{}{flattenMetricsFilter(ctx, mc.Filter)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting filter")
		}
	}
	d.Set(names.AttrName, name)

	return diags
}

func resourceBucketMetricDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, name, err := BucketMetricParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Metric: %s", d.Id())
	_, err = conn.DeleteBucketMetricsConfiguration(ctx, &s3.DeleteBucketMetricsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Metric (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findMetricsConfiguration(ctx, conn, bucket, name)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Metric (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandMetricsFilter(ctx context.Context, m map[string]interface{}) types.MetricsFilter {
	var accessPoint string
	if v, ok := m["access_point"]; ok {
		accessPoint = v.(string)
	}

	var prefix string
	if v, ok := m[names.AttrPrefix]; ok {
		prefix = v.(string)
	}

	var tags []types.Tag
	if v, ok := m[names.AttrTags]; ok {
		tags = Tags(tftags.New(ctx, v).IgnoreAWS())
	}

	var metricsFilter types.MetricsFilter

	if accessPoint != "" && prefix != "" && len(tags) > 0 {
		metricsFilter = &types.MetricsFilterMemberAnd{
			Value: types.MetricsAndOperator{
				AccessPointArn: aws.String(accessPoint),
				Prefix:         aws.String(prefix),
				Tags:           tags,
			},
		}
	} else if accessPoint != "" && prefix != "" {
		metricsFilter = &types.MetricsFilterMemberAnd{
			Value: types.MetricsAndOperator{
				AccessPointArn: aws.String(accessPoint),
				Prefix:         aws.String(prefix),
			},
		}
	} else if accessPoint != "" && len(tags) > 0 {
		metricsFilter = &types.MetricsFilterMemberAnd{
			Value: types.MetricsAndOperator{
				AccessPointArn: aws.String(accessPoint),
				Tags:           tags,
			},
		}
	} else if prefix != "" && len(tags) > 0 {
		metricsFilter = &types.MetricsFilterMemberAnd{
			Value: types.MetricsAndOperator{
				Prefix: aws.String(prefix),
				Tags:   tags,
			},
		}
	} else if len(tags) > 1 {
		metricsFilter = &types.MetricsFilterMemberAnd{
			Value: types.MetricsAndOperator{
				Tags: tags,
			},
		}
	} else if len(tags) == 1 {
		metricsFilter = &types.MetricsFilterMemberTag{
			Value: tags[0],
		}
	} else if accessPoint != "" {
		metricsFilter = &types.MetricsFilterMemberAccessPointArn{
			Value: accessPoint,
		}
	} else {
		metricsFilter = &types.MetricsFilterMemberPrefix{
			Value: prefix,
		}
	}
	return metricsFilter
}

func flattenMetricsFilter(ctx context.Context, metricsFilter types.MetricsFilter) map[string]interface{} {
	m := make(map[string]interface{})

	switch v := metricsFilter.(type) {
	case *types.MetricsFilterMemberAnd:
		if v := v.Value.AccessPointArn; v != nil {
			m["access_point"] = aws.ToString(v)
		}
		if v := v.Value.Prefix; v != nil {
			m[names.AttrPrefix] = aws.ToString(v)
		}
		if v := v.Value.Tags; v != nil {
			m[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
		}
	case *types.MetricsFilterMemberAccessPointArn:
		m["access_point"] = v.Value
	case *types.MetricsFilterMemberPrefix:
		m[names.AttrPrefix] = v.Value
	case *types.MetricsFilterMemberTag:
		tags := []types.Tag{
			v.Value,
		}
		m[names.AttrTags] = keyValueTags(ctx, tags).IgnoreAWS().Map()
	default:
		return nil
	}
	return m
}

func BucketMetricParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("please make sure the ID is in the form BUCKET:NAME (i.e. my-bucket:EntireBucket")
	}
	bucket := idParts[0]
	name := idParts[1]
	return bucket, name, nil
}

func findMetricsConfiguration(ctx context.Context, conn *s3.Client, bucket, id string) (*types.MetricsConfiguration, error) {
	input := &s3.GetBucketMetricsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
	}

	output, err := conn.GetBucketMetricsConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.MetricsConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.MetricsConfiguration, nil
}
