package ec2

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	carrierGatewayStateNotFound = "NotFound"
	carrierGatewayStateUnknown  = "Unknown"
)

// StatusCarrierGatewayState fetches the CarrierGateway and its State
func StatusCarrierGatewayState(conn *ec2.EC2, carrierGatewayID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		carrierGateway, err := FindCarrierGatewayByID(conn, carrierGatewayID)
		if tfawserr.ErrCodeEquals(err, ErrCodeInvalidCarrierGatewayIDNotFound) {
			return nil, carrierGatewayStateNotFound, nil
		}
		if err != nil {
			return nil, carrierGatewayStateUnknown, err
		}

		if carrierGateway == nil {
			return nil, carrierGatewayStateNotFound, nil
		}

		state := aws.StringValue(carrierGateway.State)

		if state == ec2.CarrierGatewayStateDeleted {
			return nil, carrierGatewayStateNotFound, nil
		}

		return carrierGateway, state, nil
	}
}

// StatusLocalGatewayRouteTableVPCAssociationState fetches the LocalGatewayRouteTableVpcAssociation and its State
func StatusLocalGatewayRouteTableVPCAssociationState(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
			LocalGatewayRouteTableVpcAssociationIds: aws.StringSlice([]string{localGatewayRouteTableVpcAssociationID}),
		}

		output, err := conn.DescribeLocalGatewayRouteTableVpcAssociations(input)

		if err != nil {
			return nil, "", err
		}

		var association *ec2.LocalGatewayRouteTableVpcAssociation

		for _, outputAssociation := range output.LocalGatewayRouteTableVpcAssociations {
			if outputAssociation == nil {
				continue
			}

			if aws.StringValue(outputAssociation.LocalGatewayRouteTableVpcAssociationId) == localGatewayRouteTableVpcAssociationID {
				association = outputAssociation
				break
			}
		}

		if association == nil {
			return association, ec2.RouteTableAssociationStateCodeDisassociated, nil
		}

		return association, aws.StringValue(association.State), nil
	}
}

const (
	ClientVPNEndpointStatusNotFound = "NotFound"

	ClientVPNEndpointStatusUnknown = "Unknown"
)

// StatusClientVPNEndpoint fetches the Client VPN endpoint and its Status
func StatusClientVPNEndpoint(conn *ec2.EC2, endpointID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeClientVpnEndpoints(&ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: aws.StringSlice([]string{endpointID}),
		})
		if tfawserr.ErrCodeEquals(err, ErrCodeClientVpnEndpointIdNotFound) {
			return nil, ClientVPNEndpointStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVPNEndpointStatusUnknown, err
		}

		if result == nil || len(result.ClientVpnEndpoints) == 0 || result.ClientVpnEndpoints[0] == nil {
			return nil, ClientVPNEndpointStatusNotFound, nil
		}

		endpoint := result.ClientVpnEndpoints[0]
		if endpoint.Status == nil || endpoint.Status.Code == nil {
			return endpoint, ClientVPNEndpointStatusUnknown, nil
		}

		return endpoint, aws.StringValue(endpoint.Status.Code), nil
	}
}

const (
	ClientVPNAuthorizationRuleStatusNotFound = "NotFound"

	ClientVPNAuthorizationRuleStatusUnknown = "Unknown"
)

// StatusClientVPNAuthorizationRule fetches the Client VPN authorization rule and its Status
func StatusClientVPNAuthorizationRule(conn *ec2.EC2, authorizationRuleID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := FindClientVPNAuthorizationRuleByID(conn, authorizationRuleID)
		if tfawserr.ErrCodeEquals(err, ErrCodeClientVpnAuthorizationRuleNotFound) {
			return nil, ClientVPNAuthorizationRuleStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVPNAuthorizationRuleStatusUnknown, err
		}

		if result == nil || len(result.AuthorizationRules) == 0 || result.AuthorizationRules[0] == nil {
			return nil, ClientVPNAuthorizationRuleStatusNotFound, nil
		}

		if len(result.AuthorizationRules) > 1 {
			return nil, ClientVPNAuthorizationRuleStatusUnknown, fmt.Errorf("internal error: found %d results for Client VPN authorization rule (%s) status, need 1", len(result.AuthorizationRules), authorizationRuleID)
		}

		rule := result.AuthorizationRules[0]
		if rule.Status == nil || rule.Status.Code == nil {
			return rule, ClientVPNAuthorizationRuleStatusUnknown, nil
		}

		return rule, aws.StringValue(rule.Status.Code), nil
	}
}

const (
	ClientVPNNetworkAssociationStatusNotFound = "NotFound"

	ClientVPNNetworkAssociationStatusUnknown = "Unknown"
)

