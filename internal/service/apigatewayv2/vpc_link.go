// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apigatewayv2_vpc_link", name="VPC Link")
// @Tags(identifierAttribute="arn")
func ResourceVPCLink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCLinkCreate,
		ReadWithoutTimeout:   resourceVPCLinkRead,
		UpdateWithoutTimeout: resourceVPCLinkUpdate,
		DeleteWithoutTimeout: resourceVPCLinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	name := d.Get("name").(string)
	input := &apigatewayv2.CreateVpcLinkInput{
		Name:             aws.String(name),
		SecurityGroupIds: flex.ExpandStringSet(d.Get("security_group_ids").(*schema.Set)),
		SubnetIds:        flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:             getTagsIn(ctx),
	}

	output, err := conn.CreateVpcLinkWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 VPC Link (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.VpcLinkId))

	if _, err := waitVPCLinkAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 VPC Link (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCLinkRead(ctx, d, meta)...)
}

func resourceVPCLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	output, err := FindVPCLinkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 VPC Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/vpclinks/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("name", output.Name)
	d.Set("security_group_ids", aws.StringValueSlice(output.SecurityGroupIds))
	d.Set("subnet_ids", aws.StringValueSlice(output.SubnetIds))

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceVPCLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &apigatewayv2.UpdateVpcLinkInput{
			Name:      aws.String(d.Get("name").(string)),
			VpcLinkId: aws.String(d.Id()),
		}

		_, err := conn.UpdateVpcLinkWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 VPC Link (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCLinkRead(ctx, d, meta)...)
}

func resourceVPCLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 VPC Link: %s", d.Id())
	_, err := conn.DeleteVpcLinkWithContext(ctx, &apigatewayv2.DeleteVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCLinkDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 VPC Link (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindVPCLinkByID(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, id string) (*apigatewayv2.GetVpcLinkOutput, error) {
	input := &apigatewayv2.GetVpcLinkInput{
		VpcLinkId: aws.String(id),
	}

	return findVPCLink(ctx, conn, input)
}

func findVPCLink(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetVpcLinkInput) (*apigatewayv2.GetVpcLinkOutput, error) {
	output, err := conn.GetVpcLinkWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusVPCLink(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCLinkByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.VpcLinkStatus), nil
	}
}

func waitVPCLinkAvailable(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, id string) (*apigatewayv2.GetVpcLinkOutput, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{apigatewayv2.VpcLinkStatusPending},
		Target:  []string{apigatewayv2.VpcLinkStatusAvailable},
		Refresh: statusVPCLink(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apigatewayv2.GetVpcLinkOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.VpcLinkStatusMessage)))

		return output, err
	}

	return nil, err
}

func waitVPCLinkDeleted(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, vpcLinkId string) (*apigatewayv2.GetVpcLinkOutput, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{apigatewayv2.VpcLinkStatusDeleting},
		Target:  []string{},
		Refresh: statusVPCLink(ctx, conn, vpcLinkId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apigatewayv2.GetVpcLinkOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.VpcLinkStatusMessage)))

		return output, err
	}

	return nil, err
}
