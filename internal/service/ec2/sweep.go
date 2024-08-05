// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
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
		F:    sweepCarrierGateways,
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

	resource.AddTestSweepers("aws_ec2_fleet", &resource.Sweeper{
		Name: "aws_ec2_fleet",
		F:    sweepFleets,
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
			"aws_eip_domain_name",
			"aws_vpc",
		},
		F: sweepEIPs,
	})

	resource.AddTestSweepers("aws_eip_domain_name", &resource.Sweeper{
		Name: "aws_eip_domain_name",
		F:    sweepEIPDomainNames,
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
			"aws_spot_instance_request",
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
			"aws_spot_instance_request",
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
			"aws_db_proxy",
			"aws_directory_service_directory",
			"aws_ec2_client_vpn_endpoint",
			"aws_ec2_traffic_mirror_session",
			"aws_ec2_transit_gateway_vpc_attachment",
			"aws_eks_cluster",
			"aws_elb",
			"aws_instance",
			"aws_lb",
			"aws_nat_gateway",
			"aws_rds_cluster",
			"aws_rds_global_cluster",
		},
	})

	resource.AddTestSweepers("aws_ec2_managed_prefix_list", &resource.Sweeper{
		Name: "aws_ec2_managed_prefix_list",
		F:    sweepManagedPrefixLists,
		Dependencies: []string{
			"aws_route_table",
			"aws_security_group",
			"aws_networkfirewall_rule_group",
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
			"aws_spot_instance_request",
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

	resource.AddTestSweepers("aws_spot_instance_request", &resource.Sweeper{
		Name: "aws_spot_instance_request",
		F:    sweepSpotInstanceRequests,
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
			"aws_dms_replication_subnet_group",
			"aws_docdb_subnet_group",
			"aws_ec2_client_vpn_endpoint",
			"aws_ec2_instance_connect_endpoint",
			"aws_ec2_transit_gateway_vpc_attachment",
			"aws_efs_file_system",
			"aws_eks_cluster",
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
			"aws_elasticache_subnet_group",
			"aws_elasticsearch_domain",
			"aws_elb",
			"aws_emr_cluster",
			"aws_emr_studio",
			"aws_fsx_lustre_file_system",
			"aws_fsx_ontap_file_system",
			"aws_fsx_openzfs_file_system",
			"aws_fsx_windows_file_system",
			"aws_grafana_workspace",
			"aws_iot_topic_rule_destination",
			"aws_lambda_function",
			"aws_lb",
			"aws_memorydb_subnet_group",
			"aws_mq_broker",
			"aws_msk_cluster",
			"aws_network_interface",
			"aws_networkfirewall_firewall",
			"aws_opensearch_domain",
			"aws_quicksight_vpc_connection",
			"aws_redshift_cluster",
			"aws_redshift_subnet_group",
			"aws_route53_resolver_endpoint",
			"aws_sagemaker_notebook_instance",
			"aws_spot_fleet_request",
			"aws_spot_instance_request",
			"aws_vpc_endpoint",
		},
	})

	resource.AddTestSweepers("aws_ec2_traffic_mirror_filter", &resource.Sweeper{
		Name: "aws_ec2_traffic_mirror_filter",
		F:    sweepTrafficMirrorFilters,
		Dependencies: []string{
			"aws_ec2_traffic_mirror_session",
		},
	})

	resource.AddTestSweepers("aws_ec2_traffic_mirror_session", &resource.Sweeper{
		Name: "aws_ec2_traffic_mirror_session",
		F:    sweepTrafficMirrorSessions,
	})

	resource.AddTestSweepers("aws_ec2_traffic_mirror_target", &resource.Sweeper{
		Name: "aws_ec2_traffic_mirror_target",
		F:    sweepTrafficMirrorTargets,
		Dependencies: []string{
			"aws_ec2_traffic_mirror_session",
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

	resource.AddTestSweepers("aws_vpc_endpoint", &resource.Sweeper{
		Name: "aws_vpc_endpoint",
		F:    sweepVPCEndpoints,
		Dependencies: []string{
			"aws_route_table",
			"aws_sagemaker_workforce",
			"aws_vpc_endpoint_connection_accepter",
		},
	})

	resource.AddTestSweepers("aws_vpc_endpoint_connection_accepter", &resource.Sweeper{
		Name: "aws_vpc_endpoint_connection_accepter",
		F:    sweepVPCEndpointConnectionAccepters,
	})

	resource.AddTestSweepers("aws_vpc_endpoint_service", &resource.Sweeper{
		Name: "aws_vpc_endpoint_service",
		F:    sweepVPCEndpointServices,
		Dependencies: []string{
			"aws_vpc_endpoint",
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
			"aws_vpclattice_service_network",
			"aws_vpclattice_target_group",
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

	resource.AddTestSweepers("aws_vpc_ipam", &resource.Sweeper{
		Name: "aws_vpc_ipam",
		F:    sweepIPAMs,
	})

	resource.AddTestSweepers("aws_vpc_ipam_resource_discovery", &resource.Sweeper{
		Name: "aws_vpc_ipam_resource_discovery",
		F:    sweepIPAMResourceDiscoveries,
	})

	resource.AddTestSweepers("aws_ami", &resource.Sweeper{
		Name: "aws_ami",
		F:    sweepAMIs,
	})

	resource.AddTestSweepers("aws_vpc_network_performance_metric_subscription", &resource.Sweeper{
		Name: "aws_vpc_network_performance_metric_subscription",
		F:    sweepNetworkPerformanceMetricSubscriptions,
	})

	resource.AddTestSweepers("aws_ec2_instance_connect_endpoint", &resource.Sweeper{
		Name: "aws_ec2_instance_connect_endpoint",
		F:    sweepInstanceConnectEndpoints,
	})

	resource.AddTestSweepers("aws_verifiedaccess_trust_provider", &resource.Sweeper{
		Name: "aws_verifiedaccess_trust_provider",
		F:    sweepVerifiedAccessTrustProviders,
		Dependencies: []string{
			"aws_verifiedaccess_instance_trust_provider_attachment",
		},
	})

	resource.AddTestSweepers("aws_verifiedaccess_instance_trust_provider_attachment", &resource.Sweeper{
		Name: "aws_verifiedaccess_instance_trust_provider_attachment",
		F:    sweepVerifiedAccessTrustProviderAttachments,
		Dependencies: []string{
			"aws_verifiedaccess_group",
		},
	})

	resource.AddTestSweepers("aws_verifiedaccess_group", &resource.Sweeper{
		Name: "aws_verifiedaccess_group",
		F:    sweepVerifiedAccessGroups,
		Dependencies: []string{
			"aws_verifiedaccess_endpoint",
		},
	})

	resource.AddTestSweepers("aws_verifiedaccess_endpoint", &resource.Sweeper{
		Name: "aws_verifiedaccess_endpoint",
		F:    sweepVerifiedAccessEndpoints,
	})

	resource.AddTestSweepers("aws_verifiedaccess_instance", &resource.Sweeper{
		Name: "aws_verifiedaccess_instance",
		F:    sweepVerifiedAccessInstances,
		Dependencies: []string{
			"aws_verifiedaccess_trust_provider",
		},
	})
}

func sweepCapacityReservations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)

	resp, err := conn.DescribeCapacityReservations(ctx, &ec2.DescribeCapacityReservationsInput{})

	if awsv2.SkipSweepError(err) {
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
		if r.State != awstypes.CapacityReservationStateCancelled && r.State != awstypes.CapacityReservationStateExpired {
			id := aws.ToString(r.CapacityReservationId)

			log.Printf("[INFO] Cancelling EC2 Capacity Reservation EC2 Instance: %s", id)

			opts := &ec2.CancelCapacityReservationInput{
				CapacityReservationId: aws.String(id),
			}

			_, err := conn.CancelCapacityReservation(ctx, opts)

			if err != nil {
				log.Printf("[ERROR] Error cancelling EC2 Capacity Reservation (%s): %s", id, err)
			}
		}
	}

	return nil
}

func sweepCarrierGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeCarrierGatewaysInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeCarrierGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Carrier Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Carrier Gateways (%s): %w", region, err)
		}

		for _, v := range page.CarrierGateways {
			r := resourceCarrierGateway()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CarrierGatewayId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Carrier Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepClientVPNEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeClientVpnEndpointsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeClientVpnEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Client VPN Endpoint sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Client VPN Endpoints (%s): %w", region, err)
		}

		for _, v := range page.ClientVpnEndpoints {
			r := resourceClientVPNEndpoint()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ClientVpnEndpointId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Client VPN Endpoints (%s): %w", region, err)
	}

	return nil
}

func sweepClientVPNNetworkAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeClientVpnEndpointsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeClientVpnEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Client VPN Network Association sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Client VPN Endpoints (%s): %w", region, err))
		}

		for _, v := range page.ClientVpnEndpoints {
			input := &ec2.DescribeClientVpnTargetNetworksInput{
				ClientVpnEndpointId: v.ClientVpnEndpointId,
			}

			pages := ec2.NewDescribeClientVpnTargetNetworksPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Client VPN Network Associations (%s): %w", region, err))
				}

				for _, v := range page.ClientVpnTargetNetworks {
					r := resourceClientVPNNetworkAssociation()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.AssociationId))
					d.Set("client_vpn_endpoint_id", v.ClientVpnEndpointId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EC2 Client VPN Network Associations (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFleets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)

	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeFleetsPaginator(conn, &ec2.DescribeFleetsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Fleet sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving EC2 Fleets: %w", err))
		}

		for _, fleet := range page.Fleets {
			if fleet.FleetState == awstypes.FleetStateCodeDeleted || fleet.FleetState == awstypes.FleetStateCodeDeletedTerminatingInstances {
				continue
			}

			r := resourceFleet()
			d := r.Data(nil)
			d.SetId(aws.ToString(fleet.FleetId))
			d.Set("terminate_instances", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EC2 Fleets for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepEBSVolumes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVolumesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVolumesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 EBS Volume sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 EBS Volumes (%s): %w", region, err)
		}

		for _, v := range page.Volumes {
			id := aws.ToString(v.VolumeId)

			if v.State != awstypes.VolumeStateAvailable {
				log.Printf("[INFO] Skipping EC2 EBS Volume (%s): %s", v.State, id)
				continue
			}

			r := resourceEBSVolume()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 EBS Volumes (%s): %w", region, err)
	}

	return nil
}

func sweepEBSSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EBS Snapshot sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EBS Snapshots (%s): %w", region, err)
		}

		for _, v := range page.Snapshots {
			r := resourceEBSSnapshot()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.SnapshotId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EBS Snapshots (%s): %w", region, err)
	}

	return nil
}

func sweepEgressOnlyInternetGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeEgressOnlyInternetGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Egress-only Internet Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Egress-only Internet Gateways (%s): %w", region, err)
		}

		for _, v := range page.EgressOnlyInternetGateways {
			r := resourceEgressOnlyInternetGateway()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.EgressOnlyInternetGatewayId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Egress-only Internet Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepEIPs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	// There is currently no paginator or Marker/NextToken
	input := &ec2.DescribeAddressesInput{}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeAddresses(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 EIP sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing EC2 EIPs: %s", err)
	}

	for _, v := range output.Addresses {
		publicIP := aws.ToString(v.PublicIp)

		if v.AssociationId != nil {
			log.Printf("[INFO] Skipping EC2 EIP (%s) with association: %s", publicIP, aws.ToString(v.AssociationId))
			continue
		}

		r := resourceEIP()
		d := r.Data(nil)
		if v.AllocationId != nil {
			d.SetId(aws.ToString(v.AllocationId))
		} else {
			d.SetId(aws.ToString(v.PublicIp))
		}

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 EIPs (%s): %w", region, err)
	}

	return nil
}

func sweepEIPDomainNames(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeAddressesAttributeInput{
		Attribute: awstypes.AddressAttributeNameDomainName,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeAddressesAttributePaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EIP Domain Name sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EIP Domain Names (%s): %w", region, err)
		}

		for _, v := range page.Addresses {
			sweepResources = append(sweepResources, framework.NewSweepResource(newEIPDomainNameResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.AllocationId)),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EIP Domain Names (%s): %w", region, err)
	}

	return nil
}

