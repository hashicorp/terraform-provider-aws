module github.com/hashicorp/terraform-provider-aws

go 1.26.6

// Disable post-quantum X25519MLKEM768 key exchange mechanism
// This causes errors with AWS Network Firewall
godebug tlsmlkem=0

require (
	github.com/ProtonMail/go-crypto v1.4.1
	github.com/YakDriver/go-version v0.2.0
	github.com/YakDriver/regexache v0.25.0
	github.com/YakDriver/smarterr v0.8.0
	github.com/aws/aws-sdk-go-v2 v1.43.6
	github.com/aws/aws-sdk-go-v2/config v1.32.37
	github.com/aws/aws-sdk-go-v2/credentials v1.19.36
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.37
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.22.43
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.51.6
	github.com/aws/aws-sdk-go-v2/service/account v1.35.6
	github.com/aws/aws-sdk-go-v2/service/acm v1.44.1
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.50.2
	github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol v1.0.2
	github.com/aws/aws-sdk-go-v2/service/amp v1.48.3
	github.com/aws/aws-sdk-go-v2/service/amplify v1.41.6
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.42.6
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.37.6
	github.com/aws/aws-sdk-go-v2/service/appconfig v1.48.6
	github.com/aws/aws-sdk-go-v2/service/appfabric v1.19.6
	github.com/aws/aws-sdk-go-v2/service/appflow v1.54.6
	github.com/aws/aws-sdk-go-v2/service/appintegrations v1.40.6
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.45.6
	github.com/aws/aws-sdk-go-v2/service/applicationinsights v1.38.6
	github.com/aws/aws-sdk-go-v2/service/applicationsignals v1.25.6
	github.com/aws/aws-sdk-go-v2/service/appmesh v1.39.1
	github.com/aws/aws-sdk-go-v2/service/apprunner v1.42.6
	github.com/aws/aws-sdk-go-v2/service/appstream v1.64.7
	github.com/aws/aws-sdk-go-v2/service/appsync v1.56.6
	github.com/aws/aws-sdk-go-v2/service/arcregionswitch v1.13.5
	github.com/aws/aws-sdk-go-v2/service/arczonalshift v1.25.6
	github.com/aws/aws-sdk-go-v2/service/athena v1.60.6
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.49.6
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.72.1
	github.com/aws/aws-sdk-go-v2/service/autoscalingplans v1.33.6
	github.com/aws/aws-sdk-go-v2/service/backup v1.60.2
	github.com/aws/aws-sdk-go-v2/service/batch v1.68.6
	github.com/aws/aws-sdk-go-v2/service/bcmdataexports v1.19.6
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.66.6
	github.com/aws/aws-sdk-go-v2/service/bedrockagent v1.58.6
	github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol v1.57.0
	github.com/aws/aws-sdk-go-v2/service/billing v1.14.3
	github.com/aws/aws-sdk-go-v2/service/budgets v1.46.6
	github.com/aws/aws-sdk-go-v2/service/chatbot v1.17.6
	github.com/aws/aws-sdk-go-v2/service/chime v1.45.1
	github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines v1.30.1
	github.com/aws/aws-sdk-go-v2/service/chimesdkvoice v1.33.1
	github.com/aws/aws-sdk-go-v2/service/cleanrooms v1.51.1
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.37.1
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.32.6
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.76.3
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.67.6
	github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore v1.15.6
	github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 v1.37.6
	github.com/aws/aws-sdk-go-v2/service/cloudsearch v1.35.6
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.58.6
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.66.5
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.82.2
	github.com/aws/aws-sdk-go-v2/service/codeartifact v1.41.6
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.72.6
	github.com/aws/aws-sdk-go-v2/service/codecatalyst v1.24.6
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.38.1
	github.com/aws/aws-sdk-go-v2/service/codeconnections v1.13.6
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.38.6
	github.com/aws/aws-sdk-go-v2/service/codeguruprofiler v1.33.1
	github.com/aws/aws-sdk-go-v2/service/codegurureviewer v1.37.6
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.49.6
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.39.1
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.34.6
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.36.6
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.67.6
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.43.6
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.57.7
	github.com/aws/aws-sdk-go-v2/service/configservice v1.68.6
	github.com/aws/aws-sdk-go-v2/service/connect v1.189.0
	github.com/aws/aws-sdk-go-v2/service/connectcases v1.45.1
	github.com/aws/aws-sdk-go-v2/service/controltower v1.32.1
	github.com/aws/aws-sdk-go-v2/service/costandusagereportservice v1.37.6
	github.com/aws/aws-sdk-go-v2/service/costexplorer v1.67.6
	github.com/aws/aws-sdk-go-v2/service/costoptimizationhub v1.26.7
	github.com/aws/aws-sdk-go-v2/service/customerprofiles v1.65.6
	github.com/aws/aws-sdk-go-v2/service/databasemigrationservice v1.66.6
	github.com/aws/aws-sdk-go-v2/service/databrew v1.42.6
	github.com/aws/aws-sdk-go-v2/service/dataexchange v1.44.6
	github.com/aws/aws-sdk-go-v2/service/datapipeline v1.33.6
	github.com/aws/aws-sdk-go-v2/service/datasync v1.61.6
	github.com/aws/aws-sdk-go-v2/service/datazone v1.69.1
	github.com/aws/aws-sdk-go-v2/service/dax v1.33.1
	github.com/aws/aws-sdk-go-v2/service/detective v1.41.6
	github.com/aws/aws-sdk-go-v2/service/devicefarm v1.42.2
	github.com/aws/aws-sdk-go-v2/service/devopsagent v1.10.6
	github.com/aws/aws-sdk-go-v2/service/devopsguru v1.43.6
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.44.3
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.41.6
	github.com/aws/aws-sdk-go-v2/service/directoryservicedata v1.10.6
	github.com/aws/aws-sdk-go-v2/service/dlm v1.40.1
	github.com/aws/aws-sdk-go-v2/service/docdb v1.51.6
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.23.6
	github.com/aws/aws-sdk-go-v2/service/drs v1.43.0
	github.com/aws/aws-sdk-go-v2/service/dsql v1.16.7
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.63.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.321.2
	github.com/aws/aws-sdk-go-v2/service/ecr v1.60.6
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.41.6
	github.com/aws/aws-sdk-go-v2/service/ecs v1.90.2
	github.com/aws/aws-sdk-go-v2/service/efs v1.44.6
	github.com/aws/aws-sdk-go-v2/service/eks v1.91.1
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.56.6
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.37.6
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.36.6
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.58.7
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.45.6
	github.com/aws/aws-sdk-go-v2/service/elastictranscoder v1.33.0
	github.com/aws/aws-sdk-go-v2/service/emr v1.64.6
	github.com/aws/aws-sdk-go-v2/service/emrcontainers v1.45.6
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.44.6
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.48.6
	github.com/aws/aws-sdk-go-v2/service/evidently v1.30.0
	github.com/aws/aws-sdk-go-v2/service/evs v1.13.6
	github.com/aws/aws-sdk-go-v2/service/finspace v1.36.6
	github.com/aws/aws-sdk-go-v2/service/firehose v1.46.6
	github.com/aws/aws-sdk-go-v2/service/fis v1.40.6
	github.com/aws/aws-sdk-go-v2/service/fms v1.48.1
	github.com/aws/aws-sdk-go-v2/service/fsx v1.68.6
	github.com/aws/aws-sdk-go-v2/service/gamelift v1.61.2
	github.com/aws/aws-sdk-go-v2/service/glacier v1.35.6
	github.com/aws/aws-sdk-go-v2/service/globalaccelerator v1.38.6
	github.com/aws/aws-sdk-go-v2/service/glue v1.153.0
	github.com/aws/aws-sdk-go-v2/service/grafana v1.38.6
	github.com/aws/aws-sdk-go-v2/service/greengrass v1.35.6
	github.com/aws/aws-sdk-go-v2/service/groundstation v1.46.1
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.85.6
	github.com/aws/aws-sdk-go-v2/service/healthlake v1.43.2
	github.com/aws/aws-sdk-go-v2/service/iam v1.59.1
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.39.6
	github.com/aws/aws-sdk-go-v2/service/imagebuilder v1.58.6
	github.com/aws/aws-sdk-go-v2/service/inspector v1.33.6
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.54.3
	github.com/aws/aws-sdk-go-v2/service/interconnect v1.4.5
	github.com/aws/aws-sdk-go-v2/service/internetmonitor v1.29.6
	github.com/aws/aws-sdk-go-v2/service/invoicing v1.13.6
	github.com/aws/aws-sdk-go-v2/service/iot v1.77.6
	github.com/aws/aws-sdk-go-v2/service/ivs v1.55.6
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.24.6
	github.com/aws/aws-sdk-go-v2/service/kafka v1.58.2
	github.com/aws/aws-sdk-go-v2/service/kafkaconnect v1.33.6
	github.com/aws/aws-sdk-go-v2/service/kendra v1.64.1
	github.com/aws/aws-sdk-go-v2/service/keyspaces v1.28.7
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.46.6
	github.com/aws/aws-sdk-go-v2/service/kinesisanalytics v1.34.1
	github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2 v1.42.1
	github.com/aws/aws-sdk-go-v2/service/kinesisvideo v1.36.6
	github.com/aws/aws-sdk-go-v2/service/kms v1.55.6
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.50.6
	github.com/aws/aws-sdk-go-v2/service/lambda v1.101.4
	github.com/aws/aws-sdk-go-v2/service/lambdacore v1.2.6
	github.com/aws/aws-sdk-go-v2/service/lambdamicrovms v1.2.6
	github.com/aws/aws-sdk-go-v2/service/launchwizard v1.17.6
	github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice v1.38.6
	github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 v1.64.6
	github.com/aws/aws-sdk-go-v2/service/licensemanager v1.41.6
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.58.6
	github.com/aws/aws-sdk-go-v2/service/location v1.54.6
	github.com/aws/aws-sdk-go-v2/service/m2 v1.29.6
	github.com/aws/aws-sdk-go-v2/service/macie2 v1.54.6
	github.com/aws/aws-sdk-go-v2/service/mailmanager v1.21.6
	github.com/aws/aws-sdk-go-v2/service/mediaconnect v1.54.1
	github.com/aws/aws-sdk-go-v2/service/mediaconvert v1.97.3
	github.com/aws/aws-sdk-go-v2/service/medialive v1.102.1
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.42.6
	github.com/aws/aws-sdk-go-v2/service/mediapackagev2 v1.44.2
	github.com/aws/aws-sdk-go-v2/service/mediapackagevod v1.42.6
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.32.6
	github.com/aws/aws-sdk-go-v2/service/memorydb v1.37.1
	github.com/aws/aws-sdk-go-v2/service/mgn v1.49.1
	github.com/aws/aws-sdk-go-v2/service/mpa v1.10.6
	github.com/aws/aws-sdk-go-v2/service/mq v1.39.6
	github.com/aws/aws-sdk-go-v2/service/mwaa v1.43.6
	github.com/aws/aws-sdk-go-v2/service/mwaaserverless v1.4.0
	github.com/aws/aws-sdk-go-v2/service/neptune v1.48.6
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.24.6
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.67.3
	github.com/aws/aws-sdk-go-v2/service/networkflowmonitor v1.14.6
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.44.6
	github.com/aws/aws-sdk-go-v2/service/networkmonitor v1.16.6
	github.com/aws/aws-sdk-go-v2/service/notifications v1.11.1
	github.com/aws/aws-sdk-go-v2/service/notificationscontacts v1.9.1
	github.com/aws/aws-sdk-go-v2/service/oam v1.26.6
	github.com/aws/aws-sdk-go-v2/service/observabilityadmin v1.23.0
	github.com/aws/aws-sdk-go-v2/service/odb v1.16.1
	github.com/aws/aws-sdk-go-v2/service/opensearch v1.75.6
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.34.6
	github.com/aws/aws-sdk-go-v2/service/organizations v1.54.0
	github.com/aws/aws-sdk-go-v2/service/osis v1.24.6
	github.com/aws/aws-sdk-go-v2/service/outposts v1.66.3
	github.com/aws/aws-sdk-go-v2/service/paymentcryptography v1.33.6
	github.com/aws/aws-sdk-go-v2/service/pcaconnectorad v1.18.6
	github.com/aws/aws-sdk-go-v2/service/pcs v1.24.6
	github.com/aws/aws-sdk-go-v2/service/pinpoint v1.42.6
	github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2 v1.33.1
	github.com/aws/aws-sdk-go-v2/service/pipes v1.26.6
	github.com/aws/aws-sdk-go-v2/service/polly v1.60.6
	github.com/aws/aws-sdk-go-v2/service/pricing v1.44.6
	github.com/aws/aws-sdk-go-v2/service/qbusiness v1.37.6
	github.com/aws/aws-sdk-go-v2/service/qldb v1.32.2
	github.com/aws/aws-sdk-go-v2/service/quicksight v1.124.1
	github.com/aws/aws-sdk-go-v2/service/ram v1.39.6
	github.com/aws/aws-sdk-go-v2/service/rbin v1.30.6
	github.com/aws/aws-sdk-go-v2/service/rds v1.124.3
	github.com/aws/aws-sdk-go-v2/service/rdsdata v1.35.6
	github.com/aws/aws-sdk-go-v2/service/redshift v1.65.6
	github.com/aws/aws-sdk-go-v2/service/redshiftdata v1.43.6
	github.com/aws/aws-sdk-go-v2/service/redshiftserverless v1.38.7
	github.com/aws/aws-sdk-go-v2/service/rekognition v1.54.6
	github.com/aws/aws-sdk-go-v2/service/resiliencehub v1.38.6
	github.com/aws/aws-sdk-go-v2/service/resiliencehubv2 v1.4.3
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.27.7
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.37.1
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.36.1
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.26.5
	github.com/aws/aws-sdk-go-v2/service/route53 v1.65.8
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.39.1
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.12.6
	github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig v1.35.6
	github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness v1.29.6
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.48.6
	github.com/aws/aws-sdk-go-v2/service/rum v1.33.6
	github.com/aws/aws-sdk-go-v2/service/s3 v1.107.2
	github.com/aws/aws-sdk-go-v2/service/s3control v1.73.6
	github.com/aws/aws-sdk-go-v2/service/s3files v1.3.6
	github.com/aws/aws-sdk-go-v2/service/s3outposts v1.37.6
	github.com/aws/aws-sdk-go-v2/service/s3tables v1.18.6
	github.com/aws/aws-sdk-go-v2/service/s3vectors v1.10.6
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.267.0
	github.com/aws/aws-sdk-go-v2/service/savingsplans v1.35.6
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.20.6
	github.com/aws/aws-sdk-go-v2/service/schemas v1.37.6
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.44.6
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.76.2
	github.com/aws/aws-sdk-go-v2/service/securitylake v1.29.1
	github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository v1.33.6
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.42.6
	github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry v1.38.6
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.43.6
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.37.6
	github.com/aws/aws-sdk-go-v2/service/ses v1.37.6
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.66.6
	github.com/aws/aws-sdk-go-v2/service/sfn v1.45.6
	github.com/aws/aws-sdk-go-v2/service/shield v1.37.6
	github.com/aws/aws-sdk-go-v2/service/signer v1.35.6
	github.com/aws/aws-sdk-go-v2/service/sns v1.42.6
	github.com/aws/aws-sdk-go-v2/service/sqs v1.46.6
	github.com/aws/aws-sdk-go-v2/service/ssm v1.73.5
	github.com/aws/aws-sdk-go-v2/service/ssmcontacts v1.34.5
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.42.5
	github.com/aws/aws-sdk-go-v2/service/ssmquicksetup v1.12.0
	github.com/aws/aws-sdk-go-v2/service/ssmsap v1.30.0
	github.com/aws/aws-sdk-go-v2/service/sso v1.33.6
	github.com/aws/aws-sdk-go-v2/service/ssoadmin v1.43.2
	github.com/aws/aws-sdk-go-v2/service/storagegateway v1.46.5
	github.com/aws/aws-sdk-go-v2/service/sts v1.45.6
	github.com/aws/aws-sdk-go-v2/service/swf v1.37.5
	github.com/aws/aws-sdk-go-v2/service/synthetics v1.47.5
	github.com/aws/aws-sdk-go-v2/service/taxsettings v1.21.0
	github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb v1.23.2
	github.com/aws/aws-sdk-go-v2/service/timestreamquery v1.39.5
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.38.5
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.58.5
	github.com/aws/aws-sdk-go-v2/service/transfer v1.75.5
	github.com/aws/aws-sdk-go-v2/service/uxc v1.3.5
	github.com/aws/aws-sdk-go-v2/service/verifiedpermissions v1.36.5
	github.com/aws/aws-sdk-go-v2/service/vpclattice v1.25.6
	github.com/aws/aws-sdk-go-v2/service/waf v1.33.5
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.33.5
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.77.4
	github.com/aws/aws-sdk-go-v2/service/wellarchitected v1.43.0
	github.com/aws/aws-sdk-go-v2/service/workmail v1.40.0
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.73.2
	github.com/aws/aws-sdk-go-v2/service/workspacesweb v1.42.5
	github.com/aws/aws-sdk-go-v2/service/xray v1.39.5
	github.com/aws/smithy-go v1.27.8
	github.com/beevik/etree v1.7.0
	github.com/cedar-policy/cedar-go v1.8.0
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dlclark/regexp2 v1.12.0
	github.com/gertd/go-pluralize v0.2.1
	github.com/goccy/go-yaml v1.19.2
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.24.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.73
	github.com/hashicorp/awspolicyequivalence v1.7.0
	github.com/hashicorp/cli v1.1.7
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.5.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-set/v3 v3.0.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.9.0
	github.com/hashicorp/hcl/v2 v2.24.0
	github.com/hashicorp/terraform-json v0.28.0
	github.com/hashicorp/terraform-plugin-framework v1.19.0
	github.com/hashicorp/terraform-plugin-framework-jsontypes v0.2.0
	github.com/hashicorp/terraform-plugin-framework-timeouts v0.7.0
	github.com/hashicorp/terraform-plugin-framework-timetypes v0.5.0
	github.com/hashicorp/terraform-plugin-framework-validators v0.19.0
	github.com/hashicorp/terraform-plugin-go v0.31.0
	github.com/hashicorp/terraform-plugin-log v0.10.0
	github.com/hashicorp/terraform-plugin-mux v0.23.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.40.1
	github.com/hashicorp/terraform-plugin-testing v1.16.0
	github.com/jaswdr/faker/v2 v2.9.1
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20240118010651-0ba75a80ca38
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pquerna/otp v1.5.0
	github.com/shopspring/decimal v1.4.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.70.0
	go.opentelemetry.io/otel v1.45.0
	golang.org/x/crypto v0.55.0
	golang.org/x/text v0.41.0
	golang.org/x/tools v0.49.0
	gopkg.in/dnaeon/go-vcr.v4 v4.0.7
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.38 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.12.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.38 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.5.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.38.6 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-plugin v1.7.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/hc-install v0.9.4 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.25.1 // indirect
	github.com/hashicorp/terraform-registry-address v0.4.0 // indirect
	github.com/hashicorp/terraform-svchost v0.2.1 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.2.0 // indirect
	github.com/posener/complete v1.2.3 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.18.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.6 // indirect
	golang.org/x/exp v0.0.0-20220921023135-46d9e7742f1e // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/grpc v1.82.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
