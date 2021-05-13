package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
			State: schema.ImportStatePassthrough,
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

// resourceAwsCloudFrontFunction maps to:
// CreateFunction in the API / SDK
func resourceAwsCloudFrontFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	functionName := d.Get("name").(string)

	log.Printf("[DEBUG] Creating CloudFront Function %s", functionName)

	params := &cloudfront.CreateFunctionInput{
		FunctionCode: []byte(d.Get("code").(string)),
		FunctionConfig: &cloudfront.FunctionConfig{
			Comment: aws.String(d.Get("comment").(string)),
			Runtime: aws.String(d.Get("runtime").(string)),
		},
		Name: aws.String(functionName),
	}

	createFunctionOutput, err := conn.CreateFunction(params)

	if err != nil {
		return fmt.Errorf("error creating CloudFront Function (%s): %w", functionName, err)
	}

	d.SetId(aws.StringValue(createFunctionOutput.FunctionSummary.Name))
	etag := createFunctionOutput.ETag

	publish := d.Get("publish").(bool)
	if publish {

		params := &cloudfront.PublishFunctionInput{
			Name:    aws.String(d.Id()),
			IfMatch: aws.String(&etag),
		}

		log.Printf("[DEBUG] Publishing Cloudfront Function: %s", params)

		PublishFunctionOutput, err := conn.PublishFunction(params)
		if err != nil {
			return fmt.Errorf("error publishing CloudFront Function (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsCloudFrontFunctionRead(d, meta)
}

// resourceAwsCloudFrontFunctionRead maps to:
// GetFunction in the API / SDK
func resourceAwsCloudFrontFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	params := &cloudfront.GetFunctionInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Get Cloudfront Function: %s", d.Id())

	getFunctionOutput, err := conn.GetFunction(params)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchFunctionExists) {
		log.Printf("[WARN] CloudFront Function (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting CloudFront Function (%s): %w", d.Id(), err)
	}

	d.Set("code", string(getFunctionOutput.FunctionCode))

	describeParams := &cloudfront.DescribeFunctionInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Fetching Cloudfront Function: %s", d.Id())

	describeFunctionOutput, err := conn.DescribeFunction(describeParams)
	if err != nil {
		return err
	}

	d.Set("etag", describeFunctionOutput.ETag)
	d.Set("arn", describeFunctionOutput.FunctionSummary.FunctionMetadata.FunctionARN)
	d.Set("last_modified", describeFunctionOutput.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339)) // 2006-01-02T15:04:05Z0700
	d.Set("stage", describeFunctionOutput.FunctionSummary.FunctionMetadata.Stage)
	d.Set("comment", describeFunctionOutput.FunctionSummary.FunctionConfig.Comment)
	d.Set("runtime", describeFunctionOutput.FunctionSummary.FunctionConfig.Runtime)
	d.Set("status", describeFunctionOutput.FunctionSummary.Status)

	return nil
}

// resourceAwsCloudFrontFunction maps to:
// DeleteFunction in the API / SDK
func resourceAwsCloudFrontFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	log.Printf("[INFO] Deleting Cloudfront Function: %s", d.Id())

	params := &cloudfront.DeleteFunctionInput{
		Name:    aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteFunction(params)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchFunctionExists) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudFront Function (%s): %w", d.Id(), err)
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
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	log.Printf("[DEBUG] Send Update Cloudfront Function Configuration request: %#v", configReq)

	UpdateFunctionOutput, err := conn.UpdateFunction(configReq)

	if err != nil {
		return fmt.Errorf("error updating CloudFront Function (%s) configuration : %w", d.Id(), err)
	}

	publish := d.Get("publish").(bool)
	if publish {

		params := &cloudfront.PublishFunctionInput{
			Name:    aws.String(d.Id()),
			IfMatch: aws.String(&UpdateFunctionOutput.ETag),
		}

		log.Printf("[DEBUG] Publishing Cloudfront Function: %s", d.Id())

		_, err := conn.PublishFunction(params)
		if err != nil {
			return fmt.Errorf("error publishing CloudFront Function (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsCloudFrontFunctionRead(d, meta)
}
