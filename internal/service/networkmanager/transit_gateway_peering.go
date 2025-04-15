// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_transit_gateway_peering", name="Transit Gateway Peering")
// @Tags(identifierAttribute="arn")
func resourceTransitGatewayPeering() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayPeeringCreate,
		ReadWithoutTimeout:   resourceTransitGatewayPeeringRead,
		UpdateWithoutTimeout: resourceTransitGatewayPeeringUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayPeeringDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"edge_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peering_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"transit_gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"transit_gateway_peering_attachment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTransitGatewayPeeringCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	coreNetworkID := d.Get("core_network_id").(string)
	transitGatewayARN := d.Get("transit_gateway_arn").(string)
	input := &networkmanager.CreateTransitGatewayPeeringInput{
		CoreNetworkId:     aws.String(coreNetworkID),
		Tags:              getTagsIn(ctx),
		TransitGatewayArn: aws.String(transitGatewayARN),
	}

	log.Printf("[DEBUG] Creating Network Manager Transit Gateway Peering: %#v", input)
	output, err := conn.CreateTransitGatewayPeering(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Transit Gateway (%s) Peering (%s): %s", transitGatewayARN, coreNetworkID, err)
	}

	d.SetId(aws.ToString(output.TransitGatewayPeering.Peering.PeeringId))

	if _, err := waitTransitGatewayPeeringCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Transit Gateway Peering (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPeeringRead(ctx, d, meta)...)
}

func resourceTransitGatewayPeeringRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	transitGatewayPeering, err := findTransitGatewayPeeringByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Transit Gateway Peering %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Peering (%s): %s", d.Id(), err)
	}

	p := transitGatewayPeering.Peering
	d.Set(names.AttrARN, peeringARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set("core_network_arn", p.CoreNetworkArn)
	d.Set("core_network_id", p.CoreNetworkId)
	d.Set("edge_location", p.EdgeLocation)
	d.Set(names.AttrOwnerAccountID, p.OwnerAccountId)
	d.Set("peering_type", p.PeeringType)
	d.Set(names.AttrResourceARN, p.ResourceArn)
	d.Set("transit_gateway_arn", transitGatewayPeering.TransitGatewayArn)
	d.Set("transit_gateway_peering_attachment_id", transitGatewayPeering.TransitGatewayPeeringAttachmentId)

	setTagsOut(ctx, p.Tags)

	return diags
}

func resourceTransitGatewayPeeringUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceTransitGatewayPeeringRead(ctx, d, meta)
}

func resourceTransitGatewayPeeringDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	log.Printf("[DEBUG] Deleting Network Manager Transit Gateway Peering: %s", d.Id())
	_, err := conn.DeletePeering(ctx, &networkmanager.DeletePeeringInput{
		PeeringId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Transit Gateway Peering (%s): %s", d.Id(), err)
	}

	if _, err := waitTransitGatewayPeeringDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Transit Gateway Peering (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findTransitGatewayPeeringByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.TransitGatewayPeering, error) {
	input := &networkmanager.GetTransitGatewayPeeringInput{
		PeeringId: aws.String(id),
	}

	output, err := conn.GetTransitGatewayPeering(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TransitGatewayPeering == nil || output.TransitGatewayPeering.Peering == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TransitGatewayPeering, nil
}

func statusTransitGatewayPeeringState(ctx context.Context, conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findTransitGatewayPeeringByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Peering.State), nil
	}
}

func waitTransitGatewayPeeringCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayPeering, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PeeringStateCreating),
		Target:  enum.Slice(awstypes.PeeringStateAvailable),
		Timeout: timeout,
		Refresh: statusTransitGatewayPeeringState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPeering); ok {
		tfresource.SetLastError(err, peeringsError(output.Peering.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitTransitGatewayPeeringDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayPeering, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PeeringStateDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusTransitGatewayPeeringState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPeering); ok {
		tfresource.SetLastError(err, peeringsError(output.Peering.LastModificationErrors))

		return output, err
	}

	return nil, err
}

// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsnetworkmanager.html#awsnetworkmanager-resources-for-iam-policies.
func peeringARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.GlobalARN(ctx, "networkmanager", "peering/"+id)
}