func sweepFlowLogs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeFlowLogsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeFlowLogsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Flow Log sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Flow Logs (%s): %w", region, err)
		}

		for _, flowLog := range page.FlowLogs {
			r := resourceFlowLog()
			d := r.Data(nil)
			d.SetId(aws.ToString(flowLog.FlowLogId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Flow Logs (%s): %w", region, err)
	}

	return nil
}

func sweepHosts(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeHostsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeHostsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Host sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Hosts (%s): %w", region, err)
		}

		for _, host := range page.Hosts {
			r := resourceHost()
			d := r.Data(nil)
			d.SetId(aws.ToString(host.HostId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Hosts (%s): %w", region, err)
	}

	return nil
}

func sweepInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Instances (%s): %w", region, err)
		}

		for _, v := range page.Reservations {
			for _, v := range v.Instances {
				id := aws.ToString(v.InstanceId)

				if v.State != nil && v.State.Name == awstypes.InstanceStateNameTerminated {
					log.Printf("[INFO] Skipping terminated EC2 Instance: %s", id)
					continue
				}

				if err := disableInstanceAPIStop(ctx, conn, id, false); err != nil {
					log.Printf("[INFO] EC2 Instance (%s): %s", id, err)
				}

				r := resourceInstance()
				d := r.Data(nil)
				d.SetId(id)

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Instances (%s): %w", region, err)
	}

	return nil
}

func sweepInternetGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)

	defaultVPCID := ""
	describeVpcsInput := &ec2.DescribeVpcsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: []string{"true"},
			},
		},
	}

	describeVpcsOutput, err := conn.DescribeVpcs(ctx, describeVpcsInput)

	if err != nil {
		return fmt.Errorf("error describing VPCs: %w", err)
	}

	if describeVpcsOutput != nil && len(describeVpcsOutput.Vpcs) == 1 {
		defaultVPCID = aws.ToString(describeVpcsOutput.Vpcs[0].VpcId)
	}

	input := &ec2.DescribeInternetGatewaysInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeInternetGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Internet Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Internet Gateways (%s): %w", region, err)
		}

		for _, internetGateway := range page.InternetGateways {
			internetGatewayID := aws.ToString(internetGateway.InternetGatewayId)
			isDefaultVPCInternetGateway := false

			for _, attachment := range internetGateway.Attachments {
				if aws.ToString(attachment.VpcId) == defaultVPCID {
					isDefaultVPCInternetGateway = true
					break
				}
			}

			if isDefaultVPCInternetGateway {
				log.Printf("[DEBUG] Skipping Default VPC Internet Gateway: %s", internetGatewayID)
				continue
			}

			r := resourceInternetGateway()
			d := r.Data(nil)
			d.SetId(internetGatewayID)
			if len(internetGateway.Attachments) > 0 {
				d.Set(names.AttrVPCID, internetGateway.Attachments[0].VpcId)
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Internet Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepKeyPairs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeKeyPairsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeKeyPairs(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Key Pair sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Key Pairs (%s): %w", region, err)
	}

	for _, v := range output.KeyPairs {
		r := resourceKeyPair()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.KeyName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Key Pairs (%s): %w", region, err)
	}

	return nil
}

func sweepLaunchTemplates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeLaunchTemplatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeLaunchTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Launch Template sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Launch Templates (%s): %w", region, err)
		}

		for _, v := range page.LaunchTemplates {
			r := resourceLaunchTemplate()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.LaunchTemplateId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Launch Templates (%s): %w", region, err)
	}

	return nil
}

func sweepNATGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeNatGatewaysInput{}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeNatGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 NAT Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 NAT Gateways (%s): %w", region, err)
		}

		for _, v := range page.NatGateways {
			r := resourceNATGateway()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.NatGatewayId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 NAT Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepNetworkACLs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &ec2.DescribeNetworkAclsInput{}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeNetworkAclsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Network ACL sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Network ACLs (%s): %w", region, err)
		}

		for _, v := range page.NetworkAcls {
			if aws.ToBool(v.IsDefault) {
				continue
			}

			r := resourceNetworkACL()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.NetworkAclId))

			var subnetIDs []string
			for _, v := range v.Associations {
				subnetIDs = append(subnetIDs, aws.ToString(v.SubnetId))
			}
			d.Set(names.AttrSubnetIDs, subnetIDs)

			d.Set(names.AttrVPCID, v.VpcId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Network ACLs (%s): %w", region, err)
	}

	return nil
}

func sweepNetworkInterfaces(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeNetworkInterfacesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeNetworkInterfacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Network Interface sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Network Interfaces (%s): %w", region, err)
		}

		for _, v := range page.NetworkInterfaces {
			id := aws.ToString(v.NetworkInterfaceId)

			if v.Status != awstypes.NetworkInterfaceStatusAvailable {
				log.Printf("[INFO] Skipping EC2 Network Interface (%s): %s", v.Status, id)
				continue
			}

			r := resourceNetworkInterface()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Network Interfaces (%s): %w", region, err)
	}

	return nil
}

func sweepManagedPrefixLists(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeManagedPrefixListsPaginator(conn, &ec2.DescribeManagedPrefixListsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Managed Prefix List sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Managed Prefix Lists (%s): %w", region, err)
		}

		for _, v := range page.PrefixLists {
			if aws.ToString(v.OwnerId) == "AWS" {
				log.Printf("[DEBUG] Skipping AWS-managed prefix list: %s", aws.ToString(v.PrefixListName))
				continue
			}

			r := resourceManagedPrefixList()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.PrefixListId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Managed Prefix Lists (%s): %w", region, err)
	}

	return nil
}

func sweepNetworkInsightsPaths(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := ec2.NewDescribeNetworkInsightsPathsPaginator(conn, &ec2.DescribeNetworkInsightsPathsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Insights Path sweep for %s: %s", region, errs)
			return nil
		}

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing Network Insights Paths for %s: %w", region, err))
		}

		for _, v := range page.NetworkInsightsPaths {
			r := resourceNetworkInsightsPath()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.NetworkInsightsPathId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Network Insights Paths for %s: %w", region, err))
	}

	return errs.ErrorOrNil()
}

func sweepPlacementGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribePlacementGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribePlacementGroups(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Placement Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Placement Groups (%s): %w", region, err)
	}

	for _, v := range output.PlacementGroups {
		r := resourcePlacementGroup()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.GroupName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Placement Groups (%s): %w", region, err)
	}

	return nil
}

