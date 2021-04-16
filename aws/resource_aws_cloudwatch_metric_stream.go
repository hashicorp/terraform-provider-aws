package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCloudWatchMetricStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchMetricStreamPut,
		Read:   resourceAwsCloudWatchMetricStreamRead,
		Update: resourceAwsCloudWatchMetricStreamPut,
		Delete: resourceAwsCloudWatchMetricStreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exclude_filter": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"include_filter"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"firehose_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"include_filter": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"exclude_filter"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"last_update_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateCloudWatchMetricStreamName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateCloudWatchMetricStreamName,
			},
			"output_format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsCloudWatchMetricStreamRead(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	log.Printf("[DEBUG] Reading CloudWatch MetricStream: %s", name)
	conn := meta.(*AWSClient).cloudwatchconn

	params := cloudwatch.GetMetricStreamInput{
		Name: aws.String(d.Id()),
	}

	resp, err := conn.GetMetricStream(&params)
	if err != nil {
		if isAWSErr(err, cloudwatch.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] CloudWatch MetricStream %q not found, removing", name)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading metric_stream failed: %s", err)
	}

	d.Set("arn", resp.Arn)
	d.Set("creation_date", resp.CreationDate.Format(time.RFC3339))
	d.Set("firehose_arn", resp.FirehoseArn)
	d.Set("last_update_date", resp.CreationDate.Format(time.RFC3339))
	d.Set("name", resp.Name)
	d.Set("output_format", resp.OutputFormat)
	d.Set("role_arn", resp.RoleArn)
	d.Set("state", resp.State)

	if resp.IncludeFilters != nil && len(resp.IncludeFilters) > 0 {
		includeFilters := make([]interface{}, len(resp.IncludeFilters))
		for i, mq := range resp.IncludeFilters {
			includeFilter := map[string]interface{}{
				"namespace": aws.StringValue(mq.Namespace),
			}
			includeFilters[i] = includeFilter
		}
		if err := d.Set("include_filter", includeFilters); err != nil {
			return fmt.Errorf("error setting include_filter: %s", err)
		}
	}

	if resp.ExcludeFilters != nil && len(resp.ExcludeFilters) > 0 {
		excludeFilters := make([]interface{}, len(resp.ExcludeFilters))
		for i, mq := range resp.ExcludeFilters {
			excludeFilter := map[string]interface{}{
				"namespace": aws.StringValue(mq.Namespace),
			}
			excludeFilters[i] = excludeFilter
		}
		if err := d.Set("exclude_filter", excludeFilters); err != nil {
			return fmt.Errorf("error setting exclude_filter: %s", err)
		}
	}

	return nil
}

func resourceAwsCloudWatchMetricStreamCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	params := cloudwatch.PutMetricStreamInput{
		Name:         aws.String(name),
		FirehoseArn:  aws.String(d.Get("firehose_arn").(string)),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		OutputFormat: aws.String(d.Get("output_format").(string)),
		Tags:         keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CloudwatchTags(),
	}

	if v, ok := d.GetOk("include_filter"); ok {
		var includeFilters []*cloudwatch.MetricStreamFilter
		for _, v := range v.(*schema.Set).List() {
			metricStreamFilterResource := v.(map[string]interface{})
			namespace := metricStreamFilterResource["namespace"].(string)
			metricStreamFilter := cloudwatch.MetricStreamFilter{
				Namespace: aws.String(namespace),
			}
			includeFilters = append(includeFilters, &metricStreamFilter)
		}
		params.IncludeFilters = includeFilters
	}

	if v, ok := d.GetOk("exclude_filter"); ok {
		var excludeFilters []*cloudwatch.MetricStreamFilter
		for _, v := range v.(*schema.Set).List() {
			metricStreamFilterResource := v.(map[string]interface{})
			namespace := metricStreamFilterResource["namespace"].(string)
			metricStreamFilter := cloudwatch.MetricStreamFilter{
				Namespace: aws.String(namespace),
			}
			excludeFilters = append(excludeFilters, &metricStreamFilter)
		}
		params.ExcludeFilters = excludeFilters
	}

	log.Printf("[DEBUG] Putting CloudWatch MetricStream: %#v", params)

	_, err := conn.PutMetricStream(&params)
	if err != nil {
		return fmt.Errorf("Putting metric_stream failed: %s", err)
	}
	d.SetId(name)
	log.Println("[INFO] CloudWatch MetricStream put finished")

	return resourceAwsCloudWatchMetricStreamRead(d, meta)
}

func resourceAwsCloudWatchMetricStreamDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Deleting CloudWatch MetricStream %s", d.Id())
	conn := meta.(*AWSClient).cloudwatchconn
	params := cloudwatch.DeleteMetricStreamInput{
		Name: aws.String(d.Id()),
	}

	if _, err := conn.DeleteMetricStream(&params); err != nil {
		return fmt.Errorf("Error deleting CloudWatch MetricStream: %s", err)
	}
	log.Printf("[INFO] CloudWatch MetricStream %s deleted", d.Id())

	return nil
}

func validateCloudWatchMetricStreamName(v interface{}, k string) (ws []string, errors []error) {
	return validation.All(
		validation.StringLenBetween(1, 255),
		validation.StringMatch(regexp.MustCompile(`^[\-_A-Za-z0-9]*$`), "must match [\\-_A-Za-z0-9]"),
	)(v, k)
}
