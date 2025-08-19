// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_connection", name="Connection")
// @Tags(identifierAttribute="arn")
func resourceConnection() *schema.Resource {
	// Resource with v0 schema (provider v5.0.1).
	resourceV0 := &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"encryption_mode": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
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
			"macsec_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"request_macsec": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"partner_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_encryption_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrProviderName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vlan_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type: resourceV0.CoreConfigSchema().ImpliedType(),
				Upgrade: func(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
					// Convert vlan_id from string to int.
					if v, ok := rawState["vlan_id"]; ok {
						if v, ok := v.(string); ok {
							if v == "" {
								rawState["vlan_id"] = 0
							} else {
								if v, err := strconv.Atoi(v); err == nil {
									rawState["vlan_id"] = v
								} else {
									return nil, err
								}
							}
						}
					}

					return rawState, nil
				},
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validConnectionBandWidth(),
			},
			// The MAC Security (MACsec) connection encryption mode.
			"encryption_mode": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"no_encrypt", "should_encrypt", "must_encrypt"}, false),
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
			// Indicates whether the connection supports MAC Security (MACsec).
			"macsec_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			// Enable or disable MAC Security (MACsec) on this connection.
			"request_macsec": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"partner_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_encryption_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrProviderName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vlan_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &directconnect.CreateConnectionInput{
		Bandwidth:      aws.String(d.Get("bandwidth").(string)),
		ConnectionName: aws.String(name),
		Location:       aws.String(d.Get(names.AttrLocation).(string)),
		RequestMACSec:  aws.Bool(d.Get("request_macsec").(bool)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrProviderName); ok {
		input.ProviderName = aws.String(v.(string))
	}

	output, err := conn.CreateConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Connection (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ConnectionId))

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	connection, err := findConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Connection (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    aws.ToString(connection.Region),
		Service:   "directconnect",
		AccountID: aws.ToString(connection.OwnerAccount),
		Resource:  fmt.Sprintf("dxcon/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("aws_device", connection.AwsDeviceV2)
	d.Set("bandwidth", connection.Bandwidth)
	d.Set("encryption_mode", connection.EncryptionMode)
	d.Set("has_logical_redundancy", connection.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", connection.JumboFrameCapable)
	d.Set(names.AttrLocation, connection.Location)
	d.Set("macsec_capable", connection.MacSecCapable)
	d.Set(names.AttrName, connection.ConnectionName)
	d.Set(names.AttrOwnerAccountID, connection.OwnerAccount)
	d.Set("partner_name", connection.PartnerName)
	d.Set("port_encryption_status", connection.PortEncryptionStatus)
	d.Set(names.AttrProviderName, connection.ProviderName)
	if !d.IsNewResource() && !d.Get("request_macsec").(bool) {
		d.Set("request_macsec", aws.Bool(false))
	}
	d.Set("vlan_id", connection.Vlan)

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if d.HasChange("encryption_mode") {
		input := &directconnect.UpdateConnectionInput{
			ConnectionId:   aws.String(d.Id()),
			EncryptionMode: aws.String(d.Get("encryption_mode").(string)),
		}

		_, err := conn.UpdateConnection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Direct Connect Connection (%s): %s", d.Id(), err)
		}

		if _, err := waitConnectionConfirmed(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Connection (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if _, ok := d.GetOk(names.AttrSkipDestroy); ok {
		return diags
	}

	if err := deleteConnection(ctx, conn, d.Id(), waitConnectionDeleted); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func deleteConnection(ctx context.Context, conn *directconnect.Client, connectionID string, waiter func(context.Context, *directconnect.Client, string) (*awstypes.Connection, error)) error {
	log.Printf("[DEBUG] Deleting Direct Connect Connection: %s", connectionID)
	input := directconnect.DeleteConnectionInput{
		ConnectionId: aws.String(connectionID),
	}
	_, err := conn.DeleteConnection(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Connection with ID") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Direct Connect Connection (%s): %w", connectionID, err)
	}

	if _, err := waiter(ctx, conn, connectionID); err != nil {
		return fmt.Errorf("waiting for Direct Connect Connection (%s): %w", connectionID, err)
	}

	return nil
}

func findConnectionByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Connection, error) {
	input := &directconnect.DescribeConnectionsInput{
		ConnectionId: aws.String(id),
	}
	output, err := findConnection(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Connection]())

	if err != nil {
		return nil, err
	}

	if state := output.ConnectionState; state == awstypes.ConnectionStateDeleted || state == awstypes.ConnectionStateRejected {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findConnection(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeConnectionsInput, filter tfslices.Predicate[*awstypes.Connection]) (*awstypes.Connection, error) {
	output, err := findConnections(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConnections(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeConnectionsInput, filter tfslices.Predicate[*awstypes.Connection]) ([]awstypes.Connection, error) {
	output, err := conn.DescribeConnections(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Connection with ID") {
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

	return tfslices.Filter(output.Connections, tfslices.PredicateValue(filter)), nil
}

func statusConnection(ctx context.Context, conn *directconnect.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ConnectionState), nil
	}
}

func waitConnectionDeleted(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Connection, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStatePending, awstypes.ConnectionStateOrdering, awstypes.ConnectionStateAvailable, awstypes.ConnectionStateRequested, awstypes.ConnectionStateDeleting),
		Target:  []string{},
		Refresh: statusConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Connection); ok {
		return output, err
	}

	return nil, err
}
