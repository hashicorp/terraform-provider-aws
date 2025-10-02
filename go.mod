module github.com/hashicorp/terraform-provider-aws

go 1.24.6

// Disable post-quantum X25519MLKEM768 key exchange mechanism
// This causes errors with AWS Network Firewall
godebug tlsmlkem=0

require (
	github.com/ProtonMail/go-crypto v1.3.0
	github.com/YakDriver/go-version v0.1.0
	github.com/YakDriver/regexache v0.24.0
	github.com/YakDriver/smarterr v0.6.0
	github.com/aws/aws-sdk-go-v2 v1.39.2
	github.com/aws/aws-sdk-go-v2/config v1.31.12
	github.com/aws/aws-sdk-go-v2/credentials v1.18.16
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.9
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.19.10
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.44.6
	github.com/aws/aws-sdk-go-v2/service/account v1.28.6
	github.com/aws/aws-sdk-go-v2/service/acm v1.37.6
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.44.5
	github.com/aws/aws-sdk-go-v2/service/amp v1.40.3
	github.com/aws/aws-sdk-go-v2/service/amplify v1.37.5
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.35.6
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.32.6
	github.com/aws/aws-sdk-go-v2/service/appconfig v1.42.6
	github.com/aws/aws-sdk-go-v2/service/appfabric v1.16.6
	github.com/aws/aws-sdk-go-v2/service/appflow v1.50.6
	github.com/aws/aws-sdk-go-v2/service/appintegrations v1.36.6
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.40.5
	github.com/aws/aws-sdk-go-v2/service/applicationinsights v1.34.5
	github.com/aws/aws-sdk-go-v2/service/applicationsignals v1.16.0
	github.com/aws/aws-sdk-go-v2/service/appmesh v1.34.6
	github.com/aws/aws-sdk-go-v2/service/apprunner v1.38.7
	github.com/aws/aws-sdk-go-v2/service/appstream v1.49.7
	github.com/aws/aws-sdk-go-v2/service/appsync v1.51.6
	github.com/aws/aws-sdk-go-v2/service/arcregionswitch v1.2.8
	github.com/aws/aws-sdk-go-v2/service/athena v1.55.6
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.45.6
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.59.3
	github.com/aws/aws-sdk-go-v2/service/autoscalingplans v1.29.5
	github.com/aws/aws-sdk-go-v2/service/backup v1.47.7
	github.com/aws/aws-sdk-go-v2/service/batch v1.57.10
	github.com/aws/aws-sdk-go-v2/service/bcmdataexports v1.11.8
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.47.2
	github.com/aws/aws-sdk-go-v2/service/bedrockagent v1.50.6
	github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol v1.6.0
	github.com/aws/aws-sdk-go-v2/service/billing v1.8.0
	github.com/aws/aws-sdk-go-v2/service/budgets v1.39.2
	github.com/aws/aws-sdk-go-v2/service/chatbot v1.14.6
	github.com/aws/aws-sdk-go-v2/service/chime v1.40.5
	github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines v1.26.6
	github.com/aws/aws-sdk-go-v2/service/chimesdkvoice v1.27.0
	github.com/aws/aws-sdk-go-v2/service/cleanrooms v1.35.0
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.33.5
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.28.6
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.66.4
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.54.4
	github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore v1.12.8
	github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 v1.34.5
	github.com/aws/aws-sdk-go-v2/service/cloudsearch v1.31.6
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.53.6
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.51.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.58.2
	github.com/aws/aws-sdk-go-v2/service/codeartifact v1.38.6
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.67.5
	github.com/aws/aws-sdk-go-v2/service/codecatalyst v1.20.8
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.32.6
	github.com/aws/aws-sdk-go-v2/service/codeconnections v1.10.5
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.34.5
	github.com/aws/aws-sdk-go-v2/service/codeguruprofiler v1.29.5
	github.com/aws/aws-sdk-go-v2/service/codegurureviewer v1.34.5
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.46.6
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.34.6
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.31.6
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.33.6
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.57.7
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.40.6
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.47.5
	github.com/aws/aws-sdk-go-v2/service/configservice v1.58.2
	github.com/aws/aws-sdk-go-v2/service/connect v1.141.0
	github.com/aws/aws-sdk-go-v2/service/connectcases v1.31.0
	github.com/aws/aws-sdk-go-v2/service/controltower v1.26.6
	github.com/aws/aws-sdk-go-v2/service/costandusagereportservice v1.33.6
	github.com/aws/aws-sdk-go-v2/service/costexplorer v1.57.0
	github.com/aws/aws-sdk-go-v2/service/costoptimizationhub v1.20.6
	github.com/aws/aws-sdk-go-v2/service/customerprofiles v1.53.0
	github.com/aws/aws-sdk-go-v2/service/databasemigrationservice v1.57.7
	github.com/aws/aws-sdk-go-v2/service/databrew v1.38.5
	github.com/aws/aws-sdk-go-v2/service/dataexchange v1.39.6
	github.com/aws/aws-sdk-go-v2/service/datapipeline v1.30.5
	github.com/aws/aws-sdk-go-v2/service/datasync v1.55.0
	github.com/aws/aws-sdk-go-v2/service/datazone v1.42.0
	github.com/aws/aws-sdk-go-v2/service/dax v1.29.1
	github.com/aws/aws-sdk-go-v2/service/detective v1.37.7
	github.com/aws/aws-sdk-go-v2/service/devicefarm v1.35.6
	github.com/aws/aws-sdk-go-v2/service/devopsguru v1.39.6
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.37.6
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.38.0
	github.com/aws/aws-sdk-go-v2/service/dlm v1.34.6
	github.com/aws/aws-sdk-go-v2/service/docdb v1.46.6
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.19.6
	github.com/aws/aws-sdk-go-v2/service/drs v1.35.6
	github.com/aws/aws-sdk-go-v2/service/dsql v1.9.8
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.50.5
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.254.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.50.5
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.37.6
	github.com/aws/aws-sdk-go-v2/service/ecs v1.65.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.40.8
	github.com/aws/aws-sdk-go-v2/service/eks v1.74.2
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.50.5
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.33.7
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.33.6
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.50.6
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.37.6
	github.com/aws/aws-sdk-go-v2/service/elastictranscoder v1.32.6
	github.com/aws/aws-sdk-go-v2/service/emr v1.54.5
	github.com/aws/aws-sdk-go-v2/service/emrcontainers v1.40.2
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.36.6
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.45.5
	github.com/aws/aws-sdk-go-v2/service/evidently v1.28.5
	github.com/aws/aws-sdk-go-v2/service/evs v1.5.2
	github.com/aws/aws-sdk-go-v2/service/finspace v1.33.6
	github.com/aws/aws-sdk-go-v2/service/firehose v1.41.6
	github.com/aws/aws-sdk-go-v2/service/fis v1.37.5
	github.com/aws/aws-sdk-go-v2/service/fms v1.44.6
	github.com/aws/aws-sdk-go-v2/service/fsx v1.62.0
	github.com/aws/aws-sdk-go-v2/service/gamelift v1.46.6
	github.com/aws/aws-sdk-go-v2/service/glacier v1.31.6
	github.com/aws/aws-sdk-go-v2/service/globalaccelerator v1.34.6
	github.com/aws/aws-sdk-go-v2/service/glue v1.129.1
	github.com/aws/aws-sdk-go-v2/service/grafana v1.31.6
	github.com/aws/aws-sdk-go-v2/service/greengrass v1.32.6
	github.com/aws/aws-sdk-go-v2/service/groundstation v1.37.6
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.64.2
	github.com/aws/aws-sdk-go-v2/service/healthlake v1.35.5
	github.com/aws/aws-sdk-go-v2/service/iam v1.47.7
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.32.6
	github.com/aws/aws-sdk-go-v2/service/imagebuilder v1.48.0
	github.com/aws/aws-sdk-go-v2/service/inspector v1.30.5
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.44.6
	github.com/aws/aws-sdk-go-v2/service/internetmonitor v1.25.5
	github.com/aws/aws-sdk-go-v2/service/invoicing v1.6.8
	github.com/aws/aws-sdk-go-v2/service/iot v1.69.5
	github.com/aws/aws-sdk-go-v2/service/ivs v1.47.6
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.21.5
	github.com/aws/aws-sdk-go-v2/service/kafka v1.43.6
	github.com/aws/aws-sdk-go-v2/service/kafkaconnect v1.27.5
	github.com/aws/aws-sdk-go-v2/service/kendra v1.60.6
	github.com/aws/aws-sdk-go-v2/service/keyspaces v1.23.6
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.40.5
	github.com/aws/aws-sdk-go-v2/service/kinesisanalytics v1.30.6
	github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2 v1.36.7
	github.com/aws/aws-sdk-go-v2/service/kinesisvideo v1.32.5
	github.com/aws/aws-sdk-go-v2/service/kms v1.45.6
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.45.5
	github.com/aws/aws-sdk-go-v2/service/lambda v1.77.6
	github.com/aws/aws-sdk-go-v2/service/launchwizard v1.13.6
	github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice v1.33.5
	github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 v1.56.6
	github.com/aws/aws-sdk-go-v2/service/licensemanager v1.36.6
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.49.1
	github.com/aws/aws-sdk-go-v2/service/location v1.49.6
	github.com/aws/aws-sdk-go-v2/service/lookoutmetrics v1.36.6
	github.com/aws/aws-sdk-go-v2/service/m2 v1.25.6
	github.com/aws/aws-sdk-go-v2/service/macie2 v1.49.6
	github.com/aws/aws-sdk-go-v2/service/mediaconnect v1.44.6
	github.com/aws/aws-sdk-go-v2/service/mediaconvert v1.82.6
	github.com/aws/aws-sdk-go-v2/service/medialive v1.83.2
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.39.6
	github.com/aws/aws-sdk-go-v2/service/mediapackagev2 v1.31.3
	github.com/aws/aws-sdk-go-v2/service/mediapackagevod v1.39.6
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.29.6
	github.com/aws/aws-sdk-go-v2/service/memorydb v1.31.6
	github.com/aws/aws-sdk-go-v2/service/mgn v1.37.5
	github.com/aws/aws-sdk-go-v2/service/mq v1.34.4
	github.com/aws/aws-sdk-go-v2/service/mwaa v1.39.6
	github.com/aws/aws-sdk-go-v2/service/neptune v1.42.5
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.21.5
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.57.1
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.39.7
	github.com/aws/aws-sdk-go-v2/service/networkmonitor v1.12.6
	github.com/aws/aws-sdk-go-v2/service/notifications v1.7.4
	github.com/aws/aws-sdk-go-v2/service/notificationscontacts v1.5.8
	github.com/aws/aws-sdk-go-v2/service/oam v1.22.5
	github.com/aws/aws-sdk-go-v2/service/odb v1.4.6
	github.com/aws/aws-sdk-go-v2/service/opensearch v1.52.5
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.26.4
	github.com/aws/aws-sdk-go-v2/service/organizations v1.45.3
	github.com/aws/aws-sdk-go-v2/service/osis v1.20.2
	github.com/aws/aws-sdk-go-v2/service/outposts v1.56.7
	github.com/aws/aws-sdk-go-v2/service/paymentcryptography v1.25.2
	github.com/aws/aws-sdk-go-v2/service/pcaconnectorad v1.15.6
	github.com/aws/aws-sdk-go-v2/service/pcs v1.14.0
	github.com/aws/aws-sdk-go-v2/service/pinpoint v1.39.6
	github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2 v1.25.5
	github.com/aws/aws-sdk-go-v2/service/pipes v1.23.5
	github.com/aws/aws-sdk-go-v2/service/polly v1.53.6
	github.com/aws/aws-sdk-go-v2/service/pricing v1.39.6
	github.com/aws/aws-sdk-go-v2/service/qbusiness v1.33.6
	github.com/aws/aws-sdk-go-v2/service/qldb v1.30.6
	github.com/aws/aws-sdk-go-v2/service/quicksight v1.94.3
	github.com/aws/aws-sdk-go-v2/service/ram v1.34.6
	github.com/aws/aws-sdk-go-v2/service/rbin v1.26.6
	github.com/aws/aws-sdk-go-v2/service/rds v1.108.0
	github.com/aws/aws-sdk-go-v2/service/redshift v1.59.0
	github.com/aws/aws-sdk-go-v2/service/redshiftdata v1.37.6
	github.com/aws/aws-sdk-go-v2/service/redshiftserverless v1.31.8
	github.com/aws/aws-sdk-go-v2/service/rekognition v1.51.5
	github.com/aws/aws-sdk-go-v2/service/resiliencehub v1.34.6
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.21.6
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.33.7
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.30.6
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.21.6
	github.com/aws/aws-sdk-go-v2/service/route53 v1.58.4
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.34.4
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.9.6
	github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig v1.31.7
	github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness v1.26.6
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.40.6
	github.com/aws/aws-sdk-go-v2/service/rum v1.28.7
	github.com/aws/aws-sdk-go-v2/service/s3 v1.88.3
	github.com/aws/aws-sdk-go-v2/service/s3control v1.66.2
	github.com/aws/aws-sdk-go-v2/service/s3outposts v1.33.6
	github.com/aws/aws-sdk-go-v2/service/s3tables v1.10.5
	github.com/aws/aws-sdk-go-v2/service/s3vectors v1.4.8
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.215.3
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.17.5
	github.com/aws/aws-sdk-go-v2/service/schemas v1.33.5
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.39.6
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.64.4
	github.com/aws/aws-sdk-go-v2/service/securitylake v1.24.6
	github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository v1.29.6
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.38.6
	github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry v1.35.6
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.39.9
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.32.5
	github.com/aws/aws-sdk-go-v2/service/ses v1.34.5
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.53.5
	github.com/aws/aws-sdk-go-v2/service/sfn v1.39.6
	github.com/aws/aws-sdk-go-v2/service/shield v1.34.6
	github.com/aws/aws-sdk-go-v2/service/signer v1.31.6
	github.com/aws/aws-sdk-go-v2/service/sns v1.38.5
	github.com/aws/aws-sdk-go-v2/service/sqs v1.42.8
	github.com/aws/aws-sdk-go-v2/service/ssm v1.65.1
	github.com/aws/aws-sdk-go-v2/service/ssmcontacts v1.30.8
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.39.5
	github.com/aws/aws-sdk-go-v2/service/ssmquicksetup v1.8.6
	github.com/aws/aws-sdk-go-v2/service/ssmsap v1.25.5
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.6
	github.com/aws/aws-sdk-go-v2/service/ssoadmin v1.36.2
	github.com/aws/aws-sdk-go-v2/service/storagegateway v1.42.6
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.6
	github.com/aws/aws-sdk-go-v2/service/swf v1.32.5
	github.com/aws/aws-sdk-go-v2/service/synthetics v1.40.6
	github.com/aws/aws-sdk-go-v2/service/taxsettings v1.16.6
	github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb v1.16.6
	github.com/aws/aws-sdk-go-v2/service/timestreamquery v1.35.5
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.35.5
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.52.6
	github.com/aws/aws-sdk-go-v2/service/transfer v1.66.0
	github.com/aws/aws-sdk-go-v2/service/verifiedpermissions v1.29.5
	github.com/aws/aws-sdk-go-v2/service/vpclattice v1.19.0
	github.com/aws/aws-sdk-go-v2/service/waf v1.30.5
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.30.6
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.67.7
	github.com/aws/aws-sdk-go-v2/service/wellarchitected v1.39.6
	github.com/aws/aws-sdk-go-v2/service/workmail v1.36.4
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.63.6
	github.com/aws/aws-sdk-go-v2/service/workspacesweb v1.32.6
	github.com/aws/aws-sdk-go-v2/service/xray v1.36.4
	github.com/aws/smithy-go v1.23.0
	github.com/beevik/etree v1.6.0
	github.com/cedar-policy/cedar-go v1.2.6
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dlclark/regexp2 v1.11.5
	github.com/gertd/go-pluralize v0.2.1
	github.com/goccy/go-yaml v1.18.0
	github.com/google/go-cmp v0.7.0
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.23.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.66
	github.com/hashicorp/awspolicyequivalence v1.7.0
	github.com/hashicorp/cli v1.1.7
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.5.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.7.0
	github.com/hashicorp/hcl/v2 v2.24.0
	github.com/hashicorp/terraform-json v0.27.2
	github.com/hashicorp/terraform-plugin-framework v1.16.0
	github.com/hashicorp/terraform-plugin-framework-jsontypes v0.2.0
	github.com/hashicorp/terraform-plugin-framework-timeouts v0.6.0
	github.com/hashicorp/terraform-plugin-framework-timetypes v0.5.0
	github.com/hashicorp/terraform-plugin-framework-validators v0.18.0
	github.com/hashicorp/terraform-plugin-go v0.29.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-mux v0.21.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.38.1
	github.com/hashicorp/terraform-plugin-testing v1.14.0-beta.1
	github.com/jaswdr/faker/v2 v2.8.1
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20240118010651-0ba75a80ca38
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pquerna/otp v1.5.0
	github.com/shopspring/decimal v1.4.0
	golang.org/x/crypto v0.42.0
	golang.org/x/text v0.29.0
	golang.org/x/tools v0.37.0
	gopkg.in/dnaeon/go-vcr.v4 v4.0.5
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.1 // indirect
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
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.63.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	golang.org/x/exp v0.0.0-20220921023135-46d9e7742f1e // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)

replace github.com/hashicorp/terraform-plugin-log => github.com/gdavison/terraform-plugin-log v0.0.0-20230928191232-6c653d8ef8fb
