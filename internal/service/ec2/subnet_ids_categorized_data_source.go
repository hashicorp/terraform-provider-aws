package ec2

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

type subnetInfo struct {
	publicSubnetIds       []string
	privateSubnetIds      []string
	privateNatSubnetIds   []string
	privateNoNatSubnetIds []string
}

func DataSourceSubnetIDsCategorized() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSubnetIDsCategorizedRead,
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"public_subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_subnet_nat_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_subnet_nonat_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSubnetIDsCategorizedRead(d *schema.ResourceData, meta interface{}) error {
	var vpcId string
	conn := meta.(*conns.AWSClient).EC2Conn

	if vpc, vpcOk := d.GetOk("vpc_id"); vpcOk {
		vpcId = vpc.(string)
		pSubnetInfo, err := getCategorizedSubnets(conn, vpcId)

		if err != nil {
			return err
		}

		log.Printf("[INFO] aws_subnet_ids_categorized: %s: %+v\n", vpcId, *pSubnetInfo)

		d.SetId(vpcId)
		d.Set("public_subnet_ids", (*pSubnetInfo).publicSubnetIds)
		d.Set("private_subnet_ids", (*pSubnetInfo).privateSubnetIds)
		d.Set("private_subnet_nat_ids", (*pSubnetInfo).privateNatSubnetIds)
		d.Set("private_subnet_nonat_ids", (*pSubnetInfo).privateNoNatSubnetIds)
	} else {
		return errors.New("Error reading vpc_id argument")
	}

	return nil
}

// Read IGW, subnets and route tables for specified VPC, and then categorize to public and private.
// Returns: Two arrays - one of public subnet IDs and one of private subnet IDs.
func getCategorizedSubnets(conn *ec2.EC2, vpcId string) (*subnetInfo, error) {
	si := subnetInfo{
		publicSubnetIds:       []string{},
		privateSubnetIds:      []string{},
		privateNatSubnetIds:   []string{},
		privateNoNatSubnetIds: []string{},
	}

	allSubnetIds, err := getAllSubnetIds(conn, vpcId)

	if err != nil {
		return nil, fmt.Errorf("error reading EC2 Subnets: %w", err)
	}

	log.Printf("[INFO] aws_subnet_ids_categorized: %s: %d subnets retrieved\n", vpcId, len(allSubnetIds))

	if len(allSubnetIds) == 0 {
		// No subnets at all - nothing more to do.
		return &si, nil
	}

	internetGateway, err := findInternetGateway(conn, vpcId)

	if err != nil {
		return nil, fmt.Errorf("error reading EC2 Internet Gateways: %w", err)
	}

	haveInternetGateway := internetGateway != nil

	if haveInternetGateway {
		log.Printf("[INFO] aws_subnet_ids_categorized: %s: IGW retrieved: %s\n", vpcId, aws.StringValue(internetGateway.InternetGatewayId))
	} else {
		log.Printf("[INFO] aws_subnet_ids_categorized: %s: No IGW retrieved\n", vpcId)
	}

	routeTables, err := FindRouteTables(conn, &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"vpc-id": vpcId,
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("error reading EC2 Route Tables: %w", err)
	}

	haveRouteTables := len(routeTables) > 0

	log.Printf("[INFO] aws_subnet_ids_categorized: %s: %d Route tables retrieved\n", vpcId, len(routeTables))

	// Condition for all subnets are private and isolated
	// NAT won't work without IGW
	if !(haveRouteTables && haveInternetGateway) {
		si.privateNoNatSubnetIds = allSubnetIds
		si.privateSubnetIds = allSubnetIds

		return &si, nil
	}

	si.publicSubnetIds, si.privateSubnetIds = categorizeSubnetIds(internetGateway, routeTables, allSubnetIds)

	// Now read NAT gateways
	natGateways, err := findNatGateways(conn, vpcId)

	if err != nil {
		return nil, fmt.Errorf("error reading EC2 NAT gateways: %w", err)
	}

	log.Printf("[INFO] aws_subnet_ids_categorized: %s: %d NAT gateways retrieved\n", vpcId, len(natGateways))

	if len(natGateways) > 0 {
		// further categorize private subnets
		si.privateNatSubnetIds, si.privateNoNatSubnetIds = categorizePrivateSubnetIds(natGateways, routeTables, si.privateSubnetIds)
	} else {
		// All private subnets are no-NAT
		si.privateNoNatSubnetIds = si.privateSubnetIds
	}

	return &si, nil
}

