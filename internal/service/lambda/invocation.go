package lambda

import (
	"crypto/md5"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceInvocation() *schema.Resource {
	return &schema.Resource{
		Create: resourceInvocationCreate,
		Read:   resourceInvocationRead,
		Delete: resourceInvocationDelete,

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"input": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  FunctionVersionLatest,
			},
			"result": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceInvocationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

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
		return fmt.Errorf("Lambda Invocation (%s) failed: %w", d.Id(), err)
	}

	if res.FunctionError != nil {
		return fmt.Errorf("Lambda function (%s) returned error: (%s)", functionName, string(res.Payload))
	}

	d.SetId(fmt.Sprintf("%s_%s_%x", functionName, qualifier, md5.Sum(input)))
	d.Set("result", string(res.Payload))

	return nil
}

func resourceInvocationRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceInvocationDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Lambda Invocation (%s) \"deleted\" by removing from state", d.Id())
	return nil
}
