//go:build sweep
// +build sweep

package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_ec2_capacity_reservation", &resource.Sweeper{
		Name: "aws_ec2_capacity_reservation",
		F:    sweepCapacityReservations,
	})

	resource.AddTestSweepers("aws_ec2_carrier_gateway", &resource.Sweeper{
		Name: "aws_ec2_carrier_gateway",
		F:    sweepCarrierGateway,
	})

	resource.AddTestSweepers("aws_ec2_client_vpn_endpoint", &resource.Sweeper{
		Name: "aws_ec2_client_vpn_endpoint",
		F:    sweepClientVPNEndpoints,
		Dependencies: []string{
			"aws_ec2_client_vpn_network_association",
		},
	})

	resource.AddTestSweepers("aws_ec2_client_vpn_network_association", &resource.Sweeper{
		Name: "aws_ec2_client_vpn_network_association",
		F:    sweepClientVPNNetworkAssociations,
	})

	resource.AddTestSweepers("aws_ebs_volume", &resource.Sweeper{
		Name: "aws_ebs_volume",
		Dependencies: []string{
			"aws_instance",
		},
		F: sweepEBSVolumes,
	})

	resource.AddTestSweepers("aws_egress_only_internet_gateway", &resource.Sweeper{
		Name: "aws_egress_only_internet_gateway",
		F:    sweepEgressOnlyInternetGateways,
	})

	resource.AddTestSweepers("aws_eip", &resource.Sweeper{
		Name: "aws_eip",
		Dependencies: []string{
			"aws_vpc",
		},
		F: sweepEIPs,
	})

	resource.AddTestSweepers("aws_flow_log", &resource.Sweeper{
		Name: "aws_flow_log",
		F:    sweepFlowLogs,
	})

	resource.AddTestSweepers("aws_ec2_host", &resource.Sweeper{
		Name: "aws_ec2_host",
		F:    sweepHosts,
		Dependencies: []string{
			"aws_instance",
		},
	})

	resource.AddTestSweepers("aws_instance", &resource.Sweeper{
		Name: "aws_instance",
		F:    sweepInstances,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_spot_fleet_request",
		},
	})

	resource.AddTestSweepers("aws_internet_gateway", &resource.Sweeper{
		Name: "aws_internet_gateway",
		Dependencies: []string{
			"aws_subnet",
		},
		F: sweepInternetGateways,
	})

	resource.AddTestSweepers("aws_key_pair", &resource.Sweeper{
		Name: "aws_key_pair",
		Dependencies: []string{
			"aws_elastic_beanstalk_environment",
			"aws_instance",
			"aws_spot_fleet_request",
		},
		F: sweepKeyPairs,
	})

	resource.AddTestSweepers("aws_launch_template", &resource.Sweeper{
		Name: "aws_launch_template",
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_batch_compute_environment",
		},
		F: sweepLaunchTemplates,
	})

	resource.AddTestSweepers("aws_nat_gateway", &resource.Sweeper{
		Name: "aws_nat_gateway",
		F:    sweepNatGateways,
	})

	resource.AddTestSweepers("aws_network_acl", &resource.Sweeper{
		Name: "aws_network_acl",
		F:    sweepNetworkACLs,
	})

	resource.AddTestSweepers("aws_network_interface", &resource.Sweeper{
		Name: "aws_network_interface",
		F:    sweepNetworkInterfaces,
		Dependencies: []string{
			"aws_instance",
		},
	})

	resource.AddTestSweepers("aws_placement_group", &resource.Sweeper{
		Name: "aws_placement_group",
		F:    sweepPlacementGroups,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_instance",
			"aws_launch_template",
			"aws_spot_fleet_request",
		},
	})

	resource.AddTestSweepers("aws_route_table", &resource.Sweeper{
		Name: "aws_route_table",
		F:    sweepRouteTables,
	})

	resource.AddTestSweepers("aws_security_group", &resource.Sweeper{
		Name: "aws_security_group",
		Dependencies: []string{
			"aws_subnet",
		},
		F: sweepSecurityGroups,
	})

	resource.AddTestSweepers("aws_spot_fleet_request", &resource.Sweeper{
		Name: "aws_spot_fleet_request",
		F:    sweepSpotFleetRequests,
	})

	resource.AddTestSweepers("aws_subnet", &resource.Sweeper{
		Name: "aws_subnet",
		F:    sweepSubnets,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_batch_compute_environment",
			"aws_elastic_beanstalk_environment",
			"aws_cloudhsm_v2_cluster",
			"aws_db_subnet_group",
			"aws_directory_service_directory",
			"aws_dms_replication_instance",
			"aws_ec2_client_vpn_endpoint",
			"aws_ec2_transit_gateway_vpc_attachment",
			"aws_efs_file_system",
			"aws_eks_cluster",
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
			"aws_elasticsearch_domain",
			"aws_elb",
			"aws_emr_cluster",
			"aws_emr_studio",
			"aws_fsx_lustre_file_system",
			"aws_fsx_ontap_file_system",
			"aws_fsx_windows_file_system",
			"aws_lambda_function",
			"aws_lb",
			"aws_mq_broker",
			"aws_msk_cluster",
			"aws_network_interface",
			"aws_networkfirewall_firewall",
			"aws_redshift_cluster",
			"aws_route53_resolver_endpoint",
			"aws_sagemaker_notebook_instance",
			"aws_spot_fleet_request",
			"aws_vpc_endpoint",
		},
	})

	resource.AddTestSweepers("aws_ec2_transit_gateway_peering_attachment", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_peering_attachment",
		F:    sweepTransitGatewayPeeringAttachments,
	})

	resource.AddTestSweepers("aws_ec2_transit_gateway", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway",
		F:    sweepTransitGateways,
		Dependencies: []string{
			"aws_dx_gateway_association",
			"aws_ec2_transit_gateway_vpc_attachment",
			"aws_ec2_transit_gateway_peering_attachment",
			"aws_vpn_connection",
		},
	})

	resource.AddTestSweepers("aws_ec2_transit_gateway_vpc_attachment", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_vpc_attachment",
		F:    sweepTransitGatewayVPCAttachments,
	})

	resource.AddTestSweepers("aws_vpc_dhcp_options", &resource.Sweeper{
		Name: "aws_vpc_dhcp_options",
		F:    sweepVPCDHCPOptions,
	})

	resource.AddTestSweepers("aws_vpc_endpoint_service", &resource.Sweeper{
		Name: "aws_vpc_endpoint_service",
		F:    sweepVPCEndpointServices,
		Dependencies: []string{
			"aws_vpc_endpoint",
		},
	})

	resource.AddTestSweepers("aws_vpc_endpoint", &resource.Sweeper{
		Name: "aws_vpc_endpoint",
		F:    sweepVPCEndpoints,
		Dependencies: []string{
			"aws_route_table",
		},
	})

	resource.AddTestSweepers("aws_vpc_peering_connection", &resource.Sweeper{
		Name: "aws_vpc_peering_connection",
		F:    sweepVPCPeeringConnections,
	})

	resource.AddTestSweepers("aws_vpc", &resource.Sweeper{
		Name: "aws_vpc",
		Dependencies: []string{
			"aws_ec2_carrier_gateway",
			"aws_egress_only_internet_gateway",
			"aws_internet_gateway",
			"aws_nat_gateway",
			"aws_network_acl",
			"aws_route_table",
			"aws_security_group",
			"aws_subnet",
			"aws_vpc_peering_connection",
			"aws_vpn_gateway",
		},
		F: sweepVPCs,
	})

	resource.AddTestSweepers("aws_vpn_connection", &resource.Sweeper{
		Name: "aws_vpn_connection",
		F:    sweepVPNConnections,
	})

	resource.AddTestSweepers("aws_vpn_gateway", &resource.Sweeper{
		Name: "aws_vpn_gateway",
		F:    sweepVPNGateways,
		Dependencies: []string{
			"aws_dx_gateway_association",
		},
	})
}

