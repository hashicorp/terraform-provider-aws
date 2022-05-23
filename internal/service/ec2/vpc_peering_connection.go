package ec2

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCPeeringConnectionCreate,
		Read:   resourceVPCPeeringConnectionRead,
		Update: resourceVPCPeeringConnectionUpdate,
		Delete: resourceVPCPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"requester": vpcPeeringConnectionOptionsSchema,
			"tags":      tftags.TagsSchema(),
			"tags_all":  tftags.TagsSchemaComputed(),
			"vpc_id": {
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
			"allow_classic_link_to_remote_vpc": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allow_remote_vpc_dns_resolution": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allow_vpc_to_remote_classic_link": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	},
}

func resourceVPCPeeringConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateVpcPeeringConnectionInput{
		PeerVpcId:         aws.String(d.Get("peer_vpc_id").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpcPeeringConnection),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("peer_owner_id"); ok {
		input.PeerOwnerId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("peer_region"); ok {
		if _, ok := d.GetOk("auto_accept"); ok {
			return fmt.Errorf("`peer_region` cannot be set whilst `auto_accept` is `true` when creating an EC2 VPC Peering Connection")
		}

		input.PeerRegion = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 VPC Peering Connection: %s", input)
	output, err := conn.CreateVpcPeeringConnection(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 VPC Peering Connection: %w", err)
	}

	d.SetId(aws.StringValue(output.VpcPeeringConnection.VpcPeeringConnectionId))

	vpcPeeringConnection, err := WaitVPCPeeringConnectionActive(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPC Peering Connection (%s) create: %w", d.Id(), err)
	}

	if _, ok := d.GetOk("auto_accept"); ok && aws.StringValue(vpcPeeringConnection.Status.Code) == ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance {
		vpcPeeringConnection, err = acceptVPCPeeringConnection(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return err
		}
	}

	if err := modifyVPCPeeringConnectionOptions(conn, d, vpcPeeringConnection, true); err != nil {
		return err
	}

	return resourceVPCPeeringConnectionRead(d, meta)
}

func resourceVPCPeeringConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vpcPeeringConnection, err := FindVPCPeeringConnectionByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC Peering Connection %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC Peering Connection (%s): %w", d.Id(), err)
	}

	d.Set("accept_status", vpcPeeringConnection.Status.Code)
	d.Set("peer_region", vpcPeeringConnection.AccepterVpcInfo.Region)

	if accountID := meta.(*conns.AWSClient).AccountID; accountID == aws.StringValue(vpcPeeringConnection.AccepterVpcInfo.OwnerId) && accountID != aws.StringValue(vpcPeeringConnection.RequesterVpcInfo.OwnerId) {
		// We're the accepter.
		d.Set("peer_owner_id", vpcPeeringConnection.RequesterVpcInfo.OwnerId)
		d.Set("peer_vpc_id", vpcPeeringConnection.RequesterVpcInfo.VpcId)
		d.Set("vpc_id", vpcPeeringConnection.AccepterVpcInfo.VpcId)
	} else {
		// We're the requester.
		d.Set("peer_owner_id", vpcPeeringConnection.AccepterVpcInfo.OwnerId)
		d.Set("peer_vpc_id", vpcPeeringConnection.AccepterVpcInfo.VpcId)
		d.Set("vpc_id", vpcPeeringConnection.RequesterVpcInfo.VpcId)
	}

	if vpcPeeringConnection.AccepterVpcInfo.PeeringOptions != nil {
		if err := d.Set("accepter", []interface{}{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.AccepterVpcInfo.PeeringOptions)}); err != nil {
			return fmt.Errorf("error setting accepter: %w", err)
		}
	} else {
		d.Set("accepter", nil)
	}

	if vpcPeeringConnection.RequesterVpcInfo.PeeringOptions != nil {
		if err := d.Set("requester", []interface{}{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.RequesterVpcInfo.PeeringOptions)}); err != nil {
			return fmt.Errorf("error setting requester: %w", err)
		}
	} else {
		d.Set("requester", nil)
	}

	tags := KeyValueTags(vpcPeeringConnection.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceVPCPeeringConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcPeeringConnection, err := FindVPCPeeringConnectionByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC Peering Connection (%s): %w", d.Id(), err)
	}

	if _, ok := d.GetOk("auto_accept"); ok && aws.StringValue(vpcPeeringConnection.Status.Code) == ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance {
		vpcPeeringConnection, err = acceptVPCPeeringConnection(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return err
		}
	}

	if d.HasChanges("accepter", "requester") {
		if err := modifyVPCPeeringConnectionOptions(conn, d, vpcPeeringConnection, true); err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 VPC Peering Connection (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceVPCPeeringConnectionRead(d, meta)
}

func resourceVPCPeeringConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 VPC Peering Connection: %s", d.Id())
	_, err := conn.DeleteVpcPeeringConnection(&ec2.DeleteVpcPeeringConnectionInput{
		VpcPeeringConnectionId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCPeeringConnectionIDNotFound) {
		return nil
	}

	// "InvalidStateTransition: Invalid state transition for pcx-0000000000000000, attempted to transition from failed to deleting"
	if tfawserr.ErrMessageContains(err, "InvalidStateTransition", "to deleting") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 VPC Peering Connection (%s): %w", d.Id(), err)
	}

	if _, err := WaitVPCPeeringConnectionDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for EC2 VPC Peering Connection (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func acceptVPCPeeringConnection(conn *ec2.EC2, vpcPeeringConnectionID string, timeout time.Duration) (*ec2.VpcPeeringConnection, error) {
	log.Printf("[INFO] Accepting EC2 VPC Peering Connection: %s", vpcPeeringConnectionID)
	_, err := conn.AcceptVpcPeeringConnection(&ec2.AcceptVpcPeeringConnectionInput{
		VpcPeeringConnectionId: aws.String(vpcPeeringConnectionID),
	})

	if err != nil {
		return nil, fmt.Errorf("error acccepting EC2 VPC Peering Connection (%s): %w", vpcPeeringConnectionID, err)
	}

	// "OperationNotPermitted: Peering pcx-0000000000000000 is not active. Peering options can be added only to active peerings."
	vpcPeeringConnection, err := WaitVPCPeeringConnectionActive(conn, vpcPeeringConnectionID, timeout)

	if err != nil {
		return nil, fmt.Errorf("error waiting for EC2 VPC Peering Connection (%s) update: %w", vpcPeeringConnectionID, err)
	}

	return vpcPeeringConnection, nil
}

func modifyVPCPeeringConnectionOptions(conn *ec2.EC2, d *schema.ResourceData, vpcPeeringConnection *ec2.VpcPeeringConnection, checkActive bool) error {
	var accepterPeeringConnectionOptions, requesterPeeringConnectionOptions *ec2.PeeringConnectionOptionsRequest
	crossRegionPeering := aws.StringValue(vpcPeeringConnection.RequesterVpcInfo.Region) != aws.StringValue(vpcPeeringConnection.AccepterVpcInfo.Region)

	if key := "accepter"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			accepterPeeringConnectionOptions = expandPeeringConnectionOptionsRequest(v.([]interface{})[0].(map[string]interface{}), crossRegionPeering)
		}
	}

	if key := "requester"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			requesterPeeringConnectionOptions = expandPeeringConnectionOptionsRequest(v.([]interface{})[0].(map[string]interface{}), crossRegionPeering)
		}
	}

	if accepterPeeringConnectionOptions == nil && requesterPeeringConnectionOptions == nil {
		return nil
	}

	if checkActive {
		switch statusCode := aws.StringValue(vpcPeeringConnection.Status.Code); statusCode {
		case ec2.VpcPeeringConnectionStateReasonCodeActive, ec2.VpcPeeringConnectionStateReasonCodeProvisioning:
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

	log.Printf("[DEBUG] Modifying VPC Peering Connection Options: %s", input)
	if _, err := conn.ModifyVpcPeeringConnectionOptions(input); err != nil {
		return fmt.Errorf("error modifying EC2 VPC Peering Connection (%s) Options: %w", d.Id(), err)
	}

	// Retry reading back the modified options to deal with eventual consistency.
	// Often this is to do with a delay transitioning from pending-acceptance to active.
	err := resource.Retry(VPCPeeringConnectionOptionsPropagationTimeout, func() *resource.RetryError { // nosem: helper-schema-resource-Retry-without-TimeoutError-check
		vpcPeeringConnection, err := FindVPCPeeringConnectionByID(conn, d.Id())

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if v := vpcPeeringConnection.AccepterVpcInfo; v != nil && accepterPeeringConnectionOptions != nil {
			if !vpcPeeringConnectionOptionsEqual(v.PeeringOptions, accepterPeeringConnectionOptions) {
				return resource.RetryableError(errors.New("Accepter Options not stable"))
			}
		}

		if v := vpcPeeringConnection.RequesterVpcInfo; v != nil && requesterPeeringConnectionOptions != nil {
			if !vpcPeeringConnectionOptionsEqual(v.PeeringOptions, requesterPeeringConnectionOptions) {
				return resource.RetryableError(errors.New("Requester Options not stable"))
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPC Peering Connection (%s) Options update: %w", d.Id(), err)
	}

	return nil
}

func vpcPeeringConnectionOptionsEqual(o1 *ec2.VpcPeeringConnectionOptionsDescription, o2 *ec2.PeeringConnectionOptionsRequest) bool {
	return aws.BoolValue(o1.AllowDnsResolutionFromRemoteVpc) == aws.BoolValue(o2.AllowDnsResolutionFromRemoteVpc) &&
		aws.BoolValue(o1.AllowEgressFromLocalClassicLinkToRemoteVpc) == aws.BoolValue(o2.AllowEgressFromLocalClassicLinkToRemoteVpc) &&
		aws.BoolValue(o1.AllowEgressFromLocalVpcToRemoteClassicLink) == aws.BoolValue(o2.AllowEgressFromLocalVpcToRemoteClassicLink)
}

func expandPeeringConnectionOptionsRequest(tfMap map[string]interface{}, crossRegionPeering bool) *ec2.PeeringConnectionOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.PeeringConnectionOptionsRequest{}

	if v, ok := tfMap["allow_remote_vpc_dns_resolution"].(bool); ok {
		apiObject.AllowDnsResolutionFromRemoteVpc = aws.Bool(v)
	}

	if !crossRegionPeering {
		if v, ok := tfMap["allow_classic_link_to_remote_vpc"].(bool); ok {
			apiObject.AllowEgressFromLocalClassicLinkToRemoteVpc = aws.Bool(v)
		}

		if v, ok := tfMap["allow_vpc_to_remote_classic_link"].(bool); ok {
			apiObject.AllowEgressFromLocalVpcToRemoteClassicLink = aws.Bool(v)
		}
	}

	return apiObject
}

func flattenVPCPeeringConnectionOptionsDescription(apiObject *ec2.VpcPeeringConnectionOptionsDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllowDnsResolutionFromRemoteVpc; v != nil {
		tfMap["allow_remote_vpc_dns_resolution"] = aws.BoolValue(v)
	}

	if v := apiObject.AllowEgressFromLocalClassicLinkToRemoteVpc; v != nil {
		tfMap["allow_classic_link_to_remote_vpc"] = aws.BoolValue(v)
	}

	if v := apiObject.AllowEgressFromLocalVpcToRemoteClassicLink; v != nil {
		tfMap["allow_vpc_to_remote_classic_link"] = aws.BoolValue(v)
	}

	return tfMap
}