func sweepRouteTables(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.EC2Client(ctx)
	var sweeperErrs *multierror.Error
	input := &ec2.DescribeRouteTablesInput{}

	pages := ec2.NewDescribeRouteTablesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Route Table sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Route Tables: %w", err))
		}

		for _, routeTable := range page.RouteTables {
			id := aws.ToString(routeTable.RouteTableId)
			isMainRouteTableAssociation := false

			for _, routeTableAssociation := range routeTable.Associations {
				if aws.ToBool(routeTableAssociation.Main) {
					isMainRouteTableAssociation = true
					break
				}

				associationID := aws.ToString(routeTableAssociation.RouteTableAssociationId)

				input := &ec2.DisassociateRouteTableInput{
					AssociationId: routeTableAssociation.RouteTableAssociationId,
				}

				log.Printf("[DEBUG] Deleting EC2 Route Table Association: %s", associationID)
				_, err := conn.DisassociateRouteTable(ctx, input)

				if err != nil {
					sweeperErr := fmt.Errorf("error deleting EC2 Route Table (%s) Association (%s): %w", id, associationID, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}

			if isMainRouteTableAssociation {
				for _, route := range routeTable.Routes {
					if gatewayID := aws.ToString(route.GatewayId); gatewayID == gatewayIDLocal || gatewayID == gatewayIDVPCLattice {
						continue
					}

					// Prevent deleting default VPC route for Internet Gateway
					// which some testing is still reliant on operating correctly
					if strings.HasPrefix(aws.ToString(route.GatewayId), "igw-") && aws.ToString(route.DestinationCidrBlock) == "0.0.0.0/0" {
						continue
					}

					input := &ec2.DeleteRouteInput{
						DestinationCidrBlock:     route.DestinationCidrBlock,
						DestinationIpv6CidrBlock: route.DestinationIpv6CidrBlock,
						RouteTableId:             routeTable.RouteTableId,
					}

					log.Printf("[DEBUG] Deleting EC2 Route Table (%s) Route", id)
					_, err := conn.DeleteRoute(ctx, input)

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
			_, err := conn.DeleteRouteTable(ctx, input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting EC2 Route Table (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSecurityGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeSecurityGroupsInput{}

	// Delete all non-default EC2 Security Group Rules to prevent DependencyViolation errors
	pages := ec2.NewDescribeSecurityGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Security Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving EC2 Security Groups: %w", err)
		}

		for _, sg := range page.SecurityGroups {
			if aws.ToString(sg.GroupName) == "default" {
				log.Printf("[DEBUG] Skipping default EC2 Security Group: %s", aws.ToString(sg.GroupId))
				continue
			}

			if sg.IpPermissions != nil {
				req := &ec2.RevokeSecurityGroupIngressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissions,
				}

				if _, err = conn.RevokeSecurityGroupIngress(ctx, req); err != nil {
					log.Printf("[ERROR] Error revoking ingress rule for Security Group (%s): %s", aws.ToString(sg.GroupId), err)
				}
			}

			if sg.IpPermissionsEgress != nil {
				req := &ec2.RevokeSecurityGroupEgressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissionsEgress,
				}

				if _, err = conn.RevokeSecurityGroupEgress(ctx, req); err != nil {
					log.Printf("[ERROR] Error revoking egress rule for Security Group (%s): %s", aws.ToString(sg.GroupId), err)
				}
			}
		}
	}

	pages = ec2.NewDescribeSecurityGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Security Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving EC2 Security Groups: %w", err)
		}

		for _, sg := range page.SecurityGroups {
			if aws.ToString(sg.GroupName) == "default" {
				log.Printf("[DEBUG] Skipping default EC2 Security Group: %s", aws.ToString(sg.GroupId))
				continue
			}

			input := &ec2.DeleteSecurityGroupInput{
				GroupId: sg.GroupId,
			}

			// Handle EC2 eventual consistency
			err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
				_, err := conn.DeleteSecurityGroup(ctx, input)

				if tfawserr.ErrCodeEquals(err, "DependencyViolation") {
					return retry.RetryableError(err)
				}
				if err != nil {
					return retry.NonRetryableError(err)
				}
				return nil
			})

			if err != nil {
				log.Printf("[ERROR] Error deleting Security Group (%s): %s", aws.ToString(sg.GroupId), err)
			}
		}
	}

	return nil
}

func sweepSpotFleetRequests(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := ec2.NewDescribeSpotFleetRequestsPaginator(conn, &ec2.DescribeSpotFleetRequestsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(errs.ErrorOrNil()) {
			log.Printf("[WARN] Skipping EC2 Spot Fleet Requests sweep for %s: %s", region, errs)
			return nil
		}

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing EC2 Spot Fleet Requests for %s: %w", region, err))
		}

		if len(page.SpotFleetRequestConfigs) == 0 {
			log.Print("[DEBUG] No Spot Fleet Requests to sweep")
			return nil
		}

		for _, v := range page.SpotFleetRequestConfigs {
			r := resourceSpotFleetRequest()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.SpotFleetRequestId))
			d.Set("terminate_instances_with_expiration", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 Spot Fleet Requests for %s: %w", region, err))
	}

	return errs.ErrorOrNil()
}