func sweepCapacityReservations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	resp, err := conn.DescribeCapacityReservations(&ec2.DescribeCapacityReservationsInput{})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Capacity Reservation sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving EC2 Capacity Reservations: %s", err)
	}

	if len(resp.CapacityReservations) == 0 {
		log.Print("[DEBUG] No EC2 Capacity Reservations to sweep")
		return nil
	}

	for _, r := range resp.CapacityReservations {
		if aws.StringValue(r.State) != ec2.CapacityReservationStateCancelled && aws.StringValue(r.State) != ec2.CapacityReservationStateExpired {
			id := aws.StringValue(r.CapacityReservationId)

			log.Printf("[INFO] Cancelling EC2 Capacity Reservation EC2 Instance: %s", id)

			opts := &ec2.CancelCapacityReservationInput{
				CapacityReservationId: aws.String(id),
			}

			_, err := conn.CancelCapacityReservation(opts)

			if err != nil {
				log.Printf("[ERROR] Error cancelling EC2 Capacity Reservation (%s): %s", id, err)
			}
		}
	}

	return nil
}

func sweepCarrierGateway(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeCarrierGatewaysInput{}
	var sweeperErrs *multierror.Error

	err = conn.DescribeCarrierGatewaysPages(input, func(page *ec2.DescribeCarrierGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, carrierGateway := range page.CarrierGateways {
			r := ResourceCarrierGateway()
			d := r.Data(nil)
			d.SetId(aws.StringValue(carrierGateway.CarrierGatewayId))
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Carrier Gateway sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Carrier Gateways: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClientVPNEndpoints(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeClientVpnEndpointsInput{}
	err = conn.DescribeClientVpnEndpointsPages(input, func(page *ec2.DescribeClientVpnEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, clientVpnEndpoint := range page.ClientVpnEndpoints {
			id := aws.StringValue(clientVpnEndpoint.ClientVpnEndpointId)
			log.Printf("[INFO] Deleting Client VPN endpoint: %s", id)
			err := DeleteClientVPNEndpoint(conn, id)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Client VPN endpoint (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Client VPN endpoint sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Client VPN endpoints: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClientVPNNetworkAssociations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeClientVpnEndpointsInput{}
	err = conn.DescribeClientVpnEndpointsPages(input, func(page *ec2.DescribeClientVpnEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, clientVpnEndpoint := range page.ClientVpnEndpoints {

			input := &ec2.DescribeClientVpnTargetNetworksInput{
				ClientVpnEndpointId: clientVpnEndpoint.ClientVpnEndpointId,
			}
			err := conn.DescribeClientVpnTargetNetworksPages(input, func(page *ec2.DescribeClientVpnTargetNetworksOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, networkAssociation := range page.ClientVpnTargetNetworks {
					networkAssociationID := aws.StringValue(networkAssociation.AssociationId)
					clientVpnEndpointID := aws.StringValue(networkAssociation.ClientVpnEndpointId)

					log.Printf("[INFO] Deleting Client VPN network association (%s,%s)", clientVpnEndpointID, networkAssociationID)
					err := DeleteClientVPNNetworkAssociation(conn, networkAssociationID, clientVpnEndpointID)

					if err != nil {
						sweeperErr := fmt.Errorf("error deleting Client VPN network association (%s,%s): %w", clientVpnEndpointID, networkAssociationID, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					}
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Client VPN network association sweeper for %q: %s", region, err)
				return false
			}
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Client VPN network associations: %w", err))
				return false
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Client VPN network association sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Client VPN network associations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepEBSVolumes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	err = conn.DescribeVolumesPages(&ec2.DescribeVolumesInput{}, func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
		for _, volume := range page.Volumes {
			id := aws.StringValue(volume.VolumeId)

			if aws.StringValue(volume.State) != ec2.VolumeStateAvailable {
				log.Printf("[INFO] Skipping unavailable EC2 EBS Volume: %s", id)
				continue
			}

			input := &ec2.DeleteVolumeInput{
				VolumeId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 EBS Volume: %s", id)
			_, err := conn.DeleteVolume(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 EBS Volume (%s): %s", id, err)
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 EBS Volume sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving EC2 EBS Volumes: %s", err)
	}

	return nil
}

func sweepEgressOnlyInternetGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{}
	err = conn.DescribeEgressOnlyInternetGatewaysPages(input, func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gateway := range page.EgressOnlyInternetGateways {
			id := aws.StringValue(gateway.EgressOnlyInternetGatewayId)
			input := &ec2.DeleteEgressOnlyInternetGatewayInput{
				EgressOnlyInternetGatewayId: gateway.EgressOnlyInternetGatewayId,
			}

			log.Printf("[INFO] Deleting EC2 Egress Only Internet Gateway: %s", id)

			_, err := conn.DeleteEgressOnlyInternetGateway(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 Egress Only Internet Gateway (%s): %s", id, err)
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Egress Only Internet Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing EC2 Egress Only Internet Gateways: %s", err)
	}

	return nil
}

func sweepEIPs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn

	// There is currently no paginator or Marker/NextToken
	input := &ec2.DescribeAddressesInput{}

	output, err := conn.DescribeAddresses(input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 EIP sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing EC2 EIPs: %s", err)
	}

	if output == nil || len(output.Addresses) == 0 {
		log.Print("[DEBUG] No EC2 EIPs to sweep")
		return nil
	}

	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	for _, address := range output.Addresses {
		publicIP := aws.StringValue(address.PublicIp)

		if address.AssociationId != nil {
			log.Printf("[INFO] Skipping EC2 EIP (%s) with association: %s", publicIP, aws.StringValue(address.AssociationId))
			continue
		}

		r := ResourceEIP()
		d := r.Data(nil)
		if address.AllocationId != nil {
			d.SetId(aws.StringValue(address.AllocationId))
		} else {
			d.SetId(aws.StringValue(address.PublicIp))
		}

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 EIPs for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping EC2 EIP sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepFlowLogs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeFlowLogsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeFlowLogsPages(input, func(page *ec2.DescribeFlowLogsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, flowLog := range page.FlowLogs {
			r := ResourceFlowLog()
			d := r.Data(nil)
			d.SetId(aws.StringValue(flowLog.FlowLogId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Flow Log sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Flow Logs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Flow Logs (%s): %w", region, err)
	}

	return nil
}

func sweepHosts(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeHostsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeHostsPages(input, func(page *ec2.DescribeHostsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, host := range page.Hosts {
			r := ResourceHost()
			d := r.Data(nil)
			d.SetId(aws.StringValue(host.HostId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Host sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Hosts (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Hosts (%s): %w", region, err)
	}

	return nil
}

func sweepInstances(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeInstancesPages(&ec2.DescribeInstancesInput{}, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, reservation := range page.Reservations {
			if reservation == nil {
				continue
			}

			for _, instance := range reservation.Instances {
				id := aws.StringValue(instance.InstanceId)

				if instance.State != nil && aws.StringValue(instance.State.Name) == ec2.InstanceStateNameTerminated {
					log.Printf("[INFO] Skipping terminated EC2 Instance: %s", id)
					continue
				}

				r := ResourceInstance()
				d := r.Data(nil)
				d.SetId(id)
				d.Set("disable_api_termination", false)

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing EC2 Instances for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 Instances for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping EC2 Instance sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepInternetGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	defaultVPCID := ""
	describeVpcsInput := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: aws.StringSlice([]string{"true"}),
			},
		},
	}

	describeVpcsOutput, err := conn.DescribeVpcs(describeVpcsInput)

	if err != nil {
		return fmt.Errorf("error describing VPCs: %w", err)
	}

	if describeVpcsOutput != nil && len(describeVpcsOutput.Vpcs) == 1 {
		defaultVPCID = aws.StringValue(describeVpcsOutput.Vpcs[0].VpcId)
	}

	input := &ec2.DescribeInternetGatewaysInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeInternetGatewaysPages(input, func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, internetGateway := range page.InternetGateways {
			internetGatewayID := aws.StringValue(internetGateway.InternetGatewayId)
			isDefaultVPCInternetGateway := false

			for _, attachment := range internetGateway.Attachments {
				if aws.StringValue(attachment.VpcId) == defaultVPCID {
					isDefaultVPCInternetGateway = true
					break
				}
			}

			if isDefaultVPCInternetGateway {
				log.Printf("[DEBUG] Skipping Default VPC Internet Gateway: %s", internetGatewayID)
				continue
			}

			r := ResourceInternetGateway()
			d := r.Data(nil)
			d.SetId(internetGatewayID)
			if len(internetGateway.Attachments) > 0 {
				d.Set("vpc_id", internetGateway.Attachments[0].VpcId)
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Internet Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Internet Gateways (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Internet Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepKeyPairs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	log.Printf("Destroying the tmp keys in (%s)", client.(*conns.AWSClient).Region)

	resp, err := conn.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Key Pair sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing key pairs in Sweeper: %s", err)
	}

	keyPairs := resp.KeyPairs
	for _, d := range keyPairs {
		_, err := conn.DeleteKeyPair(&ec2.DeleteKeyPairInput{
			KeyName: d.KeyName,
		})

		if err != nil {
			return fmt.Errorf("Error deleting key pairs in Sweeper: %s", err)
		}
	}
	return nil
}

func sweepLaunchTemplates(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeLaunchTemplatesInput{}
	var sweeperErrs *multierror.Error

	err = conn.DescribeLaunchTemplatesPages(input, func(page *ec2.DescribeLaunchTemplatesOutput, lastPage bool) bool {
		for _, launchTemplate := range page.LaunchTemplates {
			id := aws.StringValue(launchTemplate.LaunchTemplateId)
			input := &ec2.DeleteLaunchTemplateInput{
				LaunchTemplateId: launchTemplate.LaunchTemplateId,
			}

			log.Printf("[INFO] Deleting EC2 Launch Template: %s", id)
			_, err := conn.DeleteLaunchTemplate(input)

			if tfawserr.ErrMessageContains(err, "InvalidLaunchTemplateId.NotFound", "") {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting EC2 Launch Template (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Launch Template sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing EC2 Launch Templates: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepNatGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	req := &ec2.DescribeNatGatewaysInput{}
	resp, err := conn.DescribeNatGateways(req)
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 NAT Gateway sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing NAT Gateways: %s", err)
	}

	if len(resp.NatGateways) == 0 {
		log.Print("[DEBUG] No AWS NAT Gateways to sweep")
		return nil
	}

	for _, natGateway := range resp.NatGateways {
		_, err := conn.DeleteNatGateway(&ec2.DeleteNatGatewayInput{
			NatGatewayId: natGateway.NatGatewayId,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting NAT Gateway (%s): %s",
				*natGateway.NatGatewayId, err)
		}
	}

	return nil
}

func sweepNetworkACLs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	req := &ec2.DescribeNetworkAclsInput{}
	resp, err := conn.DescribeNetworkAcls(req)
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Network ACL sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing Network ACLs: %s", err)
	}

	if len(resp.NetworkAcls) == 0 {
		log.Print("[DEBUG] No Network ACLs to sweep")
		return nil
	}

	for _, nacl := range resp.NetworkAcls {
		// Delete rules first
		for _, entry := range nacl.Entries {
			// These are the rule numbers for IPv4 and IPv6 "ALL traffic" rules which cannot be deleted
			if aws.Int64Value(entry.RuleNumber) == 32767 || aws.Int64Value(entry.RuleNumber) == 32768 {
				log.Printf("[DEBUG] Skipping Network ACL rule: %q / %d", *nacl.NetworkAclId, *entry.RuleNumber)
				continue
			}

			log.Printf("[INFO] Deleting Network ACL rule: %q / %d", *nacl.NetworkAclId, *entry.RuleNumber)
			_, err := conn.DeleteNetworkAclEntry(&ec2.DeleteNetworkAclEntryInput{
				NetworkAclId: nacl.NetworkAclId,
				Egress:       entry.Egress,
				RuleNumber:   entry.RuleNumber,
			})
			if err != nil {
				return fmt.Errorf(
					"Error deleting Network ACL rule (%s / %d): %s",
					*nacl.NetworkAclId, *entry.RuleNumber, err)
			}
		}

		// Disassociate subnets
		log.Printf("[DEBUG] Found %d Network ACL associations for %q", len(nacl.Associations), *nacl.NetworkAclId)
		for _, a := range nacl.Associations {
			log.Printf("[DEBUG] Replacing subnet associations for Network ACL %q", *nacl.NetworkAclId)
			defaultAcl, err := GetDefaultNetworkACL(*nacl.VpcId, conn)
			if err != nil {
				return fmt.Errorf("Failed to find default Network ACL for VPC %q", *nacl.VpcId)
			}
			_, err = conn.ReplaceNetworkAclAssociation(&ec2.ReplaceNetworkAclAssociationInput{
				NetworkAclId:  defaultAcl.NetworkAclId,
				AssociationId: a.NetworkAclAssociationId,
			})
			if err != nil {
				return fmt.Errorf("Failed to replace subnet association for Network ACL %q: %s",
					*nacl.NetworkAclId, err)
			}
		}

		// Default Network ACLs will be deleted along with VPC
		if *nacl.IsDefault {
			log.Printf("[DEBUG] Skipping default Network ACL: %q", *nacl.NetworkAclId)
			continue
		}

		log.Printf("[INFO] Deleting Network ACL: %q", *nacl.NetworkAclId)
		_, err := conn.DeleteNetworkAcl(&ec2.DeleteNetworkAclInput{
			NetworkAclId: nacl.NetworkAclId,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting Network ACL (%s): %s",
				*nacl.NetworkAclId, err)
		}
	}

	return nil
}

func sweepNetworkInterfaces(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	err = conn.DescribeNetworkInterfacesPages(&ec2.DescribeNetworkInterfacesInput{}, func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, networkInterface := range page.NetworkInterfaces {
			id := aws.StringValue(networkInterface.NetworkInterfaceId)

			if aws.StringValue(networkInterface.Status) != ec2.NetworkInterfaceStatusAvailable {
				log.Printf("[INFO] Skipping EC2 Network Interface in unavailable (%s) status: %s", aws.StringValue(networkInterface.Status), id)
				continue
			}

			input := &ec2.DeleteNetworkInterfaceInput{
				NetworkInterfaceId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 Network Interface: %s", id)
			_, err := conn.DeleteNetworkInterface(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 Network Interface (%s): %s", id, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Network Interface sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving EC2 Network Interfaces: %s", err)
	}

	return nil
}

func sweepPlacementGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribePlacementGroupsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribePlacementGroups(input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Placement Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Placement Groups (%s): %w", region, err)
	}

	for _, placementGroup := range output.PlacementGroups {
		r := ResourcePlacementGroup()
		d := r.Data(nil)
		d.SetId(aws.StringValue(placementGroup.GroupName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Placement Groups (%s): %w", region, err)
	}

	return nil
}

func sweepRouteTables(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeRouteTablesInput{}

	err = conn.DescribeRouteTablesPages(input, func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, routeTable := range page.RouteTables {
			if routeTable == nil {
				continue
			}

			id := aws.StringValue(routeTable.RouteTableId)
			isMainRouteTableAssociation := false

			for _, routeTableAssociation := range routeTable.Associations {
				if routeTableAssociation == nil {
					continue
				}

				if aws.BoolValue(routeTableAssociation.Main) {
					isMainRouteTableAssociation = true
					break
				}

				associationID := aws.StringValue(routeTableAssociation.RouteTableAssociationId)

				input := &ec2.DisassociateRouteTableInput{
					AssociationId: routeTableAssociation.RouteTableAssociationId,
				}

				log.Printf("[DEBUG] Deleting EC2 Route Table Association: %s", associationID)
				_, err := conn.DisassociateRouteTable(input)

				if err != nil {
					sweeperErr := fmt.Errorf("error deleting EC2 Route Table (%s) Association (%s): %w", id, associationID, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}

			if isMainRouteTableAssociation {
				for _, route := range routeTable.Routes {
					if route == nil {
						continue
					}

					if aws.StringValue(route.GatewayId) == "local" {
						continue
					}

					// Prevent deleting default VPC route for Internet Gateway
					// which some testing is still reliant on operating correctly
					if strings.HasPrefix(aws.StringValue(route.GatewayId), "igw-") && aws.StringValue(route.DestinationCidrBlock) == "0.0.0.0/0" {
						continue
					}

					input := &ec2.DeleteRouteInput{
						DestinationCidrBlock:     route.DestinationCidrBlock,
						DestinationIpv6CidrBlock: route.DestinationIpv6CidrBlock,
						RouteTableId:             routeTable.RouteTableId,
					}

					log.Printf("[DEBUG] Deleting EC2 Route Table (%s) Route", id)
					_, err := conn.DeleteRoute(input)

					if err != nil {
						sweeperErr := fmt.Errorf("error deleting EC2 Route Table (%s) Route: %w", id, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
						continue
					}
				}

				continue
			}

			input := &ec2.DeleteRouteTableInput{
				RouteTableId: routeTable.RouteTableId,
			}

			log.Printf("[DEBUG] Deleting EC2 Route Table: %s", id)
			_, err := conn.DeleteRouteTable(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting EC2 Route Table (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Route Table sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Route Tables: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSecurityGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSecurityGroupsInput{}

	// Delete all non-default EC2 Security Group Rules to prevent DependencyViolation errors
	err = conn.DescribeSecurityGroupsPages(input, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		for _, sg := range page.SecurityGroups {
			if aws.StringValue(sg.GroupName) == "default" {
				log.Printf("[DEBUG] Skipping default EC2 Security Group: %s", aws.StringValue(sg.GroupId))
				continue
			}

			if sg.IpPermissions != nil {
				req := &ec2.RevokeSecurityGroupIngressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissions,
				}

				if _, err = conn.RevokeSecurityGroupIngress(req); err != nil {
					log.Printf("[ERROR] Error revoking ingress rule for Security Group (%s): %s", aws.StringValue(sg.GroupId), err)
				}
			}

			if sg.IpPermissionsEgress != nil {
				req := &ec2.RevokeSecurityGroupEgressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissionsEgress,
				}

				if _, err = conn.RevokeSecurityGroupEgress(req); err != nil {
					log.Printf("[ERROR] Error revoking egress rule for Security Group (%s): %s", aws.StringValue(sg.GroupId), err)
				}
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Security Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving EC2 Security Groups: %w", err)
	}

	err = conn.DescribeSecurityGroupsPages(input, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		for _, sg := range page.SecurityGroups {
			if aws.StringValue(sg.GroupName) == "default" {
				log.Printf("[DEBUG] Skipping default EC2 Security Group: %s", aws.StringValue(sg.GroupId))
				continue
			}

			input := &ec2.DeleteSecurityGroupInput{
				GroupId: sg.GroupId,
			}

			// Handle EC2 eventual consistency
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := conn.DeleteSecurityGroup(input)

				if tfawserr.ErrMessageContains(err, "DependencyViolation", "") {
					return resource.RetryableError(err)
				}
				if err != nil {
					return resource.NonRetryableError(err)
				}
				return nil
			})

			if err != nil {
				log.Printf("[ERROR] Error deleting Security Group (%s): %s", aws.StringValue(sg.GroupId), err)
			}
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Error retrieving EC2 Security Groups: %w", err)
	}

	return nil
}

func sweepSpotFleetRequests(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeSpotFleetRequestsPages(&ec2.DescribeSpotFleetRequestsInput{}, func(page *ec2.DescribeSpotFleetRequestsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		if len(page.SpotFleetRequestConfigs) == 0 {
			log.Print("[DEBUG] No Spot Fleet Requests to sweep")
			return false
		}

		for _, config := range page.SpotFleetRequestConfigs {
			id := aws.StringValue(config.SpotFleetRequestId)

			r := ResourceSpotFleetRequest()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("terminate_instances_with_expiration", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing EC2 Spot Fleet Requests for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 Spot Fleet Requests for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping EC2 Spot Fleet Requests sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepSubnets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &ec2.DescribeSubnetsInput{}

	err = conn.DescribeSubnetsPages(input, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, subnet := range page.Subnets {
			if subnet == nil {
				continue
			}

			id := aws.StringValue(subnet.SubnetId)

			if aws.BoolValue(subnet.DefaultForAz) {
				log.Printf("[DEBUG] Skipping default EC2 Subnet: %s", id)
				continue
			}

			r := ResourceSubnet()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing EC2 Subnets for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 Subnets for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping EC2 Subnet sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepTransitGatewayPeeringAttachments(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{}
	var sweeperErrs *multierror.Error

	err = conn.DescribeTransitGatewayPeeringAttachmentsPages(input,
		func(page *ec2.DescribeTransitGatewayPeeringAttachmentsOutput, lastPage bool) bool {
			for _, transitGatewayPeeringAttachment := range page.TransitGatewayPeeringAttachments {
				if aws.StringValue(transitGatewayPeeringAttachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
					continue
				}

				id := aws.StringValue(transitGatewayPeeringAttachment.TransitGatewayAttachmentId)

				input := &ec2.DeleteTransitGatewayPeeringAttachmentInput{
					TransitGatewayAttachmentId: aws.String(id),
				}

				log.Printf("[INFO] Deleting EC2 Transit Gateway Peering Attachment: %s", id)
				_, err := conn.DeleteTransitGatewayPeeringAttachment(input)

				if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
					continue
				}

				if err != nil {
					sweeperErr := fmt.Errorf("error deleting EC2 Transit Gateway Peering Attachment (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				if err := WaitForTransitGatewayPeeringAttachmentDeletion(conn, id); err != nil {
					sweeperErr := fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) deletion: %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}
			return !lastPage
		})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Transit Gateway Peering Attachment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving EC2 Transit Gateway Peering Attachments: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTransitGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeTransitGatewaysInput{}

	for {
		output, err := conn.DescribeTransitGateways(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving EC2 Transit Gateways: %s", err)
		}

		for _, transitGateway := range output.TransitGateways {
			if aws.StringValue(transitGateway.State) == ec2.TransitGatewayStateDeleted {
				continue
			}

			id := aws.StringValue(transitGateway.TransitGatewayId)

			input := &ec2.DeleteTransitGatewayInput{
				TransitGatewayId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 Transit Gateway: %s", id)
			err := resource.Retry(2*time.Minute, func() *resource.RetryError {
				_, err := conn.DeleteTransitGateway(input)

				if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted Transit Gateway Attachments") {
					return resource.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted DirectConnect Gateway Attachments") {
					return resource.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted VPN Attachments") {
					return resource.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayID.NotFound", "") {
					return nil
				}

				if err != nil {
					return resource.NonRetryableError(err)
				}

				return nil
			})

			if tfresource.TimedOut(err) {
				_, err = conn.DeleteTransitGateway(input)
			}

			if err != nil {
				return fmt.Errorf("error deleting EC2 Transit Gateway (%s): %s", id, err)
			}

			if err := WaitForTransitGatewayDeletion(conn, id); err != nil {
				return fmt.Errorf("error waiting for EC2 Transit Gateway (%s) deletion: %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepTransitGatewayVPCAttachments(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeTransitGatewayAttachmentsInput{}

	for {
		output, err := conn.DescribeTransitGatewayAttachments(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway VPC Attachment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving EC2 Transit Gateway VPC Attachments: %s", err)
		}

		for _, attachment := range output.TransitGatewayAttachments {
			if aws.StringValue(attachment.ResourceType) != ec2.TransitGatewayAttachmentResourceTypeVpc {
				continue
			}

			if aws.StringValue(attachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
				continue
			}

			id := aws.StringValue(attachment.TransitGatewayAttachmentId)

			input := &ec2.DeleteTransitGatewayVpcAttachmentInput{
				TransitGatewayAttachmentId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 Transit Gateway VPC Attachment: %s", id)
			_, err := conn.DeleteTransitGatewayVpcAttachment(input)

			if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
				continue
			}

			if err != nil {
				return fmt.Errorf("error deleting EC2 Transit Gateway VPC Attachment (%s): %s", id, err)
			}

			if err := WaitForTransitGatewayVPCAttachmentDeletion(conn, id); err != nil {
				return fmt.Errorf("error waiting for EC2 Transit Gateway VPC Attachment (%s) deletion: %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepVPCDHCPOptions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeDhcpOptionsInput{}

	err = conn.DescribeDhcpOptionsPages(input, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
		for _, dhcpOption := range page.DhcpOptions {
			var defaultDomainNameFound, defaultDomainNameServersFound bool

			// This skips the default dhcp configurations so they don't get deleted
			for _, dhcpConfiguration := range dhcpOption.DhcpConfigurations {
				if aws.StringValue(dhcpConfiguration.Key) == "domain-name" {
					if len(dhcpConfiguration.Values) != 1 || dhcpConfiguration.Values[0] == nil {
						continue
					}

					if aws.StringValue(dhcpConfiguration.Values[0].Value) == RegionalPrivateDNSSuffix(region) {
						defaultDomainNameFound = true
					}
				} else if aws.StringValue(dhcpConfiguration.Key) == "domain-name-servers" {
					if len(dhcpConfiguration.Values) != 1 || dhcpConfiguration.Values[0] == nil {
						continue
					}

					if aws.StringValue(dhcpConfiguration.Values[0].Value) == "AmazonProvidedDNS" {
						defaultDomainNameServersFound = true
					}
				}
			}

			if defaultDomainNameFound && defaultDomainNameServersFound {
				continue
			}

			input := &ec2.DeleteDhcpOptionsInput{
				DhcpOptionsId: dhcpOption.DhcpOptionsId,
			}

			_, err := conn.DeleteDhcpOptions(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 DHCP Option (%s): %s", aws.StringValue(dhcpOption.DhcpOptionsId), err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 DHCP Option sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing DHCP Options: %s", err)
	}

	return nil
}

func sweepVPCEndpointServices(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{}

	err = conn.DescribeVpcEndpointServiceConfigurationsPages(input, func(page *ec2.DescribeVpcEndpointServiceConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, serviceConfiguration := range page.ServiceConfigurations {
			if serviceConfiguration == nil {
				continue
			}

			if aws.StringValue(serviceConfiguration.ServiceState) == ec2.ServiceStateDeleted {
				continue
			}

			id := aws.StringValue(serviceConfiguration.ServiceId)

			log.Printf("[INFO] Deleting EC2 VPC Endpoint Service: %s", id)

			r := ResourceVPCEndpointService()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting EC2 VPC Endpoint Service (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPC Endpoint Service sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 VPC Endpoint Services: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVPCEndpoints(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeVpcEndpointsInput{}

	err = conn.DescribeVpcEndpointsPages(input, func(page *ec2.DescribeVpcEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vpcEndpoint := range page.VpcEndpoints {
			if vpcEndpoint == nil {
				continue
			}

			if aws.StringValue(vpcEndpoint.State) != "available" {
				continue
			}

			id := aws.StringValue(vpcEndpoint.VpcEndpointId)

			log.Printf("[INFO] Deleting EC2 VPC Endpoint: %s", id)

			r := ResourceVPCEndpoint()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting EC2 VPC Endpoint (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPC Endpoint sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 VPC Endpoints: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVPCPeeringConnections(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeVpcPeeringConnectionsInput{}

	err = conn.DescribeVpcPeeringConnectionsPages(input, func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vpcPeeringConnection := range page.VpcPeeringConnections {
			deletedStatuses := map[string]bool{
				ec2.VpcPeeringConnectionStateReasonCodeDeleted:  true,
				ec2.VpcPeeringConnectionStateReasonCodeExpired:  true,
				ec2.VpcPeeringConnectionStateReasonCodeFailed:   true,
				ec2.VpcPeeringConnectionStateReasonCodeRejected: true,
			}

			if _, ok := deletedStatuses[aws.StringValue(vpcPeeringConnection.Status.Code)]; ok {
				continue
			}

			id := aws.StringValue(vpcPeeringConnection.VpcPeeringConnectionId)
			input := &ec2.DeleteVpcPeeringConnectionInput{
				VpcPeeringConnectionId: vpcPeeringConnection.VpcPeeringConnectionId,
			}

			log.Printf("[INFO] Deleting EC2 VPC Peering Connection: %s", id)

			_, err := conn.DeleteVpcPeeringConnection(input)

			if tfawserr.ErrMessageContains(err, "InvalidVpcPeeringConnectionID.NotFound", "") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 VPC Peering Connection (%s): %s", id, err)
				continue
			}

			if err := WaitForVPCPeeringConnectionDeletion(conn, id, 5*time.Minute); err != nil { //nolint:gomnd
				log.Printf("[ERROR] Error waiting for EC2 VPC Peering Connection (%s) to be deleted: %s", id, err)
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPC Peering Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing EC2 VPC Peering Connections: %s", err)
	}

	return nil
}

func sweepVPCs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &ec2.DescribeVpcsInput{}

	err = conn.DescribeVpcsPages(input, func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vpc := range page.Vpcs {
			if vpc == nil {
				continue
			}

			id := aws.StringValue(vpc.VpcId)

			if aws.BoolValue(vpc.IsDefault) {
				log.Printf("[DEBUG] Skipping default EC2 VPC: %s", id)
				continue
			}

			r := ResourceVPC()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing EC2 VPCs for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 VPCs for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping EC2 VPCs sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepVPNConnections(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeVpnConnectionsInput{}

	// DescribeVpnConnections does not currently have any form of pagination
	output, err := conn.DescribeVpnConnections(input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPN Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving EC2 VPN Connections: %s", err)
	}

	for _, vpnConnection := range output.VpnConnections {
		if aws.StringValue(vpnConnection.State) == ec2.VpnStateDeleted {
			continue
		}

		id := aws.StringValue(vpnConnection.VpnConnectionId)
		input := &ec2.DeleteVpnConnectionInput{
			VpnConnectionId: vpnConnection.VpnConnectionId,
		}

		log.Printf("[INFO] Deleting EC2 VPN Connection: %s", id)

		_, err := conn.DeleteVpnConnection(input)

		if tfawserr.ErrMessageContains(err, "InvalidVpnConnectionID.NotFound", "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error deleting EC2 VPN Connection (%s): %s", id, err)
		}

		if err := WaitForVPNConnectionDeletion(conn, id); err != nil {
			return fmt.Errorf("error waiting for VPN connection (%s) to delete: %s", id, err)
		}
	}

	return nil
}

func sweepVPNGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	var sweeperErrs *multierror.Error

	req := &ec2.DescribeVpnGatewaysInput{}
	resp, err := conn.DescribeVpnGateways(req)
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPN Gateway sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing VPN Gateways: %s", err)
	}

	if len(resp.VpnGateways) == 0 {
		log.Print("[DEBUG] No VPN Gateways to sweep")
		return nil
	}

	for _, vpng := range resp.VpnGateways {
		if aws.StringValue(vpng.State) == ec2.VpnStateDeleted {
			continue
		}

		for _, vpcAttachment := range vpng.VpcAttachments {
			if aws.StringValue(vpcAttachment.State) == ec2.AttachmentStatusDetached {
				continue
			}

			r := ResourceVPNGatewayAttachment()
			d := r.Data(nil)
			d.Set("vpc_id", vpcAttachment.VpcId)
			d.Set("vpn_gateway_id", vpng.VpnGatewayId)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		r := ResourceVPNGateway()
		d := r.Data(nil)
		d.SetId(aws.StringValue(vpng.VpnGatewayId))
		err := r.Delete(d, client)

		if err != nil {
			log.Printf("[ERROR] %s", err)
			sweeperErrs = multierror.Append(sweeperErrs, err)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}
