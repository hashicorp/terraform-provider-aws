package aws

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsLambdaInvoke() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLambdaInvokeRead,

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "$LATEST",
			},

			"input": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"result": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceAwsLambdaInvokeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)

	payload, err := json.Marshal(d.Get("input"))

	if err != nil {
		return err
	}

	res, err := conn.Invoke(&lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		Payload:        payload,
		Qualifier:      aws.String(qualifier),
	})

	if err != nil {
		return err
	}

	if res.FunctionError != nil {
		return fmt.Errorf("Lambda function (%s) returned error: (%s)", functionName, string(res.Payload))
	}

	var result map[string]interface{}

	if err = json.Unmarshal(res.Payload, &result); err != nil {
		return err
	}

	if err = d.Set("result", result); err != nil {
		return fmt.Errorf("Lambda function (%s) returned invalid JSON: %s", functionName, err)
	}

	d.SetId(fmt.Sprintf("%s_%s_%x", functionName, qualifier, md5.Sum(payload)))

	return nil
}
