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
		},
	}
}

func dataSourceSubnetIDsCategorizedRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if vpc, vpcOk := d.GetOk("vpc_id"); vpcOk {
		publicSubnets, privateSubnets, err := getCategorizedSubnets(conn, vpc.(string))

		if err != nil {
			return err
		}

		d.SetId(d.Get("vpc_id").(string))
		d.Set("public_subnet_ids", publicSubnets)
		d.Set("private_subnet_ids", privateSubnets)
	} else {
		return errors.New("Error reading vpc_id argument")
	}

	return nil
}

// Read IGW, subnets and route tables for specified VPC, and then categorize to public and private.
// Returns: Two arrays - one of public subnet IDs and one of private subnet IDs.
func getCategorizedSubnets(conn *ec2.EC2, vpcId string) ([]string, []string, error) {
	allSubnetIds, err := getAllSubnetIds(conn, vpcId)

	if err != nil {
		return nil, nil, fmt.Errorf("error reading EC2 Subnets: %w", err)
	}

	log.Printf("[INFO] aws_subnet_ids_categorized: %s: %d subnets retrieved\n", vpcId, len(allSubnetIds))

	if len(allSubnetIds) == 0 {
		// No subnets at all - return empty arrays.
		return []string{}, []string{}, nil
	}

	igw, err := getInternetGateway(conn, vpcId)

	if err != nil {
		return nil, nil, fmt.Errorf("error reading EC2 Internet Gateways: %w", err)
	}

	if igw == nil {
		log.Printf("[INFO] aws_subnet_ids_categorized: %s: No IGW retrieved\n", vpcId)
		// No IGW, then all subnets are private.
		return []string{}, allSubnetIds, nil
	}

	log.Printf("[INFO] aws_subnet_ids_categorized: %s: IGW retrieved: %s\n", vpcId, aws.StringValue(igw.InternetGatewayId))

	routeTables, err := FindRouteTables(conn, &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"vpc-id": vpcId,
		}),
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error reading EC2 Route Tables: %w", err)
	}

	log.Printf("[INFO] aws_subnet_ids_categorized: %s: %d Route tables retrieved\n", vpcId, len(routeTables))

	if len(routeTables) == 0 {
		// No route tables (unlikely), then all subnets are private.
		return []string{}, allSubnetIds, nil
	}

	publicSubnetIds, privateSubnetIds := categorizeSubnetIds(igw, routeTables, allSubnetIds)

	log.Printf("[INFO] aws_subnet_ids_categorized: %s: Public subnets: %+q\n", vpcId, publicSubnetIds)
	log.Printf("[INFO] aws_subnet_ids_categorized: %s: Private subnets: %+q\n", vpcId, privateSubnetIds)

	return publicSubnetIds, privateSubnetIds, nil
}

// Categorize subnets into public and private given an IGW, at least one route table and at least one subnet.
// Returns: Two arrays - one of public subnet IDs and one of private subnet IDs.
func categorizeSubnetIds(igw *ec2.InternetGateway, routeTables []*ec2.RouteTable, allSubnetIds []string) ([]string, []string) {
	var publicRouteTable *ec2.RouteTable

	publicRouteTable = nil

	// Find the public route table (routes to IGW)
	for _, t := range routeTables {
		for _, r := range t.Routes {
			if r.GatewayId != nil && aws.StringValue(r.GatewayId) == aws.StringValue(igw.InternetGatewayId) {
				publicRouteTable = t
				break
			}
		}

		if publicRouteTable != nil {
			break
		}
	}

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
func getInternetGateway(conn *ec2.EC2, vpcId string) (*ec2.InternetGateway, error) {

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
