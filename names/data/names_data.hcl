// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

service "accessanalyzer" {

  sdk {
    id             = "AccessAnalyzer"
    client_version = [2]
  }

  names {
    provider_name_upper = "AccessAnalyzer"
    human_friendly      = "IAM Access Analyzer"
  }

  endpoint_info {
    endpoint_api_call        = "ListAnalyzers"
  }

  resource_prefix {
    correct = "aws_accessanalyzer_"
  }

  provider_package_correct = "accessanalyzer"
  doc_prefix               = ["accessanalyzer_"]
  brand                    = "AWS"
}

service "account" {

  sdk {
    id             = "Account"
    client_version = [2]
  }

  names {
    provider_name_upper = "Account"
    human_friendly      = "Account Management"
  }

  endpoint_info {
    endpoint_api_call        = "ListRegions"
  }

  resource_prefix {
    correct = "aws_account_"
  }

  provider_package_correct = "account"
  doc_prefix               = ["account_"]
  brand                    = "AWS"
}

service "acm" {

  sdk {
    id             = "ACM"
    client_version = [2]
  }

  names {
    provider_name_upper = "ACM"
    human_friendly      = "ACM (Certificate Manager)"
  }

  endpoint_info {
    endpoint_api_call        = "ListCertificates"
  }

  resource_prefix {
    correct = "aws_acm_"
  }

  provider_package_correct = "acm"
  doc_prefix               = ["acm_"]
  brand                    = "AWS"
}

service "acmpca" {

  cli_v2_command {
    aws_cli_v2_command           = "acm-pca"
    aws_cli_v2_command_no_dashes = "acmpca"
  }

  sdk {
    id             = "ACM PCA"
    client_version = [2]
  }

  names {
    provider_name_upper = "ACMPCA"
    human_friendly      = "ACM PCA (Certificate Manager Private Certificate Authority)"
  }

  endpoint_info {
    endpoint_api_call        = "ListCertificateAuthorities"
  }

  resource_prefix {
    correct = "aws_acmpca_"
  }

  provider_package_correct = "acmpca"
  doc_prefix               = ["acmpca_"]
  brand                    = "AWS"
}

service "alexaforbusiness" {

  sdk {
    id             = "Alexa For Business"
    client_version = [1]
  }

  names {
    provider_name_upper = "AlexaForBusiness"
    human_friendly      = "Alexa for Business"
  }

  client {
    go_v1_client_typename = "AlexaForBusiness"
  }

  resource_prefix {
    correct = "aws_alexaforbusiness_"
  }

  provider_package_correct = "alexaforbusiness"
  doc_prefix               = ["alexaforbusiness_"]
  brand                    = ""
  not_implemented          = true
}

service "amp" {

  go_packages {
    v1_package = "prometheusservice"
    v2_package = "amp"
  }

  sdk {
    id             = "amp"
    client_version = [2]
  }

  names {
    aliases             = ["prometheus", "prometheusservice"]
    provider_name_upper = "AMP"
    human_friendly      = "AMP (Managed Prometheus)"
  }

  endpoint_info {
    endpoint_api_call        = "ListScrapers"
  }

  resource_prefix {
    actual  = "aws_prometheus_"
    correct = "aws_amp_"
  }

  provider_package_correct = "amp"
  doc_prefix               = ["prometheus_"]
  brand                    = "AWS"
}

service "amplify" {

  sdk {
    id             = "Amplify"
    client_version = [2]
  }

  names {
    provider_name_upper = "Amplify"
    human_friendly      = "Amplify"
  }

  endpoint_info {
    endpoint_api_call        = "ListApps"
  }

  resource_prefix {
    correct = "aws_amplify_"
  }

  provider_package_correct = "amplify"
  doc_prefix               = ["amplify_"]
  brand                    = "AWS"
}

