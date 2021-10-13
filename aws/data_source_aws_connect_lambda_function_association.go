package aws

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/finder"
)

func dataSourceAwsConnectLambdaFunctionAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsConnectLambdaFunctionAssociationRead,
		Schema: map[string]*schema.Schema{
			"function_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsConnectLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn
	functionArn := d.Get("function_arn")
	instanceID := d.Get("instance_id")

	lfaArn, err := finder.LambdaFunctionAssociationByArnWithContext(ctx, conn, instanceID.(string), functionArn.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Lambda Function Association by ARN (%s): %w", functionArn, err))
	}

	if lfaArn == "" {
		return diag.FromErr(fmt.Errorf("error finding Connect Lambda Function Association by ARN (%s): not found", functionArn))
	}

	d.Set("function_arn", functionArn)
	d.Set("instance_id", instanceID)
	d.SetId(fmt.Sprintf("%s:%s", d.Get("instance_id").(string), d.Get("function_arn").(string)))

	return nil
}
