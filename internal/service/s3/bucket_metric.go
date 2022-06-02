package s3

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceBucketMetric() *schema.Resource {
	return &schema.Resource{
		Create: resourceBucketMetricPut,
		Read:   resourceBucketMetricRead,
		Update: resourceBucketMetricPut,
		Delete: resourceBucketMetricDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
							Elem:         &schema.Schema{Type: schema.TypeString},
							AtLeastOneOf: filterAtLeastOneOfKeys,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBucketMetricPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn
	bucket := d.Get("bucket").(string)
	name := d.Get("name").(string)

	metricsConfiguration := &s3.MetricsConfiguration{
		Id: aws.String(name),
	}

	if v, ok := d.GetOk("filter"); ok {
		filterList := v.([]interface{})
		if filterMap, ok := filterList[0].(map[string]interface{}); ok {
			metricsConfiguration.Filter = ExpandMetricsFilter(filterMap)
		}
	}

	input := &s3.PutBucketMetricsConfigurationInput{
		Bucket:               aws.String(bucket),
		Id:                   aws.String(name),
		MetricsConfiguration: metricsConfiguration,
	}

	log.Printf("[DEBUG] Putting S3 Bucket Metrics Configuration: %s", input)
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.PutBucketMetricsConfiguration(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBucketMetricsConfiguration(input)
	}

	if err != nil {
		return fmt.Errorf("error putting S3 Bucket Metrics Configuration: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", bucket, name))

	return resourceBucketMetricRead(d, meta)
}

func resourceBucketMetricDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, name, err := BucketMetricParseID(d.Id())
	if err != nil {
		return err
	}

	input := &s3.DeleteBucketMetricsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Metrics Configuration: %s", input)
	_, err = conn.DeleteBucketMetricsConfiguration(input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchConfiguration) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket Metrics Configuration (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceBucketMetricRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, name, err := BucketMetricParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("bucket", bucket)
	d.Set("name", name)

	input := &s3.GetBucketMetricsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Reading S3 Bucket Metrics Configuration: %s", input)
	output, err := conn.GetBucketMetricsConfiguration(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Metrics Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchConfiguration) {
		log.Printf("[WARN] S3 Bucket Metrics Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Bucket Metrics Configuration (%s): %w", d.Id(), err)
	}

	if output == nil || output.MetricsConfiguration == nil {
		return fmt.Errorf("error reading S3 Bucket Metrics Configuration (%s): empty response", d.Id())
	}

	if output.MetricsConfiguration.Filter != nil {
		if err := d.Set("filter", []interface{}{FlattenMetricsFilter(output.MetricsConfiguration.Filter)}); err != nil {
			return err
		}
	}

	return nil
}

func ExpandMetricsFilter(m map[string]interface{}) *s3.MetricsFilter {
	var prefix string
	if v, ok := m["prefix"]; ok {
		prefix = v.(string)
	}

	var tags []*s3.Tag
	if v, ok := m["tags"]; ok {
		tags = Tags(tftags.New(v).IgnoreAWS())
	}

	metricsFilter := &s3.MetricsFilter{}
	if prefix != "" && len(tags) > 0 {
		metricsFilter.And = &s3.MetricsAndOperator{
			Prefix: aws.String(prefix),
			Tags:   tags,
		}
	} else if len(tags) > 1 {
		metricsFilter.And = &s3.MetricsAndOperator{
			Tags: tags,
		}
	} else if len(tags) == 1 {
		metricsFilter.Tag = tags[0]
	} else {
		metricsFilter.Prefix = aws.String(prefix)
	}
	return metricsFilter
}

func FlattenMetricsFilter(metricsFilter *s3.MetricsFilter) map[string]interface{} {
	m := make(map[string]interface{})

	if and := metricsFilter.And; and != nil {
		if and.Prefix != nil {
			m["prefix"] = aws.StringValue(and.Prefix)
		}
		if and.Tags != nil {
			m["tags"] = KeyValueTags(and.Tags).IgnoreAWS().Map()
		}
	} else if metricsFilter.Prefix != nil {
		m["prefix"] = aws.StringValue(metricsFilter.Prefix)
	} else if metricsFilter.Tag != nil {
		tags := []*s3.Tag{
			metricsFilter.Tag,
		}
		m["tags"] = KeyValueTags(tags).IgnoreAWS().Map()
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
