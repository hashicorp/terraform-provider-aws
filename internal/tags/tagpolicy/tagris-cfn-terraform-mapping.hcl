# Tag Resource Type to Terraform Resource Type Mapping
# Each alias block represents a Tag Policies resource type and its corresponding Terraform resource type(s)

tagtype "access-analyzer:analyzer" {
  terraform_types = ["aws_accessanalyzer_analyzer"]
}

tagtype "acm-pca:certificate-authority" {
  terraform_types = ["aws_acmpca_certificate_authority"]
}

tagtype "acm:certificate" {
  terraform_types = ["aws_acm_certificate"]
}

tagtype "airflow:environment" {
  terraform_types = ["aws_mwaa_environment"]
}

tagtype "amplify:apps" {
  terraform_types = ["aws_amplify_app"]
}

tagtype "aoss:collection" {
  terraform_types = ["aws_opensearchserverless_collection"]
}

tagtype "app-integrations:application" {
  terraform_types = []
}

tagtype "app-integrations:data-integration" {
  terraform_types = ["aws_appintegrations_data_integration"]
}

tagtype "app-integrations:event-integration" {
  terraform_types = ["aws_appintegrations_event_integration"]
}

tagtype "appconfig:application" {
  terraform_types = ["aws_appconfig_application"]
}

tagtype "appconfig:application/configurationprofile" {
  terraform_types = ["aws_appconfig_configuration_profile"]
}

tagtype "appconfig:application/environment" {
  terraform_types = ["aws_appconfig_environment"]
}

tagtype "appconfig:deploymentstrategy" {
  terraform_types = ["aws_appconfig_deployment_strategy"]
}

tagtype "appconfig:extension" {
  terraform_types = ["aws_appconfig_extension"]
}

tagtype "appconfig:extensionassociation" {
  terraform_types = ["aws_appconfig_extension_association"]
}

tagtype "appflow:flow" {
  terraform_types = ["aws_appflow_flow"]
}

tagtype "application-signals:slo" {
  terraform_types = []
}

tagtype "applicationinsights:application" {
  terraform_types = ["aws_applicationinsights_application"]
}

tagtype "appmesh:mesh" {
  terraform_types = ["aws_appmesh_mesh"]
}

tagtype "appmesh:mesh/virtualGateway" {
  terraform_types = ["aws_appmesh_virtual_gateway"]
}

tagtype "appmesh:mesh/virtualGateway/gatewayRoute" {
  terraform_types = ["aws_appmesh_gateway_route"]
}

tagtype "appmesh:mesh/virtualNode" {
  terraform_types = ["aws_appmesh_virtual_node"]
}

tagtype "appmesh:mesh/virtualRouter" {
  terraform_types = ["aws_appmesh_virtual_router"]
}

tagtype "appmesh:mesh/virtualRouter/route" {
  terraform_types = ["aws_appmesh_route"]
}

tagtype "appmesh:mesh/virtualService" {
  terraform_types = ["aws_appmesh_virtual_service"]
}

tagtype "apprunner:autoscalingconfiguration" {
  terraform_types = ["aws_apprunner_auto_scaling_configuration_version"]
}

tagtype "apprunner:observabilityconfiguration" {
  terraform_types = ["aws_apprunner_observability_configuration"]
}

tagtype "apprunner:service" {
  terraform_types = ["aws_apprunner_service"]
}

tagtype "apprunner:vpcconnector" {
  terraform_types = ["aws_apprunner_vpc_connector"]
}

tagtype "apprunner:vpcingressconnection" {
  terraform_types = ["aws_apprunner_vpc_ingress_connection"]
}

tagtype "appstream:app-block" {
  terraform_types = []
}

tagtype "appstream:application" {
  terraform_types = []
}

tagtype "appstream:fleet" {
  terraform_types = ["aws_appstream_fleet"]
}

tagtype "appstream:image-builder" {
  terraform_types = ["aws_appstream_image_builder"]
}

tagtype "appstream:stack" {
  terraform_types = ["aws_appstream_stack"]
}

tagtype "apptest:testcase" {
  terraform_types = []
}

tagtype "aps:rulegroupsnamespace" {
  terraform_types = []
}

tagtype "aps:workspace" {
  terraform_types = []
}

tagtype "athena:capacity-reservation" {
  terraform_types = ["aws_athena_capacity_reservation"]
}

tagtype "athena:datacatalog" {
  terraform_types = ["aws_athena_data_catalog"]
}

tagtype "athena:workgroup" {
  terraform_types = ["aws_athena_workgroup"]
}

tagtype "auditmanager:assessment" {
  terraform_types = ["aws_auditmanager_assessment"]
}

tagtype "b2bi:capability" {
  terraform_types = []
}

tagtype "b2bi:partnership" {
  terraform_types = []
}

tagtype "b2bi:profile" {
  terraform_types = []
}

tagtype "b2bi:transformer" {
  terraform_types = []
}

tagtype "backup-gateway:hypervisor" {
  terraform_types = []
}

tagtype "backup:backup-plan" {
  terraform_types = ["aws_backup_plan"]
}

tagtype "backup:framework" {
  terraform_types = ["aws_backup_framework"]
}

tagtype "backup:report-plan" {
  terraform_types = ["aws_backup_report_plan"]
}

tagtype "backup:restore-testing-plan" {
  terraform_types = ["aws_backup_restore_testing_plan"]
}

tagtype "batch:compute-environment" {
  terraform_types = ["aws_batch_compute_environment"]
}

tagtype "batch:consumable-resource" {
  terraform_types = []
}

tagtype "batch:job-definition" {
  terraform_types = ["aws_batch_job_definition"]
}

tagtype "batch:job-queue" {
  terraform_types = ["aws_batch_job_queue"]
}

tagtype "batch:scheduling-policy" {
  terraform_types = ["aws_batch_scheduling_policy"]
}

tagtype "bcm-data-exports:export" {
  terraform_types = ["aws_bcmdataexports_export"]
}

tagtype "bedrock:agent" {
  terraform_types = ["aws_bedrockagent_agent"]
}

tagtype "bedrock:agent-alias" {
  terraform_types = ["aws_bedrockagent_agent_alias"]
}

tagtype "bedrock:application-inference-profile" {
  terraform_types = []
}

tagtype "bedrock:blueprint" {
  terraform_types = []
}

tagtype "bedrock:data-automation-project" {
  terraform_types = []
}

tagtype "bedrock:flow" {
  terraform_types = ["aws_bedrockagent_flow"]
}

tagtype "bedrock:guardrail" {
  terraform_types = ["aws_bedrock_guardrail"]
}

tagtype "bedrock:knowledge-base" {
  terraform_types = ["aws_bedrockagent_knowledge_base"]
}

tagtype "bedrock:prompt" {
  terraform_types = ["aws_bedrockagent_prompt"]
}

tagtype "budgets:budget" {
  terraform_types = ["aws_budgets_budget"]
}

tagtype "budgets:budget/action" {
  terraform_types = ["aws_budgets_budget_action"]
}

tagtype "cassandra:keyspace" {
  terraform_types = ["aws_keyspaces_keyspace"]
}

tagtype "catalog:portfolio" {
  terraform_types = ["aws_servicecatalog_portfolio"]
}

tagtype "ce:anomalymonitor" {
  terraform_types = ["aws_ce_anomaly_monitor"]
}

tagtype "ce:anomalysubscription" {
  terraform_types = ["aws_ce_anomaly_subscription"]
}

tagtype "ce:costcategory" {
  terraform_types = ["aws_ce_cost_category"]
}

tagtype "cleanrooms-ml:training-dataset" {
  terraform_types = []
}

tagtype "cleanrooms:configuredtable" {
  terraform_types = ["aws_cleanrooms_configured_table"]
}

tagtype "cloudformation:stack" {
  terraform_types = ["aws_cloudformation_stack"]
}

tagtype "cloudformation:stackset" {
  terraform_types = ["aws_cloudformation_stack_set"]
}

