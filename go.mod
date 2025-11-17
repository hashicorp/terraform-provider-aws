module github.com/hashicorp/terraform-provider-aws

go 1.24.10

// Disable post-quantum X25519MLKEM768 key exchange mechanism
// This causes errors with AWS Network Firewall
godebug tlsmlkem=0

require (
	github.com/ProtonMail/go-crypto v1.3.0
	github.com/YakDriver/go-version v0.1.0
	github.com/YakDriver/regexache v0.25.0
	github.com/YakDriver/smarterr v0.8.0
	github.com/aws/aws-sdk-go-v2 v1.40.0
	github.com/aws/aws-sdk-go-v2/config v1.32.2
	github.com/aws/aws-sdk-go-v2/credentials v1.19.2
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.14
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.20.12
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.45.4
	github.com/aws/aws-sdk-go-v2/service/account v1.29.6
	github.com/aws/aws-sdk-go-v2/service/acm v1.37.15
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.46.4
	github.com/aws/aws-sdk-go-v2/service/amp v1.42.2
	github.com/aws/aws-sdk-go-v2/service/amplify v1.38.7
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.38.1
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.33.2
	github.com/aws/aws-sdk-go-v2/service/appconfig v1.43.5
	github.com/aws/aws-sdk-go-v2/service/appfabric v1.16.14
	github.com/aws/aws-sdk-go-v2/service/appflow v1.51.5
	github.com/aws/aws-sdk-go-v2/service/appintegrations v1.36.14
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.41.6
	github.com/aws/aws-sdk-go-v2/service/applicationinsights v1.34.13
	github.com/aws/aws-sdk-go-v2/service/applicationsignals v1.18.1
	github.com/aws/aws-sdk-go-v2/service/appmesh v1.35.5
	github.com/aws/aws-sdk-go-v2/service/apprunner v1.39.7
	github.com/aws/aws-sdk-go-v2/service/appstream v1.52.3
	github.com/aws/aws-sdk-go-v2/service/appsync v1.52.5
	github.com/aws/aws-sdk-go-v2/service/arcregionswitch v1.2.16
	github.com/aws/aws-sdk-go-v2/service/athena v1.56.1
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.46.5
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.62.1
	github.com/aws/aws-sdk-go-v2/service/autoscalingplans v1.30.7
	github.com/aws/aws-sdk-go-v2/service/backup v1.54.2
	github.com/aws/aws-sdk-go-v2/service/batch v1.58.8
	github.com/aws/aws-sdk-go-v2/service/bcmdataexports v1.12.7
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.50.1
	github.com/aws/aws-sdk-go-v2/service/bedrockagent v1.51.5
	github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol v1.14.1
	github.com/aws/aws-sdk-go-v2/service/billing v1.9.2
	github.com/aws/aws-sdk-go-v2/service/budgets v1.42.1
	github.com/aws/aws-sdk-go-v2/service/chatbot v1.14.14
	github.com/aws/aws-sdk-go-v2/service/chime v1.41.5
	github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines v1.26.14
	github.com/aws/aws-sdk-go-v2/service/chimesdkvoice v1.28.6
	github.com/aws/aws-sdk-go-v2/service/cleanrooms v1.37.5
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.33.13
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.29.5
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.71.1
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.58.1
	github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore v1.12.16
	github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 v1.34.13
	github.com/aws/aws-sdk-go-v2/service/cloudsearch v1.32.5
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.55.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.52.5
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.61.1
	github.com/aws/aws-sdk-go-v2/service/codeartifact v1.38.14
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.68.6
	github.com/aws/aws-sdk-go-v2/service/codecatalyst v1.21.5
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.33.5
	github.com/aws/aws-sdk-go-v2/service/codeconnections v1.10.13
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.35.5
	github.com/aws/aws-sdk-go-v2/service/codeguruprofiler v1.29.13
	github.com/aws/aws-sdk-go-v2/service/codegurureviewer v1.34.13
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.46.14
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.35.6
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.31.14
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.33.14
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.57.15
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.40.14
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.48.5
	github.com/aws/aws-sdk-go-v2/service/configservice v1.59.6
	github.com/aws/aws-sdk-go-v2/service/connect v1.148.1
	github.com/aws/aws-sdk-go-v2/service/connectcases v1.34.5
	github.com/aws/aws-sdk-go-v2/service/controltower v1.28.1
	github.com/aws/aws-sdk-go-v2/service/costandusagereportservice v1.34.6
	github.com/aws/aws-sdk-go-v2/service/costexplorer v1.60.2
	github.com/aws/aws-sdk-go-v2/service/costoptimizationhub v1.21.2
	github.com/aws/aws-sdk-go-v2/service/customerprofiles v1.54.7
	github.com/aws/aws-sdk-go-v2/service/databasemigrationservice v1.61.1
	github.com/aws/aws-sdk-go-v2/service/databrew v1.39.7
	github.com/aws/aws-sdk-go-v2/service/dataexchange v1.40.7
	github.com/aws/aws-sdk-go-v2/service/datapipeline v1.30.13
	github.com/aws/aws-sdk-go-v2/service/datasync v1.56.1
	github.com/aws/aws-sdk-go-v2/service/datazone v1.48.2
	github.com/aws/aws-sdk-go-v2/service/dax v1.29.9
	github.com/aws/aws-sdk-go-v2/service/detective v1.38.6
	github.com/aws/aws-sdk-go-v2/service/devicefarm v1.38.1
	github.com/aws/aws-sdk-go-v2/service/devopsguru v1.40.5
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.38.7
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.38.8
	github.com/aws/aws-sdk-go-v2/service/dlm v1.35.8
	github.com/aws/aws-sdk-go-v2/service/docdb v1.48.5
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.20.6
	github.com/aws/aws-sdk-go-v2/service/drs v1.36.6
	github.com/aws/aws-sdk-go-v2/service/dsql v1.12.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.53.2
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.275.0
	github.com/aws/aws-sdk-go-v2/service/ecr v1.54.1
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.38.6
	github.com/aws/aws-sdk-go-v2/service/ecs v1.69.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.41.6
	github.com/aws/aws-sdk-go-v2/service/eks v1.75.1
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.51.5
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.33.15
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.33.15
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.54.2
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.37.15
	github.com/aws/aws-sdk-go-v2/service/elastictranscoder v1.32.14
	github.com/aws/aws-sdk-go-v2/service/emr v1.57.1
	github.com/aws/aws-sdk-go-v2/service/emrcontainers v1.40.10
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.37.5
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.45.14
	github.com/aws/aws-sdk-go-v2/service/evidently v1.28.13
	github.com/aws/aws-sdk-go-v2/service/evs v1.5.10
	github.com/aws/aws-sdk-go-v2/service/finspace v1.33.14
	github.com/aws/aws-sdk-go-v2/service/firehose v1.42.5
	github.com/aws/aws-sdk-go-v2/service/fis v1.37.13
	github.com/aws/aws-sdk-go-v2/service/fms v1.44.14
	github.com/aws/aws-sdk-go-v2/service/fsx v1.64.2
	github.com/aws/aws-sdk-go-v2/service/gamelift v1.48.4
	github.com/aws/aws-sdk-go-v2/service/glacier v1.31.14
	github.com/aws/aws-sdk-go-v2/service/globalaccelerator v1.35.7
	github.com/aws/aws-sdk-go-v2/service/glue v1.134.1
	github.com/aws/aws-sdk-go-v2/service/grafana v1.32.7
	github.com/aws/aws-sdk-go-v2/service/greengrass v1.32.14
	github.com/aws/aws-sdk-go-v2/service/groundstation v1.39.4
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.68.2
	github.com/aws/aws-sdk-go-v2/service/healthlake v1.36.6
	github.com/aws/aws-sdk-go-v2/service/iam v1.52.2
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.34.5
	github.com/aws/aws-sdk-go-v2/service/imagebuilder v1.50.1
	github.com/aws/aws-sdk-go-v2/service/inspector v1.30.13
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.45.2
	github.com/aws/aws-sdk-go-v2/service/internetmonitor v1.26.6
	github.com/aws/aws-sdk-go-v2/service/invoicing v1.9.1
	github.com/aws/aws-sdk-go-v2/service/iot v1.69.13
	github.com/aws/aws-sdk-go-v2/service/ivs v1.48.7
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.21.13
	github.com/aws/aws-sdk-go-v2/service/kafka v1.46.2
	github.com/aws/aws-sdk-go-v2/service/kafkaconnect v1.27.14
	github.com/aws/aws-sdk-go-v2/service/kendra v1.60.14
	github.com/aws/aws-sdk-go-v2/service/keyspaces v1.24.7
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.42.6
	github.com/aws/aws-sdk-go-v2/service/kinesisanalytics v1.30.14
	github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2 v1.36.15
	github.com/aws/aws-sdk-go-v2/service/kinesisvideo v1.33.1
	github.com/aws/aws-sdk-go-v2/service/kms v1.49.1
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.46.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.83.1
	github.com/aws/aws-sdk-go-v2/service/launchwizard v1.13.14
	github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice v1.34.7
	github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 v1.58.1
	github.com/aws/aws-sdk-go-v2/service/licensemanager v1.37.1
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.50.8
	github.com/aws/aws-sdk-go-v2/service/location v1.50.6
	github.com/aws/aws-sdk-go-v2/service/m2 v1.26.7
	github.com/aws/aws-sdk-go-v2/service/macie2 v1.50.6
	github.com/aws/aws-sdk-go-v2/service/mediaconnect v1.46.2
	github.com/aws/aws-sdk-go-v2/service/mediaconvert v1.85.2
	github.com/aws/aws-sdk-go-v2/service/medialive v1.87.2
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.39.14
	github.com/aws/aws-sdk-go-v2/service/mediapackagev2 v1.34.1
	github.com/aws/aws-sdk-go-v2/service/mediapackagevod v1.39.14
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.29.14
	github.com/aws/aws-sdk-go-v2/service/memorydb v1.33.7
	github.com/aws/aws-sdk-go-v2/service/mgn v1.38.6
	github.com/aws/aws-sdk-go-v2/service/mq v1.34.12
	github.com/aws/aws-sdk-go-v2/service/mwaa v1.39.14
	github.com/aws/aws-sdk-go-v2/service/mwaaserverless v1.0.2
	github.com/aws/aws-sdk-go-v2/service/neptune v1.43.5
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.21.13
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.59.0
	github.com/aws/aws-sdk-go-v2/service/networkflowmonitor v1.11.2
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.41.1
	github.com/aws/aws-sdk-go-v2/service/networkmonitor v1.13.6
	github.com/aws/aws-sdk-go-v2/service/notifications v1.7.12
	github.com/aws/aws-sdk-go-v2/service/notificationscontacts v1.5.16
	github.com/aws/aws-sdk-go-v2/service/oam v1.23.7
	github.com/aws/aws-sdk-go-v2/service/observabilityadmin v1.8.8
	github.com/aws/aws-sdk-go-v2/service/odb v1.6.1
	github.com/aws/aws-sdk-go-v2/service/opensearch v1.54.2
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.27.6
	github.com/aws/aws-sdk-go-v2/service/organizations v1.49.0
	github.com/aws/aws-sdk-go-v2/service/osis v1.21.7
	github.com/aws/aws-sdk-go-v2/service/outposts v1.57.8
	github.com/aws/aws-sdk-go-v2/service/paymentcryptography v1.26.5
	github.com/aws/aws-sdk-go-v2/service/pcaconnectorad v1.15.14
	github.com/aws/aws-sdk-go-v2/service/pcs v1.15.2
	github.com/aws/aws-sdk-go-v2/service/pinpoint v1.39.14
	github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2 v1.26.4
	github.com/aws/aws-sdk-go-v2/service/pipes v1.23.13
	github.com/aws/aws-sdk-go-v2/service/polly v1.54.7
	github.com/aws/aws-sdk-go-v2/service/pricing v1.40.7
	github.com/aws/aws-sdk-go-v2/service/qbusiness v1.33.14
	github.com/aws/aws-sdk-go-v2/service/qldb v1.32.2
	github.com/aws/aws-sdk-go-v2/service/quicksight v1.98.1
	github.com/aws/aws-sdk-go-v2/service/ram v1.34.15
	github.com/aws/aws-sdk-go-v2/service/rbin v1.27.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.111.1
	github.com/aws/aws-sdk-go-v2/service/rdsdata v1.32.14
	github.com/aws/aws-sdk-go-v2/service/redshift v1.61.1
	github.com/aws/aws-sdk-go-v2/service/redshiftdata v1.38.1
	github.com/aws/aws-sdk-go-v2/service/redshiftserverless v1.32.1
	github.com/aws/aws-sdk-go-v2/service/rekognition v1.51.13
	github.com/aws/aws-sdk-go-v2/service/resiliencehub v1.35.6
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.22.8
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.33.16
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.31.2
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.21.14
	github.com/aws/aws-sdk-go-v2/service/route53 v1.61.0
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.34.12
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.9.14
	github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig v1.32.7
	github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness v1.26.14
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.41.2
	github.com/aws/aws-sdk-go-v2/service/rum v1.30.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.92.1
	github.com/aws/aws-sdk-go-v2/service/s3control v1.66.11
	github.com/aws/aws-sdk-go-v2/service/s3outposts v1.34.5
	github.com/aws/aws-sdk-go-v2/service/s3tables v1.12.2
	github.com/aws/aws-sdk-go-v2/service/s3vectors v1.5.4
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.226.1
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.17.14
	github.com/aws/aws-sdk-go-v2/service/schemas v1.34.5
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.40.2
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.66.1
	github.com/aws/aws-sdk-go-v2/service/securitylake v1.25.6
	github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository v1.30.5
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.39.5
	github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry v1.35.14
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.39.18
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.33.9
	github.com/aws/aws-sdk-go-v2/service/ses v1.34.14
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.55.1
	github.com/aws/aws-sdk-go-v2/service/sfn v1.40.2
	github.com/aws/aws-sdk-go-v2/service/shield v1.34.14
	github.com/aws/aws-sdk-go-v2/service/signer v1.31.14
	github.com/aws/aws-sdk-go-v2/service/sns v1.39.7
	github.com/aws/aws-sdk-go-v2/service/sqs v1.42.17
	github.com/aws/aws-sdk-go-v2/service/ssm v1.67.4
	github.com/aws/aws-sdk-go-v2/service/ssmcontacts v1.31.7
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.39.13
	github.com/aws/aws-sdk-go-v2/service/ssmquicksetup v1.8.14
	github.com/aws/aws-sdk-go-v2/service/ssmsap v1.25.13
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.5
	github.com/aws/aws-sdk-go-v2/service/ssoadmin v1.36.10
	github.com/aws/aws-sdk-go-v2/service/storagegateway v1.43.6
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.2
	github.com/aws/aws-sdk-go-v2/service/swf v1.33.8
	github.com/aws/aws-sdk-go-v2/service/synthetics v1.42.6
	github.com/aws/aws-sdk-go-v2/service/taxsettings v1.16.14
	github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb v1.17.8
	github.com/aws/aws-sdk-go-v2/service/timestreamquery v1.36.7
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.35.13
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.53.8
	github.com/aws/aws-sdk-go-v2/service/transfer v1.68.1
	github.com/aws/aws-sdk-go-v2/service/verifiedpermissions v1.30.4
	github.com/aws/aws-sdk-go-v2/service/vpclattice v1.20.4
	github.com/aws/aws-sdk-go-v2/service/waf v1.30.13
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.30.14
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.70.2
	github.com/aws/aws-sdk-go-v2/service/wellarchitected v1.39.14
	github.com/aws/aws-sdk-go-v2/service/workmail v1.36.12
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.64.6
	github.com/aws/aws-sdk-go-v2/service/workspacesweb v1.34.2
	github.com/aws/aws-sdk-go-v2/service/xray v1.36.13
	github.com/aws/smithy-go v1.23.2
	github.com/beevik/etree v1.6.0
	github.com/cedar-policy/cedar-go v1.3.0
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dlclark/regexp2 v1.11.5
	github.com/gertd/go-pluralize v0.2.1
	github.com/goccy/go-yaml v1.18.0
	github.com/google/go-cmp v0.7.0
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.23.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.68
	github.com/hashicorp/awspolicyequivalence v1.7.0
	github.com/hashicorp/cli v1.1.7
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.5.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-set/v3 v3.0.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.7.0
	github.com/hashicorp/hcl/v2 v2.24.0
	github.com/hashicorp/terraform-json v0.27.2
	github.com/hashicorp/terraform-plugin-framework v1.16.1
	github.com/hashicorp/terraform-plugin-framework-jsontypes v0.2.0
	github.com/hashicorp/terraform-plugin-framework-timeouts v0.7.0
	github.com/hashicorp/terraform-plugin-framework-timetypes v0.5.0
	github.com/hashicorp/terraform-plugin-framework-validators v0.19.0
	github.com/hashicorp/terraform-plugin-go v0.29.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-mux v0.21.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.38.1
	github.com/hashicorp/terraform-plugin-testing v1.14.0-beta.1.0.20251013071646-7ed2ee242705
	github.com/jaswdr/faker/v2 v2.9.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20240118010651-0ba75a80ca38
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pquerna/otp v1.5.0
	github.com/shopspring/decimal v1.4.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.63.0
	go.opentelemetry.io/otel v1.38.0
	golang.org/x/crypto v0.45.0
	golang.org/x/text v0.31.0
	golang.org/x/tools v0.39.0
	gopkg.in/dnaeon/go-vcr.v4 v4.0.6
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.10 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-plugin v1.7.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/hc-install v0.9.2 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.24.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.4.0 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/posener/complete v1.2.3 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.17.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.3 // indirect
	golang.org/x/exp v0.0.0-20220921023135-46d9e7742f1e // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)

replace github.com/hashicorp/terraform-plugin-log => github.com/gdavison/terraform-plugin-log v0.0.0-20230928191232-6c653d8ef8fb

replace github.com/aws/aws-sdk-go-v2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2

replace github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/aws/protocol/eventstream

replace github.com/aws/aws-sdk-go-v2/config => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/config

replace github.com/aws/aws-sdk-go-v2/credentials => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/credentials

replace github.com/aws/aws-sdk-go-v2/feature/ec2/imds => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/feature/ec2/imds

replace github.com/aws/aws-sdk-go-v2/feature/s3/manager => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/feature/s3/manager

replace github.com/aws/aws-sdk-go-v2/internal/configsources => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/internal/configsources

replace github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/internal/endpoints/v2

replace github.com/aws/aws-sdk-go-v2/internal/ini => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/internal/ini

replace github.com/aws/aws-sdk-go-v2/internal/v4a => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/internal/v4a

replace github.com/aws/aws-sdk-go-v2/service/accessanalyzer => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/accessanalyzer

replace github.com/aws/aws-sdk-go-v2/service/account => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/account

replace github.com/aws/aws-sdk-go-v2/service/acm => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/acm

replace github.com/aws/aws-sdk-go-v2/service/acmpca => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/acmpca

replace github.com/aws/aws-sdk-go-v2/service/amp => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/amp

replace github.com/aws/aws-sdk-go-v2/service/amplify => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/amplify

replace github.com/aws/aws-sdk-go-v2/service/apigateway => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/apigateway

replace github.com/aws/aws-sdk-go-v2/service/apigatewayv2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/apigatewayv2

replace github.com/aws/aws-sdk-go-v2/service/appconfig => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/appconfig

replace github.com/aws/aws-sdk-go-v2/service/appfabric => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/appfabric

replace github.com/aws/aws-sdk-go-v2/service/appflow => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/appflow

replace github.com/aws/aws-sdk-go-v2/service/appintegrations => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/appintegrations

replace github.com/aws/aws-sdk-go-v2/service/applicationautoscaling => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/applicationautoscaling

replace github.com/aws/aws-sdk-go-v2/service/applicationinsights => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/applicationinsights

replace github.com/aws/aws-sdk-go-v2/service/applicationsignals => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/applicationsignals

replace github.com/aws/aws-sdk-go-v2/service/appmesh => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/appmesh

replace github.com/aws/aws-sdk-go-v2/service/apprunner => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/apprunner

replace github.com/aws/aws-sdk-go-v2/service/appstream => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/appstream

replace github.com/aws/aws-sdk-go-v2/service/appsync => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/appsync

replace github.com/aws/aws-sdk-go-v2/service/arcregionswitch => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/arcregionswitch

replace github.com/aws/aws-sdk-go-v2/service/athena => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/athena

replace github.com/aws/aws-sdk-go-v2/service/auditmanager => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/auditmanager

replace github.com/aws/aws-sdk-go-v2/service/autoscaling => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/autoscaling

replace github.com/aws/aws-sdk-go-v2/service/autoscalingplans => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/autoscalingplans

replace github.com/aws/aws-sdk-go-v2/service/backup => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/backup

replace github.com/aws/aws-sdk-go-v2/service/batch => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/batch

replace github.com/aws/aws-sdk-go-v2/service/bcmdataexports => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/bcmdataexports

replace github.com/aws/aws-sdk-go-v2/service/bedrock => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/bedrock

replace github.com/aws/aws-sdk-go-v2/service/bedrockagent => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/bedrockagent

replace github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/bedrockagentcorecontrol

replace github.com/aws/aws-sdk-go-v2/service/billing => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/billing

replace github.com/aws/aws-sdk-go-v2/service/budgets => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/budgets

replace github.com/aws/aws-sdk-go-v2/service/chatbot => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/chatbot

replace github.com/aws/aws-sdk-go-v2/service/chime => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/chime

replace github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/chimesdkmediapipelines

replace github.com/aws/aws-sdk-go-v2/service/chimesdkvoice => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/chimesdkvoice

replace github.com/aws/aws-sdk-go-v2/service/cleanrooms => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cleanrooms

replace github.com/aws/aws-sdk-go-v2/service/cloud9 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloud9

replace github.com/aws/aws-sdk-go-v2/service/cloudcontrol => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloudcontrol

replace github.com/aws/aws-sdk-go-v2/service/cloudformation => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloudformation

replace github.com/aws/aws-sdk-go-v2/service/cloudfront => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloudfront

replace github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore

replace github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloudhsmv2

replace github.com/aws/aws-sdk-go-v2/service/cloudsearch => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloudsearch

replace github.com/aws/aws-sdk-go-v2/service/cloudtrail => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloudtrail

replace github.com/aws/aws-sdk-go-v2/service/cloudwatch => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloudwatch

replace github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cloudwatchlogs

replace github.com/aws/aws-sdk-go-v2/service/codeartifact => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codeartifact

replace github.com/aws/aws-sdk-go-v2/service/codebuild => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codebuild

replace github.com/aws/aws-sdk-go-v2/service/codecatalyst => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codecatalyst

replace github.com/aws/aws-sdk-go-v2/service/codecommit => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codecommit

replace github.com/aws/aws-sdk-go-v2/service/codeconnections => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codeconnections

replace github.com/aws/aws-sdk-go-v2/service/codedeploy => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codedeploy

replace github.com/aws/aws-sdk-go-v2/service/codeguruprofiler => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codeguruprofiler

replace github.com/aws/aws-sdk-go-v2/service/codegurureviewer => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codegurureviewer

replace github.com/aws/aws-sdk-go-v2/service/codepipeline => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codepipeline

replace github.com/aws/aws-sdk-go-v2/service/codestarconnections => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codestarconnections

replace github.com/aws/aws-sdk-go-v2/service/codestarnotifications => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/codestarnotifications

replace github.com/aws/aws-sdk-go-v2/service/cognitoidentity => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cognitoidentity

replace github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/cognitoidentityprovider

replace github.com/aws/aws-sdk-go-v2/service/comprehend => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/comprehend

replace github.com/aws/aws-sdk-go-v2/service/computeoptimizer => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/computeoptimizer

replace github.com/aws/aws-sdk-go-v2/service/configservice => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/configservice

replace github.com/aws/aws-sdk-go-v2/service/connect => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/connect

replace github.com/aws/aws-sdk-go-v2/service/connectcases => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/connectcases

replace github.com/aws/aws-sdk-go-v2/service/controltower => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/controltower

replace github.com/aws/aws-sdk-go-v2/service/costandusagereportservice => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/costandusagereportservice

replace github.com/aws/aws-sdk-go-v2/service/costexplorer => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/costexplorer

replace github.com/aws/aws-sdk-go-v2/service/costoptimizationhub => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/costoptimizationhub

replace github.com/aws/aws-sdk-go-v2/service/customerprofiles => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/customerprofiles

replace github.com/aws/aws-sdk-go-v2/service/databasemigrationservice => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/databasemigrationservice

replace github.com/aws/aws-sdk-go-v2/service/databrew => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/databrew

replace github.com/aws/aws-sdk-go-v2/service/dataexchange => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/dataexchange

replace github.com/aws/aws-sdk-go-v2/service/datapipeline => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/datapipeline

replace github.com/aws/aws-sdk-go-v2/service/datasync => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/datasync

replace github.com/aws/aws-sdk-go-v2/service/datazone => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/datazone

replace github.com/aws/aws-sdk-go-v2/service/dax => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/dax

replace github.com/aws/aws-sdk-go-v2/service/detective => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/detective

replace github.com/aws/aws-sdk-go-v2/service/devicefarm => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/devicefarm

replace github.com/aws/aws-sdk-go-v2/service/devopsguru => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/devopsguru

replace github.com/aws/aws-sdk-go-v2/service/directconnect => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/directconnect

replace github.com/aws/aws-sdk-go-v2/service/directoryservice => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/directoryservice

replace github.com/aws/aws-sdk-go-v2/service/dlm => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/dlm

replace github.com/aws/aws-sdk-go-v2/service/docdb => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/docdb

replace github.com/aws/aws-sdk-go-v2/service/docdbelastic => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/docdbelastic

replace github.com/aws/aws-sdk-go-v2/service/drs => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/drs

replace github.com/aws/aws-sdk-go-v2/service/dsql => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/dsql

replace github.com/aws/aws-sdk-go-v2/service/dynamodb => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/dynamodb

replace github.com/aws/aws-sdk-go-v2/service/ec2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ec2

replace github.com/aws/aws-sdk-go-v2/service/ecr => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ecr

replace github.com/aws/aws-sdk-go-v2/service/ecrpublic => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ecrpublic

replace github.com/aws/aws-sdk-go-v2/service/ecs => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ecs

replace github.com/aws/aws-sdk-go-v2/service/efs => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/efs

replace github.com/aws/aws-sdk-go-v2/service/eks => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/eks

replace github.com/aws/aws-sdk-go-v2/service/elasticache => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/elasticache

replace github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/elasticbeanstalk

replace github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/elasticloadbalancing

replace github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/elasticloadbalancingv2

replace github.com/aws/aws-sdk-go-v2/service/elasticsearchservice => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/elasticsearchservice

replace github.com/aws/aws-sdk-go-v2/service/elastictranscoder => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/elastictranscoder

replace github.com/aws/aws-sdk-go-v2/service/emr => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/emr

replace github.com/aws/aws-sdk-go-v2/service/emrcontainers => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/emrcontainers

replace github.com/aws/aws-sdk-go-v2/service/emrserverless => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/emrserverless

replace github.com/aws/aws-sdk-go-v2/service/eventbridge => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/eventbridge

replace github.com/aws/aws-sdk-go-v2/service/evidently => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/evidently

replace github.com/aws/aws-sdk-go-v2/service/evs => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/evs

replace github.com/aws/aws-sdk-go-v2/service/finspace => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/finspace

replace github.com/aws/aws-sdk-go-v2/service/firehose => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/firehose

replace github.com/aws/aws-sdk-go-v2/service/fis => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/fis

replace github.com/aws/aws-sdk-go-v2/service/fms => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/fms

replace github.com/aws/aws-sdk-go-v2/service/fsx => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/fsx

replace github.com/aws/aws-sdk-go-v2/service/gamelift => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/gamelift

replace github.com/aws/aws-sdk-go-v2/service/glacier => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/glacier

replace github.com/aws/aws-sdk-go-v2/service/globalaccelerator => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/globalaccelerator

replace github.com/aws/aws-sdk-go-v2/service/glue => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/glue

replace github.com/aws/aws-sdk-go-v2/service/grafana => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/grafana

replace github.com/aws/aws-sdk-go-v2/service/greengrass => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/greengrass

replace github.com/aws/aws-sdk-go-v2/service/groundstation => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/groundstation

replace github.com/aws/aws-sdk-go-v2/service/guardduty => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/guardduty

replace github.com/aws/aws-sdk-go-v2/service/healthlake => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/healthlake

replace github.com/aws/aws-sdk-go-v2/service/iam => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/iam

replace github.com/aws/aws-sdk-go-v2/service/identitystore => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/identitystore

replace github.com/aws/aws-sdk-go-v2/service/imagebuilder => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/imagebuilder

replace github.com/aws/aws-sdk-go-v2/service/inspector => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/inspector

replace github.com/aws/aws-sdk-go-v2/service/inspector2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/inspector2

replace github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/internal/accept-encoding

replace github.com/aws/aws-sdk-go-v2/service/internal/checksum => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/internal/checksum

replace github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/internal/endpoint-discovery

replace github.com/aws/aws-sdk-go-v2/service/internal/presigned-url => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/internal/presigned-url

replace github.com/aws/aws-sdk-go-v2/service/internal/s3shared => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/internal/s3shared

replace github.com/aws/aws-sdk-go-v2/service/internetmonitor => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/internetmonitor

replace github.com/aws/aws-sdk-go-v2/service/invoicing => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/invoicing

replace github.com/aws/aws-sdk-go-v2/service/iot => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/iot

replace github.com/aws/aws-sdk-go-v2/service/ivs => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ivs

replace github.com/aws/aws-sdk-go-v2/service/ivschat => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ivschat

replace github.com/aws/aws-sdk-go-v2/service/kafka => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/kafka

replace github.com/aws/aws-sdk-go-v2/service/kafkaconnect => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/kafkaconnect

replace github.com/aws/aws-sdk-go-v2/service/kendra => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/kendra

replace github.com/aws/aws-sdk-go-v2/service/keyspaces => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/keyspaces

replace github.com/aws/aws-sdk-go-v2/service/kinesis => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/kinesis

replace github.com/aws/aws-sdk-go-v2/service/kinesisanalytics => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/kinesisanalytics

replace github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/kinesisanalyticsv2

replace github.com/aws/aws-sdk-go-v2/service/kinesisvideo => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/kinesisvideo

replace github.com/aws/aws-sdk-go-v2/service/kms => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/kms

replace github.com/aws/aws-sdk-go-v2/service/lakeformation => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/lakeformation

replace github.com/aws/aws-sdk-go-v2/service/lambda => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/lambda

replace github.com/aws/aws-sdk-go-v2/service/launchwizard => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/launchwizard

replace github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/lexmodelbuildingservice

replace github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/lexmodelsv2

replace github.com/aws/aws-sdk-go-v2/service/licensemanager => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/licensemanager

replace github.com/aws/aws-sdk-go-v2/service/lightsail => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/lightsail

replace github.com/aws/aws-sdk-go-v2/service/location => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/location

replace github.com/aws/aws-sdk-go-v2/service/m2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/m2

replace github.com/aws/aws-sdk-go-v2/service/macie2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/macie2

replace github.com/aws/aws-sdk-go-v2/service/mediaconnect => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/mediaconnect

replace github.com/aws/aws-sdk-go-v2/service/mediaconvert => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/mediaconvert

replace github.com/aws/aws-sdk-go-v2/service/medialive => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/medialive

replace github.com/aws/aws-sdk-go-v2/service/mediapackage => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/mediapackage

replace github.com/aws/aws-sdk-go-v2/service/mediapackagev2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/mediapackagev2

replace github.com/aws/aws-sdk-go-v2/service/mediapackagevod => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/mediapackagevod

replace github.com/aws/aws-sdk-go-v2/service/mediastore => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/mediastore

replace github.com/aws/aws-sdk-go-v2/service/memorydb => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/memorydb

replace github.com/aws/aws-sdk-go-v2/service/mgn => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/mgn

replace github.com/aws/aws-sdk-go-v2/service/mq => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/mq

replace github.com/aws/aws-sdk-go-v2/service/mwaa => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/mwaa

replace github.com/aws/aws-sdk-go-v2/service/neptune => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/neptune

replace github.com/aws/aws-sdk-go-v2/service/neptunegraph => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/neptunegraph

replace github.com/aws/aws-sdk-go-v2/service/networkfirewall => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/networkfirewall

replace github.com/aws/aws-sdk-go-v2/service/networkflowmonitor => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/networkflowmonitor

replace github.com/aws/aws-sdk-go-v2/service/networkmanager => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/networkmanager

replace github.com/aws/aws-sdk-go-v2/service/networkmonitor => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/networkmonitor

replace github.com/aws/aws-sdk-go-v2/service/notifications => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/notifications

replace github.com/aws/aws-sdk-go-v2/service/notificationscontacts => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/notificationscontacts

replace github.com/aws/aws-sdk-go-v2/service/oam => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/oam

replace github.com/aws/aws-sdk-go-v2/service/observabilityadmin => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/observabilityadmin

replace github.com/aws/aws-sdk-go-v2/service/odb => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/odb

replace github.com/aws/aws-sdk-go-v2/service/opensearch => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/opensearch

replace github.com/aws/aws-sdk-go-v2/service/opensearchserverless => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/opensearchserverless

replace github.com/aws/aws-sdk-go-v2/service/organizations => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/organizations

replace github.com/aws/aws-sdk-go-v2/service/osis => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/osis

replace github.com/aws/aws-sdk-go-v2/service/outposts => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/outposts

replace github.com/aws/aws-sdk-go-v2/service/paymentcryptography => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/paymentcryptography

replace github.com/aws/aws-sdk-go-v2/service/pcaconnectorad => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/pcaconnectorad

replace github.com/aws/aws-sdk-go-v2/service/pcs => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/pcs

replace github.com/aws/aws-sdk-go-v2/service/pinpoint => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/pinpoint

replace github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/pinpointsmsvoicev2

replace github.com/aws/aws-sdk-go-v2/service/pipes => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/pipes

replace github.com/aws/aws-sdk-go-v2/service/polly => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/polly

replace github.com/aws/aws-sdk-go-v2/service/pricing => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/pricing

replace github.com/aws/aws-sdk-go-v2/service/qbusiness => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/qbusiness

replace github.com/aws/aws-sdk-go-v2/service/quicksight => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/quicksight

replace github.com/aws/aws-sdk-go-v2/service/ram => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ram

replace github.com/aws/aws-sdk-go-v2/service/rbin => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/rbin

replace github.com/aws/aws-sdk-go-v2/service/rds => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/rds

replace github.com/aws/aws-sdk-go-v2/service/redshift => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/redshift

replace github.com/aws/aws-sdk-go-v2/service/redshiftdata => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/redshiftdata

replace github.com/aws/aws-sdk-go-v2/service/redshiftserverless => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/redshiftserverless

replace github.com/aws/aws-sdk-go-v2/service/rekognition => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/rekognition

replace github.com/aws/aws-sdk-go-v2/service/resiliencehub => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/resiliencehub

replace github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/resourceexplorer2

replace github.com/aws/aws-sdk-go-v2/service/resourcegroups => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/resourcegroups

replace github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/resourcegroupstaggingapi

replace github.com/aws/aws-sdk-go-v2/service/rolesanywhere => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/rolesanywhere

replace github.com/aws/aws-sdk-go-v2/service/route53 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/route53

replace github.com/aws/aws-sdk-go-v2/service/route53domains => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/route53domains

replace github.com/aws/aws-sdk-go-v2/service/route53profiles => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/route53profiles

replace github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/route53recoverycontrolconfig

replace github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/route53recoveryreadiness

replace github.com/aws/aws-sdk-go-v2/service/route53resolver => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/route53resolver

replace github.com/aws/aws-sdk-go-v2/service/rum => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/rum

replace github.com/aws/aws-sdk-go-v2/service/s3 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/s3

replace github.com/aws/aws-sdk-go-v2/service/s3control => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/s3control

replace github.com/aws/aws-sdk-go-v2/service/s3outposts => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/s3outposts

replace github.com/aws/aws-sdk-go-v2/service/s3tables => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/s3tables

replace github.com/aws/aws-sdk-go-v2/service/s3vectors => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/s3vectors

replace github.com/aws/aws-sdk-go-v2/service/sagemaker => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/sagemaker

replace github.com/aws/aws-sdk-go-v2/service/scheduler => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/scheduler

replace github.com/aws/aws-sdk-go-v2/service/schemas => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/schemas

replace github.com/aws/aws-sdk-go-v2/service/secretsmanager => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/secretsmanager

replace github.com/aws/aws-sdk-go-v2/service/securityhub => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/securityhub

replace github.com/aws/aws-sdk-go-v2/service/securitylake => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/securitylake

replace github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/serverlessapplicationrepository

replace github.com/aws/aws-sdk-go-v2/service/servicecatalog => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/servicecatalog

replace github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/servicecatalogappregistry

replace github.com/aws/aws-sdk-go-v2/service/servicediscovery => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/servicediscovery

replace github.com/aws/aws-sdk-go-v2/service/servicequotas => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/servicequotas

replace github.com/aws/aws-sdk-go-v2/service/ses => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ses

replace github.com/aws/aws-sdk-go-v2/service/sesv2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/sesv2

replace github.com/aws/aws-sdk-go-v2/service/sfn => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/sfn

replace github.com/aws/aws-sdk-go-v2/service/shield => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/shield

replace github.com/aws/aws-sdk-go-v2/service/signer => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/signer

replace github.com/aws/aws-sdk-go-v2/service/sns => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/sns

replace github.com/aws/aws-sdk-go-v2/service/sqs => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/sqs

replace github.com/aws/aws-sdk-go-v2/service/ssm => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ssm

replace github.com/aws/aws-sdk-go-v2/service/ssmcontacts => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ssmcontacts

replace github.com/aws/aws-sdk-go-v2/service/ssmincidents => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ssmincidents

replace github.com/aws/aws-sdk-go-v2/service/ssmquicksetup => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ssmquicksetup

replace github.com/aws/aws-sdk-go-v2/service/ssmsap => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ssmsap

replace github.com/aws/aws-sdk-go-v2/service/sso => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/sso

replace github.com/aws/aws-sdk-go-v2/service/ssoadmin => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ssoadmin

replace github.com/aws/aws-sdk-go-v2/service/ssooidc => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/ssooidc

replace github.com/aws/aws-sdk-go-v2/service/storagegateway => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/storagegateway

replace github.com/aws/aws-sdk-go-v2/service/sts => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/sts

replace github.com/aws/aws-sdk-go-v2/service/swf => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/swf

replace github.com/aws/aws-sdk-go-v2/service/synthetics => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/synthetics

replace github.com/aws/aws-sdk-go-v2/service/taxsettings => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/taxsettings

replace github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/timestreaminfluxdb

replace github.com/aws/aws-sdk-go-v2/service/timestreamquery => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/timestreamquery

replace github.com/aws/aws-sdk-go-v2/service/timestreamwrite => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/timestreamwrite

replace github.com/aws/aws-sdk-go-v2/service/transcribe => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/transcribe

replace github.com/aws/aws-sdk-go-v2/service/transfer => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/transfer

replace github.com/aws/aws-sdk-go-v2/service/verifiedpermissions => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/verifiedpermissions

replace github.com/aws/aws-sdk-go-v2/service/vpclattice => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/vpclattice

replace github.com/aws/aws-sdk-go-v2/service/waf => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/waf

replace github.com/aws/aws-sdk-go-v2/service/wafregional => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/wafregional

replace github.com/aws/aws-sdk-go-v2/service/wafv2 => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/wafv2

replace github.com/aws/aws-sdk-go-v2/service/wellarchitected => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/wellarchitected

replace github.com/aws/aws-sdk-go-v2/service/workmail => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/workmail

replace github.com/aws/aws-sdk-go-v2/service/workspaces => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/workspaces

replace github.com/aws/aws-sdk-go-v2/service/workspacesweb => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/workspacesweb

replace github.com/aws/aws-sdk-go-v2/service/xray => /home/ec2-user/terraform-provider-aws/aws-sdk-go-v2/service/xray
