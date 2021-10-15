package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//Flattens network interface attachment into a map[string]interface
func FlattenAttachment(a *ec2.NetworkInterfaceAttachment) map[string]interface{} {
	att := make(map[string]interface{})
	if a.InstanceId != nil {
		att["instance"] = *a.InstanceId
	}
	if a.DeviceIndex != nil {
		att["device_index"] = *a.DeviceIndex
	}
	if a.AttachmentId != nil {
		att["attachment_id"] = *a.AttachmentId
	}
	return att
}

func flattenAttributeValues(l []*ec2.AttributeValue) []string {
	values := make([]string, 0, len(l))
	for _, v := range l {
		values = append(values, aws.StringValue(v.Value))
	}
	return values
}

//Flattens security group identifiers into a []string, where the elements returned are the GroupIDs
func FlattenGroupIdentifiers(dtos []*ec2.GroupIdentifier) []string {
	ids := make([]string, 0, len(dtos))
	for _, v := range dtos {
		group_id := *v.GroupId
		ids = append(ids, group_id)
	}
	return ids
}

// Like ec2.GroupIdentifier but with additional rule description.
type GroupIdentifier struct {
	// The ID of the security group.
	GroupId *string

	// The name of the security group.
	GroupName *string

	Description *string
}

func expandIP6Addresses(ips []interface{}) []*ec2.InstanceIpv6Address {
	dtos := make([]*ec2.InstanceIpv6Address, 0, len(ips))
	for _, v := range ips {
		ipv6Address := &ec2.InstanceIpv6Address{
			Ipv6Address: aws.String(v.(string)),
		}

		dtos = append(dtos, ipv6Address)
	}
	return dtos
}

// Takes the result of flatmap.Expand for an array of ingress/egress security
// group rules and returns EC2 API compatible objects. This function will error
// if it finds invalid permissions input, namely a protocol of "-1" with either
// to_port or from_port set to a non-zero value.
func ExpandIPPerms(
	group *ec2.SecurityGroup, configured []interface{}) ([]*ec2.IpPermission, error) {
	vpc := group.VpcId != nil && *group.VpcId != ""

	perms := make([]*ec2.IpPermission, len(configured))
	for i, mRaw := range configured {
		var perm ec2.IpPermission
		m := mRaw.(map[string]interface{})

		perm.FromPort = aws.Int64(int64(m["from_port"].(int)))
		perm.ToPort = aws.Int64(int64(m["to_port"].(int)))
		perm.IpProtocol = aws.String(m["protocol"].(string))

		// When protocol is "-1", AWS won't store any ports for the
		// rule, but also won't error if the user specifies ports other
		// than '0'. Force the user to make a deliberate '0' port
		// choice when specifying a "-1" protocol, and tell them about
		// AWS's behavior in the error message.
		if *perm.IpProtocol == "-1" && (*perm.FromPort != 0 || *perm.ToPort != 0) {
			return nil, fmt.Errorf(
				"from_port (%d) and to_port (%d) must both be 0 to use the 'ALL' \"-1\" protocol!",
				*perm.FromPort, *perm.ToPort)
		}

		var groups []string
		if raw, ok := m["security_groups"]; ok {
			list := raw.(*schema.Set).List()
			for _, v := range list {
				groups = append(groups, v.(string))
			}
		}
		if v, ok := m["self"]; ok && v.(bool) {
			if vpc {
				groups = append(groups, *group.GroupId)
			} else {
				groups = append(groups, *group.GroupName)
			}
		}

		if len(groups) > 0 {
			perm.UserIdGroupPairs = make([]*ec2.UserIdGroupPair, len(groups))
			for i, name := range groups {
				ownerId, id := "", name
				if items := strings.Split(id, "/"); len(items) > 1 {
					ownerId, id = items[0], items[1]
				}

				perm.UserIdGroupPairs[i] = &ec2.UserIdGroupPair{
					GroupId: aws.String(id),
				}

				if ownerId != "" {
					perm.UserIdGroupPairs[i].UserId = aws.String(ownerId)
				}

				if !vpc {
					perm.UserIdGroupPairs[i].GroupId = nil
					perm.UserIdGroupPairs[i].GroupName = aws.String(id)
				}
			}
		}

		if raw, ok := m["cidr_blocks"]; ok {
			list := raw.([]interface{})
			for _, v := range list {
				perm.IpRanges = append(perm.IpRanges, &ec2.IpRange{CidrIp: aws.String(v.(string))})
			}
		}
		if raw, ok := m["ipv6_cidr_blocks"]; ok {
			list := raw.([]interface{})
			for _, v := range list {
				perm.Ipv6Ranges = append(perm.Ipv6Ranges, &ec2.Ipv6Range{CidrIpv6: aws.String(v.(string))})
			}
		}

		if raw, ok := m["prefix_list_ids"]; ok {
			list := raw.([]interface{})
			for _, v := range list {
				perm.PrefixListIds = append(perm.PrefixListIds, &ec2.PrefixListId{PrefixListId: aws.String(v.(string))})
			}
		}

		if raw, ok := m["description"]; ok {
			description := raw.(string)
			if description != "" {
				for _, v := range perm.IpRanges {
					v.Description = aws.String(description)
				}
				for _, v := range perm.Ipv6Ranges {
					v.Description = aws.String(description)
				}
				for _, v := range perm.PrefixListIds {
					v.Description = aws.String(description)
				}
				for _, v := range perm.UserIdGroupPairs {
					v.Description = aws.String(description)
				}
			}
		}

		perms[i] = &perm
	}

	return perms, nil
}

