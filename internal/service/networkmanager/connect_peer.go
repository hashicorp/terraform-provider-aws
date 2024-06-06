// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
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

// @SDKResource("aws_networkmanager_connect_peer", name="Connect Peer")
// @Tags(identifierAttribute="arn")
func ResourceConnectPeer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectPeerCreate,
		ReadWithoutTimeout:   resourceConnectPeerRead,
		UpdateWithoutTimeout: resourceConnectPeerUpdate,
		DeleteWithoutTimeout: resourceConnectPeerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_options": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"peer_asn": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 2147483647),
						},
					},
				},
			},
			names.AttrConfiguration: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bgp_configurations": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"core_network_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"core_network_asn": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"peer_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"peer_asn": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"core_network_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"inside_cidr_blocks": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"peer_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrProtocol: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"connect_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 50),
					validation.StringMatch(regexache.MustCompile(`^attachment-([0-9a-f]{8,17})$`), "Must start with attachment and then have 8 to 17 characters"),
				),
			},
			"connect_peer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_address": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 50),
					validation.StringMatch(regexache.MustCompile(`[\s\S]*`), "Anything but whitespace"),
				),
			},
			"core_network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"edge_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inside_cidr_blocks": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsCIDR,
				},
			},
			"peer_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 50),
					validation.StringMatch(regexache.MustCompile(`[\s\S]*`), "Anything but whitespace"),
				),
			},
			"subnet_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 500),
					validation.StringMatch(regexache.MustCompile(`^arn:[^:]{1,63}:ec2:[^:]{0,63}:[^:]{0,63}:subnet\/subnet-[0-9a-f]{8,17}$|^$`), "Must be a valid subnet ARN"),
				),
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConnectPeerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	connectAttachmentID := d.Get("connect_attachment_id").(string)
	// insideCIDRBlocks := flex.ExpandStringList(d.Get("inside_cidr_blocks").([]interface{}))
	peerAddress := d.Get("peer_address").(string)
	input := &networkmanager.CreateConnectPeerInput{
		ConnectAttachmentId: aws.String(connectAttachmentID),
		// InsideCidrBlocks:    insideCIDRBlocks,
		PeerAddress: aws.String(peerAddress),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("bgp_options"); ok && len(v.([]interface{})) > 0 {
		input.BgpOptions = expandPeerOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("core_network_address"); ok {
		input.CoreNetworkAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("inside_cidr_blocks"); ok {
		insideCIDRBlocks := flex.ExpandStringList(v.([]interface{}))
		input.InsideCidrBlocks = insideCIDRBlocks
	}

	if v, ok := d.GetOk("subnet_arn"); ok {
		input.SubnetArn = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.CreateConnectPeerWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			// Connect Peer doesn't have direct dependency to Connect attachment state when using Attachment Accepter.
			// Waiting for Create Timeout period for Connect Attachment to come available state.
			// Only needed if depends_on statement is not used in Connect Peer.
			//
			// ValidationException: Connect attachment state is invalid.
			// Error: creating Connect Peer: ValidationException: Connect attachment state is invalid. attachment id: attachment-06cb63ed3fe0008df
			// {
			//   RespMetadata: {
			// 	StatusCode: 400,
			// 	RequestID: "c5f0f9de-ad7f-411a-ba2e-7c37ea397255"
			//   },
			//   Message_: "Connect attachment state is invalid. attachment id: attachment-06cb63ed3fe0008df",
			//   Reason: "Other"
			// }
			if validationExceptionMessage_Contains(err, networkmanager.ValidationExceptionReasonOther, "Connect attachment state is invalid") {
				return true, err
			}

			return false, err
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Peer: %s", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*networkmanager.CreateConnectPeerOutput).ConnectPeer.ConnectPeerId))

	if _, err := waitConnectPeerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connect Peer (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConnectPeerRead(ctx, d, meta)...)
}

func resourceConnectPeerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	connectPeer, err := FindConnectPeerByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Connect Peer %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Connect Peer (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "networkmanager",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("connect-peer/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	bgpOptions := map[string]interface{}{}
	bgpOptions["peer_asn"] = connectPeer.Configuration.BgpConfigurations[0].PeerAsn
	d.Set("bgp_options", []interface{}{bgpOptions})
	d.Set(names.AttrConfiguration, []interface{}{flattenPeerConfiguration(connectPeer.Configuration)})
	d.Set("connect_peer_id", connectPeer.ConnectPeerId)
	d.Set("core_network_id", connectPeer.CoreNetworkId)
	if connectPeer.CreatedAt != nil {
		d.Set(names.AttrCreatedAt, aws.TimeValue(connectPeer.CreatedAt).Format(time.RFC3339))
	} else {
		d.Set(names.AttrCreatedAt, nil)
	}
	d.Set("edge_location", connectPeer.EdgeLocation)
	d.Set("connect_attachment_id", connectPeer.ConnectAttachmentId)
	d.Set("inside_cidr_blocks", connectPeer.Configuration.InsideCidrBlocks)
	d.Set("peer_address", connectPeer.Configuration.PeerAddress)
	d.Set("subnet_arn", connectPeer.SubnetArn)
	d.Set(names.AttrState, connectPeer.State)

	setTagsOut(ctx, connectPeer.Tags)

	return diags
}

func resourceConnectPeerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceConnectPeerRead(ctx, d, meta)
}

func resourceConnectPeerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	log.Printf("[DEBUG] Deleting Network Manager Connect Peer: %s", d.Id())
	_, err := conn.DeleteConnectPeerWithContext(ctx, &networkmanager.DeleteConnectPeerInput{
		ConnectPeerId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Connect Peer (%s): %s", d.Id(), err)
	}

	if _, err := waitConnectPeerDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connect Peer (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandPeerOptions(o map[string]interface{}) *networkmanager.BgpOptions {
	if o == nil {
		return nil
	}

	object := &networkmanager.BgpOptions{}

	if v, ok := o["peer_asn"].(int); ok {
		object.PeerAsn = aws.Int64(int64(v))
	}

	return object
}

func FindConnectPeerByID(ctx context.Context, conn *networkmanager.NetworkManager, id string) (*networkmanager.ConnectPeer, error) {
	input := &networkmanager.GetConnectPeerInput{
		ConnectPeerId: aws.String(id),
	}

	output, err := conn.GetConnectPeerWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ConnectPeer == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ConnectPeer, nil
}

func flattenPeerConfiguration(apiObject *networkmanager.ConnectPeerConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	confMap := map[string]interface{}{}

	if v := apiObject.BgpConfigurations; v != nil {
		bgpConfMap := map[string]interface{}{}

		if a := v[0].CoreNetworkAddress; a != nil {
			bgpConfMap["core_network_address"] = aws.StringValue(a)
		}
		if a := v[0].CoreNetworkAsn; a != nil {
			bgpConfMap["core_network_asn"] = aws.Int64Value(a)
		}
		if a := v[0].PeerAddress; a != nil {
			bgpConfMap["peer_address"] = aws.StringValue(a)
		}
		if a := v[0].PeerAsn; a != nil {
			bgpConfMap["peer_asn"] = aws.Int64Value(a)
		}
		confMap["bgp_configurations"] = []interface{}{bgpConfMap}
	}
	if v := apiObject.CoreNetworkAddress; v != nil {
		confMap["core_network_address"] = aws.StringValue(v)
	}
	if v := apiObject.InsideCidrBlocks; v != nil {
		confMap["inside_cidr_blocks"] = aws.StringValueSlice(v)
	}
	if v := apiObject.PeerAddress; v != nil {
		confMap["peer_address"] = aws.StringValue(v)
	}
	if v := apiObject.Protocol; v != nil {
		confMap[names.AttrProtocol] = aws.StringValue(v)
	}

	return confMap
}

func statusConnectPeerState(ctx context.Context, conn *networkmanager.NetworkManager, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindConnectPeerByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitConnectPeerCreated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.ConnectPeer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.ConnectPeerStateCreating},
		Target:  []string{networkmanager.ConnectPeerStateAvailable},
		Timeout: timeout,
		Refresh: statusConnectPeerState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.ConnectPeer); ok {
		return output, err
	}

	return nil, err
}

func waitConnectPeerDeleted(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.ConnectPeer, error) {
	stateconf := &retry.StateChangeConf{
		Pending:        []string{networkmanager.ConnectPeerStateDeleting},
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusConnectPeerState(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateconf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.ConnectPeer); ok {
		return output, err
	}

	return nil, err
}

// validationExceptionMessage_Contains returns true if the error matches all these conditions:
//   - err is of type networkmanager.ValidationException
//   - ValidationException.Reason equals reason
//   - ValidationException.Message_ contains message
func validationExceptionMessage_Contains(err error, reason string, message string) bool {
	var validationException *networkmanager.ValidationException

	if errors.As(err, &validationException) && aws.StringValue(validationException.Reason) == reason {
		if strings.Contains(aws.StringValue(validationException.Message_), message) {
			return true
		}
	}

	return false
}
