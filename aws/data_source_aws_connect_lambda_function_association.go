package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfconnect "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect"
)

func dataSourceAwsConnectLambdaFunctionAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsConnectLambdaFunctionAssociationRead,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"function_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsConnectLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn
	instanceID := d.Get("instance_id")
	functionArn := d.Get("functionArn")

	var matchedLambdaFunction string

	LambdaFunctions, err := dataSourceAwsConnectGetAllLambdaFunctionAssociations(ctx, conn, instanceID.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing Connect Lambda Functions: %s", err))
	}

	for _, LambdaFunctionArn := range LambdaFunctions {
		log.Printf("[DEBUG] Connect Lambda Function Association: %s", LambdaFunctionArn)
		if aws.StringValue(LambdaFunctionArn) == functionArn.(string) {
			matchedLambdaFunction = aws.StringValue(LambdaFunctionArn)
			break
		}
	}
	d.Set("function_arn", matchedLambdaFunction)
	d.Set("instance_id", instanceID)
	d.SetId(fmt.Sprintf("%s:%s", instanceID, d.Get("function_arn").(string)))

	return nil
}

func dataSourceAwsConnectGetAllLambdaFunctionAssociations(ctx context.Context, conn *connect.Connect, instanceID string) ([]*string, error) {
	var functionArns []*string
	var nextToken string

	for {
		input := &connect.ListLambdaFunctionsInput{
			InstanceId: aws.String(instanceID),
			MaxResults: aws.Int64(int64(tfconnect.ListLambdaFunctionsMaxResults)),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		log.Printf("[DEBUG] Listing Connect Lambda Functions: %s", input)

		output, err := conn.ListLambdaFunctionsWithContext(ctx, input)
		if err != nil {
			return functionArns, err
		}
		functionArns = append(functionArns, output.LambdaFunctions...)

		if output.NextToken == nil {
			break
		}
		nextToken = aws.StringValue(output.NextToken)
	}

	return functionArns, nil
}