func sweepSpotInstanceRequests(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := ec2.NewDescribeSpotInstanceRequestsPaginator(conn, &ec2.DescribeSpotInstanceRequestsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(errs.ErrorOrNil()) {
			log.Printf("[WARN] Skipping EC2 Spot Instance Requests sweep for %s: %s", region, errs)
			return nil
		}

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing EC2 Spot Instance Requests for %s: %w", region, err))
		}

		if len(page.SpotInstanceRequests) == 0 {
			log.Print("[DEBUG] No Spot Instance Requests to sweep")
			return nil
		}

		for _, v := range page.SpotInstanceRequests {
			r := resourceSpotInstanceRequest()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.SpotInstanceRequestId))
			d.Set("spot_instance_id", v.InstanceId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 Spot Instance Requests for %s: %w", region, err))
	}

	return errs.ErrorOrNil()
}

func sweepSubnets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeSubnetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeSubnetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Subnet sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Subnets (%s): %w", region, err)
		}

		for _, v := range page.Subnets {
			// Skip default subnets.
			if aws.ToBool(v.DefaultForAz) {
				continue
			}

			r := resourceSubnet()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.SubnetId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Subnets (%s): %w", region, err)
	}

	return nil
}

func sweepTrafficMirrorFilters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeTrafficMirrorFiltersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeTrafficMirrorFiltersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Traffic Mirror Filter sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Traffic Mirror Filters (%s): %w", region, err)
		}

		for _, v := range page.TrafficMirrorFilters {
			r := resourceTrafficMirrorFilter()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TrafficMirrorFilterId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Traffic Mirror Filters (%s): %w", region, err)
	}

	return nil
}

func sweepTrafficMirrorSessions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeTrafficMirrorSessionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeTrafficMirrorSessionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Traffic Mirror Session sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Traffic Mirror Sessions (%s): %w", region, err)
		}

		for _, v := range page.TrafficMirrorSessions {
			r := resourceTrafficMirrorSession()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TrafficMirrorSessionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Traffic Mirror Sessions (%s): %w", region, err)
	}

	return nil
}

func sweepTrafficMirrorTargets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeTrafficMirrorTargetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeTrafficMirrorTargetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Traffic Mirror Target sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Traffic Mirror Targets (%s): %w", region, err)
		}

		for _, v := range page.TrafficMirrorTargets {
			r := resourceTrafficMirrorTarget()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TrafficMirrorTargetId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Traffic Mirror Targets (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeTransitGatewaysInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeTransitGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Transit Gateways (%s): %w", region, err)
		}

		for _, v := range page.TransitGateways {
			if v.State == awstypes.TransitGatewayStateDeleted {
				continue
			}

			r := resourceTransitGateway()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TransitGatewayId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayConnectPeers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeTransitGatewayConnectPeersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeTransitGatewayConnectPeersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway Connect Peer sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Transit Gateway Connect Peers (%s): %w", region, err)
		}

		for _, v := range page.TransitGatewayConnectPeers {
			if v.State == awstypes.TransitGatewayConnectPeerStateDeleted {
				continue
			}

			r := resourceTransitGatewayConnectPeer()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TransitGatewayConnectPeerId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateway Connect Peers (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayConnects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeTransitGatewayConnectsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeTransitGatewayConnectsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway Connect sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Transit Gateway Connects (%s): %w", region, err)
		}

		for _, v := range page.TransitGatewayConnects {
			if v.State == awstypes.TransitGatewayAttachmentStateDeleted {
				continue
			}

			r := resourceTransitGatewayConnect()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TransitGatewayAttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateway Connects (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayMulticastDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeTransitGatewayMulticastDomainsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway Multicast Domain sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Transit Gateway Multicast Domains (%s): %w", region, err)
		}

		for _, v := range page.TransitGatewayMulticastDomains {
			if v.State == awstypes.TransitGatewayMulticastDomainStateDeleted {
				continue
			}

			r := resourceTransitGatewayMulticastDomain()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TransitGatewayMulticastDomainId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateway Multicast Domains (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayPeeringAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeTransitGatewayPeeringAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway Peering Attachment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Transit Gateway Peering Attachments (%s): %w", region, err)
		}

		for _, v := range page.TransitGatewayPeeringAttachments {
			if v.State == awstypes.TransitGatewayAttachmentStateDeleted {
				continue
			}

			r := resourceTransitGatewayPeeringAttachment()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TransitGatewayAttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateway Peering Attachments (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayVPCAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeTransitGatewayVpcAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway VPC Attachment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Transit Gateway VPC Attachments (%s): %w", region, err)
		}

		for _, v := range page.TransitGatewayVpcAttachments {
			if v.State == awstypes.TransitGatewayAttachmentStateDeleted {
				continue
			}

			r := resourceTransitGatewayVPCAttachment()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TransitGatewayAttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Transit Gateway VPC Attachments (%s): %w", region, err)
	}

	return nil
}

func sweepVPCDHCPOptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	input := &ec2.DescribeDhcpOptionsInput{}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeDhcpOptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 DHCP Options Set sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 DHCP Options Sets (%s): %w", region, err)
		}

		for _, v := range page.DhcpOptions {
			// Skip the default DHCP Options.
			var defaultDomainNameFound, defaultDomainNameServersFound bool

			for _, v := range v.DhcpConfigurations {
				if aws.ToString(v.Key) == "domain-name" {
					if len(v.Values) != 1 {
						continue
					}

					if aws.ToString(v.Values[0].Value) == client.EC2RegionalPrivateDNSSuffix(ctx) {
						defaultDomainNameFound = true
					}
				} else if aws.ToString(v.Key) == "domain-name-servers" {
					if len(v.Values) != 1 {
						continue
					}

					if aws.ToString(v.Values[0].Value) == "AmazonProvidedDNS" {
						defaultDomainNameServersFound = true
					}
				}
			}

			if defaultDomainNameFound && defaultDomainNameServersFound {
				continue
			}

			r := resourceVPCDHCPOptions()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DhcpOptionsId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 DHCP Options Sets (%s): %w", region, err)
	}

	return nil
}

func sweepVPCEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVpcEndpointsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVpcEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPC Endpoint sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 VPC Endpoints (%s): %w", region, err)
		}

		for _, v := range page.VpcEndpoints {
			id := aws.ToString(v.VpcEndpointId)

			if v.State == vpcEndpointStateDeleted {
				log.Printf("[INFO] Skipping EC2 VPC Endpoint %s: State=%s", id, v.State)
				continue
			}

			if requesterManaged := aws.ToBool(v.RequesterManaged); requesterManaged {
				log.Printf("[INFO] Skipping EC2 VPC Endpoint %s: RequesterManaged=%t", id, requesterManaged)
				continue
			}

			r := resourceVPCEndpoint()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPC Endpoints (%s): %w", region, err)
	}

	return nil
}

