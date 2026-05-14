# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider_source = "registry.terraform.io/hashicorp/aws"
provider_dir    = "."
schema_json     = "terraform-providers-schema/schema.json"

ignore_file_missing = [
  "aws_alb",
  "aws_alb_listener",
  "aws_alb_listener_certificate",
  "aws_alb_listener_rule",
  "aws_alb_target_group",
  "aws_alb_target_group_attachment",
  "aws_alb_trust_store",
  "aws_alb_trust_store_revocation",
  "aws_albs",
]

file_aliases = {
  "list_resource/aws_ebs_volume" = "aws_ec2_ebs_volume"
}

ignore_contents_check = [
  "data_source/aws_kms_secret",
]

check "schema_docs" {
  enabled = true

  # Sub-check toggles
  coverage    = true
  ordering    = false
  description = false
  heading     = true
  format      = true
  labels      = true

  block_heading_styles = [
    "`{Parent}` `{Block}` Block",
    "`{Block}` Block",
    "{Block} Block",
    "{Block} block",
    "{Block} Configuration Block",
    "{Block} Argument Reference",
    "{Block} Attribute Reference",
    "{Title} Arguments",
    "{Title} Argument Reference",
    "{Title} Attribute Reference",
    "`{Block}`",
    "{Block}",
    "{Title}",
  ]

  preferred_block_heading_styles = [
    "`{Parent}` `{Block}` Block",
    "`{Block}` Block",
  ]

  prefixes = ["aws_appmesh"]
}

