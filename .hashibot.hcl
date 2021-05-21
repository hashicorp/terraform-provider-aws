queued_behavior "release_commenter" "releases" {
  repo_prefix = "terraform-provider-"

  message = <<-EOF
    This has been released in [version ${var.release_version} of the Terraform AWS provider](${var.changelog_link}). Please see the [Terraform documentation on provider versioning](https://www.terraform.io/docs/configuration/providers.html#provider-versions) or reach out if you need any assistance upgrading.

    For further feature requests or bug reports with this functionality, please create a [new GitHub issue](https://github.com/hashicorp/terraform-provider-aws/issues/new/choose) following the template for triage. Thanks!
  EOF
}

# Catch the following in issues:
# *aws_XXX
# * aws_XXX
# * `aws_XXX`
# -aws_XXX
# - aws_XXX
# - `aws_XXX`
# data "aws_XXX"
# resource "aws_XXX"
# NOTE: Go regexp does not support negative lookaheads
behavior "regexp_issue_labeler_v2" "service_labels" {
  regexp = "(\\* ?`?|- ?`?|data \"|resource \")aws_(\\w+)"

  label_map = {
    "service/accessanalyzer" = [
      "aws_accessanalyzer_",
    ],
    "service/acm" = [
      "aws_acm_",
    ],
    "service/acmpca" = [
      "aws_acmpca_",
    ],
    "service/alexaforbusiness" = [
      "aws_alexaforbusiness_",
    ],
    "service/amplify" = [
      "aws_amplify_",
    ],
    "service/apigateway" = [
      # Catch aws_api_gateway_XXX but not aws_api_gateway_v2_
      "aws_api_gateway_([^v]|v[^2]|v2[^_])",
    ],
    "service/apigatewayv2" = [
      "aws_api_gateway_v2_",
      "aws_apigatewayv2_",
    ],
    "service/appconfig" = [
      "aws_appconfig_",
    ],
    "service/applicationautoscaling" = [
      "aws_appautoscaling_",
    ],
    "service/applicationdiscoveryservice" = [
      "aws_applicationdiscoveryservice_",
    ],
    "service/applicationinsights" = [
      "aws_applicationinsights_",
    ],
    "service/appmesh" = [
      "aws_appmesh_",
    ],
    "service/apprunner" = [
      "aws_apprunner_",
    ],
    "service/appstream" = [
      "aws_appstream_",
    ],
    "service/appsync" = [
      "aws_appsync_",
    ],
    "service/athena" = [
      "aws_athena_",
    ],
    "service/auditmanager" = [
      "aws_auditmanager_",
    ],
    "service/autoscaling" = [
      "aws_autoscaling_",
      "aws_launch_configuration",
    ],
    "service/autoscalingplans" = [
      "aws_autoscalingplans_",
    ],
    "service/backup" = [
      "aws_backup_",
    ],
    "service/batch" = [
      "aws_batch_",
    ],
    "service/budgets" = [
      "aws_budgets_",
    ],
    "service/chime" = [
      "aws_chime_",
    ],
    "service/cloud9" = [
      "aws_cloud9_",
    ],
    "service/clouddirectory" = [
      "aws_clouddirectory_",
    ],
    "service/cloudformation" = [
      "aws_cloudformation_",
    ],
    "service/cloudfront" = [
      "aws_cloudfront_",
    ],
    "service/cloudhsmv2" = [
      "aws_cloudhsm_v2_",
    ],
    "service/cloudsearch" = [
      "aws_cloudsearch_",
    ],
    "service/cloudtrail" = [
      "aws_cloudtrail",
    ],
    "service/cloudwatch" = [
      "aws_cloudwatch_([^e]|e[^v]|ev[^e]|eve[^n]|even[^t]|event[^_]|[^l]|l[^o]|lo[^g]|log[^_])",
    ],
    "service/cloudwatchevents" = [
      "aws_cloudwatch_event_",
    ],
    "service/cloudwatchlogs" = [
      "aws_cloudwatch_log_",
      "aws_cloudwatch_query_definition",
    ],
    "service/codeartifact" = [
      "aws_codeartifact_",
    ],
    "service/codebuild" = [
      "aws_codebuild_",
    ],
    "service/codecommit" = [
      "aws_codecommit_",
    ],
    "service/codedeploy" = [
      "aws_codedeploy_",
    ],
    "service/codepipeline" = [
      "aws_codepipeline",
    ],
    "service/codestar" = [
      "aws_codestar_",
    ],
    "service/codestarconnections" = [
      "aws_codestarconnections_",
    ],
    "service/codestarnotifications" = [
      "aws_codestarnotifications_",
    ],
    "service/cognito" = [
      "aws_cognito_",
    ],
    "service/configservice" = [
      "aws_config_",
    ],
    "service/connect" = [
      "aws_connect_",
    ],
    "service/databasemigrationservice" = [
      "aws_dms_",
    ],
    "service/dataexchange" = [
      "aws_dataexchange_",
    ],
    "service/datapipeline" = [
      "aws_datapipeline_",
    ],
    "service/datasync" = [
      "aws_datasync_",
    ],
    "service/dax" = [
      "aws_dax_",
    ],
    "service/detective" = [
      "aws_detective_"
    ],
    "service/devicefarm" = [
      "aws_devicefarm_",
    ],
    "service/directconnect" = [
      "aws_dx_",
    ],
    "service/directoryservice" = [
      "aws_directory_service_",
    ],
    "service/dlm" = [
      "aws_dlm_",
    ],
    "service/docdb" = [
      "aws_docdb_",
    ],
    "service/dynamodb" = [
      "aws_dynamodb_",
    ],
    "service/ec2" = [
      "aws_ami",
      "aws_availability_zone",
      "aws_customer_gateway",
      "aws_(default_)?(network_acl|route_table|security_group|subnet|vpc)",
      "aws_ebs_",
      "aws_ec2_",
      "aws_egress_only_internet_gateway",
      "aws_eip",
      "aws_flow_log",
      "aws_instance",
      "aws_internet_gateway",
      "aws_key_pair",
      "aws_launch_template",
      "aws_main_route_table_association",
      "aws_network_interface",
      "aws_placement_group",
      "aws_prefix_list",
      "aws_spot",
      "aws_route(\"|`|$)",
      "aws_vpn_",
      "aws_volume_attachment",
    ],
    "service/ecr" = [
      "aws_ecr_",
    ],
    "service/ecrpublic" = [
      "aws_ecrpublic_",
    ],
    "service/ecs" = [
      "aws_ecs_",
    ],
    "service/efs" = [
      "aws_efs_",
    ],
    "service/eks" = [
      "aws_eks_",
    ],
    "service/elastic-transcoder" = [
      "aws_elastictranscoder_",
    ],
    "service/elasticache" = [
      "aws_elasticache_",
    ],
    "service/elasticbeanstalk" = [
      "aws_elastic_beanstalk_",
    ],
    "service/elasticsearch" = [
      "aws_elasticsearch_",
    ],
    "service/elb" = [
      "aws_app_cookie_stickiness_policy",
      "aws_elb",
      "aws_lb_cookie_stickiness_policy",
      "aws_lb_ssl_negotiation_policy",
      "aws_load_balancer_",
      "aws_proxy_protocol_policy",
    ],
    "service/elbv2" = [
      "aws_(a)?lb(\"|`|$)",
      # Catch aws_lb_XXX but not aws_lb_cookie_ or aws_lb_ssl_ (Classic ELB)
      "aws_(a)?lb_([^c]|c[^o]|co[^o]|coo[^k]|cook[^i]|cooki[^e]|cookie[^_]|[^s]|s[^s]|ss[^l]|ssl[^_])",
    ],
    "service/emr" = [
      "aws_emr_",
    ],
    "service/emrcontainers" = [
      "aws_emrcontainers_",
    ],
    "service/eventbridge" = [
      # EventBridge is rebranded CloudWatch Events
      "aws_cloudwatch_event_",
    ],
    "service/firehose" = [
      "aws_kinesis_firehose_",
    ],
    "service/fms" = [
      "aws_fms_",
    ],
    "service/forecast" = [
      "aws_forecast_",
    ],
    "service/fsx" = [
      "aws_fsx_",
    ],
    "service/gamelift" = [
      "aws_gamelift_",
    ],
    "service/glacier" = [
      "aws_glacier_",
    ],
    "service/globalaccelerator" = [
      "aws_globalaccelerator_",
    ],
    "service/glue" = [
      "aws_glue_",
    ],
    "service/greengrass" = [
      "aws_greengrass_",
    ],
    "service/guardduty" = [
      "aws_guardduty_",
    ],
    "service/iam" = [
      "aws_iam_",
    ],
    "service/identitystore" = [
      "aws_identitystore_",
    ],
    "service/imagebuilder" = [
      "aws_imagebuilder_",
    ],
    "service/inspector" = [
      "aws_inspector_",
    ],
    "service/iot" = [
      "aws_iot_",
    ],
    "service/iotanalytics" = [
      "aws_iotanalytics_",
    ],
    "service/iotevents" = [
      "aws_iotevents_",
    ],
    "service/kafka" = [
      "aws_msk_",
    ],
    "service/kinesis" = [
      # Catch aws_kinesis_XXX but not aws_kinesis_firehose_
      "aws_kinesis_([^f]|f[^i]|fi[^r]|fir[^e]|fire[^h]|fireh[^o]|fireho[^s]|firehos[^e]|firehose[^_])",
    ],
    "service/kinesisanalytics" = [
      "aws_kinesis_analytics_",
    ],
    "service/kinesisanalyticsv2" = [
      "aws_kinesisanalyticsv2_",
    ],
    "service/kms" = [
      "aws_kms_",
    ],
    "service/lakeformation" = [
      "aws_lakeformation_",
    ],
    "service/lambda" = [
      "aws_lambda_",
    ],
    "service/lexmodelbuildingservice" = [
      "aws_lex_",
    ],
    "service/licensemanager" = [
      "aws_licensemanager_",
    ],
    "service/lightsail" = [
      "aws_lightsail_",
    ],
    "service/machinelearning" = [
      "aws_machinelearning_",
    ],
    "service/macie" = [
      "aws_macie_",
    ],
    "service/macie2" = [
      "aws_macie2_",
    ],
    "service/marketplacecatalog" = [
      "aws_marketplace_catalog_",
    ],
    "service/mediaconnect" = [
      "aws_media_connect_",
    ],
    "service/mediaconvert" = [
      "aws_media_convert_",
    ],
    "service/medialive" = [
      "aws_media_live_",
    ],
    "service/mediapackage" = [
      "aws_media_package_",
    ],
    "service/mediastore" = [
      "aws_media_store_",
    ],
    "service/mediatailor" = [
      "aws_media_tailor_",
    ],
    "service/mobile" = [
      "aws_mobile_",
    ],
    "service/mq" = [
      "aws_mq_",
    ],
    "service/mwaa" = [
      "aws_mwaa_",
    ],
    "service/neptune" = [
      "aws_neptune_",
    ],
    "service/networkfirewall" = [
      "aws_networkfirewall_",
    ],
    "service/networkmanager" = [
      "aws_networkmanager_",
    ],
    "service/opsworks" = [
      "aws_opsworks_",
    ],
    "service/organizations" = [
      "aws_organizations_",
    ],
    "service/outposts" = [
      "aws_outposts_",
    ],
    "service/personalize" = [
      "aws_personalize_",
    ],
    "service/pinpoint" = [
      "aws_pinpoint_",
    ],
    "service/polly" = [
      "aws_polly_",
    ],
    "service/pricing" = [
      "aws_pricing_",
    ],
    "service/prometheusservice" = [
      "aws_prometheus_",
    ],
    "service/qldb" = [
      "aws_qldb_",
    ],
    "service/quicksight" = [
      "aws_quicksight_",
    ],
    "service/ram" = [
      "aws_ram_",
    ],
    "service/rds" = [
      "aws_db_",
      "aws_rds_",
    ],
    "service/redshift" = [
      "aws_redshift_",
    ],
    "service/resourcegroups" = [
      "aws_resourcegroups_",
    ],
    "service/resourcegroupstaggingapi" = [
      "aws_resourcegroupstaggingapi_",
    ],
    "service/robomaker" = [
      "aws_robomaker_",
    ],
    "service/route53" = [
      # Catch aws_route53_XXX but not aws_route53_domains_ or aws_route53_resolver_
      "aws_route53_([^d]|d[^o]|do[^m]|dom[^a]|doma[^i]|domai[^n]|domain[^s]|domains[^_]|[^r]|r[^e]|re[^s]|res[^o]|reso[^l]|resol[^v]|resolv[^e]|resolve[^r]|resolver[^_])",
    ],
    "service/route53domains" = [
      "aws_route53domains_",
    ],
    "service/route53resolver" = [
      "aws_route53_resolver_",
    ],
    "service/s3" = [
      "aws_canonical_user_id",
      "aws_s3_bucket",
      "aws_s3_object",
    ],
    "service/s3control" = [
      "aws_s3_account_",
      "aws_s3control_",
    ],
    "service/s3outposts" = [
      "aws_s3outposts_",
    ],
    "service/sagemaker" = [
      "aws_sagemaker_",
    ],
    "service/secretsmanager" = [
      "aws_secretsmanager_",
    ],
    "service/securityhub" = [
      "aws_securityhub_",
    ],
    "service/serverlessapplicationrepository" = [
      "aws_serverlessapplicationrepository_",
    ],
    "service/servicecatalog" = [
      "aws_servicecatalog_",
    ],
    "service/servicediscovery" = [
      "aws_service_discovery_",
    ],
    "service/servicequotas" = [
      "aws_servicequotas_",
    ],
    "service/ses" = [
      "aws_ses_",
    ],
    "service/sfn" = [
      "aws_sfn_",
    ],
    "service/shield" = [
      "aws_shield_",
    ],
    "service/signer" = [
      "aws_signer_",
    ],
    "service/simpledb" = [
      "aws_simpledb_",
    ],
    "service/snowball" = [
      "aws_snowball_",
    ],
    "service/sns" = [
      "aws_sns_",
    ],
    "service/sqs" = [
      "aws_sqs_",
    ],
    "service/ssm" = [
      "aws_ssm_",
    ],
    "service/ssoadmin" = [
      "aws_ssoadmin_",
    ],
    "service/storagegateway" = [
      "aws_storagegateway_",
    ],
    "service/sts" = [
      "aws_caller_identity",
    ],
    "service/swf" = [
      "aws_swf_",
    ],
    "service/synthetics" = [
      "aws_synthetics_",
    ],
    "service/timestreamwrite" = [
      "aws_timestreamwrite_",
    ],
    "service/transfer" = [
      "aws_transfer_",
    ],
    "service/waf" = [
      "aws_waf_",
      "aws_wafregional_",
    ],
    "service/wafv2" = [
      "aws_wafv2_",
    ],
    "service/workdocs" = [
      "aws_workdocs_",
    ],
    "service/worklink" = [
      "aws_worklink_",
    ],
    "service/workmail" = [
      "aws_workmail_",
    ],
    "service/workspaces" = [
      "aws_workspaces_",
    ],
    "service/xray" = [
      "aws_xray_",
    ],
  }
}

behavior "remove_labels_on_reply" "remove_stale" {
    labels = ["waiting-response", "stale"]
    only_non_maintainers = true
}

behavior "pull_request_size_labeler" "size" {
    label_prefix = "size/"
    label_map = {
        "size/XS" = {
            from = 0
            to = 30
        }
        "size/S" = {
            from = 31
            to = 60
        }
        "size/M" = {
            from = 61
            to = 150
        }
        "size/L" = {
            from = 151
            to = 300
        }
        "size/XL" = {
            from = 301
            to = 1000
        }
        "size/XXL" = {
            from = 1001
            to = 0
        }
    }
}
