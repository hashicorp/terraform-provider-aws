package aws

import (
	"crypto/md5"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsLambdaInvocation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLambdaInvocationCreate,
		Read:   resourceAwsLambdaInvocationRead,
		Update: resourceAwsLambdaInvocationUpdate,
		Delete: resourceAwsLambdaInvocationDelete,

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "$LATEST",
			},

			"input": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},

			"invoke_on_update": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"result": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLambdaInvocationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	input := []byte(d.Get("input").(string))

	res, err := conn.Invoke(&lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		Payload:        input,
		Qualifier:      aws.String(qualifier),
	})

	if err != nil {
		return err
	}

	if res.FunctionError != nil {
		return fmt.Errorf("Lambda function (%s) returned error: (%s)", functionName, string(res.Payload))
	}

	if err = d.Set("result", string(res.Payload)); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s_%s_%x", functionName, qualifier, md5.Sum(input)))

	return nil
}

func resourceAwsLambdaInvocationRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsLambdaInvocationUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.Get("invoke_on_update").(bool) {
		return resourceAwsLambdaInvocationCreate(d, meta)
	} else {
		return nil
	}
}

func resourceAwsLambdaInvocationDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	d.Set("result", nil)
	return nil
}
