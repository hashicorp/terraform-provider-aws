// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_connection", name="Connection")
// @Tags(identifierAttribute="arn")
func resourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				parsedARN, err := arn.Parse(d.Id())

				if err != nil {
					return nil, fmt.Errorf("parsing ARN (%s): %w", d.Id(), err)
				}

				// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_networkmanager.html#networkmanager-resources-for-iam-policies.
				resourceParts := strings.Split(parsedARN.Resource, "/")

				if actual, expected := len(resourceParts), 3; actual < expected {
					return nil, fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, d.Id(), actual)
				}

				d.SetId(resourceParts[2])
				d.Set("global_network_id", resourceParts[1])

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connected_device_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"connected_link_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"device_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"link_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	input := &networkmanager.CreateConnectionInput{
		ConnectedDeviceId: aws.String(d.Get("connected_device_id").(string)),
		DeviceId:          aws.String(d.Get("device_id").(string)),
		GlobalNetworkId:   aws.String(globalNetworkID),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("connected_link_id"); ok {
		input.ConnectedLinkId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("link_id"); ok {
		input.LinkId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Network Manager Connection: %#v", input)
	output, err := conn.CreateConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Connection: %s", err)
	}

	d.SetId(aws.ToString(output.Connection.ConnectionId))

	if _, err := waitConnectionCreated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connection (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	connection, err := findConnectionByTwoPartKey(ctx, conn, globalNetworkID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Connection %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Connection (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, connection.ConnectionArn)
	d.Set("connected_device_id", connection.ConnectedDeviceId)
	d.Set("connected_link_id", connection.ConnectedLinkId)
	d.Set(names.AttrDescription, connection.Description)
	d.Set("device_id", connection.DeviceId)
	d.Set("global_network_id", connection.GlobalNetworkId)
	d.Set("link_id", connection.LinkId)

	setTagsOut(ctx, connection.Tags)

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		globalNetworkID := d.Get("global_network_id").(string)
		input := &networkmanager.UpdateConnectionInput{
			ConnectedLinkId: aws.String(d.Get("connected_link_id").(string)),
			ConnectionId:    aws.String(d.Id()),
			Description:     aws.String(d.Get(names.AttrDescription).(string)),
			GlobalNetworkId: aws.String(globalNetworkID),
			LinkId:          aws.String(d.Get("link_id").(string)),
		}

		log.Printf("[DEBUG] Updating Network Manager Connection: %#v", input)
		_, err := conn.UpdateConnection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Network Manager Connection (%s): %s", d.Id(), err)
		}

		if _, err := waitConnectionUpdated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connection (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)

	log.Printf("[DEBUG] Deleting Network Manager Connection: %s", d.Id())
	_, err := conn.DeleteConnection(ctx, &networkmanager.DeleteConnectionInput{
		ConnectionId:    aws.String(d.Id()),
		GlobalNetworkId: aws.String(globalNetworkID),
	})

	if globalNetworkIDNotFoundError(err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitConnectionDeleted(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connection (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findConnection(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetConnectionsInput) (*awstypes.Connection, error) {
	output, err := findConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output[0], nil
}

func findConnections(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetConnectionsInput) ([]awstypes.Connection, error) {
	var output []awstypes.Connection

	pages := networkmanager.NewGetConnectionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if globalNetworkIDNotFoundError(err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Connections...)
	}

	return output, nil
}

func findConnectionByTwoPartKey(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectionID string) (*awstypes.Connection, error) {
	input := &networkmanager.GetConnectionsInput{
		ConnectionIds:   []string{connectionID},
		GlobalNetworkId: aws.String(globalNetworkID),
	}

	output, err := findConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalNetworkId) != globalNetworkID || aws.ToString(output.ConnectionId) != connectionID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusConnectionState(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectionID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findConnectionByTwoPartKey(ctx, conn, globalNetworkID, connectionID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitConnectionCreated(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectionID string, timeout time.Duration) (*awstypes.Connection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStatePending),
		Target:  enum.Slice(awstypes.ConnectionStateAvailable),
		Timeout: timeout,
		Refresh: statusConnectionState(ctx, conn, globalNetworkID, connectionID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitConnectionDeleted(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectionID string, timeout time.Duration) (*awstypes.Connection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStateDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusConnectionState(ctx, conn, globalNetworkID, connectionID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitConnectionUpdated(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectionID string, timeout time.Duration) (*awstypes.Connection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStateUpdating),
		Target:  enum.Slice(awstypes.ConnectionStateAvailable),
		Timeout: timeout,
		Refresh: statusConnectionState(ctx, conn, globalNetworkID, connectionID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Connection); ok {
		return output, err
	}

	return nil, err
}
