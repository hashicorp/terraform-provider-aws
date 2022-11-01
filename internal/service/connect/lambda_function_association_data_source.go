package connect

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceLambdaFunctionAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceLambdaFunctionAssociationRead,
		Schema: map[string]*schema.Schema{
			"function_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	functionArn := d.Get("function_arn")
	instanceID := d.Get("instance_id")

	lfaArn, err := FindLambdaFunctionAssociationByARNWithContext(ctx, conn, instanceID.(string), functionArn.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Lambda Function Association by ARN (%s): %w", functionArn, err))
	}

	if lfaArn == "" {
		return diag.FromErr(fmt.Errorf("error finding Connect Lambda Function Association by ARN (%s): not found", functionArn))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("function_arn", functionArn)
	d.Set("instance_id", instanceID)

	return nil
}
