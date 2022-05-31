package connect

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLambdaFunctionAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLambdaFunctionAssociationCreate,
		ReadContext:   resourceLambdaFunctionAssociationRead,
		DeleteContext: resourceLambdaFunctionAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"function_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLambdaFunctionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceId := d.Get("instance_id").(string)
	functionArn := d.Get("function_arn").(string)

	input := &connect.AssociateLambdaFunctionInput{
		InstanceId:  aws.String(instanceId),
		FunctionArn: aws.String(functionArn),
	}

	_, err := conn.AssociateLambdaFunctionWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Lambda Function Association (%s,%s): %s", instanceId, functionArn, err))
	}

	d.SetId(LambdaFunctionAssociationCreateResourceID(instanceId, functionArn))

	return resourceLambdaFunctionAssociationRead(ctx, d, meta)
}

func resourceLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, functionArn, err := LambdaFunctionAssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	lfaArn, err := FindLambdaFunctionAssociationByARNWithContext(ctx, conn, instanceID, functionArn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Lambda Function Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Lambda Function Association by Function ARN (%s): %w", functionArn, err))
	}

	d.Set("function_arn", lfaArn)
	d.Set("instance_id", instanceID)

	return nil
}

func resourceLambdaFunctionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, functionArn, err := LambdaFunctionAssociationParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &connect.DisassociateLambdaFunctionInput{
		InstanceId:  aws.String(instanceID),
		FunctionArn: aws.String(functionArn),
	}

	_, err = conn.DisassociateLambdaFunctionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Lambda Function Association (%s): %w", d.Id(), err))
	}

	return nil
}
