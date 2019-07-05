behavior "pull_request_path_labeler" "service_labels" {
  label_map = {
    # label provider related changes
    "provider" = [
      "aws/auth_helpers.go",
      "aws/awserr.go",
      "aws/config.go",
      "aws/*_aws_arn*",
      "aws/*_aws_ip_ranges*",
      "aws/*_aws_partition*",
      "aws/*_aws_region*",
      "aws/provider.go",
      "aws/utils.go",
      "website/docs/index.html.markdown",
      "website/**/arn*",
      "website/**/ip_ranges*",
      "website/**/partition*",
      "website/**/region*"
    ]
    # label test related changes
    "tests" = [
      "**/*_test.go",
      ".gometalinter.json",
      ".travis.yml"
    ]
    # label services
    "service/acm" = [
      "**/*_acm_*",
      "**/acm_*",
      "aws/tagsACM*"
    ]
    "service/acmpca" = [
      "**/*_acmpca_*",
      "**/acmpca_*",
      "aws/tagsACMPCA*"
    ]
    "service/alexaforbusiness" = [
      "**/*_alexaforbusiness_*",
      "**/alexaforbusiness_*"
    ]
    "service/amplify" = [
      "**/*_amplify_*",
      "**/amplify_*"
    ]
    "service/apigateway" = [
      "**/*_api_gateway_[^v][^2][^_]*",
      "**/api_gateway_[^v][^2][^_]*",
      "aws/tags_apigateway[^v][^2]*"
    ]
    "service/apigatewayv2" = [
      "**/*_api_gateway_v2_*",
      "**/api_gateway_v2_*",
      "aws/tags_apigatewayv2*"
    ]
    "service/applicationautoscaling" = [
      "**/*_appautoscaling_*",
      "**/appautoscaling_*"
    ]
    # "service/applicationdiscoveryservice" = [
    # 	"**/*_applicationdiscoveryservice_*",
    # 	"**/applicationdiscoveryservice_*"
    # ]
    "service/applicationinsights" = [
      "**/*_applicationinsights_*",
      "**/applicationinsights_*"
    ]
    "service/appmesh" = [
      "**/*_appmesh_*",
      "**/appmesh_*"
    ]
    "service/appstream" = [
      "**/*_appstream_*",
      "**/appstream_*"
    ]
    "service/appsync" = [
      "**/*_appsync_*",
      "**/appsync_*"
    ]
    "service/athena" = [
      "service/athena",
      "**/athena_*"
    ]
    "service/autoscaling" = [
      "**/*_autoscaling_*",
      "**/autoscaling_*",
      "aws/*_aws_launch_configuration*",
      "website/**/launch_configuration*"
    ]
    "service/autoscalingplans" = [
      "**/*_autoscalingplans_*",
      "**/autoscalingplans_*"
    ]
    "service/batch" = [
      "**/*_batch_*",
      "**/batch_*"
    ]
    "service/budgets" = [
      "**/*_budgets_*",
      "**/budgets_*"
    ]
    "service/cloud9" = [
      "**/*_cloud9_*",
      "**/cloud9_*"
    ]
    "service/clouddirectory" = [
      "**/*_clouddirectory_*",
      "**/clouddirectory_*"
    ]
    "service/cloudformation" = [
      "**/*_cloudformation_*",
      "**/cloudformation_*"
    ]
    "service/cloudfront" = [
      "**/*_cloudfront_*",
      "**/cloudfront_*",
      "aws/tagsCloudFront*"
    ]
    "service/cloudhsmv2" = [
      "**/*_cloudhsm_v2_*",
      "**/cloudhsm_v2_*"
    ]
    "service/cloudsearch" = [
      "**/*_cloudsearch_*",
      "**/cloudsearch_*"
    ]
    "service/cloudtrail" = [
      "**/*_cloudtrail_*",
      "**/cloudtrail_*",
      "aws/tagsCloudtrail*"
    ]
    "service/cloudwatch" = [
      "**/*_cloudwatch_dashboard*",
      "**/*_cloudwatch_metric_alarm*",
      "**/cloudwatch_dashboard*",
      "**/cloudwatch_metric_alarm*"
    ]
    "service/cloudwatchevents" = [
      "**/*_cloudwatch_event_*",
      "**/cloudwatch_event_*"
    ]
    "service/cloudwatchlogs" = [
      "**/*_cloudwatch_log_*",
      "**/cloudwatch_log_*"
    ]
    "service/codebuild" = [
      "**/*_codebuild_*",
      "**/codebuild_*",
      "aws/tagsCodeBuild*"
    ]
    "service/codecommit" = [
      "**/*_codecommit_*",
      "**/codecommit_*"
    ]
    "service/codedeploy" = [
      "**/*_codedeploy_*",
      "**/codedeploy_*"
    ]
    "service/codepipeline" = [
      "**/*_codepipeline_*",
      "**/codepipeline_*"
    ]
    "service/codestar" = [
      "**/*_codestar_*",
      "**/codestar_*"
    ]
    "service/cognito" = [
      "**/*_cognito_*",
      "**/_cognito_*"
    ]
    "service/configservice" = [
      "aws/*_aws_config_*",
      "website/**/config_*"
    ]
    "service/databasemigrationservice" = [
      "**/*_dms_*",
      "**/dms_*",
      "aws/tags_dms*"
    ]
    "service/datapipeline" = [
      "**/*_datapipeline_*",
      "**/datapipeline_*",
    ]
    "service/datasync" = [
      "**/*_datasync_*",
      "**/datasync_*",
    ]
    "service/dax" = [
      "**/*_dax_*",
      "**/dax_*",
      "aws/tagsDAX*"
    ]
    "service/devicefarm" = [
      "**/*_devicefarm_*",
      "**/devicefarm_*"
    ]
    "service/directconnect" = [
      "**/*_dx_*",
      "**/dx_*",
      "aws/tagsDX*"
    ]
    "service/directoryservice" = [
      "**/*_directory_service_*",
      "**/directory_service_*",
      "aws/tagsDS*"
    ]
    "service/dlm" = [
      "**/*_dlm_*",
      "**/dlm_*"
    ]
    "service/dynamodb" = [
      "**/*_dynamodb_*",
      "**/dynamodb_*"
    ]
    # Special casing this one because the files aren't _ec2_
    "service/ec2" = [
      "**/*_ec2_*",
      "**/ec2_*",
      "aws/*_aws_ami*",
      "aws/*_aws_availability_zone*",
      "aws/*_aws_customer_gateway*",
      "aws/*_aws_default_network_acl*",
      "aws/*_aws_default_route_table*",
      "aws/*_aws_default_security_group*",
      "aws/*_aws_default_subnet*",
      "aws/*_aws_default_vpc*",
      "aws/*_aws_ebs_*",
      "aws/*_aws_egress_only_internet_gateway*",
      "aws/*_aws_eip*",
      "aws/*_aws_flow_log*",
      "aws/*_aws_instance*",
      "aws/*_aws_internet_gateway*",
      "aws/*_aws_key_pair*",
      "aws/*_aws_launch_template*",
      "aws/*_aws_main_route_table_association*",
      "aws/*_aws_nat_gateway*",
      "aws/*_aws_network_acl*",
      "aws/*_aws_network_interface*",
      "aws/*_aws_placement_group*",
      "aws/*_aws_route_table*",
      "aws/*_aws_route.*",
      "aws/*_aws_security_group*",
      "aws/*_aws_spot*",
      "aws/*_aws_subnet*",
      "aws/*_aws_vpc*",
      "aws/*_aws_vpn*",
      "website/**/availability_zone*",
      "website/**/customer_gateway*",
      "website/**/default_network_acl*",
      "website/**/default_route_table*",
      "website/**/default_security_group*",
      "website/**/default_subnet*",
      "website/**/default_vpc*",
      "website/**/ebs_*",
      "website/**/egress_only_internet_gateway*",
      "website/**/eip*",
      "website/**/flow_log*",
      "website/**/instance*",
      "website/**/internet_gateway*",
      "website/**/key_pair*",
      "website/**/launch_template*",
      "website/**/main_route_table_association*",
      "website/**/nat_gateway*",
      "website/**/network_acl*",
      "website/**/network_interface*",
      "website/**/placement_group*",
      "website/**/route_table*",
      "website/**/route.*",
      "website/**/security_group*",
      "website/**/spot_*",
      "website/**/subnet.*",
      "website/**/vpc*",
      "website/**/vpn*"
    ]
    "service/ecr" = [
      "**/*_ecr_*",
      "**/ecr_*"
    ]
    "service/ecs" = [
      "**/*_ecs_*",
      "**/ecs_*"
    ]
    "service/efs" = [
      "**/*_efs_*",
      "**/efs_*",
      "aws/tagsEFS*"
    ]
    "service/eks" = [
      "**/*_eks_*",
      "**/eks_*"
    ]
    "service/elastic-transcoder" = [
      "**/*_elastic_transcoder_*",
      "**/elastic_transcoder_*"
    ]
    "service/elasticache" = [
      "**/*_elasticache_*",
      "**/elasticache_*",
      "aws/tagsEC*"
    ]
    "service/elasticbeanstalk" = [
      "**/*_elastic_beanstalk_*",
      "**/elastic_beanstalk_*",
      "aws/tagsBeanstalk*"
    ]
    "service/elasticsearch" = [
      "**/*_elasticsearch_*",
      "**/elasticsearch_*",
      "**/*_elasticsearchservice*"
    ]
    "service/elb" = [
      "aws/*_aws_app_cookie_stickiness_policy*",
      "aws/*_aws_elb*",
      "aws/*_aws_lb_cookie_stickiness_policy*",
      "aws/*_aws_lb_ssl_negotiation_policy*",
      "aws/*_aws_proxy_protocol_policy*",
      "aws/tagsELB*",
      "website/**/app_cookie_stickiness_policy*",
      "website/**/elb*",
      "website/**/lb_cookie_stickiness_policy*",
      "website/**/lb_ssl_negotiation_policy*",
      "website/**/proxy_protocol_policy*"
    ]
    "service/elbv2" = [
      "aws/*_lb.*",
      "aws/*_lb_listener*",
      "aws/*_lb_target_group*",
      "website/**/lb.*",
      "website/**/lb_listener*",
      "website/**/lb_target_group*"
    ]
    "service/emr" = [
      "**/*_emr_*",
      "**/emr_*"
    ]
    "service/firehose" = [
      "**/*_firehose_*",
      "**/firehose_*"
    ]
    "service/fms" = [
      "**/*_fms_*",
      "**/fms_*"
    ]
    "service/fsx" = [
      "**/*_fsx_*",
      "**/fsx_*"
    ]
    "service/gamelift" = [
      "**/*_gamelift_*",
      "**/gamelift_*"
    ]
    "service/glacier" = [
      "**/*_glacier_*",
      "**/glacier_*"
    ]
    "service/globalaccelerator" = [
      "**/*_globalaccelerator_*",
      "**/globalaccelerator_*"
    ]
    "service/glue" = [
      "**/*_glue_*",
      "**/glue_*"
    ]
    "service/greengrass" = [
      "**/*_greengrass_*",
      "**/greengrass_*"
    ]
    "service/guardduty" = [
      "**/*_guardduty_*",
      "**/guardduty_*"
    ]
    "service/iam" = [
      "**/*_iam_*",
      "**/iam_*"
    ]
    "service/inspector" = [
      "**/*_inspector_*",
      "**/inspector_*",
      "aws/tagsInspector*"
    ]
    "service/iot" = [
      "**/*_iot_*",
      "**/iot_*"
    ]
    "service/kinesis" = [
      "aws/*_aws_kinesis_stream*",
      "aws/tags_kinesis*",
      "website/kinesis_stream*"
    ]
    "service/kinesisanalytics" = [
      "**/*_kinesisanalytics_*",
      "**/kinesisanalytics_*"
    ]
    "service/kms" = [
      "**/*_kms_*",
      "**/kms_*",
      "aws/tagsKMS*"
    ]
    "service/lambda" = [
      "**/*_lambda_*",
      "**/lambda_*",
      "aws/tagsLambda*"
    ]
    "service/lexmodelbuildingservice" = [
      "**/*_lex_*",
      "**/lex_*"
    ]
    "service/licensemanager" = [
      "**/*_licensemanager_*",
      "**/licensemanager_*"
    ]
    "service/lightsail" = [
      "**/*_lightsail_*",
      "**/lightsail_*"
    ]
    "service/machinelearning" = [
      "**/*_machinelearning_*",
      "**/machinelearning_*"
    ]
    "service/macie" = [
      "**/*_macie_*",
      "**/macie_*"
    ]
    "service/mediaconnect" = [
      "**/*_media_connect_*",
      "**/media_connect_*"
    ]
    "service/mediaconvert" = [
      "**/*_media_convert_*",
      "**/media_convert_*"
    ]
    "service/medialive" = [
      "**/*_media_live_*",
      "**/media_live_*"
    ]
    "service/mediapackage" = [
      "**/*_media_package_*",
      "**/media_package_*"
    ]
    "service/mediastore" = [
      "**/*_media_store_*",
      "**/media_store_*"
    ]
    "service/mediatailor" = [
      "**/*_media_tailor_*",
      "**/media_tailor_*",
    ]
    "service/mobile" = [
      "**/*_mobile_*",
      "**/mobile_*"
    ],
    "service/mq" = [
      "**/*_mq_*",
      "**/mq_*"
    ]
    "service/neptune" = [
      "**/*_neptune_*",
      "**/neptune_*",
      "aws/tagsNeptune*"
    ]
    "service/opsworks" = [
      "**/*_opsworks_*",
      "**/opsworks_*",
      "aws/tagsOpsworks*"
    ]
    "service/organizations" = [
      "**/*_organizations_*",
      "**/organizations_*"
    ]
    "service/pinpoint" = [
      "**/*_pinpoint_*",
      "**/pinpoint_*"
    ]
    "service/polly" = [
      "**/*_polly_*",
      "**/polly_*"
    ]
    "service/pricing" = [
      "**/*_pricing_*",
      "**/pricing_*"
    ]
    "service/ram" = [
      "**/*_ram_*",
      "**/ram_*"
    ]
    "service/rds" = [
      "aws/*_aws_db_*",
      "aws/*_aws_rds_*",
      "aws/tagsRDS*",
      "website/**/db_*",
      "website/**/rds_*"
    ]
    "service/redshift" = [
      "**/*_redshift_*",
      "**/redshift_*",
      "aws/tagsRedshift*"
    ]
    "service/resourcegroups" = [
      "**/*_resourcegroups_*",
      "**/resourcegroups_*"
    ]
    "service/route53" = [
      "**/*_route53_delegation_set*",
      "**/*_route53_health_check*",
      "**/*_route53_query_log*",
      "**/*_route53_record*",
      "**/*_route53_zone*",
      "**/route53_delegation_set*",
      "**/route53_health_check*",
      "**/route53_query_log*",
      "**/route53_record*",
      "**/route53_zone*",
      "aws/tags_route53*"
    ]
    "service/robomaker" = [
      "**/*_robomaker_*",
      "**/robomaker_*",
    ]
    "service/route53domains" = [
      "**/*_route53_domains_*",
      "**/route53_domains_*"
    ]
    "service/s3" = [
      "**/*_s3_bucket*",
      "**/s3_bucket*",
      "aws/*_aws_canonical_user_id*",
      "website/**/canonical_user_id*"
    ]
    "service/s3control" = [
      "**/*_s3_account_*",
      "**/s3_account_*"
    ]
    "service/sagemaker" = [
      "**/*_sagemaker_*",
      "**/sagemaker_*"
    ]
    "service/secretsmanager" = [
      "**/*_secretsmanager_*",
      "**/secretsmanager_*",
      "aws/tagsSecretsManager*"
    ]
    "service/securityhub" = [
      "**/*_securityhub_*",
      "**/securityhub_*"
    ]
    "service/servicecatalog" = [
      "**/*_servicecatalog_*",
      "**/servicecatalog_*"
    ]
    "service/servicediscovery" = [
      "**/*_service_discovery_*",
      "**/service_discovery_*"
    ]
    "service/servicequotas" = [
      "**/*_servicequotas_*",
      "**/servicequotas_*"
    ]
    "service/ses" = [
      "**/*_ses_*",
      "**/ses_*",
      "aws/tagsSSM*"
    ]
    "service/sfn" = [
      "**/*_sfn_*",
      "**/sfn_*"
    ]
    "service/shield" = [
      "**/*_shield_*",
      "**/shield_*",
    ],
    "service/simpledb" = [
      "**/*_simpledb_*",
      "**/simpledb_*"
    ]
    "service/snowball" = [
      "**/*_snowball_*",
      "**/snowball_*"
    ]
    "service/sns" = [
      "**/*_sns_*",
      "**/sns_*"
    ]
    "service/sqs" = [
      "**/*_sqs_*",
      "**/sqs_*"
    ]
    "service/ssm" = [
      "**/*_ssm_*",
      "**/ssm_*"
    ]
    "service/storagegateway" = [
      "**/*_storagegateway_*",
      "**/storagegateway_*"
    ]
    "service/sts" = [
      "aws/*_aws_caller_identity*",
      "website/**/caller_identity*"
    ]
    "service/swf" = [
      "**/*_swf_*",
      "**/swf_*"
    ]
    "service/transfer" = [
      "**/*_transfer_*",
      "**/transfer_*"
    ]
    "service/waf" = [
      "**/*_waf_*",
      "**/waf_*",
      "**/*_wafregional_*",
      "**/wafregional_*"
    ]
    "service/workdocs" = [
      "**/*_workdocs_*",
      "**/workdocs_*"
    ]
    "service/workmail" = [
      "**/*_workmail_*",
      "**/workmail_*"
    ]
    "service/workspaces" = [
      "**/*_workspaces_*",
      "**/workspaces_*"
    ]
    "service/xray" = [
      "**/*_xray_*",
      "**/xray_*"
    ]
  }
}

