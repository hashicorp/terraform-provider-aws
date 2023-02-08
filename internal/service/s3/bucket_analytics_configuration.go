package s3

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketAnalyticsConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketAnalyticsConfigurationPut,
		ReadWithoutTimeout:   resourceBucketAnalyticsConfigurationRead,
		UpdateWithoutTimeout: resourceBucketAnalyticsConfigurationPut,
		DeleteWithoutTimeout: resourceBucketAnalyticsConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							AtLeastOneOf: filterAtLeastOneOfKeys,
						},
						"tags": {
							Type:         schema.TypeMap,
							Optional:     true,
							AtLeastOneOf: filterAtLeastOneOfKeys,
							Elem:         &schema.Schema{Type: schema.TypeString},
						},
					},
				},
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
										Type:         schema.TypeString,
										Optional:     true,
										Default:      s3.StorageClassAnalysisSchemaVersionV1,
										ValidateFunc: validation.StringInSlice([]string{s3.StorageClassAnalysisSchemaVersionV1}, false),
									},
									"destination": {
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
															"format": {
																Type:         schema.TypeString,
																Optional:     true,
																Default:      s3.AnalyticsS3ExportFileFormatCsv,
																ValidateFunc: validation.StringInSlice([]string{s3.AnalyticsS3ExportFileFormatCsv}, false),
															},
															"prefix": {
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

var filterAtLeastOneOfKeys = []string{"filter.0.prefix", "filter.0.tags"}

func resourceBucketAnalyticsConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()

	bucket := d.Get("bucket").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] S3 bucket %q, add analytics configuration %q", bucket, name)

	analyticsConfiguration := &s3.AnalyticsConfiguration{
		Id:                   aws.String(name),
		Filter:               ExpandAnalyticsFilter(d.Get("filter").([]interface{})),
		StorageClassAnalysis: ExpandStorageClassAnalysis(d.Get("storage_class_analysis").([]interface{})),
	}

	input := &s3.PutBucketAnalyticsConfigurationInput{
		Bucket:                 aws.String(bucket),
		Id:                     aws.String(name),
		AnalyticsConfiguration: analyticsConfiguration,
	}

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := conn.PutBucketAnalyticsConfigurationWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBucketAnalyticsConfigurationWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding S3 Bucket Analytics Configuration: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", bucket, name))

	return append(diags, resourceBucketAnalyticsConfigurationRead(ctx, d, meta)...)
}

func resourceBucketAnalyticsConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()

	bucket, name, err := BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Analytics Configuration (%s): %s", d.Id(), err)
	}

	d.Set("bucket", bucket)
	d.Set("name", name)

	input := &s3.GetBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Reading S3 bucket analytics configuration: %s", input)
	output, err := conn.GetBucketAnalyticsConfigurationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Analytics Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchConfiguration) {
		log.Printf("[WARN] S3 Bucket Analytics Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting S3 Bucket Analytics Configuration (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting S3 Bucket Analytics Configuration (%s): empty response", d.Id())
	}

	if err := d.Set("filter", FlattenAnalyticsFilter(output.AnalyticsConfiguration.Filter)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting filter: %s", err)
	}

	if err = d.Set("storage_class_analysis", FlattenStorageClassAnalysis(output.AnalyticsConfiguration.StorageClassAnalysis)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage class anyalytics: %s", err)
	}

	return diags
}

func resourceBucketAnalyticsConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()

	bucket, name, err := BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 analytics configuration (%s): %s", d.Id(), err)
	}

	input := &s3.DeleteBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Deleting S3 bucket analytics configuration: %s", input)
	_, err = conn.DeleteBucketAnalyticsConfigurationWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) || tfawserr.ErrMessageContains(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting S3 analytics configuration (%s): %s", d.Id(), err)
	}

	if err := WaitForDeleteBucketAnalyticsConfiguration(ctx, conn, bucket, name, 1*time.Minute); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 analytics configuration (%s): %s", d.Id(), err)
	}
	return nil
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

func ExpandAnalyticsFilter(l []interface{}) *s3.AnalyticsFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	var prefix string
	if v, ok := m["prefix"]; ok {
		prefix = v.(string)
	}

	var tags []*s3.Tag
	if v, ok := m["tags"]; ok {
		tags = Tags(tftags.New(v).IgnoreAWS())
	}

	if prefix == "" && len(tags) == 0 {
		return nil
	}
	analyticsFilter := &s3.AnalyticsFilter{}
	if prefix != "" && len(tags) > 0 {
		analyticsFilter.And = &s3.AnalyticsAndOperator{
			Prefix: aws.String(prefix),
			Tags:   tags,
		}
	} else if len(tags) > 1 {
		analyticsFilter.And = &s3.AnalyticsAndOperator{
			Tags: tags,
		}
	} else if len(tags) == 1 {
		analyticsFilter.Tag = tags[0]
	} else {
		analyticsFilter.Prefix = aws.String(prefix)
	}
	return analyticsFilter
}

