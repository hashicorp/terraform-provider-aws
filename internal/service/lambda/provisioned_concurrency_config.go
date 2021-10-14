package lambda

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceProvisionedConcurrencyConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceProvisionedConcurrencyConfigCreate,
		Read:   resourceProvisionedConcurrencyConfigRead,
		Update: resourceProvisionedConcurrencyConfigUpdate,
		Delete: resourceProvisionedConcurrencyConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"provisioned_concurrent_executions": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"qualifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceProvisionedConcurrencyConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn
	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)

	input := &lambda.PutProvisionedConcurrencyConfigInput{
		FunctionName:                    aws.String(functionName),
		ProvisionedConcurrentExecutions: aws.Int64(int64(d.Get("provisioned_concurrent_executions").(int))),
		Qualifier:                       aws.String(qualifier),
	}

	_, err := conn.PutProvisionedConcurrencyConfig(input)

	if err != nil {
		return fmt.Errorf("error putting Lambda Provisioned Concurrency Config (%s:%s): %s", functionName, qualifier, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", functionName, qualifier))

	if err := waitForLambdaProvisionedConcurrencyConfigStatusReady(conn, functionName, qualifier, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Lambda Provisioned Concurrency Config (%s) to be ready: %s", d.Id(), err)
	}

	return resourceProvisionedConcurrencyConfigRead(d, meta)
}

func resourceProvisionedConcurrencyConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	functionName, qualifier, err := ProvisionedConcurrencyConfigParseID(d.Id())

	if err != nil {
		return err
	}

	input := &lambda.GetProvisionedConcurrencyConfigInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(qualifier),
	}

	output, err := conn.GetProvisionedConcurrencyConfig(input)

	if tfawserr.ErrMessageContains(err, lambda.ErrCodeProvisionedConcurrencyConfigNotFoundException, "") || tfawserr.ErrMessageContains(err, lambda.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Lambda Provisioned Concurrency Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Lambda Provisioned Concurrency Config (%s): %s", d.Id(), err)
	}

	d.Set("function_name", functionName)
	d.Set("provisioned_concurrent_executions", output.AllocatedProvisionedConcurrentExecutions)
	d.Set("qualifier", qualifier)

	return nil
}

func resourceProvisionedConcurrencyConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	functionName, qualifier, err := ProvisionedConcurrencyConfigParseID(d.Id())

	if err != nil {
		return err
	}

	input := &lambda.PutProvisionedConcurrencyConfigInput{
		FunctionName:                    aws.String(functionName),
		ProvisionedConcurrentExecutions: aws.Int64(int64(d.Get("provisioned_concurrent_executions").(int))),
		Qualifier:                       aws.String(qualifier),
	}

	_, err = conn.PutProvisionedConcurrencyConfig(input)

	if err != nil {
		return fmt.Errorf("error putting Lambda Provisioned Concurrency Config (%s:%s): %s", functionName, qualifier, err)
	}

	if err := waitForLambdaProvisionedConcurrencyConfigStatusReady(conn, functionName, qualifier, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for Lambda Provisioned Concurrency Config (%s) to be ready: %s", d.Id(), err)
	}

	return resourceProvisionedConcurrencyConfigRead(d, meta)
}

func resourceProvisionedConcurrencyConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	functionName, qualifier, err := ProvisionedConcurrencyConfigParseID(d.Id())

	if err != nil {
		return err
	}

	input := &lambda.DeleteProvisionedConcurrencyConfigInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(qualifier),
	}

	_, err = conn.DeleteProvisionedConcurrencyConfig(input)

	if tfawserr.ErrMessageContains(err, lambda.ErrCodeProvisionedConcurrencyConfigNotFoundException, "") || tfawserr.ErrMessageContains(err, lambda.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error putting Lambda Provisioned Concurrency Config (%s:%s): %s", functionName, qualifier, err)
	}

	return nil
}

func ProvisionedConcurrencyConfigParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected FUNCTION_NAME:QUALIFIER", id)
	}

	return parts[0], parts[1], nil
}

func refreshLambdaProvisionedConcurrencyConfigStatus(conn *lambda.Lambda, functionName, qualifier string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lambda.GetProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(functionName),
			Qualifier:    aws.String(qualifier),
		}

		output, err := conn.GetProvisionedConcurrencyConfig(input)

		if err != nil {
			return "", "", err
		}

		status := aws.StringValue(output.Status)

		if status == lambda.ProvisionedConcurrencyStatusEnumFailed {
			return output, status, fmt.Errorf("status reason: %s", aws.StringValue(output.StatusReason))
		}

		return output, status, nil
	}
}

func waitForLambdaProvisionedConcurrencyConfigStatusReady(conn *lambda.Lambda, functionName, qualifier string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{lambda.ProvisionedConcurrencyStatusEnumInProgress},
		Target:  []string{lambda.ProvisionedConcurrencyStatusEnumReady},
		Refresh: refreshLambdaProvisionedConcurrencyConfigStatus(conn, functionName, qualifier),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