behavior "regexp_issue_labeler" "aws_acm_service_labels" {
  regexp = "\*\saws_acm.+\n"
  labels = ["service/acm"]
}

behavior "regexp_issue_labeler" "aws_acm_resource_labels" {
  regexp = "resource\s.+aws_acm.+"
  labels = ["service/acm"]
}

behavior "regexp_issue_labeler" "aws_acm_data_labels" {
  regexp = "data\s.+aws_acm.+"
  labels = ["service/acm"]
}

behavior "regexp_issue_labeler" "aws_acmpca_service_labels" {
  regexp = "\*\saws_acmpca.+\n"
  labels = ["service/acmpca"]
}

behavior "regexp_issue_labeler" "aws_acmpca_resource_labels" {
  regexp = "resource\s.+aws_acmpca.+"
  labels = ["service/acmpca"]
}

behavior "regexp_issue_labeler" "aws_acmpca_data_labels" {
  regexp = "data\s.+aws_acmpca.+"
  labels = ["service/acmpca"]
}

behavior "regexp_issue_labeler" "aws_alexaforbusiness_service_labels" {
  regexp = "\*\saws_alexaforbusiness.+\n"
  labels = ["service/alexaforbusiness"]
}

behavior "regexp_issue_labeler" "aws_alexaforbusiness_resource_labels" {
  regexp = "resource\s.+aws_alexaforbusiness.+"
  labels = ["service/alexaforbusiness"]
}

behavior "regexp_issue_labeler" "aws_alexaforbusiness_data_labels" {
  regexp = "data\s.+aws_alexaforbusiness.+"
  labels = ["service/alexaforbusiness"]
}

behavior "regexp_issue_labeler" "aws_amplify_service_labels" {
  regexp = "\*\saws_amplify.+\n"
  labels = ["service/amplify"]
}

behavior "regexp_issue_labeler" "aws_amplify_resource_labels" {
  regexp = "resource\s.+aws_amplify.+"
  labels = ["service/amplify"]
}

behavior "regexp_issue_labeler" "aws_amplify_data_labels" {
  regexp = "data\s.+aws_amplify.+"
  labels = ["service/amplify"]
}

behavior "regexp_issue_labeler" "aws_api_gateway_[^v][^2][^_]_service_labels" {
  regexp = "\*\saws_api_gateway_[^v][^2][^_].+\n"
  labels = ["service/apigateway"]
}

behavior "regexp_issue_labeler" "aws_api_gateway_[^v][^2][^_]_resource_labels" {
  regexp = "resource\s.+aws_api_gateway_[^v][^2][^_].+"
  labels = ["service/apigateway"]
}

behavior "regexp_issue_labeler" "aws_api_gateway_[^v][^2][^_]_data_labels" {
  regexp = "data\s.+aws_api_gateway_[^v][^2][^_].+"
  labels = ["service/apigateway"]
}

behavior "regexp_issue_labeler" "aws_api_gateway_v2_service_labels" {
  regexp = "\*\saws_api_gateway_v2.+\n"
  labels = ["service/apigatewayv2"]
}

behavior "regexp_issue_labeler" "aws_api_gateway_v2_resource_labels" {
  regexp = "resource\s.+aws_api_gateway_v2.+"
  labels = ["service/apigatewayv2"]
}

behavior "regexp_issue_labeler" "aws_api_gateway_v2_data_labels" {
  regexp = "data\s.+aws_api_gateway_v2.+"
  labels = ["service/apigatewayv2"]
}

behavior "regexp_issue_labeler" "aws_applicationautoscaling_service_labels" {
  regexp = "\*\saws_applicationautoscaling.+\n"
  labels = ["service/applicationautoscaling"]
}

behavior "regexp_issue_labeler" "aws_applicationautoscaling_resource_labels" {
  regexp = "resource\s.+aws_applicationautoscaling.+"
  labels = ["service/applicationautoscaling"]
}

behavior "regexp_issue_labeler" "aws_applicationautoscaling_data_labels" {
  regexp = "data\s.+aws_applicationautoscaling.+"
  labels = ["service/applicationautoscaling"]
}

behavior "regexp_issue_labeler" "aws_applicationinsights_service_labels" {
  regexp = "\*\saws_applicationinsights.+\n"
  labels = ["service/applicationinsights"]
}

behavior "regexp_issue_labeler" "aws_applicationinsights_resource_labels" {
  regexp = "resource\s.+aws_applicationinsights.+"
  labels = ["service/applicationinsights"]
}

behavior "regexp_issue_labeler" "aws_applicationinsights_data_labels" {
  regexp = "data\s.+aws_applicationinsights.+"
  labels = ["service/applicationinsights"]
}

behavior "regexp_issue_labeler" "aws_appmesh_service_labels" {
  regexp = "\*\saws_appmesh.+\n"
  labels = ["service/appmesh"]
}

