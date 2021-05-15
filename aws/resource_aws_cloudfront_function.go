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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"code": {
				Type:     schema.TypeString,
				Required: true,
			},

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"publish": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"runtime": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.FunctionRuntime_Values(), false),
			},

			"stage": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
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
	input := &cloudfront.CreateFunctionInput{
		FunctionCode: []byte(d.Get("code").(string)),
		FunctionConfig: &cloudfront.FunctionConfig{
			Comment: aws.String(d.Get("comment").(string)),
			Runtime: aws.String(d.Get("runtime").(string)),
		},
		Name: aws.String(functionName),
	}

	log.Printf("[DEBUG] Creating CloudFront Function %s", functionName)
	output, err := conn.CreateFunction(input)

	if err != nil {
		return fmt.Errorf("error creating CloudFront Function (%s): %w", functionName, err)
	}

	d.SetId(aws.StringValue(output.FunctionSummary.Name))

	if d.Get("publish").(bool) {
		input := &cloudfront.PublishFunctionInput{
			Name:    aws.String(d.Id()),
			IfMatch: output.ETag,
		}

		log.Printf("[DEBUG] Publishing Cloudfront Function: %s", input)
		_, err := conn.PublishFunction(input)

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

	describeFunctionOutput, err := conn.DescribeFunction(&cloudfront.DescribeFunctionInput{
		Name: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchFunctionExists) {
		log.Printf("[WARN] CloudFront Function (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing CloudFront Function (%s): %w", d.Id(), err)
	}

	d.Set("arn", describeFunctionOutput.FunctionSummary.FunctionMetadata.FunctionARN)
	d.Set("comment", describeFunctionOutput.FunctionSummary.FunctionConfig.Comment)
	d.Set("etag", describeFunctionOutput.ETag)
	d.Set("last_modified", describeFunctionOutput.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339)) // 2006-01-02T15:04:05Z0700
	d.Set("name", describeFunctionOutput.FunctionSummary.Name)
	d.Set("runtime", describeFunctionOutput.FunctionSummary.FunctionConfig.Runtime)
	d.Set("stage", describeFunctionOutput.FunctionSummary.FunctionMetadata.Stage)
	d.Set("status", describeFunctionOutput.FunctionSummary.Status)

	getFunctionOutput, err := conn.GetFunction(&cloudfront.GetFunctionInput{
		Name: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error describing CloudFront Function (%s): %w", d.Id(), err)
	}

	d.Set("code", string(getFunctionOutput.FunctionCode))

	return nil
}

// resourceAwsCloudFrontFunction maps to:
// DeleteFunction in the API / SDK
func resourceAwsCloudFrontFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	log.Printf("[INFO] Deleting Cloudfront Function: %s", d.Id())
	_, err := conn.DeleteFunction(&cloudfront.DeleteFunctionInput{
		Name:    aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

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

	input := &cloudfront.UpdateFunctionInput{
		FunctionCode: []byte(d.Get("code").(string)),
		FunctionConfig: &cloudfront.FunctionConfig{
			Comment: aws.String(d.Get("comment").(string)),
			Runtime: aws.String(d.Get("runtime").(string)),
		},
		Name:    aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	log.Printf("[INFO] Updating Cloudfront Function: %s", d.Id())
	output, err := conn.UpdateFunction(input)

	if err != nil {
		return fmt.Errorf("error updating CloudFront Function (%s) configuration : %w", d.Id(), err)
	}

	if d.Get("publish").(bool) {
		input := &cloudfront.PublishFunctionInput{
			Name:    aws.String(d.Id()),
			IfMatch: output.ETag,
		}

		log.Printf("[DEBUG] Publishing Cloudfront Function: %s", d.Id())
		_, err := conn.PublishFunction(input)

		if err != nil {
			return fmt.Errorf("error publishing CloudFront Function (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsCloudFrontFunctionRead(d, meta)
}