tagtype "cloudfront:connection-group" {
  terraform_types = []
}

tagtype "cloudfront:distribution" {
  terraform_types = ["aws_cloudfront_distribution"]
}

tagtype "cloudtrail:channel" {
  terraform_types = []
}

tagtype "cloudtrail:dashboard" {
  terraform_types = []
}

tagtype "cloudtrail:eventdatastore" {
  terraform_types = ["aws_cloudtrail_event_data_store"]
}

tagtype "cloudtrail:trail" {
  terraform_types = ["aws_cloudtrail"]
}

tagtype "cloudwatch:alarm" {
  terraform_types = ["aws_cloudwatch_metric_alarm"]
}

tagtype "cloudwatch:insight-rule" {
  terraform_types = []
}

tagtype "cloudwatch:metric-stream" {
  terraform_types = ["aws_cloudwatch_metric_stream"]
}

tagtype "codeartifact:domain" {
  terraform_types = ["aws_codeartifact_domain"]
}

tagtype "codeartifact:repository" {
  terraform_types = ["aws_codeartifact_repository"]
}

tagtype "codebuild:project" {
  terraform_types = ["aws_codebuild_project"]
}

tagtype "codebuild:report-group" {
  terraform_types = ["aws_codebuild_report_group"]
}

tagtype "codecommit:repository" {
  terraform_types = ["aws_codecommit_repository"]
}

tagtype "codeconnections:connection" {
  terraform_types = ["aws_codeconnections_connection"]
}

tagtype "codedeploy:application" {
  terraform_types = ["aws_codedeploy_app"]
}

tagtype "codedeploy:deploymentconfig" {
  terraform_types = ["aws_codedeploy_deployment_config"]
}

tagtype "codeguru-profiler:profilingGroup" {
  terraform_types = ["aws_codeguruprofiler_profiling_group"]
}

tagtype "codeguru-reviewer:association" {
  terraform_types = ["aws_codegurureviewer_repository_association"]
}

tagtype "codepipeline:actiontype" {
  terraform_types = ["aws_codepipeline_custom_action_type"]
}

tagtype "codepipeline:pipeline" {
  terraform_types = ["aws_codepipeline"]
}

tagtype "codepipeline:webhook" {
  terraform_types = ["aws_codepipeline_webhook"]
}

tagtype "codestar-connections:connection" {
  terraform_types = ["aws_codestarconnections_connection"]
}

tagtype "codestar-connections:repository-link" {
  terraform_types = []
}

tagtype "codestar-notifications:notificationrule" {
  terraform_types = ["aws_codestarnotifications_notification_rule"]
}

tagtype "cognito-identity:identitypool" {
  terraform_types = ["aws_cognito_identity_pool"]
}

tagtype "cognito-idp:userpool" {
  terraform_types = ["aws_cognito_user_pool"]
}

tagtype "comprehend:document-classifier" {
  terraform_types = ["aws_comprehend_document_classifier"]
}

tagtype "comprehend:flywheel" {
  terraform_types = []
}

tagtype "config:aggregation-authorization" {
  terraform_types = ["aws_config_aggregate_authorization"]
}

tagtype "config:config-aggregator" {
  terraform_types = ["aws_config_configuration_aggregator"]
}

tagtype "config:config-rule" {
  terraform_types = ["aws_config_config_rule"]
}

tagtype "config:conformance-pack" {
  terraform_types = ["aws_config_conformance_pack"]
}

tagtype "config:stored-query" {
  terraform_types = []
}

tagtype "connect-campaigns:campaign" {
  terraform_types = []
}

tagtype "connect:instance" {
  terraform_types = ["aws_connect_instance"]
}

tagtype "connect:instance/agent" {
  terraform_types = ["aws_connect_user"]
}

tagtype "connect:instance/contact-flow" {
  terraform_types = ["aws_connect_contact_flow"]
}

tagtype "connect:instance/evaluation-form" {
  terraform_types = []
}

tagtype "connect:instance/flow-module" {
  terraform_types = ["aws_connect_contact_flow_module"]
}

tagtype "connect:instance/integration-association" {
  terraform_types = []
}

tagtype "connect:instance/operating-hours" {
  terraform_types = ["aws_connect_hours_of_operation"]
}

tagtype "connect:instance/prompt" {
  terraform_types = []
}

tagtype "connect:instance/queue" {
  terraform_types = ["aws_connect_queue"]
}

tagtype "connect:instance/routing-profile" {
  terraform_types = ["aws_connect_routing_profile"]
}

tagtype "connect:instance/rule" {
  terraform_types = []
}

tagtype "connect:instance/security-profile" {
  terraform_types = ["aws_connect_security_profile"]
}

tagtype "connect:instance/task-template" {
  terraform_types = []
}

tagtype "connect:instance/transfer-destination" {
  terraform_types = ["aws_connect_quick_connect"]
}

tagtype "connect:phone-number" {
  terraform_types = ["aws_connect_phone_number"]
}

tagtype "cur:definition" {
  terraform_types = ["aws_cur_report_definition"]
}

tagtype "databrew:dataset" {
  terraform_types = []
}

tagtype "databrew:job" {
  terraform_types = []
}

tagtype "databrew:project" {
  terraform_types = []
}

tagtype "databrew:recipe" {
  terraform_types = []
}

tagtype "databrew:ruleset" {
  terraform_types = []
}

tagtype "databrew:schedule" {
  terraform_types = []
}

tagtype "datasync:task" {
  terraform_types = ["aws_datasync_task"]
}

tagtype "datazone:domain" {
  terraform_types = ["aws_datazone_domain"]
}

tagtype "dax:cache" {
  terraform_types = ["aws_dax_cluster"]
}

tagtype "deadline:farm" {
  terraform_types = []
}

tagtype "deadline:license-endpoint" {
  terraform_types = []
}

tagtype "detective:graph" {
  terraform_types = ["aws_detective_graph"]
}

tagtype "devicefarm:instanceprofile" {
  terraform_types = ["aws_devicefarm_instance_profile"]
}

tagtype "devicefarm:project" {
  terraform_types = ["aws_devicefarm_project"]
}

tagtype "devicefarm:testgrid-project" {
  terraform_types = ["aws_devicefarm_test_grid_project"]
}

tagtype "dlm:policy" {
  terraform_types = ["aws_dlm_lifecycle_policy"]
}

tagtype "dms:cert" {
  terraform_types = ["aws_dms_certificate"]
}

tagtype "dms:data-provider" {
  terraform_types = []
}

tagtype "dms:endpoint" {
  terraform_types = ["aws_dms_endpoint"]
}

tagtype "dms:es" {
  terraform_types = ["aws_dms_event_subscription"]
}

tagtype "dms:instance-profile" {
  terraform_types = []
}

tagtype "dms:migration-project" {
  terraform_types = []
}

tagtype "dms:rep" {
  terraform_types = ["aws_dms_replication_instance"]
}

tagtype "dms:replication-config" {
  terraform_types = ["aws_dms_replication_config"]
}

tagtype "dms:subgrp" {
  terraform_types = ["aws_dms_replication_subnet_group"]
}

tagtype "dms:task" {
  terraform_types = ["aws_dms_replication_task"]
}

tagtype "dsql:cluster" {
  terraform_types = ["aws_dsql_cluster"]
}

tagtype "dynamodb:table" {
  terraform_types = ["aws_dynamodb_table"]
}

tagtype "ec2:capacity-reservation" {
  terraform_types = ["aws_ec2_capacity_reservation"]
}

tagtype "ec2:capacity-reservation-fleet" {
  terraform_types = []
}

tagtype "ec2:carrier-gateway" {
  terraform_types = ["aws_ec2_carrier_gateway"]
}

tagtype "ec2:client-vpn-endpoint" {
  terraform_types = ["aws_ec2_client_vpn_endpoint"]
}

tagtype "ec2:customer-gateway" {
  terraform_types = ["aws_customer_gateway"]
}

