package ec2

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCPeeringConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCPeeringAccepterCreate,
		Read:   resourceVPCPeeringConnectionRead,
		Update: resourceVPCPeeringConnectionUpdate,
		Delete: resourceVPCPeeringAccepterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, m interface{}) (result []*schema.ResourceData, err error) {
				d.Set("vpc_peering_connection_id", d.Id())

				return []*schema.ResourceData{d}, nil
			},
		},

		// Keep in sync with aws_vpc_peering_connections's schema with the following changes:
		//   - peer_owner_id is Computed-only
		//   - peer_region is Computed-only
		//   - peer_vpc_id is Computed-only
		//   - vpc_id is Computed-only
		// and additions:
		//   - vpc_peering_connection_id Required/ForceNew
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
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"requester": vpcPeeringConnectionOptionsSchema,
			"tags":      tftags.TagsSchema(),
			"tags_all":  tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_peering_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCPeeringAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if peeringConnectionOptionsAllowsClassicLink(d) {
		return errors.New(`with the retirement of EC2-Classic no VPC Peering Connections can be accepted with ClassicLink options enabled`)
	}

	vpcPeeringConnectionID := d.Get("vpc_peering_connection_id").(string)
	vpcPeeringConnection, err := FindVPCPeeringConnectionByID(conn, vpcPeeringConnectionID)

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC Peering Connection (%s): %w", vpcPeeringConnectionID, err)
	}

	d.SetId(vpcPeeringConnectionID)

	if _, ok := d.GetOk("auto_accept"); ok && aws.StringValue(vpcPeeringConnection.Status.Code) == ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance {
		vpcPeeringConnection, err = acceptVPCPeeringConnection(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return err
		}
	}

	if err := modifyVPCPeeringConnectionOptions(conn, d, vpcPeeringConnection, true); err != nil {
		return err
	}

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags.Map()); err != nil {
			return fmt.Errorf("error creating EC2 VPC Peering Connection (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceVPCPeeringConnectionRead(d, meta)
}

func resourceVPCPeeringAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN]  EC2 VPC Peering Connection (%s) not deleted, removing from state", d.Id())

	return nil
}
