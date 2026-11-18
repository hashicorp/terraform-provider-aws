// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_peering_connection", name="VPC Peering Connection")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCPeeringConnectionCreate,
		ReadWithoutTimeout:   resourceVPCPeeringConnectionRead,
		UpdateWithoutTimeout: resourceVPCPeeringConnectionUpdate,
		DeleteWithoutTimeout: resourceVPCPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		// Keep in sync with aws_vpc_peering_connection_accepter's schema.
		// See notes in vpc_peering_connection_accepter.go.
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"accept_status": {
					Type:     schema.TypeString,
					Computed: true,
				},
				// Multipath resilience is cross-Region only, and each side can only be modified from its own Region.
				"accepter": vpcPeeringConnectionOptionsSchema(false),
				"auto_accept": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"multipath_resilience_status": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"peer_owner_id": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Computed: true,
				},
				"peer_region": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Computed: true,
				},
				"peer_vpc_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"requester":       vpcPeeringConnectionOptionsSchema(true),
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				names.AttrVPCID: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
			}
		},
	}
}

// vpcPeeringConnectionOptionsSchema returns the schema of an options block.
// multipath adds the multipath resilience arguments, which can only be set for the side of
// the connection in the Region (and account) that the resource is managed in.
func vpcPeeringConnectionOptionsSchema(multipath bool) *schema.Schema {
	s := map[string]*schema.Schema{
		"allow_remote_vpc_dns_resolution": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
	}

	if multipath {
		s["allow_multipath_resilience"] = &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		}
		s["multipath_resilience_mode"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: enum.Validate[awstypes.MultipathResilienceMode](),
		}
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: s,
		},
	}
}

func resourceVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVpcPeeringConnectionInput{
		PeerVpcId:         aws.String(d.Get("peer_vpc_id").(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpcPeeringConnection),
		VpcId:             aws.String(d.Get(names.AttrVPCID).(string)),
	}

	if v, ok := d.GetOk("peer_owner_id"); ok {
		input.PeerOwnerId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("peer_region"); ok {
		if _, ok := d.GetOk("auto_accept"); ok {
			return sdkdiag.AppendErrorf(diags, "`peer_region` cannot be set whilst `auto_accept` is `true` when creating an EC2 VPC Peering Connection")
		}

		input.PeerRegion = aws.String(v.(string))
	}

	// The requester's multipath resilience is set at create time, as a connection
	// can't be modified until it's accepted, and cross-Region connections can't be auto-accepted.
	input.RequesterPeeringConnectionOptions = expandCreatePeeringConnectionOptionsRequest(d)

	log.Printf("[DEBUG] Creating EC2 VPC Peering Connection: %#v", input)
	output, err := conn.CreateVpcPeeringConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC Peering Connection: %s", err)
	}

	d.SetId(aws.ToString(output.VpcPeeringConnection.VpcPeeringConnectionId))

	vpcPeeringConnection, err := waitVPCPeeringConnectionActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Peering Connection (%s) create: %s", d.Id(), err)
	}

	if _, ok := d.GetOk("auto_accept"); ok && vpcPeeringConnection.Status.Code == awstypes.VpcPeeringConnectionStateReasonCodePendingAcceptance {
		vpcPeeringConnection, err = acceptVPCPeeringConnection(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	accepterPeeringConnectionOptions := expandPeeringConnectionOptionsRequest(d, "accepter")
	requesterPeeringConnectionOptions := expandPeeringConnectionOptionsRequest(d, "requester")
	if v := requesterPeeringConnectionOptions; v != nil {
		v.AllowMultipathResilience = nil
		v.MultipathResilienceMode = ""

		// A new connection doesn't allow DNS resolution from the remote VPC, so there's nothing left to modify.
		if !aws.ToBool(v.AllowDnsResolutionFromRemoteVpc) {
			requesterPeeringConnectionOptions = nil
		}
	}

	if err := updateVPCPeeringConnectionOptions(ctx, conn, d.Id(), vpcPeeringConnection, accepterPeeringConnectionOptions, requesterPeeringConnectionOptions, true); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceVPCPeeringConnectionRead(ctx, d, meta)...)
}

func resourceVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return readVPCPeeringConnection(ctx, d, meta, "requester")
}