tagtype "ec2:dedicated-host" {
  terraform_types = ["aws_ec2_host"]
}

tagtype "ec2:dhcp-options" {
  terraform_types = ["aws_vpc_dhcp_options"]
}

tagtype "ec2:egress-only-internet-gateway" {
  terraform_types = ["aws_egress_only_internet_gateway"]
}

tagtype "ec2:elastic-ip" {
  terraform_types = ["aws_eip"]
}

tagtype "ec2:fleet" {
  terraform_types = ["aws_ec2_fleet"]
}

tagtype "ec2:instance" {
  terraform_types = ["aws_instance"]
}

tagtype "ec2:instance-connect-endpoint" {
  terraform_types = ["aws_ec2_instance_connect_endpoint"]
}

tagtype "ec2:internet-gateway" {
  terraform_types = ["aws_internet_gateway"]
}

tagtype "ec2:ipam" {
  terraform_types = ["aws_vpc_ipam"]
}

tagtype "ec2:ipam-pool" {
  terraform_types = ["aws_vpc_ipam_pool"]
}

tagtype "ec2:ipam-resource-discovery" {
  terraform_types = ["aws_vpc_ipam_resource_discovery"]
}

tagtype "ec2:ipam-resource-discovery-association" {
  terraform_types = ["aws_vpc_ipam_resource_discovery_association"]
}

tagtype "ec2:ipam-scope" {
  terraform_types = ["aws_vpc_ipam_scope"]
}

tagtype "ec2:key-pair" {
  terraform_types = ["aws_key_pair"]
}

tagtype "ec2:launch-template" {
  terraform_types = ["aws_launch_template"]
}

tagtype "ec2:local-gateway-route-table" {
  terraform_types = []
}

tagtype "ec2:local-gateway-route-table-virtual-interface-group-association" {
  terraform_types = []
}

tagtype "ec2:local-gateway-route-table-vpc-association" {
  terraform_types = ["aws_ec2_local_gateway_route_table_vpc_association"]
}

tagtype "ec2:natgateway" {
  terraform_types = ["aws_nat_gateway"]
}

tagtype "ec2:network-acl" {
  terraform_types = ["aws_network_acl"]
}

tagtype "ec2:network-insights-access-scope" {
  terraform_types = []
}

tagtype "ec2:network-insights-access-scope-analysis" {
  terraform_types = []
}

tagtype "ec2:network-insights-analysis" {
  terraform_types = ["aws_ec2_network_insights_analysis"]
}

tagtype "ec2:network-insights-path" {
  terraform_types = ["aws_ec2_network_insights_path"]
}

tagtype "ec2:network-interface" {
  terraform_types = ["aws_network_interface"]
}

tagtype "ec2:placement-group" {
  terraform_types = ["aws_placement_group"]
}

tagtype "ec2:prefix-list" {
  terraform_types = ["aws_ec2_managed_prefix_list"]
}

tagtype "ec2:route-table" {
  terraform_types = ["aws_route_table"]
}

tagtype "ec2:security-group" {
  terraform_types = ["aws_security_group"]
}

tagtype "ec2:spot-fleet-request" {
  terraform_types = []
}

tagtype "ec2:subnet" {
  terraform_types = ["aws_subnet"]
}

tagtype "ec2:traffic-mirror-filter" {
  terraform_types = ["aws_ec2_traffic_mirror_filter"]
}

tagtype "ec2:traffic-mirror-filter-rule" {
  terraform_types = ["aws_ec2_traffic_mirror_filter_rule"]
}

tagtype "ec2:traffic-mirror-session" {
  terraform_types = ["aws_ec2_traffic_mirror_session"]
}

tagtype "ec2:traffic-mirror-target" {
  terraform_types = ["aws_ec2_traffic_mirror_target"]
}

tagtype "ec2:transit-gateway" {
  terraform_types = ["aws_ec2_transit_gateway"]
}

tagtype "ec2:transit-gateway-connect-peer" {
  terraform_types = ["aws_ec2_transit_gateway_connect_peer"]
}

tagtype "ec2:transit-gateway-multicast-domain" {
  terraform_types = ["aws_ec2_transit_gateway_multicast_domain"]
}

tagtype "ec2:transit-gateway-route-table" {
  terraform_types = ["aws_ec2_transit_gateway_route_table"]
}

tagtype "ec2:verified-access-endpoint" {
  terraform_types = ["aws_verifiedaccess_endpoint"]
}

tagtype "ec2:verified-access-group" {
  terraform_types = ["aws_verifiedaccess_group"]
}

tagtype "ec2:verified-access-instance" {
  terraform_types = ["aws_verifiedaccess_instance"]
}

tagtype "ec2:verified-access-trust-provider" {
  terraform_types = ["aws_verifiedaccess_trust_provider"]
}

tagtype "ec2:volume" {
  terraform_types = ["aws_ebs_volume"]
}

tagtype "ec2:vpc" {
  terraform_types = ["aws_vpc"]
}

tagtype "ec2:vpc-block-public-access-exclusion" {
  terraform_types = ["aws_vpc_block_public_access_exclusion"]
}

tagtype "ec2:vpc-endpoint" {
  terraform_types = ["aws_vpc_endpoint"]
}

tagtype "ec2:vpc-endpoint-service" {
  terraform_types = ["aws_vpc_endpoint_service"]
}

tagtype "ec2:vpc-endpoint-service-permission" {
  terraform_types = []
}

tagtype "ec2:vpc-flow-log" {
  terraform_types = ["aws_flow_log"]
}

tagtype "ec2:vpc-peering-connection" {
  terraform_types = ["aws_vpc_peering_connection"]
}

tagtype "ec2:vpn-connection" {
  terraform_types = ["aws_vpn_connection"]
}

tagtype "ec2:vpn-gateway" {
  terraform_types = ["aws_vpn_gateway"]
}

tagtype "ecr-public:repository" {
  terraform_types = ["aws_ecrpublic_repository"]
}

tagtype "ecr:repository" {
  terraform_types = ["aws_ecr_repository"]
}

tagtype "ecs:capacity-provider" {
  terraform_types = ["aws_ecs_capacity_provider"]
}

tagtype "ecs:cluster" {
  terraform_types = ["aws_ecs_cluster"]
}

tagtype "ecs:service" {
  terraform_types = ["aws_ecs_service"]
}

tagtype "ecs:task-definition" {
  terraform_types = ["aws_ecs_task_definition"]
}

tagtype "ecs:task-set" {
  terraform_types = ["aws_ecs_task_set"]
}

tagtype "eks:access-entry" {
  terraform_types = ["aws_eks_access_entry"]
}

tagtype "eks:addon" {
  terraform_types = ["aws_eks_addon"]
}

tagtype "eks:cluster" {
  terraform_types = ["aws_eks_cluster"]
}

tagtype "eks:fargateprofile" {
  terraform_types = ["aws_eks_fargate_profile"]
}

tagtype "eks:identityproviderconfig" {
  terraform_types = ["aws_eks_identity_provider_config"]
}

tagtype "eks:nodegroup" {
  terraform_types = ["aws_eks_node_group"]
}

tagtype "eks:podidentityassociation" {
  terraform_types = ["aws_eks_pod_identity_association"]
}

tagtype "elasticache:cluster" {
  terraform_types = ["aws_elasticache_cluster"]
}

tagtype "elasticache:parametergroup" {
  terraform_types = ["aws_elasticache_parameter_group"]
}

tagtype "elasticache:replicationgroup" {
  terraform_types = ["aws_elasticache_replication_group"]
}

tagtype "elasticache:securitygroup" {
  terraform_types = []
}

tagtype "elasticache:subnetgroup" {
  terraform_types = ["aws_elasticache_subnet_group"]
}

tagtype "elasticache:user" {
  terraform_types = ["aws_elasticache_user"]
}

tagtype "elasticache:usergroup" {
  terraform_types = ["aws_elasticache_user_group"]
}