// Categorize subnets into public and private given an IGW, at least one route table and at least one subnet.
// Returns: Two arrays - one of public subnet IDs and one of private subnet IDs.
func categorizeSubnetIds(igw *ec2.InternetGateway, routeTables []*ec2.RouteTable, allSubnetIds []string) ([]string, []string) {
	if igw == nil {
		// No IGW = All private.
		return []string{}, allSubnetIds
	}

	// Find the public route table (routes to IGW)
	publicRouteTable := findRouteTableForGateway(routeTables, aws.StringValue(igw.InternetGatewayId))

	if publicRouteTable == nil {
		// All subnets are private
		return []string{}, allSubnetIds
	}

	var publicSubnetIds []string
	if !*publicRouteTable.Associations[0].Main {
		// Gather all subnets associated with this RTB
		// which are the public subnets
		for _, assoc := range publicRouteTable.Associations {
			if assoc.SubnetId != nil {
				publicSubnetIds = append(publicSubnetIds, aws.StringValue(assoc.SubnetId))
			}
		}

		// What reamins is therefore private
		return publicSubnetIds, setDifference(allSubnetIds, publicSubnetIds)
	}

	// If we get here, public route table is main route table

	if len(routeTables) == 1 {
		// If no other route tables then
		// all subnets are therefore public
		return allSubnetIds, []string{}
	}

	// There are other route tables, thus any subnets associated with those are private
	// and what remains is public

	var privateSubnetIds []string
	for _, t := range routeTables {
		for _, assoc := range t.Associations {
			if assoc.SubnetId != nil {
				privateSubnetIds = append(privateSubnetIds, aws.StringValue(assoc.SubnetId))
			}
		}
	}

	return setDifference(allSubnetIds, privateSubnetIds), privateSubnetIds
}

// Categorize private subnets into those which have a route to a NAT gateway and those which don't
// Returns: Two arrays - one of routed private subnet IDs and one of isolated private subnet IDs.
func categorizePrivateSubnetIds(natGateways []*ec2.NatGateway, routeTables []*ec2.RouteTable, privateSubnetIds []string) ([]string, []string) {
	routedSubnetIds := []string{}

	for _, gw := range natGateways {
		routeTable := findRouteTableForGateway(routeTables, aws.StringValue(gw.NatGatewayId))

		if routeTable == nil {
			continue
		}

		for _, assoc := range routeTable.Associations {
			if assoc.SubnetId != nil {
				routedSubnetIds = append(routedSubnetIds, aws.StringValue(assoc.SubnetId))
			}
		}
	}

	return routedSubnetIds, setDifference(privateSubnetIds, routedSubnetIds)
}

// Find route table with route to specified gateway
// Returns: Pointer to route table; else nil if no route found
func findRouteTableForGateway(routeTables []*ec2.RouteTable, gatewayId string) *ec2.RouteTable {
	for _, t := range routeTables {
		for _, r := range t.Routes {
			if (r.GatewayId != nil && aws.StringValue(r.GatewayId) == gatewayId) || (r.NatGatewayId != nil && aws.StringValue(r.NatGatewayId) == gatewayId) {
				log.Printf("[INFO] aws_subnet_ids_categorized: Route table for %s: %s\n", gatewayId, aws.StringValue(t.RouteTableId))
				return t
			}
		}
	}

	log.Printf("[WARN] aws_subnet_ids_categorized: Route table for %s not found\n", gatewayId)
	return nil
}

// Find all NAT gateways (if any) on specified VPC
// Returns: Array of NAT gateways
func findNatGateways(conn *ec2.EC2, vpcId string) ([]*ec2.NatGateway, error) {
	req := &ec2.DescribeNatGatewaysInput{}

	req.Filter = BuildAttributeFilterList(
		map[string]string{
			"vpc-id": vpcId,
		},
	)

	resp, err := conn.DescribeNatGateways(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.NatGateways) == 0 {
		return []*ec2.NatGateway{}, nil
	}

	return resp.NatGateways, nil
}

// Get subnet IDs of all subnets in specified VPC
// Returns: Array of subnet IDs
func getAllSubnetIds(conn *ec2.EC2, vpcId string) ([]string, error) {
	subnets, err := FindSubnets(conn, &ec2.DescribeSubnetsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"vpc-id": vpcId,
		}),
	})

	if err != nil {
		return nil, err
	}

	if len(subnets) == 0 {
		// No subnets at all
		return []string{}, nil
	}

	var subnetIds []string

	for _, subnet := range subnets {
		subnetIds = append(subnetIds, aws.StringValue(subnet.SubnetId))
	}

	return subnetIds, nil
}

// Get the internet gateway for the specified VPC
// Returns: InternetGateway, or nil if there isn't one
func findInternetGateway(conn *ec2.EC2, vpcId string) (*ec2.InternetGateway, error) {

	igws, err := FindInternetGateways(conn, &ec2.DescribeInternetGatewaysInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.vpc-id": vpcId,
		}),
	})

	if err != nil {
		return nil, err
	}

	if len(igws) > 0 {
		return igws[0], nil
	}

	return nil, nil
}

// Compute the set difference of two string arrays
// Returns: Elements in a that are not in b
func setDifference(a, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