// readVPCPeeringConnection reads a VPC peering connection into the state of aws_vpc_peering_connection
// or aws_vpc_peering_connection_accepter. multipathKey is the options block ("accepter" or "requester")
// that has the multipath resilience arguments.
func readVPCPeeringConnection(ctx context.Context, d *schema.ResourceData, meta any, multipathKey string) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcPeeringConnection, err := findVPCPeeringConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 VPC Peering Connection %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Peering Connection (%s): %s", d.Id(), err)
	}

	d.Set("accept_status", vpcPeeringConnection.Status.Code)
	d.Set("multipath_resilience_status", vpcPeeringConnection.MultipathResilienceStatus)
	d.Set("peer_region", vpcPeeringConnection.AccepterVpcInfo.Region)

	if accountID := meta.(*conns.AWSClient).AccountID(ctx); accountID == aws.ToString(vpcPeeringConnection.AccepterVpcInfo.OwnerId) && accountID != aws.ToString(vpcPeeringConnection.RequesterVpcInfo.OwnerId) {
		// We're the accepter.
		d.Set("peer_owner_id", vpcPeeringConnection.RequesterVpcInfo.OwnerId)
		d.Set("peer_vpc_id", vpcPeeringConnection.RequesterVpcInfo.VpcId)
		d.Set(names.AttrVPCID, vpcPeeringConnection.AccepterVpcInfo.VpcId)
	} else {
		// We're the requester.
		d.Set("peer_owner_id", vpcPeeringConnection.AccepterVpcInfo.OwnerId)
		d.Set("peer_vpc_id", vpcPeeringConnection.AccepterVpcInfo.VpcId)
		d.Set(names.AttrVPCID, vpcPeeringConnection.RequesterVpcInfo.VpcId)
	}

	if vpcPeeringConnection.AccepterVpcInfo.PeeringOptions != nil {
		if err := d.Set("accepter", []any{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.AccepterVpcInfo.PeeringOptions, multipathKey == "accepter")}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting accepter: %s", err)
		}
	} else {
		d.Set("accepter", nil)
	}

	if vpcPeeringConnection.RequesterVpcInfo.PeeringOptions != nil {
		if err := d.Set("requester", []any{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.RequesterVpcInfo.PeeringOptions, multipathKey == "requester")}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting requester: %s", err)
		}
	} else {
		d.Set("requester", nil)
	}

	setTagsOut(ctx, vpcPeeringConnection.Tags)

	return diags
}

func resourceVPCPeeringConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return updateVPCPeeringConnection(ctx, d, meta, "requester")
}