tagtype "elasticbeanstalk:application" {
  terraform_types = ["aws_elastic_beanstalk_application"]
}

tagtype "elasticbeanstalk:applicationversion" {
  terraform_types = ["aws_elastic_beanstalk_application_version"]
}

tagtype "elasticbeanstalk:configurationtemplate" {
  terraform_types = ["aws_elastic_beanstalk_configuration_template"]
}

tagtype "elasticbeanstalk:environment" {
  terraform_types = ["aws_elastic_beanstalk_environment"]
}

tagtype "elasticfilesystem:access-point" {
  terraform_types = ["aws_efs_access_point"]
}

tagtype "elasticfilesystem:file-system" {
  terraform_types = ["aws_efs_file_system"]
}

tagtype "elasticloadbalancing:listener" {
  terraform_types = [
    "aws_alb_listener",
    "aws_lb_listener",
  ]
}

tagtype "elasticloadbalancing:listener-rule" {
  terraform_types = [
    "aws_alb_listener_rule",
    "aws_lb_listener_rule",
  ]
}

tagtype "elasticloadbalancing:loadbalancer" {
  terraform_types = [
    "aws_lb",
    "aws_alb",
  ]
}

tagtype "elasticloadbalancing:targetgroup" {
  terraform_types = [
    "aws_alb_target_group",
    "aws_lb_target_group",
  ]
}

tagtype "elasticloadbalancing:truststore" {
  terraform_types = ["aws_lb_trust_store"]
}

tagtype "elasticmapreduce:cluster" {
  terraform_types = ["aws_emr_cluster"]
}

tagtype "emr-containers:virtualclusters" {
  terraform_types = ["aws_emrcontainers_virtual_cluster"]
}

tagtype "emr-serverless:applications" {
  terraform_types = ["aws_emrserverless_application"]
}

tagtype "entityresolution:idmappingworkflow" {
  terraform_types = []
}

tagtype "entityresolution:idnamespace" {
  terraform_types = []
}

tagtype "entityresolution:matchingworkflow" {
  terraform_types = []
}

tagtype "entityresolution:schemamapping" {
  terraform_types = []
}

tagtype "events:event-bus" {
  terraform_types = ["aws_cloudwatch_event_bus"]
}

tagtype "firehose:deliverystream" {
  terraform_types = ["aws_kinesis_firehose_delivery_stream"]
}

tagtype "fis:experiment-template" {
  terraform_types = ["aws_fis_experiment_template"]
}

tagtype "forecast:dataset" {
  terraform_types = []
}

tagtype "forecast:dataset-group" {
  terraform_types = []
}

tagtype "frauddetector:detector" {
  terraform_types = []
}

tagtype "frauddetector:entity-type" {
  terraform_types = []
}

tagtype "frauddetector:event-type" {
  terraform_types = []
}

tagtype "frauddetector:label" {
  terraform_types = []
}

tagtype "frauddetector:list" {
  terraform_types = []
}

tagtype "frauddetector:outcome" {
  terraform_types = []
}

tagtype "frauddetector:variable" {
  terraform_types = []
}

tagtype "fsx:association" {
  terraform_types = ["aws_fsx_data_repository_association"]
}

tagtype "fsx:file-system" {
  terraform_types = [
    "aws_fsx_lustre_file_system",
    "aws_fsx_ontap_file_system",
    "aws_fsx_openzfs_file_system",
    "aws_fsx_windows_file_system",
  ]
}

tagtype "fsx:snapshot" {
  terraform_types = ["aws_fsx_openzfs_snapshot"]
}

tagtype "fsx:storage-virtual-machine" {
  terraform_types = ["aws_fsx_ontap_storage_virtual_machine"]
}

tagtype "fsx:volume" {
  terraform_types = [
    "aws_fsx_ontap_volume",
    "aws_fsx_openzfs_volume",
  ]
}

tagtype "gamelift:alias" {
  terraform_types = ["aws_gamelift_alias"]
}

tagtype "gamelift:build" {
  terraform_types = ["aws_gamelift_build"]
}

tagtype "gamelift:fleet" {
  terraform_types = ["aws_gamelift_fleet"]
}

tagtype "gamelift:gameservergroup" {
  terraform_types = ["aws_gamelift_game_server_group"]
}

tagtype "gamelift:gamesessionqueue" {
  terraform_types = ["aws_gamelift_game_session_queue"]
}

tagtype "gamelift:location" {
  terraform_types = []
}

tagtype "gamelift:matchmakingconfiguration" {
  terraform_types = []
}

tagtype "gamelift:matchmakingruleset" {
  terraform_types = []
}

tagtype "gamelift:script" {
  terraform_types = ["aws_gamelift_script"]
}

tagtype "geo:api-key" {
  terraform_types = []
}

tagtype "geo:geofence-collection" {
  terraform_types = ["aws_location_geofence_collection"]
}

tagtype "geo:map" {
  terraform_types = ["aws_location_map"]
}

tagtype "geo:place-index" {
  terraform_types = ["aws_location_place_index"]
}

tagtype "geo:route-calculator" {
  terraform_types = ["aws_location_route_calculator"]
}

tagtype "geo:tracker" {
  terraform_types = ["aws_location_tracker"]
}

tagtype "globalaccelerator:accelerator" {
  terraform_types = ["aws_globalaccelerator_accelerator"]
}

tagtype "globalaccelerator:attachment" {
  terraform_types = ["aws_globalaccelerator_cross_account_attachment"]
}

tagtype "glue:connection" {
  terraform_types = ["aws_glue_connection"]
}

tagtype "glue:crawler" {
  terraform_types = ["aws_glue_crawler"]
}

tagtype "glue:customEntityType" {
  terraform_types = []
}

tagtype "glue:dataQualityRuleset" {
  terraform_types = ["aws_glue_data_quality_ruleset"]
}

tagtype "glue:database" {
  terraform_types = ["aws_glue_catalog_database"]
}

tagtype "glue:job" {
  terraform_types = ["aws_glue_job"]
}

tagtype "glue:mlTransform" {
  terraform_types = ["aws_glue_ml_transform"]
}

tagtype "glue:registry" {
  terraform_types = ["aws_glue_registry"]
}

tagtype "glue:schema" {
  terraform_types = ["aws_glue_schema"]
}

tagtype "glue:trigger" {
  terraform_types = ["aws_glue_trigger"]
}

tagtype "glue:usageProfile" {
  terraform_types = []
}

tagtype "grafana:workspaces" {
  terraform_types = ["aws_grafana_workspace"]
}

tagtype "greengrass:connectorsDefinition" {
  terraform_types = []
}

tagtype "greengrass:coresDefinition" {
  terraform_types = []
}

tagtype "greengrass:devicesDefinition" {
  terraform_types = []
}

tagtype "greengrass:functionsDefinition" {
  terraform_types = []
}

tagtype "greengrass:groups" {
  terraform_types = []
}

tagtype "greengrass:loggersDefinition" {
  terraform_types = []
}

tagtype "greengrass:resourcesDefinition" {
  terraform_types = []
}

tagtype "greengrass:subscriptionsDefinition" {
  terraform_types = []
}

tagtype "groundstation:config" {
  terraform_types = []
}

tagtype "groundstation:dataflow-endpoint-group" {
  terraform_types = []
}

tagtype "groundstation:mission-profile" {
  terraform_types = []
}

tagtype "guardduty:detector" {
  terraform_types = ["aws_guardduty_detector"]
}

tagtype "guardduty:detector/filter" {
  terraform_types = ["aws_guardduty_filter"]
}

tagtype "guardduty:detector/ipset" {
  terraform_types = ["aws_guardduty_ipset"]
}

tagtype "guardduty:detector/threatintelset" {
  terraform_types = ["aws_guardduty_threatintelset"]
}

tagtype "guardduty:malware-protection-plan" {
  terraform_types = ["aws_guardduty_malware_protection_plan"]
}

tagtype "healthlake:datastore" {
  terraform_types = []
}

