// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_lag", name="LAG")
// @Tags(identifierAttribute="arn")
func resourceLag() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLagCreate,
		ReadWithoutTimeout:   resourceLagRead,
		UpdateWithoutTimeout: resourceLagUpdate,
		DeleteWithoutTimeout: resourceLagDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"connections_bandwidth": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validConnectionBandWidth(),
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"has_logical_redundancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrLocation: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrProviderName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceLagCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &directconnect.CreateLagInput{
		ConnectionsBandwidth: aws.String(d.Get("connections_bandwidth").(string)),
		LagName:              aws.String(name),
		Location:             aws.String(d.Get(names.AttrLocation).(string)),
		Tags:                 getTagsIn(ctx),
	}

	var connectionIDSpecified bool
	if v, ok := d.GetOk(names.AttrConnectionID); ok {
		connectionIDSpecified = true
		input.ConnectionId = aws.String(v.(string))
		input.NumberOfConnections = int32(1)
	} else {
		input.NumberOfConnections = int32(1)
	}

	if v, ok := d.GetOk(names.AttrProviderName); ok {
		input.ProviderName = aws.String(v.(string))
	}

	output, err := conn.CreateLag(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect LAG (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.LagId))

	// Delete unmanaged connection.
	if !connectionIDSpecified {
		if err := deleteConnection(ctx, conn, aws.ToString(output.Connections[0].ConnectionId), waitConnectionDeleted); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceLagRead(ctx, d, meta)...)
}

func resourceLagRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	lag, err := findLagByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect LAG (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect LAG (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    aws.ToString(lag.Region),
		Service:   "directconnect",
		AccountID: aws.ToString(lag.OwnerAccount),
		Resource:  fmt.Sprintf("dxlag/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("connections_bandwidth", lag.ConnectionsBandwidth)
	d.Set("has_logical_redundancy", lag.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", lag.JumboFrameCapable)
	d.Set(names.AttrLocation, lag.Location)
	d.Set(names.AttrName, lag.LagName)
	d.Set(names.AttrOwnerAccountID, lag.OwnerAccount)
	d.Set(names.AttrProviderName, lag.ProviderName)

	return diags
}

func resourceLagUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if d.HasChange(names.AttrName) {
		input := &directconnect.UpdateLagInput{
			LagId:   aws.String(d.Id()),
			LagName: aws.String(d.Get(names.AttrName).(string)),
		}

		_, err := conn.UpdateLag(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Direct Connect LAG (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLagRead(ctx, d, meta)...)
}

func resourceLagDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if d.Get(names.AttrForceDestroy).(bool) {
		lag, err := findLagByID(ctx, conn, d.Id())

		if tfresource.NotFound(err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Direct Connect LAG (%s): %s", d.Id(), err)
		}

		for _, connection := range lag.Connections {
			if err := deleteConnection(ctx, conn, aws.ToString(connection.ConnectionId), waitConnectionDeleted); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	} else if v, ok := d.GetOk(names.AttrConnectionID); ok {
		if err := deleteConnectionLAGAssociation(ctx, conn, v.(string), d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[DEBUG] Deleting Direct Connect LAG: %s", d.Id())
	input := directconnect.DeleteLagInput{
		LagId: aws.String(d.Id()),
	}
	_, err := conn.DeleteLag(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Lag with ID") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect LAG (%s): %s", d.Id(), err)
	}

	if _, err := waitLagDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect LAG (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findLagByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Lag, error) {
	input := &directconnect.DescribeLagsInput{
		LagId: aws.String(id),
	}
	output, err := findLag(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Lag]())

	if err != nil {
		return nil, err
	}

	if state := output.LagState; state == awstypes.LagStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findLag(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeLagsInput, filter tfslices.Predicate[*awstypes.Lag]) (*awstypes.Lag, error) {
	output, err := findLags(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLags(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeLagsInput, filter tfslices.Predicate[*awstypes.Lag]) ([]awstypes.Lag, error) {
	output, err := conn.DescribeLags(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Lag with ID") {
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

	return tfslices.Filter(output.Lags, tfslices.PredicateValue(filter)), nil
}

func statusLag(ctx context.Context, conn *directconnect.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findLagByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.LagState), nil
	}
}

func waitLagDeleted(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Lag, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LagStateAvailable, awstypes.LagStateRequested, awstypes.LagStatePending, awstypes.LagStateDeleting),
		Target:  []string{},
		Refresh: statusLag(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Lag); ok {
		return output, err
	}

	return nil, err
}
