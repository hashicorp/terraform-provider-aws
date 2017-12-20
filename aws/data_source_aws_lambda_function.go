package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsLambdaFunction() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLambdaFunctionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "$LATEST",
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLambdaFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn
	function_name := d.Get("function_name")

	log.Printf("[DEBUG] Fetching Lambda Function: %s", function_name)

	params := &lambda.GetFunctionInput{
		FunctionName: aws.String(function_name.(string)),
		Qualifier:    aws.String(d.Get("qualifier").(string)),
	}

	getFunctionOutput, err := conn.GetFunction(params)
	if err != nil {
		return fmt.Errorf("Failed getting Lambda Function \"%s\": %s", function_name, err)
	}

	function := getFunctionOutput.Configuration
	d.SetId(function_name.(string))
	d.Set("arn", function.FunctionArn)
	d.Set("role", function.Role)
	d.Set("version", function.Version)

	return nil
}