tagtype "iam:instance-profile" {
  terraform_types = ["aws_iam_instance_profile"]
}

tagtype "iam:mfa" {
  terraform_types = ["aws_iam_virtual_mfa_device"]
}

tagtype "iam:oidc-provider" {
  terraform_types = ["aws_iam_openid_connect_provider"]
}

tagtype "iam:role" {
  terraform_types = ["aws_iam_role"]
}

tagtype "iam:saml-provider" {
  terraform_types = ["aws_iam_saml_provider"]
}

tagtype "iam:server-certificate" {
  terraform_types = ["aws_iam_server_certificate"]
}

tagtype "iam:user" {
  terraform_types = ["aws_iam_user"]
}

tagtype "imagebuilder:component" {
  terraform_types = ["aws_imagebuilder_component"]
}

tagtype "imagebuilder:container-recipe" {
  terraform_types = ["aws_imagebuilder_container_recipe"]
}

tagtype "imagebuilder:distribution-configuration" {
  terraform_types = ["aws_imagebuilder_distribution_configuration"]
}

tagtype "imagebuilder:image" {
  terraform_types = ["aws_imagebuilder_image"]
}

tagtype "imagebuilder:image-pipeline" {
  terraform_types = ["aws_imagebuilder_image_pipeline"]
}

tagtype "imagebuilder:image-recipe" {
  terraform_types = ["aws_imagebuilder_image_recipe"]
}

tagtype "imagebuilder:infrastructure-configuration" {
  terraform_types = ["aws_imagebuilder_infrastructure_configuration"]
}

tagtype "imagebuilder:lifecycle-policy" {
  terraform_types = ["aws_imagebuilder_lifecycle_policy"]
}

tagtype "imagebuilder:workflow" {
  terraform_types = ["aws_imagebuilder_workflow"]
}

tagtype "inspector2:filter" {
  terraform_types = ["aws_inspector2_filter"]
}

tagtype "internetmonitor:monitor" {
  terraform_types = ["aws_internetmonitor_monitor"]
}

tagtype "invoicing:invoice-unit" {
  terraform_types = []
}

tagtype "iot:authorizer" {
  terraform_types = ["aws_iot_authorizer"]
}

tagtype "iot:billinggroup" {
  terraform_types = ["aws_iot_billing_group"]
}

tagtype "iot:cacert" {
  terraform_types = ["aws_iot_ca_certificate"]
}

tagtype "iot:certificateprovider" {
  terraform_types = []
}

tagtype "iot:custommetric" {
  terraform_types = []
}

tagtype "iot:dimension" {
  terraform_types = []
}

tagtype "iot:fleetmetric" {
  terraform_types = []
}

tagtype "iot:jobtemplate" {
  terraform_types = []
}

tagtype "iot:mitigationaction" {
  terraform_types = []
}

tagtype "iot:package" {
  terraform_types = []
}

tagtype "iot:policy" {
  terraform_types = ["aws_iot_policy"]
}

tagtype "iot:provisioningtemplate" {
  terraform_types = ["aws_iot_provisioning_template"]
}

tagtype "iot:rolealias" {
  terraform_types = ["aws_iot_role_alias"]
}

tagtype "iot:rule" {
  terraform_types = ["aws_iot_topic_rule"]
}

tagtype "iot:scheduledaudit" {
  terraform_types = []
}

tagtype "iot:securityprofile" {
  terraform_types = []
}

tagtype "iot:thinggroup" {
  terraform_types = ["aws_iot_thing_group"]
}

tagtype "iot:thingtype" {
  terraform_types = ["aws_iot_thing_type"]
}

tagtype "iotdeviceadvisor:suitedefinition" {
  terraform_types = []
}

tagtype "iotfleethub:application" {
  terraform_types = []
}

tagtype "iotfleetwise:campaign" {
  terraform_types = []
}

tagtype "iotfleetwise:decoder-manifest" {
  terraform_types = []
}

tagtype "iotfleetwise:fleet" {
  terraform_types = []
}

tagtype "iotfleetwise:model-manifest" {
  terraform_types = []
}

tagtype "iotfleetwise:signal-catalog" {
  terraform_types = []
}

tagtype "iotfleetwise:vehicle" {
  terraform_types = []
}

tagtype "iotsitewise:access-policy" {
  terraform_types = []
}

tagtype "iotsitewise:asset" {
  terraform_types = []
}

tagtype "iotsitewise:asset-model" {
  terraform_types = []
}

tagtype "iotsitewise:dashboard" {
  terraform_types = []
}

tagtype "iotsitewise:dataset" {
  terraform_types = []
}

tagtype "iotsitewise:gateway" {
  terraform_types = []
}

tagtype "iotsitewise:portal" {
  terraform_types = []
}

tagtype "iotsitewise:project" {
  terraform_types = []
}

tagtype "iotwireless:Destination" {
  terraform_types = []
}

tagtype "iotwireless:DeviceProfile" {
  terraform_types = []
}

tagtype "iotwireless:FuotaTask" {
  terraform_types = []
}

tagtype "iotwireless:ImportTask" {
  terraform_types = []
}

tagtype "iotwireless:MulticastGroup" {
  terraform_types = []
}

tagtype "iotwireless:NetworkAnalyzerConfiguration" {
  terraform_types = []
}

tagtype "iotwireless:ServiceProfile" {
  terraform_types = []
}

tagtype "iotwireless:SidewalkAccount" {
  terraform_types = []
}

tagtype "iotwireless:WirelessDevice" {
  terraform_types = []
}

tagtype "iotwireless:WirelessGateway" {
  terraform_types = []
}

tagtype "iotwireless:WirelessGatewayTaskDefinition" {
  terraform_types = []
}

tagtype "ivs:channel" {
  terraform_types = ["aws_ivs_channel"]
}

tagtype "ivs:encoder-configuration" {
  terraform_types = []
}

tagtype "ivs:ingest-configuration" {
  terraform_types = []
}

tagtype "ivs:playback-key" {
  terraform_types = ["aws_ivs_playback_key_pair"]
}

tagtype "ivs:playback-restriction-policy" {
  terraform_types = []
}

tagtype "ivs:public-key" {
  terraform_types = []
}

tagtype "ivs:recording-configuration" {
  terraform_types = ["aws_ivs_recording_configuration"]
}

tagtype "ivs:stage" {
  terraform_types = []
}

tagtype "ivs:storage-configuration" {
  terraform_types = []
}

tagtype "ivs:stream-key" {
  terraform_types = []
}

tagtype "kafka:replicator" {
  terraform_types = ["aws_msk_replicator"]
}

tagtype "kafkaconnect:custom-plugin" {
  terraform_types = ["aws_mskconnect_custom_plugin"]
}

tagtype "kafkaconnect:worker-configuration" {
  terraform_types = ["aws_mskconnect_worker_configuration"]
}

tagtype "kendra-ranking:rescore-execution-plan" {
  terraform_types = []
}

tagtype "kendra:index" {
  terraform_types = ["aws_kendra_index"]
}

tagtype "kendra:index/data-source" {
  terraform_types = ["aws_kendra_data_source"]
}

tagtype "kinesis:stream" {
  terraform_types = ["aws_kinesis_stream"]
}

tagtype "kinesis:stream/consumer" {
  terraform_types = ["aws_kinesis_stream_consumer"]
}

tagtype "kinesisanalytics:application" {
  terraform_types = ["aws_kinesisanalyticsv2_application"]
}

tagtype "kinesisvideo:channel" {
  terraform_types = []
}

tagtype "kinesisvideo:stream" {
  terraform_types = ["aws_kinesis_video_stream"]
}

tagtype "kms:key" {
  terraform_types = ["aws_kms_key"]
}

tagtype "lambda:code-signing-config" {
  terraform_types = ["aws_lambda_code_signing_config"]
}

tagtype "lambda:event-source-mapping" {
  terraform_types = ["aws_lambda_event_source_mapping"]
}

