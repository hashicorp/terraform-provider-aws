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
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_customer_gateway", &resource.Sweeper{
		Name: "aws_customer_gateway",
		F:    sweepCustomerGateways,
		Dependencies: []string{
			"aws_vpn_connection",
		},
	})

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

	resource.AddTestSweepers("aws_ebs_snapshot", &resource.Sweeper{
		Name: "aws_ebs_snapshot",
		F:    sweepEBSSnapshots,
		Dependencies: []string{
			"aws_ami",
		},
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
		F:    sweepNATGateways,
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

	resource.AddTestSweepers("aws_ec2_network_insights_path", &resource.Sweeper{
		Name: "aws_ec2_network_insights_path",
		F:    sweepNetworkInsightsPaths,
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
			"aws_appstream_fleet",
			"aws_appstream_image_builder",
			"aws_autoscaling_group",
			"aws_batch_compute_environment",
			"aws_elastic_beanstalk_environment",
			"aws_cloud9_environment_ec2",
			"aws_cloudhsm_v2_cluster",
			"aws_codestarconnections_host",
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
			"aws_fsx_openzfs_file_system",
			"aws_fsx_windows_file_system",
			"aws_iot_topic_rule_destination",
			"aws_lambda_function",
			"aws_lb",
			"aws_memorydb_subnet_group",
			"aws_mq_broker",
			"aws_msk_cluster",
			"aws_network_interface",
			"aws_networkfirewall_firewall",
			"aws_opensearch_domain",
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

	resource.AddTestSweepers("aws_ec2_transit_gateway_multicast_domain", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_multicast_domain",
		F:    sweepTransitGatewayMulticastDomains,
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

	resource.AddTestSweepers("aws_ec2_transit_gateway_connect_peer", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_connect_peer",
		F:    sweepTransitGatewayConnectPeers,
	})

	resource.AddTestSweepers("aws_ec2_transit_gateway_connect", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_connect",
		F:    sweepTransitGatewayConnects,
		Dependencies: []string{
			"aws_ec2_transit_gateway_connect_peer",
		},
	})

	resource.AddTestSweepers("aws_ec2_transit_gateway_vpc_attachment", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_vpc_attachment",
		F:    sweepTransitGatewayVPCAttachments,
		Dependencies: []string{
			"aws_ec2_transit_gateway_connect",
			"aws_ec2_transit_gateway_multicast_domain",
		},
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
			"aws_vpn_connection",
		},
	})

	resource.AddTestSweepers("aws_vpc_ipam_pool_cidr", &resource.Sweeper{
		Name: "aws_vpc_ipam_pool_cidr",
		F:    sweepIPAMPoolCIDRs,
		Dependencies: []string{
			"aws_vpc",
		},
	})

	resource.AddTestSweepers("aws_vpc_ipam_pool", &resource.Sweeper{
		Name: "aws_vpc_ipam_pool",
		F:    sweepIPAMPools,
		Dependencies: []string{
			"aws_vpc_ipam_pool_cidr",
		},
	})

	resource.AddTestSweepers("aws_vpc_ipam_scope", &resource.Sweeper{
		Name: "aws_vpc_ipam_scope",
		F:    sweepIPAMScopes,
		Dependencies: []string{
			"aws_vpc_ipam_pool",
		},
	})

	resource.AddTestSweepers("aws_vpc_ipam", &resource.Sweeper{
		Name: "aws_vpc_ipam",
		F:    sweepIPAMs,
		Dependencies: []string{
			"aws_vpc_ipam_scope",
		},
	})

	resource.AddTestSweepers("aws_ami", &resource.Sweeper{
		Name: "aws_ami",
		F:    sweepAMIs,
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
	input := &ec2.DescribeClientVpnEndpointsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeClientVpnEndpointsPages(input, func(page *ec2.DescribeClientVpnEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClientVpnEndpoints {
			r := ResourceClientVPNEndpoint()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ClientVpnEndpointId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Client VPN Endpoint sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Client VPN Endpoints (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Client VPN Endpoints (%s): %w", region, err)
	}

	return nil
}

func sweepClientVPNNetworkAssociations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeClientVpnEndpointsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeClientVpnEndpointsPages(input, func(page *ec2.DescribeClientVpnEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClientVpnEndpoints {
			input := &ec2.DescribeClientVpnTargetNetworksInput{
				ClientVpnEndpointId: v.ClientVpnEndpointId,
			}

			err := conn.DescribeClientVpnTargetNetworksPages(input, func(page *ec2.DescribeClientVpnTargetNetworksOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.ClientVpnTargetNetworks {
					r := ResourceClientVPNNetworkAssociation()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.AssociationId))
					d.Set("client_vpn_endpoint_id", v.ClientVpnEndpointId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Client VPN Network Associations (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Client VPN Network Association sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Client VPN Endpoints (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EC2 Client VPN Network Associations (%s): %w", region, err))
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

func sweepEBSSnapshots(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: aws.StringSlice([]string{"self"}),
	}
	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeSnapshotsPages(input, func(page *ec2.DescribeSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Snapshots {
			r := ResourceEBSSnapshot()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.SnapshotId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EBS Snapshot sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EBS Snapshots (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EBS Snapshots (%s): %w", region, err)
	}

	return nil
}

func sweepEgressOnlyInternetGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{}
	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeEgressOnlyInternetGatewaysPages(input, func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EgressOnlyInternetGateways {
			r := ResourceEgressOnlyInternetGateway()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.EgressOnlyInternetGatewayId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Egress-only Internet Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Egress-only Internet Gateways (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Egress-only Internet Gateways (%s): %w", region, err)
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

			if tfawserr.ErrCodeEquals(err, "InvalidLaunchTemplateId.NotFound") {
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

func sweepNATGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeNatGatewaysInput{}
	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeNatGatewaysPages(input, func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NatGateways {
			r := ResourceNATGateway()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.NatGatewayId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 NAT Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 NAT Gateways (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 NAT Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepNetworkACLs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeNetworkAclsInput{}
	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeNetworkAclsPages(input, func(page *ec2.DescribeNetworkAclsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkAcls {
			if aws.BoolValue(v.IsDefault) {
				continue
			}

			r := ResourceNetworkACL()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.NetworkAclId))

			var subnetIDs []string
			for _, v := range v.Associations {
				subnetIDs = append(subnetIDs, aws.StringValue(v.SubnetId))
			}
			d.Set("subnet_ids", subnetIDs)

			d.Set("vpc_id", v.VpcId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Network ACL sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Network ACLs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Network ACLs (%s): %w", region, err)
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

func sweepNetworkInsightsPaths(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeNetworkInsightsPathsPages(&ec2.DescribeNetworkInsightsPathsInput{}, func(page *ec2.DescribeNetworkInsightsPathsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, nip := range page.NetworkInsightsPaths {
			id := aws.StringValue(nip.NetworkInsightsPathId)

			r := ResourceNetworkInsightsPath()
			d := r.Data(nil)

			d.SetId(id)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Network Insights Paths for %s: %w", region, err))
	}
	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Network Insights Paths for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Insights Path sweep for %s: %s", region, errs)
		return nil
	}
	return errs.ErrorOrNil()
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

				if tfawserr.ErrCodeEquals(err, "DependencyViolation") {
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
	input := &ec2.DescribeSubnetsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeSubnetsPages(input, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Subnets {
			// Skip default subnets.
			if aws.BoolValue(v.DefaultForAz) {
				continue
			}

			r := ResourceSubnet()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.SubnetId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Subnet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Subnets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Subnets (%s): %w", region, err)
	}

	return nil
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

				if tfawserr.ErrCodeEquals(err, "InvalidTransitGatewayAttachmentID.NotFound") {
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

func sweepTransitGatewayMulticastDomains(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeTransitGatewayMulticastDomainsPages(input, func(page *ec2.DescribeTransitGatewayMulticastDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayMulticastDomains {
			if aws.StringValue(v.State) == ec2.TransitGatewayMulticastDomainStateDeleted {
				continue
			}

			r := ResourceTransitGatewayMulticastDomain()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.TransitGatewayMulticastDomainId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Transit Gateway Multicast Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Transit Gateway Multicast Domains (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateway Multicast Domains (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeTransitGatewaysInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeTransitGatewaysPages(input, func(page *ec2.DescribeTransitGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGateways {
			if aws.StringValue(v.State) == ec2.TransitGatewayStateDeleted {
				continue
			}

			r := ResourceTransitGateway()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.TransitGatewayId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Transit Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Transit Gateways (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayConnectPeers(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeTransitGatewayConnectPeersInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeTransitGatewayConnectPeersPages(input, func(page *ec2.DescribeTransitGatewayConnectPeersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayConnectPeers {
			if aws.StringValue(v.State) == ec2.TransitGatewayConnectPeerStateDeleted {
				continue
			}

			r := ResourceTransitGatewayConnectPeer()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.TransitGatewayConnectPeerId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Transit Gateway Connect Peer sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Transit Gateway Connect Peers (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateway Connect Peers (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayConnects(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeTransitGatewayConnectsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeTransitGatewayConnectsPages(input, func(page *ec2.DescribeTransitGatewayConnectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayConnects {
			if aws.StringValue(v.State) == ec2.TransitGatewayAttachmentStateDeleted {
				continue
			}

			r := ResourceTransitGatewayConnect()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.TransitGatewayAttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Transit Gateway Connect sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Transit Gateway Connects (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateway Connects (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayVPCAttachments(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeTransitGatewayVpcAttachmentsPages(input, func(page *ec2.DescribeTransitGatewayVpcAttachmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayVpcAttachments {
			if aws.StringValue(v.State) == ec2.TransitGatewayAttachmentStateDeleted {
				continue
			}

			r := ResourceTransitGatewayVPCAttachment()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.TransitGatewayAttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Transit Gateway VPC Attachment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Transit Gateway VPC Attachments (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateway VPC Attachments (%s): %w", region, err)
	}

	return nil
}

func sweepVPCDHCPOptions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeDhcpOptionsInput{}
	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeDhcpOptionsPages(input, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DhcpOptions {
			// Skip the default DHCP Options.
			var defaultDomainNameFound, defaultDomainNameServersFound bool

			for _, v := range v.DhcpConfigurations {
				if aws.StringValue(v.Key) == "domain-name" {
					if len(v.Values) != 1 || v.Values[0] == nil {
						continue
					}

					if aws.StringValue(v.Values[0].Value) == RegionalPrivateDNSSuffix(region) {
						defaultDomainNameFound = true
					}
				} else if aws.StringValue(v.Key) == "domain-name-servers" {
					if len(v.Values) != 1 || v.Values[0] == nil {
						continue
					}

					if aws.StringValue(v.Values[0].Value) == "AmazonProvidedDNS" {
						defaultDomainNameServersFound = true
					}
				}
			}

			if defaultDomainNameFound && defaultDomainNameServersFound {
				continue
			}

			r := ResourceVPCDHCPOptions()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DhcpOptionsId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 DHCP Options Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 DHCP Options Sets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 DHCP Options Sets (%s): %w", region, err)
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
	input := &ec2.DescribeVpcPeeringConnectionsInput{}
	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeVpcPeeringConnectionsPages(input, func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcPeeringConnections {
			r := ResourceVPCPeeringConnection()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.VpcPeeringConnectionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPC Peering Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 VPC Peering Connections (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPC Peering Connections (%s): %w", region, err)
	}

	return nil
}

func sweepVPCs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeVpcsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeVpcsPages(input, func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Vpcs {
			// Skip default VPCs.
			if aws.BoolValue(v.IsDefault) {
				continue
			}

			r := ResourceVPC()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.VpcId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPC sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 VPCs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPCs (%s): %w", region, err)
	}

	return nil
}

func sweepVPNConnections(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeVpnConnectionsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeVpnConnections(input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPN Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 VPN Connections (%s): %w", region, err)
	}

	for _, v := range output.VpnConnections {
		if aws.StringValue(v.State) == ec2.VpnStateDeleted {
			continue
		}

		r := ResourceVPNConnection()
		d := r.Data(nil)
		d.SetId(aws.StringValue(v.VpnConnectionId))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPN Connections (%s): %w", region, err)
	}

	return nil
}

func sweepVPNGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeVpnGatewaysInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeVpnGateways(input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPN Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 VPN Gateways (%s): %w", region, err)
	}

	for _, v := range output.VpnGateways {
		if aws.StringValue(v.State) == ec2.VpnStateDeleted {
			continue
		}

		r := ResourceVPNGateway()
		d := r.Data(nil)
		d.SetId(aws.StringValue(v.VpnGatewayId))

		for _, v := range v.VpcAttachments {
			if aws.StringValue(v.State) != ec2.AttachmentStatusDetached {
				d.Set("vpc_id", v.VpcId)

				break
			}
		}

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPN Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepCustomerGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeCustomerGatewaysInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeCustomerGateways(input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Customer Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Customer Gateways (%s): %w", region, err)
	}

	for _, v := range output.CustomerGateways {
		if aws.StringValue(v.State) == CustomerGatewayStateDeleted {
			continue
		}

		r := ResourceCustomerGateway()
		d := r.Data(nil)
		d.SetId(aws.StringValue(v.CustomerGatewayId))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Customer Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepIPAMPoolCIDRs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeIpamPoolsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeIpamPoolsPages(input, func(page *ec2.DescribeIpamPoolsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IpamPools {
			poolID := aws.StringValue(v.IpamPoolId)
			input := &ec2.GetIpamPoolCidrsInput{
				IpamPoolId: aws.String(poolID),
			}

			err := conn.GetIpamPoolCidrsPages(input, func(page *ec2.GetIpamPoolCidrsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.IpamPoolCidrs {
					r := ResourceIPAMPoolCIDR()
					d := r.Data(nil)
					d.SetId(encodeIPAMPoolCIDRId(aws.StringValue(v.Cidr), poolID))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IPAM Pool (%s) CIDRs (%s): %w", poolID, region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IPAM Pool sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing IPAM Pools (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping IPAM Pools (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepIPAMPools(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeIpamPoolsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeIpamPoolsPages(input, func(page *ec2.DescribeIpamPoolsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IpamPools {
			r := ResourceIPAMPool()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.IpamPoolId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IPAM Pool sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IPAM Pools (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IPAM Pools (%s): %w", region, err)
	}

	return nil
}

func sweepIPAMScopes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeIpamScopesInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeIpamScopesPages(input, func(page *ec2.DescribeIpamScopesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IpamScopes {
			scopeID := aws.StringValue(v.IpamScopeId)

			if aws.BoolValue(v.IsDefault) {
				log.Printf("[DEBUG] Skipping default IPAM Scope (%s)", scopeID)
				continue
			}

			r := ResourceIPAMScope()
			d := r.Data(nil)
			d.SetId(scopeID)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IPAM Scope sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IPAM Scopes (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IPAM Scopes (%s): %w", region, err)
	}

	return nil
}

func sweepIPAMs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribeIpamsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeIpamsPages(input, func(page *ec2.DescribeIpamsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Ipams {
			r := ResourceIPAM()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.IpamId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IPAM sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing IPAMs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IPAMs (%s): %w", region, err)
	}

	return nil
}

func sweepAMIs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeImagesInput{
		Owners: aws.StringSlice([]string{"self"}),
	}
	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)

	output, err := conn.DescribeImages(input)

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AMI sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AMIs (%s): %w", region, err)
	}

	for _, v := range output.Images {
		r := ResourceAMI()
		d := r.Data(nil)
		d.SetId(aws.StringValue(v.ImageId))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AMIs (%s): %w", region, err)
	}

	return nil
}
