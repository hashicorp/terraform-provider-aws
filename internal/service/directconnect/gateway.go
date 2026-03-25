// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package directconnect

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_gateway", name="Gateway")
// @Region(global=true)
// @IdentityAttribute("id")
// @V60SDKv2Fix
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/directconnect/types;awstypes;awstypes.DirectConnectGateway")
// @Testing(randomBgpAsn="64512;65534")
func resourceGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayCreate,
		ReadWithoutTimeout:   resourceGatewayRead,
		UpdateWithoutTimeout: resourceGatewayUpdate,
		DeleteWithoutTimeout: resourceGatewayDelete,

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAmazonSideASN,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceGatewayCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &directconnect.CreateDirectConnectGatewayInput{
		DirectConnectGatewayName: aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.Get("amazon_side_asn").(string); ok && v != "" {
		input.AmazonSideAsn = flex.StringValueToInt64(v)
	}

	output, err := conn.CreateDirectConnectGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Gateway (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.DirectConnectGateway.DirectConnectGatewayId))

	if _, err := waitGatewayCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGatewayRead(ctx, d, meta)...)
}

func resourceGatewayRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	output, err := findGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Direct Connect Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Gateway (%s): %s", d.Id(), err)
	}

	d.Set("amazon_side_asn", flex.Int64ToStringValue(output.AmazonSideAsn))
	d.Set(names.AttrARN, gatewayARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrName, output.DirectConnectGatewayName)
	d.Set(names.AttrOwnerAccountID, output.OwnerAccount)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if d.HasChange(names.AttrName) {
		input := &directconnect.UpdateDirectConnectGatewayInput{
			DirectConnectGatewayId:      aws.String(d.Id()),
			NewDirectConnectGatewayName: aws.String(d.Get(names.AttrName).(string)),
		}

		_, err := conn.UpdateDirectConnectGateway(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Direct Connect Gateway (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGatewayRead(ctx, d, meta)...)
}

func resourceGatewayDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	log.Printf("[DEBUG] Deleting Direct Connect Gateway: %s", d.Id())
	input := directconnect.DeleteDirectConnectGatewayInput{
		DirectConnectGatewayId: aws.String(d.Id()),
	}
	_, err := conn.DeleteDirectConnectGateway(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect Gateway (%s): %s", d.Id(), err)
	}

	if _, err := waitGatewayDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findGatewayByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.DirectConnectGateway, error) {
	input := &directconnect.DescribeDirectConnectGatewaysInput{
		DirectConnectGatewayId: aws.String(id),
	}
	output, err := findGateway(ctx, conn, input, tfslices.PredicateTrue[*awstypes.DirectConnectGateway]())

	if err != nil {
		return nil, err
	}

	if state := output.DirectConnectGatewayState; state == awstypes.DirectConnectGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message: string(state),
		}
	}

	return output, nil
}

func findGateway(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeDirectConnectGatewaysInput, filter tfslices.Predicate[*awstypes.DirectConnectGateway]) (*awstypes.DirectConnectGateway, error) {
	output, err := findGateways(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGateways(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeDirectConnectGatewaysInput, filter tfslices.Predicate[*awstypes.DirectConnectGateway]) ([]awstypes.DirectConnectGateway, error) {
	var output []awstypes.DirectConnectGateway

	err := describeDirectConnectGatewaysPages(ctx, conn, input, func(page *directconnect.DescribeDirectConnectGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DirectConnectGateways {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusGateway(conn *directconnect.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGatewayByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.DirectConnectGatewayState), nil
	}
}

func waitGatewayCreated(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectConnectGatewayStatePending),
		Target:  enum.Slice(awstypes.DirectConnectGatewayStateAvailable),
		Refresh: statusGateway(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGateway); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayDeleted(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectConnectGatewayStatePending, awstypes.DirectConnectGatewayStateAvailable, awstypes.DirectConnectGatewayStateDeleting),
		Target:  []string{},
		Refresh: statusGateway(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGateway); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsdirectconnect.html#awsdirectconnect-resources-for-iam-policies.
func gatewayARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.GlobalARN(ctx, "directconnect", "dx-gateway/"+id)
}
