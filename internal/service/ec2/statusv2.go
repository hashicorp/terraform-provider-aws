// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusVPCStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPCIPv6CIDRBlockAssociationStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _, err := findVPCIPv6CIDRBlockAssociationByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, string(output.Ipv6CidrBlockState.State), nil
	}
}

func statusVPCAttributeValueV2(ctx context.Context, conn *ec2.Client, id string, attribute awstypes.VpcAttributeName) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attributeValue, err := findVPCAttributeV2(ctx, conn, id, attribute)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return attributeValue, strconv.FormatBool(attributeValue), nil
	}
}

func statusNetworkInterfaceV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNetworkInterfaceByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNetworkInterfaceAttachmentV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNetworkInterfaceAttachmentByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusVPCEndpointStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCEndpointByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusRouteV2(ctx context.Context, conn *ec2.Client, routeFinder routeFinderV2, routeTableID, destination string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := routeFinder(ctx, conn, routeTableID, destination)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, RouteStatusReady, nil
	}
}

func statusRouteTableV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRouteTableByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, RouteTableStatusReady, nil
	}
}

func statusRouteTableAssociationStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRouteTableAssociationByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.AssociationState == nil {
			// In ISO partitions AssociationState can be nil.
			// If the association has been found then we assume it's associated.
			state := awstypes.RouteTableAssociationStateCodeAssociated

			return &awstypes.RouteTableAssociationState{State: state}, string(state), nil
		}

		return output.AssociationState, string(output.AssociationState.State), nil
	}
}

func statusVPCEndpointServiceStateAvailableV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindVPCEndpointServiceConfigurationByID as it maps useful status codes to NotFoundError.
		output, err := findVPCEndpointServiceConfigurationV2(ctx, conn, &ec2.DescribeVpcEndpointServiceConfigurationsInput{
			ServiceIds: []string{id},
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ServiceState), nil
	}
}

func statusVPCEndpointServiceStateDeletedV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCEndpointServiceConfigurationByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ServiceState), nil
	}
}

func statusVPCEndpointRouteTableAssociationV2(ctx context.Context, conn *ec2.Client, vpcEndpointID, routeTableID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := findVPCEndpointRouteTableAssociationExistsV2(ctx, conn, vpcEndpointID, routeTableID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return "", VPCEndpointRouteTableAssociationStatusReady, nil
	}
}

func statusVPCEndpointConnectionVPCEndpointStateV2(ctx context.Context, conn *ec2.Client, serviceID, vpcEndpointID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCEndpointConnectionByServiceIDAndVPCEndpointIDV2(ctx, conn, serviceID, vpcEndpointID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.VpcEndpointState), nil
	}
}

func StatusIPAMState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusIPAMPoolState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMPoolByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusIPAMPoolCIDRState(ctx context.Context, conn *ec2.Client, cidrBlock, poolID, poolCidrId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		if cidrBlock == "" {
			output, err := FindIPAMPoolCIDRByPoolCIDRId(ctx, conn, poolCidrId, poolID)

			if tfresource.NotFound(err) {
				return nil, "", nil
			}

			if err != nil {
				return nil, "", err
			}
			cidrBlock = aws.ToString(output.Cidr)
		}

		output, err := FindIPAMPoolCIDRByTwoPartKey(ctx, conn, cidrBlock, poolID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

const (
	// naming mapes to the SDK constants that exist for IPAM
	IpamPoolCIDRAllocationCreateComplete = "create-complete" // nosemgrep:ci.caps2-in-const-name, ci.caps2-in-var-name, ci.caps5-in-const-name, ci.caps5-in-var-name
)

func StatusIPAMPoolCIDRAllocationState(ctx context.Context, conn *ec2.Client, allocationID, poolID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMPoolAllocationByTwoPartKey(ctx, conn, allocationID, poolID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, IpamPoolCIDRAllocationCreateComplete, nil
	}
}

func StatusIPAMResourceDiscoveryState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMResourceDiscoveryByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusIPAMResourceDiscoveryAssociationStatus(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMResourceDiscoveryAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusIPAMScopeState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMScopeByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}