func ExpandStorageClassAnalysis(l []interface{}) *s3.StorageClassAnalysis {
	result := &s3.StorageClassAnalysis{}

	if len(l) == 0 || l[0] == nil {
		return result
	}

	m := l[0].(map[string]interface{})
	if v, ok := m["data_export"]; ok {
		dataExport := &s3.StorageClassAnalysisDataExport{}
		result.DataExport = dataExport

		foo := v.([]interface{})
		if len(foo) != 0 && foo[0] != nil {
			bar := foo[0].(map[string]interface{})
			if v, ok := bar["output_schema_version"]; ok {
				dataExport.OutputSchemaVersion = aws.String(v.(string))
			}

			dataExport.Destination = expandAnalyticsExportDestination(bar["destination"].([]interface{}))
		}
	}

	return result
}

func expandAnalyticsExportDestination(edl []interface{}) *s3.AnalyticsExportDestination {
	result := &s3.AnalyticsExportDestination{}

	if len(edl) != 0 && edl[0] != nil {
		edm := edl[0].(map[string]interface{})
		result.S3BucketDestination = expandAnalyticsBucketDestination(edm["s3_bucket_destination"].([]interface{}))
	}
	return result
}

func expandAnalyticsBucketDestination(bdl []interface{}) *s3.AnalyticsS3BucketDestination {
	result := &s3.AnalyticsS3BucketDestination{}

	if len(bdl) != 0 && bdl[0] != nil {
		bdm := bdl[0].(map[string]interface{})
		result.Bucket = aws.String(bdm["bucket_arn"].(string))
		result.Format = aws.String(bdm["format"].(string))

		if v, ok := bdm["bucket_account_id"]; ok && v != "" {
			result.BucketAccountId = aws.String(v.(string))
		}

		if v, ok := bdm["prefix"]; ok && v != "" {
			result.Prefix = aws.String(v.(string))
		}
	}

	return result
}

func FlattenAnalyticsFilter(analyticsFilter *s3.AnalyticsFilter) []map[string]interface{} {
	if analyticsFilter == nil {
		return nil
	}

	result := make(map[string]interface{})
	if and := analyticsFilter.And; and != nil {
		if and.Prefix != nil {
			result["prefix"] = aws.StringValue(and.Prefix)
		}
		if and.Tags != nil {
			result["tags"] = KeyValueTags(and.Tags).IgnoreAWS().Map()
		}
	} else if analyticsFilter.Prefix != nil {
		result["prefix"] = aws.StringValue(analyticsFilter.Prefix)
	} else if analyticsFilter.Tag != nil {
		tags := []*s3.Tag{
			analyticsFilter.Tag,
		}
		result["tags"] = KeyValueTags(tags).IgnoreAWS().Map()
	} else {
		return nil
	}
	return []map[string]interface{}{result}
}

func FlattenStorageClassAnalysis(storageClassAnalysis *s3.StorageClassAnalysis) []map[string]interface{} {
	if storageClassAnalysis == nil || storageClassAnalysis.DataExport == nil {
		return []map[string]interface{}{}
	}

	dataExport := storageClassAnalysis.DataExport
	de := make(map[string]interface{})
	if dataExport.OutputSchemaVersion != nil {
		de["output_schema_version"] = aws.StringValue(dataExport.OutputSchemaVersion)
	}
	if dataExport.Destination != nil {
		de["destination"] = flattenAnalyticsExportDestination(dataExport.Destination)
	}
	result := map[string]interface{}{
		"data_export": []interface{}{de},
	}

	return []map[string]interface{}{result}
}

func flattenAnalyticsExportDestination(destination *s3.AnalyticsExportDestination) []interface{} {
	if destination == nil || destination.S3BucketDestination == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"s3_bucket_destination": flattenAnalyticsBucketDestination(destination.S3BucketDestination),
		},
	}
}

func flattenAnalyticsBucketDestination(bucketDestination *s3.AnalyticsS3BucketDestination) []interface{} {
	if bucketDestination == nil {
		return nil
	}

	result := map[string]interface{}{
		"bucket_arn": aws.StringValue(bucketDestination.Bucket),
		"format":     aws.StringValue(bucketDestination.Format),
	}
	if bucketDestination.BucketAccountId != nil {
		result["bucket_account_id"] = aws.StringValue(bucketDestination.BucketAccountId)
	}
	if bucketDestination.Prefix != nil {
		result["prefix"] = aws.StringValue(bucketDestination.Prefix)
	}

	return []interface{}{result}
}

func WaitForDeleteBucketAnalyticsConfiguration(ctx context.Context, conn *s3.S3, bucket, name string, timeout time.Duration) error {
	input := &s3.GetBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	err := resource.RetryContext(ctx, timeout, func() *resource.RetryError {
		output, err := conn.GetBucketAnalyticsConfigurationWithContext(ctx, input)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if output != nil && output.AnalyticsConfiguration != nil {
			return resource.RetryableError(fmt.Errorf("S3 bucket analytics configuration exists: %v", output))
		}

		return nil
	})

	if tfresource.TimedOut(err) { // nosemgrep:ci.helper-schema-TimeoutError-check-doesnt-return-output
		_, err = conn.GetBucketAnalyticsConfigurationWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) || tfawserr.ErrMessageContains(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket Analytics Configuration \"%s:%s\": %w", bucket, name, err)
	}

	return nil
}
