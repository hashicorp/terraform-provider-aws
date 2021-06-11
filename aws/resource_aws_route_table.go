package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfnet "github.com/terraform-providers/terraform-provider-aws/aws/internal/net"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

var routeTableValidDestinations = []string{
	"cidr_block",
	"ipv6_cidr_block",
	"destination_prefix_list_id",
}

var routeTableValidTargets = []string{
	"carrier_gateway_id",
	"egress_only_gateway_id",
	"gateway_id",
	"instance_id",
	"local_gateway_id",
	"nat_gateway_id",
	"network_interface_id",
	"transit_gateway_id",
	"vpc_endpoint_id",
	"vpc_peering_connection_id",
}

func resourceAwsRouteTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRouteTableCreate,
		Read:   resourceAwsRouteTableRead,
		Update: resourceAwsRouteTableUpdate,
		Delete: resourceAwsRouteTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"propagating_vgws": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"route": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						///
						// Destinations.
						///
						"cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validateIpv4CIDRNetworkAddress,
							),
						},
						"destination_prefix_list_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipv6_cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validateIpv6CIDRNetworkAddress,
							),
						},

						//
						// Targets.
						//
						"carrier_gateway_id": {
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
						"local_gateway_id": {
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
				Set: resourceAwsRouteTableHash,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsRouteTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateRouteTableInput{
		VpcId:             aws.String(d.Get("vpc_id").(string)),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeRouteTable),
	}

	log.Printf("[DEBUG] Creating Route Table: %s", input)
	output, err := conn.CreateRouteTable(input)

	if err != nil {
		return fmt.Errorf("error creating Route Table: %w", err)
	}

	d.SetId(aws.StringValue(output.RouteTable.RouteTableId))

	if _, err := waiter.RouteTableReady(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Route Table (%s) to become available: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("propagating_vgws"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(string)

			log.Printf("[DEBUG] Enabling Route Table (%s) VPN Gateway (%s) route propagation", d.Id(), v)
			err = enableVgwRoutePropagation(conn, d.Id(), v, waiter.PropagationTimeout)

			if err != nil {
				return fmt.Errorf("error enabling Route Table (%s) VPN Gateway (%s) route propagation: %w", d.Id(), v, err)
			}
		}
	}

	if v, ok := d.GetOk("route"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(map[string]interface{})

			if err := validateNestedExactlyOneOf(v, routeTableValidDestinations); err != nil {
				return fmt.Errorf("error creating route: %w", err)
			}
			if err := validateNestedExactlyOneOf(v, routeTableValidTargets); err != nil {
				return fmt.Errorf("error creating route: %w", err)
			}

			destinationAttributeKey, destination := routeTableRouteDestinationAttribute(v)

			var routeFinder finder.RouteFinder

			switch destinationAttributeKey {
			case "cidr_block":
				routeFinder = finder.RouteByIPv4Destination
			case "ipv6_cidr_block":
				routeFinder = finder.RouteByIPv6Destination
			case "destination_prefix_list_id":
				routeFinder = finder.RouteByPrefixListIDDestination
			default:
				return fmt.Errorf("error creating Route: unexpected route destination attribute: %q", destinationAttributeKey)
			}

			input := expandEc2CreateRouteInput(v)

			if input == nil {
				continue
			}

			input.RouteTableId = aws.String(d.Id())

			log.Printf("[DEBUG] Creating Route: %s", input)
			_, err = tfresource.RetryWhenAwsErrCodeEquals(
				waiter.PropagationTimeout,
				func() (interface{}, error) {
					return conn.CreateRoute(input)
				},
				tfec2.ErrCodeInvalidParameterException,
				tfec2.ErrCodeInvalidTransitGatewayIDNotFound,
			)

			if err != nil {
				return fmt.Errorf("error creating Route for Route Table (%s) with destination (%s): %w", d.Id(), destination, err)
			}

			_, err = waiter.RouteReady(conn, routeFinder, d.Id(), destination)

			if err != nil {
				return fmt.Errorf("error waiting for Route in Route Table (%s) with destination (%s) to become available: %w", d.Id(), destination, err)
			}
		}
	}

	return resourceAwsRouteTableRead(d, meta)
}

func resourceAwsRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	routeTable, err := finder.RouteTableByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route Table (%s): %w", d.Id(), err)
	}

	d.Set("vpc_id", routeTable.VpcId)

	propagatingVGWs := make([]string, 0, len(routeTable.PropagatingVgws))
	for _, v := range routeTable.PropagatingVgws {
		propagatingVGWs = append(propagatingVGWs, aws.StringValue(v.GatewayId))
	}
	if err := d.Set("propagating_vgws", propagatingVGWs); err != nil {
		return fmt.Errorf("error setting propagating_vgws: %w", err)
	}

	if err := d.Set("route", flattenEc2Routes(routeTable.Routes)); err != nil {
		return fmt.Errorf("error setting route: %w", err)
	}

	tags := keyvaluetags.Ec2KeyValueTags(routeTable.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	ownerID := aws.StringValue(routeTable.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("route-table/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", ownerID)

	return nil
}

func resourceAwsRouteTableUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("propagating_vgws") {
		o, n := d.GetChange("propagating_vgws")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := os.Difference(ns).List()
		add := ns.Difference(os).List()

		// Now first loop through all the old propagations and disable any obsolete ones
		for _, vgw := range remove {
			id := vgw.(string)

			// Disable the propagation as it no longer exists in the config
			log.Printf("[INFO] Deleting VGW propagation from %s: %s", d.Id(), id)
			_, err := conn.DisableVgwRoutePropagation(&ec2.DisableVgwRoutePropagationInput{
				RouteTableId: aws.String(d.Id()),
				GatewayId:    aws.String(id),
			})
			if err != nil {
				return err
			}
		}

		// Make sure we save the state of the currently configured rules
		propagatingVGWs := os.Intersection(ns)
		d.Set("propagating_vgws", propagatingVGWs)

		// Then loop through all the newly configured propagations and enable them
		for _, vgw := range add {
			id := vgw.(string)

			var err error
			for i := 0; i < 5; i++ {
				log.Printf("[INFO] Enabling VGW propagation for %s: %s", d.Id(), id)
				_, err = conn.EnableVgwRoutePropagation(&ec2.EnableVgwRoutePropagationInput{
					RouteTableId: aws.String(d.Id()),
					GatewayId:    aws.String(id),
				})
				if err == nil {
					break
				}

				// If we get a Gateway.NotAttached, it is usually some
				// eventually consistency stuff. So we have to just wait a
				// bit...
				if isAWSErr(err, "Gateway.NotAttached", "") {
					time.Sleep(20 * time.Second)
					continue
				}
			}
			if err != nil {
				return err
			}

			propagatingVGWs.Add(vgw)
			d.Set("propagating_vgws", propagatingVGWs)
		}
	}

	// Check if the route set as a whole has changed
	if d.HasChange("route") {
		o, n := d.GetChange("route")
		ors := o.(*schema.Set).Difference(n.(*schema.Set))
		nrs := n.(*schema.Set).Difference(o.(*schema.Set))

		// Now first loop through all the old routes and delete any obsolete ones
		for _, route := range ors.List() {
			m := route.(map[string]interface{})

			deleteOpts := &ec2.DeleteRouteInput{
				RouteTableId: aws.String(d.Id()),
			}

			if s, ok := m["ipv6_cidr_block"].(string); ok && s != "" {
				deleteOpts.DestinationIpv6CidrBlock = aws.String(s)

				log.Printf("[INFO] Deleting route from %s: %s", d.Id(), m["ipv6_cidr_block"].(string))
			}

			if s, ok := m["cidr_block"].(string); ok && s != "" {
				deleteOpts.DestinationCidrBlock = aws.String(s)

				log.Printf("[INFO] Deleting route from %s: %s", d.Id(), m["cidr_block"].(string))
			}

			if s, ok := m["destination_prefix_list_id"].(string); ok && s != "" {
				deleteOpts.DestinationPrefixListId = aws.String(s)

				log.Printf("[INFO] Deleting route from %s: %s", d.Id(), m["destination_prefix_list_id"].(string))
			}

			_, err := conn.DeleteRoute(deleteOpts)
			if err != nil {
				return err
			}
		}

		// Make sure we save the state of the currently configured rules
		routes := o.(*schema.Set).Intersection(n.(*schema.Set))
		d.Set("route", routes)

		// Then loop through all the newly configured routes and create them
		for _, route := range nrs.List() {
			m := route.(map[string]interface{})

			if err := validateNestedExactlyOneOf(m, routeTableValidDestinations); err != nil {
				return fmt.Errorf("error creating route: %w", err)
			}
			if err := validateNestedExactlyOneOf(m, routeTableValidTargets); err != nil {
				return fmt.Errorf("error creating route: %w", err)
			}

			opts := ec2.CreateRouteInput{
				RouteTableId: aws.String(d.Id()),
			}

			if s, ok := m["transit_gateway_id"].(string); ok && s != "" {
				opts.TransitGatewayId = aws.String(s)
			}

			if s, ok := m["vpc_endpoint_id"].(string); ok && s != "" {
				opts.VpcEndpointId = aws.String(s)
			}

			if s, ok := m["vpc_peering_connection_id"].(string); ok && s != "" {
				opts.VpcPeeringConnectionId = aws.String(s)
			}

			if s, ok := m["network_interface_id"].(string); ok && s != "" {
				opts.NetworkInterfaceId = aws.String(s)
			}

			if s, ok := m["instance_id"].(string); ok && s != "" {
				opts.InstanceId = aws.String(s)
			}

			if s, ok := m["ipv6_cidr_block"].(string); ok && s != "" {
				opts.DestinationIpv6CidrBlock = aws.String(s)
			}

			if s, ok := m["cidr_block"].(string); ok && s != "" {
				opts.DestinationCidrBlock = aws.String(s)
			}

			if s, ok := m["destination_prefix_list_id"].(string); ok && s != "" {
				opts.DestinationPrefixListId = aws.String(s)
			}

			if s, ok := m["gateway_id"].(string); ok && s != "" {
				opts.GatewayId = aws.String(s)
			}

			if s, ok := m["carrier_gateway_id"].(string); ok && s != "" {
				opts.CarrierGatewayId = aws.String(s)
			}

			if s, ok := m["egress_only_gateway_id"].(string); ok && s != "" {
				opts.EgressOnlyInternetGatewayId = aws.String(s)
			}

			if s, ok := m["nat_gateway_id"].(string); ok && s != "" {
				opts.NatGatewayId = aws.String(s)
			}

			if s, ok := m["local_gateway_id"].(string); ok && s != "" {
				opts.LocalGatewayId = aws.String(s)
			}

			log.Printf("[INFO] Creating route for %s: %#v", d.Id(), opts)
			err := resource.Retry(waiter.RouteTableUpdatedTimeout, func() *resource.RetryError {
				_, err := conn.CreateRoute(&opts)

				if isAWSErr(err, "InvalidRouteTableID.NotFound", "") {
					return resource.RetryableError(err)
				}

				if isAWSErr(err, "InvalidTransitGatewayID.NotFound", "") {
					return resource.RetryableError(err)
				}

				if err != nil {
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if isResourceTimeoutError(err) {
				_, err = conn.CreateRoute(&opts)
			}
			if err != nil {
				return fmt.Errorf("error creating route: %w", err)
			}

			routes.Add(route)
			d.Set("route", routes)
		}
	}

	if d.HasChange("tags_all") && !d.IsNewResource() {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Route Table (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsRouteTableRead(d, meta)
}

func resourceAwsRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	routeTable, err := finder.RouteTableByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading Route Table (%s): %w", d.Id(), err)
	}

	// Do all the disassociations
	for _, v := range routeTable.Associations {
		v := aws.StringValue(v.RouteTableAssociationId)

		r := resourceAwsRouteTableAssociation()
		d := r.Data(nil)
		d.SetId(v)

		if err := tfresource.Delete(r, d, meta); err != nil {
			return err
		}
	}

	log.Printf("[INFO] Deleting Route Table: %s", d.Id())
	_, err = conn.DeleteRouteTable(&ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route Table (%s): %w", d.Id(), err)
	}

	// Wait for the route table to really destroy
	log.Printf("[DEBUG] Waiting for route table (%s) deletion", d.Id())
	if _, err := waiter.RouteTableDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Route Table (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsRouteTableHash(v interface{}) int {
	var buf bytes.Buffer
	m, castOk := v.(map[string]interface{})
	if !castOk {
		return 0
	}

	if v, ok := m["ipv6_cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", tfnet.CanonicalCIDRBlock(v.(string))))
	}

	if v, ok := m["cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["destination_prefix_list_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["carrier_gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["egress_only_gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	natGatewaySet := false
	if v, ok := m["nat_gateway_id"]; ok {
		natGatewaySet = v.(string) != ""
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	instanceSet := false
	if v, ok := m["instance_id"]; ok {
		instanceSet = v.(string) != ""
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["transit_gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["local_gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["vpc_endpoint_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["vpc_peering_connection_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["network_interface_id"]; ok && !(instanceSet || natGatewaySet) {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}

// enableVgwRoutePropagation attempts to enable VGW route propagation.
// The specified eventual consistency timeout is respected.
// Any error is returned.
func enableVgwRoutePropagation(conn *ec2.EC2, routeTableID, gatewayID string, timeout time.Duration) error {
	input := &ec2.EnableVgwRoutePropagationInput{
		GatewayId:    aws.String(gatewayID),
		RouteTableId: aws.String(routeTableID),
	}

	err := resource.Retry(timeout, func() *resource.RetryError {
		_, err := conn.EnableVgwRoutePropagation(input)

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeGatewayNotAttached) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.EnableVgwRoutePropagation(input)
	}

	return err
}

func expandEc2CreateRouteInput(tfMap map[string]interface{}) *ec2.CreateRouteInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.CreateRouteInput{}

	if v, ok := tfMap["cidr_block"].(string); ok && v != "" {
		apiObject.DestinationCidrBlock = aws.String(v)
	}

	if v, ok := tfMap["ipv6_cidr_block"].(string); ok && v != "" {
		apiObject.DestinationIpv6CidrBlock = aws.String(v)
	}

	if v, ok := tfMap["destination_prefix_list_id"].(string); ok && v != "" {
		apiObject.DestinationPrefixListId = aws.String(v)
	}

	if v, ok := tfMap["carrier_gateway_id"].(string); ok && v != "" {
		apiObject.CarrierGatewayId = aws.String(v)
	}

	if v, ok := tfMap["egress_only_gateway_id"].(string); ok && v != "" {
		apiObject.EgressOnlyInternetGatewayId = aws.String(v)
	}

	if v, ok := tfMap["gateway_id"].(string); ok && v != "" {
		apiObject.GatewayId = aws.String(v)
	}

	if v, ok := tfMap["instance_id"].(string); ok && v != "" {
		apiObject.InstanceId = aws.String(v)
	}

	if v, ok := tfMap["local_gateway_id"].(string); ok && v != "" {
		apiObject.LocalGatewayId = aws.String(v)
	}

	if v, ok := tfMap["nat_gateway_id"].(string); ok && v != "" {
		apiObject.NatGatewayId = aws.String(v)
	}

	if v, ok := tfMap["network_interface_id"].(string); ok && v != "" {
		apiObject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap["transit_gateway_id"].(string); ok && v != "" {
		apiObject.TransitGatewayId = aws.String(v)
	}

	if v, ok := tfMap["vpc_endpoint_id"].(string); ok && v != "" {
		apiObject.VpcEndpointId = aws.String(v)
	}

	if v, ok := tfMap["vpc_peering_connection_id"].(string); ok && v != "" {
		apiObject.VpcPeeringConnectionId = aws.String(v)
	}

	return apiObject
}

func flattenEc2Route(apiObject *ec2.Route) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DestinationCidrBlock; v != nil {
		tfMap["cidr_block"] = aws.StringValue(v)
	}

	if v := apiObject.DestinationIpv6CidrBlock; v != nil {
		tfMap["ipv6_cidr_block"] = aws.StringValue(v)
	}

	if v := apiObject.DestinationPrefixListId; v != nil {
		tfMap["destination_prefix_list_id"] = aws.StringValue(v)
	}

	if v := apiObject.CarrierGatewayId; v != nil {
		tfMap["carrier_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.EgressOnlyInternetGatewayId; v != nil {
		tfMap["egress_only_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.GatewayId; v != nil {
		if strings.HasPrefix(aws.StringValue(v), "vpce-") {
			tfMap["vpc_endpoint_id"] = aws.StringValue(v)
		} else {
			tfMap["gateway_id"] = aws.StringValue(v)
		}
	}

	if v := apiObject.InstanceId; v != nil {
		tfMap["instance_id"] = aws.StringValue(v)
	}

	if v := apiObject.LocalGatewayId; v != nil {
		tfMap["local_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap["nat_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.TransitGatewayId; v != nil {
		tfMap["network_interface_id"] = aws.StringValue(v)
	}

	if v := apiObject.TransitGatewayId; v != nil {
		tfMap["transit_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.VpcPeeringConnectionId; v != nil {
		tfMap["vpc_peering_connection_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEc2Routes(apiObjects []*ec2.Route) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if aws.StringValue(apiObject.GatewayId) == "local" {
			continue
		}

		if aws.StringValue(apiObject.Origin) == ec2.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if apiObject.DestinationPrefixListId != nil && strings.HasPrefix(aws.StringValue(apiObject.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		tfList = append(tfList, flattenEc2Route(apiObject))
	}

	return tfList
}

// routeTableRouteDestinationAttribute returns the attribute key and value of the route table route's destination.
func routeTableRouteDestinationAttribute(m map[string]interface{}) (string, string) {
	for _, key := range routeTableValidDestinations {
		if v, ok := m[key].(string); ok && v != "" {
			return key, v
		}
	}

	return "", ""
}

// routeTableRouteTargetAttribute returns the attribute key and value of the route table route's target.
func routeTableRouteTargetAttribute(m map[string]interface{}) (string, string) {
	for _, key := range routeTableValidTargets {
		if v, ok := m[key].(string); ok && v != "" {
			return key, v
		}
	}

	return "", ""
}
