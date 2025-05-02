// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	lambdaFunctionAssociationResourceIDPartCount = 2
)

// @SDKResource("aws_connect_lambda_function_association", name="Lambda Function Association")
func resourceLambdaFunctionAssociation() *schema.Resource {
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

func resourceLambdaFunctionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	functionARN := d.Get(names.AttrFunctionARN).(string)
	id, err := flex.FlattenResourceId([]string{instanceID, functionARN}, lambdaFunctionAssociationResourceIDPartCount, true)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &connect.AssociateLambdaFunctionInput{
		FunctionArn: aws.String(functionARN),
		InstanceId:  aws.String(instanceID),
	}

	if _, err := conn.AssociateLambdaFunction(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Lambda Function Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceLambdaFunctionAssociationRead(ctx, d, meta)...)
}

func resourceLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), lambdaFunctionAssociationResourceIDPartCount, true)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	instanceID, functionARN := parts[0], parts[1]
	_, err = findLambdaFunctionAssociationByTwoPartKey(ctx, conn, instanceID, functionARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Lambda Function Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Lambda Function Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrFunctionARN, functionARN)
	d.Set(names.AttrInstanceID, instanceID)

	return diags
}

func resourceLambdaFunctionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), lambdaFunctionAssociationResourceIDPartCount, true)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	instanceID, functionARN := parts[0], parts[1]

	log.Printf("[DEBUG] Deleting Connect Lambda Function Association: %s", d.Id())
	input := connect.DisassociateLambdaFunctionInput{
		InstanceId:  aws.String(instanceID),
		FunctionArn: aws.String(functionARN),
	}
	_, err = conn.DisassociateLambdaFunction(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Lambda Function Association (%s): %s", d.Id(), err)
	}

	return diags
}

func findLambdaFunctionAssociationByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, functionARN string) (*string, error) {
	const maxResults = 25
	input := &connect.ListLambdaFunctionsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findLambdaFunction(ctx, conn, input, func(v string) bool {
		return v == functionARN
	})
}

func findLambdaFunction(ctx context.Context, conn *connect.Client, input *connect.ListLambdaFunctionsInput, filter tfslices.Predicate[string]) (*string, error) {
	output, err := findLambdaFunctions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLambdaFunctions(ctx context.Context, conn *connect.Client, input *connect.ListLambdaFunctionsInput, filter tfslices.Predicate[string]) ([]string, error) {
	var output []string

	pages := connect.NewListLambdaFunctionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.LambdaFunctions {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
