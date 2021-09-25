package aws

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfconnect "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/waiter"
)

func resourceAwsConnectLambdaFunctionAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsConnectLambdaFunctionAssociationCreate,
		ReadContext:   resourceAwsConnectLambdaFunctionAssociationRead,
		UpdateContext: resourceAwsConnectLambdaFunctionAssociationRead,
		DeleteContext: resourceAwsConnectLambdaFunctionAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(waiter.ConnectLambdaFunctionAssociationCreatedTimeout),
			Delete: schema.DefaultTimeout(waiter.ConnectInstanceDeletedTimeout),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"function_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsConnectLambdaFunctionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	input := &connect.AssociateLambdaFunctionInput{
		InstanceId: aws.String(d.Get("instance_id").(string)),
		FunctionArn:     aws.String(d.Get("function_arn").(string)),
	}

	log.Printf("[DEBUG] Creating Connect Lambda Association %s", input)

	_, err := conn.AssociateLambdaFunctionWithContext(ctx, input)

	d.SetId(fmt.Sprintf("%s:%s", d.Get("instance_id").(string), d.Get("function_arn").(string)))

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Lambda Function Association (%s): %s", d.Id(), err))
	}

	return resourceAwsConnectLambdaFunctionAssociationRead(ctx, d, meta)
}

func resourceAwsConnectLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn
	instanceID := d.Get("instance_id")
	functionArn := d.Get("function_arn")

	FunctionArn, err := resourceAwsConnectGetLambdaAssociationByFunctionArn(ctx, conn, instanceID.(string), functionArn.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Lambda Association by Function ARN (%s): %w", functionArn, err))
	}

	if FunctionArn == "" {
		return diag.FromErr(fmt.Errorf("error finding Lambda Association Function ARN (%s): not found", functionArn))
	}

	d.Set("function_arn", FunctionArn)
	d.Set("instance_id", instanceID)
	d.SetId(fmt.Sprintf("%s:%s", instanceID, d.Get("function_arn").(string)))

	return nil
}

func resourceAwsConnectLambdaFunctionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	instanceID, functionArn , err := resourceAwsConnectLambdaFunctionAssociationParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &connect.DisassociateLambdaFunctionInput{
		InstanceId: aws.String(instanceID),
		FunctionArn:    aws.String(functionArn),
	}

	log.Printf("[DEBUG] Deleting Connect Lambda Function Association %s", d.Id())

	_, dissErr := conn.DisassociateLambdaFunction(input)

	if dissErr != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Lambda Function Association (%s): %s", d.Id(), err))
	}
	return nil
}

func resourceAwsConnectGetLambdaAssociationByFunctionArn(ctx context.Context, conn *connect.Connect, instanceID string, functionArn string) (string, error) {
	var result string

	input := &connect.ListLambdaFunctionsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(tfconnect.ListLambdaFunctionsMaxResults),
	}

	err := conn.ListLambdaFunctionsPagesWithContext(ctx, input, func(page *connect.ListLambdaFunctionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.LambdaFunctions {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf) == functionArn {
				result = aws.StringValue(cf)
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	return result, nil
}

func resourceAwsConnectLambdaFunctionAssociationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 2 || parts[0] == "" || parts[1] == ""  {
		return "", "",  fmt.Errorf("unexpected format of ID (%s), expected instanceID:functionArn", id)
	}

	return parts[0], parts[1], nil
}
