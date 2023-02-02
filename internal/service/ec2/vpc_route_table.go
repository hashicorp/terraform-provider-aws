package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var routeTableValidDestinations = []string{
	"cidr_block",
	"ipv6_cidr_block",
	"destination_prefix_list_id",
}

var routeTableValidTargets = []string{
	"carrier_gateway_id",
	"core_network_arn",
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

func ResourceRouteTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteTableCreate,
		ReadWithoutTimeout:   resourceRouteTableRead,
		UpdateWithoutTimeout: resourceRouteTableUpdate,
		DeleteWithoutTimeout: resourceRouteTableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
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
						//
						"carrier_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
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
							Type:       schema.TypeString,
							Optional:   true,
							Deprecated: "Use network_interface_id instead",
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
				Set: resourceRouteTableHash,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRouteTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateRouteTableInput{
		VpcId:             aws.String(d.Get("vpc_id").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeRouteTable),
	}

	log.Printf("[DEBUG] Creating Route Table: %s", input)
	output, err := conn.CreateRouteTableWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route Table: %s", err)
	}

	d.SetId(aws.StringValue(output.RouteTable.RouteTableId))

	if _, err := WaitRouteTableReady(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route Table (%s) to become available: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("propagating_vgws"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(string)

			if err := routeTableEnableVGWRoutePropagation(ctx, conn, d.Id(), v, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if v, ok := d.GetOk("route"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(map[string]interface{})

			if err := routeTableAddRoute(ctx, conn, d.Id(), v, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceRouteTableRead(ctx, d, meta)...)
}

func resourceRouteTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	routeTable, err := FindRouteTableByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route Table (%s): %s", d.Id(), err)
	}

	d.Set("vpc_id", routeTable.VpcId)

	propagatingVGWs := make([]string, 0, len(routeTable.PropagatingVgws))
	for _, v := range routeTable.PropagatingVgws {
		propagatingVGWs = append(propagatingVGWs, aws.StringValue(v.GatewayId))
	}
	if err := d.Set("propagating_vgws", propagatingVGWs); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting propagating_vgws: %s", err)
	}

	if err := d.Set("route", flattenRoutes(ctx, conn, routeTable.Routes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting route: %s", err)
	}

	//Ignore the AmazonFSx service tag in addition to standard ignores
	tags := KeyValueTags(routeTable.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Ignore(tftags.New([]string{"AmazonFSx"}))

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	ownerID := aws.StringValue(routeTable.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("route-table/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", ownerID)

	return diags
}

func resourceRouteTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("propagating_vgws") {
		o, n := d.GetChange("propagating_vgws")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		del := os.Difference(ns).List()
		add := ns.Difference(os).List()

		for _, v := range del {
			v := v.(string)

			if err := routeTableDisableVGWRoutePropagation(ctx, conn, d.Id(), v); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		for _, v := range add {
			v := v.(string)

			if err := routeTableEnableVGWRoutePropagation(ctx, conn, d.Id(), v, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("route") {
		o, n := d.GetChange("route")

		for _, new := range n.(*schema.Set).List() {
			vNew := new.(map[string]interface{})

			_, newDestination := routeTableRouteDestinationAttribute(vNew)
			_, newTarget := routeTableRouteTargetAttribute(vNew)

			addRoute := true

			for _, old := range o.(*schema.Set).List() {
				vOld := old.(map[string]interface{})

				_, oldDestination := routeTableRouteDestinationAttribute(vOld)
				_, oldTarget := routeTableRouteTargetAttribute(vOld)

				if oldDestination == newDestination {
					addRoute = false

					if oldTarget != newTarget {
						if err := routeTableUpdateRoute(ctx, conn, d.Id(), vNew, d.Timeout(schema.TimeoutUpdate)); err != nil {
							return sdkdiag.AppendFromErr(diags, err)
						}
					}
				}
			}

			if addRoute {
				if err := routeTableAddRoute(ctx, conn, d.Id(), vNew, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}

		for _, old := range o.(*schema.Set).List() {
			vOld := old.(map[string]interface{})

			_, oldDestination := routeTableRouteDestinationAttribute(vOld)

			delRoute := true

			for _, new := range n.(*schema.Set).List() {
				vNew := new.(map[string]interface{})

				_, newDestination := routeTableRouteDestinationAttribute(vNew)

				if newDestination == oldDestination {
					delRoute = false
				}
			}

			if delRoute {
				if err := routeTableDeleteRoute(ctx, conn, d.Id(), vOld, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	if d.HasChange("tags_all") && !d.IsNewResource() {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Route Table (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRouteTableRead(ctx, d, meta)...)
}

func resourceRouteTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	routeTable, err := FindRouteTableByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route Table (%s): %s", d.Id(), err)
	}

	// Do all the disassociations
	for _, v := range routeTable.Associations {
		v := aws.StringValue(v.RouteTableAssociationId)

		if err := routeTableAssociationDelete(ctx, conn, v); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Route Table (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting Route Table: %s", d.Id())
	_, err = conn.DeleteRouteTableWithContext(ctx, &ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route Table (%s): %s", d.Id(), err)
	}

	// Wait for the route table to really destroy
	log.Printf("[DEBUG] Waiting for route table (%s) deletion", d.Id())
	if _, err := WaitRouteTableDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route Table (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func resourceRouteTableHash(v interface{}) int {
	var buf bytes.Buffer
	m, castOk := v.(map[string]interface{})
	if !castOk {
		return 0
	}

	if v, ok := m["ipv6_cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", verify.CanonicalCIDRBlock(v.(string))))
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

	if v, ok := m["core_network_arn"]; ok {
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

	return create.StringHashcode(buf.String())
}

// routeTableAddRoute adds a route to the specified route table.
func routeTableAddRoute(ctx context.Context, conn *ec2.EC2, routeTableID string, tfMap map[string]interface{}, timeout time.Duration) error {
	if err := validNestedExactlyOneOf(tfMap, routeTableValidDestinations); err != nil {
		return fmt.Errorf("creating route: %w", err)
	}
	if err := validNestedExactlyOneOf(tfMap, routeTableValidTargets); err != nil {
		return fmt.Errorf("creating route: %w", err)
	}

	destinationAttributeKey, destination := routeTableRouteDestinationAttribute(tfMap)

	var routeFinder RouteFinder

	switch destinationAttributeKey {
	case "cidr_block":
		routeFinder = FindRouteByIPv4Destination
	case "ipv6_cidr_block":
		routeFinder = FindRouteByIPv6Destination
	case "destination_prefix_list_id":
		routeFinder = FindRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("creating Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	input := expandCreateRouteInput(tfMap)

	if input == nil {
		return nil
	}

	input.RouteTableId = aws.String(routeTableID)

	log.Printf("[DEBUG] Creating Route: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout,
		func() (interface{}, error) {
			return conn.CreateRouteWithContext(ctx, input)
		},
		errCodeInvalidParameterException,
		errCodeInvalidTransitGatewayIDNotFound,
	)

	if err != nil {
		return fmt.Errorf("creating Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	_, err = WaitRouteReady(ctx, conn, routeFinder, routeTableID, destination, timeout)

	if err != nil {
		return fmt.Errorf("waiting for Route in Route Table (%s) with destination (%s) to become available: %w", routeTableID, destination, err)
	}

	return nil
}

// routeTableDeleteRoute deletes a route from the specified route table.
func routeTableDeleteRoute(ctx context.Context, conn *ec2.EC2, routeTableID string, tfMap map[string]interface{}, timeout time.Duration) error {
	destinationAttributeKey, destination := routeTableRouteDestinationAttribute(tfMap)

	input := &ec2.DeleteRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder RouteFinder

	switch destination := aws.String(destination); destinationAttributeKey {
	case "cidr_block":
		input.DestinationCidrBlock = destination
		routeFinder = FindRouteByIPv4Destination
	case "ipv6_cidr_block":
		input.DestinationIpv6CidrBlock = destination
		routeFinder = FindRouteByIPv6Destination
	case "destination_prefix_list_id":
		input.DestinationPrefixListId = destination
		routeFinder = FindRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("deleting Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	log.Printf("[DEBUG] Deleting Route: %s", input)
	_, err := conn.DeleteRouteWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	_, err = WaitRouteDeleted(ctx, conn, routeFinder, routeTableID, destination, timeout)

	if err != nil {
		return fmt.Errorf("waiting for Route in Route Table (%s) with destination (%s) to delete: %w", routeTableID, destination, err)
	}

	return nil
}

// routeTableUpdateRoute updates a route in the specified route table.
func routeTableUpdateRoute(ctx context.Context, conn *ec2.EC2, routeTableID string, tfMap map[string]interface{}, timeout time.Duration) error {
	if err := validNestedExactlyOneOf(tfMap, routeTableValidDestinations); err != nil {
		return fmt.Errorf("updating route: %w", err)
	}
	if err := validNestedExactlyOneOf(tfMap, routeTableValidTargets); err != nil {
		return fmt.Errorf("updating route: %w", err)
	}

	destinationAttributeKey, destination := routeTableRouteDestinationAttribute(tfMap)

	var routeFinder RouteFinder

	switch destinationAttributeKey {
	case "cidr_block":
		routeFinder = FindRouteByIPv4Destination
	case "ipv6_cidr_block":
		routeFinder = FindRouteByIPv6Destination
	case "destination_prefix_list_id":
		routeFinder = FindRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("creating Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	input := expandReplaceRouteInput(tfMap)

	if input == nil {
		return nil
	}

	input.RouteTableId = aws.String(routeTableID)

	log.Printf("[DEBUG] Updating Route: %s", input)
	_, err := conn.ReplaceRouteWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("updating Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	_, err = WaitRouteReady(ctx, conn, routeFinder, routeTableID, destination, timeout)

	if err != nil {
		return fmt.Errorf("waiting for Route in Route Table (%s) with destination (%s) to become available: %w", routeTableID, destination, err)
	}

	return nil
}

// routeTableDisableVGWRoutePropagation attempts to disable VGW route propagation.
// Any error is returned.
func routeTableDisableVGWRoutePropagation(ctx context.Context, conn *ec2.EC2, routeTableID, gatewayID string) error {
	input := &ec2.DisableVgwRoutePropagationInput{
		GatewayId:    aws.String(gatewayID),
		RouteTableId: aws.String(routeTableID),
	}

	log.Printf("[DEBUG] Disabling Route Table (%s) VPN Gateway (%s) route propagation", routeTableID, gatewayID)
	_, err := conn.DisableVgwRoutePropagationWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("disabling Route Table (%s) VPN Gateway (%s) route propagation: %w", routeTableID, gatewayID, err)
	}

	return nil
}

// routeTableEnableVGWRoutePropagation attempts to enable VGW route propagation.
// The specified eventual consistency timeout is respected.
// Any error is returned.
func routeTableEnableVGWRoutePropagation(ctx context.Context, conn *ec2.EC2, routeTableID, gatewayID string, timeout time.Duration) error {
	input := &ec2.EnableVgwRoutePropagationInput{
		GatewayId:    aws.String(gatewayID),
		RouteTableId: aws.String(routeTableID),
	}

	log.Printf("[DEBUG] Enabling Route Table (%s) VPN Gateway (%s) route propagation", routeTableID, gatewayID)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout,
		func() (interface{}, error) {
			return conn.EnableVgwRoutePropagationWithContext(ctx, input)
		},
		errCodeGatewayNotAttached,
	)

	if err != nil {
		return fmt.Errorf("enabling Route Table (%s) VPN Gateway (%s) route propagation: %w", routeTableID, gatewayID, err)
	}

	return nil
}

func expandCreateRouteInput(tfMap map[string]interface{}) *ec2.CreateRouteInput {
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

	if v, ok := tfMap["core_network_arn"].(string); ok && v != "" {
		apiObject.CoreNetworkArn = aws.String(v)
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

func expandReplaceRouteInput(tfMap map[string]interface{}) *ec2.ReplaceRouteInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.ReplaceRouteInput{}

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

	if v, ok := tfMap["core_network_arn"].(string); ok && v != "" {
		apiObject.CoreNetworkArn = aws.String(v)
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

func flattenRoute(apiObject *ec2.Route) map[string]interface{} {
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

	if v := apiObject.CoreNetworkArn; v != nil {
		tfMap["core_network_arn"] = aws.StringValue(v)
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

	if v := apiObject.NatGatewayId; v != nil {
		tfMap["nat_gateway_id"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
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

func flattenRoutes(ctx context.Context, conn *ec2.EC2, apiObjects []*ec2.Route) []interface{} {
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

		// Skip cross-account ENIs for AWS services.
		if networkInterfaceID := aws.StringValue(apiObject.NetworkInterfaceId); networkInterfaceID != "" {
			networkInterface, err := FindNetworkInterfaceByID(ctx, conn, networkInterfaceID)

			if err == nil && networkInterface.Attachment != nil {
				if ownerID, instanceOwnerID := aws.StringValue(networkInterface.OwnerId), aws.StringValue(networkInterface.Attachment.InstanceOwnerId); ownerID != "" && instanceOwnerID != ownerID {
					log.Printf("[DEBUG] Skip cross-account ENI (%s)", networkInterfaceID)
					continue
				}
			}
		}

		tfList = append(tfList, flattenRoute(apiObject))
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
func routeTableRouteTargetAttribute(m map[string]interface{}) (string, string) { //nolint:unparam
	for _, key := range routeTableValidTargets {
		if v, ok := m[key].(string); ok && v != "" {
			return key, v
		}
	}

	return "", ""
}
