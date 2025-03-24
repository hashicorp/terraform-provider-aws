// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_connect_peer", name="Connect Peer")
// @Tags(identifierAttribute="arn")
func resourceConnectPeer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectPeerCreate,
		ReadWithoutTimeout:   resourceConnectPeerRead,
		UpdateWithoutTimeout: resourceConnectPeerUpdate,
		DeleteWithoutTimeout: resourceConnectPeerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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

func resourceConnectPeerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	connectAttachmentID := d.Get("connect_attachment_id").(string)
	peerAddress := d.Get("peer_address").(string)
	input := &networkmanager.CreateConnectPeerInput{
		ConnectAttachmentId: aws.String(connectAttachmentID),
		PeerAddress:         aws.String(peerAddress),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk("bgp_options"); ok && len(v.([]any)) > 0 {
		input.BgpOptions = expandPeerOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("core_network_address"); ok {
		input.CoreNetworkAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("inside_cidr_blocks"); ok {
		input.InsideCidrBlocks = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk("subnet_arn"); ok {
		input.SubnetArn = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutCreate),
		func() (any, error) {
			return conn.CreateConnectPeer(ctx, input)
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
			if validationExceptionMessageContains(err, awstypes.ValidationExceptionReasonOther, "Connect attachment state is invalid") {
				return true, err
			}

			return false, err
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Peer: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*networkmanager.CreateConnectPeerOutput).ConnectPeer.ConnectPeerId))

	if _, err := waitConnectPeerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connect Peer (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConnectPeerRead(ctx, d, meta)...)
}

func resourceConnectPeerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	connectPeer, err := findConnectPeerByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Connect Peer %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Connect Peer (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, connectPeerARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set("bgp_options", []any{map[string]any{
		"peer_asn": connectPeer.Configuration.BgpConfigurations[0].PeerAsn,
	}})
	d.Set(names.AttrConfiguration, []any{flattenPeerConfiguration(connectPeer.Configuration)})
	d.Set("connect_peer_id", connectPeer.ConnectPeerId)
	d.Set("core_network_id", connectPeer.CoreNetworkId)
	if connectPeer.CreatedAt != nil {
		d.Set(names.AttrCreatedAt, aws.ToTime(connectPeer.CreatedAt).Format(time.RFC3339))
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

func resourceConnectPeerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceConnectPeerRead(ctx, d, meta)
}

func resourceConnectPeerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	log.Printf("[DEBUG] Deleting Network Manager Connect Peer: %s", d.Id())
	_, err := conn.DeleteConnectPeer(ctx, &networkmanager.DeleteConnectPeerInput{
		ConnectPeerId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func expandPeerOptions(o map[string]any) *awstypes.BgpOptions {
	if o == nil {
		return nil
	}

	object := &awstypes.BgpOptions{}

	if v, ok := o["peer_asn"].(int); ok {
		object.PeerAsn = aws.Int64(int64(v))
	}

	return object
}

func findConnectPeerByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.ConnectPeer, error) {
	input := &networkmanager.GetConnectPeerInput{
		ConnectPeerId: aws.String(id),
	}

	output, err := conn.GetConnectPeer(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func flattenPeerConfiguration(apiObject *awstypes.ConnectPeerConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	confMap := map[string]any{}

	for _, v := range apiObject.BgpConfigurations {
		bgpConfMap := map[string]any{}

		if a := v.CoreNetworkAddress; a != nil {
			bgpConfMap["core_network_address"] = aws.ToString(a)
		}
		if a := v.CoreNetworkAsn; a != nil {
			bgpConfMap["core_network_asn"] = aws.ToInt64(a)
		}
		if a := v.PeerAddress; a != nil {
			bgpConfMap["peer_address"] = aws.ToString(a)
		}
		if a := v.PeerAsn; a != nil {
			bgpConfMap["peer_asn"] = aws.ToInt64(a)
		}
		var existing []any
		if c, ok := confMap["bgp_configurations"]; ok {
			existing = c.([]any)
		}
		confMap["bgp_configurations"] = append(existing, bgpConfMap)
	}
	if v := apiObject.CoreNetworkAddress; v != nil {
		confMap["core_network_address"] = aws.ToString(v)
	}
	if v := apiObject.InsideCidrBlocks; v != nil {
		confMap["inside_cidr_blocks"] = v
	}
	if v := apiObject.PeerAddress; v != nil {
		confMap["peer_address"] = aws.ToString(v)
	}

	confMap[names.AttrProtocol] = apiObject.Protocol

	return confMap
}

func statusConnectPeerState(ctx context.Context, conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findConnectPeerByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitConnectPeerCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.ConnectPeer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectPeerStateCreating),
		Target:  enum.Slice(awstypes.ConnectPeerStateAvailable),
		Timeout: timeout,
		Refresh: statusConnectPeerState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ConnectPeer); ok {
		tfresource.SetLastError(err, connectPeersError(output.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitConnectPeerDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.ConnectPeer, error) {
	stateconf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.ConnectPeerStateDeleting),
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusConnectPeerState(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateconf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ConnectPeer); ok {
		tfresource.SetLastError(err, connectPeersError(output.LastModificationErrors))

		return output, err
	}

	return nil, err
}

// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsnetworkmanager.html#awsnetworkmanager-resources-for-iam-policies.
func connectPeerARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.GlobalARN(ctx, "networkmanager", "connect-peer/"+id)
}
