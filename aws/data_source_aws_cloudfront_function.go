package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
			"etag": {
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

	getFunctionOutput, err := conn.GetFunction(params)
	if err != nil {
		return fmt.Errorf("error getting CloudFront Function (%s): %w", d.Get("name").(string), err)
	}

	d.Set("code", string(getFunctionOutput.FunctionCode))

	describeParams := &cloudfront.DescribeFunctionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	describeFunctionOutput, err := conn.DescribeFunction(describeParams)
	if err != nil {
		return err
	}

	d.Set("etag", describeFunctionOutput.ETag)
	d.Set("arn", describeFunctionOutput.FunctionSummary.FunctionMetadata.FunctionARN)
	d.Set("last_modified", describeFunctionOutput.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339))
	d.Set("stage", describeFunctionOutput.FunctionSummary.FunctionMetadata.Stage)
	d.Set("comment", describeFunctionOutput.FunctionSummary.FunctionConfig.Comment)
	d.Set("runtime", describeFunctionOutput.FunctionSummary.FunctionConfig.Runtime)
	d.Set("status", describeFunctionOutput.FunctionSummary.Status)

	d.SetId(d.Get("name").(string))

	return nil
}