check "import_section" {
  enabled = true
  require_identity_section = true

  prefixes = [
    "aws_accessanalyzer",
    "aws_account",
    "aws_alb",
    "aws_ami",
    "aws_amplify",
    "aws_api",
    "aws_apigatewayv2",
    "aws_app",
    "aws_appautoscaling",
    "aws_appconfig",
    "aws_appfabric",
    "aws_appflow",
    "aws_appintegrations",
    "aws_applicationinsights",
    "aws_appmesh",
    "aws_apprunner",
    "aws_appstream",
    "aws_appsync",
    "aws_arczonalshift",
    "aws_arn",
    "aws_athena",
    "aws_autoscaling",
    "aws_autoscalingplans",
    "aws_availability",
    "aws_batch",
    "aws_bcmdataexports",
    "aws_bedrockagent",
    "aws_bedrockagentcore",
    "aws_billing",
    "aws_budgets",
    "aws_caller",
    "aws_canonical",
    "aws_ce",
    "aws_chatbot",
    "aws_chime",
    "aws_chimesdkmediapipelines",
    "aws_chimesdkvoice",
    "aws_cleanrooms",
    "aws_cloud9",
    "aws_cloudcontrolapi",
    "aws_cloudformation",
    "aws_cloudfront",
    "aws_cloudfrontkeyvaluestore",
    "aws_cloudhsm",
    "aws_cloudsearch",
    "aws_codecatalyst",
    "aws_codecommit",
    "aws_codeconnections",
    "aws_codedeploy",
    "aws_codeguruprofiler",
    "aws_codegurureviewer",
    "aws_codepipeline",
    "aws_codestarconnections",
    "aws_codestarnotifications",
    "aws_cognito",
    "aws_comprehend",
    "aws_computeoptimizer",
    "aws_config",
    "aws_connect",
    "aws_controltower",
    "aws_costoptimizationhub",
    "aws_cur",
    "aws_customer",
    "aws_customerprofiles",
    "aws_dataexchange",
    "aws_datapipeline",
    "aws_datasync",
    "aws_datazone",
    "aws_dax",
    "aws_db",
    "aws_default",
    "aws_detective",
    "aws_devicefarm",
    "aws_directory",
    "aws_dlm",
    "aws_dms",
    "aws_docdb",
    "aws_docdbelastic",
    "aws_drs",
    "aws_dsql",
    "aws_ecr",
    "aws_ecrpublic",
    "aws_ecs",
    "aws_efs",
    "aws_egress",
    "aws_eip",
    "aws_eips",
    "aws_eks",
    "aws_elastic",
    "aws_elasticache",
    "aws_elasticsearch",
    "aws_elastictranscoder",
    "aws_elb",
    "aws_emr",
    "aws_emrcontainers",
    "aws_emrserverless",
    "aws_events",
    "aws_evidently",
    "aws_finspace",
    "aws_fis",
    "aws_flow",
    "aws_fms",
    "aws_fsx",
    "aws_gamelift",
    "aws_glacier",
    "aws_globalaccelerator",
    "aws_grafana",
    "aws_guardduty",
    "aws_identitystore",
    "aws_imagebuilder",
    "aws_instance",
    "aws_instances",
    "aws_internet",
    "aws_internetmonitor",
    "aws_ip",
    "aws_ivs",
    "aws_ivschat",
    "aws_kendra",
    "aws_key",
    "aws_keyspaces",
    "aws_kinesisanalyticsv2",
    "aws_kms",
    "aws_launch",
    "aws_lbs",
    "aws_lex",
    "aws_lexv2models",
    "aws_licensemanager",
    "aws_lightsail",
    "aws_load",
    "aws_location",
    "aws_m2",
    "aws_main",
    "aws_media",
    "aws_medialive",
    "aws_memorydb",
    "aws_mq",
    "aws_mskconnect",
    "aws_mwaa",
    "aws_nat",
    "aws_neptune",
    "aws_neptunegraph",
    "aws_network",
    "aws_networkfirewall",
    "aws_networkflowmonitor",
    "aws_networkmanager",
    "aws_networkmonitor",
    "aws_notifications",
    "aws_notificationscontacts",
    "aws_oam",
    "aws_observabilityadmin",
    "aws_odb",
    "aws_opensearch",
    "aws_opensearchserverless",
    "aws_osis",
    "aws_outposts",
    "aws_partition",
    "aws_paymentcryptography",
    "aws_pinpoint",
    "aws_pinpointsmsvoicev2",
    "aws_pipes",
    "aws_placement",
    "aws_polly",
    "aws_prefix",
    "aws_pricing",
    "aws_proxy",
    "aws_qbusiness",
    "aws_qldb",
    "aws_quicksight",
    "aws_rbin",
    "aws_redshiftdata",
    "aws_redshiftserverless",
    "aws_region",
    "aws_regions",
    "aws_rekognition",
    "aws_resiliencehub",
    "aws_resourceexplorer2",
    "aws_resourcegroups",
    "aws_resourcegroupstaggingapi",
    "aws_rolesanywhere",
    "aws_route53domains",
    "aws_route53profiles",
    "aws_route53recoverycontrolconfig",
    "aws_route53recoveryreadiness",
    "aws_rum",
    "aws_s3control",
    "aws_s3files",
    "aws_s3outposts",
    "aws_savingsplans",
    "aws_scheduler",
    "aws_schemas",
    "aws_security",
    "aws_securityhub",
    "aws_securitylake",
    "aws_serverlessapplicationrepository",
    "aws_servicecatalog",
    "aws_servicecatalogappregistry",
    "aws_servicequotas",
    "aws_ses",
    "aws_sesv2",
    "aws_shield",
    "aws_signer",
    "aws_snapshot",
    "aws_sns",
    "aws_spot",
    "aws_ssm",
    "aws_ssmcontacts",
    "aws_ssmincidents",
    "aws_ssmquicksetup",
    "aws_storagegateway",
    "aws_sts",
    "aws_subnet",
    "aws_subnets",
    "aws_swf",
    "aws_synthetics",
    "aws_timestreaminfluxdb",
    "aws_timestreamquery",
    "aws_timestreamwrite",
    "aws_transcribe",
    "aws_transfer",
    "aws_verifiedaccess",
    "aws_vpclattice",
    "aws_vpcs",
    "aws_vpn",
    "aws_wafregional",
    "aws_workspaces",
    "aws_workspacesweb",
  ]
}

check "frontmatter" {
  enabled = true

  require_subcategory = true
  require_page_title = true
  require_description = true
  require_layout = true

  allowed_subcategories_file = "website/allowed-subcategories.txt"

  allow_empty_subcategory_targets = [
    "arn_build",
    "arn_parse",
    "trim_iam_role_path",
    "user_agent",
  ]
}

check "section_presence" {
  enabled = false
}

check "timeouts_section" {
  enabled = true

  ignored_targets = [
    "aws_autoscaling_group",
    "aws_bedrock_custom_model",
    "aws_bedrockagent_agent_action_group",
    "aws_bedrockagent_agent_knowledge_base_association",
    "aws_budgets_budget_action",
    "aws_dx_hosted_private_virtual_interface",
    "aws_dx_hosted_transit_virtual_interface",
    "aws_eks_access_entry",
    "aws_eks_access_policy_association",
    "aws_fsx_openzfs_snapshot",
    "aws_globalaccelerator_custom_routing_endpoint_group",
    "aws_oam_sink_policy",
    "aws_quicksight_account_subscription",
    "aws_route53profiles_profile",
    "aws_route53profiles_resource_association",
    "aws_s3control_multi_region_access_point",
    "aws_spot_fleet_request",
    "aws_vpclattice_service",
    "aws_vpclattice_service_network_service_association",
    "aws_vpclattice_service_network_vpc_association",
    "aws_vpclattice_target_group",
    "aws_workspaces_connection_alias",
  ]
}
