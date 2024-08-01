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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_analytics_configuration", name="Bucket Analytics Configuration")
func resourceBucketAnalyticsConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketAnalyticsConfigurationPut,
		ReadWithoutTimeout:   resourceBucketAnalyticsConfigurationRead,
		UpdateWithoutTimeout: resourceBucketAnalyticsConfigurationPut,
		DeleteWithoutTimeout: resourceBucketAnalyticsConfigurationDelete,

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
						names.AttrPrefix: {
							Type:         schema.TypeString,
							Optional:     true,
							AtLeastOneOf: []string{"filter.0.prefix", "filter.0.tags"},
						},
						names.AttrTags: {
							Type:         schema.TypeMap,
							Optional:     true,
							Elem:         &schema.Schema{Type: schema.TypeString},
							AtLeastOneOf: []string{"filter.0.prefix", "filter.0.tags"},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_class_analysis": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_export": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"output_schema_version": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          types.StorageClassAnalysisSchemaVersionV1,
										ValidateDiagFunc: enum.Validate[types.StorageClassAnalysisSchemaVersion](),
									},
									names.AttrDestination: {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_bucket_destination": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: verify.ValidARN,
															},
															"bucket_account_id": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: verify.ValidAccountID,
															},
															names.AttrFormat: {
																Type:             schema.TypeString,
																Optional:         true,
																Default:          types.AnalyticsS3ExportFileFormatCsv,
																ValidateDiagFunc: enum.Validate[types.AnalyticsS3ExportFileFormat](),
															},
															names.AttrPrefix: {
																Type:     schema.TypeString,
																Optional: true,
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
					},
				},
			},
		},
	}
}

func resourceBucketAnalyticsConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	name := d.Get(names.AttrName).(string)
	analyticsConfiguration := &types.AnalyticsConfiguration{
		Id:                   aws.String(name),
		StorageClassAnalysis: expandStorageClassAnalysis(d.Get("storage_class_analysis").([]interface{})),
	}

	if v, ok := d.GetOk(names.AttrFilter); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		analyticsConfiguration.Filter = expandAnalyticsFilter(ctx, v.([]interface{})[0].(map[string]interface{}))
	}

	bucket := d.Get(names.AttrBucket).(string)
	input := &s3.PutBucketAnalyticsConfigurationInput{
		Bucket:                 aws.String(bucket),
		Id:                     aws.String(name),
		AnalyticsConfiguration: analyticsConfiguration,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketAnalyticsConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "AnalyticsConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Analytics Configuration (%s): %s", bucket, name, err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("%s:%s", bucket, name))

		_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
			return findAnalyticsConfiguration(ctx, conn, bucket, name)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Analytics Configuration (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBucketAnalyticsConfigurationRead(ctx, d, meta)...)
}

func resourceBucketAnalyticsConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, name, err := BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	ac, err := findAnalyticsConfiguration(ctx, conn, bucket, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Analytics Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Analytics Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	if err := d.Set(names.AttrFilter, flattenAnalyticsFilter(ctx, ac.Filter)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting filter: %s", err)
	}
	d.Set(names.AttrName, name)
	if err = d.Set("storage_class_analysis", flattenStorageClassAnalysis(ac.StorageClassAnalysis)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_class_analysis: %s", err)
	}

	return diags
}

func resourceBucketAnalyticsConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, name, err := BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Analytics Configuration: %s", d.Id())
	_, err = conn.DeleteBucketAnalyticsConfiguration(ctx, &s3.DeleteBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Analytics Configuration (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findAnalyticsConfiguration(ctx, conn, bucket, name)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Analytics Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func BucketAnalyticsConfigurationParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("please make sure the ID is in the form BUCKET:NAME (i.e. my-bucket:EntireBucket")
	}
	bucket := idParts[0]
	name := idParts[1]
	return bucket, name, nil
}

func expandAnalyticsFilter(ctx context.Context, m map[string]interface{}) types.AnalyticsFilter {
	var prefix string
	if v, ok := m[names.AttrPrefix]; ok {
		prefix = v.(string)
	}

	var tags []types.Tag
	if v, ok := m[names.AttrTags]; ok {
		tags = Tags(tftags.New(ctx, v).IgnoreAWS())
	}

	if prefix == "" && len(tags) == 0 {
		return nil
	}

	var analyticsFilter types.AnalyticsFilter

	if prefix != "" && len(tags) > 0 {
		analyticsFilter = &types.AnalyticsFilterMemberAnd{
			Value: types.AnalyticsAndOperator{
				Prefix: aws.String(prefix),
				Tags:   tags,
			},
		}
	} else if len(tags) > 1 {
		analyticsFilter = &types.AnalyticsFilterMemberAnd{
			Value: types.AnalyticsAndOperator{
				Tags: tags,
			},
		}
	} else if len(tags) == 1 {
		analyticsFilter = &types.AnalyticsFilterMemberTag{
			Value: tags[0],
		}
	} else {
		analyticsFilter = &types.AnalyticsFilterMemberPrefix{
			Value: prefix,
		}
	}
	return analyticsFilter
}

