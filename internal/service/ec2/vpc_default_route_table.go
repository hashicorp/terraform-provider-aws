package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDefaultRouteTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceDefaultRouteTableCreate,
		Read:   resourceDefaultRouteTableRead,
		Update: resourceRouteTableUpdate,
		Delete: resourceDefaultRouteTableDelete,

		Importer: &schema.ResourceImporter{
			State: resourceDefaultRouteTableImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
		},

		//
		// The top-level attributes must be a superset of the aws_route_table resource's attributes as common CRUD handlers are used.
		//
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"default_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"propagating_vgws": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"route": {
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Computed:   true,
				Optional:   true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						///
						// Destinations.
						///
						"cidr_block": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
						},
						"destination_prefix_list_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipv6_cidr_block": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
						},

						//
						// Targets.
						// These target attributes are a subset of the aws_route_table resource's target attributes
						// as there are some targets that are not allowed in the default route table for a VPC.
						//
						"core_network_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"egress_only_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"instance_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"nat_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"network_interface_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"transit_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vpc_endpoint_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vpc_peering_connection_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceRouteTableHash,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDefaultRouteTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	routeTableID := d.Get("default_route_table_id").(string)

	routeTable, err := FindRouteTableByID(conn, routeTableID)

	if err != nil {
		return fmt.Errorf("error reading EC2 Default Route Table (%s): %w", routeTableID, err)
	}

	d.SetId(aws.StringValue(routeTable.RouteTableId))

	// Remove all existing VGW associations.
	for _, v := range routeTable.PropagatingVgws {
		if err := ec2RouteTableDisableVgwRoutePropagation(conn, d.Id(), aws.StringValue(v.GatewayId)); err != nil {
			return err
		}
	}

	// Delete all existing routes.
	for _, v := range routeTable.Routes {
		// you cannot delete the local route
		if aws.StringValue(v.GatewayId) == "local" {
			continue
		}

		if aws.StringValue(v.Origin) == ec2.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if v.DestinationPrefixListId != nil && strings.HasPrefix(aws.StringValue(v.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		input := &ec2.DeleteRouteInput{
			RouteTableId: aws.String(d.Id()),
		}

		var destination string
		var routeFinder RouteFinder

		if v.DestinationCidrBlock != nil {
			input.DestinationCidrBlock = v.DestinationCidrBlock
			destination = aws.StringValue(v.DestinationCidrBlock)
			routeFinder = FindRouteByIPv4Destination
		} else if v.DestinationIpv6CidrBlock != nil {
			input.DestinationIpv6CidrBlock = v.DestinationIpv6CidrBlock
			destination = aws.StringValue(v.DestinationIpv6CidrBlock)
			routeFinder = FindRouteByIPv6Destination
		} else if v.DestinationPrefixListId != nil {
			input.DestinationPrefixListId = v.DestinationPrefixListId
			destination = aws.StringValue(v.DestinationPrefixListId)
			routeFinder = FindRouteByPrefixListIDDestination
		}

		log.Printf("[DEBUG] Deleting Route: %s", input)
		_, err := conn.DeleteRoute(input)

		if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error deleting Route in EC2 Default Route Table (%s) with destination (%s): %w", d.Id(), destination, err)
		}

		_, err = WaitRouteDeleted(conn, routeFinder, routeTableID, destination, d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return fmt.Errorf("error waiting for Route in EC2 Default Route Table (%s) with destination (%s) to delete: %w", d.Id(), destination, err)
		}
	}

	// Add new VGW associations.
	if v, ok := d.GetOk("propagating_vgws"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(string)

			if err := ec2RouteTableEnableVgwRoutePropagation(conn, d.Id(), v, d.Timeout(schema.TimeoutCreate)); err != nil {
				return err
			}
		}
	}

	// Add new routes.
	if v, ok := d.GetOk("route"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(map[string]interface{})

			if err := ec2RouteTableAddRoute(conn, d.Id(), v, d.Timeout(schema.TimeoutCreate)); err != nil {
				return err
			}
		}
	}

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("error adding tags: %w", err)
		}
	}

	return resourceDefaultRouteTableRead(d, meta)
}

func resourceDefaultRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	d.Set("default_route_table_id", d.Id())

	// re-use regular AWS Route Table READ. This is an extra API call but saves us
	// from trying to manually keep parity
	return resourceRouteTableRead(d, meta)
}

func resourceDefaultRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default Route Table. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func resourceDefaultRouteTableImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).EC2Conn

	routeTable, err := FindMainRouteTableByVPCID(conn, d.Id())

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(routeTable.RouteTableId))

	return []*schema.ResourceData{d}, nil
}
