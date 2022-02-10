package ec2

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	id := d.Get("vpc_peering_connection_id").(string)

	_, statusCode, err := vpcPeeringConnectionRefreshState(conn, id)()

	if err != nil && statusCode != ec2.VpcPeeringConnectionStateReasonCodeFailed {
		return fmt.Errorf("error reading VPC Peering Connection (%s): %s", id, err)
	}

	status := map[string]bool{
		ec2.VpcPeeringConnectionStateReasonCodeDeleted:  true,
		ec2.VpcPeeringConnectionStateReasonCodeDeleting: true,
		ec2.VpcPeeringConnectionStateReasonCodeExpired:  true,
		ec2.VpcPeeringConnectionStateReasonCodeFailed:   true,
		ec2.VpcPeeringConnectionStateReasonCodeRejected: true,
		"": true, // AWS consistency issue, see vpcPeeringConnectionRefreshState
	}
	if _, ok := status[statusCode]; ok {
		return fmt.Errorf("VPC Peering Connection (%s) in unexpected status for acceptance: %s", id, statusCode)
	}

	d.SetId(id)

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags.Map()); err != nil {
			return fmt.Errorf("error adding tags: %s", err)
		}
	}

	// TODO: Call Read.
	return resourceVPCPeeringConnectionUpdate(d, meta)
}

func resourceVPCPeeringAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete VPC peering connection. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func vpcPeeringConnectionRefreshState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeVpcPeeringConnections(&ec2.DescribeVpcPeeringConnectionsInput{
			VpcPeeringConnectionIds: aws.StringSlice([]string{id}),
		})
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidVpcPeeringConnectionID.NotFound", "") {
				return nil, ec2.VpcPeeringConnectionStateReasonCodeDeleted, nil
			}

			return nil, "", err
		}

		if resp == nil || resp.VpcPeeringConnections == nil ||
			len(resp.VpcPeeringConnections) == 0 || resp.VpcPeeringConnections[0] == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our peering connection yet. Return an empty state.
			return nil, "", nil
		}
		pc := resp.VpcPeeringConnections[0]
		if pc.Status == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our peering connection yet. Return an empty state.
			return nil, "", nil
		}
		statusCode := aws.StringValue(pc.Status.Code)

		// A VPC Peering Connection can exist in a failed state due to
		// incorrect VPC ID, account ID, or overlapping IP address range,
		// thus we short circuit before the time out would occur.
		if statusCode == ec2.VpcPeeringConnectionStateReasonCodeFailed {
			return nil, statusCode, errors.New(aws.StringValue(pc.Status.Message))
		}

		return pc, statusCode, nil
	}
}