func flattenNetworkInterfaceAssociation(a *ec2.NetworkInterfaceAssociation) []interface{} {
	tfMap := map[string]interface{}{}

	if a.AllocationId != nil {
		tfMap["allocation_id"] = aws.StringValue(a.AllocationId)
	}
	if a.AssociationId != nil {
		tfMap["association_id"] = aws.StringValue(a.AssociationId)
	}
	if a.CarrierIp != nil {
		tfMap["carrier_ip"] = aws.StringValue(a.CarrierIp)
	}
	if a.CustomerOwnedIp != nil {
		tfMap["customer_owned_ip"] = aws.StringValue(a.CustomerOwnedIp)
	}
	if a.IpOwnerId != nil {
		tfMap["ip_owner_id"] = aws.StringValue(a.IpOwnerId)
	}
	if a.PublicDnsName != nil {
		tfMap["public_dns_name"] = aws.StringValue(a.PublicDnsName)
	}
	if a.PublicIp != nil {
		tfMap["public_ip"] = aws.StringValue(a.PublicIp)
	}

	return []interface{}{tfMap}
}

func flattenNetworkInterfaceIPv6Address(niia []*ec2.NetworkInterfaceIpv6Address) []string {
	ips := make([]string, 0, len(niia))
	for _, v := range niia {
		ips = append(ips, *v.Ipv6Address)
	}
	return ips
}

//Flattens an array of private ip addresses into a []string, where the elements returned are the IP strings e.g. "192.168.0.0"
func FlattenNetworkInterfacesPrivateIPAddresses(dtos []*ec2.NetworkInterfacePrivateIpAddress) []string {
	ips := make([]string, 0, len(dtos))
	for _, v := range dtos {
		ip := *v.PrivateIpAddress
		ips = append(ips, ip)
	}
	return ips
}

//Expands an array of IPs into a ec2 Private IP Address Spec
func ExpandPrivateIPAddresses(ips []interface{}) []*ec2.PrivateIpAddressSpecification {
	dtos := make([]*ec2.PrivateIpAddressSpecification, 0, len(ips))
	for i, v := range ips {
		new_private_ip := &ec2.PrivateIpAddressSpecification{
			PrivateIpAddress: aws.String(v.(string)),
		}

		new_private_ip.Primary = aws.Bool(i == 0)

		dtos = append(dtos, new_private_ip)
	}
	return dtos
}

// Flattens an array of UserSecurityGroups into a []*GroupIdentifier
func FlattenSecurityGroups(list []*ec2.UserIdGroupPair, ownerId *string) []*GroupIdentifier {
	result := make([]*GroupIdentifier, 0, len(list))
	for _, g := range list {
		var userId *string
		if g.UserId != nil && *g.UserId != "" && (ownerId == nil || *ownerId != *g.UserId) {
			userId = g.UserId
		}
		// userid nil here for same vpc groups

		vpc := g.GroupName == nil || *g.GroupName == ""
		var id *string
		if vpc {
			id = g.GroupId
		} else {
			id = g.GroupName
		}

		// id is groupid for vpcs
		// id is groupname for non vpc (classic)

		if userId != nil {
			id = aws.String(*userId + "/" + *id)
		}

		if vpc {
			result = append(result, &GroupIdentifier{
				GroupId:     id,
				Description: g.Description,
			})
		} else {
			result = append(result, &GroupIdentifier{
				GroupId:     g.GroupId,
				GroupName:   id,
				Description: g.Description,
			})
		}
	}
	return result
}

func expandVPCPeeringConnectionOptions(vOptions []interface{}, crossRegionPeering bool) *ec2.PeeringConnectionOptionsRequest {
	if len(vOptions) == 0 || vOptions[0] == nil {
		return nil
	}

	mOptions := vOptions[0].(map[string]interface{})

	options := &ec2.PeeringConnectionOptionsRequest{}

	if v, ok := mOptions["allow_remote_vpc_dns_resolution"].(bool); ok {
		options.AllowDnsResolutionFromRemoteVpc = aws.Bool(v)
	}
	if !crossRegionPeering {
		if v, ok := mOptions["allow_classic_link_to_remote_vpc"].(bool); ok {
			options.AllowEgressFromLocalClassicLinkToRemoteVpc = aws.Bool(v)
		}
		if v, ok := mOptions["allow_vpc_to_remote_classic_link"].(bool); ok {
			options.AllowEgressFromLocalVpcToRemoteClassicLink = aws.Bool(v)
		}
	}

	return options
}

func flattenVPCPeeringConnectionOptions(options *ec2.VpcPeeringConnectionOptionsDescription) []interface{} {
	// When the VPC Peering Connection is pending acceptance,
	// the details about accepter and/or requester peering
	// options would not be included in the response.
	if options == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"allow_remote_vpc_dns_resolution":  aws.BoolValue(options.AllowDnsResolutionFromRemoteVpc),
		"allow_classic_link_to_remote_vpc": aws.BoolValue(options.AllowEgressFromLocalClassicLinkToRemoteVpc),
		"allow_vpc_to_remote_classic_link": aws.BoolValue(options.AllowEgressFromLocalVpcToRemoteClassicLink),
	}}
}
