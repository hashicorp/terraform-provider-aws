// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_lambda_function_association")
func ResourceLambdaFunctionAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLambdaFunctionAssociationCreate,
		ReadWithoutTimeout:   resourceLambdaFunctionAssociationRead,
		DeleteWithoutTimeout: resourceLambdaFunctionAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrFunctionARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLambdaFunctionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceId := d.Get(names.AttrInstanceID).(string)
	functionArn := d.Get(names.AttrFunctionARN).(string)

	input := &connect.AssociateLambdaFunctionInput{
		InstanceId:  aws.String(instanceId),
		FunctionArn: aws.String(functionArn),
	}

	_, err := conn.AssociateLambdaFunctionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Lambda Function Association (%s,%s): %s", instanceId, functionArn, err)
	}

	d.SetId(LambdaFunctionAssociationCreateResourceID(instanceId, functionArn))

	return append(diags, resourceLambdaFunctionAssociationRead(ctx, d, meta)...)
}

func resourceLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, functionArn, err := LambdaFunctionAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	lfaArn, err := FindLambdaFunctionAssociationByARNWithContext(ctx, conn, instanceID, functionArn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Lambda Function Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Lambda Function Association by Function ARN (%s): %s", functionArn, err)
	}

	d.Set(names.AttrFunctionARN, lfaArn)
	d.Set(names.AttrInstanceID, instanceID)

	return diags
}

func resourceLambdaFunctionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, functionArn, err := LambdaFunctionAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &connect.DisassociateLambdaFunctionInput{
		InstanceId:  aws.String(instanceID),
		FunctionArn: aws.String(functionArn),
	}

	_, err = conn.DisassociateLambdaFunctionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Lambda Function Association (%s): %s", d.Id(), err)
	}

	return diags
}
