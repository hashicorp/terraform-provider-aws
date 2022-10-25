package cloudfront

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceFunctionCreate,
		Read:   resourceFunctionRead,
		Update: resourceFunctionUpdate,
		Delete: resourceFunctionDelete,
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
			"live_stage_etag": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	functionName := d.Get("name").(string)
	input := &cloudfront.CreateFunctionInput{
		FunctionCode: []byte(d.Get("code").(string)),
		FunctionConfig: &cloudfront.FunctionConfig{
			Comment: aws.String(d.Get("comment").(string)),
			Runtime: aws.String(d.Get("runtime").(string)),
		},
		Name: aws.String(functionName),
	}

	log.Printf("[DEBUG] Creating CloudFront Function: %s", functionName)
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

		log.Printf("[DEBUG] Publishing CloudFront Function: %s", input)
		_, err := conn.PublishFunction(input)

		if err != nil {
			return fmt.Errorf("error publishing CloudFront Function (%s): %w", d.Id(), err)
		}
	}

	return resourceFunctionRead(d, meta)
}

func resourceFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	describeFunctionOutput, err := FindFunctionByNameAndStage(conn, d.Id(), cloudfront.FunctionStageDevelopment)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Function (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFront Function (%s) DEVELOPMENT stage: %w", d.Id(), err)
	}

	d.Set("arn", describeFunctionOutput.FunctionSummary.FunctionMetadata.FunctionARN)
	d.Set("comment", describeFunctionOutput.FunctionSummary.FunctionConfig.Comment)
	d.Set("etag", describeFunctionOutput.ETag)
	d.Set("name", describeFunctionOutput.FunctionSummary.Name)
	d.Set("runtime", describeFunctionOutput.FunctionSummary.FunctionConfig.Runtime)
	d.Set("status", describeFunctionOutput.FunctionSummary.Status)

	getFunctionOutput, err := conn.GetFunction(&cloudfront.GetFunctionInput{
		Name:  aws.String(d.Id()),
		Stage: aws.String(cloudfront.FunctionStageDevelopment),
	})

	if err != nil {
		return fmt.Errorf("error reading CloudFront Function (%s) DEVELOPMENT stage code: %w", d.Id(), err)
	}

	d.Set("code", string(getFunctionOutput.FunctionCode))

	describeFunctionOutput, err = FindFunctionByNameAndStage(conn, d.Id(), cloudfront.FunctionStageLive)

	if tfresource.NotFound(err) {
		d.Set("live_stage_etag", "")
	} else if err != nil {
		return fmt.Errorf("error reading CloudFront Function (%s) LIVE stage: %w", d.Id(), err)
	} else {
		d.Set("live_stage_etag", describeFunctionOutput.ETag)
	}

	return nil
}

func resourceFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn
	etag := d.Get("etag").(string)

	if d.HasChanges("code", "comment", "runtime") {
		input := &cloudfront.UpdateFunctionInput{
			FunctionCode: []byte(d.Get("code").(string)),
			FunctionConfig: &cloudfront.FunctionConfig{
				Comment: aws.String(d.Get("comment").(string)),
				Runtime: aws.String(d.Get("runtime").(string)),
			},
			Name:    aws.String(d.Id()),
			IfMatch: aws.String(etag),
		}

		log.Printf("[INFO] Updating CloudFront Function: %s", d.Id())
		output, err := conn.UpdateFunction(input)

		if err != nil {
			return fmt.Errorf("error updating CloudFront Function (%s): %w", d.Id(), err)
		}

		etag = aws.StringValue(output.ETag)
	}

	if d.Get("publish").(bool) {
		input := &cloudfront.PublishFunctionInput{
			Name:    aws.String(d.Id()),
			IfMatch: aws.String(etag),
		}

		log.Printf("[DEBUG] Publishing CloudFront Function: %s", d.Id())
		_, err := conn.PublishFunction(input)

		if err != nil {
			return fmt.Errorf("error publishing CloudFront Function (%s): %w", d.Id(), err)
		}
	}

	return resourceFunctionRead(d, meta)
}

func resourceFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	log.Printf("[INFO] Deleting CloudFront Function: %s", d.Id())
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
