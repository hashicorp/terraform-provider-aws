package connect

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLambdaFunctionAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLambdaFunctionAssociationCreate,
		ReadContext:   resourceLambdaFunctionAssociationRead,
		UpdateContext: resourceLambdaFunctionAssociationRead,
		DeleteContext: resourceLambdaFunctionAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				instanceID, functionArn, err := LambdaFunctionAssociationParseResourceID(d.Id())
				if err != nil {
					return nil, err
				}
				d.Set("function_arn", functionArn)
				d.Set("instance_id", instanceID)
				d.SetId(LambdaFunctionAssociationCreateResourceID(instanceID, functionArn))

				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(connectLambdaFunctionAssociationCreatedTimeout),
			Delete: schema.DefaultTimeout(connectLambdaFunctionAssociationDeletedTimeout),
		},
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

func resourceLambdaFunctionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	input := &connect.AssociateLambdaFunctionInput{
		InstanceId:  aws.String(d.Get("instance_id").(string)),
		FunctionArn: aws.String(d.Get("function_arn").(string)),
	}

	lfaId := LambdaFunctionAssociationCreateResourceID(d.Get("instance_id").(string), d.Get("function_arn").(string))

	lfaArn, err := FindLambdaFunctionAssociationByArnWithContext(ctx, conn, d.Get("instance_id").(string), d.Get("function_arn").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Lambda Function Association by Function ARN (%s): %w", lfaArn, err))
	}
	log.Printf("[DEBUG] Creating Connect Lambda Association %s", input)

	_, err = conn.AssociateLambdaFunctionWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Lambda Function Association (%s): %s", lfaArn, err))
	}

	d.SetId(lfaId)
	return resourceLambdaFunctionAssociationRead(ctx, d, meta)
}

func resourceLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, functionArn, err := LambdaFunctionAssociationParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	lfaArn, err := FindLambdaFunctionAssociationByArnWithContext(ctx, conn, instanceID, functionArn)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Lambda Function Association by Function ARN (%s): %w", functionArn, err))
	}

	if !d.IsNewResource() && lfaArn == "" {
		log.Printf("[WARN] Connect Lambda Function Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("function_arn", functionArn)
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

	log.Printf("[DEBUG] Deleting Connect Lambda Function Association %s", d.Id())

	_, err = conn.DisassociateLambdaFunction(input)

	if err != nil && !tfawserr.ErrCodeEquals(err, "ResourceNotFoundException", "") {
		return diag.FromErr(fmt.Errorf("error deleting Connect Lambda Function Association (%s): %w", d.Id(), err))
	}
	return nil
}