func sweepVPCEndpointConnectionAccepters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVpcEndpointConnectionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVpcEndpointConnectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPC Endpoint Connection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 VPC Endpoint Connections (%s): %w", region, err)
		}

		for _, v := range page.VpcEndpointConnections {
			id := vpcEndpointConnectionAccepterCreateResourceID(aws.ToString(v.ServiceId), aws.ToString(v.VpcEndpointId))

			r := resourceVPCEndpointConnectionAccepter()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPC Endpoint Connections (%s): %w", region, err)
	}

	return nil
}

func sweepVPCEndpointServices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVpcEndpointServiceConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPC Endpoint Service sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 VPC Endpoint Services (%s): %w", region, err)
		}

		for _, v := range page.ServiceConfigurations {
			id := aws.ToString(v.ServiceId)

			if v.ServiceState == awstypes.ServiceStateDeleted {
				log.Printf("[INFO] Skipping EC2 VPC Endpoint Service %s: ServiceState=%s", id, v.ServiceState)
				continue
			}

			r := resourceVPCEndpointService()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPC Endpoint Services (%s): %w", region, err)
	}

	return nil
}

func sweepVPCPeeringConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	input := &ec2.DescribeVpcPeeringConnectionsInput{}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVpcPeeringConnectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPC Peering Connection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 VPC Peering Connections (%s): %w", region, err)
		}

		for _, v := range page.VpcPeeringConnections {
			r := resourceVPCPeeringConnection()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VpcPeeringConnectionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPC Peering Connections (%s): %w", region, err)
	}

	return nil
}

func sweepVPCs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVpcsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVpcsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPC sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 VPCs (%s): %w", region, err)
		}

		for _, v := range page.Vpcs {
			// Skip default VPCs.
			if aws.ToBool(v.IsDefault) {
				continue
			}

			r := resourceVPC()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VpcId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPCs (%s): %w", region, err)
	}

	return nil
}

func sweepVPNConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVpnConnectionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeVpnConnections(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPN Connection sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 VPN Connections (%s): %w", region, err)
	}

	for _, v := range output.VpnConnections {
		if v.State == awstypes.VpnStateDeleted {
			continue
		}

		r := resourceVPNConnection()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.VpnConnectionId))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPN Connections (%s): %w", region, err)
	}

	return nil
}

func sweepVPNGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVpnGatewaysInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeVpnGateways(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPN Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 VPN Gateways (%s): %w", region, err)
	}

	for _, v := range output.VpnGateways {
		if v.State == awstypes.VpnStateDeleted {
			continue
		}

		r := resourceVPNGateway()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.VpnGatewayId))

		for _, v := range v.VpcAttachments {
			if v.State != awstypes.AttachmentStatusDetached {
				d.Set(names.AttrVPCID, v.VpcId)

				break
			}
		}

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 VPN Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepCustomerGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeCustomerGatewaysInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeCustomerGateways(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Customer Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EC2 Customer Gateways (%s): %w", region, err)
	}

	for _, v := range output.CustomerGateways {
		if aws.ToString(v.State) == customerGatewayStateDeleted {
			continue
		}

		r := resourceCustomerGateway()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.CustomerGatewayId))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Customer Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepIPAMs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeIpamsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeIpamsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IPAM sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IPAMs (%s): %w", region, err)
		}

		for _, v := range page.Ipams {
			r := resourceIPAM()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.IpamId))
			d.Set("cascade", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping IPAMs (%s): %w", region, err)
	}

	return nil
}