behavior "regexp_issue_labeler" "aws_appmesh_resource_labels" {
  regexp = "resource\s.+aws_appmesh.+"
  labels = ["service/appmesh"]
}

behavior "regexp_issue_labeler" "aws_appmesh_data_labels" {
  regexp = "data\s.+aws_appmesh.+"
  labels = ["service/appmesh"]
}

behavior "regexp_issue_labeler" "aws_appstream_service_labels" {
  regexp = "\*\saws_appstream.+\n"
  labels = ["service/appstream"]
}

behavior "regexp_issue_labeler" "aws_appstream_resource_labels" {
  regexp = "resource\s.+aws_appstream.+"
  labels = ["service/appstream"]
}

behavior "regexp_issue_labeler" "aws_appstream_data_labels" {
  regexp = "data\s.+aws_appstream.+"
  labels = ["service/appstream"]
}

behavior "regexp_issue_labeler" "aws_appsync_service_labels" {
  regexp = "\*\saws_appsync.+\n"
  labels = ["service/appsync"]
}

behavior "regexp_issue_labeler" "aws_appsync_resource_labels" {
  regexp = "resource\s.+aws_appsync.+"
  labels = ["service/appsync"]
}

behavior "regexp_issue_labeler" "aws_appsync_data_labels" {
  regexp = "data\s.+aws_appsync.+"
  labels = ["service/appsync"]
}

behavior "regexp_issue_labeler" "aws_athena_service_labels" {
  regexp = "\*\saws_athena.+\n"
  labels = ["service/athena"]
}

behavior "regexp_issue_labeler" "aws_athena_resource_labels" {
  regexp = "resource\s.+aws_athena.+"
  labels = ["service/athena"]
}

behavior "regexp_issue_labeler" "aws_athena_data_labels" {
  regexp = "data\s.+aws_athena.+"
  labels = ["service/athena"]
}

behavior "regexp_issue_labeler" "aws_autoscaling_service_labels" {
  regexp = "\*\saws_autoscaling.+\n"
  labels = ["service/autoscaling"]
}

behavior "regexp_issue_labeler" "aws_autoscaling_resource_labels" {
  regexp = "resource\s.+aws_autoscaling.+"
  labels = ["service/autoscaling"]
}

behavior "regexp_issue_labeler" "aws_autoscaling_data_labels" {
  regexp = "data\s.+aws_autoscaling.+"
  labels = ["service/autoscaling"]
}

behavior "regexp_issue_labeler" "aws_launch_configuration_service_labels" {
  regexp = "\*\saws_launch_configuration.+\n"
  labels = ["service/autoscaling"]
}

behavior "regexp_issue_labeler" "aws_launch_configuration_resource_labels" {
  regexp = "resource\s.+aws_launch_configuration.+"
  labels = ["service/autoscaling"]
}

behavior "regexp_issue_labeler" "aws_launch_configuration_data_labels" {
  regexp = "data\s.+aws_launch_configuration.+"
  labels = ["service/autoscaling"]
}

behavior "regexp_issue_labeler" "aws_autoscalingplans_service_labels" {
  regexp = "\*\saws_autoscalingplans.+\n"
  labels = ["service/autoscalingplans"]
}

behavior "regexp_issue_labeler" "aws_autoscalingplans_resource_labels" {
  regexp = "resource\s.+aws_autoscalingplans.+"
  labels = ["service/autoscalingplans"]
}

behavior "regexp_issue_labeler" "aws_autoscalingplans_data_labels" {
  regexp = "data\s.+aws_autoscalingplans.+"
  labels = ["service/autoscalingplans"]
}

behavior "regexp_issue_labeler" "aws_batch_service_labels" {
  regexp = "\*\saws_batch.+\n"
  labels = ["service/batch"]
}

behavior "regexp_issue_labeler" "aws_batch_resource_labels" {
  regexp = "resource\s.+aws_batch.+"
  labels = ["service/batch"]
}

behavior "regexp_issue_labeler" "aws_batch_data_labels" {
  regexp = "data\s.+aws_batch.+"
  labels = ["service/batch"]
}

behavior "regexp_issue_labeler" "aws_budgets_service_labels" {
  regexp = "\*\saws_budgets.+\n"
  labels = ["service/budgets"]
}

behavior "regexp_issue_labeler" "aws_budgets_resource_labels" {
  regexp = "resource\s.+aws_budgets.+"
  labels = ["service/budgets"]
}

behavior "regexp_issue_labeler" "aws_budgets_data_labels" {
  regexp = "data\s.+aws_budgets.+"
  labels = ["service/budgets"]
}

behavior "regexp_issue_labeler" "aws_cloud9_service_labels" {
  regexp = "\*\saws_cloud9.+\n"
  labels = ["service/cloud9"]
}

behavior "regexp_issue_labeler" "aws_cloud9_resource_labels" {
  regexp = "resource\s.+aws_cloud9.+"
  labels = ["service/cloud9"]
}

behavior "regexp_issue_labeler" "aws_cloud9_data_labels" {
  regexp = "data\s.+aws_cloud9.+"
  labels = ["service/cloud9"]
}

behavior "regexp_issue_labeler" "aws_clouddirectory_service_labels" {
  regexp = "\*\saws_clouddirectory.+\n"
  labels = ["service/clouddirectory"]
}

behavior "regexp_issue_labeler" "aws_clouddirectory_resource_labels" {
  regexp = "resource\s.+aws_clouddirectory.+"
  labels = ["service/clouddirectory"]
}

behavior "regexp_issue_labeler" "aws_clouddirectory_data_labels" {
  regexp = "data\s.+aws_clouddirectory.+"
  labels = ["service/clouddirectory"]
}

behavior "regexp_issue_labeler" "aws_cloudformation_service_labels" {
  regexp = "\*\saws_cloudformation.+\n"
  labels = ["service/cloudformation"]
}

behavior "regexp_issue_labeler" "aws_cloudformation_resource_labels" {
  regexp = "resource\s.+aws_cloudformation.+"
  labels = ["service/cloudformation"]
}

behavior "regexp_issue_labeler" "aws_cloudformation_data_labels" {
  regexp = "data\s.+aws_cloudformation.+"
  labels = ["service/cloudformation"]
}

behavior "regexp_issue_labeler" "aws_cloudfront_service_labels" {
  regexp = "\*\saws_cloudfront.+\n"
  labels = ["service/cloudfront"]
}

behavior "regexp_issue_labeler" "aws_cloudfront_resource_labels" {
  regexp = "resource\s.+aws_cloudfront.+"
  labels = ["service/cloudfront"]
}

behavior "regexp_issue_labeler" "aws_cloudfront_data_labels" {
  regexp = "data\s.+aws_cloudfront.+"
  labels = ["service/cloudfront"]
}

behavior "regexp_issue_labeler" "aws_cloudhsm_v2_service_labels" {
  regexp = "\*\saws_cloudhsm_v2.+\n"
  labels = ["service/cloudhsmv2"]
}

behavior "regexp_issue_labeler" "aws_cloudhsm_v2_resource_labels" {
  regexp = "resource\s.+aws_cloudhsm_v2.+"
  labels = ["service/cloudhsmv2"]
}

behavior "regexp_issue_labeler" "aws_cloudhsm_v2_data_labels" {
  regexp = "data\s.+aws_cloudhsm_v2.+"
  labels = ["service/cloudhsmv2"]
}

behavior "regexp_issue_labeler" "aws_cloudsearch_service_labels" {
  regexp = "\*\saws_cloudsearch.+\n"
  labels = ["service/cloudsearch"]
}

behavior "regexp_issue_labeler" "aws_cloudsearch_resource_labels" {
  regexp = "resource\s.+aws_cloudsearch.+"
  labels = ["service/cloudsearch"]
}

behavior "regexp_issue_labeler" "aws_cloudsearch_data_labels" {
  regexp = "data\s.+aws_cloudsearch.+"
  labels = ["service/cloudsearch"]
}

behavior "regexp_issue_labeler" "aws_cloudtrail_service_labels" {
  regexp = "\*\saws_cloudtrail.+\n"
  labels = ["service/cloudtrail"]
}

behavior "regexp_issue_labeler" "aws_cloudtrail_resource_labels" {
  regexp = "resource\s.+aws_cloudtrail.+"
  labels = ["service/cloudtrail"]
}