service "amplifybackend" {

  sdk {
    id             = "AmplifyBackend"
    client_version = [1]
  }

  names {
    provider_name_upper = "AmplifyBackend"
    human_friendly      = "Amplify Backend"
  }

  client {
    go_v1_client_typename = "AmplifyBackend"
  }

  resource_prefix {
    correct = "aws_amplifybackend_"
  }

  provider_package_correct = "amplifybackend"
  doc_prefix               = ["amplifybackend_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "amplifyuibuilder" {

  sdk {
    id             = "AmplifyUIBuilder"
    client_version = [1]
  }

  names {
    provider_name_upper = "AmplifyUIBuilder"
    human_friendly      = "Amplify UI Builder"
  }

  client {
    go_v1_client_typename = "AmplifyUIBuilder"
  }

  resource_prefix {
    correct = "aws_amplifyuibuilder_"
  }

  provider_package_correct = "amplifyuibuilder"
  doc_prefix               = ["amplifyuibuilder_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "apigateway" {

  sdk {
    id             = "API Gateway"
    client_version = [2]
  }

  names {
    provider_name_upper = "APIGateway"
    human_friendly      = "API Gateway"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "GetAccount"
  }

  resource_prefix {
    actual  = "aws_api_gateway_"
    correct = "aws_apigateway_"
  }

  provider_package_correct = "apigateway"
  doc_prefix               = ["api_gateway_"]
  brand                    = "AWS"
}

service "apigatewaymanagementapi" {

  sdk {
    id             = "ApiGatewayManagementApi"
    client_version = [1]
  }

  names {
    provider_name_upper = "APIGatewayManagementAPI"
    human_friendly      = "API Gateway Management API"
  }

  client {
    go_v1_client_typename = "ApiGatewayManagementApi"
  }

  resource_prefix {
    correct = "aws_apigatewaymanagementapi_"
  }

  provider_package_correct = "apigatewaymanagementapi"
  doc_prefix               = ["apigatewaymanagementapi_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "apigatewayv2" {

  sdk {
    id             = "ApiGatewayV2"
    client_version = [2]
  }

  names {
    provider_name_upper = "APIGatewayV2"
    human_friendly      = "API Gateway V2"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "GetApis"
  }

  resource_prefix {
    correct = "aws_apigatewayv2_"
  }

  provider_package_correct = "apigatewayv2"
  doc_prefix               = ["apigatewayv2_"]
  brand                    = "AWS"
}

service "appfabric" {

  sdk {
    id             = "AppFabric"
    client_version = [2]
  }

  names {
    provider_name_upper = "AppFabric"
    human_friendly      = "AppFabric"
  }

  endpoint_info {
    endpoint_api_call        = "ListAppBundles"
  }

  resource_prefix {
    correct = "aws_appfabric_"
  }

  provider_package_correct = "appfabric"
  doc_prefix               = ["appfabric_"]
  brand                    = "AWS"
}

service "appmesh" {

  sdk {
    id             = "App Mesh"
    client_version = [1]
  }

  names {
    provider_name_upper = "AppMesh"
    human_friendly      = "App Mesh"
  }

  client {
    go_v1_client_typename = "AppMesh"
  }

  endpoint_info {
    endpoint_api_call        = "ListMeshes"
  }

  resource_prefix {
    correct = "aws_appmesh_"
  }

  provider_package_correct = "appmesh"
  doc_prefix               = ["appmesh_"]
  brand                    = "AWS"
}

service "apprunner" {

  sdk {
    id             = "AppRunner"
    client_version = [2]
  }

  names {
    provider_name_upper = "AppRunner"
    human_friendly      = "App Runner"
  }

  endpoint_info {
    endpoint_api_call        = "ListConnections"
  }

  resource_prefix {
    correct = "aws_apprunner_"
  }

  provider_package_correct = "apprunner"
  doc_prefix               = ["apprunner_"]
  brand                    = "AWS"
}

service "appconfig" {

  sdk {
    id             = "AppConfig"
    client_version = [2]
  }

  names {
    provider_name_upper = "AppConfig"
    human_friendly      = "AppConfig"
  }

  client {
    go_v1_client_typename = "AppConfig"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    correct = "aws_appconfig_"
  }

  provider_package_correct = "appconfig"
  doc_prefix               = ["appconfig_"]
  brand                    = "AWS"
}

service "appconfigdata" {

  sdk {
    id             = "AppConfigData"
    client_version = [1]
  }

  names {
    provider_name_upper = "AppConfigData"
    human_friendly      = "AppConfig Data"
  }

  client {
    go_v1_client_typename = "AppConfigData"
  }

  resource_prefix {
    correct = "aws_appconfigdata_"
  }

  provider_package_correct = "appconfigdata"
  doc_prefix               = ["appconfigdata_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "appflow" {

  sdk {
    id             = "Appflow"
    client_version = [2]
  }

  names {
    provider_name_upper = "AppFlow"
    human_friendly      = "AppFlow"
  }

  endpoint_info {
    endpoint_api_call        = "ListFlows"
  }

  resource_prefix {
    correct = "aws_appflow_"
  }

  provider_package_correct = "appflow"
  doc_prefix               = ["appflow_"]
  brand                    = "AWS"
}

service "appintegrations" {

  go_packages {
    v1_package = "appintegrationsservice"
    v2_package = "appintegrations"
  }

  sdk {
    id             = "AppIntegrations"
    client_version = [2]
  }

  names {
    aliases             = ["appintegrationsservice"]
    provider_name_upper = "AppIntegrations"
    human_friendly      = "AppIntegrations"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    correct = "aws_appintegrations_"
  }

  provider_package_correct = "appintegrations"
  doc_prefix               = ["appintegrations_"]
  brand                    = "AWS"
}

service "appautoscaling" {

  cli_v2_command {
    aws_cli_v2_command           = "application-autoscaling"
    aws_cli_v2_command_no_dashes = "applicationautoscaling"
  }

  go_packages {
    v1_package = "applicationautoscaling"
    v2_package = "applicationautoscaling"
  }

  sdk {
    id             = "Application Auto Scaling"
    client_version = [2]
  }

  names {
    aliases             = ["applicationautoscaling"]
    provider_name_upper = "AppAutoScaling"
    human_friendly      = "Application Auto Scaling"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeScalableTargets"
    endpoint_api_params      = "ServiceNamespace: awstypes.ServiceNamespaceEcs"
  }

  resource_prefix {
    actual  = "aws_appautoscaling_"
    correct = "aws_applicationautoscaling_"
  }

  provider_package_correct = "applicationautoscaling"
  doc_prefix               = ["appautoscaling_"]
  brand                    = ""
}
service "applicationcostprofiler" {

  sdk {
    id             = "ApplicationCostProfiler"
    client_version = [1]
  }

  names {
    provider_name_upper = "ApplicationCostProfiler"
    human_friendly      = "Application Cost Profiler"
  }

  client {
    go_v1_client_typename = "ApplicationCostProfiler"
  }

  resource_prefix {
    correct = "aws_applicationcostprofiler_"
  }

  provider_package_correct = "applicationcostprofiler"
  doc_prefix               = ["applicationcostprofiler_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "applicationsignals" {

  cli_v2_command {
    aws_cli_v2_command           = "application-signals"
    aws_cli_v2_command_no_dashes = "applicationsignals"
  }

  sdk {
    id             = "Application Signals"
    client_version = [2]
  }

  names {
    provider_name_upper = "ApplicationSignals"
    human_friendly      = "Application Signals"
  }

  endpoint_info {
    endpoint_api_call        = "ListServiceLevelObjectives"
  }

  resource_prefix {
    correct = "aws_applicationsignals_"
  }

  provider_package_correct = "applicationsignals"
  doc_prefix               = ["applicationsignals_"]
  brand                    = "Amazon"
}

service "discovery" {

  go_packages {
    v1_package = "applicationdiscoveryservice"
    v2_package = "applicationdiscoveryservice"
  }

  sdk {
    id             = "Application Discovery Service"
    client_version = [1]
  }

  names {
    aliases             = ["applicationdiscovery", "applicationdiscoveryservice"]
    provider_name_upper = "Discovery"
    human_friendly      = "Application Discovery"
  }

  client {
    go_v1_client_typename = "ApplicationDiscoveryService"
  }

  resource_prefix {
    correct = "aws_discovery_"
  }

  provider_package_correct = "discovery"
  doc_prefix               = ["discovery_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "mgn" {

  sdk {
    id             = "mgn"
    client_version = [1]
  }

  names {
    provider_name_upper = "Mgn"
    human_friendly      = "Application Migration (Mgn)"
  }

  client {
    go_v1_client_typename = "Mgn"
  }

  resource_prefix {
    correct = "aws_mgn_"
  }

  provider_package_correct = "mgn"
  doc_prefix               = ["mgn_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "appstream" {

  sdk {
    id             = "AppStream"
    client_version = [2]
  }

  names {
    provider_name_upper = "AppStream"
    human_friendly      = "AppStream 2.0"
  }

  endpoint_info {
    endpoint_api_call        = "ListAssociatedFleets"
    endpoint_api_params      = "StackName: aws_sdkv2.String(\"test\")"
  }

  resource_prefix {
    correct = "aws_appstream_"
  }

  provider_package_correct = "appstream"
  doc_prefix               = ["appstream_"]
  brand                    = "AWS"
}

service "appsync" {

  sdk {
    id             = "AppSync"
    client_version = [2]
  }

  names {
    provider_name_upper = "AppSync"
    human_friendly      = "AppSync"
  }

  client {
    go_v1_client_typename = "AppSync"
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListDomainNames"
  }

  resource_prefix {
    correct = "aws_appsync_"
  }

  provider_package_correct = "appsync"
  doc_prefix               = ["appsync_"]
  brand                    = "AWS"
}

service "athena" {

  sdk {
    id             = "Athena"
    client_version = [2]
  }

  names {
    provider_name_upper = "Athena"
    human_friendly      = "Athena"
  }

  endpoint_info {
    endpoint_api_call        = "ListDataCatalogs"
  }

  resource_prefix {
    correct = "aws_athena_"
  }

  provider_package_correct = "athena"
  doc_prefix               = ["athena_"]
  brand                    = "AWS"
}

service "auditmanager" {

  sdk {
    id             = "AuditManager"
    client_version = [2]
  }

  names {
    provider_name_upper = "AuditManager"
    human_friendly      = "Audit Manager"
  }

  endpoint_info {
    endpoint_api_call        = "GetAccountStatus"
  }

  resource_prefix {
    correct = "aws_auditmanager_"
  }

  provider_package_correct = "auditmanager"
  doc_prefix               = ["auditmanager_"]
  brand                    = "AWS"
}

service "autoscaling" {

  sdk {
    id             = "Auto Scaling"
    client_version = [2]
  }

  names {
    provider_name_upper = "AutoScaling"
    human_friendly      = "Auto Scaling"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeAutoScalingGroups"
  }

  resource_prefix {
    actual  = "aws_(autoscaling_|launch_configuration)"
    correct = "aws_autoscaling_"
  }

  provider_package_correct = "autoscaling"
  doc_prefix               = ["autoscaling_", "launch_configuration"]
  brand                    = ""
}
service "autoscalingplans" {

  cli_v2_command {
    aws_cli_v2_command           = "autoscaling-plans"
    aws_cli_v2_command_no_dashes = "autoscalingplans"
  }

  sdk {
    id             = "Auto Scaling Plans"
    client_version = [2]
  }

  names {
    provider_name_upper = "AutoScalingPlans"
    human_friendly      = "Auto Scaling Plans"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeScalingPlans"
  }

  resource_prefix {
    correct = "aws_autoscalingplans_"
  }

  provider_package_correct = "autoscalingplans"
  doc_prefix               = ["autoscalingplans_"]
  brand                    = ""
}
service "backup" {

  sdk {
    id             = "Backup"
    client_version = [2]
  }

  names {
    provider_name_upper = "Backup"
    human_friendly      = "Backup"
  }

  client {
    go_v1_client_typename = "Backup"
  }

  endpoint_info {
    endpoint_api_call        = "ListBackupPlans"
  }

  resource_prefix {
    correct = "aws_backup_"
  }

  provider_package_correct = "backup"
  doc_prefix               = ["backup_"]
  brand                    = "AWS"
}

service "backupgateway" {

  cli_v2_command {
    aws_cli_v2_command           = "backup-gateway"
    aws_cli_v2_command_no_dashes = "backupgateway"
  }

  sdk {
    id             = "Backup Gateway"
    client_version = [1]
  }

  names {
    provider_name_upper = "BackupGateway"
    human_friendly      = "Backup Gateway"
  }

  client {
    go_v1_client_typename = "BackupGateway"
  }

  resource_prefix {
    correct = "aws_backupgateway_"
  }

  provider_package_correct = "backupgateway"
  doc_prefix               = ["backupgateway_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "batch" {

  sdk {
    id             = "Batch"
    client_version = [1, 2]
  }

  names {
    provider_name_upper = "Batch"
    human_friendly      = "Batch"
  }

  client {
    go_v1_client_typename = "Batch"
  }

  endpoint_info {
    endpoint_api_call        = "ListJobs"
  }

  resource_prefix {
    correct = "aws_batch_"
  }

  provider_package_correct = "batch"
  doc_prefix               = ["batch_"]
  brand                    = "AWS"
}

service "bedrock" {
  sdk {
    id             = "Bedrock"
    client_version = [2]
  }

  names {
    provider_name_upper = "Bedrock"
    human_friendly      = "Bedrock"
  }

  endpoint_info {
    endpoint_api_call = "ListFoundationModels"
  }

  resource_prefix {
    correct = "aws_bedrock_"
  }

  provider_package_correct = "bedrock"
  doc_prefix               = ["bedrock_"]
  brand                    = "Amazon"
}

service "bedrockagent" {
  cli_v2_command {
    aws_cli_v2_command           = "bedrock-agent"
    aws_cli_v2_command_no_dashes = "bedrockagent"
  }

  sdk {
    id             = "Bedrock Agent"
    client_version = [2]
  }

  names {
    provider_name_upper = "BedrockAgent"
    human_friendly      = "Bedrock Agents"
  }

  endpoint_info {
    endpoint_api_call = "ListAgents"
  }

  resource_prefix {
    correct = "aws_bedrockagent_"
  }

  provider_package_correct = "bedrockagent"
  doc_prefix               = ["bedrockagent_"]
  brand                    = "Amazon"
}

service "bcmdataexports" {

  sdk {
    id             = "BCM Data Exports"
    client_version = [2]
  }

  names {
    provider_name_upper = "BCMDataExports"
    human_friendly      = "BCM Data Exports"
  }

  endpoint_info {
    endpoint_api_call        = "ListExports"
  }

  resource_prefix {
    correct = "aws_bcmdataexports_"
  }

  provider_package_correct = "bcmdataexports"
  doc_prefix               = ["bcmdataexports_"]
  brand                    = "AWS"
}

service "billingconductor" {

  go_packages {
    v1_package = "billingconductor"
    v2_package = ""
  }

  sdk {
    id             = "billingconductor"
    client_version = [1]
  }

  names {
    provider_name_upper = "BillingConductor"
    human_friendly      = "Billing Conductor"
  }

  client {
    go_v1_client_typename = "BillingConductor"
  }

  resource_prefix {
    correct = "aws_billingconductor_"
  }

  provider_package_correct = "billingconductor"
  doc_prefix               = ["billingconductor_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "braket" {

  sdk {
    id             = "Braket"
    client_version = [1]
  }

  names {
    provider_name_upper = "Braket"
    human_friendly      = "Braket"
  }

  client {
    go_v1_client_typename = "Braket"
  }

  resource_prefix {
    correct = "aws_braket_"
  }

  provider_package_correct = "braket"
  doc_prefix               = ["braket_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "ce" {

  go_packages {
    v1_package = "costexplorer"
    v2_package = "costexplorer"
  }

  sdk {
    id             = "Cost Explorer"
    client_version = [2]
  }

  names {
    aliases             = ["costexplorer"]
    provider_name_upper = "CE"
    human_friendly      = "CE (Cost Explorer)"
  }

  endpoint_info {
    endpoint_api_call        = "ListCostCategoryDefinitions"
  }

  resource_prefix {
    correct = "aws_ce_"
  }

  provider_package_correct = "ce"
  doc_prefix               = ["ce_"]
  brand                    = "AWS"
}

service "chatbot" {

  sdk {
    id             = "Chatbot"
    client_version = [2]
  }

  names {
    provider_name_upper = "Chatbot"
    human_friendly      = "Chatbot"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "GetAccountPreferences"
  }

  resource_prefix {
    correct = "aws_chatbot_"
  }

  provider_package_correct = "chatbot"
  doc_prefix               = ["chatbot_"]
  brand                    = "AWS"
}

service "chime" {

  sdk {
    id             = "Chime"
    client_version = [2]
  }

  names {
    provider_name_upper = "Chime"
    human_friendly      = "Chime"
  }

  client {
    go_v1_client_typename = "Chime"
  }

  endpoint_info {
    endpoint_api_call        = "ListAccounts"
  }

  resource_prefix {
    correct = "aws_chime_"
  }

  provider_package_correct = "chime"
  doc_prefix               = ["chime_"]
  brand                    = "AWS"
}

service "chimesdkidentity" {

  cli_v2_command {
    aws_cli_v2_command           = "chime-sdk-identity"
    aws_cli_v2_command_no_dashes = "chimesdkidentity"
  }

  sdk {
    id             = "Chime SDK Identity"
    client_version = [1]
  }

  names {
    provider_name_upper = "ChimeSDKIdentity"
    human_friendly      = "Chime SDK Identity"
  }

  client {
    go_v1_client_typename = "ChimeSDKIdentity"
  }

  resource_prefix {
    correct = "aws_chimesdkidentity_"
  }

  provider_package_correct = "chimesdkidentity"
  doc_prefix               = ["chimesdkidentity_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "chimesdkmediapipelines" {

  cli_v2_command {
    aws_cli_v2_command           = "chime-sdk-mediapipelines"
    aws_cli_v2_command_no_dashes = "chimesdkmediapipelines"
  }

  sdk {
    id             = "Chime SDK Media Pipelines"
    client_version = [2]
  }

  names {
    provider_name_upper = "ChimeSDKMediaPipelines"
    human_friendly      = "Chime SDK Media Pipelines"
  }

  endpoint_info {
    endpoint_api_call        = "ListMediaPipelines"
  }

  resource_prefix {
    correct = "aws_chimesdkmediapipelines_"
  }

  provider_package_correct = "chimesdkmediapipelines"
  doc_prefix               = ["chimesdkmediapipelines_"]
  brand                    = "AWS"
}

service "chimesdkmeetings" {

  cli_v2_command {
    aws_cli_v2_command           = "chime-sdk-meetings"
    aws_cli_v2_command_no_dashes = "chimesdkmeetings"
  }

  sdk {
    id             = "Chime SDK Meetings"
    client_version = [1]
  }

  names {
    provider_name_upper = "ChimeSDKMeetings"
    human_friendly      = "Chime SDK Meetings"
  }

  client {
    go_v1_client_typename = "ChimeSDKMeetings"
  }

  resource_prefix {
    correct = "aws_chimesdkmeetings_"
  }

  provider_package_correct = "chimesdkmeetings"
  doc_prefix               = ["chimesdkmeetings_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "chimesdkmessaging" {

  cli_v2_command {
    aws_cli_v2_command           = "chime-sdk-messaging"
    aws_cli_v2_command_no_dashes = "chimesdkmessaging"
  }

  sdk {
    id             = "Chime SDK Messaging"
    client_version = [1]
  }

  names {
    provider_name_upper = "ChimeSDKMessaging"
    human_friendly      = "Chime SDK Messaging"
  }

  client {
    go_v1_client_typename = "ChimeSDKMessaging"
  }

  resource_prefix {
    correct = "aws_chimesdkmessaging_"
  }

  provider_package_correct = "chimesdkmessaging"
  doc_prefix               = ["chimesdkmessaging_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "chimesdkvoice" {

  cli_v2_command {
    aws_cli_v2_command           = "chime-sdk-voice"
    aws_cli_v2_command_no_dashes = "chimesdkvoice"
  }

  sdk {
    id             = "Chime SDK Voice"
    client_version = [2]
  }

  names {
    provider_name_upper = "ChimeSDKVoice"
    human_friendly      = "Chime SDK Voice"
  }

  endpoint_info {
    endpoint_api_call        = "ListPhoneNumbers"
  }

  resource_prefix {
    correct = "aws_chimesdkvoice_"
  }

  provider_package_correct = "chimesdkvoice"
  doc_prefix               = ["chimesdkvoice_"]
  brand                    = "AWS"
}

service "cleanrooms" {

  sdk {
    id             = "CleanRooms"
    client_version = [2]
  }

  names {
    provider_name_upper = "CleanRooms"
    human_friendly      = "Clean Rooms"
  }

  endpoint_info {
    endpoint_api_call        = "ListCollaborations"
  }

  resource_prefix {
    correct = "aws_cleanrooms_"
  }

  provider_package_correct = "cleanrooms"
  doc_prefix               = ["cleanrooms_"]
  brand                    = "AWS"
}

service "cloudcontrol" {

  go_packages {
    v1_package = "cloudcontrolapi"
    v2_package = "cloudcontrol"
  }

  sdk {
    id             = "CloudControl"
    client_version = [2]
  }

  names {
    aliases             = ["cloudcontrolapi"]
    provider_name_upper = "CloudControl"
    human_friendly      = "Cloud Control API"
  }

    client {
    go_v1_client_typename = "CloudControlApi"
  }

  endpoint_info {
    endpoint_api_call        = "ListResourceRequests"
  }

  resource_prefix {
    actual  = "aws_cloudcontrolapi_"
    correct = "aws_cloudcontrol_"
  }

  provider_package_correct = "cloudcontrol"
  doc_prefix               = ["cloudcontrolapi_"]
  brand                    = "AWS"
}

service "clouddirectory" {

  sdk {
    id             = "CloudDirectory"
    client_version = [1]
  }

  names {
    provider_name_upper = "CloudDirectory"
    human_friendly      = "Cloud Directory"
  }

  client {
    go_v1_client_typename = "CloudDirectory"
  }

  resource_prefix {
    correct = "aws_clouddirectory_"
  }

  provider_package_correct = "clouddirectory"
  doc_prefix               = ["clouddirectory_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "servicediscovery" {

  sdk {
    id             = "ServiceDiscovery"
    client_version = [2]
  }

  names {
    provider_name_upper = "ServiceDiscovery"
    human_friendly      = "Cloud Map"
  }

  endpoint_info {
    endpoint_api_call        = "ListNamespaces"
  }

  resource_prefix {
    actual  = "aws_service_discovery_"
    correct = "aws_servicediscovery_"
  }

  provider_package_correct = "servicediscovery"
  doc_prefix               = ["service_discovery_"]
  brand                    = "AWS"
}

service "cloud9" {

  sdk {
    id             = "Cloud9"
    client_version = [2]
  }

  names {
    provider_name_upper = "Cloud9"
    human_friendly      = "Cloud9"
  }

  endpoint_info {
    endpoint_api_call        = "ListEnvironments"
  }

  resource_prefix {
    correct = "aws_cloud9_"
  }

  provider_package_correct = "cloud9"
  doc_prefix               = ["cloud9_"]
  brand                    = "AWS"
}

service "cloudformation" {

  sdk {
    id             = "CloudFormation"
    client_version = [2]
  }

  names {
    provider_name_upper = "CloudFormation"
    human_friendly      = "CloudFormation"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListStackInstances"
    endpoint_api_params      = "StackSetName: aws_sdkv2.String(\"test\")"
  }

  resource_prefix {
    correct = "aws_cloudformation_"
  }

  provider_package_correct = "cloudformation"
  doc_prefix               = ["cloudformation_"]
  brand                    = "AWS"
}

service "cloudfront" {

  sdk {
    id             = "CloudFront"
    client_version = [2]
  }

  names {
    provider_name_upper = "CloudFront"
    human_friendly      = "CloudFront"
  }

  endpoint_info {
    endpoint_api_call        = "ListDistributions"
  }

  resource_prefix {
    correct = "aws_cloudfront_"
  }

  provider_package_correct = "cloudfront"
  doc_prefix               = ["cloudfront_"]
  brand                    = "AWS"
}

service "cloudfrontkeyvaluestore" {

  cli_v2_command {
    aws_cli_v2_command           = "cloudfront-keyvaluestore"
    aws_cli_v2_command_no_dashes = "cloudfrontkeyvaluestore"
  }

  go_packages {
    v1_package = ""
    v2_package = "cloudfrontkeyvaluestore"
  }

  sdk {
    id             = "CloudFront KeyValueStore"
    client_version = [2]
  }

  names {
    provider_name_upper = "CloudFrontKeyValueStore"
    human_friendly      = "CloudFront KeyValueStore"
  }

  endpoint_info {
    endpoint_api_call        = "ListKeys"
    endpoint_api_params      = "KvsARN: aws_sdkv2.String(\"arn:aws:cloudfront::111122223333:key-value-store/MaxAge\")"
  }

  resource_prefix {
    correct = "aws_cloudfrontkeyvaluestore_"
  }

  provider_package_correct = "cloudfrontkeyvaluestore"
  doc_prefix               = ["cloudfrontkeyvaluestore_"]
  brand                    = "AWS"
}

service "cloudhsmv2" {

  sdk {
    id             = "CloudHSM V2"
    client_version = [2]
  }

  names {
    aliases             = ["cloudhsm"]
    provider_name_upper = "CloudHSMV2"
    human_friendly      = "CloudHSM"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "DescribeClusters"
  }

  resource_prefix {
    actual  = "aws_cloudhsm_v2_"
    correct = "aws_cloudhsmv2_"
  }

  provider_package_correct = "cloudhsmv2"
  doc_prefix               = ["cloudhsm"]
  brand                    = "AWS"
}

service "cloudsearch" {

  sdk {
    id             = "CloudSearch"
    client_version = [2]
  }

  names {
    provider_name_upper = "CloudSearch"
    human_friendly      = "CloudSearch"
  }

  endpoint_info {
    endpoint_api_call        = "ListDomainNames"
  }

  resource_prefix {
    correct = "aws_cloudsearch_"
  }

  provider_package_correct = "cloudsearch"
  doc_prefix               = ["cloudsearch_"]
  brand                    = "AWS"
}

service "cloudsearchdomain" {

  sdk {
    id             = "CloudSearch Domain"
    client_version = [1]
  }

  names {
    provider_name_upper = "CloudSearchDomain"
    human_friendly      = "CloudSearch Domain"
  }

  client {
    go_v1_client_typename = "CloudSearchDomain"
  }

  resource_prefix {
    correct = "aws_cloudsearchdomain_"
  }

  provider_package_correct = "cloudsearchdomain"
  doc_prefix               = ["cloudsearchdomain_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "cloudtrail" {

  sdk {
    id             = "CloudTrail"
    client_version = [2]
  }

  names {
    provider_name_upper = "CloudTrail"
    human_friendly      = "CloudTrail"
  }

  endpoint_info {
    endpoint_api_call        = "ListChannels"
  }

  resource_prefix {
    actual  = "aws_cloudtrail"
    correct = "aws_cloudtrail_"
  }

  provider_package_correct = "cloudtrail"
  doc_prefix               = ["cloudtrail"]
  brand                    = "AWS"
}

service "cloudwatch" {

  sdk {
    id             = "CloudWatch"
    client_version = [2]
  }

  names {
    provider_name_upper = "CloudWatch"
    human_friendly      = "CloudWatch"
  }

  endpoint_info {
    endpoint_api_call        = "ListDashboards"
  }

  resource_prefix {
    actual  = "aws_cloudwatch_(?!(event_|log_|query_))"
    correct = "aws_cloudwatch_"
  }

  provider_package_correct = "cloudwatch"
  doc_prefix               = ["cloudwatch_dashboard", "cloudwatch_metric_", "cloudwatch_composite_"]
  brand                    = "AWS"
}

service "applicationinsights" {

  cli_v2_command {
    aws_cli_v2_command           = "application-insights"
    aws_cli_v2_command_no_dashes = "applicationinsights"
  }

  sdk {
    id             = "Application Insights"
    client_version = [2]
  }

  names {
    provider_name_upper = "ApplicationInsights"
    human_friendly      = "CloudWatch Application Insights"
  }

  endpoint_info {
    endpoint_api_call        = "CreateApplication"
  }

  resource_prefix {
    correct = "aws_applicationinsights_"
  }

  provider_package_correct = "applicationinsights"
  doc_prefix               = ["applicationinsights_"]
  brand                    = "AWS"
}

service "evidently" {

  go_packages {
    v1_package = "cloudwatchevidently"
    v2_package = "evidently"
  }

  sdk {
    id             = "Evidently"
    client_version = [2]
  }

  names {
    aliases             = ["cloudwatchevidently"]
    provider_name_upper = "Evidently"
    human_friendly      = "CloudWatch Evidently"
  }

  endpoint_info {
    endpoint_api_call        = "ListProjects"
  }

  resource_prefix {
    correct = "aws_evidently_"
  }

  provider_package_correct = "evidently"
  doc_prefix               = ["evidently_"]
  brand                    = "Amazon"
}

service "internetmonitor" {

  sdk {
    id             = "InternetMonitor"
    client_version = [2]
  }

  names {
    provider_name_upper = "InternetMonitor"
    human_friendly      = "CloudWatch Internet Monitor"
  }

  endpoint_info {
    endpoint_api_call        = "ListMonitors"
  }

  resource_prefix {
    correct = "aws_internetmonitor_"
  }

  provider_package_correct = "internetmonitor"
  doc_prefix               = ["internetmonitor_"]
  brand                    = "AWS"
}

service "logs" {

  go_packages {
    v1_package = "cloudwatchlogs"
    v2_package = "cloudwatchlogs"
  }

  sdk {
    id             = "CloudWatch Logs"
    client_version = [2]
  }

  names {
    aliases             = ["cloudwatchlog", "cloudwatchlogs"]
    provider_name_upper = "Logs"
    human_friendly      = "CloudWatch Logs"
  }

  client {
    go_v1_client_typename = "CloudWatchLogs"
  }

  endpoint_info {
    endpoint_api_call        = "ListAnomalies"
  }

  resource_prefix {
    actual  = "aws_cloudwatch_(log_|query_)"
    correct = "aws_logs_"
  }

  provider_package_correct = "logs"
  doc_prefix               = ["cloudwatch_log_", "cloudwatch_query_"]
  brand                    = "AWS"
}

service "networkmonitor" {

  sdk {
    id             = "NetworkMonitor"
    client_version = [2]
  }

  names {
    provider_name_upper = "NetworkMonitor"
    human_friendly      = "CloudWatch Network Monitor"
  }

  endpoint_info {
    endpoint_api_call        = "ListMonitors"
  }

  resource_prefix {
    correct = "aws_networkmonitor_"
  }

  provider_package_correct = "networkmonitor"
  doc_prefix               = ["networkmonitor_"]
  brand                    = "Amazon"
}

service "rum" {

  go_packages {
    v1_package = "cloudwatchrum"
    v2_package = "rum"
  }

  sdk {
    id             = "RUM"
    client_version = [2]
  }

  names {
    aliases             = ["cloudwatchrum"]
    provider_name_upper = "RUM"
    human_friendly      = "CloudWatch RUM"
  }

  client {
    go_v1_client_typename = "CloudWatchRUM"
  }

  endpoint_info {
    endpoint_api_call        = "ListAppMonitors"
  }

  resource_prefix {
    correct = "aws_rum_"
  }

  provider_package_correct = "rum"
  doc_prefix               = ["rum_"]
  brand                    = "AWS"
}

service "synthetics" {

  sdk {
    id             = "synthetics"
    client_version = [2]
  }

  names {
    provider_name_upper = "Synthetics"
    human_friendly      = "CloudWatch Synthetics"
  }

  endpoint_info {
    endpoint_api_call        = "ListGroups"
  }

  resource_prefix {
    correct = "aws_synthetics_"
  }

  provider_package_correct = "synthetics"
  doc_prefix               = ["synthetics_"]
  brand                    = "Amazon"
}

service "codeartifact" {

  sdk {
    id             = "codeartifact"
    client_version = [2]
  }

  names {
    provider_name_upper = "CodeArtifact"
    human_friendly      = "CodeArtifact"
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
  }

  resource_prefix {
    correct = "aws_codeartifact_"
  }

  provider_package_correct = "codeartifact"
  doc_prefix               = ["codeartifact_"]
  brand                    = "AWS"
}

service "codebuild" {

  sdk {
    id             = "CodeBuild"
    client_version = [2]
  }

  names {
    provider_name_upper = "CodeBuild"
    human_friendly      = "CodeBuild"
  }

  endpoint_info {
    endpoint_api_call        = "ListBuildBatches"
  }

  resource_prefix {
    correct = "aws_codebuild_"
  }

  provider_package_correct = "codebuild"
  doc_prefix               = ["codebuild_"]
  brand                    = "AWS"
}

service "codecommit" {

  sdk {
    id             = "CodeCommit"
    client_version = [2]
  }

  names {
    provider_name_upper = "CodeCommit"
    human_friendly      = "CodeCommit"
  }

  endpoint_info {
    endpoint_api_call        = "ListRepositories"
  }

  resource_prefix {
    correct = "aws_codecommit_"
  }

  provider_package_correct = "codecommit"
  doc_prefix               = ["codecommit_"]
  brand                    = "AWS"
}

service "deploy" {

  go_packages {
    v1_package = "codedeploy"
    v2_package = "codedeploy"
  }

  sdk {
    id             = "CodeDeploy"
    client_version = [2]
  }

  names {
    aliases             = ["codedeploy"]
    provider_name_upper = "Deploy"
    human_friendly      = "CodeDeploy"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  client {
  go_v1_client_typename = "CodeDeploy"
  }

  resource_prefix {
    actual  = "aws_codedeploy_"
    correct = "aws_deploy_"
  }

  provider_package_correct = "deploy"
  doc_prefix               = ["codedeploy_"]
  brand                    = "AWS"
}

service "codeguruprofiler" {

  sdk {
    id             = "CodeGuruProfiler"
    client_version = [2]
  }

  names {
    provider_name_upper = "CodeGuruProfiler"
    human_friendly      = "CodeGuru Profiler"
  }

  endpoint_info {
    endpoint_api_call        = "ListProfilingGroups"
  }

  resource_prefix {
    correct = "aws_codeguruprofiler_"
  }

  provider_package_correct = "codeguruprofiler"
  doc_prefix               = ["codeguruprofiler_"]
  brand                    = "AWS"
}

service "codegurureviewer" {

  cli_v2_command {
    aws_cli_v2_command           = "codeguru-reviewer"
    aws_cli_v2_command_no_dashes = "codegurureviewer"
  }

  sdk {
    id             = "CodeGuru Reviewer"
    client_version = [2]
  }

  names {
    provider_name_upper = "CodeGuruReviewer"
    human_friendly      = "CodeGuru Reviewer"
  }

  endpoint_info {
    endpoint_api_call        = "ListCodeReviews"
    endpoint_api_params      = "Type: awstypes.TypePullRequest"
  }

  resource_prefix {
    correct = "aws_codegurureviewer_"
  }

  provider_package_correct = "codegurureviewer"
  doc_prefix               = ["codegurureviewer_"]
  brand                    = "AWS"
}

service "codepipeline" {

  sdk {
    id             = "CodePipeline"
    client_version = [2]
  }

  names {
    provider_name_upper = "CodePipeline"
    human_friendly      = "CodePipeline"
  }

  endpoint_info {
    endpoint_api_call        = "ListPipelines"
  }

  resource_prefix {
    actual  = "aws_codepipeline"
    correct = "aws_codepipeline_"
  }

  provider_package_correct = "codepipeline"
  doc_prefix               = ["codepipeline"]
  brand                    = "AWS"
}

service "codestar" {

  sdk {
    id             = "CodeStar"
    client_version = [1]
  }

  names {
    provider_name_upper = "CodeStar"
    human_friendly      = "CodeStar"
  }

  client {
    go_v1_client_typename = "CodeStar"
  }

  resource_prefix {
    correct = "aws_codestar_"
  }

  provider_package_correct = "codestar"
  doc_prefix               = ["codestar_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "codestarconnections" {

  cli_v2_command {
    aws_cli_v2_command           = "codestar-connections"
    aws_cli_v2_command_no_dashes = "codestarconnections"
  }

  sdk {
    id             = "CodeStar connections"
    client_version = [2]
  }

  names {
    provider_name_upper = "CodeStarConnections"
    human_friendly      = "CodeStar Connections"
  }

  endpoint_info {
    endpoint_api_call        = "ListConnections"
  }

  resource_prefix {
    correct = "aws_codestarconnections_"
  }

  provider_package_correct = "codestarconnections"
  doc_prefix               = ["codestarconnections_"]
  brand                    = "AWS"
}

service "codestarnotifications" {

  cli_v2_command {
    aws_cli_v2_command           = "codestar-notifications"
    aws_cli_v2_command_no_dashes = "codestarnotifications"
  }

  sdk {
    id             = "codestar notifications"
    client_version = [2]
  }

  names {
    provider_name_upper = "CodeStarNotifications"
    human_friendly      = "CodeStar Notifications"
  }

  endpoint_info {
    endpoint_api_call        = "ListTargets"
  }

  resource_prefix {
    correct = "aws_codestarnotifications_"
  }

  provider_package_correct = "codestarnotifications"
  doc_prefix               = ["codestarnotifications_"]
  brand                    = "AWS"
}

service "cognitoidentity" {

  cli_v2_command {
    aws_cli_v2_command           = "cognito-identity"
    aws_cli_v2_command_no_dashes = "cognitoidentity"
  }

  sdk {
    id             = "Cognito Identity"
    client_version = [2]
  }

  names {
    provider_name_upper = "CognitoIdentity"
    human_friendly      = "Cognito Identity"
  }

  endpoint_info {
    endpoint_api_call        = "ListIdentityPools"
    endpoint_api_params      = "MaxResults: aws_sdkv2.Int32(1)"
  }

  resource_prefix {
    actual  = "aws_cognito_identity_(?!provider)"
    correct = "aws_cognitoidentity_"
  }

  provider_package_correct = "cognitoidentity"
  doc_prefix               = ["cognito_identity_pool"]
  brand                    = "AWS"
}

service "cognitoidp" {

  cli_v2_command {
    aws_cli_v2_command           = "cognito-idp"
    aws_cli_v2_command_no_dashes = "cognitoidp"
  }

  go_packages {
    v1_package = "cognitoidentityprovider"
    v2_package = "cognitoidentityprovider"
  }

  sdk {
    id             = "Cognito Identity Provider"
    client_version = [2]
  }

  names {
    aliases             = ["cognitoidentityprovider"]
    provider_name_upper = "CognitoIDP"
    human_friendly      = "Cognito IDP (Identity Provider)"
  }

  client {
    go_v1_client_typename = "CognitoIdentityProvider"
  }

  endpoint_info {
    endpoint_api_call        = "ListUserPools"
    endpoint_api_params      = "MaxResults: aws_sdkv2.Int32(1)"
  }

  resource_prefix {
    actual  = "aws_cognito_(identity_provider|resource|user|risk)"
    correct = "aws_cognitoidp_"
  }

  provider_package_correct = "cognitoidp"
  doc_prefix               = ["cognito_identity_provider", "cognito_managed_user", "cognito_resource_", "cognito_user", "cognito_risk"]
  brand                    = "AWS"
}

service "cognitosync" {

  cli_v2_command {
    aws_cli_v2_command           = "cognito-sync"
    aws_cli_v2_command_no_dashes = "cognitosync"
  }

  sdk {
    id             = "Cognito Sync"
    client_version = [1]
  }

  names {
    provider_name_upper = "CognitoSync"
    human_friendly      = "Cognito Sync"
  }

  client {
    go_v1_client_typename = "CognitoSync"
  }

  resource_prefix {
    correct = "aws_cognitosync_"
  }

  provider_package_correct = "cognitosync"
  doc_prefix               = ["cognitosync_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "comprehend" {

  sdk {
    id             = "Comprehend"
    client_version = [2]
  }

  names {
    provider_name_upper = "Comprehend"
    human_friendly      = "Comprehend"
  }

  endpoint_info {
    endpoint_api_call        = "ListDocumentClassifiers"
  }

  resource_prefix {
    correct = "aws_comprehend_"
  }

  provider_package_correct = "comprehend"
  doc_prefix               = ["comprehend_"]
  brand                    = "AWS"
}

service "comprehendmedical" {

  sdk {
    id             = "ComprehendMedical"
    client_version = [1]
  }

  names {
    provider_name_upper = "ComprehendMedical"
    human_friendly      = "Comprehend Medical"
  }

  client {
    go_v1_client_typename = "ComprehendMedical"
  }

  resource_prefix {
    correct = "aws_comprehendmedical_"
  }

  provider_package_correct = "comprehendmedical"
  doc_prefix               = ["comprehendmedical_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "computeoptimizer" {

  cli_v2_command {
    aws_cli_v2_command           = "compute-optimizer"
    aws_cli_v2_command_no_dashes = "computeoptimizer"
  }

  sdk {
    id             = "Compute Optimizer"
    client_version = [2]
  }

  names {
    provider_name_upper = "ComputeOptimizer"
    human_friendly      = "Compute Optimizer"
  }

  endpoint_info {
    endpoint_api_call        = "GetEnrollmentStatus"
  }

  resource_prefix {
    correct = "aws_computeoptimizer_"
  }

  provider_package_correct = "computeoptimizer"
  doc_prefix               = ["computeoptimizer_"]
  brand                    = "AWS"
}

service "configservice" {

  sdk {
    id             = "Config Service"
    client_version = [2]
  }

  names {
    aliases             = ["config"]
    provider_name_upper = "ConfigService"
    human_friendly      = "Config"
  }

  endpoint_info {
    endpoint_api_call        = "ListStoredQueries"
  }

  resource_prefix {
    actual  = "aws_config_"
    correct = "aws_configservice_"
  }

  provider_package_correct = "configservice"
  doc_prefix               = ["config_"]
  brand                    = "AWS"
}

service "connect" {

  sdk {
    id             = "Connect"
    client_version = [1]
  }

  names {
    provider_name_upper = "Connect"
    human_friendly      = "Connect"
  }

  client {
    go_v1_client_typename = "Connect"
  }

  endpoint_info {
    endpoint_api_call        = "ListInstances"
  }

  resource_prefix {
    correct = "aws_connect_"
  }

  provider_package_correct = "connect"
  doc_prefix               = ["connect_"]
  brand                    = "AWS"
}

service "connectcases" {

  sdk {
    id             = "ConnectCases"
    client_version = [2]
  }

  names {
    provider_name_upper = "ConnectCases"
    human_friendly      = "Connect Cases"
  }

  client {
    go_v1_client_typename = "ConnectCases"
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
  }

  resource_prefix {
    correct = "aws_connectcases_"
  }

  provider_package_correct = "connectcases"
  doc_prefix               = ["connectcases_"]
  brand                    = "AWS"
}

service "connectcontactlens" {

  cli_v2_command {
    aws_cli_v2_command           = "connect-contact-lens"
    aws_cli_v2_command_no_dashes = "connectcontactlens"
  }

  sdk {
    id             = "Connect Contact Lens"
    client_version = [1]
  }

  names {
    provider_name_upper = "ConnectContactLens"
    human_friendly      = "Connect Contact Lens"
  }

  client {
    go_v1_client_typename = "ConnectContactLens"
  }

  resource_prefix {
    correct = "aws_connectcontactlens_"
  }

  provider_package_correct = "connectcontactlens"
  doc_prefix               = ["connectcontactlens_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "customerprofiles" {

  cli_v2_command {
    aws_cli_v2_command           = "customer-profiles"
    aws_cli_v2_command_no_dashes = "customerprofiles"
  }

  sdk {
    id             = "Customer Profiles"
    client_version = [2]
  }

  names {
    provider_name_upper = "CustomerProfiles"
    human_friendly      = "Connect Customer Profiles"
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
  }

  resource_prefix {
    correct = "aws_customerprofiles_"
  }

  provider_package_correct = "customerprofiles"
  doc_prefix               = ["customerprofiles_"]
  brand                    = "AWS"
}

service "connectparticipant" {

  sdk {
    id             = "ConnectParticipant"
    client_version = [1]
  }

  names {
    provider_name_upper = "ConnectParticipant"
    human_friendly      = "Connect Participant"
  }

  client {
    go_v1_client_typename = "ConnectParticipant"
  }

  resource_prefix {
    correct = "aws_connectparticipant_"
  }

  provider_package_correct = "connectparticipant"
  doc_prefix               = ["connectparticipant_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "voiceid" {

  cli_v2_command {
    aws_cli_v2_command           = "voice-id"
    aws_cli_v2_command_no_dashes = "voiceid"
  }

  sdk {
    id             = "Voice ID"
    client_version = [1]
  }

  names {
    provider_name_upper = "VoiceID"
    human_friendly      = "Connect Voice ID"
  }

  client {
    go_v1_client_typename = "VoiceID"
  }

  resource_prefix {
    correct = "aws_voiceid_"
  }

  provider_package_correct = "voiceid"
  doc_prefix               = ["voiceid_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "wisdom" {

  go_packages {
    v1_package = "connectwisdomservice"
    v2_package = "wisdom"
  }

  sdk {
    id             = "Wisdom"
    client_version = [1]
  }

  names {
    aliases             = ["connectwisdomservice"]
    provider_name_upper = "Wisdom"
    human_friendly      = "Connect Wisdom"
  }

  client {
    go_v1_client_typename = "ConnectWisdomService"
  }

  resource_prefix {
    correct = "aws_wisdom_"
  }

  provider_package_correct = "wisdom"
  doc_prefix               = ["wisdom_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "controltower" {

  sdk {
    id             = "ControlTower"
    client_version = [2]
  }

  names {
    provider_name_upper = "ControlTower"
    human_friendly      = "Control Tower"
  }

  endpoint_info {
    endpoint_api_call        = "ListLandingZones"
  }

  resource_prefix {
    correct = "aws_controltower_"
  }

  provider_package_correct = "controltower"
  doc_prefix               = ["controltower_"]
  brand                    = "AWS"
}

service "costoptimizationhub" {

  cli_v2_command {
    aws_cli_v2_command           = "cost-optimization-hub"
    aws_cli_v2_command_no_dashes = "costoptimizationhub"
  }

  sdk {
    id             = "Cost Optimization Hub"
    client_version = [2]
  }

  names {
    provider_name_upper = "CostOptimizationHub"
    human_friendly      = "Cost Optimization Hub"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "GetPreferences"
    endpoint_region_override = "us-east-1"
  }

  resource_prefix {
    correct = "aws_costoptimizationhub_"
  }

  provider_package_correct = "costoptimizationhub"
  doc_prefix               = ["costoptimizationhub_"]
  brand                    = "AWS"
}

service "cur" {

  go_packages {
    v1_package = "costandusagereportservice"
    v2_package = "costandusagereportservice"
  }

  sdk {
    id             = "Cost and Usage Report Service"
    client_version = [2]
  }

  names {
    aliases             = ["costandusagereportservice"]
    provider_name_upper = "CUR"
    human_friendly      = "Cost and Usage Report"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "DescribeReportDefinitions"
    endpoint_region_override = "us-east-1"
  }

  resource_prefix {
    correct = "aws_cur_"
  }

  provider_package_correct = "cur"
  doc_prefix               = ["cur_"]
  brand                    = "AWS"
}

service "dataexchange" {

  sdk {
    id             = "DataExchange"
    client_version = [2]
  }

  names {
    provider_name_upper = "DataExchange"
    human_friendly      = "Data Exchange"
  }

  client {
    go_v1_client_typename = "DataExchange"
  }

  endpoint_info {
    endpoint_api_call        = "ListDataSets"
  }

  resource_prefix {
    correct = "aws_dataexchange_"
  }

  provider_package_correct = "dataexchange"
  doc_prefix               = ["dataexchange_"]
  brand                    = "AWS"
}

service "datapipeline" {

  sdk {
    id             = "Data Pipeline"
    client_version = [2]
  }

  names {
    provider_name_upper = "DataPipeline"
    human_friendly      = "Data Pipeline"
  }

  client {
    go_v1_client_typename = "DataPipeline"
  }

  endpoint_info {
    endpoint_api_call        = "ListPipelines"
  }

  resource_prefix {
    correct = "aws_datapipeline_"
  }

  provider_package_correct = "datapipeline"
  doc_prefix               = ["datapipeline_"]
  brand                    = "AWS"
}

service "datasync" {

  sdk {
    id             = "DataSync"
    client_version = [2]
  }

  names {
    provider_name_upper = "DataSync"
    human_friendly      = "DataSync"
  }

  endpoint_info {
    endpoint_api_call        = "ListAgents"
  }

  resource_prefix {
    correct = "aws_datasync_"
  }

  provider_package_correct = "datasync"
  doc_prefix               = ["datasync_"]
  brand                    = "AWS"
}

service "datazone" {

  sdk {
    id             = "DataZone"
    client_version = [2]
  }

  names {
    provider_name_upper = "DataZone"
    human_friendly      = "DataZone"
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
  }

  resource_prefix {
    correct = "aws_datazone_"
  }

  provider_package_correct = "datazone"
  doc_prefix               = ["datazone_"]
  brand                    = "AWS"
}

service "detective" {

  sdk {
    id             = "Detective"
    client_version = [2]
  }

  names {
    provider_name_upper = "Detective"
    human_friendly      = "Detective"
  }

  client {
    go_v1_client_typename = "Detective"
  }

  endpoint_info {
    endpoint_api_call        = "ListGraphs"
  }

  resource_prefix {
    correct = "aws_detective_"
  }

  provider_package_correct = "detective"
  doc_prefix               = ["detective_"]
  brand                    = "AWS"
}

service "devicefarm" {

  sdk {
    id             = "Device Farm"
    client_version = [2]
  }

  names {
    provider_name_upper = "DeviceFarm"
    human_friendly      = "Device Farm"
  }

  endpoint_info {
    endpoint_api_call        = "ListDeviceInstances"
  }

  resource_prefix {
    correct = "aws_devicefarm_"
  }

  provider_package_correct = "devicefarm"
  doc_prefix               = ["devicefarm_"]
  brand                    = "AWS"
}

service "devopsguru" {

  cli_v2_command {
    aws_cli_v2_command           = "devops-guru"
    aws_cli_v2_command_no_dashes = "devopsguru"
  }

  sdk {
    id             = "DevOps Guru"
    client_version = [2]
  }

  names {
    provider_name_upper = "DevOpsGuru"
    human_friendly      = "DevOps Guru"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeAccountHealth"
  }

  resource_prefix {
    correct = "aws_devopsguru_"
  }

  provider_package_correct = "devopsguru"
  doc_prefix               = ["devopsguru_"]
  brand                    = "AWS"
}

service "directconnect" {

  sdk {
    id             = "Direct Connect"
    client_version = [1]
  }

  names {
    provider_name_upper = "DirectConnect"
    human_friendly      = "Direct Connect"
  }

  client {
    go_v1_client_typename = "DirectConnect"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeConnections"
  }

  resource_prefix {
    actual  = "aws_dx_"
    correct = "aws_directconnect_"
  }

  provider_package_correct = "directconnect"
  doc_prefix               = ["dx_"]
  brand                    = "AWS"
}

service "dlm" {

  sdk {
    id             = "DLM"
    client_version = [2]
  }

  names {
    provider_name_upper = "DLM"
    human_friendly      = "DLM (Data Lifecycle Manager)"
  }

  endpoint_info {
    endpoint_api_call        = "GetLifecyclePolicies"
  }

  resource_prefix {
    correct = "aws_dlm_"
  }

  provider_package_correct = "dlm"
  doc_prefix               = ["dlm_"]
  brand                    = "AWS"
}

service "dms" {

  go_packages {
    v1_package = "databasemigrationservice"
    v2_package = "databasemigrationservice"
  }

  sdk {
    id             = "Database Migration Service"
    client_version = [2]
  }

  names {
    aliases             = ["databasemigration", "databasemigrationservice"]
    provider_name_upper = "DMS"
    human_friendly      = "DMS (Database Migration)"
  }

  client {
    go_v1_client_typename = "DatabaseMigrationService"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeCertificates"
  }

  resource_prefix {
    correct = "aws_dms_"
  }

  provider_package_correct = "dms"
  doc_prefix               = ["dms_"]
  brand                    = "AWS"
}

service "docdb" {

  sdk {
    id             = "DocDB"
    client_version = [2]
  }

  names {
    provider_name_upper = "DocDB"
    human_friendly      = "DocumentDB"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDBClusters"
  }

  resource_prefix {
    correct = "aws_docdb_"
  }

  provider_package_correct = "docdb"
  doc_prefix               = ["docdb_"]
  brand                    = "AWS"
}

service "docdbelastic" {

  cli_v2_command {
    aws_cli_v2_command           = "docdb-elastic"
    aws_cli_v2_command_no_dashes = "docdbelastic"
  }

  sdk {
    id             = "DocDB Elastic"
    client_version = [2]
  }

  names {
    provider_name_upper = "DocDBElastic"
    human_friendly      = "DocumentDB Elastic"
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
  }

  resource_prefix {
    correct = "aws_docdbelastic_"
  }

  provider_package_correct = "docdbelastic"
  doc_prefix               = ["docdbelastic_"]
  brand                    = "AWS"
}

service "drs" {

  sdk {
    id             = "DRS"
    client_version = [2]
  }

  names {
    provider_name_upper = "DRS"
    human_friendly      = "DRS (Elastic Disaster Recovery)"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeJobs"
  }

  resource_prefix {
    correct = "aws_drs_"
  }

  provider_package_correct = "drs"
  doc_prefix               = ["drs_"]
  brand                    = "AWS"
}

service "ds" {

  go_packages {
    v1_package = "directoryservice"
    v2_package = "directoryservice"
  }

  sdk {
    id             = "Directory Service"
    client_version = [2]
  }

  names {
    aliases             = ["directoryservice"]
    provider_name_upper = "DS"
    human_friendly      = "Directory Service"
  }

  client {
    go_v1_client_typename = "DirectoryService"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDirectories"
  }

  resource_prefix {
    actual  = "aws_directory_service_"
    correct = "aws_ds_"
  }

  provider_package_correct = "ds"
  doc_prefix               = ["directory_service_"]
  brand                    = "AWS"
}

service "dax" {

  sdk {
    id             = "DAX"
    client_version = [2]
  }

  names {
    provider_name_upper = "DAX"
    human_friendly      = "DynamoDB Accelerator (DAX)"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeClusters"
  }

  resource_prefix {
    correct = "aws_dax_"
  }

  provider_package_correct = "dax"
  doc_prefix               = ["dax_"]
  brand                    = "AWS"
}

service "dynamodbstreams" {

  sdk {
    id             = "DynamoDB Streams"
    client_version = [1]
  }

  names {
    provider_name_upper = "DynamoDBStreams"
    human_friendly      = "DynamoDB Streams"
  }

  client {
    go_v1_client_typename = "DynamoDBStreams"
  }

  resource_prefix {
    correct = "aws_dynamodbstreams_"
  }

  provider_package_correct = "dynamodbstreams"
  doc_prefix               = ["dynamodbstreams_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "ebs" {

  sdk {
    id             = "EBS"
    client_version = [1]
  }

  names {
    provider_name_upper = "EBS"
    human_friendly      = "EBS (Elastic Block Store)"
  }

  client {
    go_v1_client_typename = "EBS"
  }

  resource_prefix {
    correct = "aws_ebs_"
  }

  provider_package_correct = "ebs"
  doc_prefix               = ["changewhenimplemented"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "imagebuilder" {

  sdk {
    id             = "imagebuilder"
    client_version = [1]
  }

  names {
    provider_name_upper = "ImageBuilder"
    human_friendly      = "EC2 Image Builder"
  }

  client {
    go_v1_client_typename = "Imagebuilder"
  }

  endpoint_info {
    endpoint_api_call        = "ListImages"
  }

  resource_prefix {
    correct = "aws_imagebuilder_"
  }

  provider_package_correct = "imagebuilder"
  doc_prefix               = ["imagebuilder_"]
  brand                    = "AWS"
}

service "ec2instanceconnect" {

  cli_v2_command {
    aws_cli_v2_command           = "ec2-instance-connect"
    aws_cli_v2_command_no_dashes = "ec2instanceconnect"
  }

  sdk {
    id             = "EC2 Instance Connect"
    client_version = [1]
  }

  names {
    provider_name_upper = "EC2InstanceConnect"
    human_friendly      = "EC2 Instance Connect"
  }

  client {
    go_v1_client_typename = "EC2InstanceConnect"
  }

  resource_prefix {
    correct = "aws_ec2instanceconnect_"
  }

  provider_package_correct = "ec2instanceconnect"
  doc_prefix               = ["ec2instanceconnect_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "ecr" {

  sdk {
    id             = "ECR"
    client_version = [2]
  }

  names {
    provider_name_upper = "ECR"
    human_friendly      = "ECR (Elastic Container Registry)"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeRepositories"
  }

  resource_prefix {
    correct = "aws_ecr_"
  }

  provider_package_correct = "ecr"
  doc_prefix               = ["ecr_"]
  brand                    = "AWS"
}

service "ecrpublic" {

  cli_v2_command {
    aws_cli_v2_command           = "ecr-public"
    aws_cli_v2_command_no_dashes = "ecrpublic"
  }

  sdk {
    id             = "ECR PUBLIC"
    client_version = [2]
  }

  names {
    provider_name_upper = "ECRPublic"
    human_friendly      = "ECR Public"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeRepositories"
  }

  resource_prefix {
    correct = "aws_ecrpublic_"
  }

  provider_package_correct = "ecrpublic"
  doc_prefix               = ["ecrpublic_"]
  brand                    = "AWS"
}

service "ecs" {

  sdk {
    id             = "ECS"
    client_version = [2]
  }

  names {
    provider_name_upper = "ECS"
    human_friendly      = "ECS (Elastic Container)"
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
  }

  resource_prefix {
    correct = "aws_ecs_"
  }

  provider_package_correct = "ecs"
  doc_prefix               = ["ecs_"]
  brand                    = "AWS"
}

service "efs" {

  sdk {
    id             = "EFS"
    client_version = [2]
  }

  names {
    provider_name_upper = "EFS"
    human_friendly      = "EFS (Elastic File System)"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeFileSystems"
  }

  resource_prefix {
    correct = "aws_efs_"
  }

  provider_package_correct = "efs"
  doc_prefix               = ["efs_"]
  brand                    = "AWS"
}

service "eks" {

  sdk {
    id             = "EKS"
    client_version = [2]
  }

  names {
    provider_name_upper = "EKS"
    human_friendly      = "EKS (Elastic Kubernetes)"
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
  }

  resource_prefix {
    correct = "aws_eks_"
  }

  provider_package_correct = "eks"
  doc_prefix               = ["eks_"]
  brand                    = "AWS"
}

service "elasticbeanstalk" {

  sdk {
    id             = "Elastic Beanstalk"
    client_version = [2]
  }

  names {
    aliases             = ["beanstalk"]
    provider_name_upper = "ElasticBeanstalk"
    human_friendly      = "Elastic Beanstalk"
  }

  endpoint_info {
    endpoint_api_call        = "ListAvailableSolutionStacks"
  }

  resource_prefix {
    actual  = "aws_elastic_beanstalk_"
    correct = "aws_elasticbeanstalk_"
  }

  provider_package_correct = "elasticbeanstalk"
  doc_prefix               = ["elastic_beanstalk_"]
  brand                    = "AWS"
}

service "elasticinference" {

  cli_v2_command {
    aws_cli_v2_command           = "elastic-inference"
    aws_cli_v2_command_no_dashes = "elasticinference"
  }

  sdk {
    id             = "Elastic Inference"
    client_version = [1]
  }

  names {
    provider_name_upper = "ElasticInference"
    human_friendly      = "Elastic Inference"
  }

  client {
    go_v1_client_typename = "ElasticInference"
  }

  resource_prefix {
    correct = "aws_elasticinference_"
  }

  provider_package_correct = "elasticinference"
  doc_prefix               = ["elasticinference_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "elastictranscoder" {

  sdk {
    id             = "Elastic Transcoder"
    client_version = [2]
  }

  names {
    provider_name_upper = "ElasticTranscoder"
    human_friendly      = "Elastic Transcoder"
  }

  endpoint_info {
    endpoint_api_call        = "ListPipelines"
  }

  resource_prefix {
    correct = "aws_elastictranscoder_"
  }

  provider_package_correct = "elastictranscoder"
  doc_prefix               = ["elastictranscoder_"]
  brand                    = "AWS"
}

service "elasticache" {

  sdk {
    id             = "ElastiCache"
    client_version = [2]
  }

  names {
    provider_name_upper = "ElastiCache"
    human_friendly      = "ElastiCache"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeCacheClusters"
  }

  resource_prefix {
    correct = "aws_elasticache_"
  }

  provider_package_correct = "elasticache"
  doc_prefix               = ["elasticache_"]
  brand                    = "AWS"
}

service "elasticsearch" {

  cli_v2_command {
    aws_cli_v2_command           = "es"
    aws_cli_v2_command_no_dashes = "es"
  }

  go_packages {
    v1_package = "elasticsearchservice"
    v2_package = "elasticsearchservice"
  }

  sdk {
    id             = "Elasticsearch Service"
    client_version = [1]
  }

  names {
    aliases             = ["es", "elasticsearchservice"]
    provider_name_upper = "Elasticsearch"
    human_friendly      = "Elasticsearch"
  }

  client {
    go_v1_client_typename = "ElasticsearchService"
  }

  endpoint_info {
    endpoint_api_call        = "ListDomainNames"
  }

  resource_prefix {
    actual  = "aws_elasticsearch_"
    correct = "aws_es_"
  }

  provider_package_correct = "es"
  doc_prefix               = ["elasticsearch_"]
  brand                    = "AWS"
}

service "elbv2" {

  go_packages {
    v1_package = "elbv2"
    v2_package = "elasticloadbalancingv2"
  }

  sdk {
    id             = "Elastic Load Balancing v2"
    client_version = [2]
  }

  names {
    aliases             = ["elasticloadbalancingv2"]
    provider_name_upper = "ELBV2"
    human_friendly      = "ELB (Elastic Load Balancing)"
  }

  client {
    go_v1_client_typename = "ELBV2"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeLoadBalancers"
  }

  resource_prefix {
    actual  = "aws_a?lb(\\b|_listener|_target_group|s|_trust_store)"
    correct = "aws_elbv2_"
  }

  provider_package_correct = "elbv2"
  doc_prefix               = ["lbs?\\.", "lb_listener", "lb_target_group", "lb_hosted", "lb_trust_store"]
  brand                    = ""
}
service "elb" {

  go_packages {
    v1_package = "elb"
    v2_package = "elasticloadbalancing"
  }

  sdk {
    id             = "Elastic Load Balancing"
    client_version = [2]
  }

  names {
    aliases             = ["elasticloadbalancing"]
    provider_name_upper = "ELB"
    human_friendly      = "ELB Classic"
  }

  client {
    go_v1_client_typename = "ELB"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeLoadBalancers"
  }

  resource_prefix {
    actual  = "aws_(app_cookie_stickiness_policy|elb|lb_cookie_stickiness_policy|lb_ssl_negotiation_policy|load_balancer_|proxy_protocol_policy)"
    correct = "aws_elb_"
  }

  provider_package_correct = "elb"
  doc_prefix               = ["app_cookie_stickiness_policy", "elb", "lb_cookie_stickiness_policy", "lb_ssl_negotiation_policy", "load_balancer", "proxy_protocol_policy"]
  brand                    = ""
}
service "mediaconnect" {

  sdk {
    id             = "MediaConnect"
    client_version = [2]
  }

  names {
    provider_name_upper = "MediaConnect"
    human_friendly      = "Elemental MediaConnect"
  }

  endpoint_info {
    endpoint_api_call        = "ListBridges"
  }

  resource_prefix {
    correct = "aws_mediaconnect_"
  }

  provider_package_correct = "mediaconnect"
  doc_prefix               = ["mediaconnect_"]
  brand                    = "AWS"
}

service "mediaconvert" {

  sdk {
    id             = "MediaConvert"
    client_version = [2]
  }

  names {
    provider_name_upper = "MediaConvert"
    human_friendly      = "Elemental MediaConvert"
  }
  endpoint_info {
    endpoint_api_call        = "ListJobs"
  }

  resource_prefix {
    actual  = "aws_media_convert_"
    correct = "aws_mediaconvert_"
  }

  provider_package_correct = "mediaconvert"
  doc_prefix               = ["media_convert_"]
  brand                    = "AWS"
}

service "medialive" {

  sdk {
    id             = "MediaLive"
    client_version = [2]
  }

  names {
    provider_name_upper = "MediaLive"
    human_friendly      = "Elemental MediaLive"
  }

  endpoint_info {
    endpoint_api_call        = "ListOfferings"
  }

  resource_prefix {
    correct = "aws_medialive_"
  }

  provider_package_correct = "medialive"
  doc_prefix               = ["medialive_"]
  brand                    = "AWS"
}

service "mediapackage" {

  sdk {
    id             = "MediaPackage"
    client_version = [2]
  }

  names {
    provider_name_upper = "MediaPackage"
    human_friendly      = "Elemental MediaPackage"
  }

  endpoint_info {
    endpoint_api_call        = "ListChannels"
  }

  resource_prefix {
    actual  = "aws_media_package_"
    correct = "aws_mediapackage_"
  }

  provider_package_correct = "mediapackage"
  doc_prefix               = ["media_package_"]
  brand                    = "AWS"
}

service "mediapackagevod" {

  cli_v2_command {
    aws_cli_v2_command           = "mediapackage-vod"
    aws_cli_v2_command_no_dashes = "mediapackagevod"
  }

  sdk {
    id             = "MediaPackage Vod"
    client_version = [1]
  }

  names {
    provider_name_upper = "MediaPackageVOD"
    human_friendly      = "Elemental MediaPackage VOD"
  }

  client {
    go_v1_client_typename = "MediaPackageVod"
  }

  resource_prefix {
    correct = "aws_mediapackagevod_"
  }

  provider_package_correct = "mediapackagevod"
  doc_prefix               = ["mediapackagevod_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "mediastore" {

  sdk {
    id             = "MediaStore"
    client_version = [2]
  }

  names {
    provider_name_upper = "MediaStore"
    human_friendly      = "Elemental MediaStore"
  }

  endpoint_info {
    endpoint_api_call        = "ListContainers"
  }

  resource_prefix {
    actual  = "aws_media_store_"
    correct = "aws_mediastore_"
  }

  provider_package_correct = "mediastore"
  doc_prefix               = ["media_store_"]
  brand                    = "AWS"
}

service "mediastoredata" {

  cli_v2_command {
    aws_cli_v2_command           = "mediastore-data"
    aws_cli_v2_command_no_dashes = "mediastoredata"
  }

  sdk {
    id             = "MediaStore Data"
    client_version = [1]
  }

  names {
    provider_name_upper = "MediaStoreData"
    human_friendly      = "Elemental MediaStore Data"
  }

  client {
    go_v1_client_typename = "MediaStoreData"
  }

  resource_prefix {
    correct = "aws_mediastoredata_"
  }

  provider_package_correct = "mediastoredata"
  doc_prefix               = ["mediastoredata_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "mediatailor" {

  sdk {
    id             = "MediaTailor"
    client_version = [1]
  }

  names {
    provider_name_upper = "MediaTailor"
    human_friendly      = "Elemental MediaTailor"
  }

  client {
    go_v1_client_typename = "MediaTailor"
  }

  resource_prefix {
    correct = "aws_mediatailor_"
  }

  provider_package_correct = "mediatailor"
  doc_prefix               = ["media_tailor_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "emr" {

  sdk {
    id             = "EMR"
    client_version = [1, 2]
  }

  names {
    provider_name_upper = "EMR"
    human_friendly      = "EMR"
  }

  client {
    go_v1_client_typename = "EMR"
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
  }

  resource_prefix {
    correct = "aws_emr_"
  }

  provider_package_correct = "emr"
  doc_prefix               = ["emr_"]
  brand                    = "AWS"
}

service "emrcontainers" {

  cli_v2_command {
    aws_cli_v2_command           = "emr-containers"
    aws_cli_v2_command_no_dashes = "emrcontainers"
  }

  sdk {
    id             = "EMR containers"
    client_version = [2]
  }

  names {
    provider_name_upper = "EMRContainers"
    human_friendly      = "EMR Containers"
  }

  client {
    go_v1_client_typename = "EMRContainers"
  }

  endpoint_info {
    endpoint_api_call        = "ListVirtualClusters"
  }

  resource_prefix {
    correct = "aws_emrcontainers_"
  }

  provider_package_correct = "emrcontainers"
  doc_prefix               = ["emrcontainers_"]
  brand                    = "AWS"
}

service "emrserverless" {

  cli_v2_command {
    aws_cli_v2_command           = "emr-serverless"
    aws_cli_v2_command_no_dashes = "emrserverless"
  }

  sdk {
    id             = "EMR Serverless"
    client_version = [2]
  }

  names {
    provider_name_upper = "EMRServerless"
    human_friendly      = "EMR Serverless"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    correct = "aws_emrserverless_"
  }

  provider_package_correct = "emrserverless"
  doc_prefix               = ["emrserverless_"]
  brand                    = "AWS"
}

service "events" {

  go_packages {
    v1_package = "eventbridge"
    v2_package = "eventbridge"
  }

  sdk {
    id             = "EventBridge"
    client_version = [2]
  }

  names {
    aliases             = ["eventbridge", "cloudwatchevents"]
    provider_name_upper = "Events"
    human_friendly      = "EventBridge"
  }

  endpoint_info {
    endpoint_api_call        = "ListEventBuses"
  }

  resource_prefix {
    actual  = "aws_cloudwatch_event_"
    correct = "aws_events_"
  }

  provider_package_correct = "events"
  doc_prefix               = ["cloudwatch_event_"]
  brand                    = "AWS"
}

service "schemas" {

  sdk {
    id             = "schemas"
    client_version = [2]
  }

  names {
    provider_name_upper = "Schemas"
    human_friendly      = "EventBridge Schemas"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListRegistries"
  }

  resource_prefix {
    correct = "aws_schemas_"
  }

  provider_package_correct = "schemas"
  doc_prefix               = ["schemas_"]
  brand                    = "AWS"
}

service "fis" {

  sdk {
    id             = "fis"
    client_version = [2]
  }

  names {
    provider_name_upper = "FIS"
    human_friendly      = "FIS (Fault Injection Simulator)"
  }

  endpoint_info {
    endpoint_api_call        = "ListExperiments"
  }

  resource_prefix {
    correct = "aws_fis_"
  }

  provider_package_correct = "fis"
  doc_prefix               = ["fis_"]
  brand                    = "AWS"
}

service "finspace" {

  sdk {
    id             = "finspace"
    client_version = [2]
  }

  names {
    provider_name_upper = "FinSpace"
    human_friendly      = "FinSpace"
  }

  endpoint_info {
    endpoint_api_call        = "ListEnvironments"
  }

  resource_prefix {
    correct = "aws_finspace_"
  }

  provider_package_correct = "finspace"
  doc_prefix               = ["finspace_"]
  brand                    = "AWS"
}

service "finspacedata" {

  cli_v2_command {
    aws_cli_v2_command           = "finspace-data"
    aws_cli_v2_command_no_dashes = "finspacedata"
  }

  sdk {
    id             = "finspace data"
    client_version = [1]
  }

  names {
    provider_name_upper = "FinSpaceData"
    human_friendly      = "FinSpace Data"
  }

  client {
    go_v1_client_typename = "FinSpaceData"
  }

  resource_prefix {
    correct = "aws_finspacedata_"
  }

  provider_package_correct = "finspacedata"
  doc_prefix               = ["finspacedata_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "fms" {

  sdk {
    id             = "FMS"
    client_version = [2]
  }

  names {
    provider_name_upper = "FMS"
    human_friendly      = "FMS (Firewall Manager)"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListAppsLists"
    endpoint_api_params      = "MaxResults: aws_sdkv2.Int32(1)"
  }

  resource_prefix {
    correct = "aws_fms_"
  }

  provider_package_correct = "fms"
  doc_prefix               = ["fms_"]
  brand                    = "AWS"
}

service "forecast" {

  go_packages {
    v1_package = "forecastservice"
    v2_package = "forecast"
  }

  sdk {
    id             = "forecast"
    client_version = [1]
  }

  names {
    aliases             = ["forecastservice"]
    provider_name_upper = "Forecast"
    human_friendly      = "Forecast"
  }

  client {
    go_v1_client_typename = "ForecastService"
  }

  resource_prefix {
    correct = "aws_forecast_"
  }

  provider_package_correct = "forecast"
  doc_prefix               = ["forecast_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "forecastquery" {

  go_packages {
    v1_package = "forecastqueryservice"
    v2_package = "forecastquery"
  }

  sdk {
    id             = "forecastquery"
    client_version = [1]
  }

  names {
    aliases             = ["forecastqueryservice"]
    provider_name_upper = "ForecastQuery"
    human_friendly      = "Forecast Query"
  }

  client {
    go_v1_client_typename = "ForecastQueryService"
  }

  resource_prefix {
    correct = "aws_forecastquery_"
  }

  provider_package_correct = "forecastquery"
  doc_prefix               = ["forecastquery_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "frauddetector" {

  sdk {
    id             = "FraudDetector"
    client_version = [1]
  }

  names {
    provider_name_upper = "FraudDetector"
    human_friendly      = "Fraud Detector"
  }

  client {
    go_v1_client_typename = "FraudDetector"
  }

  resource_prefix {
    correct = "aws_frauddetector_"
  }

  provider_package_correct = "frauddetector"
  doc_prefix               = ["frauddetector_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "fsx" {

  sdk {
    id             = "FSx"
    client_version = [2]
  }

  names {
    provider_name_upper = "FSx"
    human_friendly      = "FSx"
  }

  client {
    go_v1_client_typename = "FSx"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeFileSystems"
  }

  resource_prefix {
    correct = "aws_fsx_"
  }

  provider_package_correct = "fsx"
  doc_prefix               = ["fsx_"]
  brand                    = "AWS"
}

service "gamelift" {

  sdk {
    id             = "GameLift"
    client_version = [1]
  }

  names {
    provider_name_upper = "GameLift"
    human_friendly      = "GameLift"
  }

  client {
    go_v1_client_typename = "GameLift"
  }

  endpoint_info {
    endpoint_api_call        = "ListGameServerGroups"
  }

  resource_prefix {
    correct = "aws_gamelift_"
  }

  provider_package_correct = "gamelift"
  doc_prefix               = ["gamelift_"]
  brand                    = "AWS"
}

service "globalaccelerator" {

  sdk {
    id             = "Global Accelerator"
    client_version = [2]
  }

  names {
    provider_name_upper = "GlobalAccelerator"
    human_friendly      = "Global Accelerator"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListAccelerators"
    endpoint_region_override = "us-west-2"
  }

  resource_prefix {
    correct = "aws_globalaccelerator_"
  }

  provider_package_correct = "globalaccelerator"
  doc_prefix               = ["globalaccelerator_"]
  brand                    = "AWS"
}

service "glue" {

  sdk {
    id             = "Glue"
    client_version = [2]
  }

  names {
    provider_name_upper = "Glue"
    human_friendly      = "Glue"
  }

  client {
    go_v1_client_typename = "Glue"
  }

  endpoint_info {
    endpoint_api_call        = "ListRegistries"
  }

  resource_prefix {
    correct = "aws_glue_"
  }

  provider_package_correct = "glue"
  doc_prefix               = ["glue_"]
  brand                    = "AWS"
}

service "databrew" {

  sdk {
    id             = "DataBrew"
    client_version = [2]
  }

  names {
    aliases             = ["gluedatabrew"]
    provider_name_upper = "DataBrew"
    human_friendly      = "Glue DataBrew"
  }

  endpoint_info {
    endpoint_api_call = "ListProjects"
  }

  resource_prefix {
    correct = "aws_databrew_"
  }

  provider_package_correct = "databrew"
  doc_prefix               = ["databrew_"]
  brand                    = "AWS"
}

service "groundstation" {

  sdk {
    id             = "GroundStation"
    client_version = [2]
  }

  names {
    provider_name_upper = "GroundStation"
    human_friendly      = "Ground Station"
  }

  endpoint_info {
    endpoint_api_call        = "ListConfigs"
  }

  resource_prefix {
    correct = "aws_groundstation_"
  }

  provider_package_correct = "groundstation"
  doc_prefix               = ["groundstation_"]
  brand                    = "AWS"
}

service "guardduty" {

  sdk {
    id             = "GuardDuty"
    client_version = [1, 2]
  }

  names {
    provider_name_upper = "GuardDuty"
    human_friendly      = "GuardDuty"
  }

  client {
    go_v1_client_typename = "GuardDuty"
  }

  endpoint_info {
    endpoint_api_call        = "ListDetectors"
  }

  resource_prefix {
    correct = "aws_guardduty_"
  }

  provider_package_correct = "guardduty"
  doc_prefix               = ["guardduty_"]
  brand                    = "AWS"
}

service "health" {

  sdk {
    id             = "Health"
    client_version = [1]
  }

  names {
    provider_name_upper = "Health"
    human_friendly      = "Health"
  }

  client {
    go_v1_client_typename = "Health"
  }

  resource_prefix {
    correct = "aws_health_"
  }

  provider_package_correct = "health"
  doc_prefix               = ["health_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "healthlake" {

  sdk {
    id             = "HealthLake"
    client_version = [2]
  }

  names {
    provider_name_upper = "HealthLake"
    human_friendly      = "HealthLake"
  }

  endpoint_info {
    endpoint_api_call        = "ListFHIRDatastores"
  }

  resource_prefix {
    correct = "aws_healthlake_"
  }

  provider_package_correct = "healthlake"
  doc_prefix               = ["healthlake_"]
  brand                    = "AWS"
}

service "honeycode" {

  sdk {
    id             = "Honeycode"
    client_version = [1]
  }

  names {
    provider_name_upper = "Honeycode"
    human_friendly      = "Honeycode"
  }

  client {
    go_v1_client_typename = "Honeycode"
  }

  resource_prefix {
    correct = "aws_honeycode_"
  }

  provider_package_correct = "honeycode"
  doc_prefix               = ["honeycode_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "iam" {

  sdk {
    id             = "IAM"
    client_version = [2]
  }

  names {
    provider_name_upper = "IAM"
    human_friendly      = "IAM (Identity & Access Management)"
  }

  client {
    go_v1_client_typename = "IAM"
  }

  env_var {
    deprecated_env_var = "AWS_IAM_ENDPOINT"
    tf_aws_env_var     = "TF_AWS_IAM_ENDPOINT"
  }
  endpoint_info {
    endpoint_api_call        = "ListRoles"
  }

  resource_prefix {
    correct = "aws_iam_"
  }

  provider_package_correct = "iam"
  doc_prefix               = ["iam_"]
  brand                    = "AWS"
}

service "inspector" {

  sdk {
    id             = "Inspector"
    client_version = [2]
  }

  names {
    provider_name_upper = "Inspector"
    human_friendly      = "Inspector Classic"
  }

  client {
    go_v1_client_typename = "Inspector"
  }

  endpoint_info {
    endpoint_api_call        = "ListRulesPackages"
  }

  resource_prefix {
    correct = "aws_inspector_"
  }

  provider_package_correct = "inspector"
  doc_prefix               = ["inspector_"]
  brand                    = "AWS"
}

service "inspector2" {

  sdk {
    id             = "Inspector2"
    client_version = [2]
  }

  names {
    aliases             = ["inspectorv2"]
    provider_name_upper = "Inspector2"
    human_friendly      = "Inspector"
  }
  endpoint_info {
    endpoint_api_call        = "ListAccountPermissions"
  }

  resource_prefix {
    correct = "aws_inspector2_"
  }

  provider_package_correct = "inspector2"
  doc_prefix               = ["inspector2_"]
  brand                    = "AWS"
}

service "iot1clickdevices" {

  cli_v2_command {
    aws_cli_v2_command           = "iot1click-devices"
    aws_cli_v2_command_no_dashes = "iot1clickdevices"
  }

  go_packages {
    v1_package = "iot1clickdevicesservice"
    v2_package = "iot1clickdevicesservice"
  }

  sdk {
    id             = "IoT 1Click Devices Service"
    client_version = [1]
  }

  names {
    aliases             = ["iot1clickdevicesservice"]
    provider_name_upper = "IoT1ClickDevices"
    human_friendly      = "IoT 1-Click Devices"
  }

  client {
    go_v1_client_typename = "IoT1ClickDevicesService"
  }

  resource_prefix {
    correct = "aws_iot1clickdevices_"
  }

  provider_package_correct = "iot1clickdevices"
  doc_prefix               = ["iot1clickdevices_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iot1clickprojects" {

  cli_v2_command {
    aws_cli_v2_command           = "iot1click-projects"
    aws_cli_v2_command_no_dashes = "iot1clickprojects"
  }

  sdk {
    id             = "IoT 1Click Projects"
    client_version = [1]
  }

  names {
    provider_name_upper = "IoT1ClickProjects"
    human_friendly      = "IoT 1-Click Projects"
  }

  client {
    go_v1_client_typename = "IoT1ClickProjects"
  }

  resource_prefix {
    correct = "aws_iot1clickprojects_"
  }

  provider_package_correct = "iot1clickprojects"
  doc_prefix               = ["iot1clickprojects_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iotanalytics" {

  sdk {
    id             = "IoTAnalytics"
    client_version = [2]
  }

  names {
    provider_name_upper = "IoTAnalytics"
    human_friendly      = "IoT Analytics"
  }

  client {
    go_v1_client_typename = "IoTAnalytics"
  }

  endpoint_info {
    endpoint_api_call        = "ListChannels"
  }

  resource_prefix {
    correct = "aws_iotanalytics_"
  }

  provider_package_correct = "iotanalytics"
  doc_prefix               = ["iotanalytics_"]
  brand                    = "AWS"
}

service "iotdata" {

  cli_v2_command {
    aws_cli_v2_command           = "iot-data"
    aws_cli_v2_command_no_dashes = "iotdata"
  }

  go_packages {
    v1_package = "iotdataplane"
    v2_package = "iotdataplane"
  }

  sdk {
    id             = "IoT Data Plane"
    client_version = [1]
  }

  names {
    aliases             = ["iotdataplane"]
    provider_name_upper = "IoTData"
    human_friendly      = "IoT Data Plane"
  }

  client {
    go_v1_client_typename = "IoTDataPlane"
  }

  resource_prefix {
    correct = "aws_iotdata_"
  }

  provider_package_correct = "iotdata"
  doc_prefix               = ["iotdata_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iotdeviceadvisor" {

  sdk {
    id             = "IotDeviceAdvisor"
    client_version = [1]
  }

  names {
    provider_name_upper = "IoTDeviceAdvisor"
    human_friendly      = "IoT Device Management"
  }

  client {
    go_v1_client_typename = "IoTDeviceAdvisor"
  }

  resource_prefix {
    correct = "aws_iotdeviceadvisor_"
  }

  provider_package_correct = "iotdeviceadvisor"
  doc_prefix               = ["iotdeviceadvisor_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iotevents" {

  sdk {
    id             = "IoT Events"
    client_version = [2]
  }

  names {
    provider_name_upper = "IoTEvents"
    human_friendly      = "IoT Events"
  }

  client {
    go_v1_client_typename = "IoTEvents"
  }

  endpoint_info {
    endpoint_api_call        = "ListAlarmModels"
  }

  resource_prefix {
    correct = "aws_iotevents_"
  }

  provider_package_correct = "iotevents"
  doc_prefix               = ["iotevents_"]
  brand                    = "AWS"
}

service "ioteventsdata" {

  cli_v2_command {
    aws_cli_v2_command           = "iotevents-data"
    aws_cli_v2_command_no_dashes = "ioteventsdata"
  }

  sdk {
    id             = "IoT Events Data"
    client_version = [1]
  }

  names {
    provider_name_upper = "IoTEventsData"
    human_friendly      = "IoT Events Data"
  }

  client {
    go_v1_client_typename = "IoTEventsData"
  }

  resource_prefix {
    correct = "aws_ioteventsdata_"
  }

  provider_package_correct = "ioteventsdata"
  doc_prefix               = ["ioteventsdata_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iotfleethub" {

  sdk {
    id             = "IoTFleetHub"
    client_version = [1]
  }

  names {
    provider_name_upper = "IoTFleetHub"
    human_friendly      = "IoT Fleet Hub"
  }

  client {
    go_v1_client_typename = "IoTFleetHub"
  }

  resource_prefix {
    correct = "aws_iotfleethub_"
  }

  provider_package_correct = "iotfleethub"
  doc_prefix               = ["iotfleethub_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "greengrass" {

  sdk {
    id             = "Greengrass"
    client_version = [2]
  }

  names {
    provider_name_upper = "Greengrass"
    human_friendly      = "IoT Greengrass"
  }

  client {
    go_v1_client_typename = "Greengrass"
  }

  endpoint_info {
    endpoint_api_call        = "ListGroups"
  }

  resource_prefix {
    correct = "aws_greengrass_"
  }

  provider_package_correct = "greengrass"
  doc_prefix               = ["greengrass_"]
  brand                    = "AWS"
}

service "greengrassv2" {

  sdk {
    id             = "GreengrassV2"
    client_version = [1]
  }

  names {
    provider_name_upper = "GreengrassV2"
    human_friendly      = "IoT Greengrass V2"
  }

  client {
    go_v1_client_typename = "GreengrassV2"
  }

  resource_prefix {
    correct = "aws_greengrassv2_"
  }

  provider_package_correct = "greengrassv2"
  doc_prefix               = ["greengrassv2_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iotjobsdata" {

  cli_v2_command {
    aws_cli_v2_command           = "iot-jobs-data"
    aws_cli_v2_command_no_dashes = "iotjobsdata"
  }

  go_packages {
    v1_package = "iotjobsdataplane"
    v2_package = "iotjobsdataplane"
  }

  sdk {
    id             = "IoT Jobs Data Plane"
    client_version = [1]
  }

  names {
    aliases             = ["iotjobsdataplane"]
    provider_name_upper = "IoTJobsData"
    human_friendly      = "IoT Jobs Data Plane"
  }

  client {
    go_v1_client_typename = "IoTJobsDataPlane"
  }

  resource_prefix {
    correct = "aws_iotjobsdata_"
  }

  provider_package_correct = "iotjobsdata"
  doc_prefix               = ["iotjobsdata_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iotsecuretunneling" {

  sdk {
    id             = "IoTSecureTunneling"
    client_version = [1]
  }

  names {
    provider_name_upper = "IoTSecureTunneling"
    human_friendly      = "IoT Secure Tunneling"
  }

  client {
    go_v1_client_typename = "IoTSecureTunneling"
  }

  resource_prefix {
    correct = "aws_iotsecuretunneling_"
  }

  provider_package_correct = "iotsecuretunneling"
  doc_prefix               = ["iotsecuretunneling_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iotsitewise" {

  sdk {
    id             = "IoTSiteWise"
    client_version = [1]
  }

  names {
    provider_name_upper = "IoTSiteWise"
    human_friendly      = "IoT SiteWise"
  }

  client {
    go_v1_client_typename = "IoTSiteWise"
  }

  resource_prefix {
    correct = "aws_iotsitewise_"
  }

  provider_package_correct = "iotsitewise"
  doc_prefix               = ["iotsitewise_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iotthingsgraph" {

  sdk {
    id             = "IoTThingsGraph"
    client_version = [1]
  }

  names {
    provider_name_upper = "IoTThingsGraph"
    human_friendly      = "IoT Things Graph"
  }

  client {
    go_v1_client_typename = "IoTThingsGraph"
  }

  resource_prefix {
    correct = "aws_iotthingsgraph_"
  }

  provider_package_correct = "iotthingsgraph"
  doc_prefix               = ["iotthingsgraph_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iottwinmaker" {

  sdk {
    id             = "IoTTwinMaker"
    client_version = [1]
  }

  names {
    provider_name_upper = "IoTTwinMaker"
    human_friendly      = "IoT TwinMaker"
  }

  client {
    go_v1_client_typename = "IoTTwinMaker"
  }

  resource_prefix {
    correct = "aws_iottwinmaker_"
  }

  provider_package_correct = "iottwinmaker"
  doc_prefix               = ["iottwinmaker_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "iotwireless" {

  sdk {
    id             = "IoT Wireless"
    client_version = [1]
  }

  names {
    provider_name_upper = "IoTWireless"
    human_friendly      = "IoT Wireless"
  }

  client {
    go_v1_client_typename = "IoTWireless"
  }

  resource_prefix {
    correct = "aws_iotwireless_"
  }

  provider_package_correct = "iotwireless"
  doc_prefix               = ["iotwireless_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "ivs" {

  sdk {
    id             = "ivs"
    client_version = [2]
  }

  names {
    provider_name_upper = "IVS"
    human_friendly      = "IVS (Interactive Video)"
  }

  client {
    go_v1_client_typename = "IVS"
  }

  endpoint_info {
    endpoint_api_call        = "ListChannels"
  }

  resource_prefix {
    correct = "aws_ivs_"
  }

  provider_package_correct = "ivs"
  doc_prefix               = ["ivs_"]
  brand                    = "AWS"
}

service "ivschat" {

  sdk {
    id             = "ivschat"
    client_version = [2]
  }

  names {
    provider_name_upper = "IVSChat"
    human_friendly      = "IVS (Interactive Video) Chat"
  }

  endpoint_info {
    endpoint_api_call        = "ListRooms"
  }

  resource_prefix {
    correct = "aws_ivschat_"
  }

  provider_package_correct = "ivschat"
  doc_prefix               = ["ivschat_"]
  brand                    = "AWS"
}

service "kendra" {

  sdk {
    id             = "kendra"
    client_version = [2]
  }

  names {
    provider_name_upper = "Kendra"
    human_friendly      = "Kendra"
  }

  client {
    go_v1_client_typename = "Kendra"
  }

  endpoint_info {
    endpoint_api_call        = "ListIndices"
  }

  resource_prefix {
    correct = "aws_kendra_"
  }

  provider_package_correct = "kendra"
  doc_prefix               = ["kendra_"]
  brand                    = "AWS"
}

service "keyspaces" {

  sdk {
    id             = "Keyspaces"
    client_version = [2]
  }

  names {
    provider_name_upper = "Keyspaces"
    human_friendly      = "Keyspaces (for Apache Cassandra)"
  }

  endpoint_info {
    endpoint_api_call        = "ListKeyspaces"
  }

  resource_prefix {
    correct = "aws_keyspaces_"
  }

  provider_package_correct = "keyspaces"
  doc_prefix               = ["keyspaces_"]
  brand                    = "AWS"
}

service "kinesis" {

  sdk {
    id             = "Kinesis"
    client_version = [2]
  }

  names {
    provider_name_upper = "Kinesis"
    human_friendly      = "Kinesis"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListStreams"
  }

  resource_prefix {
    actual  = "aws_kinesis_stream"
    correct = "aws_kinesis_"
  }

  provider_package_correct = "kinesis"
  doc_prefix               = ["kinesis_stream", "kinesis_resource_policy"]
  brand                    = "AWS"
}

service "kinesisanalytics" {

  sdk {
    id             = "Kinesis Analytics"
    client_version = [2]
  }

  names {
    provider_name_upper = "KinesisAnalytics"
    human_friendly      = "Kinesis Analytics"
  }

  client {
    go_v1_client_typename = "KinesisAnalytics"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    actual  = "aws_kinesis_analytics_"
    correct = "aws_kinesisanalytics_"
  }

  provider_package_correct = "kinesisanalytics"
  doc_prefix               = ["kinesis_analytics_"]
  brand                    = "AWS"
}

service "kinesisanalyticsv2" {

  sdk {
    id             = "Kinesis Analytics V2"
    client_version = [2]
  }

  names {
    provider_name_upper = "KinesisAnalyticsV2"
    human_friendly      = "Kinesis Analytics V2"
  }

  client {
    go_v1_client_typename = "KinesisAnalyticsV2"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    correct = "aws_kinesisanalyticsv2_"
  }

  provider_package_correct = "kinesisanalyticsv2"
  doc_prefix               = ["kinesisanalyticsv2_"]
  brand                    = "AWS"
}

service "firehose" {

  sdk {
    id             = "Firehose"
    client_version = [2]
  }

  names {
    provider_name_upper = "Firehose"
    human_friendly      = "Kinesis Firehose"
  }

  endpoint_info {
    endpoint_api_call        = "ListDeliveryStreams"
  }

  resource_prefix {
    actual  = "aws_kinesis_firehose_"
    correct = "aws_firehose_"
  }

  provider_package_correct = "firehose"
  doc_prefix               = ["kinesis_firehose_"]
  brand                    = "AWS"
}

service "kinesisvideo" {

  sdk {
    id             = "Kinesis Video"
    client_version = [1]
  }

  names {
    provider_name_upper = "KinesisVideo"
    human_friendly      = "Kinesis Video"
  }

  client {
    go_v1_client_typename = "KinesisVideo"
  }

  endpoint_info {
    endpoint_api_call        = "ListStreams"
  }

  resource_prefix {
    correct = "aws_kinesisvideo_"
  }

  provider_package_correct = "kinesisvideo"
  doc_prefix               = ["kinesis_video_"]
  brand                    = "AWS"
}

service "kinesisvideoarchivedmedia" {

  cli_v2_command {
    aws_cli_v2_command           = "kinesis-video-archived-media"
    aws_cli_v2_command_no_dashes = "kinesisvideoarchivedmedia"
  }

  sdk {
    id             = "Kinesis Video Archived Media"
    client_version = [1]
  }

  names {
    provider_name_upper = "KinesisVideoArchivedMedia"
    human_friendly      = "Kinesis Video Archived Media"
  }

  client {
    go_v1_client_typename = "KinesisVideoArchivedMedia"
  }

  resource_prefix {
    correct = "aws_kinesisvideoarchivedmedia_"
  }

  provider_package_correct = "kinesisvideoarchivedmedia"
  doc_prefix               = ["kinesisvideoarchivedmedia_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "kinesisvideomedia" {

  cli_v2_command {
    aws_cli_v2_command           = "kinesis-video-media"
    aws_cli_v2_command_no_dashes = "kinesisvideomedia"
  }

  sdk {
    id             = "Kinesis Video Media"
    client_version = [1]
  }

  names {
    provider_name_upper = "KinesisVideoMedia"
    human_friendly      = "Kinesis Video Media"
  }

  client {
    go_v1_client_typename = "KinesisVideoMedia"
  }

  resource_prefix {
    correct = "aws_kinesisvideomedia_"
  }

  provider_package_correct = "kinesisvideomedia"
  doc_prefix               = ["kinesisvideomedia_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "kinesisvideosignaling" {

  cli_v2_command {
    aws_cli_v2_command           = "kinesis-video-signaling"
    aws_cli_v2_command_no_dashes = "kinesisvideosignaling"
  }

  go_packages {
    v1_package = "kinesisvideosignalingchannels"
    v2_package = "kinesisvideosignaling"
  }

  sdk {
    id             = "Kinesis Video Signaling"
    client_version = [1]
  }

  names {
    aliases             = ["kinesisvideosignalingchannels"]
    provider_name_upper = "KinesisVideoSignaling"
    human_friendly      = "Kinesis Video Signaling"
  }

  client {
    go_v1_client_typename = "KinesisVideoSignalingChannels"
  }

  resource_prefix {
    correct = "aws_kinesisvideosignaling_"
  }

  provider_package_correct = "kinesisvideosignaling"
  doc_prefix               = ["kinesisvideosignaling_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "kms" {

  sdk {
    id             = "KMS"
    client_version = [2]
  }

  names {
    provider_name_upper = "KMS"
    human_friendly      = "KMS (Key Management)"
  }

  endpoint_info {
    endpoint_api_call        = "ListKeys"
  }

  resource_prefix {
    correct = "aws_kms_"
  }

  provider_package_correct = "kms"
  doc_prefix               = ["kms_"]
  brand                    = "AWS"
}

service "lakeformation" {

  sdk {
    id             = "LakeFormation"
    client_version = [2]
  }

  names {
    provider_name_upper = "LakeFormation"
    human_friendly      = "Lake Formation"
  }

  endpoint_info {
    endpoint_api_call        = "ListResources"
  }

  resource_prefix {
    correct = "aws_lakeformation_"
  }

  provider_package_correct = "lakeformation"
  doc_prefix               = ["lakeformation_"]
  brand                    = "AWS"
}

service "lambda" {

  sdk {
    id             = "Lambda"
    client_version = [2]
  }

  names {
    provider_name_upper = "Lambda"
    human_friendly      = "Lambda"
  }

  endpoint_info {
    endpoint_api_call        = "ListFunctions"
  }

  resource_prefix {
    correct = "aws_lambda_"
  }

  provider_package_correct = "lambda"
  doc_prefix               = ["lambda_"]
  brand                    = "AWS"
}

service "launchwizard" {

  cli_v2_command {
    aws_cli_v2_command           = "launch-wizard"
    aws_cli_v2_command_no_dashes = "launchwizard"
  }

  sdk {
    id             = "Launch Wizard"
    client_version = [2]
  }

  names {
    provider_name_upper = "LaunchWizard"
    human_friendly      = "Launch Wizard"
  }

  endpoint_info {
    endpoint_api_call        = "ListWorkloads"
  }

  resource_prefix {
    correct = "aws_launchwizard_"
  }

  provider_package_correct = "launchwizard"
  doc_prefix               = ["launchwizard_"]
  brand                    = "AWS"
}

service "lexmodels" {

  cli_v2_command {
    aws_cli_v2_command           = "lex-models"
    aws_cli_v2_command_no_dashes = "lexmodels"
  }

  go_packages {
    v1_package = "lexmodelbuildingservice"
    v2_package = "lexmodelbuildingservice"
  }

  sdk {
    id             = "Lex Model Building Service"
    client_version = [1]
  }

  names {
    aliases             = ["lexmodelbuilding", "lexmodelbuildingservice", "lex"]
    provider_name_upper = "LexModels"
    human_friendly      = "Lex Model Building"
  }

  client {
    go_v1_client_typename = "LexModelBuildingService"
  }

  endpoint_info {
    endpoint_api_call        = "GetBots"
  }

  resource_prefix {
    actual  = "aws_lex_"
    correct = "aws_lexmodels_"
  }

  provider_package_correct = "lexmodels"
  doc_prefix               = ["lex_"]
  brand                    = "AWS"
}

service "lexv2models" {

  cli_v2_command {
    aws_cli_v2_command           = "lexv2-models"
    aws_cli_v2_command_no_dashes = "lexv2models"
  }

  go_packages {
    v1_package = "lexmodelsv2"
    v2_package = "lexmodelsv2"
  }

  sdk {
    id             = "Lex Models V2"
    client_version = [2]
  }

  names {
    aliases             = ["lexmodelsv2"]
    provider_name_upper = "LexV2Models"
    human_friendly      = "Lex V2 Models"
  }

  endpoint_info {
    endpoint_api_call        = "ListBots"
  }

  resource_prefix {
    correct = "aws_lexv2models_"
  }

  provider_package_correct = "lexv2models"
  doc_prefix               = ["lexv2models_"]
  brand                    = "AWS"
}

service "lexruntime" {

  cli_v2_command {
    aws_cli_v2_command           = "lex-runtime"
    aws_cli_v2_command_no_dashes = "lexruntime"
  }

  go_packages {
    v1_package = "lexruntimeservice"
    v2_package = "lexruntimeservice"
  }

  sdk {
    id             = "Lex Runtime Service"
    client_version = [1]
  }

  names {
    aliases             = ["lexruntimeservice"]
    provider_name_upper = "LexRuntime"
    human_friendly      = "Lex Runtime"
  }

  client {
    go_v1_client_typename = "LexRuntimeService"
  }

  resource_prefix {
    correct = "aws_lexruntime_"
  }

  provider_package_correct = "lexruntime"
  doc_prefix               = ["lexruntime_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "lexruntimev2" {

  cli_v2_command {
    aws_cli_v2_command           = "lexv2-runtime"
    aws_cli_v2_command_no_dashes = "lexv2runtime"
  }

  sdk {
    id             = "Lex Runtime V2"
    client_version = [1]
  }

  names {
    aliases             = ["lexv2runtime"]
    provider_name_upper = "LexRuntimeV2"
    human_friendly      = "Lex Runtime V2"
  }

  client {
    go_v1_client_typename = "LexRuntimeV2"
  }

  resource_prefix {
    correct = "aws_lexruntimev2_"
  }

  provider_package_correct = "lexruntimev2"
  doc_prefix               = ["lexruntimev2_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "licensemanager" {

  cli_v2_command {
    aws_cli_v2_command           = "license-manager"
    aws_cli_v2_command_no_dashes = "licensemanager"
  }

  sdk {
    id             = "License Manager"
    client_version = [1]
  }

  names {
    provider_name_upper = "LicenseManager"
    human_friendly      = "License Manager"
  }

  client {
    go_v1_client_typename = "LicenseManager"
  }

  endpoint_info {
    endpoint_api_call        = "ListLicenseConfigurations"
  }

  resource_prefix {
    correct = "aws_licensemanager_"
  }

  provider_package_correct = "licensemanager"
  doc_prefix               = ["licensemanager_"]
  brand                    = "AWS"
}

service "lightsail" {

  sdk {
    id             = "Lightsail"
    client_version = [2]
  }

  names {
    provider_name_upper = "Lightsail"
    human_friendly      = "Lightsail"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "GetInstances"
  }

  resource_prefix {
    correct = "aws_lightsail_"
  }

  provider_package_correct = "lightsail"
  doc_prefix               = ["lightsail_"]
  brand                    = "AWS"
}

service "location" {

  go_packages {
    v1_package = "locationservice"
    v2_package = "location"
  }

  sdk {
    id             = "Location"
    client_version = [1]
  }

  names {
    aliases             = ["locationservice"]
    provider_name_upper = "Location"
    human_friendly      = "Location"
  }

  client {
    go_v1_client_typename = "LocationService"
  }

  endpoint_info {
    endpoint_api_call        = "ListGeofenceCollections"
  }

  resource_prefix {
    correct = "aws_location_"
  }

  provider_package_correct = "location"
  doc_prefix               = ["location_"]
  brand                    = "AWS"
}

service "lookoutequipment" {

  sdk {
    id             = "LookoutEquipment"
    client_version = [1]
  }

  names {
    provider_name_upper = "LookoutEquipment"
    human_friendly      = "Lookout for Equipment"
  }

  client {
    go_v1_client_typename = "LookoutEquipment"
  }

  resource_prefix {
    correct = "aws_lookoutequipment_"
  }

  provider_package_correct = "lookoutequipment"
  doc_prefix               = ["lookoutequipment_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "lookoutmetrics" {

  sdk {
    id             = "LookoutMetrics"
    client_version = [2]
  }

  names {
    provider_name_upper = "LookoutMetrics"
    human_friendly      = "Lookout for Metrics"
  }

  endpoint_info {
    endpoint_api_call        = "ListMetricSets"
  }

  resource_prefix {
    correct = "aws_lookoutmetrics_"
  }

  provider_package_correct = "lookoutmetrics"
  doc_prefix               = ["lookoutmetrics_"]
  brand                    = "AWS"
}

service "lookoutvision" {

  go_packages {
    v1_package = "lookoutforvision"
    v2_package = "lookoutvision"
  }

  sdk {
    id             = "LookoutVision"
    client_version = [1]
  }

  names {
    aliases             = ["lookoutforvision"]
    provider_name_upper = "LookoutVision"
    human_friendly      = "Lookout for Vision"
  }

  client {
    go_v1_client_typename = "LookoutForVision"
  }

  resource_prefix {
    correct = "aws_lookoutvision_"
  }

  provider_package_correct = "lookoutvision"
  doc_prefix               = ["lookoutvision_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "machinelearning" {

  sdk {
    id             = "Machine Learning"
    client_version = [1]
  }

  names {
    provider_name_upper = "MachineLearning"
    human_friendly      = "Machine Learning"
  }

  client {
    go_v1_client_typename = "MachineLearning"
  }

  resource_prefix {
    correct = "aws_machinelearning_"
  }

  provider_package_correct = "machinelearning"
  doc_prefix               = ["machinelearning_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "macie2" {

  sdk {
    id             = "Macie2"
    client_version = [1]
  }

  names {
    provider_name_upper = "Macie2"
    human_friendly      = "Macie"
  }

  client {
    go_v1_client_typename = "Macie2"
  }

  endpoint_info {
    endpoint_api_call        = "ListFindings"
  }

  resource_prefix {
    correct = "aws_macie2_"
  }

  provider_package_correct = "macie2"
  doc_prefix               = ["macie2_"]
  brand                    = "AWS"
}

service "macie" {

  sdk {
    id             = "Macie"
    client_version = [1]
  }

  names {
    provider_name_upper = "Macie"
    human_friendly      = "Macie Classic"
  }

  client {
    go_v1_client_typename = "Macie"
  }

  resource_prefix {
    correct = "aws_macie_"
  }

  provider_package_correct = "macie"
  doc_prefix               = ["macie_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "m2" {

  sdk {
    id             = "m2"
    client_version = [2]
  }

  names {
    provider_name_upper = "M2"
    human_friendly      = "Mainframe Modernization"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    correct = "aws_m2_"
  }

  provider_package_correct = "m2"
  doc_prefix               = ["m2_"]
  brand                    = "AWS"
}

service "managedblockchain" {

  sdk {
    id             = "ManagedBlockchain"
    client_version = [1]
  }

  names {
    provider_name_upper = "ManagedBlockchain"
    human_friendly      = "Managed Blockchain"
  }

  client {
    go_v1_client_typename = "ManagedBlockchain"
  }

  resource_prefix {
    correct = "aws_managedblockchain_"
  }

  provider_package_correct = "managedblockchain"
  doc_prefix               = ["managedblockchain_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "grafana" {

  go_packages {
    v1_package = "managedgrafana"
    v2_package = "grafana"
  }

  sdk {
    id             = "grafana"
    client_version = [2]
  }

  names {
    aliases             = ["managedgrafana", "amg"]
    provider_name_upper = "Grafana"
    human_friendly      = "Managed Grafana"
  }

  client {
    go_v1_client_typename = "ManagedGrafana"
  }

  endpoint_info {
    endpoint_api_call        = "ListWorkspaces"
  }

  resource_prefix {
    correct = "aws_grafana_"
  }

  provider_package_correct = "grafana"
  doc_prefix               = ["grafana_"]
  brand                    = "AWS"
}

service "kafka" {

  sdk {
    id             = "Kafka"
    client_version = [2]
  }

  names {
    aliases             = ["msk"]
    provider_name_upper = "Kafka"
    human_friendly      = "Managed Streaming for Kafka"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
  }

  resource_prefix {
    actual  = "aws_msk_"
    correct = "aws_kafka_"
  }

  provider_package_correct = "kafka"
  doc_prefix               = ["msk_"]
  brand                    = "AWS"
}

service "kafkaconnect" {

  sdk {
    id             = "KafkaConnect"
    client_version = [2]
  }

  names {
    provider_name_upper = "KafkaConnect"
    human_friendly      = "Managed Streaming for Kafka Connect"
  }

  client {
    go_v1_client_typename = "KafkaConnect"
  }

  endpoint_info {
    endpoint_api_call        = "ListConnectors"
  }

  resource_prefix {
    actual  = "aws_mskconnect_"
    correct = "aws_kafkaconnect_"
  }

  provider_package_correct = "kafkaconnect"
  doc_prefix               = ["mskconnect_"]
  brand                    = "AWS"
}

service "marketplacecatalog" {

  cli_v2_command {
    aws_cli_v2_command           = "marketplace-catalog"
    aws_cli_v2_command_no_dashes = "marketplacecatalog"
  }

  sdk {
    id             = "Marketplace Catalog"
    client_version = [1]
  }

  names {
    provider_name_upper = "MarketplaceCatalog"
    human_friendly      = "Marketplace Catalog"
  }

  client {
    go_v1_client_typename = "MarketplaceCatalog"
  }

  resource_prefix {
    correct = "aws_marketplacecatalog_"
  }

  provider_package_correct = "marketplacecatalog"
  doc_prefix               = ["marketplace_catalog_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "marketplacecommerceanalytics" {

  sdk {
    id             = "Marketplace Commerce Analytics"
    client_version = [1]
  }

  names {
    provider_name_upper = "MarketplaceCommerceAnalytics"
    human_friendly      = "Marketplace Commerce Analytics"
  }

  client {
    go_v1_client_typename = "MarketplaceCommerceAnalytics"
  }

  resource_prefix {
    correct = "aws_marketplacecommerceanalytics_"
  }

  provider_package_correct = "marketplacecommerceanalytics"
  doc_prefix               = ["marketplacecommerceanalytics_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "marketplaceentitlement" {

  cli_v2_command {
    aws_cli_v2_command           = "marketplace-entitlement"
    aws_cli_v2_command_no_dashes = "marketplaceentitlement"
  }

  go_packages {
    v1_package = "marketplaceentitlementservice"
    v2_package = "marketplaceentitlementservice"
  }

  sdk {
    id             = "Marketplace Entitlement Service"
    client_version = [1]
  }

  names {
    aliases             = ["marketplaceentitlementservice"]
    provider_name_upper = "MarketplaceEntitlement"
    human_friendly      = "Marketplace Entitlement"
  }

  client {
    go_v1_client_typename = "MarketplaceEntitlementService"
  }

  resource_prefix {
    correct = "aws_marketplaceentitlement_"
  }

  provider_package_correct = "marketplaceentitlement"
  doc_prefix               = ["marketplaceentitlement_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "marketplacemetering" {

  cli_v2_command {
    aws_cli_v2_command           = "meteringmarketplace"
    aws_cli_v2_command_no_dashes = "meteringmarketplace"
  }

  sdk {
    id             = "Marketplace Metering"
    client_version = [1]
  }

  names {
    aliases             = ["meteringmarketplace"]
    provider_name_upper = "MarketplaceMetering"
    human_friendly      = "Marketplace Metering"
  }

  client {
    go_v1_client_typename = "MarketplaceMetering"
  }

  resource_prefix {
    correct = "aws_marketplacemetering_"
  }

  provider_package_correct = "marketplacemetering"
  doc_prefix               = ["marketplacemetering_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "memorydb" {

  sdk {
    id             = "MemoryDB"
    client_version = [1]
  }

  names {
    provider_name_upper = "MemoryDB"
    human_friendly      = "MemoryDB for Redis"
  }

  client {
    go_v1_client_typename = "MemoryDB"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeClusters"
  }

  resource_prefix {
    correct = "aws_memorydb_"
  }

  provider_package_correct = "memorydb"
  doc_prefix               = ["memorydb_"]
  brand                    = "AWS"
}

service "meta" {

  go_packages {
    v1_package = ""
    v2_package = ""
  }

  sdk {
    id             = ""
    client_version = null
  }

  names {
    provider_name_upper = "Meta"
    human_friendly      = "Meta Data Sources"
  }

  client {
    skip_client_generate = true
  }

  resource_prefix {
    actual  = "aws_(arn|billing_service_account|default_tags|ip_ranges|partition|regions?|service|service_principal)$"
    correct = "aws_meta_"
  }

  provider_package_correct = "meta"
  doc_prefix               = ["arn", "ip_ranges", "billing_service_account", "default_tags", "partition", "region", "service\\.", "service_principal"]
  brand                    = ""
  exclude                  = true
  allowed_subcategory      = true
  note                     = "Not an AWS service (metadata)"
}
service "mgh" {

  go_packages {
    v1_package = "migrationhub"
    v2_package = "migrationhub"
  }

  sdk {
    id             = "Migration Hub"
    client_version = [1]
  }

  names {
    aliases             = ["migrationhub"]
    provider_name_upper = "MgH"
    human_friendly      = "MgH (Migration Hub)"
  }

  client {
    go_v1_client_typename = "MigrationHub"
  }

  resource_prefix {
    correct = "aws_mgh_"
  }

  provider_package_correct = "mgh"
  doc_prefix               = ["mgh_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "migrationhubconfig" {

  cli_v2_command {
    aws_cli_v2_command           = "migrationhub-config"
    aws_cli_v2_command_no_dashes = "migrationhubconfig"
  }

  sdk {
    id             = "MigrationHub Config"
    client_version = [1]
  }

  names {
    provider_name_upper = "MigrationHubConfig"
    human_friendly      = "Migration Hub Config"
  }

  client {
    go_v1_client_typename = "MigrationHubConfig"
  }

  resource_prefix {
    correct = "aws_migrationhubconfig_"
  }

  provider_package_correct = "migrationhubconfig"
  doc_prefix               = ["migrationhubconfig_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "migrationhubrefactorspaces" {

  cli_v2_command {
    aws_cli_v2_command           = "migration-hub-refactor-spaces"
    aws_cli_v2_command_no_dashes = "migrationhubrefactorspaces"
  }

  sdk {
    id             = "Migration Hub Refactor Spaces"
    client_version = [1]
  }

  names {
    provider_name_upper = "MigrationHubRefactorSpaces"
    human_friendly      = "Migration Hub Refactor Spaces"
  }

  client {
    go_v1_client_typename = "MigrationHubRefactorSpaces"
  }

  resource_prefix {
    correct = "aws_migrationhubrefactorspaces_"
  }

  provider_package_correct = "migrationhubrefactorspaces"
  doc_prefix               = ["migrationhubrefactorspaces_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "migrationhubstrategy" {

  go_packages {
    v1_package = "migrationhubstrategyrecommendations"
    v2_package = "migrationhubstrategy"
  }

  sdk {
    id             = "MigrationHubStrategy"
    client_version = [1]
  }

  names {
    aliases             = ["migrationhubstrategyrecommendations"]
    provider_name_upper = "MigrationHubStrategy"
    human_friendly      = "Migration Hub Strategy"
  }

  client {
    go_v1_client_typename = "MigrationHubStrategyRecommendations"
  }

  resource_prefix {
    correct = "aws_migrationhubstrategy_"
  }

  provider_package_correct = "migrationhubstrategy"
  doc_prefix               = ["migrationhubstrategy_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "mobile" {

  sdk {
    id             = "Mobile"
    client_version = [1]
  }

  names {
    provider_name_upper = "Mobile"
    human_friendly      = "Mobile"
  }

  client {
    go_v1_client_typename = "Mobile"
  }

  resource_prefix {
    correct = "aws_mobile_"
  }

  provider_package_correct = "mobile"
  doc_prefix               = ["mobile_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "mq" {

  sdk {
    id             = "mq"
    client_version = [2]
  }

  names {
    provider_name_upper = "MQ"
    human_friendly      = "MQ"
  }

  endpoint_info {
    endpoint_api_call        = "ListBrokers"
  }

  resource_prefix {
    correct = "aws_mq_"
  }

  provider_package_correct = "mq"
  doc_prefix               = ["mq_"]
  brand                    = "AWS"
}

service "mturk" {

  sdk {
    id             = "MTurk"
    client_version = [1]
  }

  names {
    provider_name_upper = "MTurk"
    human_friendly      = "MTurk (Mechanical Turk)"
  }

  client {
    go_v1_client_typename = "MTurk"
  }

  resource_prefix {
    correct = "aws_mturk_"
  }

  provider_package_correct = "mturk"
  doc_prefix               = ["mturk_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "mwaa" {

  sdk {
    id             = "MWAA"
    client_version = [2]
  }

  names {
    provider_name_upper = "MWAA"
    human_friendly      = "MWAA (Managed Workflows for Apache Airflow)"
  }

  endpoint_info {
    endpoint_api_call        = "ListEnvironments"
  }

  resource_prefix {
    correct = "aws_mwaa_"
  }

  provider_package_correct = "mwaa"
  doc_prefix               = ["mwaa_"]
  brand                    = "AWS"
}

service "neptune" {

  sdk {
    id             = "Neptune"
    client_version = [1]
  }

  names {
    provider_name_upper = "Neptune"
    human_friendly      = "Neptune"
  }

  client {
    go_v1_client_typename = "Neptune"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDBClusters"
  }

  resource_prefix {
    correct = "aws_neptune_"
  }

  provider_package_correct = "neptune"
  doc_prefix               = ["neptune_"]
  brand                    = "AWS"
}

service "neptunegraph" {

  cli_v2_command {
    aws_cli_v2_command           = "neptune-graph"
    aws_cli_v2_command_no_dashes = "neptunegraph"
  }

  go_packages {
    v1_package = ""
    v2_package = "neptunegraph"
  }

  sdk {
    id             = "Neptune Graph"
    client_version = [2]
  }

  names {
    provider_name_upper = "NeptuneGraph"
    human_friendly      = "Neptune Analytics"
  }

  endpoint_info {
    endpoint_api_call        = "ListGraphs"
  }

  resource_prefix {
    correct = "aws_neptunegraph_"
  }

  provider_package_correct = "neptunegraph"
  doc_prefix               = ["neptunegraph_"]
  brand                    = "AWS"
}

service "networkfirewall" {

  cli_v2_command {
    aws_cli_v2_command           = "network-firewall"
    aws_cli_v2_command_no_dashes = "networkfirewall"
  }

  sdk {
    id             = "Network Firewall"
    client_version = [2]
  }

  names {
    provider_name_upper = "NetworkFirewall"
    human_friendly      = "Network Firewall"
  }

  client {
    go_v1_client_typename = "NetworkFirewall"
  }

  endpoint_info {
    endpoint_api_call        = "ListFirewalls"
  }

  resource_prefix {
    correct = "aws_networkfirewall_"
  }

  provider_package_correct = "networkfirewall"
  doc_prefix               = ["networkfirewall_"]
  brand                    = "AWS"
}

service "networkmanager" {

  sdk {
    id             = "NetworkManager"
    client_version = [1]
  }

  names {
    provider_name_upper = "NetworkManager"
    human_friendly      = "Network Manager"
  }

  client {
    go_v1_client_typename = "NetworkManager"
  }

  endpoint_info {
    endpoint_api_call        = "ListCoreNetworks"
  }

  resource_prefix {
    correct = "aws_networkmanager_"
  }

  provider_package_correct = "networkmanager"
  doc_prefix               = ["networkmanager_"]
  brand                    = "AWS"
}

service "nimble" {

  go_packages {
    v1_package = "nimblestudio"
    v2_package = "nimble"
  }

  sdk {
    id             = "nimble"
    client_version = [1]
  }

  names {
    aliases             = ["nimblestudio"]
    provider_name_upper = "Nimble"
    human_friendly      = "Nimble Studio"
  }

  client {
    go_v1_client_typename = "NimbleStudio"
  }

  resource_prefix {
    correct = "aws_nimble_"
  }

  provider_package_correct = "nimble"
  doc_prefix               = ["nimble_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "oam" {

  sdk {
    id             = "OAM"
    client_version = [2]
  }

  names {
    aliases             = ["cloudwatchobservabilityaccessmanager"]
    provider_name_upper = "ObservabilityAccessManager"
    human_friendly      = "CloudWatch Observability Access Manager"
  }

  endpoint_info {
    endpoint_api_call        = "ListLinks"
  }

  resource_prefix {
    correct = "aws_oam_"
  }

  provider_package_correct = "oam"
  doc_prefix               = ["oam_"]
  brand                    = "AWS"
}

service "opensearch" {

  go_packages {
    v1_package = "opensearchservice"
    v2_package = "opensearch"
  }

  sdk {
    id             = "OpenSearch"
    client_version = [1]
  }

  names {
    aliases             = ["opensearchservice"]
    provider_name_upper = "OpenSearch"
    human_friendly      = "OpenSearch"
  }

  client {
    go_v1_client_typename = "OpenSearchService"
  }

  endpoint_info {
    endpoint_api_call        = "ListDomainNames"
  }

  resource_prefix {
    correct = "aws_opensearch_"
  }

  provider_package_correct = "opensearch"
  doc_prefix               = ["opensearch_"]
  brand                    = "AWS"
}

service "opensearchserverless" {

  sdk {
    id             = "OpenSearchServerless"
    client_version = [2]
  }

  names {
    provider_name_upper = "OpenSearchServerless"
    human_friendly      = "OpenSearch Serverless"
  }

  endpoint_info {
    endpoint_api_call        = "ListCollections"
  }

  resource_prefix {
    correct = "aws_opensearchserverless_"
  }

  provider_package_correct = "opensearchserverless"
  doc_prefix               = ["opensearchserverless_"]
  brand                    = "AWS"
}

service "osis" {

  sdk {
    id             = "OSIS"
    client_version = [2]
  }

  names {
    aliases             = ["opensearchingestion"]
    provider_name_upper = "OpenSearchIngestion"
    human_friendly      = "OpenSearch Ingestion"
  }

  endpoint_info {
    endpoint_api_call        = "ListPipelines"
  }

  resource_prefix {
    correct = "aws_osis_"
  }

  provider_package_correct = "osis"
  doc_prefix               = ["osis_"]
  brand                    = "AWS"
}

service "opsworks" {

  sdk {
    id             = "OpsWorks"
    client_version = [1]
  }

  names {
    provider_name_upper = "OpsWorks"
    human_friendly      = "OpsWorks"
  }

  client {
    go_v1_client_typename = "OpsWorks"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeApps"
  }

  resource_prefix {
    correct = "aws_opsworks_"
  }

  provider_package_correct = "opsworks"
  doc_prefix               = ["opsworks_"]
  brand                    = "AWS"
}

service "opsworkscm" {

  cli_v2_command {
    aws_cli_v2_command           = "opsworks-cm"
    aws_cli_v2_command_no_dashes = "opsworkscm"
  }

  sdk {
    id             = "OpsWorksCM"
    client_version = [1]
  }

  names {
    provider_name_upper = "OpsWorksCM"
    human_friendly      = "OpsWorks CM"
  }

  client {
    go_v1_client_typename = "OpsWorksCM"
  }

  resource_prefix {
    correct = "aws_opsworkscm_"
  }

  provider_package_correct = "opsworkscm"
  doc_prefix               = ["opsworkscm_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "organizations" {

  sdk {
    id             = "Organizations"
    client_version = [2]
  }

  names {
    provider_name_upper = "Organizations"
    human_friendly      = "Organizations"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListAccounts"
  }

  resource_prefix {
    correct = "aws_organizations_"
  }

  provider_package_correct = "organizations"
  doc_prefix               = ["organizations_"]
  brand                    = "AWS"
}

service "outposts" {

  sdk {
    id             = "Outposts"
    client_version = [1]
  }

  names {
    provider_name_upper = "Outposts"
    human_friendly      = "Outposts"
  }

  client {
    go_v1_client_typename = "Outposts"
  }

  endpoint_info {
    endpoint_api_call        = "ListSites"
  }

  resource_prefix {
    correct = "aws_outposts_"
  }

  provider_package_correct = "outposts"
  doc_prefix               = ["outposts_"]
  brand                    = "AWS"
}

service "panorama" {

  sdk {
    id             = "Panorama"
    client_version = [1]
  }

  names {
    provider_name_upper = "Panorama"
    human_friendly      = "Panorama"
  }

  client {
    go_v1_client_typename = "Panorama"
  }

  resource_prefix {
    correct = "aws_panorama_"
  }

  provider_package_correct = "panorama"
  doc_prefix               = ["panorama_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "paymentcryptography" {

  cli_v2_command {
    aws_cli_v2_command           = "payment-cryptography"
    aws_cli_v2_command_no_dashes = "paymentcryptography"
  }

  sdk {
    id             = "PaymentCryptography"
    client_version = [2]
  }

  names {
    provider_name_upper = "PaymentCryptography"
    human_friendly      = "Payment Cryptography Control Plane"
  }

  endpoint_info {
    endpoint_api_call        = "ListKeys"
  }

  resource_prefix {
    correct = "aws_paymentcryptography_"
  }

  provider_package_correct = "paymentcryptography"
  doc_prefix               = ["paymentcryptography_"]
  brand                    = "AWS"
}

service "pcaconnectorad" {

  cli_v2_command {
    aws_cli_v2_command           = "pca-connector-ad"
    aws_cli_v2_command_no_dashes = "pcaconnectorad"
  }

  sdk {
    id             = "Pca Connector Ad"
    client_version = [2]
  }

  names {
    provider_name_upper = "PCAConnectorAD"
    human_friendly      = "Private CA Connector for Active Directory"
  }

  endpoint_info {
    endpoint_api_call        = "ListConnectors"
  }

  resource_prefix {
    correct = "aws_pcaconnectorad_"
  }

  provider_package_correct = "pcaconnectorad"
  doc_prefix               = ["pcaconnectorad_"]
  brand                    = "AWS"
}

service "personalize" {

  sdk {
    id             = "Personalize"
    client_version = [1]
  }

  names {
    provider_name_upper = "Personalize"
    human_friendly      = "Personalize"
  }

  client {
    go_v1_client_typename = "Personalize"
  }

  resource_prefix {
    correct = "aws_personalize_"
  }

  provider_package_correct = "personalize"
  doc_prefix               = ["personalize_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "personalizeevents" {

  cli_v2_command {
    aws_cli_v2_command           = "personalize-events"
    aws_cli_v2_command_no_dashes = "personalizeevents"
  }

  sdk {
    id             = "Personalize Events"
    client_version = [1]
  }

  names {
    provider_name_upper = "PersonalizeEvents"
    human_friendly      = "Personalize Events"
  }

  client {
    go_v1_client_typename = "PersonalizeEvents"
  }

  resource_prefix {
    correct = "aws_personalizeevents_"
  }

  provider_package_correct = "personalizeevents"
  doc_prefix               = ["personalizeevents_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "personalizeruntime" {

  cli_v2_command {
    aws_cli_v2_command           = "personalize-runtime"
    aws_cli_v2_command_no_dashes = "personalizeruntime"
  }

  sdk {
    id             = "Personalize Runtime"
    client_version = [1]
  }

  names {
    provider_name_upper = "PersonalizeRuntime"
    human_friendly      = "Personalize Runtime"
  }

  client {
    go_v1_client_typename = "PersonalizeRuntime"
  }

  resource_prefix {
    correct = "aws_personalizeruntime_"
  }

  provider_package_correct = "personalizeruntime"
  doc_prefix               = ["personalizeruntime_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "pinpoint" {

  sdk {
    id             = "Pinpoint"
    client_version = [1]
  }

  names {
    provider_name_upper = "Pinpoint"
    human_friendly      = "Pinpoint"
  }

  client {
    go_v1_client_typename = "Pinpoint"
  }

  endpoint_info {
    endpoint_api_call        = "GetApps"
  }

  resource_prefix {
    correct = "aws_pinpoint_"
  }

  provider_package_correct = "pinpoint"
  doc_prefix               = ["pinpoint_"]
  brand                    = "AWS"
}

service "pinpointemail" {

  cli_v2_command {
    aws_cli_v2_command           = "pinpoint-email"
    aws_cli_v2_command_no_dashes = "pinpointemail"
  }

  sdk {
    id             = "Pinpoint Email"
    client_version = [1]
  }

  names {
    provider_name_upper = "PinpointEmail"
    human_friendly      = "Pinpoint Email"
  }

  client {
    go_v1_client_typename = "PinpointEmail"
  }

  resource_prefix {
    correct = "aws_pinpointemail_"
  }

  provider_package_correct = "pinpointemail"
  doc_prefix               = ["pinpointemail_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "pinpointsmsvoice" {

  cli_v2_command {
    aws_cli_v2_command           = "pinpoint-sms-voice"
    aws_cli_v2_command_no_dashes = "pinpointsmsvoice"
  }

  sdk {
    id             = "Pinpoint SMS Voice"
    client_version = [1]
  }

  names {
    provider_name_upper = "PinpointSMSVoice"
    human_friendly      = "Pinpoint SMS and Voice"
  }

  client {
    go_v1_client_typename = "PinpointSMSVoice"
  }

  resource_prefix {
    correct = "aws_pinpointsmsvoice_"
  }

  provider_package_correct = "pinpointsmsvoice"
  doc_prefix               = ["pinpointsmsvoice_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "pipes" {

  sdk {
    id             = "Pipes"
    client_version = [2]
  }

  names {
    provider_name_upper = "Pipes"
    human_friendly      = "EventBridge Pipes"
  }

  endpoint_info {
    endpoint_api_call        = "ListPipes"
  }

  resource_prefix {
    correct = "aws_pipes_"
  }

  provider_package_correct = "pipes"
  doc_prefix               = ["pipes_"]
  brand                    = "AWS"
}

service "polly" {

  sdk {
    id             = "Polly"
    client_version = [2]
  }

  names {
    provider_name_upper = "Polly"
    human_friendly      = "Polly"
  }

  endpoint_info {
    endpoint_api_call        = "ListLexicons"
  }

  resource_prefix {
    correct = "aws_polly_"
  }

  provider_package_correct = "polly"
  doc_prefix               = ["polly_"]
  brand                    = "AWS"
}

service "pricing" {

  sdk {
    id             = "Pricing"
    client_version = [2]
  }

  names {
    provider_name_upper = "Pricing"
    human_friendly      = "Pricing Calculator"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeServices"
  }

  resource_prefix {
    correct = "aws_pricing_"
  }

  provider_package_correct = "pricing"
  doc_prefix               = ["pricing_"]
  brand                    = "AWS"
}

service "proton" {

  sdk {
    id             = "Proton"
    client_version = [1]
  }

  names {
    provider_name_upper = "Proton"
    human_friendly      = "Proton"
  }

  client {
    go_v1_client_typename = "Proton"
  }

  resource_prefix {
    correct = "aws_proton_"
  }

  provider_package_correct = "proton"
  doc_prefix               = ["proton_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "qbusiness" {

  sdk {
    id             = "QBusiness"
    client_version = [2]
  }

  names {
    provider_name_upper = "QBusiness"
    human_friendly      = "Amazon Q Business"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    correct = "aws_qbusiness_"
  }

  provider_package_correct = "qbusiness"
  doc_prefix               = ["qbusiness_"]
  brand                    = "AWS"
}

service "qldb" {

  sdk {
    id             = "QLDB"
    client_version = [2]
  }

  names {
    provider_name_upper = "QLDB"
    human_friendly      = "QLDB (Quantum Ledger Database)"
  }

  endpoint_info {
    endpoint_api_call        = "ListLedgers"
  }

  resource_prefix {
    correct = "aws_qldb_"
  }

  provider_package_correct = "qldb"
  doc_prefix               = ["qldb_"]
  brand                    = "AWS"
}

service "qldbsession" {

  cli_v2_command {
    aws_cli_v2_command           = "qldb-session"
    aws_cli_v2_command_no_dashes = "qldbsession"
  }

  sdk {
    id             = "QLDB Session"
    client_version = [1]
  }

  names {
    provider_name_upper = "QLDBSession"
    human_friendly      = "QLDB Session"
  }

  client {
    go_v1_client_typename = "QLDBSession"
  }

  resource_prefix {
    correct = "aws_qldbsession_"
  }

  provider_package_correct = "qldbsession"
  doc_prefix               = ["qldbsession_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "quicksight" {

  sdk {
    id             = "QuickSight"
    client_version = [1]
  }

  names {
    provider_name_upper = "QuickSight"
    human_friendly      = "QuickSight"
  }

  client {
    go_v1_client_typename = "QuickSight"
  }

  endpoint_info {
    endpoint_api_call        = "ListDashboards"
    endpoint_api_params      = "AwsAccountId: aws_sdkv1.String(\"123456789012\")"
  }

  resource_prefix {
    correct = "aws_quicksight_"
  }

  provider_package_correct = "quicksight"
  doc_prefix               = ["quicksight_"]
  brand                    = "AWS"
}

service "ram" {

  sdk {
    id             = "RAM"
    client_version = [2]
  }

  names {
    provider_name_upper = "RAM"
    human_friendly      = "RAM (Resource Access Manager)"
  }

  endpoint_info {
    endpoint_api_call        = "ListPermissions"
  }

  resource_prefix {
    correct = "aws_ram_"
  }

  provider_package_correct = "ram"
  doc_prefix               = ["ram_"]
  brand                    = "AWS"
}

service "rds" {

  sdk {
    id             = "RDS"
    client_version = [1, 2]
  }

  names {
    provider_name_upper = "RDS"
    human_friendly      = "RDS (Relational Database)"
  }

  client {
    go_v1_client_typename = "RDS"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDBInstances"
  }

  resource_prefix {
    actual  = "aws_(db_|rds_)"
    correct = "aws_rds_"
  }

  provider_package_correct = "rds"
  doc_prefix               = ["rds_", "db_"]
  brand                    = "AWS"
}

service "rdsdata" {

  cli_v2_command {
    aws_cli_v2_command           = "rds-data"
    aws_cli_v2_command_no_dashes = "rdsdata"
  }

  go_packages {
    v1_package = "rdsdataservice"
    v2_package = "rdsdata"
  }

  sdk {
    id             = "RDS Data"
    client_version = [1]
  }

  names {
    aliases             = ["rdsdataservice"]
    provider_name_upper = "RDSData"
    human_friendly      = "RDS Data"
  }

  client {
    go_v1_client_typename = "RDSDataService"
  }

  resource_prefix {
    correct = "aws_rdsdata_"
  }

  provider_package_correct = "rdsdata"
  doc_prefix               = ["rdsdata_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "pi" {

  sdk {
    id             = "PI"
    client_version = [1]
  }

  names {
    provider_name_upper = "PI"
    human_friendly      = "RDS Performance Insights (PI)"
  }

  client {
    go_v1_client_typename = "PI"
  }

  resource_prefix {
    correct = "aws_pi_"
  }

  provider_package_correct = "pi"
  doc_prefix               = ["pi_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "rbin" {

  go_packages {
    v1_package = "recyclebin"
    v2_package = "rbin"
  }

  sdk {
    id             = "rbin"
    client_version = [2]
  }

  names {
    aliases             = ["recyclebin"]
    provider_name_upper = "RBin"
    human_friendly      = "Recycle Bin (RBin)"
  }

  endpoint_info {
    endpoint_api_call        = "ListRules"
    endpoint_api_params      = "ResourceType: awstypes.ResourceTypeEc2Image"
  }

  resource_prefix {
    correct = "aws_rbin_"
  }

  provider_package_correct = "rbin"
  doc_prefix               = ["rbin_"]
  brand                    = "AWS"
}

service "redshift" {

  sdk {
    id             = "Redshift"
    client_version = [1, 2]
  }

  names {
    provider_name_upper = "Redshift"
    human_friendly      = "Redshift"
  }

  client {
    go_v1_client_typename = "Redshift"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeClusters"
  }

  resource_prefix {
    correct = "aws_redshift_"
  }

  provider_package_correct = "redshift"
  doc_prefix               = ["redshift_"]
  brand                    = "AWS"
}

service "redshiftdata" {

  cli_v2_command {
    aws_cli_v2_command           = "redshift-data"
    aws_cli_v2_command_no_dashes = "redshiftdata"
  }

  go_packages {
    v1_package = "redshiftdataapiservice"
    v2_package = "redshiftdata"
  }

  sdk {
    id             = "Redshift Data"
    client_version = [2]
  }

  names {
    aliases             = ["redshiftdataapiservice"]
    provider_name_upper = "RedshiftData"
    human_friendly      = "Redshift Data"
  }

  endpoint_info {
    endpoint_api_call        = "ListDatabases"
    endpoint_api_params      = "Database: aws_sdkv2.String(\"test\")"
  }

  resource_prefix {
    correct = "aws_redshiftdata_"
  }

  provider_package_correct = "redshiftdata"
  doc_prefix               = ["redshiftdata_"]
  brand                    = "AWS"
}

service "redshiftserverless" {

  cli_v2_command {
    aws_cli_v2_command           = "redshift-serverless"
    aws_cli_v2_command_no_dashes = "redshiftserverless"
  }

  sdk {
    id             = "Redshift Serverless"
    client_version = [1, 2]
  }

  names {
    provider_name_upper = "RedshiftServerless"
    human_friendly      = "Redshift Serverless"
  }

  client {
    go_v1_client_typename = "RedshiftServerless"
  }

  endpoint_info {
    endpoint_api_call        = "ListNamespaces"
  }

  resource_prefix {
    correct = "aws_redshiftserverless_"
  }

  provider_package_correct = "redshiftserverless"
  doc_prefix               = ["redshiftserverless_"]
  brand                    = "AWS"
}

service "rekognition" {

  sdk {
    id             = "Rekognition"
    client_version = [2]
  }

  names {
    provider_name_upper = "Rekognition"
    human_friendly      = "Rekognition"
  }

  endpoint_info {
    endpoint_api_call        = "ListCollections"
  }

  resource_prefix {
    correct = "aws_rekognition_"
  }

  provider_package_correct = "rekognition"
  doc_prefix               = ["rekognition_"]
  brand                    = "AWS"
}

service "resiliencehub" {

  sdk {
    id             = "resiliencehub"
    client_version = [1]
  }

  names {
    provider_name_upper = "ResilienceHub"
    human_friendly      = "Resilience Hub"
  }

  client {
    go_v1_client_typename = "ResilienceHub"
  }

  resource_prefix {
    correct = "aws_resiliencehub_"
  }

  provider_package_correct = "resiliencehub"
  doc_prefix               = ["resiliencehub_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "resourceexplorer2" {

  cli_v2_command {
    aws_cli_v2_command           = "resource-explorer-2"
    aws_cli_v2_command_no_dashes = "resourceexplorer2"
  }

  sdk {
    id             = "Resource Explorer 2"
    client_version = [2]
  }

  names {
    provider_name_upper = "ResourceExplorer2"
    human_friendly      = "Resource Explorer"
  }


  endpoint_info {
    endpoint_api_call        = "ListIndexes"
  }

  resource_prefix {
    correct = "aws_resourceexplorer2_"
  }

  provider_package_correct = "resourceexplorer2"
  doc_prefix               = ["resourceexplorer2_"]
  brand                    = "AWS"
}

service "resourcegroups" {

  cli_v2_command {
    aws_cli_v2_command           = "resource-groups"
    aws_cli_v2_command_no_dashes = "resourcegroups"
  }

  sdk {
    id             = "Resource Groups"
    client_version = [2]
  }

  names {
    provider_name_upper = "ResourceGroups"
    human_friendly      = "Resource Groups"
  }

  endpoint_info {
    endpoint_api_call        = "ListGroups"
  }

  resource_prefix {
    correct = "aws_resourcegroups_"
  }

  provider_package_correct = "resourcegroups"
  doc_prefix               = ["resourcegroups_"]
  brand                    = "AWS"
}

service "resourcegroupstaggingapi" {

  sdk {
    id             = "Resource Groups Tagging API"
    client_version = [2]
  }

  names {
    aliases             = ["resourcegroupstagging"]
    provider_name_upper = "ResourceGroupsTaggingAPI"
    human_friendly      = "Resource Groups Tagging"
  }

  endpoint_info {
    endpoint_api_call        = "GetResources"
  }

  resource_prefix {
    correct = "aws_resourcegroupstaggingapi_"
  }

  provider_package_correct = "resourcegroupstaggingapi"
  doc_prefix               = ["resourcegroupstaggingapi_"]
  brand                    = "AWS"
}

service "robomaker" {

  sdk {
    id             = "RoboMaker"
    client_version = [1]
  }

  names {
    provider_name_upper = "RoboMaker"
    human_friendly      = "RoboMaker"
  }

  client {
    go_v1_client_typename = "RoboMaker"
  }

  resource_prefix {
    correct = "aws_robomaker_"
  }

  provider_package_correct = "robomaker"
  doc_prefix               = ["robomaker_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "rolesanywhere" {

  sdk {
    id             = "RolesAnywhere"
    client_version = [2]
  }

  names {
    provider_name_upper = "RolesAnywhere"
    human_friendly      = "Roles Anywhere"
  }

  endpoint_info {
    endpoint_api_call        = "ListProfiles"
  }

  resource_prefix {
    correct = "aws_rolesanywhere_"
  }

  provider_package_correct = "rolesanywhere"
  doc_prefix               = ["rolesanywhere_"]
  brand                    = "AWS"
}

service "route53" {

  sdk {
    id             = "Route 53"
    client_version = [2]
  }

  names {
    provider_name_upper = "Route53"
    human_friendly      = "Route 53"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListHostedZones"
    endpoint_region_override = "us-east-1"
  }

  resource_prefix {
    actual  = "aws_route53_(?!resolver_)"
    correct = "aws_route53_"
  }

  provider_package_correct = "route53"
  doc_prefix               = ["route53_cidr_", "route53_delegation_", "route53_health_", "route53_hosted_", "route53_key_", "route53_query_", "route53_record", "route53_traffic_", "route53_vpc_", "route53_zone"]
  brand                    = "AWS"
}

service "route53domains" {

  sdk {
    id             = "Route 53 Domains"
    client_version = [2]
  }

  names {
    provider_name_upper = "Route53Domains"
    human_friendly      = "Route 53 Domains"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
    endpoint_region_override = "us-east-1"
  }

  resource_prefix {
    correct = "aws_route53domains_"
  }

  provider_package_correct = "route53domains"
  doc_prefix               = ["route53domains_"]
  brand                    = "AWS"
}

service "route53profiles" {

  sdk {
    id             = "Route 53 Profiles"
    client_version = [2]
  }

  names {
    provider_name_upper = "Route53Profiles"
    human_friendly      = "Route 53 Profiles"
  }

  endpoint_info {
    endpoint_api_call        = "ListProfiles"
  }

  resource_prefix {
    correct = "aws_route53profiles_"
  }

  provider_package_correct = "route53profiles"
  doc_prefix               = ["route53profiles_"]
  brand                    = "AWS"
}

service "route53recoverycluster" {

  cli_v2_command {
    aws_cli_v2_command           = "route53-recovery-cluster"
    aws_cli_v2_command_no_dashes = "route53recoverycluster"
  }

  sdk {
    id             = "Route53 Recovery Cluster"
    client_version = [1]
  }

  names {
    provider_name_upper = "Route53RecoveryCluster"
    human_friendly      = "Route 53 Recovery Cluster"
  }

  client {
    go_v1_client_typename = "Route53RecoveryCluster"
  }

  resource_prefix {
    correct = "aws_route53recoverycluster_"
  }

  provider_package_correct = "route53recoverycluster"
  doc_prefix               = ["route53recoverycluster_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "route53recoverycontrolconfig" {

  cli_v2_command {
    aws_cli_v2_command           = "route53-recovery-control-config"
    aws_cli_v2_command_no_dashes = "route53recoverycontrolconfig"
  }

  sdk {
    id             = "Route53 Recovery Control Config"
    client_version = [1]
  }

  names {
    provider_name_upper = "Route53RecoveryControlConfig"
    human_friendly      = "Route 53 Recovery Control Config"
  }

  client {
    go_v1_client_typename = "Route53RecoveryControlConfig"
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
  }

  resource_prefix {
    correct = "aws_route53recoverycontrolconfig_"
  }

  provider_package_correct = "route53recoverycontrolconfig"
  doc_prefix               = ["route53recoverycontrolconfig_"]
  brand                    = "AWS"
}

service "route53recoveryreadiness" {

  cli_v2_command {
    aws_cli_v2_command           = "route53-recovery-readiness"
    aws_cli_v2_command_no_dashes = "route53recoveryreadiness"
  }

  sdk {
    id             = "Route53 Recovery Readiness"
    client_version = [1]
  }

  names {
    provider_name_upper = "Route53RecoveryReadiness"
    human_friendly      = "Route 53 Recovery Readiness"
  }

  client {
    go_v1_client_typename = "Route53RecoveryReadiness"
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListCells"
  }

  resource_prefix {
    correct = "aws_route53recoveryreadiness_"
  }

  provider_package_correct = "route53recoveryreadiness"
  doc_prefix               = ["route53recoveryreadiness_"]
  brand                    = "AWS"
}

service "route53resolver" {

  sdk {
    id             = "Route53Resolver"
    client_version = [1]
  }

  names {
    provider_name_upper = "Route53Resolver"
    human_friendly      = "Route 53 Resolver"
  }

  client {
    go_v1_client_typename = "Route53Resolver"
  }

  endpoint_info {
    endpoint_api_call        = "ListFirewallDomainLists"
  }

  resource_prefix {
    actual  = "aws_route53_resolver_"
    correct = "aws_route53resolver_"
  }

  provider_package_correct = "route53resolver"
  doc_prefix               = ["route53_resolver_"]
  brand                    = "AWS"
}

service "s3" {

  cli_v2_command {
    aws_cli_v2_command           = "s3api"
    aws_cli_v2_command_no_dashes = "s3api"
  }

  sdk {
    id             = "S3"
    client_version = [2]
  }

  names {
    aliases             = ["s3api"]
    provider_name_upper = "S3"
    human_friendly      = "S3 (Simple Storage)"
  }

  client {
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = "AWS_S3_ENDPOINT"
    tf_aws_env_var     = "TF_AWS_S3_ENDPOINT"
  }
  endpoint_info {
    endpoint_api_call        = "ListBuckets"
  }

  resource_prefix {
    actual  = "aws_(canonical_user_id|s3_bucket|s3_object|s3_directory_bucket)"
    correct = "aws_s3_"
  }

  provider_package_correct = "s3"
  doc_prefix               = ["s3_bucket", "s3_directory_bucket", "s3_object", "canonical_user_id"]
  brand                    = "AWS"
}

service "s3control" {

  sdk {
    id             = "S3 Control"
    client_version = [2]
  }

  names {
    provider_name_upper = "S3Control"
    human_friendly      = "S3 Control"
  }

  endpoint_info {
    endpoint_api_call        = "ListJobs"
  }

  resource_prefix {
    actual  = "aws_(s3_account_|s3control_|s3_access_)"
    correct = "aws_s3control_"
  }

  provider_package_correct = "s3control"
  doc_prefix               = ["s3control", "s3_account_", "s3_access_"]
  brand                    = "AWS"
}

service "glacier" {

  sdk {
    id             = "Glacier"
    client_version = [2]
  }

  names {
    provider_name_upper = "Glacier"
    human_friendly      = "S3 Glacier"
  }

  endpoint_info {
    endpoint_api_call        = "ListVaults"
  }

  resource_prefix {
    correct = "aws_glacier_"
  }

  provider_package_correct = "glacier"
  doc_prefix               = ["glacier_"]
  brand                    = "AWS"
}

service "s3outposts" {

  sdk {
    id             = "S3Outposts"
    client_version = [1]
  }

  names {
    provider_name_upper = "S3Outposts"
    human_friendly      = "S3 on Outposts"
  }

  client {
    go_v1_client_typename = "S3Outposts"
  }

  endpoint_info {
    endpoint_api_call        = "ListEndpoints"
  }

  resource_prefix {
    correct = "aws_s3outposts_"
  }

  provider_package_correct = "s3outposts"
  doc_prefix               = ["s3outposts_"]
  brand                    = "AWS"
}

service "sagemaker" {

  sdk {
    id             = "SageMaker"
    client_version = [1]
  }

  names {
    provider_name_upper = "SageMaker"
    human_friendly      = "SageMaker"
  }

  client {
    go_v1_client_typename = "SageMaker"
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
  }

  resource_prefix {
    correct = "aws_sagemaker_"
  }

  provider_package_correct = "sagemaker"
  doc_prefix               = ["sagemaker_"]
  brand                    = "AWS"
}

service "sagemakera2iruntime" {

  cli_v2_command {
    aws_cli_v2_command           = "sagemaker-a2i-runtime"
    aws_cli_v2_command_no_dashes = "sagemakera2iruntime"
  }

  go_packages {
    v1_package = "augmentedairuntime"
    v2_package = "sagemakera2iruntime"
  }

  sdk {
    id             = "SageMaker A2I Runtime"
    client_version = [1]
  }

  names {
    aliases             = ["augmentedairuntime"]
    provider_name_upper = "SageMakerA2IRuntime"
    human_friendly      = "SageMaker A2I (Augmented AI)"
  }

  client {
    go_v1_client_typename = "AugmentedAIRuntime"
  }

  resource_prefix {
    correct = "aws_sagemakera2iruntime_"
  }

  provider_package_correct = "sagemakera2iruntime"
  doc_prefix               = ["sagemakera2iruntime_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "sagemakeredge" {

  cli_v2_command {
    aws_cli_v2_command           = "sagemaker-edge"
    aws_cli_v2_command_no_dashes = "sagemakeredge"
  }

  go_packages {
    v1_package = "sagemakeredgemanager"
    v2_package = "sagemakeredge"
  }

  sdk {
    id             = "Sagemaker Edge"
    client_version = [1]
  }

  names {
    aliases             = ["sagemakeredgemanager"]
    provider_name_upper = "SageMakerEdge"
    human_friendly      = "SageMaker Edge Manager"
  }

  client {
    go_v1_client_typename = "SagemakerEdgeManager"
  }

  resource_prefix {
    correct = "aws_sagemakeredge_"
  }

  provider_package_correct = "sagemakeredge"
  doc_prefix               = ["sagemakeredge_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "sagemakerfeaturestoreruntime" {

  cli_v2_command {
    aws_cli_v2_command           = "sagemaker-featurestore-runtime"
    aws_cli_v2_command_no_dashes = "sagemakerfeaturestoreruntime"
  }

  sdk {
    id             = "SageMaker FeatureStore Runtime"
    client_version = [1]
  }

  names {
    provider_name_upper = "SageMakerFeatureStoreRuntime"
    human_friendly      = "SageMaker Feature Store Runtime"
  }

  client {
    go_v1_client_typename = "SageMakerFeatureStoreRuntime"
  }

  resource_prefix {
    correct = "aws_sagemakerfeaturestoreruntime_"
  }

  provider_package_correct = "sagemakerfeaturestoreruntime"
  doc_prefix               = ["sagemakerfeaturestoreruntime_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "sagemakerruntime" {

  cli_v2_command {
    aws_cli_v2_command           = "sagemaker-runtime"
    aws_cli_v2_command_no_dashes = "sagemakerruntime"
  }

  sdk {
    id             = "SageMaker Runtime"
    client_version = [1]
  }

  names {
    provider_name_upper = "SageMakerRuntime"
    human_friendly      = "SageMaker Runtime"
  }

  client {
    go_v1_client_typename = "SageMakerRuntime"
  }

  resource_prefix {
    correct = "aws_sagemakerruntime_"
  }

  provider_package_correct = "sagemakerruntime"
  doc_prefix               = ["sagemakerruntime_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "savingsplans" {

  sdk {
    id             = "savingsplans"
    client_version = [1]
  }

  names {
    provider_name_upper = "SavingsPlans"
    human_friendly      = "Savings Plans"
  }

  client {
    go_v1_client_typename = "SavingsPlans"
  }

  resource_prefix {
    correct = "aws_savingsplans_"
  }

  provider_package_correct = "savingsplans"
  doc_prefix               = ["savingsplans_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "simpledb" {

  cli_v2_command {
    aws_cli_v2_command           = "sdb"
    aws_cli_v2_command_no_dashes = "sdb"
  }

  go_packages {
    v1_package = "simpledb"
    v2_package = ""
  }

  sdk {
    id             = "SimpleDB"
    client_version = [1]
  }

  names {
    aliases             = ["sdb"]
    provider_name_upper = "SimpleDB"
    human_friendly      = "SDB (SimpleDB)"
  }

  client {
    go_v1_client_typename = "SimpleDB"
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
  }

  resource_prefix {
    actual  = "aws_simpledb_"
    correct = "aws_sdb_"
  }

  provider_package_correct = "sdb"
  doc_prefix               = ["simpledb_"]
  brand                    = "AWS"
}

service "scheduler" {

  sdk {
    id             = "Scheduler"
    client_version = [2]
  }

  names {
    provider_name_upper = "Scheduler"
    human_friendly      = "EventBridge Scheduler"
  }

  endpoint_info {
    endpoint_api_call        = "ListSchedules"
  }

  resource_prefix {
    correct = "aws_scheduler_"
  }

  provider_package_correct = "scheduler"
  doc_prefix               = ["scheduler_"]
  brand                    = "AWS"
}

service "secretsmanager" {

  sdk {
    id             = "Secrets Manager"
    client_version = [2]
  }

  names {
    provider_name_upper = "SecretsManager"
    human_friendly      = "Secrets Manager"
  }

  endpoint_info {
    endpoint_api_call        = "ListSecrets"
  }

  resource_prefix {
    correct = "aws_secretsmanager_"
  }

  provider_package_correct = "secretsmanager"
  doc_prefix               = ["secretsmanager_"]
  brand                    = "AWS"
}

service "securityhub" {

  sdk {
    id             = "SecurityHub"
    client_version = [2]
  }

  names {
    provider_name_upper = "SecurityHub"
    human_friendly      = "Security Hub"
  }

  endpoint_info {
    endpoint_api_call        = "ListAutomationRules"
  }

  resource_prefix {
    correct = "aws_securityhub_"
  }

  provider_package_correct = "securityhub"
  doc_prefix               = ["securityhub_"]
  brand                    = "AWS"
}

service "securitylake" {

  sdk {
    id             = "SecurityLake"
    client_version = [2]
  }

  names {
    provider_name_upper = "SecurityLake"
    human_friendly      = "Security Lake"
  }

  endpoint_info {
    endpoint_api_call        = "ListDataLakes"
  }

  resource_prefix {
    correct = "aws_securitylake_"
  }

  provider_package_correct = "securitylake"
  doc_prefix               = ["securitylake_"]
  brand                    = "AWS"
}

service "serverlessrepo" {

  go_packages {
    v1_package = "serverlessapplicationrepository"
    v2_package = "serverlessapplicationrepository"
  }

  sdk {
    id             = "ServerlessApplicationRepository"
    client_version = [2]
  }

  names {
    aliases             = ["serverlessapprepo", "serverlessapplicationrepository"]
    provider_name_upper = "ServerlessRepo"
    human_friendly      = "Serverless Application Repository"
  }

  client {
    go_v1_client_typename = "ServerlessApplicationRepository"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    actual  = "aws_serverlessapplicationrepository_"
    correct = "aws_serverlessrepo_"
  }

  provider_package_correct = "serverlessrepo"
  doc_prefix               = ["serverlessapplicationrepository_"]
  brand                    = "AWS"
}

service "servicecatalog" {

  sdk {
    id             = "Service Catalog"
    client_version = [1]
  }

  names {
    provider_name_upper = "ServiceCatalog"
    human_friendly      = "Service Catalog"
  }

  client {
    go_v1_client_typename = "ServiceCatalog"
  }

  endpoint_info {
    endpoint_api_call        = "ListPortfolios"
  }

  resource_prefix {
    correct = "aws_servicecatalog_"
  }

  provider_package_correct = "servicecatalog"
  doc_prefix               = ["servicecatalog_"]
  brand                    = "AWS"
}

service "servicecatalogappregistry" {

  cli_v2_command {
    aws_cli_v2_command           = "servicecatalog-appregistry"
    aws_cli_v2_command_no_dashes = "servicecatalogappregistry"
  }

  go_packages {
    v1_package = "appregistry"
    v2_package = "servicecatalogappregistry"
  }

  sdk {
    id             = "Service Catalog AppRegistry"
    client_version = [2]
  }

  names {
    aliases             = ["appregistry"]
    provider_name_upper = "ServiceCatalogAppRegistry"
    human_friendly      = "Service Catalog AppRegistry"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    correct = "aws_servicecatalogappregistry_"
  }

  provider_package_correct = "servicecatalogappregistry"
  doc_prefix               = ["servicecatalogappregistry_"]
  brand                    = "AWS"
}

service "servicequotas" {

  cli_v2_command {
    aws_cli_v2_command           = "service-quotas"
    aws_cli_v2_command_no_dashes = "servicequotas"
  }

  sdk {
    id             = "Service Quotas"
    client_version = [2]
  }

  names {
    provider_name_upper = "ServiceQuotas"
    human_friendly      = "Service Quotas"
  }

  endpoint_info {
    endpoint_api_call        = "ListServices"
  }

  resource_prefix {
    correct = "aws_servicequotas_"
  }

  provider_package_correct = "servicequotas"
  doc_prefix               = ["servicequotas_"]
  brand                    = ""
}
service "ses" {

  sdk {
    id             = "SES"
    client_version = [1]
  }

  names {
    provider_name_upper = "SES"
    human_friendly      = "SES (Simple Email)"
  }

  client {
    go_v1_client_typename = "SES"
  }

  endpoint_info {
    endpoint_api_call        = "ListIdentities"
  }

  resource_prefix {
    correct = "aws_ses_"
  }

  provider_package_correct = "ses"
  doc_prefix               = ["ses_"]
  brand                    = "AWS"
}

service "sesv2" {

  sdk {
    id             = "SESv2"
    client_version = [2]
  }

  names {
    provider_name_upper = "SESV2"
    human_friendly      = "SESv2 (Simple Email V2)"
  }

  endpoint_info {
    endpoint_api_call        = "ListContactLists"
  }

  resource_prefix {
    correct = "aws_sesv2_"
  }

  provider_package_correct = "sesv2"
  doc_prefix               = ["sesv2_"]
  brand                    = "AWS"
}

service "sfn" {

  cli_v2_command {
    aws_cli_v2_command           = "stepfunctions"
    aws_cli_v2_command_no_dashes = "stepfunctions"
  }

  sdk {
    id             = "SFN"
    client_version = [2]
  }

  names {
    aliases             = ["stepfunctions"]
    provider_name_upper = "SFN"
    human_friendly      = "SFN (Step Functions)"
  }

  client {
    go_v1_client_typename = "SFN"
  }

  endpoint_info {
    endpoint_api_call        = "ListActivities"
  }

  resource_prefix {
    correct = "aws_sfn_"
  }

  provider_package_correct = "sfn"
  doc_prefix               = ["sfn_"]
  brand                    = "AWS"
}

service "shield" {

  sdk {
    id             = "Shield"
    client_version = [2]
  }

  names {
    provider_name_upper = "Shield"
    human_friendly      = "Shield"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListProtectionGroups"
    endpoint_region_override = "us-east-1"
  }

  resource_prefix {
    correct = "aws_shield_"
  }

  provider_package_correct = "shield"
  doc_prefix               = ["shield_"]
  brand                    = "AWS"
}

service "signer" {

  sdk {
    id             = "signer"
    client_version = [2]
  }

  names {
    provider_name_upper = "Signer"
    human_friendly      = "Signer"
  }

  endpoint_info {
    endpoint_api_call        = "ListSigningJobs"
  }

  resource_prefix {
    correct = "aws_signer_"
  }

  provider_package_correct = "signer"
  doc_prefix               = ["signer_"]
  brand                    = "AWS"
}

service "sms" {

  sdk {
    id             = "SMS"
    client_version = [1]
  }

  names {
    provider_name_upper = "SMS"
    human_friendly      = "SMS (Server Migration)"
  }

  client {
    go_v1_client_typename = "SMS"
  }

  resource_prefix {
    correct = "aws_sms_"
  }

  provider_package_correct = "sms"
  doc_prefix               = ["sms_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "snowdevicemanagement" {

  cli_v2_command {
    aws_cli_v2_command           = "snow-device-management"
    aws_cli_v2_command_no_dashes = "snowdevicemanagement"
  }

  sdk {
    id             = "Snow Device Management"
    client_version = [1]
  }

  names {
    provider_name_upper = "SnowDeviceManagement"
    human_friendly      = "Snow Device Management"
  }

  client {
    go_v1_client_typename = "SnowDeviceManagement"
  }

  resource_prefix {
    correct = "aws_snowdevicemanagement_"
  }

  provider_package_correct = "snowdevicemanagement"
  doc_prefix               = ["snowdevicemanagement_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "snowball" {

  sdk {
    id             = "Snowball"
    client_version = [1]
  }

  names {
    provider_name_upper = "Snowball"
    human_friendly      = "Snow Family"
  }

  client {
    go_v1_client_typename = "Snowball"
  }

  resource_prefix {
    correct = "aws_snowball_"
  }

  provider_package_correct = "snowball"
  doc_prefix               = ["snowball_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "sns" {

  sdk {
    id             = "SNS"
    client_version = [2]
  }

  names {
    provider_name_upper = "SNS"
    human_friendly      = "SNS (Simple Notification)"
  }

  endpoint_info {
    endpoint_api_call        = "ListSubscriptions"
  }

  resource_prefix {
    correct = "aws_sns_"
  }

  provider_package_correct = "sns"
  doc_prefix               = ["sns_"]
  brand                    = "AWS"
}

service "sqs" {

  sdk {
    id             = "SQS"
    client_version = [2]
  }

  names {
    provider_name_upper = "SQS"
    human_friendly      = "SQS (Simple Queue)"
  }

  endpoint_info {
    endpoint_api_call        = "ListQueues"
  }

  resource_prefix {
    correct = "aws_sqs_"
  }

  provider_package_correct = "sqs"
  doc_prefix               = ["sqs_"]
  brand                    = "AWS"
}

service "ssm" {

  sdk {
    id             = "SSM"
    client_version = [2]
  }

  names {
    provider_name_upper = "SSM"
    human_friendly      = "SSM (Systems Manager)"
  }

  endpoint_info {
    endpoint_api_call        = "ListDocuments"
  }

  resource_prefix {
    correct = "aws_ssm_"
  }

  provider_package_correct = "ssm"
  doc_prefix               = ["ssm_"]
  brand                    = "AWS"
}

service "ssmcontacts" {

  cli_v2_command {
    aws_cli_v2_command           = "ssm-contacts"
    aws_cli_v2_command_no_dashes = "ssmcontacts"
  }

  sdk {
    id             = "SSM Contacts"
    client_version = [2]
  }

  names {
    provider_name_upper = "SSMContacts"
    human_friendly      = "SSM Contacts"
  }

  endpoint_info {
    endpoint_api_call        = "ListContacts"
  }

  resource_prefix {
    correct = "aws_ssmcontacts_"
  }

  provider_package_correct = "ssmcontacts"
  doc_prefix               = ["ssmcontacts_"]
  brand                    = "AWS"
}

service "ssmincidents" {

  cli_v2_command {
    aws_cli_v2_command           = "ssm-incidents"
    aws_cli_v2_command_no_dashes = "ssmincidents"
  }

  sdk {
    id             = "SSM Incidents"
    client_version = [2]
  }

  names {
    provider_name_upper = "SSMIncidents"
    human_friendly      = "SSM Incident Manager Incidents"
  }

  endpoint_info {
    endpoint_api_call        = "ListResponsePlans"
  }

  resource_prefix {
    correct = "aws_ssmincidents_"
  }

  provider_package_correct = "ssmincidents"
  doc_prefix               = ["ssmincidents_"]
  brand                    = "AWS"
}

service "ssmsap" {

  cli_v2_command {
    aws_cli_v2_command           = "ssm-sap"
    aws_cli_v2_command_no_dashes = "ssmsap"
  }

  sdk {
    id             = "Ssm Sap"
    client_version = [2]
  }

  names {
    provider_name_upper = "SSMSAP"
    human_friendly      = "Systems Manager for SAP"
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
  }

  resource_prefix {
    correct = "aws_ssmsap_"
  }

  provider_package_correct = "ssmsap"
  doc_prefix               = ["ssmsap_"]
  brand                    = "AWS"
}

service "sso" {

  sdk {
    id             = "SSO"
    client_version = [2]
  }

  names {
    provider_name_upper = "SSO"
    human_friendly      = "SSO (Single Sign-On)"
  }

  endpoint_info {
    endpoint_api_call        = "ListAccounts"
    endpoint_api_params      = "AccessToken: aws_sdkv2.String(\"mock-access-token\")"
    endpoint_only            = true
  }

  resource_prefix {
    correct = "aws_sso_"
  }

  provider_package_correct = "sso"
  doc_prefix               = ["sso_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "ssoadmin" {

  cli_v2_command {
    aws_cli_v2_command           = "sso-admin"
    aws_cli_v2_command_no_dashes = "ssoadmin"
  }

  sdk {
    id             = "SSO Admin"
    client_version = [2]
  }

  names {
    provider_name_upper = "SSOAdmin"
    human_friendly      = "SSO Admin"
  }

  client {
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "ListInstances"
  }

  resource_prefix {
    correct = "aws_ssoadmin_"
  }

  provider_package_correct = "ssoadmin"
  doc_prefix               = ["ssoadmin_"]
  brand                    = "AWS"
}

service "identitystore" {

  sdk {
    id             = "identitystore"
    client_version = [2]
  }

  names {
    provider_name_upper = "IdentityStore"
    human_friendly      = "SSO Identity Store"
  }

  endpoint_info {
    endpoint_api_call        = "ListUsers"
    endpoint_api_params      = "IdentityStoreId: aws_sdkv2.String(\"d-1234567890\")"
  }

  resource_prefix {
    correct = "aws_identitystore_"
  }

  provider_package_correct = "identitystore"
  doc_prefix               = ["identitystore_"]
  brand                    = "AWS"
}

service "ssooidc" {

  cli_v2_command {
    aws_cli_v2_command           = "sso-oidc"
    aws_cli_v2_command_no_dashes = "ssooidc"
  }

  sdk {
    id             = "SSO OIDC"
    client_version = [1]
  }

  names {
    provider_name_upper = "SSOOIDC"
    human_friendly      = "SSO OIDC"
  }

  client {
    go_v1_client_typename = "SSOOIDC"
  }

  resource_prefix {
    correct = "aws_ssooidc_"
  }

  provider_package_correct = "ssooidc"
  doc_prefix               = ["ssooidc_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "storagegateway" {

  sdk {
    id             = "Storage Gateway"
    client_version = [1]
  }

  names {
    provider_name_upper = "StorageGateway"
    human_friendly      = "Storage Gateway"
  }

  client {
    go_v1_client_typename = "StorageGateway"
  }

  endpoint_info {
    endpoint_api_call        = "ListGateways"
  }

  resource_prefix {
    correct = "aws_storagegateway_"
  }

  provider_package_correct = "storagegateway"
  doc_prefix               = ["storagegateway_"]
  brand                    = "AWS"
}

service "sts" {

  sdk {
    id             = "STS"
    client_version = [2]
  }

  names {
    provider_name_upper = "STS"
    human_friendly      = "STS (Security Token)"
  }

  client {
    skip_client_generate = true
  }

  env_var {
    deprecated_env_var = "AWS_STS_ENDPOINT"
    tf_aws_env_var     = "TF_AWS_STS_ENDPOINT"
  }

  endpoint_info {
    endpoint_api_call        = "GetCallerIdentity"
  }

  resource_prefix {
    actual  = "aws_caller_identity"
    correct = "aws_sts_"
  }

  provider_package_correct = "sts"
  doc_prefix               = ["caller_identity"]
  brand                    = "AWS"
}

service "support" {

  sdk {
    id             = "Support"
    client_version = [1]
  }

  names {
    provider_name_upper = "Support"
    human_friendly      = "Support"
  }

  client {
    go_v1_client_typename = "Support"
  }

  resource_prefix {
    correct = "aws_support_"
  }

  provider_package_correct = "support"
  doc_prefix               = ["support_"]
  brand                    = "AWS"
  not_implemented          = true
}

service "swf" {

  sdk {
    id             = "SWF"
    client_version = [2]
  }

  names {
    provider_name_upper = "SWF"
    human_friendly      = "SWF (Simple Workflow)"
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
    endpoint_api_params      = "RegistrationStatus: \"REGISTERED\""
  }

  resource_prefix {
    correct = "aws_swf_"
  }

  provider_package_correct = "swf"
  doc_prefix               = ["swf_"]
  brand                    = "AWS"
}

service "textract" {

  sdk {
    id             = "Textract"
    client_version = [1]
  }

  names {
    provider_name_upper = "Textract"
    human_friendly      = "Textract"
  }

  client {
    go_v1_client_typename = "Textract"
  }

  resource_prefix {
    correct = "aws_textract_"
  }

  provider_package_correct = "textract"
  doc_prefix               = ["textract_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "timestreaminfluxdb" {

  cli_v2_command {
    aws_cli_v2_command           = "timestream-influxdb"
    aws_cli_v2_command_no_dashes = "timestreaminfluxdb"
  }

  sdk {
    id             = "Timestream InfluxDB"
    client_version = [2]
  }

  names {
    provider_name_upper = "TimestreamInfluxDB"
    human_friendly      = "Timestream for InfluxDB"
  }

  endpoint_info {
    endpoint_api_call        = "ListDbInstances"
  }

  resource_prefix {
    correct = "aws_timestreaminfluxdb_"
  }

  provider_package_correct = "timestreaminfluxdb"
  doc_prefix               = ["timestreaminfluxdb_"]
  brand                    = "AWS"
}

service "timestreamquery" {

  cli_v2_command {
    aws_cli_v2_command           = "timestream-query"
    aws_cli_v2_command_no_dashes = "timestreamquery"
  }

  sdk {
    id             = "Timestream Query"
    client_version = [1]
  }

  names {
    provider_name_upper = "TimestreamQuery"
    human_friendly      = "Timestream Query"
  }

  client {
    go_v1_client_typename = "TimestreamQuery"
  }

  resource_prefix {
    correct = "aws_timestreamquery_"
  }

  provider_package_correct = "timestreamquery"
  doc_prefix               = ["timestreamquery_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "timestreamwrite" {

  cli_v2_command {
    aws_cli_v2_command           = "timestream-write"
    aws_cli_v2_command_no_dashes = "timestreamwrite"
  }

  sdk {
    id             = "Timestream Write"
    client_version = [2]
  }

  names {
    provider_name_upper = "TimestreamWrite"
    human_friendly      = "Timestream Write"
  }

  endpoint_info {
    endpoint_api_call        = "ListDatabases"
  }

  resource_prefix {
    correct = "aws_timestreamwrite_"
  }

  provider_package_correct = "timestreamwrite"
  doc_prefix               = ["timestreamwrite_"]
  brand                    = "AWS"
}

service "transcribe" {

  go_packages {
    v1_package = "transcribeservice"
    v2_package = "transcribe"
  }

  sdk {
    id             = "Transcribe"
    client_version = [2]
  }

  names {
    aliases             = ["transcribeservice"]
    provider_name_upper = "Transcribe"
    human_friendly      = "Transcribe"
  }

  endpoint_info {
    endpoint_api_call        = "ListLanguageModels"
  }

  resource_prefix {
    correct = "aws_transcribe_"
  }

  provider_package_correct = "transcribe"
  doc_prefix               = ["transcribe_"]
  brand                    = "Amazon"
}

service "transcribestreaming" {

  go_packages {
    v1_package = "transcribestreamingservice"
    v2_package = "transcribestreaming"
  }

  sdk {
    id             = "Transcribe Streaming"
    client_version = [1]
  }

  names {
    aliases             = ["transcribestreamingservice"]
    provider_name_upper = "TranscribeStreaming"
    human_friendly      = "Transcribe Streaming"
  }

  client {
    go_v1_client_typename = "TranscribeStreamingService"
  }

  resource_prefix {
    correct = "aws_transcribestreaming_"
  }

  provider_package_correct = "transcribestreaming"
  doc_prefix               = ["transcribestreaming_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "transfer" {

  sdk {
    id             = "Transfer"
    client_version = [2]
  }

  names {
    provider_name_upper = "Transfer"
    human_friendly      = "Transfer Family"
  }


  endpoint_info {
    endpoint_api_call        = "ListConnectors"
  }

  resource_prefix {
    correct = "aws_transfer_"
  }

  provider_package_correct = "transfer"
  doc_prefix               = ["transfer_"]
  brand                    = "AWS"
}

service "translate" {

  sdk {
    id             = "Translate"
    client_version = [1]
  }

  names {
    provider_name_upper = "Translate"
    human_friendly      = "Translate"
  }

  client {
    go_v1_client_typename = "Translate"
  }

  resource_prefix {
    correct = "aws_translate_"
  }

  provider_package_correct = "translate"
  doc_prefix               = ["translate_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "vpclattice" {

  cli_v2_command {
    aws_cli_v2_command           = "vpc-lattice"
    aws_cli_v2_command_no_dashes = "vpclattice"
  }

  sdk {
    id             = "VPC Lattice"
    client_version = [2]
  }

  names {
    provider_name_upper = "VPCLattice"
    human_friendly      = "VPC Lattice"
  }

  endpoint_info {
    endpoint_api_call        = "ListServices"
  }

  resource_prefix {
    correct = "aws_vpclattice_"
  }

  provider_package_correct = "vpclattice"
  doc_prefix               = ["vpclattice_"]
  brand                    = "AWS"
}

service "wafv2" {

  sdk {
    id             = "WAFV2"
    client_version = [2]
  }

  names {
    provider_name_upper = "WAFV2"
    human_friendly      = "WAF"
  }

  endpoint_info {
    endpoint_api_call        = "ListRuleGroups"
    endpoint_api_params      = "Scope: awstypes.ScopeRegional"
  }

  resource_prefix {
    correct = "aws_wafv2_"
  }

  provider_package_correct = "wafv2"
  doc_prefix               = ["wafv2_"]
  brand                    = "AWS"
}

service "waf" {

  sdk {
    id             = "WAF"
    client_version = [2]
  }

  names {
    provider_name_upper = "WAF"
    human_friendly      = "WAF Classic"
  }

  endpoint_info {
    endpoint_api_call        = "ListRules"
  }

  resource_prefix {
    correct = "aws_waf_"
  }

  provider_package_correct = "waf"
  doc_prefix               = ["waf_"]
  brand                    = "AWS"
}

service "wafregional" {

  cli_v2_command {
    aws_cli_v2_command           = "waf-regional"
    aws_cli_v2_command_no_dashes = "wafregional"
  }

  sdk {
    id             = "WAF Regional"
    client_version = [2]
  }

  names {
    provider_name_upper = "WAFRegional"
    human_friendly      = "WAF Classic Regional"
  }

  endpoint_info {
    endpoint_api_call        = "ListRules"
  }

  resource_prefix {
    correct = "aws_wafregional_"
  }

  provider_package_correct = "wafregional"
  doc_prefix               = ["wafregional_"]
  brand                    = "AWS"
}

service "budgets" {

  sdk {
    id             = "Budgets"
    client_version = [2]
  }

  names {
    provider_name_upper = "Budgets"
    human_friendly      = "Web Services Budgets"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeBudgets"
    endpoint_api_params      = "AccountId: aws_sdkv2.String(\"012345678901\")"
  }

  resource_prefix {
    correct = "aws_budgets_"
  }

  provider_package_correct = "budgets"
  doc_prefix               = ["budgets_"]
  brand                    = "AWS"
}

service "wellarchitected" {

  sdk {
    id             = "WellArchitected"
    client_version = [2]
  }

  names {
    provider_name_upper = "WellArchitected"
    human_friendly      = "Well-Architected Tool"
  }

  endpoint_info {
    endpoint_api_call        = "ListProfiles"
  }

  resource_prefix {
    correct = "aws_wellarchitected_"
  }

  provider_package_correct = "wellarchitected"
  doc_prefix               = ["wellarchitected_"]
  brand                    = "AWS"
}

service "workdocs" {

  sdk {
    id             = "WorkDocs"
    client_version = [1]
  }

  names {
    provider_name_upper = "WorkDocs"
    human_friendly      = "WorkDocs"
  }

  client {
    go_v1_client_typename = "WorkDocs"
  }

  resource_prefix {
    correct = "aws_workdocs_"
  }

  provider_package_correct = "workdocs"
  doc_prefix               = ["workdocs_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "worklink" {

  sdk {
    id             = "WorkLink"
    client_version = [1]
  }

  names {
    provider_name_upper = "WorkLink"
    human_friendly      = "WorkLink"
  }

  client {
    go_v1_client_typename = "WorkLink"
  }

  endpoint_info {
    endpoint_api_call        = "ListFleets"
  }

  resource_prefix {
    correct = "aws_worklink_"
  }

  provider_package_correct = "worklink"
  doc_prefix               = ["worklink_"]
  brand                    = "AWS"
}

service "workmail" {

  sdk {
    id             = "WorkMail"
    client_version = [1]
  }

  names {
    provider_name_upper = "WorkMail"
    human_friendly      = "WorkMail"
  }

  client {
    go_v1_client_typename = "WorkMail"
  }

  resource_prefix {
    correct = "aws_workmail_"
  }

  provider_package_correct = "workmail"
  doc_prefix               = ["workmail_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "workmailmessageflow" {

  sdk {
    id             = "WorkMailMessageFlow"
    client_version = [1]
  }

  names {
    provider_name_upper = "WorkMailMessageFlow"
    human_friendly      = "WorkMail Message Flow"
  }

  client {
    go_v1_client_typename = "WorkMailMessageFlow"
  }

  resource_prefix {
    correct = "aws_workmailmessageflow_"
  }

  provider_package_correct = "workmailmessageflow"
  doc_prefix               = ["workmailmessageflow_"]
  brand                    = "Amazon"
  not_implemented          = true
}

service "workspaces" {

  sdk {
    id             = "WorkSpaces"
    client_version = [2]
  }

  names {
    provider_name_upper = "WorkSpaces"
    human_friendly      = "WorkSpaces"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeWorkspaces"
  }

  resource_prefix {
    correct = "aws_workspaces_"
  }

  provider_package_correct = "workspaces"
  doc_prefix               = ["workspaces_"]
  brand                    = "AWS"
}

service "workspacesweb" {

  cli_v2_command {
    aws_cli_v2_command           = "workspaces-web"
    aws_cli_v2_command_no_dashes = "workspacesweb"
  }

  sdk {
    id             = "WorkSpaces Web"
    client_version = [2]
  }

  names {
    provider_name_upper = "WorkSpacesWeb"
    human_friendly      = "WorkSpaces Web"
  }

  endpoint_info {
    endpoint_api_call        = "ListPortals"
  }

  resource_prefix {
    correct = "aws_workspacesweb_"
  }

  provider_package_correct = "workspacesweb"
  doc_prefix               = ["workspacesweb_"]
  brand                    = "AWS"
}

service "xray" {

  sdk {
    id             = "XRay"
    client_version = [2]
  }

  names {
    provider_name_upper = "XRay"
    human_friendly      = "X-Ray"
  }

  endpoint_info {
    endpoint_api_call        = "ListResourcePolicies"
  }

  resource_prefix {
    correct = "aws_xray_"
  }

  provider_package_correct = "xray"
  doc_prefix               = ["xray_"]
  brand                    = "AWS"
}

service "verifiedpermissions" {

  sdk {
    id             = "VerifiedPermissions"
    client_version = [2]
  }

  names {
    provider_name_upper = "VerifiedPermissions"
    human_friendly      = "Verified Permissions"
  }

  endpoint_info {
    endpoint_api_call        = "ListPolicyStores"
  }

  resource_prefix {
    correct = "aws_verifiedpermissions_"
  }

  provider_package_correct = "verifiedpermissions"
  doc_prefix               = ["verifiedpermissions_"]
  brand                    = "AWS"
}

service "codecatalyst" {

  sdk {
    id             = "CodeCatalyst"
    client_version = [2]
  }

  names {
    provider_name_upper = "CodeCatalyst"
    human_friendly      = "CodeCatalyst"
  }

  endpoint_info {
    endpoint_api_call        = "ListAccessTokens"
  }

  resource_prefix {
    correct = "aws_codecatalyst_"
  }

  provider_package_correct = "codecatalyst"
  doc_prefix               = ["codecatalyst_"]
  brand                    = "AWS"
}

service "mediapackagev2" {

  sdk {
    id             = "MediaPackageV2"
    client_version = [2]
  }

  names {
    provider_name_upper = "MediaPackageV2"
    human_friendly      = "Elemental MediaPackage Version 2"
  }

  endpoint_info {
    endpoint_api_call        = "ListChannelGroups"
  }

  resource_prefix {
    actual  = "aws_media_packagev2_"
    correct = "aws_mediapackagev2_"
  }

  provider_package_correct = "mediapackagev2"
  doc_prefix               = ["media_packagev2_"]
  brand                    = "AWS"
}

service "iot" {

  sdk {
    id             = "IoT"
    client_version = [2]
  }

  names {
    provider_name_upper = "IoT"
    human_friendly      = "IoT Core"
  }

  client {
    go_v1_client_typename = "IoT"
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDefaultAuthorizer"
  }

  resource_prefix {
    correct = "aws_iot_"
  }

  provider_package_correct = "iot"
  doc_prefix               = ["iot_"]
  brand                    = "AWS"
}

service "dynamodb" {

  sdk {
    id             = "DynamoDB"
    client_version = [2]
  }

  names {
    provider_name_upper = "DynamoDB"
    human_friendly      = "DynamoDB"
  }

  client {
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = "AWS_DYNAMODB_ENDPOINT"
    tf_aws_env_var     = "TF_AWS_DYNAMODB_ENDPOINT"
  }
  endpoint_info {
    endpoint_api_call        = "ListTables"
  }

  resource_prefix {
    correct = "aws_dynamodb_"
  }

  provider_package_correct = "dynamodb"
  doc_prefix               = ["dynamodb_"]
  brand                    = "AWS"
}

service "ec2" {

  sdk {
    id             = "EC2"
    client_version = [2]
  }

  names {
    provider_name_upper = "EC2"
    human_friendly      = "EC2 (Elastic Compute Cloud)"
  }

  client {
    go_v1_client_typename = "EC2"
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "DescribeVpcs"
  }

  resource_prefix {
    actual  = "aws_(ami|availability_zone|ec2_(availability|capacity|fleet|host|instance|public_ipv4_pool|serial|spot|tag)|eip|instance|key_pair|launch_template|placement_group|spot)"
    correct = "aws_ec2_"
  }

  sub_service "ec2ebs" {

    cli_v2_command {
          aws_cli_v2_command                    = ""
          aws_cli_v2_command_no_dashes          = ""
    }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "EC2EBS"
      human_friendly      = "EBS (EC2)"
    }

    resource_prefix {
      actual  = "aws_(ebs_|volume_attach|snapshot_create)"
      correct = "aws_ec2ebs_"
    }

    split_package            = "ec2"
    file_prefix              = "ebs_"
    doc_prefix               = ["ebs_", "volume_attachment", "snapshot_"]
    brand                    = "Amazon"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "ec2outposts" {

    cli_v2_command {
          aws_cli_v2_command           = ""
          aws_cli_v2_command_no_dashes          = ""
      }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
      }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "EC2Outposts"
      human_friendly      = "Outposts (EC2)"
    }

    resource_prefix {
      actual  = "aws_ec2_(coip_pool|local_gateway)"
      correct = "aws_ec2outposts_"
    }

    split_package            = "ec2"
    file_prefix              = "outposts_"
    doc_prefix               = ["ec2_coip_pool", "ec2_local_gateway"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "transitgateway" {

  cli_v2_command {
        aws_cli_v2_command           = ""
        aws_cli_v2_command_no_dashes           = ""
    }

  go_packages {
        v1_package                   = ""
        v2_package                   = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "TransitGateway"
      human_friendly      = "Transit Gateway"
    }

    resource_prefix {
      actual  = "aws_ec2_transit_gateway"
      correct = "aws_transitgateway_"
    }

    split_package            = "ec2"
    file_prefix              = "transitgateway_"
    doc_prefix               = ["ec2_transit_gateway"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "verifiedaccess" {

    cli_v2_command {
          aws_cli_v2_command           = ""
          aws_cli_v2_command_no_dashes           = ""
      }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
      }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "VerifiedAccess"
      human_friendly      = "Verified Access"
    }

    resource_prefix {
      actual  = "aws_verifiedaccess"
      correct = "aws_verifiedaccess_"
    }
    
    split_package            = "ec2"
    file_prefix              = "verifiedaccess_"
    doc_prefix               = ["verifiedaccess_"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "vpc" {

    cli_v2_command {
          aws_cli_v2_command           = ""
          aws_cli_v2_command_no_dashes          = ""
      }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
      }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "VPC"
      human_friendly      = "VPC (Virtual Private Cloud)"
    }

    resource_prefix {
      actual  = "aws_((default_)?(network_acl|route_table|security_group|subnet|vpc(?!_ipam))|ec2_(managed|network|subnet|traffic)|egress_only_internet|flow_log|internet_gateway|main_route_table_association|nat_gateway|network_interface|prefix_list|route\\b)"
      correct = "aws_vpc_"
    }

    split_package            = "ec2"
    file_prefix              = "vpc_"
    doc_prefix               = ["default_network_", "default_route_", "default_security_", "default_subnet", "default_vpc", "ec2_managed_", "ec2_network_", "ec2_subnet_", "ec2_traffic_", "egress_only_", "flow_log", "internet_gateway", "main_route_", "nat_", "network_", "prefix_list", "route_", "route\\.", "security_group", "subnet", "vpc_dhcp_", "vpc_endpoint", "vpc_ipv", "vpc_network_performance", "vpc_peering_", "vpc_security_group_", "vpc\\.", "vpcs\\."]
    brand                    = "Amazon"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "ipam" {

    cli_v2_command {
          aws_cli_v2_command                    = ""
          aws_cli_v2_command_no_dashes          = ""
    }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
    }
      
    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "IPAM"
      human_friendly      = "VPC IPAM (IP Address Manager)"
    }

    resource_prefix {
      actual  = "aws_vpc_ipam"
      correct = "aws_ipam_"
    }
    split_package            = "ec2"
    file_prefix              = "ipam_"
    doc_prefix               = ["vpc_ipam"]
    brand                    = "Amazon"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "vpnclient" {

    cli_v2_command {
          aws_cli_v2_command                    = ""
          aws_cli_v2_command_no_dashes          = ""
    }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "ClientVPN"
      human_friendly      = "VPN (Client)"
    }

    resource_prefix {
      actual  = "aws_ec2_client_vpn"
      correct = "aws_vpnclient_"
    }
    split_package            = "ec2"
    file_prefix              = "vpnclient_"
    doc_prefix               = ["ec2_client_vpn_"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "vpnsite" {

    cli_v2_command {
          aws_cli_v2_command           = ""
          aws_cli_v2_command_no_dashes          = ""
    }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "SiteVPN"
      human_friendly      = "VPN (Site-to-Site)"
    }

    resource_prefix {
      actual  = "aws_(customer_gateway|vpn_)"
      correct = "aws_vpnsite_"
    }

    split_package            = "ec2"
    file_prefix              = "vpnsite_"
    doc_prefix               = ["customer_gateway", "vpn_"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "wavelength" {

cli_v2_command {
       aws_cli_v2_command                     = ""
       aws_cli_v2_command_no_dashes           = ""
   }

go_packages {
       v1_package                   = ""
       v2_package                   = ""
   }
    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "Wavelength"
      human_friendly      = "Wavelength"
    }

    resource_prefix {
      actual  = "aws_ec2_carrier_gateway"
      correct = "aws_wavelength_"
    }

    split_package            = "ec2"
    file_prefix              = "wavelength_"
    doc_prefix               = ["ec2_carrier_"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  provider_package_correct = "ec2"
  split_package            = "ec2"
  file_prefix              = "ec2_"
  doc_prefix               = ["ami", "availability_zone", "ec2_availability_", "ec2_capacity_", "ec2_fleet", "ec2_host", "ec2_image_", "ec2_instance_", "ec2_public_ipv4_pool", "ec2_serial_", "ec2_spot_", "ec2_tag", "eip", "instance", "key_pair", "launch_template", "placement_group", "spot_"]
  brand                    = "Amazon"
}
