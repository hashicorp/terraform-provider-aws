// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_vpc_link", name="VPC Link")
// @Tags(identifierAttribute="arn")
func resourceVPCLink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCLinkCreate,
		ReadWithoutTimeout:   resourceVPCLinkRead,
		UpdateWithoutTimeout: resourceVPCLinkUpdate,
		DeleteWithoutTimeout: resourceVPCLinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_arns": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &apigateway.CreateVpcLinkInput{
		Name:       aws.String(name),
		Tags:       getTagsIn(ctx),
		TargetArns: flex.ExpandStringValueList(d.Get("target_arns").([]interface{})),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateVpcLink(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway VPC Link (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	if _, err := waitVPCLinkAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway VPC Link (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCLinkRead(ctx, d, meta)...)
}

func resourceVPCLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	vpcLink, err := findVPCLinkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway VPC Link %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway VPC Link (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, vpcLinkARN(meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrDescription, vpcLink.Description)
	d.Set(names.AttrName, vpcLink.Name)
	d.Set("target_arns", vpcLink.TargetArns)

	setTagsOut(ctx, vpcLink.Tags)

	return diags
}

func resourceVPCLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		operations := make([]types.PatchOperation, 0)

		if d.HasChange(names.AttrDescription) {
			operations = append(operations, types.PatchOperation{
				Op:    types.Op("replace"),
				Path:  aws.String("/description"),
				Value: aws.String(d.Get(names.AttrDescription).(string)),
			})
		}

		if d.HasChange(names.AttrName) {
			operations = append(operations, types.PatchOperation{
				Op:    types.Op("replace"),
				Path:  aws.String("/name"),
				Value: aws.String(d.Get(names.AttrName).(string)),
			})
		}

		input := &apigateway.UpdateVpcLinkInput{
			PatchOperations: operations,
			VpcLinkId:       aws.String(d.Id()),
		}

		_, err := conn.UpdateVpcLink(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway VPC Link (%s): %s", d.Id(), err)
		}

		if _, err := waitVPCLinkAvailable(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for API Gateway VPC Link (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCLinkRead(ctx, d, meta)...)
}

func resourceVPCLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway VPC Link: %s", d.Id())
	_, err := conn.DeleteVpcLink(ctx, &apigateway.DeleteVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway VPC Link (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCLinkDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway VPC Link (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findVPCLinkByID(ctx context.Context, conn *apigateway.Client, id string) (*apigateway.GetVpcLinkOutput, error) {
	input := &apigateway.GetVpcLinkInput{
		VpcLinkId: aws.String(id),
	}

	output, err := conn.GetVpcLink(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
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

func vpcLinkStatus(ctx context.Context, conn *apigateway.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCLinkByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitVPCLinkAvailable(ctx context.Context, conn *apigateway.Client, id string) (*apigateway.GetVpcLinkOutput, error) { //nolint:unparam
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VpcLinkStatusPending),
		Target:     enum.Slice(types.VpcLinkStatusAvailable),
		Refresh:    vpcLinkStatus(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apigateway.GetVpcLinkOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitVPCLinkDeleted(ctx context.Context, conn *apigateway.Client, id string) (*apigateway.GetVpcLinkOutput, error) {
	const (
		timeout = 20 * time.Minute
	)
	stateConf := retry.StateChangeConf{
		Pending:    enum.Slice(types.VpcLinkStatusPending, types.VpcLinkStatusAvailable, types.VpcLinkStatusDeleting),
		Target:     []string{},
		Timeout:    timeout,
		MinTimeout: 1 * time.Second,
		Refresh:    vpcLinkStatus(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apigateway.GetVpcLinkOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func vpcLinkARN(c *conns.AWSClient, vpcLinkID string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   "apigateway",
		Region:    c.Region,
		Resource:  fmt.Sprintf("/vpclinks/%s", vpcLinkID),
	}.String()
}