func StatusClientVPNNetworkAssociation(conn *ec2.EC2, cvnaID string, cvepID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(cvepID),
			AssociationIds:      []*string{aws.String(cvnaID)},
		})

		if tfawserr.ErrCodeEquals(err, ErrCodeClientVpnAssociationIdNotFound) || tfawserr.ErrCodeEquals(err, ErrCodeClientVpnEndpointIdNotFound) {
			return nil, ClientVPNNetworkAssociationStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVPNNetworkAssociationStatusUnknown, err
		}

		if result == nil || len(result.ClientVpnTargetNetworks) == 0 || result.ClientVpnTargetNetworks[0] == nil {
			return nil, ClientVPNNetworkAssociationStatusNotFound, nil
		}

		network := result.ClientVpnTargetNetworks[0]
		if network.Status == nil || network.Status.Code == nil {
			return network, ClientVPNNetworkAssociationStatusUnknown, nil
		}

		return network, aws.StringValue(network.Status.Code), nil
	}
}

const (
	ClientVPNRouteStatusNotFound = "NotFound"

	ClientVPNRouteStatusUnknown = "Unknown"
)

// StatusClientVPNRoute fetches the Client VPN route and its Status
func StatusClientVPNRoute(conn *ec2.EC2, routeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := FindClientVPNRouteByID(conn, routeID)
		if tfawserr.ErrCodeEquals(err, ErrCodeClientVpnRouteNotFound) {
			return nil, ClientVPNRouteStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVPNRouteStatusUnknown, err
		}

		if result == nil || len(result.Routes) == 0 || result.Routes[0] == nil {
			return nil, ClientVPNRouteStatusNotFound, nil
		}

		if len(result.Routes) > 1 {
			return nil, ClientVPNRouteStatusUnknown, fmt.Errorf("internal error: found %d results for Client VPN route (%s) status, need 1", len(result.Routes), routeID)
		}

		rule := result.Routes[0]
		if rule.Status == nil || rule.Status.Code == nil {
			return rule, ClientVPNRouteStatusUnknown, nil
		}

		return rule, aws.StringValue(rule.Status.Code), nil
	}
}

// StatusInstanceIAMInstanceProfile fetches the Instance and its IamInstanceProfile
//
// The EC2 API accepts a name and always returns an ARN, so it is converted
// back to the name to prevent unexpected differences.
func StatusInstanceIAMInstanceProfile(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := FindInstanceByID(conn, id)

		if tfawserr.ErrCodeEquals(err, ErrCodeInvalidInstanceIDNotFound) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if instance == nil {
			return nil, "", nil
		}

		if instance.IamInstanceProfile == nil || instance.IamInstanceProfile.Arn == nil {
			return instance, "", nil
		}

		name, err := tfiam.InstanceProfileARNToName(aws.StringValue(instance.IamInstanceProfile.Arn))

		if err != nil {
			return instance, "", err
		}

		return instance, name, nil
	}
}

func StatusNATGatewayState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNATGatewayByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

const (
	RouteStatusReady = "ready"
)

func StatusRoute(conn *ec2.EC2, routeFinder RouteFinder, routeTableID, destination string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := routeFinder(conn, routeTableID, destination)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, RouteStatusReady, nil
	}
}

const (
	RouteTableStatusReady = "ready"
)

func StatusRouteTable(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRouteTableByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, RouteTableStatusReady, nil
	}
}

func StatusRouteTableAssociationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRouteTableAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.AssociationState, aws.StringValue(output.AssociationState.State), nil
	}
}

const (
	SecurityGroupStatusCreated = "Created"
)

func StatusSecurityGroup(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSecurityGroupByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, SecurityGroupStatusCreated, nil
	}
}

func StatusSubnetState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusSubnetIPv6CIDRBlockAssociationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetIPv6CIDRBlockAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, aws.StringValue(output.Ipv6CidrBlockState.State), nil
	}
}

func StatusSubnetAssignIpv6AddressOnCreation(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.AssignIpv6AddressOnCreation)), nil
	}
}

func StatusSubnetEnableDns64(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.EnableDns64)), nil
	}
}

func StatusSubnetEnableResourceNameDnsAAAARecordOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord)), nil
	}
}

func StatusSubnetEnableResourceNameDnsARecordOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord)), nil
	}
}

func StatusSubnetMapCustomerOwnedIPOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.MapCustomerOwnedIpOnLaunch)), nil
	}
}

func StatusSubnetMapPublicIPOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.MapPublicIpOnLaunch)), nil
	}
}

func StatusSubnetPrivateDNSHostnameTypeOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.PrivateDnsNameOptionsOnLaunch.HostnameType), nil
	}
}

func StatusTransitGatewayPrefixListReferenceState(conn *ec2.EC2, transitGatewayRouteTableID string, prefixListID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		transitGatewayPrefixListReference, err := FindTransitGatewayPrefixListReference(conn, transitGatewayRouteTableID, prefixListID)

		if err != nil {
			return nil, "", err
		}

		if transitGatewayPrefixListReference == nil {
			return nil, "", nil
		}

		return transitGatewayPrefixListReference, aws.StringValue(transitGatewayPrefixListReference.State), nil
	}
}

func StatusTransitGatewayRouteTablePropagationState(conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		transitGatewayRouteTablePropagation, err := FindTransitGatewayRouteTablePropagation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if err != nil {
			return nil, "", err
		}

		if transitGatewayRouteTablePropagation == nil {
			return nil, "", nil
		}

		return transitGatewayRouteTablePropagation, aws.StringValue(transitGatewayRouteTablePropagation.State), nil
	}
}

