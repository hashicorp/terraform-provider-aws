package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"reflect"
	"time"
)

func resourceAwsKinesisAnalyticsApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisAnalyticsApplicationCreate,
		Read:   resourceAwsKinesisAnalyticsApplicationRead,
		Update: resourceAwsKinesisAnalyticsApplicationUpdate,
		Delete: resourceAwsKinesisAnalyticsApplicationDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"code": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"create_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"last_update_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"cloudwatch_logging_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"log_stream": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},

						"role": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"inputs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},

			"outputs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
		},
	}
}

func resourceAwsKinesisAnalyticsApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn
	name := d.Get("name").(string)
	createOpts := &kinesisanalytics.CreateApplicationInput{
		ApplicationName: aws.String(name),
	}

	if v, ok := d.GetOk("code"); ok && v.(string) != "" {
		createOpts.ApplicationCode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
		var cloudwatchLoggingOptions []*kinesisanalytics.CloudWatchLoggingOption
		clos := v.([]interface{})

		if len(clos) > 0 {
			clo := clos[0].(map[string]interface{})
			cloudwatchLoggingOption := &kinesisanalytics.CloudWatchLoggingOption{
				LogStreamARN: aws.String(clo["log_stream"].(string)),
				RoleARN:      aws.String(clo["role"].(string)),
			}
			cloudwatchLoggingOptions = append(cloudwatchLoggingOptions, cloudwatchLoggingOption)
		}

		createOpts.CloudWatchLoggingOptions = cloudwatchLoggingOptions
	}

	_, err := conn.CreateApplication(createOpts)
	if err != nil {
		return fmt.Errorf("Unable to create Kinesis Analytics Application: %s", err)
	}

	return resourceAwsKinesisAnalyticsApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn
	name := d.Get("name").(string)

	describeOpts := &kinesisanalytics.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}
	resp, err := conn.DescribeApplication(describeOpts)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("[WARN] Error reading Kinesis Analytics Application: \"%s\", code: \"%s\"", awsErr.Message(), awsErr.Code())
		}
		return err
	}

	d.SetId(aws.StringValue(resp.ApplicationDetail.ApplicationARN))
	d.Set("name", aws.StringValue(resp.ApplicationDetail.ApplicationName))
	d.Set("arn", aws.StringValue(resp.ApplicationDetail.ApplicationARN))
	d.Set("code", aws.StringValue(resp.ApplicationDetail.ApplicationCode))
	d.Set("create_timestamp", aws.TimeValue(resp.ApplicationDetail.CreateTimestamp).Format(time.RFC3339))
	d.Set("description", aws.StringValue(resp.ApplicationDetail.ApplicationDescription))
	d.Set("last_update_timestamp", aws.TimeValue(resp.ApplicationDetail.LastUpdateTimestamp).Format(time.RFC3339))
	d.Set("status", aws.StringValue(resp.ApplicationDetail.ApplicationStatus))
	d.Set("version", int(aws.Int64Value(resp.ApplicationDetail.ApplicationVersionId)))

	if err := d.Set("cloudwatch_logging_options", getCloudwatchLoggingOptions(resp.ApplicationDetail.CloudWatchLoggingOptionDescriptions)); err != nil {
		return err
	}

	return nil
}

func resourceAwsKinesisAnalyticsApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn

	if !d.IsNewResource() {
		applicationUpdate := &kinesisanalytics.ApplicationUpdate{}
		name := d.Get("name").(string)
		version := d.Get("version").(int)

		updateApplicationOpts := &kinesisanalytics.UpdateApplicationInput{
			ApplicationName:             aws.String(name),
			CurrentApplicationVersionId: aws.Int64(int64(version)),
		}

		applicationUpdate, err := createApplicationUpdateOpts(d)
		if err != nil {
			return err
		}

		if !reflect.DeepEqual(applicationUpdate, &kinesisanalytics.ApplicationUpdate{}) {
			updateApplicationOpts.SetApplicationUpdate(applicationUpdate)
			_, updateErr := conn.UpdateApplication(updateApplicationOpts)
			if updateErr != nil {
				return updateErr
			}
			version = version + 1
		}

		oldLoggingOptions, newLoggingOptions := d.GetChange("cloudwatch_logging_options")
		if len(oldLoggingOptions.([]interface{})) == 0 && len(newLoggingOptions.([]interface{})) > 0 {
			if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
				clo := v.([]interface{})[0].(map[string]interface{})
				cloudwatchLoggingOption := &kinesisanalytics.CloudWatchLoggingOption{
					LogStreamARN: aws.String(clo["log_stream"].(string)),
					RoleARN:      aws.String(clo["role"].(string)),
				}
				addOpts := &kinesisanalytics.AddApplicationCloudWatchLoggingOptionInput{
					ApplicationName:             aws.String(name),
					CurrentApplicationVersionId: aws.Int64(int64(version)),
					CloudWatchLoggingOption:     cloudwatchLoggingOption,
				}
				conn.AddApplicationCloudWatchLoggingOption(addOpts)
			}
		}
	}

	return resourceAwsKinesisAnalyticsApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn
	name := d.Get("name").(string)
	createTimestamp, parseErr := time.Parse(time.RFC3339, d.Get("create_timestamp").(string))
	if parseErr != nil {
		return parseErr
	}

	log.Printf("[DEBUG] Kinesis Analytics Application destroy: %v", d.Id())
	deleteOpts := &kinesisanalytics.DeleteApplicationInput{
		ApplicationName: aws.String(name),
		CreateTimestamp: aws.Time(createTimestamp),
	}
	_, deleteErr := conn.DeleteApplication(deleteOpts)
	if deleteErr != nil {
		return deleteErr
	}

	log.Printf("[DEBUG] Kinesis Analytics Application deleted: %v", d.Id())
	return nil
}

func createApplicationUpdateOpts(d *schema.ResourceData) (*kinesisanalytics.ApplicationUpdate, error) {
	applicationUpdate := &kinesisanalytics.ApplicationUpdate{}

	if d.HasChange("code") {
		if v, ok := d.GetOk("code"); ok && v.(string) != "" {
			applicationUpdate.ApplicationCodeUpdate = aws.String(v.(string))
		}
	}

	oldLoggingOptions, _ := d.GetChange("cloudwatch_logging_options")
	if len(oldLoggingOptions.([]interface{})) > 0 {
		if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
			var cloudwatchLoggingOptions []*kinesisanalytics.CloudWatchLoggingOptionUpdate
			clo := v.([]interface{})[0].(map[string]interface{})
			cloudwatchLoggingOption := &kinesisanalytics.CloudWatchLoggingOptionUpdate{
				CloudWatchLoggingOptionId: aws.String(clo["id"].(string)),
				LogStreamARNUpdate:        aws.String(clo["log_stream"].(string)),
				RoleARNUpdate:             aws.String(clo["role"].(string)),
			}
			cloudwatchLoggingOptions = append(cloudwatchLoggingOptions, cloudwatchLoggingOption)
			applicationUpdate.CloudWatchLoggingOptionUpdates = cloudwatchLoggingOptions
		}
	}

	return applicationUpdate, nil
}

func getCloudwatchLoggingOptions(options []*kinesisanalytics.CloudWatchLoggingOptionDescription) []interface{} {
	s := []interface{}{}
	for _, v := range options {
		option := map[string]interface{}{
			"id":         aws.StringValue(v.CloudWatchLoggingOptionId),
			"log_stream": aws.StringValue(v.LogStreamARN),
			"role":       aws.StringValue(v.RoleARN),
		}
		s = append(s, option)
	}
	return s
}