tagtype "lambda:function" {
  terraform_types = ["aws_lambda_function"]
}

tagtype "lex:bot" {
  terraform_types = ["aws_lex_bot"]
}

tagtype "lex:bot-alias" {
  terraform_types = ["aws_lex_bot_alias"]
}

tagtype "license-manager:grant" {
  terraform_types = ["aws_licensemanager_grant"]
}

tagtype "license-manager:license" {
  terraform_types = []
}

tagtype "lightsail:Bucket" {
  terraform_types = ["aws_lightsail_bucket"]
}

tagtype "lightsail:Certificate" {
  terraform_types = ["aws_lightsail_certificate"]
}

tagtype "lightsail:ContainerService" {
  terraform_types = ["aws_lightsail_container_service"]
}

tagtype "lightsail:Disk" {
  terraform_types = ["aws_lightsail_disk"]
}

tagtype "lightsail:Distribution" {
  terraform_types = ["aws_lightsail_distribution"]
}

tagtype "lightsail:Domain" {
  terraform_types = ["aws_lightsail_domain"]
}

tagtype "lightsail:Instance" {
  terraform_types = ["aws_lightsail_instance"]
}

tagtype "lightsail:LoadBalancer" {
  terraform_types = ["aws_lightsail_lb"]
}

tagtype "lightsail:StaticIp" {
  terraform_types = ["aws_lightsail_static_ip"]
}

tagtype "logs:anomaly-detector" {
  terraform_types = ["aws_cloudwatch_log_anomaly_detector"]
}

tagtype "logs:delivery" {
  terraform_types = ["aws_cloudwatch_log_delivery"]
}

tagtype "logs:delivery-destination" {
  terraform_types = ["aws_cloudwatch_log_delivery_destination"]
}

tagtype "logs:delivery-source" {
  terraform_types = ["aws_cloudwatch_log_delivery_source"]
}

tagtype "logs:destination" {
  terraform_types = ["aws_cloudwatch_log_destination"]
}

tagtype "logs:log-group" {
  terraform_types = ["aws_cloudwatch_log_group"]
}

tagtype "lookoutmetrics:Alert" {
  terraform_types = []
}

tagtype "lookoutmetrics:AnomalyDetector" {
  terraform_types = []
}

tagtype "m2:env" {
  terraform_types = []
}

tagtype "managedblockchain:accessors" {
  terraform_types = []
}

tagtype "mediaconnect:entitlement" {
  terraform_types = []
}

tagtype "mediaconnect:flow" {
  terraform_types = []
}

tagtype "mediaconnect:output" {
  terraform_types = []
}

tagtype "mediaconnect:source" {
  terraform_types = []
}

tagtype "mediaconvert:presets" {
  terraform_types = []
}

tagtype "mediaconvert:queues" {
  terraform_types = []
}

tagtype "medialive:channelPlacementGroup" {
  terraform_types = []
}

tagtype "medialive:cloudwatch-alarm-template" {
  terraform_types = []
}

tagtype "medialive:cloudwatch-alarm-template-group" {
  terraform_types = []
}

tagtype "medialive:eventbridge-rule-template" {
  terraform_types = []
}

tagtype "medialive:eventbridge-rule-template-group" {
  terraform_types = []
}

tagtype "medialive:inputSecurityGroup" {
  terraform_types = ["aws_medialive_input_security_group"]
}

tagtype "medialive:multiplex" {
  terraform_types = ["aws_medialive_multiplex"]
}

tagtype "medialive:network" {
  terraform_types = []
}

tagtype "medialive:sdiSource" {
  terraform_types = []
}

tagtype "medialive:signal-map" {
  terraform_types = []
}

tagtype "mediapackage-vod:assets" {
  terraform_types = []
}

tagtype "mediapackage-vod:packaging-configurations" {
  terraform_types = []
}

tagtype "mediapackage-vod:packaging-groups" {
  terraform_types = []
}

tagtype "mediapackage:channels" {
  terraform_types = []
}

tagtype "mediapackage:origin_endpoints" {
  terraform_types = []
}

tagtype "mediapackagev2:channelGroup" {
  terraform_types = []
}

tagtype "mediapackagev2:channelGroup/channel" {
  terraform_types = []
}

tagtype "mediapackagev2:channelGroup/channel/originEndpoint" {
  terraform_types = []
}

tagtype "mediatailor:channel" {
  terraform_types = []
}

tagtype "mediatailor:liveSource" {
  terraform_types = []
}

tagtype "mediatailor:playbackConfiguration" {
  terraform_types = []
}

tagtype "mediatailor:sourceLocation" {
  terraform_types = []
}

tagtype "mediatailor:vodSource" {
  terraform_types = []
}

tagtype "medical-imaging:datastore" {
  terraform_types = []
}

tagtype "memorydb:acl" {
  terraform_types = ["aws_memorydb_acl"]
}

tagtype "memorydb:cluster" {
  terraform_types = ["aws_memorydb_cluster"]
}

tagtype "memorydb:parametergroup" {
  terraform_types = ["aws_memorydb_parameter_group"]
}

tagtype "memorydb:subnetgroup" {
  terraform_types = ["aws_memorydb_subnet_group"]
}

tagtype "memorydb:user" {
  terraform_types = ["aws_memorydb_user"]
}

tagtype "mobiletargeting:apps" {
  terraform_types = ["aws_pinpoint_app"]
}

tagtype "mq:broker" {
  terraform_types = ["aws_mq_broker"]
}

tagtype "mq:configuration" {
  terraform_types = ["aws_mq_configuration"]
}

tagtype "network-firewall:firewall" {
  terraform_types = ["aws_networkfirewall_firewall"]
}

tagtype "network-firewall:firewall-policy" {
  terraform_types = ["aws_networkfirewall_firewall_policy"]
}

tagtype "network-firewall:stateless-rulegroup" {
  terraform_types = ["aws_networkfirewall_rule_group"]
}

tagtype "networkmanager:connect-peer" {
  terraform_types = ["aws_networkmanager_connect_peer"]
}

tagtype "networkmanager:core-network" {
  terraform_types = ["aws_networkmanager_core_network"]
}

tagtype "networkmanager:device" {
  terraform_types = ["aws_networkmanager_device"]
}

tagtype "networkmanager:global-network" {
  terraform_types = ["aws_networkmanager_global_network"]
}

tagtype "networkmanager:link" {
  terraform_types = ["aws_networkmanager_link"]
}

tagtype "networkmanager:site" {
  terraform_types = ["aws_networkmanager_site"]
}

tagtype "notifications-contacts:emailcontact" {
  terraform_types = ["aws_notificationscontacts_email_contact"]
}

tagtype "oam:sink" {
  terraform_types = ["aws_oam_sink"]
}

tagtype "omics:referenceStore" {
  terraform_types = []
}

tagtype "omics:runGroup" {
  terraform_types = []
}

tagtype "omics:sequenceStore" {
  terraform_types = []
}

tagtype "omics:workflow" {
  terraform_types = []
}

tagtype "organizations:account" {
  terraform_types = ["aws_organizations_account"]
}

tagtype "organizations:ou" {
  terraform_types = ["aws_organizations_organizational_unit"]
}

tagtype "organizations:resourcepolicy" {
  terraform_types = ["aws_organizations_resource_policy"]
}

tagtype "osis:pipeline" {
  terraform_types = ["aws_osis_pipeline"]
}

tagtype "panorama:package" {
  terraform_types = []
}

tagtype "payment-cryptography:key" {
  terraform_types = ["aws_paymentcryptography_key"]
}

tagtype "personalize:dataset" {
  terraform_types = []
}

tagtype "personalize:dataset-group" {
  terraform_types = []
}

tagtype "personalize:solution" {
  terraform_types = []
}

tagtype "pipes:pipe" {
  terraform_types = ["aws_pipes_pipe"]
}

tagtype "profile:domains" {
  terraform_types = ["aws_customerprofiles_domain"]
}

