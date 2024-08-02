// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
		Schema: map[string]*schema.Schema{
			"accept_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"accepter": vpcPeeringConnectionOptionsSchema,
			"auto_accept": {
				Type:     schema.TypeBool,
				Optional: true,
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
			"requester":       vpcPeeringConnectionOptionsSchema,
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

var vpcPeeringConnectionOptionsSchema = &schema.Schema{
	Type:     schema.TypeList,
	Optional: true,
	Computed: true,
	MaxItems: 1,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"allow_remote_vpc_dns_resolution": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	},
}

func resourceVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if err := modifyVPCPeeringConnectionOptions(ctx, conn, d, vpcPeeringConnection, true); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceVPCPeeringConnectionRead(ctx, d, meta)...)
}

func resourceVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcPeeringConnection, err := findVPCPeeringConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC Peering Connection %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Peering Connection (%s): %s", d.Id(), err)
	}

	d.Set("accept_status", vpcPeeringConnection.Status.Code)
	d.Set("peer_region", vpcPeeringConnection.AccepterVpcInfo.Region)

	if accountID := meta.(*conns.AWSClient).AccountID; accountID == aws.ToString(vpcPeeringConnection.AccepterVpcInfo.OwnerId) && accountID != aws.ToString(vpcPeeringConnection.RequesterVpcInfo.OwnerId) {
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
		if err := d.Set("accepter", []interface{}{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.AccepterVpcInfo.PeeringOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting accepter: %s", err)
		}
	} else {
		d.Set("accepter", nil)
	}

	if vpcPeeringConnection.RequesterVpcInfo.PeeringOptions != nil {
		if err := d.Set("requester", []interface{}{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.RequesterVpcInfo.PeeringOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting requester: %s", err)
		}
	} else {
		d.Set("requester", nil)
	}

	setTagsOut(ctx, vpcPeeringConnection.Tags)

	return diags
}

func resourceVPCPeeringConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return append(diags, resourceVPCPeeringConnectionRead(ctx, d, meta)...)
}

func resourceVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 VPC Peering Connection: %s", d.Id())
	_, err := conn.DeleteVpcPeeringConnection(ctx, &ec2.DeleteVpcPeeringConnectionInput{
		VpcPeeringConnectionId: aws.String(d.Id()),
	})

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
	_, err := conn.AcceptVpcPeeringConnection(ctx, &ec2.AcceptVpcPeeringConnectionInput{
		VpcPeeringConnectionId: aws.String(vpcPeeringConnectionID),
	})

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
		if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			accepterPeeringConnectionOptions = expandPeeringConnectionOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	if key := "requester"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			requesterPeeringConnectionOptions = expandPeeringConnectionOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
		}
	}

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
				d.Id(), statusCode)
		}
	}

	input := &ec2.ModifyVpcPeeringConnectionOptionsInput{
		AccepterPeeringConnectionOptions:  accepterPeeringConnectionOptions,
		RequesterPeeringConnectionOptions: requesterPeeringConnectionOptions,
		VpcPeeringConnectionId:            aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Modifying VPC Peering Connection Options: %#v", input)
	if _, err := conn.ModifyVpcPeeringConnectionOptions(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 VPC Peering Connection (%s) Options: %w", d.Id(), err)
	}

	// Retry reading back the modified options to deal with eventual consistency.
	// Often this is to do with a delay transitioning from pending-acceptance to active.
	err := retry.RetryContext(ctx, ec2PropagationTimeout, func() *retry.RetryError { // nosemgrep:ci.helper-schema-retry-RetryContext-without-TimeoutError-check
		vpcPeeringConnection, err := findVPCPeeringConnectionByID(ctx, conn, d.Id())

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if v := vpcPeeringConnection.AccepterVpcInfo; v != nil && v.PeeringOptions != nil && accepterPeeringConnectionOptions != nil {
			if !vpcPeeringConnectionOptionsEqual(v.PeeringOptions, accepterPeeringConnectionOptions) {
				return retry.RetryableError(errors.New("Accepter Options not stable"))
			}
		}

		if v := vpcPeeringConnection.RequesterVpcInfo; v != nil && v.PeeringOptions != nil && requesterPeeringConnectionOptions != nil {
			if !vpcPeeringConnectionOptionsEqual(v.PeeringOptions, requesterPeeringConnectionOptions) {
				return retry.RetryableError(errors.New("Requester Options not stable"))
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("modifying EC2 VPC Peering Connection (%s) Options: waiting for completion: %w", d.Id(), err)
	}

	return nil
}

func vpcPeeringConnectionOptionsEqual(o1 *awstypes.VpcPeeringConnectionOptionsDescription, o2 *awstypes.PeeringConnectionOptionsRequest) bool {
	return aws.ToBool(o1.AllowDnsResolutionFromRemoteVpc) == aws.ToBool(o2.AllowDnsResolutionFromRemoteVpc)
}

func expandPeeringConnectionOptionsRequest(tfMap map[string]interface{}) *awstypes.PeeringConnectionOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PeeringConnectionOptionsRequest{}

	if v, ok := tfMap["allow_remote_vpc_dns_resolution"].(bool); ok {
		apiObject.AllowDnsResolutionFromRemoteVpc = aws.Bool(v)
	}

	return apiObject
}

func flattenVPCPeeringConnectionOptionsDescription(apiObject *awstypes.VpcPeeringConnectionOptionsDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllowDnsResolutionFromRemoteVpc; v != nil {
		tfMap["allow_remote_vpc_dns_resolution"] = aws.ToBool(v)
	}

	return tfMap
}