behavior "regexp_issue_labeler" "aws_cloudtrail_data_labels" {
  regexp = "data\s.+aws_cloudtrail.+"
  labels = ["service/cloudtrail"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_dashboard_service_labels" {
  regexp = "\*\saws_cloudwatch_dashboard.+\n"
  labels = ["service/cloudwatch"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_dashboard_resource_labels" {
  regexp = "resource\s.+aws_cloudwatch_dashboard.+"
  labels = ["service/cloudwatch"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_dashboard_data_labels" {
  regexp = "data\s.+aws_cloudwatch_dashboard.+"
  labels = ["service/cloudwatch"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_metric_alarm_service_labels" {
  regexp = "\*\saws_cloudwatch_metric_alarm.+\n"
  labels = ["service/cloudwatch"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_metric_alarm_resource_labels" {
  regexp = "resource\s.+aws_cloudwatch_metric_alarm.+"
  labels = ["service/cloudwatch"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_metric_alarm_data_labels" {
  regexp = "data\s.+aws_cloudwatch_metric_alarm.+"
  labels = ["service/cloudwatch"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_event_service_labels" {
  regexp = "\*\saws_cloudwatch_event.+\n"
  labels = ["service/cloudwatchevents"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_event_resource_labels" {
  regexp = "resource\s.+aws_cloudwatch_event.+"
  labels = ["service/cloudwatchevents"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_event_data_labels" {
  regexp = "data\s.+aws_cloudwatch_event.+"
  labels = ["service/cloudwatchevents"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_log_service_labels" {
  regexp = "\*\saws_cloudwatch_log.+\n"
  labels = ["service/cloudwatchlogs"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_log_resource_labels" {
  regexp = "resource\s.+aws_cloudwatch_log.+"
  labels = ["service/cloudwatchlogs"]
}

behavior "regexp_issue_labeler" "aws_cloudwatch_log_data_labels" {
  regexp = "data\s.+aws_cloudwatch_log.+"
  labels = ["service/cloudwatchlogs"]
}

behavior "regexp_issue_labeler" "aws_codebuild_service_labels" {
  regexp = "\*\saws_codebuild.+\n"
  labels = ["service/codebuild"]
}

behavior "regexp_issue_labeler" "aws_codebuild_resource_labels" {
  regexp = "resource\s.+aws_codebuild.+"
  labels = ["service/codebuild"]
}

behavior "regexp_issue_labeler" "aws_codebuild_data_labels" {
  regexp = "data\s.+aws_codebuild.+"
  labels = ["service/codebuild"]
}

behavior "regexp_issue_labeler" "aws_codecommit_service_labels" {
  regexp = "\*\saws_codecommit.+\n"
  labels = ["service/codecommit"]
}

behavior "regexp_issue_labeler" "aws_codecommit_resource_labels" {
  regexp = "resource\s.+aws_codecommit.+"
  labels = ["service/codecommit"]
}

behavior "regexp_issue_labeler" "aws_codecommit_data_labels" {
  regexp = "data\s.+aws_codecommit.+"
  labels = ["service/codecommit"]
}

behavior "regexp_issue_labeler" "aws_codedeploy_service_labels" {
  regexp = "\*\saws_codedeploy.+\n"
  labels = ["service/codedeploy"]
}

behavior "regexp_issue_labeler" "aws_codedeploy_resource_labels" {
  regexp = "resource\s.+aws_codedeploy.+"
  labels = ["service/codedeploy"]
}

behavior "regexp_issue_labeler" "aws_codedeploy_data_labels" {
  regexp = "data\s.+aws_codedeploy.+"
  labels = ["service/codedeploy"]
}

behavior "regexp_issue_labeler" "aws_codepipeline_service_labels" {
  regexp = "\*\saws_codepipeline.+\n"
  labels = ["service/codepipeline"]
}

behavior "regexp_issue_labeler" "aws_codepipeline_resource_labels" {
  regexp = "resource\s.+aws_codepipeline.+"
  labels = ["service/codepipeline"]
}

behavior "regexp_issue_labeler" "aws_codepipeline_data_labels" {
  regexp = "data\s.+aws_codepipeline.+"
  labels = ["service/codepipeline"]
}

behavior "regexp_issue_labeler" "aws_codestar_service_labels" {
  regexp = "\*\saws_codestar.+\n"
  labels = ["service/codestar"]
}

behavior "regexp_issue_labeler" "aws_codestar_resource_labels" {
  regexp = "resource\s.+aws_codestar.+"
  labels = ["service/codestar"]
}

behavior "regexp_issue_labeler" "aws_codestar_data_labels" {
  regexp = "data\s.+aws_codestar.+"
  labels = ["service/codestar"]
}

behavior "regexp_issue_labeler" "aws_cognito_service_labels" {
  regexp = "\*\saws_cognito.+\n"
  labels = ["service/cognito"]
}

behavior "regexp_issue_labeler" "aws_cognito_resource_labels" {
  regexp = "resource\s.+aws_cognito.+"
  labels = ["service/cognito"]
}

behavior "regexp_issue_labeler" "aws_cognito_data_labels" {
  regexp = "data\s.+aws_cognito.+"
  labels = ["service/cognito"]
}

behavior "regexp_issue_labeler" "aws_config_service_labels" {
  regexp = "\*\saws_config.+\n"
  labels = ["service/configservice"]
}

behavior "regexp_issue_labeler" "aws_config_resource_labels" {
  regexp = "resource\s.+aws_config.+"
  labels = ["service/configservice"]
}

behavior "regexp_issue_labeler" "aws_config_data_labels" {
  regexp = "data\s.+aws_config.+"
  labels = ["service/configservice"]
}

behavior "regexp_issue_labeler" "aws_dms_service_labels" {
  regexp = "\*\saws_dms.+\n"
  labels = ["service/databasemigrationservice"]
}

behavior "regexp_issue_labeler" "aws_dms_resource_labels" {
  regexp = "resource\s.+aws_dms.+"
  labels = ["service/databasemigrationservice"]
}

behavior "regexp_issue_labeler" "aws_dms_data_labels" {
  regexp = "data\s.+aws_dms.+"
  labels = ["service/databasemigrationservice"]
}

behavior "regexp_issue_labeler" "aws_datapipeline_service_labels" {
  regexp = "\*\saws_datapipeline.+\n"
  labels = ["service/datapipeline"]
}

behavior "regexp_issue_labeler" "aws_datapipeline_resource_labels" {
  regexp = "resource\s.+aws_datapipeline.+"
  labels = ["service/datapipeline"]
}

behavior "regexp_issue_labeler" "aws_datapipeline_data_labels" {
  regexp = "data\s.+aws_datapipeline.+"
  labels = ["service/datapipeline"]
}

behavior "regexp_issue_labeler" "aws_datasync_service_labels" {
  regexp = "\*\saws_datasync.+\n"
  labels = ["service/datasync"]
}

behavior "regexp_issue_labeler" "aws_datasync_resource_labels" {
  regexp = "resource\s.+aws_datasync.+"
  labels = ["service/datasync"]
}

behavior "regexp_issue_labeler" "aws_datasync_data_labels" {
  regexp = "data\s.+aws_datasync.+"
  labels = ["service/datasync"]
}

behavior "regexp_issue_labeler" "aws_dax_service_labels" {
  regexp = "\*\saws_dax.+\n"
  labels = ["service/dax"]
}

behavior "regexp_issue_labeler" "aws_dax_resource_labels" {
  regexp = "resource\s.+aws_dax.+"
  labels = ["service/dax"]
}

behavior "regexp_issue_labeler" "aws_dax_data_labels" {
  regexp = "data\s.+aws_dax.+"
  labels = ["service/dax"]
}

behavior "regexp_issue_labeler" "aws_devicefarm_service_labels" {
  regexp = "\*\saws_devicefarm.+\n"
  labels = ["service/devicefarm"]
}

behavior "regexp_issue_labeler" "aws_devicefarm_resource_labels" {
  regexp = "resource\s.+aws_devicefarm.+"
  labels = ["service/devicefarm"]
}

behavior "regexp_issue_labeler" "aws_devicefarm_data_labels" {
  regexp = "data\s.+aws_devicefarm.+"
  labels = ["service/devicefarm"]
}

behavior "regexp_issue_labeler" "aws_directconnect_service_labels" {
  regexp = "\*\saws_directconnect.+\n"
  labels = ["service/directconnect"]
}

behavior "regexp_issue_labeler" "aws_directconnect_resource_labels" {
  regexp = "resource\s.+aws_directconnect.+"
  labels = ["service/directconnect"]
}

behavior "regexp_issue_labeler" "aws_directconnect_data_labels" {
  regexp = "data\s.+aws_directconnect.+"
  labels = ["service/directconnect"]
}

behavior "regexp_issue_labeler" "aws_directoryservice_service_labels" {
  regexp = "\*\saws_directoryservice.+\n"
  labels = ["service/directoryservice"]
}

behavior "regexp_issue_labeler" "aws_directoryservice_resource_labels" {
  regexp = "resource\s.+aws_directoryservice.+"
  labels = ["service/directoryservice"]
}

behavior "regexp_issue_labeler" "aws_directoryservice_data_labels" {
  regexp = "data\s.+aws_directoryservice.+"
  labels = ["service/directoryservice"]
}

behavior "regexp_issue_labeler" "aws_dlm_service_labels" {
  regexp = "\*\saws_dlm.+\n"
  labels = ["service/dlm"]
}

behavior "regexp_issue_labeler" "aws_dlm_resource_labels" {
  regexp = "resource\s.+aws_dlm.+"
  labels = ["service/dlm"]
}

behavior "regexp_issue_labeler" "aws_dlm_data_labels" {
  regexp = "data\s.+aws_dlm.+"
  labels = ["service/dlm"]
}

behavior "regexp_issue_labeler" "aws_dynamodb_service_labels" {
  regexp = "\*\saws_dynamodb.+\n"
  labels = ["service/dynamodb"]
}

behavior "regexp_issue_labeler" "aws_dynamodb_resource_labels" {
  regexp = "resource\s.+aws_dynamodb.+"
  labels = ["service/dynamodb"]
}

behavior "regexp_issue_labeler" "aws_dynamodb_data_labels" {
  regexp = "data\s.+aws_dynamodb.+"
  labels = ["service/dynamodb"]
}

behavior "regexp_issue_labeler" "aws_ec2_service_labels" {
  regexp = "\*\saws_ec2.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_ec2_resource_labels" {
  regexp = "resource\s.+aws_ec2.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_ec2_data_labels" {
  regexp = "data\s.+aws_ec2.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_ami_service_labels" {
  regexp = "\*\saws_ami.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_ami_resource_labels" {
  regexp = "resource\s.+aws_ami.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_ami_data_labels" {
  regexp = "data\s.+aws_ami.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_availability_zone_service_labels" {
  regexp = "\*\saws_availability_zone.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_availability_zone_resource_labels" {
  regexp = "resource\s.+aws_availability_zone.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_availability_zone_data_labels" {
  regexp = "data\s.+aws_availability_zone.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_customer_gateway_service_labels" {
  regexp = "\*\saws_customer_gateway.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_customer_gateway_resource_labels" {
  regexp = "resource\s.+aws_customer_gateway.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_customer_gateway_data_labels" {
  regexp = "data\s.+aws_customer_gateway.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_network_acl_service_labels" {
  regexp = "\*\saws_default_network_acl.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_network_acl_resource_labels" {
  regexp = "resource\s.+aws_default_network_acl.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_network_acl_data_labels" {
  regexp = "data\s.+aws_default_network_acl.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_route_table_service_labels" {
  regexp = "\*\saws_default_route_table.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_route_table_resource_labels" {
  regexp = "resource\s.+aws_default_route_table.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_route_table_data_labels" {
  regexp = "data\s.+aws_default_route_table.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_security_group_service_labels" {
  regexp = "\*\saws_default_security_group.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_security_group_resource_labels" {
  regexp = "resource\s.+aws_default_security_group.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_security_group_data_labels" {
  regexp = "data\s.+aws_default_security_group.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_subnet_service_labels" {
  regexp = "\*\saws_default_subnet.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_subnet_resource_labels" {
  regexp = "resource\s.+aws_default_subnet.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_subnet_data_labels" {
  regexp = "data\s.+aws_default_subnet.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_vpc_service_labels" {
  regexp = "\*\saws_default_vpc.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_vpc_resource_labels" {
  regexp = "resource\s.+aws_default_vpc.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_default_vpc_data_labels" {
  regexp = "data\s.+aws_default_vpc.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_ebs_service_labels" {
  regexp = "\*\saws_ebs.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_ebs_resource_labels" {
  regexp = "resource\s.+aws_ebs.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_ebs_data_labels" {
  regexp = "data\s.+aws_ebs.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_egress_only_internet_gateway_service_labels" {
  regexp = "\*\saws_egress_only_internet_gateway.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_egress_only_internet_gateway_resource_labels" {
  regexp = "resource\s.+aws_egress_only_internet_gateway.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_egress_only_internet_gateway_data_labels" {
  regexp = "data\s.+aws_egress_only_internet_gateway.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_eip_service_labels" {
  regexp = "\*\saws_eip.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_eip_resource_labels" {
  regexp = "resource\s.+aws_eip.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_eip_data_labels" {
  regexp = "data\s.+aws_eip.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_flow_log_service_labels" {
  regexp = "\*\saws_flow_log.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_flow_log_resource_labels" {
  regexp = "resource\s.+aws_flow_log.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_flow_log_data_labels" {
  regexp = "data\s.+aws_flow_log.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_instance_service_labels" {
  regexp = "\*\saws_instance.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_instance_resource_labels" {
  regexp = "resource\s.+aws_instance.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_instance_data_labels" {
  regexp = "data\s.+aws_instance.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_internet_gateway_service_labels" {
  regexp = "\*\saws_internet_gateway.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_internet_gateway_resource_labels" {
  regexp = "resource\s.+aws_internet_gateway.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_internet_gateway_data_labels" {
  regexp = "data\s.+aws_internet_gateway.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_key_pair_service_labels" {
  regexp = "\*\saws_key_pair.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_key_pair_resource_labels" {
  regexp = "resource\s.+aws_key_pair.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_key_pair_data_labels" {
  regexp = "data\s.+aws_key_pair.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_launch_template_service_labels" {
  regexp = "\*\saws_launch_template.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_launch_template_resource_labels" {
  regexp = "resource\s.+aws_launch_template.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_launch_template_data_labels" {
  regexp = "data\s.+aws_launch_template.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_main_route_table_association_service_labels" {
  regexp = "\*\saws_main_route_table_association.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_main_route_table_association_resource_labels" {
  regexp = "resource\s.+aws_main_route_table_association.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_main_route_table_association_data_labels" {
  regexp = "data\s.+aws_main_route_table_association.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_nat_gateway_service_labels" {
  regexp = "\*\saws_nat_gateway.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_nat_gateway_resource_labels" {
  regexp = "resource\s.+aws_nat_gateway.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_nat_gateway_data_labels" {
  regexp = "data\s.+aws_nat_gateway.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_network_acl_service_labels" {
  regexp = "\*\saws_network_acl.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_network_acl_resource_labels" {
  regexp = "resource\s.+aws_network_acl.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_network_acl_data_labels" {
  regexp = "data\s.+aws_network_acl.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_network_interface_service_labels" {
  regexp = "\*\saws_network_interface.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_network_interface_resource_labels" {
  regexp = "resource\s.+aws_network_interface.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_network_interface_data_labels" {
  regexp = "data\s.+aws_network_interface.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_placement_group_service_labels" {
  regexp = "\*\saws_placement_group.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_placement_group_resource_labels" {
  regexp = "resource\s.+aws_placement_group.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_placement_group_data_labels" {
  regexp = "data\s.+aws_placement_group.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_route_table_service_labels" {
  regexp = "\*\saws_route_table.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_route_table_resource_labels" {
  regexp = "resource\s.+aws_route_table.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_route_table_data_labels" {
  regexp = "data\s.+aws_route_table.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_route_service_labels" {
  regexp = "\*\saws_route.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_route_resource_labels" {
  regexp = "resource\s.+aws_route.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_route_data_labels" {
  regexp = "data\s.+aws_route.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_security_group_service_labels" {
  regexp = "\*\saws_security_group.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_security_group_resource_labels" {
  regexp = "resource\s.+aws_security_group.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_security_group_data_labels" {
  regexp = "data\s.+aws_security_group.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_spot_service_labels" {
  regexp = "\*\saws_spot.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_spot_resource_labels" {
  regexp = "resource\s.+aws_spot.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_spot_data_labels" {
  regexp = "data\s.+aws_spot.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_subnet_service_labels" {
  regexp = "\*\saws_subnet.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_subnet_resource_labels" {
  regexp = "resource\s.+aws_subnet.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_subnet_data_labels" {
  regexp = "data\s.+aws_subnet.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_vpc_service_labels" {
  regexp = "\*\saws_vpc.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_vpc_resource_labels" {
  regexp = "resource\s.+aws_vpc.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_vpc_data_labels" {
  regexp = "data\s.+aws_vpc.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_vpn_service_labels" {
  regexp = "\*\saws_vpn.+\n"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_vpn_resource_labels" {
  regexp = "resource\s.+aws_vpn.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_vpn_data_labels" {
  regexp = "data\s.+aws_vpn.+"
  labels = ["service/ec2"]
}

behavior "regexp_issue_labeler" "aws_ecr_service_labels" {
  regexp = "\*\saws_ecr.+\n"
  labels = ["service/ecr"]
}

behavior "regexp_issue_labeler" "aws_ecr_resource_labels" {
  regexp = "resource\s.+aws_ecr.+"
  labels = ["service/ecr"]
}

behavior "regexp_issue_labeler" "aws_ecr_data_labels" {
  regexp = "data\s.+aws_ecr.+"
  labels = ["service/ecr"]
}

behavior "regexp_issue_labeler" "aws_esc_service_labels" {
  regexp = "\*\saws_esc.+\n"
  labels = ["service/ecs"]
}

behavior "regexp_issue_labeler" "aws_esc_resource_labels" {
  regexp = "resource\s.+aws_esc.+"
  labels = ["service/ecs"]
}

behavior "regexp_issue_labeler" "aws_esc_data_labels" {
  regexp = "data\s.+aws_esc.+"
  labels = ["service/ecs"]
}

behavior "regexp_issue_labeler" "aws_efs_service_labels" {
  regexp = "\*\saws_efs.+\n"
  labels = ["service/efs"]
}

behavior "regexp_issue_labeler" "aws_efs_resource_labels" {
  regexp = "resource\s.+aws_efs.+"
  labels = ["service/efs"]
}

behavior "regexp_issue_labeler" "aws_efs_data_labels" {
  regexp = "data\s.+aws_efs.+"
  labels = ["service/efs"]
}

behavior "regexp_issue_labeler" "aws_eks_service_labels" {
  regexp = "\*\saws_eks.+\n"
  labels = ["service/eks"]
}

behavior "regexp_issue_labeler" "aws_eks_resource_labels" {
  regexp = "resource\s.+aws_eks.+"
  labels = ["service/eks"]
}

behavior "regexp_issue_labeler" "aws_eks_data_labels" {
  regexp = "data\s.+aws_eks.+"
  labels = ["service/eks"]
}

behavior "regexp_issue_labeler" "aws_elastic_transcoder_service_labels" {
  regexp = "\*\saws_elastic_transcoder.+\n"
  labels = ["service/elastic-transcoder"]
}

behavior "regexp_issue_labeler" "aws_elastic_transcoder_resource_labels" {
  regexp = "resource\s.+aws_elastic_transcoder.+"
  labels = ["service/elastic-transcoder"]
}

behavior "regexp_issue_labeler" "aws_elastic_transcoder_data_labels" {
  regexp = "data\s.+aws_elastic_transcoder.+"
  labels = ["service/elastic-transcoder"]
}

behavior "regexp_issue_labeler" "aws_elasticache_service_labels" {
  regexp = "\*\saws_elasticache.+\n"
  labels = ["service/elasticache"]
}

behavior "regexp_issue_labeler" "aws_elasticache_resource_labels" {
  regexp = "resource\s.+aws_elasticache.+"
  labels = ["service/elasticache"]
}

behavior "regexp_issue_labeler" "aws_elasticache_data_labels" {
  regexp = "data\s.+aws_elasticache.+"
  labels = ["service/elasticache"]
}

behavior "regexp_issue_labeler" "aws_elasticbeanstalk_service_labels" {
  regexp = "\*\saws_elasticbeanstalk.+\n"
  labels = ["service/elasticbeanstalk"]
}

behavior "regexp_issue_labeler" "aws_elasticbeanstalk_resource_labels" {
  regexp = "resource\s.+aws_elasticbeanstalk.+"
  labels = ["service/elasticbeanstalk"]
}

behavior "regexp_issue_labeler" "aws_elasticbeanstalk_data_labels" {
  regexp = "data\s.+aws_elasticbeanstalk.+"
  labels = ["service/elasticbeanstalk"]
}

behavior "regexp_issue_labeler" "aws_elasticsearch_service_labels" {
  regexp = "\*\saws_elasticsearch.+\n"
  labels = ["service/elasticsearch"]
}

behavior "regexp_issue_labeler" "aws_elasticsearch_resource_labels" {
  regexp = "resource\s.+aws_elasticsearch.+"
  labels = ["service/elasticsearch"]
}

behavior "regexp_issue_labeler" "aws_elasticsearch_data_labels" {
  regexp = "data\s.+aws_elasticsearch.+"
  labels = ["service/elasticsearch"]
}

behavior "regexp_issue_labeler" "aws_app_cookie_stickiness_policy_service_labels" {
  regexp = "\*\saws_app_cookie_stickiness_policy.+\n"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_app_cookie_stickiness_policy_resource_labels" {
  regexp = "resource\s.+aws_app_cookie_stickiness_policy.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_app_cookie_stickiness_policy_data_labels" {
  regexp = "data\s.+aws_app_cookie_stickiness_policy.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_elb_service_labels" {
  regexp = "\*\saws_elb.+\n"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_elb_resource_labels" {
  regexp = "resource\s.+aws_elb.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_elb_data_labels" {
  regexp = "data\s.+aws_elb.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_lb_cookie_stickiness_policy_service_labels" {
  regexp = "\*\saws_lb_cookie_stickiness_policy.+\n"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_lb_cookie_stickiness_policy_resource_labels" {
  regexp = "resource\s.+aws_lb_cookie_stickiness_policy.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_lb_cookie_stickiness_policy_data_labels" {
  regexp = "data\s.+aws_lb_cookie_stickiness_policy.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_lb_ssl_negotiation_policy_service_labels" {
  regexp = "\*\saws_lb_ssl_negotiation_policy.+\n"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_lb_ssl_negotiation_policy_resource_labels" {
  regexp = "resource\s.+aws_lb_ssl_negotiation_policy.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_lb_ssl_negotiation_policy_data_labels" {
  regexp = "data\s.+aws_lb_ssl_negotiation_policy.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_proxy_protocol_policy_service_labels" {
  regexp = "\*\saws_proxy_protocol_policy.+\n"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_proxy_protocol_policy_resource_labels" {
  regexp = "resource\s.+aws_proxy_protocol_policy.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_proxy_protocol_policy_data_labels" {
  regexp = "data\s.+aws_proxy_protocol_policy.+"
  labels = ["service/elb"]
}

behavior "regexp_issue_labeler" "aws_lb_service_labels" {
  regexp = "\*\saws_lb.+\n"
  labels = ["service/elbv2"]
}

behavior "regexp_issue_labeler" "aws_lb_resource_labels" {
  regexp = "resource\s.+aws_lb.+"
  labels = ["service/elbv2"]
}

behavior "regexp_issue_labeler" "aws_lb_data_labels" {
  regexp = "data\s.+aws_lb.+"
  labels = ["service/elbv2"]
}

behavior "regexp_issue_labeler" "aws_lb_listener_service_labels" {
  regexp = "\*\saws_lb_listener.+\n"
  labels = ["service/elbv2"]
}

behavior "regexp_issue_labeler" "aws_lb_listener_resource_labels" {
  regexp = "resource\s.+aws_lb_listener.+"
  labels = ["service/elbv2"]
}

behavior "regexp_issue_labeler" "aws_lb_listener_data_labels" {
  regexp = "data\s.+aws_lb_listener.+"
  labels = ["service/elbv2"]
}

behavior "regexp_issue_labeler" "aws_lb_target_group_service_labels" {
  regexp = "\*\saws_lb_target_group.+\n"
  labels = ["service/elbv2"]
}

behavior "regexp_issue_labeler" "aws_lb_target_group_resource_labels" {
  regexp = "resource\s.+aws_lb_target_group.+"
  labels = ["service/elbv2"]
}

behavior "regexp_issue_labeler" "aws_lb_target_group_data_labels" {
  regexp = "data\s.+aws_lb_target_group.+"
  labels = ["service/elbv2"]
}

behavior "regexp_issue_labeler" "aws_emr_service_labels" {
  regexp = "\*\saws_emr.+\n"
  labels = ["service/emr"]
}

behavior "regexp_issue_labeler" "aws_emr_resource_labels" {
  regexp = "resource\s.+aws_emr.+"
  labels = ["service/emr"]
}

behavior "regexp_issue_labeler" "aws_emr_data_labels" {
  regexp = "data\s.+aws_emr.+"
  labels = ["service/emr"]
}

behavior "regexp_issue_labeler" "aws_firehose_service_labels" {
  regexp = "\*\saws_firehose.+\n"
  labels = ["service/firehose"]
}

behavior "regexp_issue_labeler" "aws_firehose_resource_labels" {
  regexp = "resource\s.+aws_firehose.+"
  labels = ["service/firehose"]
}

behavior "regexp_issue_labeler" "aws_firehose_data_labels" {
  regexp = "data\s.+aws_firehose.+"
  labels = ["service/firehose"]
}

behavior "regexp_issue_labeler" "aws_fms_service_labels" {
  regexp = "\*\saws_fms.+\n"
  labels = ["service/fms"]
}

behavior "regexp_issue_labeler" "aws_fms_resource_labels" {
  regexp = "resource\s.+aws_fms.+"
  labels = ["service/fms"]
}

behavior "regexp_issue_labeler" "aws_fms_data_labels" {
  regexp = "data\s.+aws_fms.+"
  labels = ["service/fms"]
}

behavior "regexp_issue_labeler" "aws_fsx_service_labels" {
  regexp = "\*\saws_fsx.+\n"
  labels = ["service/fsx"]
}

behavior "regexp_issue_labeler" "aws_fsx_resource_labels" {
  regexp = "resource\s.+aws_fsx.+"
  labels = ["service/fsx"]
}

behavior "regexp_issue_labeler" "aws_fsx_data_labels" {
  regexp = "data\s.+aws_fsx.+"
  labels = ["service/fsx"]
}

behavior "regexp_issue_labeler" "aws_gamelift_service_labels" {
  regexp = "\*\saws_gamelift.+\n"
  labels = ["service/gamelift"]
}

behavior "regexp_issue_labeler" "aws_gamelift_resource_labels" {
  regexp = "resource\s.+aws_gamelift.+"
  labels = ["service/gamelift"]
}

behavior "regexp_issue_labeler" "aws_gamelift_data_labels" {
  regexp = "data\s.+aws_gamelift.+"
  labels = ["service/gamelift"]
}

behavior "regexp_issue_labeler" "aws_glacier_service_labels" {
  regexp = "\*\saws_glacier.+\n"
  labels = ["service/glacier"]
}

behavior "regexp_issue_labeler" "aws_glacier_resource_labels" {
  regexp = "resource\s.+aws_glacier.+"
  labels = ["service/glacier"]
}

behavior "regexp_issue_labeler" "aws_glacier_data_labels" {
  regexp = "data\s.+aws_glacier.+"
  labels = ["service/glacier"]
}

behavior "regexp_issue_labeler" "aws_globalaccelerator_service_labels" {
  regexp = "\*\saws_globalaccelerator.+\n"
  labels = ["service/globalaccelerator"]
}

behavior "regexp_issue_labeler" "aws_globalaccelerator_resource_labels" {
  regexp = "resource\s.+aws_globalaccelerator.+"
  labels = ["service/globalaccelerator"]
}

behavior "regexp_issue_labeler" "aws_globalaccelerator_data_labels" {
  regexp = "data\s.+aws_globalaccelerator.+"
  labels = ["service/globalaccelerator"]
}

behavior "regexp_issue_labeler" "aws_glue_service_labels" {
  regexp = "\*\saws_glue.+\n"
  labels = ["service/glue"]
}

behavior "regexp_issue_labeler" "aws_glue_resource_labels" {
  regexp = "resource\s.+aws_glue.+"
  labels = ["service/glue"]
}

behavior "regexp_issue_labeler" "aws_glue_data_labels" {
  regexp = "data\s.+aws_glue.+"
  labels = ["service/glue"]
}

behavior "regexp_issue_labeler" "aws_greengrass_service_labels" {
  regexp = "\*\saws_greengrass.+\n"
  labels = ["service/greengrass"]
}

behavior "regexp_issue_labeler" "aws_greengrass_resource_labels" {
  regexp = "resource\s.+aws_greengrass.+"
  labels = ["service/greengrass"]
}

behavior "regexp_issue_labeler" "aws_greengrass_data_labels" {
  regexp = "data\s.+aws_greengrass.+"
  labels = ["service/greengrass"]
}

behavior "regexp_issue_labeler" "aws_guardduty_service_labels" {
  regexp = "\*\saws_guardduty.+\n"
  labels = ["service/guardduty"]
}

behavior "regexp_issue_labeler" "aws_guardduty_resource_labels" {
  regexp = "resource\s.+aws_guardduty.+"
  labels = ["service/guardduty"]
}

behavior "regexp_issue_labeler" "aws_guardduty_data_labels" {
  regexp = "data\s.+aws_guardduty.+"
  labels = ["service/guardduty"]
}

behavior "regexp_issue_labeler" "aws_iam_service_labels" {
  regexp = "\*\saws_iam.+\n"
  labels = ["service/iam"]
}

behavior "regexp_issue_labeler" "aws_iam_resource_labels" {
  regexp = "resource\s.+aws_iam.+"
  labels = ["service/iam"]
}

behavior "regexp_issue_labeler" "aws_iam_data_labels" {
  regexp = "data\s.+aws_iam.+"
  labels = ["service/iam"]
}

behavior "regexp_issue_labeler" "aws_inspector_service_labels" {
  regexp = "\*\saws_inspector.+\n"
  labels = ["service/inspector"]
}

behavior "regexp_issue_labeler" "aws_inspector_resource_labels" {
  regexp = "resource\s.+aws_inspector.+"
  labels = ["service/inspector"]
}

behavior "regexp_issue_labeler" "aws_inspector_data_labels" {
  regexp = "data\s.+aws_inspector.+"
  labels = ["service/inspector"]
}

behavior "regexp_issue_labeler" "aws_iot_service_labels" {
  regexp = "\*\saws_iot.+\n"
  labels = ["service/iot"]
}

behavior "regexp_issue_labeler" "aws_iot_resource_labels" {
  regexp = "resource\s.+aws_iot.+"
  labels = ["service/iot"]
}

behavior "regexp_issue_labeler" "aws_iot_data_labels" {
  regexp = "data\s.+aws_iot.+"
  labels = ["service/iot"]
}

behavior "regexp_issue_labeler" "aws_kinesis_service_labels" {
  regexp = "\*\saws_kinesis.+\n"
  labels = ["service/kinesis"]
}

behavior "regexp_issue_labeler" "aws_kinesis_resource_labels" {
  regexp = "resource\s.+aws_kinesis.+"
  labels = ["service/kinesis"]
}

behavior "regexp_issue_labeler" "aws_kinesis_data_labels" {
  regexp = "data\s.+aws_kinesis.+"
  labels = ["service/kinesis"]
}

behavior "regexp_issue_labeler" "aws_kms_service_labels" {
  regexp = "\*\saws_kms.+\n"
  labels = ["service/kms"]
}

behavior "regexp_issue_labeler" "aws_kms_resource_labels" {
  regexp = "resource\s.+aws_kms.+"
  labels = ["service/kms"]
}

behavior "regexp_issue_labeler" "aws_kms_data_labels" {
  regexp = "data\s.+aws_kms.+"
  labels = ["service/kms"]
}

behavior "regexp_issue_labeler" "aws_lambda_service_labels" {
  regexp = "\*\saws_lambda.+\n"
  labels = ["service/lambda"]
}

behavior "regexp_issue_labeler" "aws_lambda_resource_labels" {
  regexp = "resource\s.+aws_lambda.+"
  labels = ["service/lambda"]
}

behavior "regexp_issue_labeler" "aws_lambda_data_labels" {
  regexp = "data\s.+aws_lambda.+"
  labels = ["service/lambda"]
}

behavior "regexp_issue_labeler" "aws_lex_service_labels" {
  regexp = "\*\saws_lex.+\n"
  labels = ["service/lexmodelbuildingservice"]
}

behavior "regexp_issue_labeler" "aws_lex_resource_labels" {
  regexp = "resource\s.+aws_lex.+"
  labels = ["service/lexmodelbuildingservice"]
}

behavior "regexp_issue_labeler" "aws_lex_data_labels" {
  regexp = "data\s.+aws_lex.+"
  labels = ["service/lexmodelbuildingservice"]
}

behavior "regexp_issue_labeler" "aws_licensemanager_service_labels" {
  regexp = "\*\saws_licensemanager.+\n"
  labels = ["service/licensemanager"]
}

behavior "regexp_issue_labeler" "aws_licensemanager_resource_labels" {
  regexp = "resource\s.+aws_licensemanager.+"
  labels = ["service/licensemanager"]
}

behavior "regexp_issue_labeler" "aws_licensemanager_data_labels" {
  regexp = "data\s.+aws_licensemanager.+"
  labels = ["service/licensemanager"]
}

behavior "regexp_issue_labeler" "aws_lightsail_service_labels" {
  regexp = "\*\saws_lightsail.+\n"
  labels = ["service/lightsail"]
}

behavior "regexp_issue_labeler" "aws_lightsail_resource_labels" {
  regexp = "resource\s.+aws_lightsail.+"
  labels = ["service/lightsail"]
}

behavior "regexp_issue_labeler" "aws_lightsail_data_labels" {
  regexp = "data\s.+aws_lightsail.+"
  labels = ["service/lightsail"]
}

behavior "regexp_issue_labeler" "aws_machinelearning_service_labels" {
  regexp = "\*\saws_machinelearning.+\n"
  labels = ["service/machinelearning"]
}

behavior "regexp_issue_labeler" "aws_machinelearning_resource_labels" {
  regexp = "resource\s.+aws_machinelearning.+"
  labels = ["service/machinelearning"]
}

behavior "regexp_issue_labeler" "aws_machinelearning_data_labels" {
  regexp = "data\s.+aws_machinelearning.+"
  labels = ["service/machinelearning"]
}

behavior "regexp_issue_labeler" "aws_macie_service_labels" {
  regexp = "\*\saws_macie.+\n"
  labels = ["service/macie"]
}

behavior "regexp_issue_labeler" "aws_macie_resource_labels" {
  regexp = "resource\s.+aws_macie.+"
  labels = ["service/macie"]
}

behavior "regexp_issue_labeler" "aws_macie_data_labels" {
  regexp = "data\s.+aws_macie.+"
  labels = ["service/macie"]
}

behavior "regexp_issue_labeler" "aws_media_connect_service_labels" {
  regexp = "\*\saws_media_connect.+\n"
  labels = ["service/mediaconnect"]
}

behavior "regexp_issue_labeler" "aws_media_connect_resource_labels" {
  regexp = "resource\s.+aws_media_connect.+"
  labels = ["service/mediaconnect"]
}

behavior "regexp_issue_labeler" "aws_media_connect_data_labels" {
  regexp = "data\s.+aws_media_connect.+"
  labels = ["service/mediaconnect"]
}

behavior "regexp_issue_labeler" "aws_media_convert_service_labels" {
  regexp = "\*\saws_media_convert.+\n"
  labels = ["service/mediaconvert"]
}

behavior "regexp_issue_labeler" "aws_media_convert_resource_labels" {
  regexp = "resource\s.+aws_media_convert.+"
  labels = ["service/mediaconvert"]
}

behavior "regexp_issue_labeler" "aws_media_convert_data_labels" {
  regexp = "data\s.+aws_media_convert.+"
  labels = ["service/mediaconvert"]
}

behavior "regexp_issue_labeler" "aws_media_live_service_labels" {
  regexp = "\*\saws_media_live.+\n"
  labels = ["service/medialive"]
}

behavior "regexp_issue_labeler" "aws_media_live_resource_labels" {
  regexp = "resource\s.+aws_media_live.+"
  labels = ["service/medialive"]
}

behavior "regexp_issue_labeler" "aws_media_live_data_labels" {
  regexp = "data\s.+aws_media_live.+"
  labels = ["service/medialive"]
}

behavior "regexp_issue_labeler" "aws_media_package_service_labels" {
  regexp = "\*\saws_media_package.+\n"
  labels = ["service/mediapackage"]
}

behavior "regexp_issue_labeler" "aws_media_package_resource_labels" {
  regexp = "resource\s.+aws_media_package.+"
  labels = ["service/mediapackage"]
}

behavior "regexp_issue_labeler" "aws_media_package_data_labels" {
  regexp = "data\s.+aws_media_package.+"
  labels = ["service/mediapackage"]
}

behavior "regexp_issue_labeler" "aws_media_store_service_labels" {
  regexp = "\*\saws_media_store.+\n"
  labels = ["service/mediastore"]
}

behavior "regexp_issue_labeler" "aws_media_store_resource_labels" {
  regexp = "resource\s.+aws_media_store.+"
  labels = ["service/mediastore"]
}

behavior "regexp_issue_labeler" "aws_media_store_data_labels" {
  regexp = "data\s.+aws_media_store.+"
  labels = ["service/mediastore"]
}

behavior "regexp_issue_labeler" "aws_media_tailor_service_labels" {
  regexp = "\*\saws_media_tailor.+\n"
  labels = ["service/mediatailor"]
}

behavior "regexp_issue_labeler" "aws_media_tailor_resource_labels" {
  regexp = "resource\s.+aws_media_tailor.+"
  labels = ["service/mediatailor"]
}

behavior "regexp_issue_labeler" "aws_media_tailor_data_labels" {
  regexp = "data\s.+aws_media_tailor.+"
  labels = ["service/mediatailor"]
}

behavior "regexp_issue_labeler" "aws_mobile_service_labels" {
  regexp = "\*\saws_mobile.+\n"
  labels = ["service/mobile"]
}

behavior "regexp_issue_labeler" "aws_mobile_resource_labels" {
  regexp = "resource\s.+aws_mobile.+"
  labels = ["service/mobile"]
}

behavior "regexp_issue_labeler" "aws_mobile_data_labels" {
  regexp = "data\s.+aws_mobile.+"
  labels = ["service/mobile"]
}

behavior "regexp_issue_labeler" "aws_mq_service_labels" {
  regexp = "\*\saws_mq.+\n"
  labels = ["service/mq"]
}

behavior "regexp_issue_labeler" "aws_mq_resource_labels" {
  regexp = "resource\s.+aws_mq.+"
  labels = ["service/mq"]
}

behavior "regexp_issue_labeler" "aws_mq_data_labels" {
  regexp = "data\s.+aws_mq.+"
  labels = ["service/mq"]
}

behavior "regexp_issue_labeler" "aws_neptune_service_labels" {
  regexp = "\*\saws_neptune.+\n"
  labels = ["service/neptune"]
}

behavior "regexp_issue_labeler" "aws_neptune_resource_labels" {
  regexp = "resource\s.+aws_neptune.+"
  labels = ["service/neptune"]
}

behavior "regexp_issue_labeler" "aws_neptune_data_labels" {
  regexp = "data\s.+aws_neptune.+"
  labels = ["service/neptune"]
}

behavior "regexp_issue_labeler" "aws_opsworks_service_labels" {
  regexp = "\*\saws_opsworks.+\n"
  labels = ["service/opsworks"]
}

behavior "regexp_issue_labeler" "aws_opsworks_resource_labels" {
  regexp = "resource\s.+aws_opsworks.+"
  labels = ["service/opsworks"]
}

behavior "regexp_issue_labeler" "aws_opsworks_data_labels" {
  regexp = "data\s.+aws_opsworks.+"
  labels = ["service/opsworks"]
}

behavior "regexp_issue_labeler" "aws_organizations_service_labels" {
  regexp = "\*\saws_organizations.+\n"
  labels = ["service/organizations"]
}

behavior "regexp_issue_labeler" "aws_organizations_resource_labels" {
  regexp = "resource\s.+aws_organizations.+"
  labels = ["service/organizations"]
}

behavior "regexp_issue_labeler" "aws_organizations_data_labels" {
  regexp = "data\s.+aws_organizations.+"
  labels = ["service/organizations"]
}

behavior "regexp_issue_labeler" "aws_pinpoint_service_labels" {
  regexp = "\*\saws_pinpoint.+\n"
  labels = ["service/pinpoint"]
}

behavior "regexp_issue_labeler" "aws_pinpoint_resource_labels" {
  regexp = "resource\s.+aws_pinpoint.+"
  labels = ["service/pinpoint"]
}

behavior "regexp_issue_labeler" "aws_pinpoint_data_labels" {
  regexp = "data\s.+aws_pinpoint.+"
  labels = ["service/pinpoint"]
}

behavior "regexp_issue_labeler" "aws_polly_service_labels" {
  regexp = "\*\saws_polly.+\n"
  labels = ["service/polly"]
}

behavior "regexp_issue_labeler" "aws_polly_resource_labels" {
  regexp = "resource\s.+aws_polly.+"
  labels = ["service/polly"]
}

behavior "regexp_issue_labeler" "aws_polly_data_labels" {
  regexp = "data\s.+aws_polly.+"
  labels = ["service/polly"]
}

behavior "regexp_issue_labeler" "aws_pricing_service_labels" {
  regexp = "\*\saws_pricing.+\n"
  labels = ["service/pricing"]
}

behavior "regexp_issue_labeler" "aws_pricing_resource_labels" {
  regexp = "resource\s.+aws_pricing.+"
  labels = ["service/pricing"]
}

behavior "regexp_issue_labeler" "aws_pricing_data_labels" {
  regexp = "data\s.+aws_pricing.+"
  labels = ["service/pricing"]
}

behavior "regexp_issue_labeler" "aws_ram_service_labels" {
  regexp = "\*\saws_ram.+\n"
  labels = ["service/ram"]
}

behavior "regexp_issue_labeler" "aws_ram_resource_labels" {
  regexp = "resource\s.+aws_ram.+"
  labels = ["service/ram"]
}

behavior "regexp_issue_labeler" "aws_ram_data_labels" {
  regexp = "data\s.+aws_ram.+"
  labels = ["service/ram"]
}

behavior "regexp_issue_labeler" "aws_db_service_labels" {
  regexp = "\*\saws_db.+\n"
  labels = ["service/rds"]
}

behavior "regexp_issue_labeler" "aws_db_resource_labels" {
  regexp = "resource\s.+aws_db.+"
  labels = ["service/rds"]
}

behavior "regexp_issue_labeler" "aws_db_data_labels" {
  regexp = "data\s.+aws_db.+"
  labels = ["service/rds"]
}

behavior "regexp_issue_labeler" "aws_rds_service_labels" {
  regexp = "\*\saws_rds.+\n"
  labels = ["service/rds"]
}

behavior "regexp_issue_labeler" "aws_rds_resource_labels" {
  regexp = "resource\s.+aws_rds.+"
  labels = ["service/rds"]
}

behavior "regexp_issue_labeler" "aws_rds_data_labels" {
  regexp = "data\s.+aws_rds.+"
  labels = ["service/rds"]
}

behavior "regexp_issue_labeler" "aws_redshift_service_labels" {
  regexp = "\*\saws_redshift.+\n"
  labels = ["service/redshift"]
}

behavior "regexp_issue_labeler" "aws_redshift_resource_labels" {
  regexp = "resource\s.+aws_redshift.+"
  labels = ["service/redshift"]
}

behavior "regexp_issue_labeler" "aws_redshift_data_labels" {
  regexp = "data\s.+aws_redshift.+"
  labels = ["service/redshift"]
}

behavior "regexp_issue_labeler" "aws_resourcegroups_service_labels" {
  regexp = "\*\saws_resourcegroups.+\n"
  labels = ["service/resourcegroups"]
}

behavior "regexp_issue_labeler" "aws_resourcegroups_resource_labels" {
  regexp = "resource\s.+aws_resourcegroups.+"
  labels = ["service/resourcegroups"]
}

behavior "regexp_issue_labeler" "aws_resourcegroups_data_labels" {
  regexp = "data\s.+aws_resourcegroups.+"
  labels = ["service/resourcegroups"]
}

behavior "regexp_issue_labeler" "aws_route53_delegation_set_service_labels" {
  regexp = "\*\saws_route53_delegation_set.+\n"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_delegation_set_resource_labels" {
  regexp = "resource\s.+aws_route53_delegation_set.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_delegation_set_data_labels" {
  regexp = "data\s.+aws_route53_delegation_set.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_health_check_service_labels" {
  regexp = "\*\saws_route53_health_check.+\n"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_health_check_resource_labels" {
  regexp = "resource\s.+aws_route53_health_check.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_health_check_data_labels" {
  regexp = "data\s.+aws_route53_health_check.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_query_log_service_labels" {
  regexp = "\*\saws_route53_query_log.+\n"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_query_log_resource_labels" {
  regexp = "resource\s.+aws_route53_query_log.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_query_log_data_labels" {
  regexp = "data\s.+aws_route53_query_log.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_record_service_labels" {
  regexp = "\*\saws_route53_record.+\n"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_record_resource_labels" {
  regexp = "resource\s.+aws_route53_record.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_record_data_labels" {
  regexp = "data\s.+aws_route53_record.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_zone_service_labels" {
  regexp = "\*\saws_route53_zone.+\n"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_zone_resource_labels" {
  regexp = "resource\s.+aws_route53_zone.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_route53_zone_data_labels" {
  regexp = "data\s.+aws_route53_zone.+"
  labels = ["service/route53"]
}

behavior "regexp_issue_labeler" "aws_robomaker_service_labels" {
  regexp = "\*\saws_robomaker.+\n"
  labels = ["service/robomaker"]
}

behavior "regexp_issue_labeler" "aws_robomaker_resource_labels" {
  regexp = "resource\s.+aws_robomaker.+"
  labels = ["service/robomaker"]
}

behavior "regexp_issue_labeler" "aws_robomaker_data_labels" {
  regexp = "data\s.+aws_robomaker.+"
  labels = ["service/robomaker"]
}

behavior "regexp_issue_labeler" "aws_route53_domains_service_labels" {
  regexp = "\*\saws_route53_domains.+\n"
  labels = ["service/route53domains"]
}

behavior "regexp_issue_labeler" "aws_route53_domains_resource_labels" {
  regexp = "resource\s.+aws_route53_domains.+"
  labels = ["service/route53domains"]
}

behavior "regexp_issue_labeler" "aws_route53_domains_data_labels" {
  regexp = "data\s.+aws_route53_domains.+"
  labels = ["service/route53domains"]
}

behavior "regexp_issue_labeler" "aws_s3_bucket_service_labels" {
  regexp = "\*\saws_s3_bucket.+\n"
  labels = ["service/s3"]
}

behavior "regexp_issue_labeler" "aws_s3_bucket_resource_labels" {
  regexp = "resource\s.+aws_s3_bucket.+"
  labels = ["service/s3"]
}

behavior "regexp_issue_labeler" "aws_s3_bucket_data_labels" {
  regexp = "data\s.+aws_s3_bucket.+"
  labels = ["service/s3"]
}

behavior "regexp_issue_labeler" "aws_aws_canonical_user_id_service_labels" {
  regexp = "\*\saws_aws_canonical_user_id.+\n"
  labels = ["service/s3"]
}

behavior "regexp_issue_labeler" "aws_aws_canonical_user_id_resource_labels" {
  regexp = "resource\s.+aws_aws_canonical_user_id.+"
  labels = ["service/s3"]
}

behavior "regexp_issue_labeler" "aws_aws_canonical_user_id_data_labels" {
  regexp = "data\s.+aws_aws_canonical_user_id.+"
  labels = ["service/s3"]
}

behavior "regexp_issue_labeler" "aws_s3_account_service_labels" {
  regexp = "\*\saws_s3_account.+\n"
  labels = ["service/s3control"]
}

behavior "regexp_issue_labeler" "aws_s3_account_resource_labels" {
  regexp = "resource\s.+aws_s3_account.+"
  labels = ["service/s3control"]
}

behavior "regexp_issue_labeler" "aws_s3_account_data_labels" {
  regexp = "data\s.+aws_s3_account.+"
  labels = ["service/s3control"]
}

behavior "regexp_issue_labeler" "aws_sagemaker_service_labels" {
  regexp = "\*\saws_sagemaker.+\n"
  labels = ["service/sagemaker"]
}

behavior "regexp_issue_labeler" "aws_sagemaker_resource_labels" {
  regexp = "resource\s.+aws_sagemaker.+"
  labels = ["service/sagemaker"]
}

behavior "regexp_issue_labeler" "aws_sagemaker_data_labels" {
  regexp = "data\s.+aws_sagemaker.+"
  labels = ["service/sagemaker"]
}

behavior "regexp_issue_labeler" "aws_secretsmanager_service_labels" {
  regexp = "\*\saws_secretsmanager.+\n"
  labels = ["service/secretsmanager"]
}

behavior "regexp_issue_labeler" "aws_secretsmanager_resource_labels" {
  regexp = "resource\s.+aws_secretsmanager.+"
  labels = ["service/secretsmanager"]
}

behavior "regexp_issue_labeler" "aws_secretsmanager_data_labels" {
  regexp = "data\s.+aws_secretsmanager.+"
  labels = ["service/secretsmanager"]
}

behavior "regexp_issue_labeler" "aws_securityhub_service_labels" {
  regexp = "\*\saws_securityhub.+\n"
  labels = ["service/securityhub"]
}

behavior "regexp_issue_labeler" "aws_securityhub_resource_labels" {
  regexp = "resource\s.+aws_securityhub.+"
  labels = ["service/securityhub"]
}

behavior "regexp_issue_labeler" "aws_securityhub_data_labels" {
  regexp = "data\s.+aws_securityhub.+"
  labels = ["service/securityhub"]
}

behavior "regexp_issue_labeler" "aws_servicediscovery_service_labels" {
  regexp = "\*\saws_servicediscovery.+\n"
  labels = ["service/servicediscovery"]
}

behavior "regexp_issue_labeler" "aws_servicediscovery_resource_labels" {
  regexp = "resource\s.+aws_servicediscovery.+"
  labels = ["service/servicediscovery"]
}

behavior "regexp_issue_labeler" "aws_servicediscovery_data_labels" {
  regexp = "data\s.+aws_servicediscovery.+"
  labels = ["service/servicediscovery"]
}

behavior "regexp_issue_labeler" "aws_servicequotas_service_labels" {
  regexp = "\*\saws_servicequotas.+\n"
  labels = ["service/servicequotas"]
}

behavior "regexp_issue_labeler" "aws_servicequotas_resource_labels" {
  regexp = "resource\s.+aws_servicequotas.+"
  labels = ["service/servicequotas"]
}

behavior "regexp_issue_labeler" "aws_servicequotas_data_labels" {
  regexp = "data\s.+aws_servicequotas.+"
  labels = ["service/servicequotas"]
}

behavior "regexp_issue_labeler" "aws_ses_service_labels" {
  regexp = "\*\saws_ses.+\n"
  labels = ["service/ses"]
}

behavior "regexp_issue_labeler" "aws_ses_resource_labels" {
  regexp = "resource\s.+aws_ses.+"
  labels = ["service/ses"]
}

behavior "regexp_issue_labeler" "aws_ses_data_labels" {
  regexp = "data\s.+aws_ses.+"
  labels = ["service/ses"]
}

behavior "regexp_issue_labeler" "aws_sfn_service_labels" {
  regexp = "\*\saws_sfn.+\n"
  labels = ["service/sfn"]
}

behavior "regexp_issue_labeler" "aws_sfn_resource_labels" {
  regexp = "resource\s.+aws_sfn.+"
  labels = ["service/sfn"]
}

behavior "regexp_issue_labeler" "aws_sfn_data_labels" {
  regexp = "data\s.+aws_sfn.+"
  labels = ["service/sfn"]
}

behavior "regexp_issue_labeler" "aws_shield_service_labels" {
  regexp = "\*\saws_shield.+\n"
  labels = ["service/shield"]
}

behavior "regexp_issue_labeler" "aws_shield_resource_labels" {
  regexp = "resource\s.+aws_shield.+"
  labels = ["service/shield"]
}

behavior "regexp_issue_labeler" "aws_shield_data_labels" {
  regexp = "data\s.+aws_shield.+"
  labels = ["service/shield"]
}

behavior "regexp_issue_labeler" "aws_simpledb_service_labels" {
  regexp = "\*\saws_simpledb.+\n"
  labels = ["service/simpledb"]
}

behavior "regexp_issue_labeler" "aws_simpledb_resource_labels" {
  regexp = "resource\s.+aws_simpledb.+"
  labels = ["service/simpledb"]
}

behavior "regexp_issue_labeler" "aws_simpledb_data_labels" {
  regexp = "data\s.+aws_simpledb.+"
  labels = ["service/simpledb"]
}

behavior "regexp_issue_labeler" "aws_snowball_service_labels" {
  regexp = "\*\saws_snowball.+\n"
  labels = ["service/snowball"]
}

behavior "regexp_issue_labeler" "aws_snowball_resource_labels" {
  regexp = "resource\s.+aws_snowball.+"
  labels = ["service/snowball"]
}

behavior "regexp_issue_labeler" "aws_snowball_data_labels" {
  regexp = "data\s.+aws_snowball.+"
  labels = ["service/snowball"]
}

behavior "regexp_issue_labeler" "aws_sns_service_labels" {
  regexp = "\*\saws_sns.+\n"
  labels = ["service/sns"]
}

behavior "regexp_issue_labeler" "aws_sns_resource_labels" {
  regexp = "resource\s.+aws_sns.+"
  labels = ["service/sns"]
}

behavior "regexp_issue_labeler" "aws_sns_data_labels" {
  regexp = "data\s.+aws_sns.+"
  labels = ["service/sns"]
}

behavior "regexp_issue_labeler" "aws_sqs_service_labels" {
  regexp = "\*\saws_sqs.+\n"
  labels = ["service/sqs"]
}

behavior "regexp_issue_labeler" "aws_sqs_resource_labels" {
  regexp = "resource\s.+aws_sqs.+"
  labels = ["service/sqs"]
}

behavior "regexp_issue_labeler" "aws_sqs_data_labels" {
  regexp = "data\s.+aws_sqs.+"
  labels = ["service/sqs"]
}

behavior "regexp_issue_labeler" "aws_ssm_service_labels" {
  regexp = "\*\saws_ssm.+\n"
  labels = ["service/ssm"]
}

behavior "regexp_issue_labeler" "aws_ssm_resource_labels" {
  regexp = "resource\s.+aws_ssm.+"
  labels = ["service/ssm"]
}

behavior "regexp_issue_labeler" "aws_ssm_data_labels" {
  regexp = "data\s.+aws_ssm.+"
  labels = ["service/ssm"]
}

behavior "regexp_issue_labeler" "aws_storagegateway_service_labels" {
  regexp = "\*\saws_storagegateway.+\n"
  labels = ["service/storagegateway"]
}

behavior "regexp_issue_labeler" "aws_storagegateway_resource_labels" {
  regexp = "resource\s.+aws_storagegateway.+"
  labels = ["service/storagegateway"]
}

behavior "regexp_issue_labeler" "aws_storagegateway_data_labels" {
  regexp = "data\s.+aws_storagegateway.+"
  labels = ["service/storagegateway"]
}

behavior "regexp_issue_labeler" "aws_caller_identity_service_labels" {
  regexp = "\*\saws_caller_identity.+\n"
  labels = ["service/sts"]
}

behavior "regexp_issue_labeler" "aws_caller_identity_resource_labels" {
  regexp = "resource\s.+aws_caller_identity.+"
  labels = ["service/sts"]
}

behavior "regexp_issue_labeler" "aws_caller_identity_data_labels" {
  regexp = "data\s.+aws_caller_identity.+"
  labels = ["service/sts"]
}

behavior "regexp_issue_labeler" "aws_swf_service_labels" {
  regexp = "\*\saws_swf.+\n"
  labels = ["service/swf"]
}

behavior "regexp_issue_labeler" "aws_swf_resource_labels" {
  regexp = "resource\s.+aws_swf.+"
  labels = ["service/swf"]
}

behavior "regexp_issue_labeler" "aws_swf_data_labels" {
  regexp = "data\s.+aws_swf.+"
  labels = ["service/swf"]
}

behavior "regexp_issue_labeler" "aws_transfer_service_labels" {
  regexp = "\*\saws_transfer.+\n"
  labels = ["service/transfer"]
}

behavior "regexp_issue_labeler" "aws_transfer_resource_labels" {
  regexp = "resource\s.+aws_transfer.+"
  labels = ["service/transfer"]
}

behavior "regexp_issue_labeler" "aws_transfer_data_labels" {
  regexp = "data\s.+aws_transfer.+"
  labels = ["service/transfer"]
}

behavior "regexp_issue_labeler" "aws_waf_service_labels" {
  regexp = "\*\saws_waf.+\n"
  labels = ["service/waf"]
}

behavior "regexp_issue_labeler" "aws_waf_resource_labels" {
  regexp = "resource\s.+aws_waf.+"
  labels = ["service/waf"]
}

behavior "regexp_issue_labeler" "aws_waf_data_labels" {
  regexp = "data\s.+aws_waf.+"
  labels = ["service/waf"]
}

behavior "regexp_issue_labeler" "aws_workdocs_service_labels" {
  regexp = "\*\saws_workdocs.+\n"
  labels = ["service/workdocs"]
}

behavior "regexp_issue_labeler" "aws_workdocs_resource_labels" {
  regexp = "resource\s.+aws_workdocs.+"
  labels = ["service/workdocs"]
}

behavior "regexp_issue_labeler" "aws_workdocs_data_labels" {
  regexp = "data\s.+aws_workdocs.+"
  labels = ["service/workdocs"]
}

behavior "regexp_issue_labeler" "aws_workmail_service_labels" {
  regexp = "\*\saws_workmail.+\n"
  labels = ["service/workmail"]
}

behavior "regexp_issue_labeler" "aws_workmail_resource_labels" {
  regexp = "resource\s.+aws_workmail.+"
  labels = ["service/workmail"]
}

behavior "regexp_issue_labeler" "aws_workmail_data_labels" {
  regexp = "data\s.+aws_workmail.+"
  labels = ["service/workmail"]
}

behavior "regexp_issue_labeler" "aws_xray_service_labels" {
  regexp = "\*\saws_xray.+\n"
  labels = ["service/xray"]
}

behavior "regexp_issue_labeler" "aws_xray_resource_labels" {
  regexp = "resource\s.+aws_xray.+"
  labels = ["service/xray"]
}

behavior "regexp_issue_labeler" "aws_xray_data_labels" {
  regexp = "data\s.+aws_xray.+"
  labels = ["service/xray"]
}