func StatusVPCState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPCAttributeValue(conn *ec2.EC2, id string, attribute string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attributeValue, err := FindVPCAttribute(conn, id, attribute)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return attributeValue, strconv.FormatBool(attributeValue), nil
	}
}

func StatusVPCCIDRBlockAssociationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _, err := FindVPCCIDRBlockAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.CidrBlockState, aws.StringValue(output.CidrBlockState.State), nil
	}
}

func StatusVPCIPv6CIDRBlockAssociationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _, err := FindVPCIPv6CIDRBlockAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, aws.StringValue(output.Ipv6CidrBlockState.State), nil
	}
}

const (
	vpcPeeringConnectionStatusNotFound = "NotFound"
	vpcPeeringConnectionStatusUnknown  = "Unknown"
)

// StatusVPCPeeringConnection fetches the VPC peering connection and its status
func StatusVPCPeeringConnection(conn *ec2.EC2, vpcPeeringConnectionID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpcPeeringConnection, err := FindVPCPeeringConnectionByID(conn, vpcPeeringConnectionID)
		if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpcPeeringConnectionIDNotFound) {
			return nil, vpcPeeringConnectionStatusNotFound, nil
		}
		if err != nil {
			return nil, vpcPeeringConnectionStatusUnknown, err
		}

		// Sometimes AWS just has consistency issues and doesn't see
		// our peering connection yet. Return an empty state.
		if vpcPeeringConnection == nil || vpcPeeringConnection.Status == nil {
			return nil, vpcPeeringConnectionStatusNotFound, nil
		}

		statusCode := aws.StringValue(vpcPeeringConnection.Status.Code)

		// https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-basics.html#vpc-peering-lifecycle
		switch statusCode {
		case ec2.VpcPeeringConnectionStateReasonCodeFailed:
			log.Printf("[WARN] VPC Peering Connection (%s): %s: %s", vpcPeeringConnectionID, statusCode, aws.StringValue(vpcPeeringConnection.Status.Message))
			fallthrough
		case ec2.VpcPeeringConnectionStateReasonCodeDeleted, ec2.VpcPeeringConnectionStateReasonCodeExpired, ec2.VpcPeeringConnectionStateReasonCodeRejected:
			return nil, vpcPeeringConnectionStatusNotFound, nil
		}

		return vpcPeeringConnection, statusCode, nil
	}
}

func StatusVPNGatewayVPCAttachmentState(conn *ec2.EC2, vpnGatewayID, vpcID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNGatewayVPCAttachment(conn, vpnGatewayID, vpcID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusCustomerGatewayState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCustomerGatewayByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPNConnectionState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNConnectionByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPNConnectionRouteState(conn *ec2.EC2, vpnConnectionID, cidrBlock string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNConnectionRouteByVPNConnectionIDAndCIDR(conn, vpnConnectionID, cidrBlock)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusHostState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindHostByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusInternetGatewayAttachmentState(conn *ec2.EC2, internetGatewayID, vpcID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInternetGatewayAttachment(conn, internetGatewayID, vpcID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusManagedPrefixListState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindManagedPrefixListByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusNetworkInterfaceStatus(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNetworkInterfaceByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusNetworkInterfaceAttachmentStatus(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNetworkInterfaceAttachmentByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusPlacementGroupState(conn *ec2.EC2, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindPlacementGroupByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPCEndpointState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCEndpointByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

const (
	VPCEndpointRouteTableAssociationStatusReady = "ready"
)

func StatusVPCEndpointRouteTableAssociation(conn *ec2.EC2, vpcEndpointID, routeTableID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := FindVPCEndpointRouteTableAssociationExists(conn, vpcEndpointID, routeTableID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return "", VPCEndpointRouteTableAssociationStatusReady, nil
	}
}

const (
	snapshotImportNotFound = "NotFound"
)

func StatusEBSSnapshotImport(conn *ec2.EC2, importTaskId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		params := &ec2.DescribeImportSnapshotTasksInput{
			ImportTaskIds: []*string{aws.String(importTaskId)},
		}

		resp, err := conn.DescribeImportSnapshotTasks(params)
		if err != nil {
			return nil, "", err
		}

		if resp == nil || len(resp.ImportSnapshotTasks) < 1 {
			return nil, snapshotImportNotFound, nil
		}

		if task := resp.ImportSnapshotTasks[0]; task != nil {
			detail := task.SnapshotTaskDetail
			if detail.Status != nil && aws.StringValue(detail.Status) == EBSSnapshotImportStateDeleting {
				err = fmt.Errorf("Snapshot import task is deleting")
			}
			return detail, aws.StringValue(detail.Status), err
		} else {
			return nil, snapshotImportNotFound, nil
		}
	}
}

func statusVPCEndpointConnectionVPCEndpointState(conn *ec2.EC2, serviceID, vpcEndpointID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(conn, serviceID, vpcEndpointID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.VpcEndpointState), nil
	}
}

func StatusSnapshotTierStatus(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSnapshotTierStatusById(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.StorageTier), nil
	}
}
