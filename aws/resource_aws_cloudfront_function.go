package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsCloudFrontFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFrontFunctionCreate,
		Read:   resourceAwsCloudFrontFunctionRead,
		Update: resourceAwsCloudFrontFunctionUpdate,
		Delete: resourceAwsCloudFrontFunctionDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"runtime": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.FunctionRuntime_Values(), false),
			},
			"publish": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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

// resourceAwsCloudFrontFunction maps to:
// CreateFunction in the API / SDK
func resourceAwsCloudFrontFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	functionName := d.Get("name").(string)

	log.Printf("[DEBUG] Creating Lambda Function %s", functionName)

	params := &cloudfront.CreateFunctionInput{
		FunctionCode: []byte(d.Get("code").(string)),
		FunctionConfig: &cloudfront.FunctionConfig{
			Comment: aws.String(d.Get("comment").(string)),
			Runtime: aws.String(d.Get("runtime").(string)),
		},
		Name: aws.String(functionName),
	}

	CreateFunctionOutput, err := conn.CreateFunction(params)

	if err != nil {
		return fmt.Errorf("error creating Cloudfront Function: %w", err)
	}

	d.SetId(d.Get("name").(string))
	d.Set("version", CreateFunctionOutput.ETag)
	d.Set("arn", CreateFunctionOutput.FunctionSummary.FunctionMetadata.FunctionARN)
	d.Set("last_modified", CreateFunctionOutput.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339))
	d.Set("stage", CreateFunctionOutput.FunctionSummary.FunctionMetadata.Stage)
	d.Set("status", CreateFunctionOutput.FunctionSummary.Status)

	publish := d.Get("publish").(bool)
	if publish {

		params := &cloudfront.PublishFunctionInput{
			Name:    aws.String(d.Get("name").(string)),
			IfMatch: aws.String(d.Get("version").(string)),
		}

		log.Printf("[DEBUG] Publishing Cloudfront Function: %s", d.Id())

		PublishFunctionOutput, err := conn.PublishFunction(params)
		if err != nil {
			return err
		}
		d.Set("status", PublishFunctionOutput.FunctionSummary.Status)
		d.Set("last_modified", PublishFunctionOutput.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339))
	}

	return nil
}

// resourceAwsCloudFrontFunctionRead maps to:
// GetFunction in the API / SDK
func resourceAwsCloudFrontFunctionRead(d *schema.ResourceData, meta interface{}) error {
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
	d.Set("last_modified", DescribeFunctionOutput.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339)) // 2006-01-02T15:04:05Z0700
	d.Set("stage", DescribeFunctionOutput.FunctionSummary.FunctionMetadata.Stage)
	d.Set("comment", DescribeFunctionOutput.FunctionSummary.FunctionConfig.Comment)
	d.Set("runtime", DescribeFunctionOutput.FunctionSummary.FunctionConfig.Runtime)
	d.Set("status", DescribeFunctionOutput.FunctionSummary.Status)
	d.Set("publish", true)

	return nil
}

// resourceAwsCloudFrontFunction maps to:
// DeleteFunction in the API / SDK
func resourceAwsCloudFrontFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	log.Printf("[INFO] Deleting Cloudfront Function: %s", d.Id())

	params := &cloudfront.DeleteFunctionInput{
		Name:    aws.String(d.Get("name").(string)),
		IfMatch: aws.String(d.Get("version").(string)),
	}

	_, err := conn.DeleteFunction(params)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchFunctionExists) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Cloudfront Function (%s): %w", d.Id(), err)
	}

	return nil
}

// resourceAwsCloudFrontFunctionUpdate maps to:
// UpdateFunctionCode in the API / SDK
func resourceAwsCloudFrontFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	// arn := d.Get("arn").(string)

	configReq := &cloudfront.UpdateFunctionInput{
		FunctionCode: []byte(d.Get("code").(string)),
		FunctionConfig: &cloudfront.FunctionConfig{
			Comment: aws.String(d.Get("comment").(string)),
			Runtime: aws.String(d.Get("runtime").(string)),
		},
		Name:    aws.String(d.Id()),
		IfMatch: aws.String(d.Get("version").(string)),
	}

	log.Printf("[DEBUG] Send Update Cloudfront Function Configuration request: %#v", configReq)

	UpdateFunctionOutput, err := conn.UpdateFunction(configReq)

	if err != nil {
		return fmt.Errorf("error modifying Lambda Function (%s) configuration : %w", d.Id(), err)
	}

	d.Set("version", UpdateFunctionOutput.ETag)
	d.Set("last_modified", UpdateFunctionOutput.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339))

	publish := d.Get("publish").(bool)
	if publish {

		params := &cloudfront.PublishFunctionInput{
			Name:    aws.String(d.Get("name").(string)),
			IfMatch: aws.String(d.Get("version").(string)),
		}

		log.Printf("[DEBUG] Publishing Cloudfront Function: %s", d.Id())

		_, err := conn.PublishFunction(params)
		if err != nil {
			return err
		}
	}

	return resourceAwsCloudFrontFunctionRead(d, meta)
}

/*
func refreshCloudfrontFunctionLastUpdateStatus(conn *lambda.Lambda, functionName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lambda.GetFunctionInput{
			FunctionName: aws.String(functionName),
		}

		output, err := conn.GetFunction(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.Configuration == nil {
			return nil, "", nil
		}

		lastUpdateStatus := aws.StringValue(output.Configuration.LastUpdateStatus)

		if lastUpdateStatus == lambda.LastUpdateStatusFailed {
			return output.Configuration, lastUpdateStatus, fmt.Errorf("%s: %s", aws.StringValue(output.Configuration.LastUpdateStatusReasonCode), aws.StringValue(output.Configuration.LastUpdateStatusReason))
		}

		return output.Configuration, lastUpdateStatus, nil
	}
}

func refreshCloudfrontFunctionState(conn *lambda.Lambda, functionName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lambda.GetFunctionInput{
			FunctionName: aws.String(functionName),
		}

		output, err := conn.GetFunction(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.Configuration == nil {
			return nil, "", nil
		}

		state := aws.StringValue(output.Configuration.State)

		if state == lambda.StateFailed {
			return output.Configuration, state, fmt.Errorf("%s: %s", aws.StringValue(output.Configuration.StateReasonCode), aws.StringValue(output.Configuration.StateReason))
		}

		return output.Configuration, state, nil
	}
}

*/
