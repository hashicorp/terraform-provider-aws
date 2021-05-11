package aws

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCloudFrontFunction() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudFrontFunctionRead,

		Schema: map[string]*schema.Schema{
			"code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"runtime": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCloudFrontFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	params := &cloudfront.GetFunctionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Get Cloudfront Function: %s", d.Id())

	GetFunctionOutput, err := conn.GetFunction(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == cloudfront.ErrCodeNoSuchFunctionExists && !d.IsNewResource() {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("code", string(GetFunctionOutput.FunctionCode))

	describeParams := &cloudfront.DescribeFunctionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Fetching Cloudfront Function: %s", d.Id())

	DescribeFunctionOutput, err := conn.DescribeFunction(describeParams)
	if err != nil {
		return err
	}

	d.Set("version", DescribeFunctionOutput.ETag)
	d.Set("arn", DescribeFunctionOutput.FunctionSummary.FunctionMetadata.FunctionARN)
	d.Set("last_modified", DescribeFunctionOutput.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339))
	d.Set("stage", DescribeFunctionOutput.FunctionSummary.FunctionMetadata.Stage)
	d.Set("comment", DescribeFunctionOutput.FunctionSummary.FunctionConfig.Comment)
	d.Set("runtime", DescribeFunctionOutput.FunctionSummary.FunctionConfig.Runtime)
	d.Set("status", DescribeFunctionOutput.FunctionSummary.Status)

	d.SetId(d.Get("name").(string))

	return nil
}