func sweepIPAMResourceDiscoveries(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeIpamResourceDiscoveriesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeIpamResourceDiscoveriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping IPAM Resource Discovery sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing IPAM Resource Discoveries (%s): %w", region, err)
		}

		for _, v := range page.IpamResourceDiscoveries {
			// do not attempt to delete default resource created by each ipam
			if !aws.ToBool(v.IsDefault) {
				r := resourceIPAMResourceDiscovery()
				d := r.Data(nil)
				d.SetId(aws.ToString(v.IpamResourceDiscoveryId))

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Resource Discoveries (%s): %w", region, err)
	}

	return nil
}

func sweepAMIs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	input := &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	}
	conn := client.EC2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeImagesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping AMI sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing AMIs (%s): %w", region, err)
		}

		for _, v := range page.Images {
			r := resourceAMI()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ImageId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AMIs (%s): %w", region, err)
	}

	return nil
}

func sweepNetworkPerformanceMetricSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeAwsNetworkPerformanceMetricSubscriptionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeAwsNetworkPerformanceMetricSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 AWS Network Performance Metric Subscription sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 AWS Network Performance Metric Subscriptions (%s): %w", region, err)
		}

		for _, v := range page.Subscriptions {
			r := resourceNetworkPerformanceMetricSubscription()
			id := networkPerformanceMetricSubscriptionCreateResourceID(aws.ToString(v.Source), aws.ToString(v.Destination), string(v.Metric), string(v.Statistic))
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 AWS Network Performance Metric Subscriptions (%s): %w", region, err)
	}

	return nil
}

func sweepInstanceConnectEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EC2Client(ctx)
	input := &ec2.DescribeInstanceConnectEndpointsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeInstanceConnectEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Instance Connect Endpoint sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EC2 Instance Connect Endpoints (%s): %w", region, err)
		}

		for _, v := range page.InstanceConnectEndpoints {
			if v.State == awstypes.Ec2InstanceConnectEndpointStateDeleteComplete {
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newInstanceConnectEndpointResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.InstanceConnectEndpointId)),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EC2 Instance Connect Endpoints (%s): %w", region, err)
	}

	return nil
}

func sweepVerifiedAccessEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVerifiedAccessEndpointsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVerifiedAccessEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Verified Access Endpoint sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Verified Access Endpoints (%s): %w", region, err)
		}

		for _, v := range page.VerifiedAccessEndpoints {
			r := resourceVerifiedAccessEndpoint()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VerifiedAccessEndpointId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Verified Access Endpoints (%s): %w", region, err)
	}

	return nil
}

func sweepVerifiedAccessGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVerifiedAccessGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVerifiedAccessGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Verified Access Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Verified Access Groups (%s): %w", region, err)
		}

		for _, v := range page.VerifiedAccessGroups {
			r := resourceVerifiedAccessGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VerifiedAccessGroupId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Verified Access Groups (%s): %w", region, err)
	}

	return nil
}

func sweepVerifiedAccessInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVerifiedAccessInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVerifiedAccessInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Verified Access Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Verified Access Instances (%s): %w", region, err)
		}

		for _, v := range page.VerifiedAccessInstances {
			r := resourceVerifiedAccessInstance()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VerifiedAccessInstanceId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Verified Access Instances (%s): %w", region, err)
	}

	return nil
}

func sweepVerifiedAccessTrustProviders(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVerifiedAccessTrustProvidersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVerifiedAccessTrustProvidersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Verified Access Trust Provider sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Verified Access Trust Providers (%s): %w", region, err)
		}

		for _, v := range page.VerifiedAccessTrustProviders {
			r := resourceVerifiedAccessTrustProvider()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VerifiedAccessTrustProviderId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Verified Access Trust Providers (%s): %w", region, err)
	}

	return nil
}

func sweepVerifiedAccessTrustProviderAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.EC2Client(ctx)
	input := &ec2.DescribeVerifiedAccessInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ec2.NewDescribeVerifiedAccessInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Verified Access Instance Trust Provider Attachment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Verified Access Instance Trust Provider Attachments (%s): %w", region, err)
		}

		for _, v := range page.VerifiedAccessInstances {
			vaiID := aws.ToString(v.VerifiedAccessInstanceId)

			for _, v := range v.VerifiedAccessTrustProviders {
				vatpID := aws.ToString(v.VerifiedAccessTrustProviderId)

				r := resourceVerifiedAccessInstanceTrustProviderAttachment()
				d := r.Data(nil)
				d.SetId(verifiedAccessInstanceTrustProviderAttachmentCreateResourceID(vaiID, vatpID))

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Verified Access Instance Trust Provider Attachments (%s): %w", region, err)
	}

	return nil
}