tagtype "profile:domains/integrations" {
  terraform_types = []
}

tagtype "profile:domains/object-types" {
  terraform_types = []
}

tagtype "proton:environment-account-connection" {
  terraform_types = []
}

tagtype "proton:environment-template" {
  terraform_types = []
}

tagtype "proton:service-template" {
  terraform_types = []
}

tagtype "ram:resource-share" {
  terraform_types = ["aws_ram_resource_share"]
}

tagtype "rbin:rule" {
  terraform_types = ["aws_rbin_rule"]
}

tagtype "rds:cev" {
  terraform_types = ["aws_rds_custom_db_engine_version"]
}

tagtype "rds:cluster" {
  terraform_types = [
    "aws_docdb_cluster",
    "aws_neptune_cluster",
    "aws_rds_cluster",
  ]
}

tagtype "rds:db" {
  terraform_types = [
    "aws_db_instance",
    "aws_docdb_cluster_instance",
    "aws_neptune_cluster_instance",
  ]
}

tagtype "rds:db-proxy" {
  terraform_types = ["aws_db_proxy"]
}

tagtype "rds:db-proxy-endpoint" {
  terraform_types = ["aws_db_proxy_endpoint"]
}

tagtype "rds:es" {
  terraform_types = ["aws_db_event_subscription"]
}

tagtype "rds:global-cluster" {
  terraform_types = [
    "aws_docdb_global_cluster",
    "aws_neptune_global_cluster",
    "aws_rds_global_cluster",
  ]
}

tagtype "rds:og" {
  terraform_types = ["aws_db_option_group"]
}

tagtype "rds:pg" {
  terraform_types = ["aws_db_parameter_group"]
}

tagtype "rds:secgrp" {
  terraform_types = []
}

tagtype "rds:subgrp" {
  terraform_types = ["aws_db_subnet_group"]
}

tagtype "rds:target-group" {
  terraform_types = []
}

tagtype "redshift-serverless:namespace" {
  terraform_types = ["aws_redshiftserverless_namespace"]
}

tagtype "redshift-serverless:workgroup" {
  terraform_types = ["aws_redshiftserverless_workgroup"]
}

tagtype "redshift:cluster" {
  terraform_types = ["aws_redshift_cluster"]
}

tagtype "redshift:eventsubscription" {
  terraform_types = ["aws_redshift_event_subscription"]
}

tagtype "redshift:integration" {
  terraform_types = ["aws_redshift_integration"]
}

tagtype "redshift:parametergroup" {
  terraform_types = ["aws_redshift_parameter_group"]
}

tagtype "redshift:subnetgroup" {
  terraform_types = ["aws_redshift_subnet_group"]
}

tagtype "refactor-spaces:environment" {
  terraform_types = []
}

tagtype "refactor-spaces:environment/application" {
  terraform_types = []
}

tagtype "refactor-spaces:environment/application/route" {
  terraform_types = []
}

tagtype "refactor-spaces:environment/application/service" {
  terraform_types = []
}

tagtype "rekognition:collection" {
  terraform_types = []
}

tagtype "resiliencehub:app" {
  terraform_types = []
}

tagtype "resiliencehub:resiliency-policy" {
  terraform_types = ["aws_resiliencehub_resiliency_policy"]
}

tagtype "resource-groups:group" {
  terraform_types = ["aws_resourcegroups_group"]
}

tagtype "robomaker:robot-application" {
  terraform_types = []
}

tagtype "route53-recovery-control:cluster" {
  terraform_types = ["aws_route53recoverycontrolconfig_cluster"]
}

tagtype "route53-recovery-control:controlpanel" {
  terraform_types = ["aws_route53recoverycontrolconfig_control_panel"]
}

tagtype "route53-recovery-control:controlpanel/safetyrule" {
  terraform_types = ["aws_route53recoverycontrolconfig_safety_rule"]
}

tagtype "route53-recovery-readiness:cell" {
  terraform_types = ["aws_route53recoveryreadiness_cell"]
}

tagtype "route53-recovery-readiness:readiness-check" {
  terraform_types = ["aws_route53recoveryreadiness_readiness_check"]
}

tagtype "route53-recovery-readiness:recovery-group" {
  terraform_types = ["aws_route53recoveryreadiness_recovery_group"]
}

tagtype "route53-recovery-readiness:resource-set" {
  terraform_types = ["aws_route53recoveryreadiness_resource_set"]
}

tagtype "route53:healthcheck" {
  terraform_types = ["aws_route53_health_check"]
}

tagtype "route53:hostedzone" {
  terraform_types = ["aws_route53_zone"]
}

tagtype "route53profiles:profile" {
  terraform_types = ["aws_route53profiles_profile"]
}

tagtype "route53profiles:profile-association" {
  terraform_types = ["aws_route53profiles_association"]
}

tagtype "route53resolver:firewall-domain-list" {
  terraform_types = ["aws_route53_resolver_firewall_domain_list"]
}

tagtype "route53resolver:firewall-rule-group" {
  terraform_types = ["aws_route53_resolver_firewall_rule_group"]
}

tagtype "route53resolver:firewall-rule-group-association" {
  terraform_types = ["aws_route53_resolver_firewall_rule_group_association"]
}

tagtype "route53resolver:resolver-endpoint" {
  terraform_types = ["aws_route53_resolver_endpoint"]
}

tagtype "route53resolver:resolver-query-log-config" {
  terraform_types = ["aws_route53_resolver_query_log_config"]
}

tagtype "route53resolver:resolver-rule" {
  terraform_types = ["aws_route53_resolver_rule"]
}

tagtype "rum:appmonitor" {
  terraform_types = ["aws_rum_app_monitor"]
}

tagtype "s3:accesspoint" {
  terraform_types = ["aws_s3_access_point"]
}

tagtype "s3:bucket" {
  terraform_types = ["aws_s3_bucket"]
}

tagtype "s3:storage-lens" {
  terraform_types = []
}

tagtype "s3:storage-lens-group" {
  terraform_types = []
}

tagtype "s3express:bucket" {
  terraform_types = ["aws_s3_directory_bucket"]
}

tagtype "sagemaker:app" {
  terraform_types = ["aws_sagemaker_app"]
}

tagtype "sagemaker:app-image-config" {
  terraform_types = ["aws_sagemaker_app_image_config"]
}

tagtype "sagemaker:cluster" {
  terraform_types = []
}

tagtype "sagemaker:code-repository" {
  terraform_types = ["aws_sagemaker_code_repository"]
}

tagtype "sagemaker:data-quality-job-definition" {
  terraform_types = ["aws_sagemaker_data_quality_job_definition"]
}

tagtype "sagemaker:domain" {
  terraform_types = ["aws_sagemaker_domain"]
}

tagtype "sagemaker:endpoint" {
  terraform_types = ["aws_sagemaker_endpoint"]
}

tagtype "sagemaker:endpoint-config" {
  terraform_types = ["aws_sagemaker_endpoint_configuration"]
}

tagtype "sagemaker:feature-group" {
  terraform_types = ["aws_sagemaker_feature_group"]
}

tagtype "sagemaker:image" {
  terraform_types = ["aws_sagemaker_image"]
}

tagtype "sagemaker:inference-component" {
  terraform_types = []
}

tagtype "sagemaker:inference-experiment" {
  terraform_types = []
}

tagtype "sagemaker:mlflow-tracking-server" {
  terraform_types = ["aws_sagemaker_mlflow_tracking_server"]
}

tagtype "sagemaker:model" {
  terraform_types = ["aws_sagemaker_model"]
}

tagtype "sagemaker:model-bias-job-definition" {
  terraform_types = []
}

tagtype "sagemaker:model-card" {
  terraform_types = []
}

tagtype "sagemaker:model-explainability-job-definition" {
  terraform_types = []
}

tagtype "sagemaker:model-package" {
  terraform_types = []
}

tagtype "sagemaker:model-package-group" {
  terraform_types = ["aws_sagemaker_model_package_group"]
}