func expandStorageClassAnalysis(l []interface{}) *types.StorageClassAnalysis {
	result := &types.StorageClassAnalysis{}

	if len(l) == 0 || l[0] == nil {
		return result
	}

	m := l[0].(map[string]interface{})
	if v, ok := m["data_export"]; ok {
		dataExport := &types.StorageClassAnalysisDataExport{}
		result.DataExport = dataExport

		foo := v.([]interface{})
		if len(foo) != 0 && foo[0] != nil {
			bar := foo[0].(map[string]interface{})
			if v, ok := bar["output_schema_version"]; ok {
				dataExport.OutputSchemaVersion = types.StorageClassAnalysisSchemaVersion(v.(string))
			}

			dataExport.Destination = expandAnalyticsExportDestination(bar[names.AttrDestination].([]interface{}))
		}
	}

	return result
}

func expandAnalyticsExportDestination(edl []interface{}) *types.AnalyticsExportDestination {
	result := &types.AnalyticsExportDestination{}

	if len(edl) != 0 && edl[0] != nil {
		edm := edl[0].(map[string]interface{})
		result.S3BucketDestination = expandAnalyticsS3BucketDestination(edm["s3_bucket_destination"].([]interface{}))
	}
	return result
}

func expandAnalyticsS3BucketDestination(bdl []interface{}) *types.AnalyticsS3BucketDestination { // nosemgrep:ci.s3-in-func-name
	result := &types.AnalyticsS3BucketDestination{}

	if len(bdl) != 0 && bdl[0] != nil {
		bdm := bdl[0].(map[string]interface{})
		result.Bucket = aws.String(bdm["bucket_arn"].(string))
		result.Format = types.AnalyticsS3ExportFileFormat(bdm[names.AttrFormat].(string))

		if v, ok := bdm["bucket_account_id"]; ok && v != "" {
			result.BucketAccountId = aws.String(v.(string))
		}

		if v, ok := bdm[names.AttrPrefix]; ok && v != "" {
			result.Prefix = aws.String(v.(string))
		}
	}

	return result
}

func flattenAnalyticsFilter(ctx context.Context, analyticsFilter types.AnalyticsFilter) []map[string]interface{} {
	result := make(map[string]interface{})

	switch v := analyticsFilter.(type) {
	case *types.AnalyticsFilterMemberAnd:
		if v := v.Value.Prefix; v != nil {
			result[names.AttrPrefix] = aws.ToString(v)
		}
		if v := v.Value.Tags; v != nil {
			result[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
		}
	case *types.AnalyticsFilterMemberPrefix:
		result[names.AttrPrefix] = v.Value
	case *types.AnalyticsFilterMemberTag:
		tags := []types.Tag{
			v.Value,
		}
		result[names.AttrTags] = keyValueTags(ctx, tags).IgnoreAWS().Map()
	default:
		return nil
	}

	return []map[string]interface{}{result}
}

func flattenStorageClassAnalysis(storageClassAnalysis *types.StorageClassAnalysis) []map[string]interface{} {
	if storageClassAnalysis == nil || storageClassAnalysis.DataExport == nil {
		return []map[string]interface{}{}
	}

	dataExport := storageClassAnalysis.DataExport
	de := map[string]interface{}{
		"output_schema_version": dataExport.OutputSchemaVersion,
	}
	if dataExport.Destination != nil {
		de[names.AttrDestination] = flattenAnalyticsExportDestination(dataExport.Destination)
	}
	result := map[string]interface{}{
		"data_export": []interface{}{de},
	}

	return []map[string]interface{}{result}
}

func flattenAnalyticsExportDestination(destination *types.AnalyticsExportDestination) []interface{} {
	if destination == nil || destination.S3BucketDestination == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"s3_bucket_destination": flattenAnalyticsS3BucketDestination(destination.S3BucketDestination),
		},
	}
}

func flattenAnalyticsS3BucketDestination(bucketDestination *types.AnalyticsS3BucketDestination) []interface{} { // nosemgrep:ci.s3-in-func-name
	if bucketDestination == nil {
		return nil
	}

	result := map[string]interface{}{
		"bucket_arn":     aws.ToString(bucketDestination.Bucket),
		names.AttrFormat: bucketDestination.Format,
	}
	if bucketDestination.BucketAccountId != nil {
		result["bucket_account_id"] = aws.ToString(bucketDestination.BucketAccountId)
	}
	if bucketDestination.Prefix != nil {
		result[names.AttrPrefix] = aws.ToString(bucketDestination.Prefix)
	}

	return []interface{}{result}
}

func findAnalyticsConfiguration(ctx context.Context, conn *s3.Client, bucket, id string) (*types.AnalyticsConfiguration, error) {
	input := &s3.GetBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
	}

	output, err := conn.GetBucketAnalyticsConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AnalyticsConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AnalyticsConfiguration, nil
}
