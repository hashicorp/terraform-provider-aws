
service "accessanalyzer" {

  sdk {
    id             = "AccessAnalyzer"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AccessAnalyzer"
    human_friendly      = "IAM Access Analyzer"
  }

  client {
    go_v1_client_typename = "AccessAnalyzer"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAnalyzers"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_accessanalyzer_"
  }
  	
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["accessanalyzer_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "account" {

  sdk {
    id             = "Account"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Account"
    human_friendly      = "Account Management"
  }

  client {
    go_v1_client_typename = "Account"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRegions"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_account_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["account_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "acm" {

  sdk {
    id             = "ACM"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ACM"
    human_friendly      = "ACM (Certificate Manager)"
  }

  client {
    go_v1_client_typename = "ACM"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListCertificates"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_acm_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["acm_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ACMPCA"
    human_friendly      = "ACM PCA (Certificate Manager Private Certificate Authority)"
  }

  client {
    go_v1_client_typename = "ACMPCA"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListCertificateAuthorities"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_acmpca_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["acmpca_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "alexaforbusiness" {

  sdk {
    id             = "Alexa For Business"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AlexaForBusiness"
    human_friendly      = "Alexa for Business"
  }

  client {
    go_v1_client_typename = "AlexaForBusiness"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_alexaforbusiness_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["alexaforbusiness_"]
  brand               = ""
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "PrometheusService"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListScrapers"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_prometheus_"
    correct = "aws_amp_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["prometheus_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "amplify" {

  sdk {
    id             = "Amplify"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Amplify"
    human_friendly      = "Amplify"
  }

  client {
    go_v1_client_typename = "Amplify"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApps"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_amplify_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["amplify_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "amplifybackend" {

  sdk {
    id             = "AmplifyBackend"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AmplifyBackend"
    human_friendly      = "Amplify Backend"
  }

  client {
    go_v1_client_typename = "AmplifyBackend"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_amplifybackend_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["amplifybackend_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "amplifyuibuilder" {

  sdk {
    id             = "AmplifyUIBuilder"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AmplifyUIBuilder"
    human_friendly      = "Amplify UI Builder"
  }

  client {
    go_v1_client_typename = "AmplifyUIBuilder"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_amplifyuibuilder_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["amplifyuibuilder_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "apigateway" {

  sdk {
    id             = "API Gateway"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "APIGateway"
    human_friendly      = "API Gateway"
  }

  client {
    go_v1_client_typename = "APIGateway"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetAccount"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_api_gateway_"
    correct = "aws_apigateway_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["api_gateway_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "apigatewaymanagementapi" {

  sdk {
    id             = "ApiGatewayManagementApi"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "APIGatewayManagementAPI"
    human_friendly      = "API Gateway Management API"
  }

  client {
    go_v1_client_typename = "ApiGatewayManagementApi"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_apigatewaymanagementapi_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["apigatewaymanagementapi_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "apigatewayv2" {

  sdk {
    id             = "ApiGatewayV2"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "APIGatewayV2"
    human_friendly      = "API Gateway V2"
  }

  client {
    go_v1_client_typename = "ApiGatewayV2"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetApis"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_apigatewayv2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["apigatewayv2_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "appfabric" {

  sdk {
    id             = "AppFabric"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AppFabric"
    human_friendly      = "AppFabric"
  }

  client {
    go_v1_client_typename = "AppFabric"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAppBundles"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_appfabric_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["appfabric_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "appmesh" {

  sdk {
    id             = "App Mesh"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AppMesh"
    human_friendly      = "App Mesh"
  }

  client {
    go_v1_client_typename = "AppMesh"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListMeshes"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_appmesh_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["appmesh_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "apprunner" {

  sdk {
    id             = "AppRunner"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AppRunner"
    human_friendly      = "App Runner"
  }

  client {
    go_v1_client_typename = "AppRunner"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListConnections"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_apprunner_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["apprunner_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "appconfig" {

  sdk {
    id             = "AppConfig"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AppConfig"
    human_friendly      = "AppConfig"
  }

  client {
    go_v1_client_typename = "AppConfig"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_appconfig_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["appconfig_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "appconfigdata" {

  sdk {
    id             = "AppConfigData"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AppConfigData"
    human_friendly      = "AppConfig Data"
  }

  client {
    go_v1_client_typename = "AppConfigData"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_appconfigdata_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["appconfigdata_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "appflow" {

  sdk {
    id             = "Appflow"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AppFlow"
    human_friendly      = "AppFlow"
  }

  client {
    go_v1_client_typename = "Appflow"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListFlows"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_appflow_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["appflow_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "AppIntegrationsService"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_appintegrations_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["appintegrations_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "ApplicationAutoScaling"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeScalableTargets"
    endpoint_api_params      = "ServiceNamespace: awstypes.ServiceNamespaceEcs"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_appautoscaling_"
    correct = "aws_applicationautoscaling_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["appautoscaling_"]
  brand               = ""
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "applicationcostprofiler" {

  sdk {
    id             = "ApplicationCostProfiler"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ApplicationCostProfiler"
    human_friendly      = "Application Cost Profiler"
  }

  client {
    go_v1_client_typename = "ApplicationCostProfiler"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_applicationcostprofiler_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["applicationcostprofiler_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_discovery_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["discovery_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "mgn" {

  sdk {
    id             = "mgn"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Mgn"
    human_friendly      = "Application Migration (Mgn)"
  }

  client {
    go_v1_client_typename = "Mgn"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mgn_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mgn_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "appstream" {

  sdk {
    id             = "AppStream"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AppStream"
    human_friendly      = "AppStream 2.0"
  }

  client {
    go_v1_client_typename = "AppStream"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAssociatedFleets"
    endpoint_api_params      = "StackName: aws_sdkv2.String(\"test\")"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_appstream_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["appstream_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "appsync" {

  sdk {
    id             = "AppSync"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AppSync"
    human_friendly      = "AppSync"
  }

  client {
    go_v1_client_typename = "AppSync"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomainNames"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_appsync_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["appsync_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "athena" {

  sdk {
    id             = "Athena"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Athena"
    human_friendly      = "Athena"
  }

  client {
    go_v1_client_typename = "Athena"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDataCatalogs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_athena_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["athena_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "auditmanager" {

  sdk {
    id             = "AuditManager"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AuditManager"
    human_friendly      = "Audit Manager"
  }

  client {
    go_v1_client_typename = "AuditManager"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetAccountStatus"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_auditmanager_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["auditmanager_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "autoscaling" {

  sdk {
    id             = "Auto Scaling"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "AutoScaling"
    human_friendly      = "Auto Scaling"
  }

  client {
    go_v1_client_typename = "AutoScaling"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeAutoScalingGroups"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_(autoscaling_|launch_configuration)"
    correct = "aws_autoscaling_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["autoscaling_", "launch_configuration"]
  brand               = ""
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "AutoScalingPlans"
    human_friendly      = "Auto Scaling Plans"
  }

  client {
    go_v1_client_typename = "AutoScalingPlans"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeScalingPlans"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_autoscalingplans_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["autoscalingplans_"]
  brand               = ""
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "backup" {

  sdk {
    id             = "Backup"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Backup"
    human_friendly      = "Backup"
  }

  client {
    go_v1_client_typename = "Backup"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListBackupPlans"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_backup_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["backup_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "BackupGateway"
    human_friendly      = "Backup Gateway"
  }

  client {
    go_v1_client_typename = "BackupGateway"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_backupgateway_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["backupgateway_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "batch" {

  sdk {
    id             = "Batch"
    client_version = [1, 2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Batch"
    human_friendly      = "Batch"
  }

  client {
    go_v1_client_typename = "Batch"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListJobs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_batch_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["batch_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "bedrock" {

  sdk {
    id             = "Bedrock"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Bedrock"
    human_friendly      = "Amazon Bedrock"
  }

  client {
    go_v1_client_typename = "Bedrock"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListFoundationModels"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_bedrock_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["bedrock_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "BedrockAgent"
    human_friendly      = "Agents for Amazon Bedrock"
  }

  client {
    go_v1_client_typename = "BedrockAgent"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAgents"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_bedrockagent_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["bedrockagent_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "bcmdataexports" {

  sdk {
    id             = "BCM Data Exports"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "BCMDataExports"
    human_friendly      = "BCM Data Exports"
  }

  client {
    go_v1_client_typename = "BCMDataExports"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListExports"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_bcmdataexports_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["bcmdataexports_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "BillingConductor"
    human_friendly      = "Billing Conductor"
  }

  client {
    go_v1_client_typename = "BillingConductor"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_billingconductor_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["billingconductor_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "braket" {

  sdk {
    id             = "Braket"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Braket"
    human_friendly      = "Braket"
  }

  client {
    go_v1_client_typename = "Braket"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_braket_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["braket_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "CostExplorer"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListCostCategoryDefinitions"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ce_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ce_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "chatbot" {

  sdk {
    id             = "Chatbot"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Chatbot"
    human_friendly      = "Chatbot"
  }

  client {
    go_v1_client_typename = ""
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetAccountPreferences"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_chatbot_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["chatbot_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "chime" {

  sdk {
    id             = "Chime"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Chime"
    human_friendly      = "Chime"
  }

  client {
    go_v1_client_typename = "Chime"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAccounts"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_chime_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["chime_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ChimeSDKIdentity"
    human_friendly      = "Chime SDK Identity"
  }

  client {
    go_v1_client_typename = "ChimeSDKIdentity"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_chimesdkidentity_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["chimesdkidentity_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ChimeSDKMediaPipelines"
    human_friendly      = "Chime SDK Media Pipelines"
  }

  client {
    go_v1_client_typename = "ChimeSDKMediaPipelines"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListMediaPipelines"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_chimesdkmediapipelines_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["chimesdkmediapipelines_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ChimeSDKMeetings"
    human_friendly      = "Chime SDK Meetings"
  }

  client {
    go_v1_client_typename = "ChimeSDKMeetings"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_chimesdkmeetings_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["chimesdkmeetings_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ChimeSDKMessaging"
    human_friendly      = "Chime SDK Messaging"
  }

  client {
    go_v1_client_typename = "ChimeSDKMessaging"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_chimesdkmessaging_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["chimesdkmessaging_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ChimeSDKVoice"
    human_friendly      = "Chime SDK Voice"
  }

  client {
    go_v1_client_typename = "ChimeSDKVoice"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPhoneNumbers"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_chimesdkvoice_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["chimesdkvoice_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "cleanrooms" {

  sdk {
    id             = "CleanRooms"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CleanRooms"
    human_friendly      = "Clean Rooms"
  }

  client {
    go_v1_client_typename = "CleanRooms"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListCollaborations"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_cleanrooms_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cleanrooms_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListResourceRequests"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_cloudcontrolapi_"
    correct = "aws_cloudcontrol_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudcontrolapi_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "clouddirectory" {

  sdk {
    id             = "CloudDirectory"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CloudDirectory"
    human_friendly      = "Cloud Directory"
  }

  client {
    go_v1_client_typename = "CloudDirectory"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_clouddirectory_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["clouddirectory_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "servicediscovery" {

  sdk {
    id             = "ServiceDiscovery"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ServiceDiscovery"
    human_friendly      = "Cloud Map"
  }

  client {
    go_v1_client_typename = "ServiceDiscovery"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListNamespaces"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_service_discovery_"
    correct = "aws_servicediscovery_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["service_discovery_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "cloud9" {

  sdk {
    id             = "Cloud9"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Cloud9"
    human_friendly      = "Cloud9"
  }

  client {
    go_v1_client_typename = "Cloud9"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListEnvironments"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_cloud9_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloud9_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "cloudformation" {

  sdk {
    id             = "CloudFormation"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CloudFormation"
    human_friendly      = "CloudFormation"
  }

  client {
    go_v1_client_typename = "CloudFormation"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListStackInstances"
    endpoint_api_params      = "StackSetName: aws_sdkv2.String(\"test\")"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_cloudformation_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudformation_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "cloudfront" {

  sdk {
    id             = "CloudFront"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CloudFront"
    human_friendly      = "CloudFront"
  }

  client {
    go_v1_client_typename = "CloudFront"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDistributions"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_cloudfront_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudfront_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "CloudFrontKeyValueStore"
    human_friendly      = "CloudFront KeyValueStore"
  }

  client {
    go_v1_client_typename = "CloudFrontKeyValueStore"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListKeys"
    endpoint_api_params      = "KvsARN: aws_sdkv2.String(\"arn:aws:cloudfront::111122223333:key-value-store/MaxAge\")"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_cloudfrontkeyvaluestore_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudfrontkeyvaluestore_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    go_v1_client_typename = "CloudHSMV2"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_cloudhsm_v2_"
    correct = "aws_cloudhsmv2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudhsm"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "cloudsearch" {

  sdk {
    id             = "CloudSearch"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CloudSearch"
    human_friendly      = "CloudSearch"
  }

  client {
    go_v1_client_typename = "CloudSearch"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomainNames"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_cloudsearch_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudsearch_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "cloudsearchdomain" {

  sdk {
    id             = "CloudSearch Domain"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CloudSearchDomain"
    human_friendly      = "CloudSearch Domain"
  }

  client {
    go_v1_client_typename = "CloudSearchDomain"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_cloudsearchdomain_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudsearchdomain_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "cloudtrail" {

  sdk {
    id             = "CloudTrail"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CloudTrail"
    human_friendly      = "CloudTrail"
  }

  client {
    go_v1_client_typename = "CloudTrail"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListChannels"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_cloudtrail"
    correct = "aws_cloudtrail_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudtrail"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "cloudwatch" {

  sdk {
    id             = "CloudWatch"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CloudWatch"
    human_friendly      = "CloudWatch"
  }

  client {
    go_v1_client_typename = "CloudWatch"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDashboards"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_cloudwatch_(?!(event_|log_|query_))"
    correct = "aws_cloudwatch_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudwatch_dashboard", "cloudwatch_metric_", "cloudwatch_composite_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "applicationinsights" {

  cli_v2_command {
    aws_cli_v2_command           = "application-insights"
    aws_cli_v2_command_no_dashes = "applicationinsights"
  }

  sdk {
    id             = "Application Insights"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ApplicationInsights"
    human_friendly      = "CloudWatch Application Insights"
  }

  client {
    go_v1_client_typename = "ApplicationInsights"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "CreateApplication"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_applicationinsights_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["applicationinsights_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "CloudWatchEvidently"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListProjects"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_evidently_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["evidently_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "internetmonitor" {

  sdk {
    id             = "InternetMonitor"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "InternetMonitor"
    human_friendly      = "CloudWatch Internet Monitor"
  }

  client {
    go_v1_client_typename = "InternetMonitor"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListMonitors"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_internetmonitor_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["internetmonitor_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAnomalies"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_cloudwatch_(log_|query_)"
    correct = "aws_logs_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudwatch_log_", "cloudwatch_query_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "rum" {

  go_packages {
    v1_package = "cloudwatchrum"
    v2_package = "rum"
  }

  sdk {
    id             = "RUM"
    client_version = [1]
  }

  names {
    aliases             = ["cloudwatchrum"]
    provider_name_upper = "RUM"
    human_friendly      = "CloudWatch RUM"
  }

  client {
    go_v1_client_typename = "CloudWatchRUM"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAppMonitors"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_rum_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["rum_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "synthetics" {

  sdk {
    id             = "synthetics"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Synthetics"
    human_friendly      = "CloudWatch Synthetics"
  }

  client {
    go_v1_client_typename = "Synthetics"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListGroups"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_synthetics_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["synthetics_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "codeartifact" {

  sdk {
    id             = "codeartifact"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CodeArtifact"
    human_friendly      = "CodeArtifact"
  }

  client {
    go_v1_client_typename = "CodeArtifact"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_codeartifact_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codeartifact_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "codebuild" {

  sdk {
    id             = "CodeBuild"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CodeBuild"
    human_friendly      = "CodeBuild"
  }

  client {
    go_v1_client_typename = "CodeBuild"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListBuildBatches"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_codebuild_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codebuild_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "codecommit" {

  sdk {
    id             = "CodeCommit"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CodeCommit"
    human_friendly      = "CodeCommit"
  }

  client {
    go_v1_client_typename = "CodeCommit"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRepositories"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_codecommit_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codecommit_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "CodeDeploy"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_codedeploy_"
    correct = "aws_deploy_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codedeploy_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "codeguruprofiler" {

  sdk {
    id             = "CodeGuruProfiler"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CodeGuruProfiler"
    human_friendly      = "CodeGuru Profiler"
  }

  client {
    go_v1_client_typename = "CodeGuruProfiler"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListProfilingGroups"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_codeguruprofiler_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codeguruprofiler_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "CodeGuruReviewer"
    human_friendly      = "CodeGuru Reviewer"
  }

  client {
    go_v1_client_typename = "CodeGuruReviewer"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListCodeReviews"
    endpoint_api_params      = "Type: awstypes.TypePullRequest"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_codegurureviewer_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codegurureviewer_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "codepipeline" {

  sdk {
    id             = "CodePipeline"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CodePipeline"
    human_friendly      = "CodePipeline"
  }

  client {
    go_v1_client_typename = "CodePipeline"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPipelines"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_codepipeline"
    correct = "aws_codepipeline_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codepipeline"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "codestar" {

  sdk {
    id             = "CodeStar"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CodeStar"
    human_friendly      = "CodeStar"
  }

  client {
    go_v1_client_typename = "CodeStar"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_codestar_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codestar_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "CodeStarConnections"
    human_friendly      = "CodeStar Connections"
  }

  client {
    go_v1_client_typename = "CodeStarConnections"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListConnections"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_codestarconnections_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codestarconnections_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "CodeStarNotifications"
    human_friendly      = "CodeStar Notifications"
  }

  client {
    go_v1_client_typename = "CodeStarNotifications"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListTargets"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_codestarnotifications_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codestarnotifications_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "CognitoIdentity"
    human_friendly      = "Cognito Identity"
  }

  client {
    go_v1_client_typename = "CognitoIdentity"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListIdentityPools"
    endpoint_api_params      = "MaxResults: aws_sdkv2.Int32(1)"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_cognito_identity_(?!provider)"
    correct = "aws_cognitoidentity_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cognito_identity_pool"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    client_version = [1]
  }

  names {
    aliases             = ["cognitoidentityprovider"]
    provider_name_upper = "CognitoIDP"
    human_friendly      = "Cognito IDP (Identity Provider)"
  }

  client {
    go_v1_client_typename = "CognitoIdentityProvider"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListUserPools"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_cognito_(identity_provider|resource|user|risk)"
    correct = "aws_cognitoidp_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cognito_identity_provider", "cognito_managed_user", "cognito_resource_", "cognito_user", "cognito_risk"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "CognitoSync"
    human_friendly      = "Cognito Sync"
  }

  client {
    go_v1_client_typename = "CognitoSync"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_cognitosync_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cognitosync_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "comprehend" {

  sdk {
    id             = "Comprehend"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Comprehend"
    human_friendly      = "Comprehend"
  }

  client {
    go_v1_client_typename = "Comprehend"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDocumentClassifiers"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_comprehend_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["comprehend_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "comprehendmedical" {

  sdk {
    id             = "ComprehendMedical"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ComprehendMedical"
    human_friendly      = "Comprehend Medical"
  }

  client {
    go_v1_client_typename = "ComprehendMedical"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_comprehendmedical_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["comprehendmedical_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ComputeOptimizer"
    human_friendly      = "Compute Optimizer"
  }

  client {
    go_v1_client_typename = "ComputeOptimizer"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetEnrollmentStatus"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_computeoptimizer_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["computeoptimizer_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "ConfigService"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListStoredQueries"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_config_"
    correct = "aws_configservice_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["config_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "connect" {

  sdk {
    id             = "Connect"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Connect"
    human_friendly      = "Connect"
  }

  client {
    go_v1_client_typename = "Connect"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListInstances"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_connect_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["connect_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "connectcases" {

  sdk {
    id             = "ConnectCases"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ConnectCases"
    human_friendly      = "Connect Cases"
  }

  client {
    go_v1_client_typename = "ConnectCases"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_connectcases_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["connectcases_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ConnectContactLens"
    human_friendly      = "Connect Contact Lens"
  }

  client {
    go_v1_client_typename = "ConnectContactLens"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_connectcontactlens_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["connectcontactlens_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "CustomerProfiles"
    human_friendly      = "Connect Customer Profiles"
  }

  client {
    go_v1_client_typename = "CustomerProfiles"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_customerprofiles_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["customerprofiles_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "connectparticipant" {

  sdk {
    id             = "ConnectParticipant"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ConnectParticipant"
    human_friendly      = "Connect Participant"
  }

  client {
    go_v1_client_typename = "ConnectParticipant"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_connectparticipant_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["connectparticipant_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "VoiceID"
    human_friendly      = "Connect Voice ID"
  }

  client {
    go_v1_client_typename = "VoiceID"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_voiceid_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["voiceid_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_wisdom_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["wisdom_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "controltower" {

  sdk {
    id             = "ControlTower"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ControlTower"
    human_friendly      = "Control Tower"
  }

  client {
    go_v1_client_typename = "ControlTower"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListLandingZones"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_controltower_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["controltower_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "CostOptimizationHub"
    human_friendly      = "Cost Optimization Hub"
  }

  client {
    go_v1_client_typename = "CostOptimizationHub"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetPreferences"
    endpoint_api_params      = ""
    endpoint_region_override = "us-east-1"
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_costoptimizationhub_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["costoptimizationhub_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    go_v1_client_typename = "CostandUsageReportService"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeReportDefinitions"
    endpoint_api_params      = ""
    endpoint_region_override = "us-east-1"
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_cur_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cur_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "dataexchange" {

  sdk {
    id             = "DataExchange"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DataExchange"
    human_friendly      = "Data Exchange"
  }

  client {
    go_v1_client_typename = "DataExchange"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDataSets"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_dataexchange_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["dataexchange_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "datapipeline" {

  sdk {
    id             = "Data Pipeline"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DataPipeline"
    human_friendly      = "Data Pipeline"
  }

  client {
    go_v1_client_typename = "DataPipeline"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPipelines"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_datapipeline_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["datapipeline_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "datasync" {

  sdk {
    id             = "DataSync"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DataSync"
    human_friendly      = "DataSync"
  }

  client {
    go_v1_client_typename = "DataSync"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAgents"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_datasync_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["datasync_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "datazone" {

  sdk {
    id             = "DataZone"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DataZone"
    human_friendly      = "DataZone"
  }

  client {
    go_v1_client_typename = "DataZone"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_datazone_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["datazone_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "detective" {

  sdk {
    id             = "Detective"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Detective"
    human_friendly      = "Detective"
  }

  client {
    go_v1_client_typename = "Detective"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListGraphs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_detective_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["detective_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "devicefarm" {

  sdk {
    id             = "Device Farm"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DeviceFarm"
    human_friendly      = "Device Farm"
  }

  client {
    go_v1_client_typename = "DeviceFarm"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDeviceInstances"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_devicefarm_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["devicefarm_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "DevOpsGuru"
    human_friendly      = "DevOps Guru"
  }

  client {
    go_v1_client_typename = "DevOpsGuru"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeAccountHealth"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_devopsguru_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["devopsguru_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "directconnect" {

  sdk {
    id             = "Direct Connect"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DirectConnect"
    human_friendly      = "Direct Connect"
  }

  client {
    go_v1_client_typename = "DirectConnect"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeConnections"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_dx_"
    correct = "aws_directconnect_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["dx_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "dlm" {

  sdk {
    id             = "DLM"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DLM"
    human_friendly      = "DLM (Data Lifecycle Manager)"
  }

  client {
    go_v1_client_typename = "DLM"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetLifecyclePolicies"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_dlm_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["dlm_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "dms" {

  go_packages {
    v1_package = "databasemigrationservice"
    v2_package = "databasemigrationservice"
  }

  sdk {
    id             = "Database Migration Service"
    client_version = [1]
  }

  names {
    aliases             = ["databasemigration", "databasemigrationservice"]
    provider_name_upper = "DMS"
    human_friendly      = "DMS (Database Migration)"
  }

  client {
    go_v1_client_typename = "DatabaseMigrationService"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeCertificates"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_dms_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["dms_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "docdb" {

  sdk {
    id             = "DocDB"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DocDB"
    human_friendly      = "DocumentDB"
  }

  client {
    go_v1_client_typename = "DocDB"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDBClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_docdb_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["docdb_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "DocDBElastic"
    human_friendly      = "DocumentDB Elastic"
  }

  client {
    go_v1_client_typename = "DocDBElastic"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_docdbelastic_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["docdbelastic_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "drs" {

  sdk {
    id             = "DRS"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DRS"
    human_friendly      = "DRS (Elastic Disaster Recovery)"
  }

  client {
    go_v1_client_typename = "Drs"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeJobs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_drs_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["drs_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "ds" {

  go_packages {
    v1_package = "directoryservice"
    v2_package = "directoryservice"
  }

  sdk {
    id             = "Directory Service"
    client_version = [1, 2]
  }

  names {
    aliases             = ["directoryservice"]
    provider_name_upper = "DS"
    human_friendly      = "Directory Service"
  }

  client {
    go_v1_client_typename = "DirectoryService"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDirectories"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_directory_service_"
    correct = "aws_ds_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["directory_service_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "dax" {

  sdk {
    id             = "DAX"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DAX"
    human_friendly      = "DynamoDB Accelerator (DAX)"
  }

  client {
    go_v1_client_typename = "DAX"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_dax_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["dax_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "dynamodbstreams" {

  sdk {
    id             = "DynamoDB Streams"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DynamoDBStreams"
    human_friendly      = "DynamoDB Streams"
  }

  client {
    go_v1_client_typename = "DynamoDBStreams"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_dynamodbstreams_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["dynamodbstreams_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "ebs" {

  sdk {
    id             = "EBS"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "EBS"
    human_friendly      = "EBS (Elastic Block Store)"
  }

  client {
    go_v1_client_typename = "EBS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ebs_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["changewhenimplemented"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "imagebuilder" {

  sdk {
    id             = "imagebuilder"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ImageBuilder"
    human_friendly      = "EC2 Image Builder"
  }

  client {
    go_v1_client_typename = "Imagebuilder"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListImages"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_imagebuilder_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["imagebuilder_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "EC2InstanceConnect"
    human_friendly      = "EC2 Instance Connect"
  }

  client {
    go_v1_client_typename = "EC2InstanceConnect"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ec2instanceconnect_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ec2instanceconnect_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "ecr" {

  sdk {
    id             = "ECR"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ECR"
    human_friendly      = "ECR (Elastic Container Registry)"
  }

  client {
    go_v1_client_typename = "ECR"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeRepositories"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ecr_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ecr_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ECRPublic"
    human_friendly      = "ECR Public"
  }

  client {
    go_v1_client_typename = "ECRPublic"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeRepositories"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ecrpublic_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ecrpublic_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "ecs" {

  sdk {
    id             = "ECS"
    client_version = [1, 2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ECS"
    human_friendly      = "ECS (Elastic Container)"
  }

  client {
    go_v1_client_typename = "ECS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ecs_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ecs_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "efs" {

  sdk {
    id             = "EFS"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "EFS"
    human_friendly      = "EFS (Elastic File System)"
  }

  client {
    go_v1_client_typename = "EFS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeFileSystems"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_efs_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["efs_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "eks" {

  sdk {
    id             = "EKS"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "EKS"
    human_friendly      = "EKS (Elastic Kubernetes)"
  }

  client {
    go_v1_client_typename = "EKS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_eks_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["eks_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "ElasticBeanstalk"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAvailableSolutionStacks"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_elastic_beanstalk_"
    correct = "aws_elasticbeanstalk_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["elastic_beanstalk_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ElasticInference"
    human_friendly      = "Elastic Inference"
  }

  client {
    go_v1_client_typename = "ElasticInference"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_elasticinference_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["elasticinference_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "elastictranscoder" {

  sdk {
    id             = "Elastic Transcoder"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ElasticTranscoder"
    human_friendly      = "Elastic Transcoder"
  }

  client {
    go_v1_client_typename = "ElasticTranscoder"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPipelines"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_elastictranscoder_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["elastictranscoder_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "elasticache" {

  sdk {
    id             = "ElastiCache"
    client_version = [1, 2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ElastiCache"
    human_friendly      = "ElastiCache"
  }

  client {
    go_v1_client_typename = "ElastiCache"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeCacheClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_elasticache_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["elasticache_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomainNames"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_elasticsearch_"
    correct = "aws_es_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["elasticsearch_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "elbv2" {

  go_packages {
    v1_package = "elbv2"
    v2_package = "elasticloadbalancingv2"
  }

  sdk {
    id             = "Elastic Load Balancing v2"
    client_version = [1, 2]
  }

  names {
    aliases             = ["elasticloadbalancingv2"]
    provider_name_upper = "ELBV2"
    human_friendly      = "ELB (Elastic Load Balancing)"
  }

  client {
    go_v1_client_typename = "ELBV2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeLoadBalancers"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_a?lb(\\b|_listener|_target_group|s|_trust_store)"
    correct = "aws_elbv2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lbs?\\.", "lb_listener", "lb_target_group", "lb_hosted", "lb_trust_store"]
  brand               = ""
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "elb" {

  go_packages {
    v1_package = "elb"
    v2_package = "elasticloadbalancing"
  }

  sdk {
    id             = "Elastic Load Balancing"
    client_version = [1]
  }

  names {
    aliases             = ["elasticloadbalancing"]
    provider_name_upper = "ELB"
    human_friendly      = "ELB Classic"
  }

  client {
    go_v1_client_typename = "ELB"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeLoadBalancers"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_(app_cookie_stickiness_policy|elb|lb_cookie_stickiness_policy|lb_ssl_negotiation_policy|load_balancer_|proxy_protocol_policy)"
    correct = "aws_elb_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["app_cookie_stickiness_policy", "elb", "lb_cookie_stickiness_policy", "lb_ssl_negotiation_policy", "load_balancer", "proxy_protocol_policy"]
  brand               = ""
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "mediaconnect" {

  sdk {
    id             = "MediaConnect"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MediaConnect"
    human_friendly      = "Elemental MediaConnect"
  }

  client {
    go_v1_client_typename = "MediaConnect"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListBridges"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mediaconnect_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mediaconnect_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "mediaconvert" {

  sdk {
    id             = "MediaConvert"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MediaConvert"
    human_friendly      = "Elemental MediaConvert"
  }

  client {
    go_v1_client_typename = "MediaConvert"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListJobs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_media_convert_"
    correct = "aws_mediaconvert_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["media_convert_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "medialive" {

  sdk {
    id             = "MediaLive"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MediaLive"
    human_friendly      = "Elemental MediaLive"
  }

  client {
    go_v1_client_typename = "MediaLive"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListOfferings"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_medialive_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["medialive_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "mediapackage" {

  sdk {
    id             = "MediaPackage"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MediaPackage"
    human_friendly      = "Elemental MediaPackage"
  }

  client {
    go_v1_client_typename = "MediaPackage"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListChannels"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_media_package_"
    correct = "aws_mediapackage_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["media_package_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "MediaPackageVOD"
    human_friendly      = "Elemental MediaPackage VOD"
  }

  client {
    go_v1_client_typename = "MediaPackageVod"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mediapackagevod_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mediapackagevod_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "mediastore" {

  sdk {
    id             = "MediaStore"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MediaStore"
    human_friendly      = "Elemental MediaStore"
  }

  client {
    go_v1_client_typename = "MediaStore"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListContainers"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_media_store_"
    correct = "aws_mediastore_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["media_store_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "MediaStoreData"
    human_friendly      = "Elemental MediaStore Data"
  }

  client {
    go_v1_client_typename = "MediaStoreData"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mediastoredata_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mediastoredata_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "mediatailor" {

  sdk {
    id             = "MediaTailor"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MediaTailor"
    human_friendly      = "Elemental MediaTailor"
  }

  client {
    go_v1_client_typename = "MediaTailor"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mediatailor_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["media_tailor_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "emr" {

  sdk {
    id             = "EMR"
    client_version = [1, 2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "EMR"
    human_friendly      = "EMR"
  }

  client {
    go_v1_client_typename = "EMR"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_emr_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["emr_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "emrcontainers" {

  cli_v2_command {
    aws_cli_v2_command           = "emr-containers"
    aws_cli_v2_command_no_dashes = "emrcontainers"
  }

  sdk {
    id             = "EMR containers"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "EMRContainers"
    human_friendly      = "EMR Containers"
  }

  client {
    go_v1_client_typename = "EMRContainers"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListVirtualClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_emrcontainers_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["emrcontainers_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "EMRServerless"
    human_friendly      = "EMR Serverless"
  }

  client {
    go_v1_client_typename = "EMRServerless"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_emrserverless_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["emrserverless_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "EventBridge"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListEventBuses"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_cloudwatch_event_"
    correct = "aws_events_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["cloudwatch_event_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "schemas" {

  sdk {
    id             = "schemas"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Schemas"
    human_friendly      = "EventBridge Schemas"
  }

  client {
    go_v1_client_typename = "Schemas"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRegistries"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_schemas_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["schemas_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "fis" {

  sdk {
    id             = "fis"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "FIS"
    human_friendly      = "FIS (Fault Injection Simulator)"
  }

  client {
    go_v1_client_typename = "FIS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListExperiments"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_fis_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["fis_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "finspace" {

  sdk {
    id             = "finspace"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "FinSpace"
    human_friendly      = "FinSpace"
  }

  client {
    go_v1_client_typename = "Finspace"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListEnvironments"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_finspace_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["finspace_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "FinSpaceData"
    human_friendly      = "FinSpace Data"
  }

  client {
    go_v1_client_typename = "FinSpaceData"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_finspacedata_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["finspacedata_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "fms" {

  sdk {
    id             = "FMS"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "FMS"
    human_friendly      = "FMS (Firewall Manager)"
  }

  client {
    go_v1_client_typename = "FMS"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAppsLists"
    endpoint_api_params      = "MaxResults: aws_sdkv2.Int32(1)"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_fms_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["fms_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_forecast_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["forecast_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_forecastquery_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["forecastquery_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "frauddetector" {

  sdk {
    id             = "FraudDetector"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "FraudDetector"
    human_friendly      = "Fraud Detector"
  }

  client {
    go_v1_client_typename = "FraudDetector"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_frauddetector_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["frauddetector_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "fsx" {

  sdk {
    id             = "FSx"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "FSx"
    human_friendly      = "FSx"
  }

  client {
    go_v1_client_typename = "FSx"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeFileSystems"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_fsx_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["fsx_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "gamelift" {

  sdk {
    id             = "GameLift"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "GameLift"
    human_friendly      = "GameLift"
  }

  client {
    go_v1_client_typename = "GameLift"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListGameServerGroups"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_gamelift_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["gamelift_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "globalaccelerator" {

  sdk {
    id             = "Global Accelerator"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "GlobalAccelerator"
    human_friendly      = "Global Accelerator"
  }

  client {
    go_v1_client_typename = "GlobalAccelerator"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAccelerators"
    endpoint_api_params      = ""
    endpoint_region_override = "us-west-2"
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_globalaccelerator_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["globalaccelerator_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "glue" {

  sdk {
    id             = "Glue"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Glue"
    human_friendly      = "Glue"
  }

  client {
    go_v1_client_typename = "Glue"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRegistries"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_glue_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["glue_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "databrew" {

  go_packages {
    v1_package = "gluedatabrew"
    v2_package = "databrew"
  }

  sdk {
    id             = "DataBrew"
    client_version = [1]
  }

  names {
    aliases             = ["gluedatabrew"]
    provider_name_upper = "DataBrew"
    human_friendly      = "Glue DataBrew"
  }

  client {
    go_v1_client_typename = "GlueDataBrew"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_databrew_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["databrew_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "groundstation" {

  sdk {
    id             = "GroundStation"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "GroundStation"
    human_friendly      = "Ground Station"
  }

  client {
    go_v1_client_typename = "GroundStation"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListConfigs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_groundstation_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["groundstation_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "guardduty" {

  sdk {
    id             = "GuardDuty"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "GuardDuty"
    human_friendly      = "GuardDuty"
  }

  client {
    go_v1_client_typename = "GuardDuty"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDetectors"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_guardduty_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["guardduty_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "health" {

  sdk {
    id             = "Health"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Health"
    human_friendly      = "Health"
  }

  client {
    go_v1_client_typename = "Health"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_health_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["health_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "healthlake" {

  sdk {
    id             = "HealthLake"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "HealthLake"
    human_friendly      = "HealthLake"
  }

  client {
    go_v1_client_typename = "HealthLake"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListFHIRDatastores"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_healthlake_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["healthlake_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "honeycode" {

  sdk {
    id             = "Honeycode"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Honeycode"
    human_friendly      = "Honeycode"
  }

  client {
    go_v1_client_typename = "Honeycode"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_honeycode_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["honeycode_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iam" {

  sdk {
    id             = "IAM"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IAM"
    human_friendly      = "IAM (Identity & Access Management)"
  }

  client {
    go_v1_client_typename = "IAM"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = "AWS_IAM_ENDPOINT"
    tf_aws_env_var     = "TF_AWS_IAM_ENDPOINT"
  }

  endpoint_info {
    endpoint_api_call        = "ListRoles"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iam_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iam_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "inspector" {

  sdk {
    id             = "Inspector"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Inspector"
    human_friendly      = "Inspector Classic"
  }

  client {
    go_v1_client_typename = "Inspector"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRulesPackages"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_inspector_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["inspector_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "Inspector2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAccountPermissions"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_inspector2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["inspector2_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iot1clickdevices_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iot1clickdevices_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "IoT1ClickProjects"
    human_friendly      = "IoT 1-Click Projects"
  }

  client {
    go_v1_client_typename = "IoT1ClickProjects"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iot1clickprojects_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iot1clickprojects_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iotanalytics" {

  sdk {
    id             = "IoTAnalytics"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoTAnalytics"
    human_friendly      = "IoT Analytics"
  }

  client {
    go_v1_client_typename = "IoTAnalytics"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListChannels"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotanalytics_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotanalytics_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotdata_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotdata_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iotdeviceadvisor" {

  sdk {
    id             = "IotDeviceAdvisor"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoTDeviceAdvisor"
    human_friendly      = "IoT Device Management"
  }

  client {
    go_v1_client_typename = "IoTDeviceAdvisor"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotdeviceadvisor_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotdeviceadvisor_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iotevents" {

  sdk {
    id             = "IoT Events"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoTEvents"
    human_friendly      = "IoT Events"
  }

  client {
    go_v1_client_typename = "IoTEvents"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAlarmModels"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotevents_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotevents_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "IoTEventsData"
    human_friendly      = "IoT Events Data"
  }

  client {
    go_v1_client_typename = "IoTEventsData"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ioteventsdata_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ioteventsdata_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iotfleethub" {

  sdk {
    id             = "IoTFleetHub"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoTFleetHub"
    human_friendly      = "IoT Fleet Hub"
  }

  client {
    go_v1_client_typename = "IoTFleetHub"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotfleethub_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotfleethub_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "greengrass" {

  sdk {
    id             = "Greengrass"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Greengrass"
    human_friendly      = "IoT Greengrass"
  }

  client {
    go_v1_client_typename = "Greengrass"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListGroups"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_greengrass_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["greengrass_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "greengrassv2" {

  sdk {
    id             = "GreengrassV2"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "GreengrassV2"
    human_friendly      = "IoT Greengrass V2"
  }

  client {
    go_v1_client_typename = "GreengrassV2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_greengrassv2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["greengrassv2_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotjobsdata_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotjobsdata_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iotsecuretunneling" {

  sdk {
    id             = "IoTSecureTunneling"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoTSecureTunneling"
    human_friendly      = "IoT Secure Tunneling"
  }

  client {
    go_v1_client_typename = "IoTSecureTunneling"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotsecuretunneling_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotsecuretunneling_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iotsitewise" {

  sdk {
    id             = "IoTSiteWise"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoTSiteWise"
    human_friendly      = "IoT SiteWise"
  }

  client {
    go_v1_client_typename = "IoTSiteWise"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotsitewise_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotsitewise_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iotthingsgraph" {

  sdk {
    id             = "IoTThingsGraph"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoTThingsGraph"
    human_friendly      = "IoT Things Graph"
  }

  client {
    go_v1_client_typename = "IoTThingsGraph"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotthingsgraph_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotthingsgraph_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iottwinmaker" {

  sdk {
    id             = "IoTTwinMaker"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoTTwinMaker"
    human_friendly      = "IoT TwinMaker"
  }

  client {
    go_v1_client_typename = "IoTTwinMaker"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iottwinmaker_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iottwinmaker_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "iotwireless" {

  sdk {
    id             = "IoT Wireless"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoTWireless"
    human_friendly      = "IoT Wireless"
  }

  client {
    go_v1_client_typename = "IoTWireless"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iotwireless_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iotwireless_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "ivs" {

  sdk {
    id             = "ivs"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IVS"
    human_friendly      = "IVS (Interactive Video)"
  }

  client {
    go_v1_client_typename = "IVS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListChannels"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ivs_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ivs_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "ivschat" {

  sdk {
    id             = "ivschat"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IVSChat"
    human_friendly      = "IVS (Interactive Video) Chat"
  }

  client {
    go_v1_client_typename = "Ivschat"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRooms"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ivschat_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ivschat_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "kendra" {

  sdk {
    id             = "kendra"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Kendra"
    human_friendly      = "Kendra"
  }

  client {
    go_v1_client_typename = "Kendra"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListIndices"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_kendra_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kendra_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "keyspaces" {

  sdk {
    id             = "Keyspaces"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Keyspaces"
    human_friendly      = "Keyspaces (for Apache Cassandra)"
  }

  client {
    go_v1_client_typename = "Keyspaces"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListKeyspaces"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_keyspaces_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["keyspaces_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "kinesis" {

  sdk {
    id             = "Kinesis"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Kinesis"
    human_friendly      = "Kinesis"
  }

  client {
    go_v1_client_typename = "Kinesis"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListStreams"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_kinesis_stream"
    correct = "aws_kinesis_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kinesis_stream", "kinesis_resource_policy"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "kinesisanalytics" {

  sdk {
    id             = "Kinesis Analytics"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "KinesisAnalytics"
    human_friendly      = "Kinesis Analytics"
  }

  client {
    go_v1_client_typename = "KinesisAnalytics"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_kinesis_analytics_"
    correct = "aws_kinesisanalytics_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kinesis_analytics_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "kinesisanalyticsv2" {

  sdk {
    id             = "Kinesis Analytics V2"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "KinesisAnalyticsV2"
    human_friendly      = "Kinesis Analytics V2"
  }

  client {
    go_v1_client_typename = "KinesisAnalyticsV2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_kinesisanalyticsv2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kinesisanalyticsv2_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "firehose" {

  sdk {
    id             = "Firehose"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Firehose"
    human_friendly      = "Kinesis Firehose"
  }

  client {
    go_v1_client_typename = "Firehose"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDeliveryStreams"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_kinesis_firehose_"
    correct = "aws_firehose_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kinesis_firehose_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "kinesisvideo" {

  sdk {
    id             = "Kinesis Video"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "KinesisVideo"
    human_friendly      = "Kinesis Video"
  }

  client {
    go_v1_client_typename = "KinesisVideo"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListStreams"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_kinesisvideo_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kinesis_video_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "KinesisVideoArchivedMedia"
    human_friendly      = "Kinesis Video Archived Media"
  }

  client {
    go_v1_client_typename = "KinesisVideoArchivedMedia"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_kinesisvideoarchivedmedia_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kinesisvideoarchivedmedia_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "KinesisVideoMedia"
    human_friendly      = "Kinesis Video Media"
  }

  client {
    go_v1_client_typename = "KinesisVideoMedia"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_kinesisvideomedia_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kinesisvideomedia_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_kinesisvideosignaling_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kinesisvideosignaling_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "kms" {

  sdk {
    id             = "KMS"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "KMS"
    human_friendly      = "KMS (Key Management)"
  }

  client {
    go_v1_client_typename = "KMS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListKeys"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_kms_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["kms_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "lakeformation" {

  sdk {
    id             = "LakeFormation"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "LakeFormation"
    human_friendly      = "Lake Formation"
  }

  client {
    go_v1_client_typename = "LakeFormation"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListResources"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_lakeformation_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lakeformation_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "lambda" {

  sdk {
    id             = "Lambda"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Lambda"
    human_friendly      = "Lambda"
  }

  client {
    go_v1_client_typename = "Lambda"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListFunctions"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_lambda_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lambda_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "LaunchWizard"
    human_friendly      = "Launch Wizard"
  }

  client {
    go_v1_client_typename = "LaunchWizard"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListWorkloads"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_launchwizard_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["launchwizard_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetBots"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_lex_"
    correct = "aws_lexmodels_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lex_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "LexModelsV2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListBots"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_lexv2models_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lexv2models_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_lexruntime_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lexruntime_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_lexruntimev2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lexruntimev2_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "LicenseManager"
    human_friendly      = "License Manager"
  }

  client {
    go_v1_client_typename = "LicenseManager"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListLicenseConfigurations"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_licensemanager_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["licensemanager_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "lightsail" {

  sdk {
    id             = "Lightsail"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Lightsail"
    human_friendly      = "Lightsail"
  }

  client {
    go_v1_client_typename = "Lightsail"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetInstances"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_lightsail_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lightsail_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListGeofenceCollections"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_location_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["location_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "lookoutequipment" {

  sdk {
    id             = "LookoutEquipment"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "LookoutEquipment"
    human_friendly      = "Lookout for Equipment"
  }

  client {
    go_v1_client_typename = "LookoutEquipment"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_lookoutequipment_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lookoutequipment_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "lookoutmetrics" {

  sdk {
    id             = "LookoutMetrics"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "LookoutMetrics"
    human_friendly      = "Lookout for Metrics"
  }

  client {
    go_v1_client_typename = "LookoutMetrics"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListMetricSets"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_lookoutmetrics_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lookoutmetrics_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_lookoutvision_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["lookoutvision_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "machinelearning" {

  sdk {
    id             = "Machine Learning"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MachineLearning"
    human_friendly      = "Machine Learning"
  }

  client {
    go_v1_client_typename = "MachineLearning"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_machinelearning_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["machinelearning_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "macie2" {

  sdk {
    id             = "Macie2"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Macie2"
    human_friendly      = "Macie"
  }

  client {
    go_v1_client_typename = "Macie2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListFindings"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_macie2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["macie2_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "macie" {

  sdk {
    id             = "Macie"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Macie"
    human_friendly      = "Macie Classic"
  }

  client {
    go_v1_client_typename = "Macie"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_macie_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["macie_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "m2" {

  sdk {
    id             = "m2"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "M2"
    human_friendly      = "Mainframe Modernization"
  }

  client {
    go_v1_client_typename = "M2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_m2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["m2_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "managedblockchain" {

  sdk {
    id             = "ManagedBlockchain"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ManagedBlockchain"
    human_friendly      = "Managed Blockchain"
  }

  client {
    go_v1_client_typename = "ManagedBlockchain"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_managedblockchain_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["managedblockchain_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "grafana" {

  go_packages {
    v1_package = "managedgrafana"
    v2_package = "grafana"
  }

  sdk {
    id             = "grafana"
    client_version = [1]
  }

  names {
    aliases             = ["managedgrafana", "amg"]
    provider_name_upper = "Grafana"
    human_friendly      = "Managed Grafana"
  }

  client {
    go_v1_client_typename = "ManagedGrafana"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListWorkspaces"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_grafana_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["grafana_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    go_v1_client_typename = "Kafka"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_msk_"
    correct = "aws_kafka_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["msk_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "kafkaconnect" {

  sdk {
    id             = "KafkaConnect"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "KafkaConnect"
    human_friendly      = "Managed Streaming for Kafka Connect"
  }

  client {
    go_v1_client_typename = "KafkaConnect"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListConnectors"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_mskconnect_"
    correct = "aws_kafkaconnect_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mskconnect_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "MarketplaceCatalog"
    human_friendly      = "Marketplace Catalog"
  }

  client {
    go_v1_client_typename = "MarketplaceCatalog"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_marketplacecatalog_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["marketplace_catalog_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "marketplacecommerceanalytics" {

  sdk {
    id             = "Marketplace Commerce Analytics"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MarketplaceCommerceAnalytics"
    human_friendly      = "Marketplace Commerce Analytics"
  }

  client {
    go_v1_client_typename = "MarketplaceCommerceAnalytics"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_marketplacecommerceanalytics_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["marketplacecommerceanalytics_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_marketplaceentitlement_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["marketplaceentitlement_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_marketplacemetering_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["marketplacemetering_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "memorydb" {

  sdk {
    id             = "MemoryDB"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MemoryDB"
    human_friendly      = "MemoryDB for Redis"
  }

  client {
    go_v1_client_typename = "MemoryDB"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_memorydb_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["memorydb_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "meta" {

  cli_v2_command {
    aws_cli_v2_command           = ""
    aws_cli_v2_command_no_dashes = ""
  }

  go_packages {
    v1_package = ""
    v2_package = ""
  }

  sdk {
    id             = ""
    client_version = null
  }

  names {
    aliases             = [""]
    provider_name_upper = "Meta"
    human_friendly      = "Meta Data Sources"
  }

  client {
    go_v1_client_typename = ""
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_(arn|billing_service_account|default_tags|ip_ranges|partition|regions?|service)$"
    correct = "aws_meta_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["arn", "ip_ranges", "billing_service_account", "default_tags", "partition", "region", "service\\."]
  brand               = ""
  exclude             = true
  not_implemented     = false
  allowed_subcategory = true
  note                = "Not an AWS service (metadata)"
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mgh_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mgh_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "MigrationHubConfig"
    human_friendly      = "Migration Hub Config"
  }

  client {
    go_v1_client_typename = "MigrationHubConfig"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_migrationhubconfig_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["migrationhubconfig_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "MigrationHubRefactorSpaces"
    human_friendly      = "Migration Hub Refactor Spaces"
  }

  client {
    go_v1_client_typename = "MigrationHubRefactorSpaces"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_migrationhubrefactorspaces_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["migrationhubrefactorspaces_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_migrationhubstrategy_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["migrationhubstrategy_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "mobile" {

  sdk {
    id             = "Mobile"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Mobile"
    human_friendly      = "Mobile"
  }

  client {
    go_v1_client_typename = "Mobile"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mobile_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mobile_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "mq" {

  sdk {
    id             = "mq"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MQ"
    human_friendly      = "MQ"
  }

  client {
    go_v1_client_typename = "MQ"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListBrokers"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mq_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mq_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "mturk" {

  sdk {
    id             = "MTurk"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MTurk"
    human_friendly      = "MTurk (Mechanical Turk)"
  }

  client {
    go_v1_client_typename = "MTurk"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mturk_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mturk_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "mwaa" {

  sdk {
    id             = "MWAA"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MWAA"
    human_friendly      = "MWAA (Managed Workflows for Apache Airflow)"
  }

  client {
    go_v1_client_typename = "MWAA"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListEnvironments"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_mwaa_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["mwaa_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "neptune" {

  sdk {
    id             = "Neptune"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Neptune"
    human_friendly      = "Neptune"
  }

  client {
    go_v1_client_typename = "Neptune"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDBClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_neptune_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["neptune_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "NeptuneGraph"
    human_friendly      = "Neptune Analytics"
  }

  client {
    go_v1_client_typename = ""
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListGraphs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_neptunegraph_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["neptunegraph_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "networkfirewall" {

  cli_v2_command {
    aws_cli_v2_command           = "network-firewall"
    aws_cli_v2_command_no_dashes = "networkfirewall"
  }

  sdk {
    id             = "Network Firewall"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "NetworkFirewall"
    human_friendly      = "Network Firewall"
  }

  client {
    go_v1_client_typename = "NetworkFirewall"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListFirewalls"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_networkfirewall_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["networkfirewall_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "networkmanager" {

  sdk {
    id             = "NetworkManager"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "NetworkManager"
    human_friendly      = "Network Manager"
  }

  client {
    go_v1_client_typename = "NetworkManager"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListCoreNetworks"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_networkmanager_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["networkmanager_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_nimble_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["nimble_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "OAM"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListLinks"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_oam_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["oam_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomainNames"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_opensearch_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["opensearch_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "opensearchserverless" {

  sdk {
    id             = "OpenSearchServerless"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "OpenSearchServerless"
    human_friendly      = "OpenSearch Serverless"
  }

  client {
    go_v1_client_typename = "OpenSearchServerless"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListCollections"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_opensearchserverless_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["opensearchserverless_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "OSIS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPipelines"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_osis_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["osis_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "opsworks" {

  sdk {
    id             = "OpsWorks"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "OpsWorks"
    human_friendly      = "OpsWorks"
  }

  client {
    go_v1_client_typename = "OpsWorks"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeApps"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_opsworks_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["opsworks_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "OpsWorksCM"
    human_friendly      = "OpsWorks CM"
  }

  client {
    go_v1_client_typename = "OpsWorksCM"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_opsworkscm_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["opsworkscm_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "organizations" {

  sdk {
    id             = "Organizations"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Organizations"
    human_friendly      = "Organizations"
  }

  client {
    go_v1_client_typename = "Organizations"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAccounts"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_organizations_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["organizations_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "outposts" {

  sdk {
    id             = "Outposts"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Outposts"
    human_friendly      = "Outposts"
  }

  client {
    go_v1_client_typename = "Outposts"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListSites"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_outposts_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["outposts_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "panorama" {

  sdk {
    id             = "Panorama"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Panorama"
    human_friendly      = "Panorama"
  }

  client {
    go_v1_client_typename = "Panorama"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_panorama_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["panorama_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "PaymentCryptography"
    human_friendly      = "Payment Cryptography Control Plane"
  }

  client {
    go_v1_client_typename = "PaymentCryptography"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListKeys"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_paymentcryptography_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["paymentcryptography_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "PCAConnectorAD"
    human_friendly      = "Private CA Connector for Active Directory"
  }

  client {
    go_v1_client_typename = "PcaConnectorAd"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListConnectors"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_pcaconnectorad_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["pcaconnectorad_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "personalize" {

  sdk {
    id             = "Personalize"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Personalize"
    human_friendly      = "Personalize"
  }

  client {
    go_v1_client_typename = "Personalize"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_personalize_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["personalize_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "PersonalizeEvents"
    human_friendly      = "Personalize Events"
  }

  client {
    go_v1_client_typename = "PersonalizeEvents"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_personalizeevents_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["personalizeevents_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "PersonalizeRuntime"
    human_friendly      = "Personalize Runtime"
  }

  client {
    go_v1_client_typename = "PersonalizeRuntime"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_personalizeruntime_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["personalizeruntime_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "pinpoint" {

  sdk {
    id             = "Pinpoint"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Pinpoint"
    human_friendly      = "Pinpoint"
  }

  client {
    go_v1_client_typename = "Pinpoint"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetApps"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_pinpoint_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["pinpoint_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "PinpointEmail"
    human_friendly      = "Pinpoint Email"
  }

  client {
    go_v1_client_typename = "PinpointEmail"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_pinpointemail_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["pinpointemail_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "PinpointSMSVoice"
    human_friendly      = "Pinpoint SMS and Voice"
  }

  client {
    go_v1_client_typename = "PinpointSMSVoice"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_pinpointsmsvoice_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["pinpointsmsvoice_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "pipes" {

  sdk {
    id             = "Pipes"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Pipes"
    human_friendly      = "EventBridge Pipes"
  }

  client {
    go_v1_client_typename = "Pipes"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPipes"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_pipes_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["pipes_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "polly" {

  sdk {
    id             = "Polly"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Polly"
    human_friendly      = "Polly"
  }

  client {
    go_v1_client_typename = "Polly"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListLexicons"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_polly_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["polly_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "pricing" {

  sdk {
    id             = "Pricing"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Pricing"
    human_friendly      = "Pricing Calculator"
  }

  client {
    go_v1_client_typename = "Pricing"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeServices"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_pricing_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["pricing_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "proton" {

  sdk {
    id             = "Proton"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Proton"
    human_friendly      = "Proton"
  }

  client {
    go_v1_client_typename = "Proton"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_proton_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["proton_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "qbusiness" {

  sdk {
    id             = "QBusiness"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "QBusiness"
    human_friendly      = "Amazon Q Business"
  }

  client {
    go_v1_client_typename = "QBusiness"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_qbusiness_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["qbusiness_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "qldb" {

  sdk {
    id             = "QLDB"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "QLDB"
    human_friendly      = "QLDB (Quantum Ledger Database)"
  }

  client {
    go_v1_client_typename = "QLDB"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListLedgers"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_qldb_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["qldb_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "QLDBSession"
    human_friendly      = "QLDB Session"
  }

  client {
    go_v1_client_typename = "QLDBSession"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_qldbsession_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["qldbsession_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "quicksight" {

  sdk {
    id             = "QuickSight"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "QuickSight"
    human_friendly      = "QuickSight"
  }

  client {
    go_v1_client_typename = "QuickSight"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDashboards"
    endpoint_api_params      = "AwsAccountId: aws_sdkv1.String(\"123456789012\")"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_quicksight_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["quicksight_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "ram" {

  sdk {
    id             = "RAM"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "RAM"
    human_friendly      = "RAM (Resource Access Manager)"
  }

  client {
    go_v1_client_typename = "RAM"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPermissions"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ram_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ram_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "rds" {

  sdk {
    id             = "RDS"
    client_version = [1, 2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "RDS"
    human_friendly      = "RDS (Relational Database)"
  }

  client {
    go_v1_client_typename = "RDS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDBInstances"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_(db_|rds_)"
    correct = "aws_rds_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["rds_", "db_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_rdsdata_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["rdsdata_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "pi" {

  sdk {
    id             = "PI"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "PI"
    human_friendly      = "RDS Performance Insights (PI)"
  }

  client {
    go_v1_client_typename = "PI"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_pi_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["pi_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "RecycleBin"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRules"
    endpoint_api_params      = "ResourceType: awstypes.ResourceTypeEc2Image"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_rbin_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["rbin_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "redshift" {

  sdk {
    id             = "Redshift"
    client_version = [1, 2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Redshift"
    human_friendly      = "Redshift"
  }

  client {
    go_v1_client_typename = "Redshift"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_redshift_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["redshift_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "RedshiftDataAPIService"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDatabases"
    endpoint_api_params      = "Database: aws_sdkv2.String(\"test\")"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_redshiftdata_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["redshiftdata_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "RedshiftServerless"
    human_friendly      = "Redshift Serverless"
  }

  client {
    go_v1_client_typename = "RedshiftServerless"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListNamespaces"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_redshiftserverless_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["redshiftserverless_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "rekognition" {

  sdk {
    id             = "Rekognition"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Rekognition"
    human_friendly      = "Rekognition"
  }

  client {
    go_v1_client_typename = "Rekognition"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListCollections"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_rekognition_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["rekognition_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "resiliencehub" {

  sdk {
    id             = "resiliencehub"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ResilienceHub"
    human_friendly      = "Resilience Hub"
  }

  client {
    go_v1_client_typename = "ResilienceHub"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_resiliencehub_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["resiliencehub_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ResourceExplorer2"
    human_friendly      = "Resource Explorer"
  }

  client {
    go_v1_client_typename = "ResourceExplorer2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListIndexes"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_resourceexplorer2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["resourceexplorer2_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ResourceGroups"
    human_friendly      = "Resource Groups"
  }

  client {
    go_v1_client_typename = "ResourceGroups"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListGroups"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_resourcegroups_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["resourcegroups_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "ResourceGroupsTaggingAPI"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "GetResources"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_resourcegroupstaggingapi_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["resourcegroupstaggingapi_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "robomaker" {

  sdk {
    id             = "RoboMaker"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "RoboMaker"
    human_friendly      = "RoboMaker"
  }

  client {
    go_v1_client_typename = "RoboMaker"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_robomaker_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["robomaker_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "rolesanywhere" {

  sdk {
    id             = "RolesAnywhere"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "RolesAnywhere"
    human_friendly      = "Roles Anywhere"
  }

  client {
    go_v1_client_typename = "RolesAnywhere"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListProfiles"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_rolesanywhere_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["rolesanywhere_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "route53" {

  sdk {
    id             = "Route 53"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Route53"
    human_friendly      = "Route 53"
  }

  client {
    go_v1_client_typename = "Route53"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListHostedZones"
    endpoint_api_params      = ""
    endpoint_region_override = "us-east-1"
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_route53_(?!resolver_)"
    correct = "aws_route53_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["route53_cidr_", "route53_delegation_", "route53_health_", "route53_hosted_", "route53_key_", "route53_query_", "route53_record", "route53_traffic_", "route53_vpc_", "route53_zone"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "route53domains" {

  sdk {
    id             = "Route 53 Domains"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Route53Domains"
    human_friendly      = "Route 53 Domains"
  }

  client {
    go_v1_client_typename = "Route53Domains"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
    endpoint_api_params      = ""
    endpoint_region_override = "us-east-1"
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_route53domains_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["route53domains_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "route53profiles" {

  sdk {
    id             = "Route 53 Profiles"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Route53Profiles"
    human_friendly      = "Route 53 Profiles"
  }

  client {
    go_v1_client_typename = "Route53Profiles"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListProfiles"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_route53profiles_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["route53profiles_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "Route53RecoveryCluster"
    human_friendly      = "Route 53 Recovery Cluster"
  }

  client {
    go_v1_client_typename = "Route53RecoveryCluster"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_route53recoverycluster_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["route53recoverycluster_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "Route53RecoveryControlConfig"
    human_friendly      = "Route 53 Recovery Control Config"
  }

  client {
    go_v1_client_typename = "Route53RecoveryControlConfig"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_route53recoverycontrolconfig_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["route53recoverycontrolconfig_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "Route53RecoveryReadiness"
    human_friendly      = "Route 53 Recovery Readiness"
  }

  client {
    go_v1_client_typename = "Route53RecoveryReadiness"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListCells"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_route53recoveryreadiness_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["route53recoveryreadiness_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "route53resolver" {

  sdk {
    id             = "Route53Resolver"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Route53Resolver"
    human_friendly      = "Route 53 Resolver"
  }

  client {
    go_v1_client_typename = "Route53Resolver"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListFirewallDomainLists"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_route53_resolver_"
    correct = "aws_route53resolver_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["route53_resolver_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    go_v1_client_typename = "S3"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = "AWS_S3_ENDPOINT"
    tf_aws_env_var     = "TF_AWS_S3_ENDPOINT"
  }

  endpoint_info {
    endpoint_api_call        = "ListBuckets"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_(canonical_user_id|s3_bucket|s3_object|s3_directory_bucket)"
    correct = "aws_s3_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["s3_bucket", "s3_directory_bucket", "s3_object", "canonical_user_id"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "s3control" {

  sdk {
    id             = "S3 Control"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "S3Control"
    human_friendly      = "S3 Control"
  }

  client {
    go_v1_client_typename = "S3Control"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListJobs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_(s3_account_|s3control_|s3_access_)"
    correct = "aws_s3control_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["s3control", "s3_account_", "s3_access_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "glacier" {

  sdk {
    id             = "Glacier"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Glacier"
    human_friendly      = "S3 Glacier"
  }

  client {
    go_v1_client_typename = "Glacier"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListVaults"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_glacier_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["glacier_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "s3outposts" {

  sdk {
    id             = "S3Outposts"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "S3Outposts"
    human_friendly      = "S3 on Outposts"
  }

  client {
    go_v1_client_typename = "S3Outposts"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListEndpoints"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_s3outposts_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["s3outposts_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "sagemaker" {

  sdk {
    id             = "SageMaker"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SageMaker"
    human_friendly      = "SageMaker"
  }

  client {
    go_v1_client_typename = "SageMaker"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListClusters"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sagemaker_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sagemaker_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sagemakera2iruntime_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sagemakera2iruntime_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sagemakeredge_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sagemakeredge_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "SageMakerFeatureStoreRuntime"
    human_friendly      = "SageMaker Feature Store Runtime"
  }

  client {
    go_v1_client_typename = "SageMakerFeatureStoreRuntime"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sagemakerfeaturestoreruntime_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sagemakerfeaturestoreruntime_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "SageMakerRuntime"
    human_friendly      = "SageMaker Runtime"
  }

  client {
    go_v1_client_typename = "SageMakerRuntime"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sagemakerruntime_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sagemakerruntime_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "savingsplans" {

  sdk {
    id             = "savingsplans"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SavingsPlans"
    human_friendly      = "Savings Plans"
  }

  client {
    go_v1_client_typename = "SavingsPlans"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_savingsplans_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["savingsplans_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_simpledb_"
    correct = "aws_sdb_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["simpledb_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "scheduler" {

  sdk {
    id             = "Scheduler"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Scheduler"
    human_friendly      = "EventBridge Scheduler"
  }

  client {
    go_v1_client_typename = "Scheduler"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListSchedules"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_scheduler_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["scheduler_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "secretsmanager" {

  sdk {
    id             = "Secrets Manager"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SecretsManager"
    human_friendly      = "Secrets Manager"
  }

  client {
    go_v1_client_typename = "SecretsManager"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListSecrets"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_secretsmanager_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["secretsmanager_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "securityhub" {

  sdk {
    id             = "SecurityHub"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SecurityHub"
    human_friendly      = "Security Hub"
  }

  client {
    go_v1_client_typename = "SecurityHub"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAutomationRules"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_securityhub_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["securityhub_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "securitylake" {

  sdk {
    id             = "SecurityLake"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SecurityLake"
    human_friendly      = "Security Lake"
  }

  client {
    go_v1_client_typename = "SecurityLake"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDataLakes"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_securitylake_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["securitylake_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "serverlessrepo" {

  go_packages {
    v1_package = "serverlessapplicationrepository"
    v2_package = "serverlessapplicationrepository"
  }

  sdk {
    id             = "ServerlessApplicationRepository"
    client_version = [1]
  }

  names {
    aliases             = ["serverlessapprepo", "serverlessapplicationrepository"]
    provider_name_upper = "ServerlessRepo"
    human_friendly      = "Serverless Application Repository"
  }

  client {
    go_v1_client_typename = "ServerlessApplicationRepository"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_serverlessapplicationrepository_"
    correct = "aws_serverlessrepo_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["serverlessapplicationrepository_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "servicecatalog" {

  sdk {
    id             = "Service Catalog"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "ServiceCatalog"
    human_friendly      = "Service Catalog"
  }

  client {
    go_v1_client_typename = "ServiceCatalog"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPortfolios"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_servicecatalog_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["servicecatalog_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "AppRegistry"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_servicecatalogappregistry_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["servicecatalogappregistry_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "ServiceQuotas"
    human_friendly      = "Service Quotas"
  }

  client {
    go_v1_client_typename = "ServiceQuotas"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListServices"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_servicequotas_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["servicequotas_"]
  brand               = ""
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "ses" {

  sdk {
    id             = "SES"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SES"
    human_friendly      = "SES (Simple Email)"
  }

  client {
    go_v1_client_typename = "SES"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListIdentities"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ses_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ses_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "sesv2" {

  sdk {
    id             = "SESv2"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SESV2"
    human_friendly      = "SESv2 (Simple Email V2)"
  }

  client {
    go_v1_client_typename = "SESV2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListContactLists"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sesv2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sesv2_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "sfn" {

  cli_v2_command {
    aws_cli_v2_command           = "stepfunctions"
    aws_cli_v2_command_no_dashes = "stepfunctions"
  }

  sdk {
    id             = "SFN"
    client_version = [1]
  }

  names {
    aliases             = ["stepfunctions"]
    provider_name_upper = "SFN"
    human_friendly      = "SFN (Step Functions)"
  }

  client {
    go_v1_client_typename = "SFN"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListActivities"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sfn_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sfn_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "shield" {

  sdk {
    id             = "Shield"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Shield"
    human_friendly      = "Shield"
  }

  client {
    go_v1_client_typename = "Shield"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListProtectionGroups"
    endpoint_api_params      = ""
    endpoint_region_override = "us-east-1"
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_shield_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["shield_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "signer" {

  sdk {
    id             = "signer"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Signer"
    human_friendly      = "Signer"
  }

  client {
    go_v1_client_typename = "Signer"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListSigningJobs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_signer_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["signer_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "sms" {

  sdk {
    id             = "SMS"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SMS"
    human_friendly      = "SMS (Server Migration)"
  }

  client {
    go_v1_client_typename = "SMS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sms_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sms_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "SnowDeviceManagement"
    human_friendly      = "Snow Device Management"
  }

  client {
    go_v1_client_typename = "SnowDeviceManagement"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_snowdevicemanagement_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["snowdevicemanagement_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "snowball" {

  sdk {
    id             = "Snowball"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Snowball"
    human_friendly      = "Snow Family"
  }

  client {
    go_v1_client_typename = "Snowball"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_snowball_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["snowball_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "sns" {

  sdk {
    id             = "SNS"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SNS"
    human_friendly      = "SNS (Simple Notification)"
  }

  client {
    go_v1_client_typename = "SNS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListSubscriptions"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sns_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sns_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "sqs" {

  sdk {
    id             = "SQS"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SQS"
    human_friendly      = "SQS (Simple Queue)"
  }

  client {
    go_v1_client_typename = "SQS"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListQueues"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sqs_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sqs_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "ssm" {

  sdk {
    id             = "SSM"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SSM"
    human_friendly      = "SSM (Systems Manager)"
  }

  client {
    go_v1_client_typename = "SSM"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDocuments"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ssm_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ssm_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "SSMContacts"
    human_friendly      = "SSM Contacts"
  }

  client {
    go_v1_client_typename = "SSMContacts"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListContacts"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ssmcontacts_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ssmcontacts_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "SSMIncidents"
    human_friendly      = "SSM Incident Manager Incidents"
  }

  client {
    go_v1_client_typename = "SSMIncidents"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListResponsePlans"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ssmincidents_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ssmincidents_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "SSMSAP"
    human_friendly      = "Systems Manager for SAP"
  }

  client {
    go_v1_client_typename = "SsmSap"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListApplications"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ssmsap_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ssmsap_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "sso" {

  sdk {
    id             = "SSO"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SSO"
    human_friendly      = "SSO (Single Sign-On)"
  }

  client {
    go_v1_client_typename = "SSO"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAccounts"
    endpoint_api_params      = "AccessToken: aws_sdkv2.String(\"mock-access-token\")"
    endpoint_region_override = ""
    endpoint_only            = true
  }

  resource_prefix {
    actual  = ""
    correct = "aws_sso_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["sso_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "SSOAdmin"
    human_friendly      = "SSO Admin"
  }

  client {
    go_v1_client_typename = "SSOAdmin"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListInstances"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ssoadmin_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ssoadmin_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "identitystore" {

  sdk {
    id             = "identitystore"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IdentityStore"
    human_friendly      = "SSO Identity Store"
  }

  client {
    go_v1_client_typename = "IdentityStore"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListUsers"
    endpoint_api_params      = "IdentityStoreId: aws_sdkv2.String(\"d-1234567890\")"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_identitystore_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["identitystore_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "SSOOIDC"
    human_friendly      = "SSO OIDC"
  }

  client {
    go_v1_client_typename = "SSOOIDC"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_ssooidc_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["ssooidc_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "storagegateway" {

  sdk {
    id             = "Storage Gateway"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "StorageGateway"
    human_friendly      = "Storage Gateway"
  }

  client {
    go_v1_client_typename = "StorageGateway"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListGateways"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_storagegateway_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["storagegateway_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "sts" {

  sdk {
    id             = "STS"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "STS"
    human_friendly      = "STS (Security Token)"
  }

  client {
    go_v1_client_typename = "STS"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = "AWS_STS_ENDPOINT"
    tf_aws_env_var     = "TF_AWS_STS_ENDPOINT"
  }

  endpoint_info {
    endpoint_api_call        = "GetCallerIdentity"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_caller_identity"
    correct = "aws_sts_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["caller_identity"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "support" {

  sdk {
    id             = "Support"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Support"
    human_friendly      = "Support"
  }

  client {
    go_v1_client_typename = "Support"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_support_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["support_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "swf" {

  sdk {
    id             = "SWF"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "SWF"
    human_friendly      = "SWF (Simple Workflow)"
  }

  client {
    go_v1_client_typename = "SWF"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDomains"
    endpoint_api_params      = "RegistrationStatus: \"REGISTERED\""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_swf_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["swf_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "textract" {

  sdk {
    id             = "Textract"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Textract"
    human_friendly      = "Textract"
  }

  client {
    go_v1_client_typename = "Textract"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_textract_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["textract_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "TimestreamInfluxDB"
    human_friendly      = "Timestream for InfluxDB"
  }

  client {
    go_v1_client_typename = "TimestreamInfluxDB"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDbInstances"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_timestreaminfluxdb_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["timestreaminfluxdb_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "TimestreamQuery"
    human_friendly      = "Timestream Query"
  }

  client {
    go_v1_client_typename = "TimestreamQuery"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_timestreamquery_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["timestreamquery_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "TimestreamWrite"
    human_friendly      = "Timestream Write"
  }

  client {
    go_v1_client_typename = "TimestreamWrite"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListDatabases"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_timestreamwrite_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["timestreamwrite_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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

  client {
    go_v1_client_typename = "TranscribeService"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListLanguageModels"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_transcribe_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["transcribe_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "transcribestreaming" {

  cli_v2_command {
    aws_cli_v2_command           = ""
    aws_cli_v2_command_no_dashes = ""
  }

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
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_transcribestreaming_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["transcribestreaming_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "transfer" {

  sdk {
    id             = "Transfer"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Transfer"
    human_friendly      = "Transfer Family"
  }

  client {
    go_v1_client_typename = "Transfer"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListConnectors"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_transfer_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["transfer_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "translate" {

  sdk {
    id             = "Translate"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Translate"
    human_friendly      = "Translate"
  }

  client {
    go_v1_client_typename = "Translate"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_translate_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["translate_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "VPCLattice"
    human_friendly      = "VPC Lattice"
  }

  client {
    go_v1_client_typename = "VPCLattice"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListServices"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_vpclattice_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["vpclattice_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "wafv2" {

  sdk {
    id             = "WAFV2"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "WAFV2"
    human_friendly      = "WAF"
  }

  client {
    go_v1_client_typename = "WAFV2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRuleGroups"
    endpoint_api_params      = "Scope: awstypes.ScopeRegional"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_wafv2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["wafv2_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "waf" {

  sdk {
    id             = "WAF"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "WAF"
    human_friendly      = "WAF Classic"
  }

  client {
    go_v1_client_typename = "WAF"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRules"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_waf_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["waf_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "WAFRegional"
    human_friendly      = "WAF Classic Regional"
  }

  client {
    go_v1_client_typename = "WAFRegional"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListRules"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_wafregional_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["wafregional_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "budgets" {

  sdk {
    id             = "Budgets"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "Budgets"
    human_friendly      = "Web Services Budgets"
  }

  client {
    go_v1_client_typename = "Budgets"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeBudgets"
    endpoint_api_params      = "AccountId: aws_sdkv2.String(\"012345678901\")"
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_budgets_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["budgets_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "wellarchitected" {

  sdk {
    id             = "WellArchitected"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "WellArchitected"
    human_friendly      = "Well-Architected Tool"
  }

  client {
    go_v1_client_typename = "WellArchitected"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListProfiles"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_wellarchitected_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["wellarchitected_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "workdocs" {

  sdk {
    id             = "WorkDocs"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "WorkDocs"
    human_friendly      = "WorkDocs"
  }

  client {
    go_v1_client_typename = "WorkDocs"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_workdocs_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["workdocs_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "worklink" {

  sdk {
    id             = "WorkLink"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "WorkLink"
    human_friendly      = "WorkLink"
  }

  client {
    go_v1_client_typename = "WorkLink"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListFleets"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_worklink_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["worklink_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "workmail" {

  sdk {
    id             = "WorkMail"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "WorkMail"
    human_friendly      = "WorkMail"
  }

  client {
    go_v1_client_typename = "WorkMail"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_workmail_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["workmail_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "workmailmessageflow" {

  sdk {
    id             = "WorkMailMessageFlow"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "WorkMailMessageFlow"
    human_friendly      = "WorkMail Message Flow"
  }

  client {
    go_v1_client_typename = "WorkMailMessageFlow"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = ""
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_workmailmessageflow_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["workmailmessageflow_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = true
  allowed_subcategory = false
  note                = ""
}
service "workspaces" {

  sdk {
    id             = "WorkSpaces"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "WorkSpaces"
    human_friendly      = "WorkSpaces"
  }

  client {
    go_v1_client_typename = "WorkSpaces"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeWorkspaces"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_workspaces_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["workspaces_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
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
    aliases             = [""]
    provider_name_upper = "WorkSpacesWeb"
    human_friendly      = "WorkSpaces Web"
  }

  client {
    go_v1_client_typename = "WorkSpacesWeb"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPortals"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_workspacesweb_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["workspacesweb_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "xray" {

  sdk {
    id             = "XRay"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "XRay"
    human_friendly      = "X-Ray"
  }

  client {
    go_v1_client_typename = "XRay"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListResourcePolicies"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_xray_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["xray_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "verifiedpermissions" {

  sdk {
    id             = "VerifiedPermissions"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "VerifiedPermissions"
    human_friendly      = "Verified Permissions"
  }

  client {
    go_v1_client_typename = "VerifiedPermissions"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListPolicyStores"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_verifiedpermissions_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["verifiedpermissions_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "codecatalyst" {

  sdk {
    id             = "CodeCatalyst"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "CodeCatalyst"
    human_friendly      = "CodeCatalyst"
  }

  client {
    go_v1_client_typename = "CodeCatalyst"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListAccessTokens"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_codecatalyst_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["codecatalyst_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "mediapackagev2" {

  sdk {
    id             = "MediaPackageV2"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "MediaPackageV2"
    human_friendly      = "Elemental MediaPackage Version 2"
  }

  client {
    go_v1_client_typename = "MediaPackageV2"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "ListChannelGroups"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_media_packagev2_"
    correct = "aws_mediapackagev2_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["media_packagev2_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "iot" {

  sdk {
    id             = "IoT"
    client_version = [1]
  }

  names {
    aliases             = [""]
    provider_name_upper = "IoT"
    human_friendly      = "IoT Core"
  }

  client {
    go_v1_client_typename = "IoT"
    skip_client_generate  = false
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeDefaultAuthorizer"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_iot_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["iot_"]
  brand               = "AWS"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "dynamodb" {

  sdk {
    id             = "DynamoDB"
    client_version = [2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "DynamoDB"
    human_friendly      = "DynamoDB"
  }

  client {
    go_v1_client_typename = "DynamoDB"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = "AWS_DYNAMODB_ENDPOINT"
    tf_aws_env_var     = "TF_AWS_DYNAMODB_ENDPOINT"
  }

  endpoint_info {
    endpoint_api_call        = "ListTables"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = ""
    correct = "aws_dynamodb_"
  }
  split_package       = ""
  file_prefix         = ""
  doc_prefix          = ["dynamodb_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}
service "ec2" {

  sdk {
    id             = "EC2"
    client_version = [1, 2]
  }

  names {
    aliases             = [""]
    provider_name_upper = "EC2"
    human_friendly      = "EC2 (Elastic Compute Cloud)"
  }

  client {
    go_v1_client_typename = "EC2"
    skip_client_generate  = true
  }

  env_var {
    deprecated_env_var = ""
    tf_aws_env_var     = ""
  }

  endpoint_info {
    endpoint_api_call        = "DescribeVpcs"
    endpoint_api_params      = ""
    endpoint_region_override = ""
    endpoint_only            = false
  }

  resource_prefix {
    actual  = "aws_(ami|availability_zone|ec2_(availability|capacity|fleet|host|instance|public_ipv4_pool|serial|spot|tag)|eip|instance|key_pair|launch_template|placement_group|spot)"
    correct = "aws_ec2_"
  }

  sub_service "ec2ebs" {

    cli_v2_command {
      aws_cli_v2_command           = ""
      aws_cli_v2_command_no_dashes = ""
    }

    go_packages {
      v1_package = ""
      v2_package = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
      aliases             = [""]
      provider_name_upper = "EC2EBS"
      human_friendly      = "EBS (EC2)"
    }

    client {
      go_v1_client_typename = ""
      skip_client_generate  = false
    }

    env_var {
      deprecated_env_var = ""
      tf_aws_env_var     = ""
    }

    endpoint_info {
      endpoint_api_call        = ""
      endpoint_api_params      = ""
      endpoint_region_override = ""
      endpoint_only            = false
    }

    resource_prefix {
      actual  = "aws_(ebs_|volume_attach|snapshot_create)"
      correct = "aws_ec2ebs_"
    }
    split_package       = "ec2"
    file_prefix         = "ebs_"
    doc_prefix          = ["ebs_", "volume_attachment", "snapshot_"]
    brand               = "Amazon"
    exclude             = true
    not_implemented     = false
    allowed_subcategory = true
    note                = "Part of EC2"
  }
  sub_service "ec2outposts" {

    cli_v2_command {
      aws_cli_v2_command           = ""
      aws_cli_v2_command_no_dashes = ""
    }

    go_packages {
      v1_package = ""
      v2_package = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
      aliases             = [""]
      provider_name_upper = "EC2Outposts"
      human_friendly      = "Outposts (EC2)"
    }

    client {
      go_v1_client_typename = ""
      skip_client_generate  = false
    }

    env_var {
      deprecated_env_var = ""
      tf_aws_env_var     = ""
    }

    endpoint_info {
      endpoint_api_call        = ""
      endpoint_api_params      = ""
      endpoint_region_override = ""
      endpoint_only            = false
    }

    resource_prefix {
      actual  = "aws_ec2_(coip_pool|local_gateway)"
      correct = "aws_ec2outposts_"
    }
    split_package       = "ec2"
    file_prefix         = "outposts_"
    doc_prefix          = ["ec2_coip_pool", "ec2_local_gateway"]
    brand               = "AWS"
    exclude             = true
    not_implemented     = false
    allowed_subcategory = true
    note                = "Part of EC2"
  }
  sub_service "transitgateway" {

    cli_v2_command {
      aws_cli_v2_command           = ""
      aws_cli_v2_command_no_dashes = ""
    }

    go_packages {
      v1_package = ""
      v2_package = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
      aliases             = [""]
      provider_name_upper = "TransitGateway"
      human_friendly      = "Transit Gateway"
    }

    client {
      go_v1_client_typename = ""
      skip_client_generate  = false
    }

    env_var {
      deprecated_env_var = ""
      tf_aws_env_var     = ""
    }

    endpoint_info {
      endpoint_api_call        = ""
      endpoint_api_params      = ""
      endpoint_region_override = ""
      endpoint_only            = false
    }

    resource_prefix {
      actual  = "aws_ec2_transit_gateway"
      correct = "aws_transitgateway_"
    }
    split_package       = "ec2"
    file_prefix         = "transitgateway_"
    doc_prefix          = ["ec2_transit_gateway"]
    brand               = "AWS"
    exclude             = true
    not_implemented     = false
    allowed_subcategory = true
    note                = "Part of EC2"
  }
  sub_service "verifiedaccess" {

    cli_v2_command {
      aws_cli_v2_command           = ""
      aws_cli_v2_command_no_dashes = ""
    }

    go_packages {
      v1_package = ""
      v2_package = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
      aliases             = [""]
      provider_name_upper = "VerifiedAccess"
      human_friendly      = "Verified Access"
    }

    client {
      go_v1_client_typename = ""
      skip_client_generate  = false
    }

    env_var {
      deprecated_env_var = ""
      tf_aws_env_var     = ""
    }

    endpoint_info {
      endpoint_api_call        = ""
      endpoint_api_params      = ""
      endpoint_region_override = ""
      endpoint_only            = false
    }

    resource_prefix {
      actual  = "aws_verifiedaccess"
      correct = "aws_verifiedaccess_"
    }
    split_package       = "ec2"
    file_prefix         = "verifiedaccess_"
    doc_prefix          = ["verifiedaccess_"]
    brand               = "AWS"
    exclude             = true
    not_implemented     = false
    allowed_subcategory = true
    note                = "Part of EC2"
  }
  sub_service "vpc" {

    cli_v2_command {
      aws_cli_v2_command           = ""
      aws_cli_v2_command_no_dashes = ""
    }

    go_packages {
      v1_package = ""
      v2_package = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
      aliases             = [""]
      provider_name_upper = "VPC"
      human_friendly      = "VPC (Virtual Private Cloud)"
    }

    client {
      go_v1_client_typename = ""
      skip_client_generate  = false
    }

    env_var {
      deprecated_env_var = ""
      tf_aws_env_var     = ""
    }

    endpoint_info {
      endpoint_api_call        = ""
      endpoint_api_params      = ""
      endpoint_region_override = ""
      endpoint_only            = false
    }

    resource_prefix {
      actual  = "aws_((default_)?(network_acl|route_table|security_group|subnet|vpc(?!_ipam))|ec2_(managed|network|subnet|traffic)|egress_only_internet|flow_log|internet_gateway|main_route_table_association|nat_gateway|network_interface|prefix_list|route\\b)"
      correct = "aws_vpc_"
    }
    split_package       = "ec2"
    file_prefix         = "vpc_"
    doc_prefix          = ["default_network_", "default_route_", "default_security_", "default_subnet", "default_vpc", "ec2_managed_", "ec2_network_", "ec2_subnet_", "ec2_traffic_", "egress_only_", "flow_log", "internet_gateway", "main_route_", "nat_", "network_", "prefix_list", "route_", "route\\.", "security_group", "subnet", "vpc_dhcp_", "vpc_endpoint", "vpc_ipv", "vpc_network_performance", "vpc_peering_", "vpc_security_group_", "vpc\\.", "vpcs\\."]
    brand               = "Amazon"
    exclude             = true
    not_implemented     = false
    allowed_subcategory = true
    note                = "Part of EC2"
  }
  sub_service "ipam" {

    cli_v2_command {
      aws_cli_v2_command           = ""
      aws_cli_v2_command_no_dashes = ""
    }

    go_packages {
      v1_package = ""
      v2_package = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
      aliases             = [""]
      provider_name_upper = "IPAM"
      human_friendly      = "VPC IPAM (IP Address Manager)"
    }

    client {
      go_v1_client_typename = ""
      skip_client_generate  = false
    }

    env_var {
      deprecated_env_var = ""
      tf_aws_env_var     = ""
    }

    endpoint_info {
      endpoint_api_call        = ""
      endpoint_api_params      = ""
      endpoint_region_override = ""
      endpoint_only            = false
    }

    resource_prefix {
      actual  = "aws_vpc_ipam"
      correct = "aws_ipam_"
    }
    split_package       = "ec2"
    file_prefix         = "ipam_"
    doc_prefix          = ["vpc_ipam"]
    brand               = "Amazon"
    exclude             = true
    not_implemented     = false
    allowed_subcategory = true
    note                = "Part of EC2"
  }
  sub_service "vpnclient" {

    cli_v2_command {
      aws_cli_v2_command           = ""
      aws_cli_v2_command_no_dashes = ""
    }

    go_packages {
      v1_package = ""
      v2_package = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
      aliases             = [""]
      provider_name_upper = "ClientVPN"
      human_friendly      = "VPN (Client)"
    }

    client {
      go_v1_client_typename = ""
      skip_client_generate  = false
    }

    env_var {
      deprecated_env_var = ""
      tf_aws_env_var     = ""
    }

    endpoint_info {
      endpoint_api_call        = ""
      endpoint_api_params      = ""
      endpoint_region_override = ""
      endpoint_only            = false
    }

    resource_prefix {
      actual  = "aws_ec2_client_vpn"
      correct = "aws_vpnclient_"
    }
    split_package       = "ec2"
    file_prefix         = "vpnclient_"
    doc_prefix          = ["ec2_client_vpn_"]
    brand               = "AWS"
    exclude             = true
    not_implemented     = false
    allowed_subcategory = true
    note                = "Part of EC2"
  }
  sub_service "vpnsite" {

    cli_v2_command {
      aws_cli_v2_command           = ""
      aws_cli_v2_command_no_dashes = ""
    }

    go_packages {
      v1_package = ""
      v2_package = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
      aliases             = [""]
      provider_name_upper = "SiteVPN"
      human_friendly      = "VPN (Site-to-Site)"
    }

    client {
      go_v1_client_typename = ""
      skip_client_generate  = false
    }

    env_var {
      deprecated_env_var = ""
      tf_aws_env_var     = ""
    }

    endpoint_info {
      endpoint_api_call        = ""
      endpoint_api_params      = ""
      endpoint_region_override = ""
      endpoint_only            = false
    }

    resource_prefix {
      actual  = "aws_(customer_gateway|vpn_)"
      correct = "aws_vpnsite_"
    }
    split_package       = "ec2"
    file_prefix         = "vpnsite_"
    doc_prefix          = ["customer_gateway", "vpn_"]
    brand               = "AWS"
    exclude             = true
    not_implemented     = false
    allowed_subcategory = true
    note                = "Part of EC2"
  }
  sub_service "wavelength" {

    cli_v2_command {
      aws_cli_v2_command           = ""
      aws_cli_v2_command_no_dashes = ""
    }

    go_packages {
      v1_package = ""
      v2_package = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
      aliases             = [""]
      provider_name_upper = "Wavelength"
      human_friendly      = "Wavelength"
    }

    client {
      go_v1_client_typename = ""
      skip_client_generate  = false
    }

    env_var {
      deprecated_env_var = ""
      tf_aws_env_var     = ""
    }

    endpoint_info {
      endpoint_api_call        = ""
      endpoint_api_params      = ""
      endpoint_region_override = ""
      endpoint_only            = false
    }

    resource_prefix {
      actual  = "aws_ec2_carrier_gateway"
      correct = "aws_wavelength_"
    }
    split_package       = "ec2"
    file_prefix         = "wavelength_"
    doc_prefix          = ["ec2_carrier_"]
    brand               = "AWS"
    exclude             = true
    not_implemented     = false
    allowed_subcategory = true
    note                = "Part of EC2"
  }

  split_package       = "ec2"
  file_prefix         = "ec2_"
  doc_prefix          = ["ami", "availability_zone", "ec2_availability_", "ec2_capacity_", "ec2_fleet", "ec2_host", "ec2_image_", "ec2_instance_", "ec2_public_ipv4_pool", "ec2_serial_", "ec2_spot_", "ec2_tag", "eip", "instance", "key_pair", "launch_template", "placement_group", "spot_"]
  brand               = "Amazon"
  exclude             = false
  not_implemented     = false
  allowed_subcategory = false
  note                = ""
}