tagtype "sagemaker:model-quality-job-definition" {
  terraform_types = []
}

tagtype "sagemaker:monitoring-schedule" {
  terraform_types = ["aws_sagemaker_monitoring_schedule"]
}

tagtype "sagemaker:notebook-instance" {
  terraform_types = ["aws_sagemaker_notebook_instance"]
}

tagtype "sagemaker:notebook-instance-lifecycle-config" {
  terraform_types = ["aws_sagemaker_notebook_instance_lifecycle_configuration"]
}

tagtype "sagemaker:pipeline" {
  terraform_types = ["aws_sagemaker_pipeline"]
}

tagtype "sagemaker:processing-job" {
  terraform_types = []
}

tagtype "sagemaker:project" {
  terraform_types = ["aws_sagemaker_project"]
}

tagtype "sagemaker:space" {
  terraform_types = ["aws_sagemaker_space"]
}

tagtype "sagemaker:studio-lifecycle-config" {
  terraform_types = ["aws_sagemaker_studio_lifecycle_config"]
}

tagtype "sagemaker:user-profile" {
  terraform_types = ["aws_sagemaker_user_profile"]
}

tagtype "sagemaker:workteam" {
  terraform_types = ["aws_sagemaker_workteam"]
}

tagtype "scheduler:schedule-group" {
  terraform_types = ["aws_scheduler_schedule_group"]
}

tagtype "schemas:discoverer" {
  terraform_types = ["aws_schemas_discoverer"]
}

tagtype "schemas:registry" {
  terraform_types = ["aws_schemas_registry"]
}

tagtype "schemas:schema" {
  terraform_types = ["aws_schemas_schema"]
}

tagtype "secretsmanager:secret" {
  terraform_types = ["aws_secretsmanager_secret"]
}

tagtype "securityhub:hubv2" {
  terraform_types = []
}

tagtype "servicecatalog:applications" {
  terraform_types = ["aws_servicecatalogappregistry_application"]
}

tagtype "servicecatalog:attribute-groups" {
  terraform_types = ["aws_servicecatalogappregistry_attribute_group"]
}

tagtype "servicediscovery:service" {
  terraform_types = ["aws_service_discovery_service"]
}

tagtype "ses:configuration-set" {
  terraform_types = ["aws_sesv2_configuration_set"]
}

tagtype "ses:contact-list" {
  terraform_types = ["aws_sesv2_contact_list"]
}

tagtype "ses:dedicated-ip-pool" {
  terraform_types = ["aws_sesv2_dedicated_ip_pool"]
}

tagtype "ses:identity" {
  terraform_types = ["aws_sesv2_email_identity"]
}

tagtype "ses:mailmanager-archive" {
  terraform_types = []
}

tagtype "ses:mailmanager-ingress-point" {
  terraform_types = []
}

tagtype "ses:mailmanager-rule-set" {
  terraform_types = []
}

tagtype "ses:mailmanager-traffic-policy" {
  terraform_types = []
}

tagtype "signer:signing-profiles" {
  terraform_types = ["aws_signer_signing_profile"]
}

tagtype "sns:topic" {
  terraform_types = ["aws_sns_topic"]
}

tagtype "sqs:queue" {
  terraform_types = ["aws_sqs_queue"]
}

tagtype "ssm-incidents:replication-set" {
  terraform_types = ["aws_ssmincidents_replication_set"]
}

tagtype "ssm-incidents:response-plan" {
  terraform_types = ["aws_ssmincidents_response_plan"]
}

tagtype "ssm:association" {
  terraform_types = ["aws_ssm_association"]
}

tagtype "ssm:document" {
  terraform_types = ["aws_ssm_document"]
}

tagtype "ssm:maintenancewindow" {
  terraform_types = ["aws_ssm_maintenance_window"]
}

tagtype "ssm:parameter" {
  terraform_types = ["aws_ssm_parameter"]
}

tagtype "ssm:patchbaseline" {
  terraform_types = ["aws_ssm_patch_baseline"]
}

tagtype "states:activity" {
  terraform_types = ["aws_sfn_activity"]
}

tagtype "states:stateMachine" {
  terraform_types = ["aws_sfn_state_machine"]
}

tagtype "synthetics:canary" {
  terraform_types = ["aws_synthetics_canary"]
}

tagtype "synthetics:group" {
  terraform_types = ["aws_synthetics_group"]
}

tagtype "timestream:database" {
  terraform_types = ["aws_timestreamwrite_database"]
}

tagtype "timestream:database/table" {
  terraform_types = ["aws_timestreamwrite_table"]
}

tagtype "timestream:scheduled-query" {
  terraform_types = ["aws_timestreamquery_scheduled_query"]
}

tagtype "transfer:agreement" {
  terraform_types = ["aws_transfer_agreement"]
}

tagtype "transfer:certificate" {
  terraform_types = ["aws_transfer_certificate"]
}

tagtype "transfer:connector" {
  terraform_types = ["aws_transfer_connector"]
}

tagtype "transfer:profile" {
  terraform_types = ["aws_transfer_profile"]
}

tagtype "transfer:server" {
  terraform_types = ["aws_transfer_server"]
}

tagtype "transfer:user" {
  terraform_types = ["aws_transfer_user"]
}

tagtype "transfer:workflow" {
  terraform_types = ["aws_transfer_workflow"]
}

tagtype "verifiedpermissions:policy-store" {
  terraform_types = ["aws_verifiedpermissions_policy_store"]
}

tagtype "vpc-lattice:accesslogsubscription" {
  terraform_types = ["aws_vpclattice_access_log_subscription"]
}

tagtype "vpc-lattice:service" {
  terraform_types = ["aws_vpclattice_service"]
}

tagtype "vpc-lattice:service/listener" {
  terraform_types = ["aws_vpclattice_listener"]
}

tagtype "vpc-lattice:service/listener/rule" {
  terraform_types = ["aws_vpclattice_listener_rule"]
}

tagtype "vpc-lattice:servicenetwork" {
  terraform_types = ["aws_vpclattice_service_network"]
}

tagtype "vpc-lattice:servicenetworkserviceassociation" {
  terraform_types = ["aws_vpclattice_service_network_service_association"]
}

tagtype "vpc-lattice:servicenetworkvpcassociation" {
  terraform_types = ["aws_vpclattice_service_network_vpc_association"]
}

tagtype "vpc-lattice:targetgroup" {
  terraform_types = ["aws_vpclattice_target_group"]
}

tagtype "wisdom:ai-agent" {
  terraform_types = []
}

tagtype "wisdom:assistant" {
  terraform_types = []
}

tagtype "wisdom:association" {
  terraform_types = []
}

tagtype "wisdom:knowledge-base" {
  terraform_types = []
}

tagtype "workspaces-web:browserSettings" {
  terraform_types = ["aws_workspacesweb_browser_settings"]
}

tagtype "workspaces-web:ipAccessSettings" {
  terraform_types = ["aws_workspacesweb_ip_access_settings"]
}

tagtype "workspaces-web:networkSettings" {
  terraform_types = ["aws_workspacesweb_network_settings"]
}

tagtype "workspaces-web:portal" {
  terraform_types = ["aws_workspacesweb_portal"]
}

tagtype "workspaces-web:trustStore" {
  terraform_types = ["aws_workspacesweb_trust_store"]
}

tagtype "workspaces-web:userAccessLoggingSettings" {
  terraform_types = ["aws_workspacesweb_user_access_logging_settings"]
}

tagtype "workspaces-web:userSettings" {
  terraform_types = ["aws_workspacesweb_user_settings"]
}

tagtype "workspaces:connectionalias" {
  terraform_types = ["aws_workspaces_connection_alias"]
}

tagtype "workspaces:workspacespool" {
  terraform_types = []
}

tagtype "xray:group" {
  terraform_types = ["aws_xray_group"]
}

tagtype "xray:sampling-rule" {
  terraform_types = ["aws_xray_sampling_rule"]
}

