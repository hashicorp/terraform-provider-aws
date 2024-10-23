// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var routeTableValidDestinations = []string{
	names.AttrCIDRBlock,
	"ipv6_cidr_block",
	"destination_prefix_list_id",
}

var routeTableValidTargets = []string{
	"carrier_gateway_id",
	"core_network_arn",
	"egress_only_gateway_id",
	"gateway_id",
	"local_gateway_id",
	"nat_gateway_id",
	names.AttrNetworkInterfaceID,
	names.AttrTransitGatewayID,
	names.AttrVPCEndpointID,
	"vpc_peering_connection_id",
}

// @SDKResource("aws_route_table", name="Route Table")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.RouteTable")
// @Testing(generator=false)
func resourceRouteTable() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"propagating_vgws": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
						names.AttrCIDRBlock: {
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
						"local_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"nat_gateway_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrTransitGatewayID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrVPCEndpointID: {
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

func resourceRouteTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateRouteTableInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeRouteTable),
		VpcId:             aws.String(d.Get(names.AttrVPCID).(string)),
	}

	output, err := conn.CreateRouteTable(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route Table: %s", err)
	}

	d.SetId(aws.ToString(output.RouteTable.RouteTableId))

	if _, err := waitRouteTableReady(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route Table (%s) create: %s", d.Id(), err)
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
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findRouteTableByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route Table (%s): %s", d.Id(), err)
	}

	routeTable := outputRaw.(*awstypes.RouteTable)
	ownerID := aws.ToString(routeTable.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("route-table/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrOwnerID, ownerID)
	propagatingVGWs := make([]string, 0, len(routeTable.PropagatingVgws))
	for _, v := range routeTable.PropagatingVgws {
		propagatingVGWs = append(propagatingVGWs, aws.ToString(v.GatewayId))
	}
	if err := d.Set("propagating_vgws", propagatingVGWs); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting propagating_vgws: %s", err)
	}
	if err := d.Set("route", flattenRoutes(ctx, conn, d, routeTable.Routes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting route: %s", err)
	}
	d.Set(names.AttrVPCID, routeTable.VpcId)

	// Ignore the AmazonFSx service tag in addition to standard ignores.
	setTagsOut(ctx, Tags(keyValueTags(ctx, routeTable.Tags).Ignore(tftags.New(ctx, []string{"AmazonFSx"}))))

	return diags
}

func resourceRouteTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

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

	return append(diags, resourceRouteTableRead(ctx, d, meta)...)
}

func resourceRouteTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	routeTable, err := findRouteTableByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route Table (%s): %s", d.Id(), err)
	}

	// Do all the disassociations
	for _, v := range routeTable.Associations {
		v := aws.ToString(v.RouteTableAssociationId)

		if err := routeTableAssociationDelete(ctx, conn, v, d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Route Table (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting Route Table: %s", d.Id())
	_, err = conn.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route Table (%s): %s", d.Id(), err)
	}

	if _, err := waitRouteTableDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route Table (%s) delete: %s", d.Id(), err)
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
		buf.WriteString(fmt.Sprintf("%s-", itypes.CanonicalCIDRBlock(v.(string))))
	}

	if v, ok := m[names.AttrCIDRBlock]; ok {
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

	if v, ok := m[names.AttrTransitGatewayID]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["local_gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m[names.AttrVPCEndpointID]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["vpc_peering_connection_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m[names.AttrNetworkInterfaceID]; ok && !natGatewaySet {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}

// routeTableAddRoute adds a route to the specified route table.
func routeTableAddRoute(ctx context.Context, conn *ec2.Client, routeTableID string, tfMap map[string]interface{}, timeout time.Duration) error {
	if err := validNestedExactlyOneOf(tfMap, routeTableValidDestinations); err != nil {
		return fmt.Errorf("creating route: %w", err)
	}
	if err := validNestedExactlyOneOf(tfMap, routeTableValidTargets); err != nil {
		return fmt.Errorf("creating route: %w", err)
	}

	destinationAttributeKey, destination := routeTableRouteDestinationAttribute(tfMap)

	var routeFinder routeFinder

	switch destinationAttributeKey {
	case "cidr_block":
		routeFinder = findRouteByIPv4Destination
	case "ipv6_cidr_block":
		routeFinder = findRouteByIPv6Destination
	case "destination_prefix_list_id":
		routeFinder = findRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("creating Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	input := expandCreateRouteInput(tfMap)

	if input == nil {
		return nil
	}

	input.RouteTableId = aws.String(routeTableID)

	_, target := routeTableRouteTargetAttribute(tfMap)

	if target == gatewayIDLocal {
		// created by AWS so probably doesn't need a retry but just to be sure
		// we provide a small one
		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, time.Second*15,
			func() (interface{}, error) {
				return routeFinder(ctx, conn, routeTableID, destination)
			},
			errCodeInvalidRouteNotFound,
		)

		if tfresource.NotFound(err) {
			return fmt.Errorf("local route cannot be created but must exist to be adopted, %s %s does not exist", target, destination)
		}

		if err != nil {
			return fmt.Errorf("finding local route %s %s: %w", target, destination, err)
		}

		return nil
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout,
		func() (interface{}, error) {
			return conn.CreateRoute(ctx, input)
		},
		errCodeInvalidParameterException,
		errCodeInvalidTransitGatewayIDNotFound,
	)

	if err != nil {
		return fmt.Errorf("creating Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	if _, err := waitRouteReady(ctx, conn, routeFinder, routeTableID, destination, timeout); err != nil {
		return fmt.Errorf("waiting for Route in Route Table (%s) with destination (%s) create: %w", routeTableID, destination, err)
	}

	return nil
}

// routeTableDeleteRoute deletes a route from the specified route table.
func routeTableDeleteRoute(ctx context.Context, conn *ec2.Client, routeTableID string, tfMap map[string]interface{}, timeout time.Duration) error {
	destinationAttributeKey, destination := routeTableRouteDestinationAttribute(tfMap)

	input := &ec2.DeleteRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder routeFinder

	switch destination := aws.String(destination); destinationAttributeKey {
	case "cidr_block":
		input.DestinationCidrBlock = destination
		routeFinder = findRouteByIPv4Destination
	case "ipv6_cidr_block":
		input.DestinationIpv6CidrBlock = destination
		routeFinder = findRouteByIPv6Destination
	case "destination_prefix_list_id":
		input.DestinationPrefixListId = destination
		routeFinder = findRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("deleting Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	_, err := conn.DeleteRoute(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	if _, err := waitRouteDeleted(ctx, conn, routeFinder, routeTableID, destination, timeout); err != nil {
		return fmt.Errorf("waiting for Route in Route Table (%s) with destination (%s) delete: %w", routeTableID, destination, err)
	}

	return nil
}

// routeTableUpdateRoute updates a route in the specified route table.
func routeTableUpdateRoute(ctx context.Context, conn *ec2.Client, routeTableID string, tfMap map[string]interface{}, timeout time.Duration) error {
	if err := validNestedExactlyOneOf(tfMap, routeTableValidDestinations); err != nil {
		return fmt.Errorf("updating route: %w", err)
	}
	if err := validNestedExactlyOneOf(tfMap, routeTableValidTargets); err != nil {
		return fmt.Errorf("updating route: %w", err)
	}

	destinationAttributeKey, destination := routeTableRouteDestinationAttribute(tfMap)

	var routeFinder routeFinder

	switch destinationAttributeKey {
	case "cidr_block":
		routeFinder = findRouteByIPv4Destination
	case "ipv6_cidr_block":
		routeFinder = findRouteByIPv6Destination
	case "destination_prefix_list_id":
		routeFinder = findRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("creating Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	input := expandReplaceRouteInput(tfMap)

	if input == nil {
		return nil
	}

	input.RouteTableId = aws.String(routeTableID)

	_, err := conn.ReplaceRoute(ctx, input)

	if err != nil {
		return fmt.Errorf("updating Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	if _, err := waitRouteReady(ctx, conn, routeFinder, routeTableID, destination, timeout); err != nil {
		return fmt.Errorf("waiting for Route in Route Table (%s) with destination (%s) update: %w", routeTableID, destination, err)
	}

	return nil
}

// routeTableDisableVGWRoutePropagation attempts to disable VGW route propagation.
// Any error is returned.
func routeTableDisableVGWRoutePropagation(ctx context.Context, conn *ec2.Client, routeTableID, gatewayID string) error {
	input := &ec2.DisableVgwRoutePropagationInput{
		GatewayId:    aws.String(gatewayID),
		RouteTableId: aws.String(routeTableID),
	}

	_, err := conn.DisableVgwRoutePropagation(ctx, input)

	if err != nil {
		return fmt.Errorf("disabling Route Table (%s) VPN Gateway (%s) route propagation: %w", routeTableID, gatewayID, err)
	}

	return nil
}

// routeTableEnableVGWRoutePropagation attempts to enable VGW route propagation.
// The specified eventual consistency timeout is respected.
// Any error is returned.
func routeTableEnableVGWRoutePropagation(ctx context.Context, conn *ec2.Client, routeTableID, gatewayID string, timeout time.Duration) error {
	input := &ec2.EnableVgwRoutePropagationInput{
		GatewayId:    aws.String(gatewayID),
		RouteTableId: aws.String(routeTableID),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout,
		func() (interface{}, error) {
			return conn.EnableVgwRoutePropagation(ctx, input)
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

	if v, ok := tfMap[names.AttrCIDRBlock].(string); ok && v != "" {
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

	if v, ok := tfMap["local_gateway_id"].(string); ok && v != "" {
		apiObject.LocalGatewayId = aws.String(v)
	}

	if v, ok := tfMap["nat_gateway_id"].(string); ok && v != "" {
		apiObject.NatGatewayId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrNetworkInterfaceID].(string); ok && v != "" {
		apiObject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTransitGatewayID].(string); ok && v != "" {
		apiObject.TransitGatewayId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVPCEndpointID].(string); ok && v != "" {
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

	if v, ok := tfMap[names.AttrCIDRBlock].(string); ok && v != "" {
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
		if v == gatewayIDLocal {
			apiObject.LocalTarget = aws.Bool(true)
		} else {
			apiObject.GatewayId = aws.String(v)
		}
	}

	if v, ok := tfMap["local_gateway_id"].(string); ok && v != "" {
		apiObject.LocalGatewayId = aws.String(v)
	}

	if v, ok := tfMap["nat_gateway_id"].(string); ok && v != "" {
		apiObject.NatGatewayId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrNetworkInterfaceID].(string); ok && v != "" {
		apiObject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTransitGatewayID].(string); ok && v != "" {
		apiObject.TransitGatewayId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVPCEndpointID].(string); ok && v != "" {
		apiObject.VpcEndpointId = aws.String(v)
	}

	if v, ok := tfMap["vpc_peering_connection_id"].(string); ok && v != "" {
		apiObject.VpcPeeringConnectionId = aws.String(v)
	}

	return apiObject
}

func flattenRoute(apiObject *awstypes.Route) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DestinationCidrBlock; v != nil {
		tfMap[names.AttrCIDRBlock] = aws.ToString(v)
	}

	if v := apiObject.DestinationIpv6CidrBlock; v != nil {
		tfMap["ipv6_cidr_block"] = aws.ToString(v)
	}

	if v := apiObject.DestinationPrefixListId; v != nil {
		tfMap["destination_prefix_list_id"] = aws.ToString(v)
	}

	if v := apiObject.CarrierGatewayId; v != nil {
		tfMap["carrier_gateway_id"] = aws.ToString(v)
	}

	if v := apiObject.CoreNetworkArn; v != nil {
		tfMap["core_network_arn"] = aws.ToString(v)
	}

	if v := apiObject.EgressOnlyInternetGatewayId; v != nil {
		tfMap["egress_only_gateway_id"] = aws.ToString(v)
	}

	if v := apiObject.GatewayId; v != nil {
		if strings.HasPrefix(aws.ToString(v), "vpce-") {
			tfMap[names.AttrVPCEndpointID] = aws.ToString(v)
		} else {
			tfMap["gateway_id"] = aws.ToString(v)
		}
	}

	if v := apiObject.LocalGatewayId; v != nil {
		tfMap["local_gateway_id"] = aws.ToString(v)
	}

	if v := apiObject.NatGatewayId; v != nil {
		tfMap["nat_gateway_id"] = aws.ToString(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap[names.AttrNetworkInterfaceID] = aws.ToString(v)
	}

	if v := apiObject.TransitGatewayId; v != nil {
		tfMap[names.AttrTransitGatewayID] = aws.ToString(v)
	}

	if v := apiObject.VpcPeeringConnectionId; v != nil {
		tfMap["vpc_peering_connection_id"] = aws.ToString(v)
	}

	return tfMap
}

func flattenRoutes(ctx context.Context, conn *ec2.Client, d *schema.ResourceData, apiObjects []awstypes.Route) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if gatewayID := aws.ToString(apiObject.GatewayId); gatewayID == gatewayIDVPCLattice {
			continue
		}

		// local routes from config need to be included but not default local routes, as determined by hasLocalConfig
		// see local route tests
		if gatewayID := aws.ToString(apiObject.GatewayId); gatewayID == gatewayIDLocal && !hasLocalConfig(d, apiObject) {
			continue
		}

		if apiObject.Origin == awstypes.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if apiObject.DestinationPrefixListId != nil && strings.HasPrefix(aws.ToString(apiObject.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		// Skip cross-account ENIs for AWS services.
		if networkInterfaceID := aws.ToString(apiObject.NetworkInterfaceId); networkInterfaceID != "" {
			networkInterface, err := findNetworkInterfaceByID(ctx, conn, networkInterfaceID)

			if err == nil && networkInterface.Attachment != nil {
				if ownerID, instanceOwnerID := aws.ToString(networkInterface.OwnerId), aws.ToString(networkInterface.Attachment.InstanceOwnerId); ownerID != "" && instanceOwnerID != ownerID {
					continue
				}
			}
		}

		tfList = append(tfList, flattenRoute(&apiObject))
	}

	return tfList
}

// hasLocalConfig along with flattenRoutes prevents default local routes from
// being stored in state but allows configured local routes to be stored in
// state. hasLocalConfig checks the ResourceData and flattenRoutes skips or
// includes the route. Normally, you can't count on ResourceData to represent
// config. However, in this case, a local gateway route in ResourceData must
// come from config because of the gatekeeping done by hasLocalConfig and
// flattenRoutes.
func hasLocalConfig(d *schema.ResourceData, apiObject awstypes.Route) bool {
	if v, ok := d.GetOk("route"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			v := v.(map[string]interface{})
			if v[names.AttrCIDRBlock].(string) != aws.ToString(apiObject.DestinationCidrBlock) &&
				v["destination_prefix_list_id"] != aws.ToString(apiObject.DestinationPrefixListId) &&
				v["ipv6_cidr_block"] != aws.ToString(apiObject.DestinationIpv6CidrBlock) {
				continue
			}

			if v["gateway_id"].(string) == gatewayIDLocal {
				return true
			}
		}
	}

	return false
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