// updateVPCPeeringConnection updates aws_vpc_peering_connection or aws_vpc_peering_connection_accepter.
// See readVPCPeeringConnection for multipathKey.
func updateVPCPeeringConnection(ctx context.Context, d *schema.ResourceData, meta any, multipathKey string) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcPeeringConnection, err := findVPCPeeringConnectionByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Peering Connection (%s): %s", d.Id(), err)
	}

	if _, ok := d.GetOk("auto_accept"); ok && vpcPeeringConnection.Status.Code == awstypes.VpcPeeringConnectionStateReasonCodePendingAcceptance {
		vpcPeeringConnection, err = acceptVPCPeeringConnection(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChanges("accepter", "requester") {
		if err := modifyVPCPeeringConnectionOptions(ctx, conn, d, vpcPeeringConnection, true); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, readVPCPeeringConnection(ctx, d, meta, multipathKey)...)
}

func resourceVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 VPC Peering Connection: %s", d.Id())
	input := ec2.DeleteVpcPeeringConnectionInput{
		VpcPeeringConnectionId: aws.String(d.Id()),
	}
	_, err := conn.DeleteVpcPeeringConnection(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCPeeringConnectionIDNotFound) {
		return diags
	}

	// "InvalidStateTransition: Invalid state transition for pcx-0000000000000000, attempted to transition from failed to deleting"
	if tfawserr.ErrMessageContains(err, "InvalidStateTransition", "to deleting") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPC Peering Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCPeeringConnectionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC Peering Connection (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func acceptVPCPeeringConnection(ctx context.Context, conn *ec2.Client, vpcPeeringConnectionID string, timeout time.Duration) (*awstypes.VpcPeeringConnection, error) {
	log.Printf("[INFO] Accepting EC2 VPC Peering Connection: %s", vpcPeeringConnectionID)
	input := ec2.AcceptVpcPeeringConnectionInput{
		VpcPeeringConnectionId: aws.String(vpcPeeringConnectionID),
	}
	_, err := conn.AcceptVpcPeeringConnection(ctx, &input)

	if err != nil {
		return nil, fmt.Errorf("accepting EC2 VPC Peering Connection (%s): %w", vpcPeeringConnectionID, err)
	}

	// "OperationNotPermitted: Peering pcx-0000000000000000 is not active. Peering options can be added only to active peerings."
	vpcPeeringConnection, err := waitVPCPeeringConnectionActive(ctx, conn, vpcPeeringConnectionID, timeout)

	if err != nil {
		return nil, fmt.Errorf("accepting EC2 VPC Peering Connection (%s): waiting for completion: %w", vpcPeeringConnectionID, err)
	}

	return vpcPeeringConnection, nil
}

func modifyVPCPeeringConnectionOptions(ctx context.Context, conn *ec2.Client, d *schema.ResourceData, vpcPeeringConnection *awstypes.VpcPeeringConnection, checkActive bool) error {
	var accepterPeeringConnectionOptions, requesterPeeringConnectionOptions *awstypes.PeeringConnectionOptionsRequest

	if key := "accepter"; d.HasChange(key) {
		accepterPeeringConnectionOptions = expandPeeringConnectionOptionsRequest(d, key)
	}

	if key := "requester"; d.HasChange(key) {
		requesterPeeringConnectionOptions = expandPeeringConnectionOptionsRequest(d, key)
	}

	return updateVPCPeeringConnectionOptions(ctx, conn, d.Id(), vpcPeeringConnection, accepterPeeringConnectionOptions, requesterPeeringConnectionOptions, checkActive)
}

func updateVPCPeeringConnectionOptions(ctx context.Context, conn *ec2.Client, id string, vpcPeeringConnection *awstypes.VpcPeeringConnection, accepterPeeringConnectionOptions, requesterPeeringConnectionOptions *awstypes.PeeringConnectionOptionsRequest, checkActive bool) error {
	if accepterPeeringConnectionOptions == nil && requesterPeeringConnectionOptions == nil {
		return nil
	}

	if checkActive {
		switch statusCode := vpcPeeringConnection.Status.Code; statusCode {
		case awstypes.VpcPeeringConnectionStateReasonCodeActive, awstypes.VpcPeeringConnectionStateReasonCodeProvisioning:
		default:
			return fmt.Errorf(
				"Unable to modify EC2 VPC Peering Connection Options. EC2 VPC Peering Connection (%s) is not active (current status: %s). "+
					"Please set the `auto_accept` attribute to `true` or activate the EC2 VPC Peering Connection manually.",
				id, statusCode)
		}
	}

	input := &ec2.ModifyVpcPeeringConnectionOptionsInput{
		AccepterPeeringConnectionOptions:  accepterPeeringConnectionOptions,
		RequesterPeeringConnectionOptions: requesterPeeringConnectionOptions,
		VpcPeeringConnectionId:            aws.String(id),
	}

	log.Printf("[DEBUG] Modifying VPC Peering Connection Options: %#v", input)
	if _, err := conn.ModifyVpcPeeringConnectionOptions(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 VPC Peering Connection (%s) Options: %w", id, err)
	}

	// Retry reading back the modified options to deal with eventual consistency.
	// Often this is to do with a delay transitioning from pending-acceptance to active.
	err := tfresource.Retry(ctx, ec2PropagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		vpcPeeringConnection, err := findVPCPeeringConnectionByID(ctx, conn, id)

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		if v := vpcPeeringConnection.AccepterVpcInfo; v != nil && v.PeeringOptions != nil && accepterPeeringConnectionOptions != nil {
			if !vpcPeeringConnectionOptionsEqual(v.PeeringOptions, accepterPeeringConnectionOptions) {
				return tfresource.RetryableError(errors.New("Accepter Options not stable"))
			}
		}

		if v := vpcPeeringConnection.RequesterVpcInfo; v != nil && v.PeeringOptions != nil && requesterPeeringConnectionOptions != nil {
			if !vpcPeeringConnectionOptionsEqual(v.PeeringOptions, requesterPeeringConnectionOptions) {
				return tfresource.RetryableError(errors.New("Requester Options not stable"))
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("modifying EC2 VPC Peering Connection (%s) Options: waiting for completion: %w", id, err)
	}

	return nil
}

func vpcPeeringConnectionOptionsEqual(o1 *awstypes.VpcPeeringConnectionOptionsDescription, o2 *awstypes.PeeringConnectionOptionsRequest) bool {
	if aws.ToBool(o1.AllowDnsResolutionFromRemoteVpc) != aws.ToBool(o2.AllowDnsResolutionFromRemoteVpc) {
		return false
	}

	if v := o2.AllowMultipathResilience; v != nil && aws.ToBool(o1.AllowMultipathResilience) != aws.ToBool(v) {
		return false
	}

	if v := o2.MultipathResilienceMode; v != "" && o1.MultipathResilienceMode != v {
		return false
	}

	return true
}

// vpcPeeringConnectionOptionConfigured reports whether attribute attr is set in
// the configuration of options block key ("accepter" or "requester").
func vpcPeeringConnectionOptionConfigured(d *schema.ResourceData, key, attr string) bool {
	v := d.GetRawConfig().GetAttr(key)
	if !v.IsKnown() || v.IsNull() || v.LengthInt() == 0 {
		return false
	}

	v = v.Index(cty.NumberIntVal(0))
	if !v.IsKnown() || v.IsNull() {
		return false
	}

	// Not every options block has every attribute.
	if !v.Type().HasAttribute(attr) {
		return false
	}

	v = v.GetAttr(attr)

	return v.IsKnown() && !v.IsNull()
}

func expandPeeringConnectionOptionsRequest(d *schema.ResourceData, key string) *awstypes.PeeringConnectionOptionsRequest {
	v, ok := d.GetOk(key)
	if !ok || len(v.([]any)) == 0 || v.([]any)[0] == nil {
		return nil
	}

	tfMap := v.([]any)[0].(map[string]any)
	apiObject := &awstypes.PeeringConnectionOptionsRequest{}

	if v, ok := tfMap["allow_remote_vpc_dns_resolution"].(bool); ok {
		apiObject.AllowDnsResolutionFromRemoteVpc = aws.Bool(v)
	}

	// Multipath resilience is Optional+Computed: when it isn't configured, leave the current setting unchanged.
	// Accepting a connection whose requester allows it also allows it for the accepter.
	if vpcPeeringConnectionOptionConfigured(d, key, "allow_multipath_resilience") {
		allow := tfMap["allow_multipath_resilience"].(bool)
		apiObject.AllowMultipathResilience = aws.Bool(allow)

		if v := tfMap["multipath_resilience_mode"].(string); allow && v != "" && vpcPeeringConnectionOptionConfigured(d, key, "multipath_resilience_mode") {
			apiObject.MultipathResilienceMode = awstypes.MultipathResilienceMode(v)
		}
	}

	return apiObject
}

func expandCreatePeeringConnectionOptionsRequest(d *schema.ResourceData) *awstypes.CreatePeeringConnectionOptionsRequest {
	const key = "requester"

	if !vpcPeeringConnectionOptionConfigured(d, key, "allow_multipath_resilience") || !d.Get(key+".0.allow_multipath_resilience").(bool) {
		return nil
	}

	apiObject := &awstypes.CreatePeeringConnectionOptionsRequest{
		AllowMultipathResilience: aws.Bool(true),
	}

	if vpcPeeringConnectionOptionConfigured(d, key, "multipath_resilience_mode") {
		apiObject.MultipathResilienceMode = awstypes.MultipathResilienceMode(d.Get(key + ".0.multipath_resilience_mode").(string))
	}

	return apiObject
}

// flattenVPCPeeringConnectionOptionsDescription flattens an options block.
// multipath includes the multipath resilience attributes; see vpcPeeringConnectionOptionsSchema.
func flattenVPCPeeringConnectionOptionsDescription(apiObject *awstypes.VpcPeeringConnectionOptionsDescription, multipath bool) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AllowDnsResolutionFromRemoteVpc; v != nil {
		tfMap["allow_remote_vpc_dns_resolution"] = aws.ToBool(v)
	}

	if !multipath {
		return tfMap
	}

	if v := apiObject.AllowMultipathResilience; v != nil {
		tfMap["allow_multipath_resilience"] = aws.ToBool(v)
	}

	if v := apiObject.MultipathResilienceMode; v != "" {
		tfMap["multipath_resilience_mode"] = string(v)
	}

	return tfMap
}
