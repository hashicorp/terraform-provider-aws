package lambda

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceInvocation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInvocationCreate,
		ReadWithoutTimeout:   resourceInvocationRead,
		DeleteWithoutTimeout: resourceInvocationDelete,

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

func resourceInvocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn()

	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	input := []byte(d.Get("input").(string))

	res, err := conn.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		Payload:        input,
		Qualifier:      aws.String(qualifier),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Lambda Invocation (%s) failed: %s", d.Id(), err)
	}

	if res.FunctionError != nil {
		return sdkdiag.AppendErrorf(diags, "Lambda function (%s) returned error: (%s)", functionName, string(res.Payload))
	}

	d.SetId(fmt.Sprintf("%s_%s_%x", functionName, qualifier, md5.Sum(input)))
	d.Set("result", string(res.Payload))

	return diags
}

func resourceInvocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceInvocationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[DEBUG] Lambda Invocation (%s) \"deleted\" by removing from state", d.Id())
	return diags
}
