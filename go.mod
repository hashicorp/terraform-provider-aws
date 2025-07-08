module github.com/hashicorp/terraform-provider-aws

go 1.24.4

// Disable post-quantum X25519MLKEM768 key exchange mechanism
// This causes errors with AWS Network Firewall
godebug tlsmlkem=0

require (
	github.com/ProtonMail/go-crypto v1.3.0
	github.com/YakDriver/go-version v0.1.0
	github.com/YakDriver/regexache v0.24.0
	github.com/aws/aws-sdk-go-v2 v1.36.5
	github.com/aws/aws-sdk-go-v2/config v1.29.17
	github.com/aws/aws-sdk-go-v2/credentials v1.17.70
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.32
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.83
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.40.0
	github.com/aws/aws-sdk-go-v2/service/account v1.24.2
	github.com/aws/aws-sdk-go-v2/service/acm v1.33.0
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.40.5
	github.com/aws/aws-sdk-go-v2/service/amp v1.34.3
	github.com/aws/aws-sdk-go-v2/service/amplify v1.33.3
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.31.4
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.28.4
	github.com/aws/aws-sdk-go-v2/service/appconfig v1.38.3
	github.com/aws/aws-sdk-go-v2/service/appfabric v1.12.4
	github.com/aws/aws-sdk-go-v2/service/appflow v1.46.4
	github.com/aws/aws-sdk-go-v2/service/appintegrations v1.31.4
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.36.4
	github.com/aws/aws-sdk-go-v2/service/applicationinsights v1.30.6
	github.com/aws/aws-sdk-go-v2/service/applicationsignals v1.11.3
	github.com/aws/aws-sdk-go-v2/service/appmesh v1.30.4
	github.com/aws/aws-sdk-go-v2/service/apprunner v1.34.2
	github.com/aws/aws-sdk-go-v2/service/appstream v1.45.5
	github.com/aws/aws-sdk-go-v2/service/appsync v1.47.3
	github.com/aws/aws-sdk-go-v2/service/athena v1.51.3
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.39.2
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.54.0
	github.com/aws/aws-sdk-go-v2/service/autoscalingplans v1.25.4
	github.com/aws/aws-sdk-go-v2/service/backup v1.43.1
	github.com/aws/aws-sdk-go-v2/service/batch v1.53.0
	github.com/aws/aws-sdk-go-v2/service/bcmdataexports v1.8.4
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.38.0
	github.com/aws/aws-sdk-go-v2/service/bedrockagent v1.44.2
	github.com/aws/aws-sdk-go-v2/service/billing v1.2.4
	github.com/aws/aws-sdk-go-v2/service/budgets v1.31.2
	github.com/aws/aws-sdk-go-v2/service/chatbot v1.10.4
	github.com/aws/aws-sdk-go-v2/service/chime v1.36.4
	github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines v1.22.4
	github.com/aws/aws-sdk-go-v2/service/chimesdkvoice v1.22.2
	github.com/aws/aws-sdk-go-v2/service/cleanrooms v1.25.2
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.29.4
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.24.6
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.61.0
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.46.3
	github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore v1.9.4
	github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 v1.30.5
	github.com/aws/aws-sdk-go-v2/service/cloudsearch v1.27.4
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.49.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.45.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.51.0
	github.com/aws/aws-sdk-go-v2/service/codeartifact v1.34.4
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.61.2
	github.com/aws/aws-sdk-go-v2/service/codecatalyst v1.17.21
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.28.4
	github.com/aws/aws-sdk-go-v2/service/codeconnections v1.6.4
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.30.6
	github.com/aws/aws-sdk-go-v2/service/codeguruprofiler v1.25.4
	github.com/aws/aws-sdk-go-v2/service/codegurureviewer v1.30.4
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.42.2
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.30.4
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.27.4
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.29.6
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.53.2
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.36.6
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.43.2
	github.com/aws/aws-sdk-go-v2/service/configservice v1.53.0
	github.com/aws/aws-sdk-go-v2/service/connect v1.131.0
	github.com/aws/aws-sdk-go-v2/service/connectcases v1.26.0
	github.com/aws/aws-sdk-go-v2/service/controltower v1.22.3
	github.com/aws/aws-sdk-go-v2/service/costandusagereportservice v1.29.4
	github.com/aws/aws-sdk-go-v2/service/costexplorer v1.51.2
	github.com/aws/aws-sdk-go-v2/service/costoptimizationhub v1.16.2
	github.com/aws/aws-sdk-go-v2/service/customerprofiles v1.47.0
	github.com/aws/aws-sdk-go-v2/service/databasemigrationservice v1.53.0
	github.com/aws/aws-sdk-go-v2/service/databrew v1.34.4
	github.com/aws/aws-sdk-go-v2/service/dataexchange v1.35.2
	github.com/aws/aws-sdk-go-v2/service/datapipeline v1.26.4
	github.com/aws/aws-sdk-go-v2/service/datasync v1.49.3
	github.com/aws/aws-sdk-go-v2/service/datazone v1.31.0
	github.com/aws/aws-sdk-go-v2/service/dax v1.24.4
	github.com/aws/aws-sdk-go-v2/service/detective v1.33.2
	github.com/aws/aws-sdk-go-v2/service/devicefarm v1.31.2
	github.com/aws/aws-sdk-go-v2/service/devopsguru v1.35.4
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.32.5
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.31.7
	github.com/aws/aws-sdk-go-v2/service/dlm v1.30.7
	github.com/aws/aws-sdk-go-v2/service/docdb v1.41.6
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.15.4
	github.com/aws/aws-sdk-go-v2/service/drs v1.31.4
	github.com/aws/aws-sdk-go-v2/service/dsql v1.5.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.44.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.230.0
	github.com/aws/aws-sdk-go-v2/service/ecr v1.45.1
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.33.2
	github.com/aws/aws-sdk-go-v2/service/ecs v1.58.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.36.2
	github.com/aws/aws-sdk-go-v2/service/eks v1.66.1
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.46.3
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.29.4
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.29.6
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.46.0
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.33.6
	github.com/aws/aws-sdk-go-v2/service/elastictranscoder v1.28.4
	github.com/aws/aws-sdk-go-v2/service/emr v1.49.3
	github.com/aws/aws-sdk-go-v2/service/emrcontainers v1.35.4
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.32.0
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.40.0
	github.com/aws/aws-sdk-go-v2/service/evidently v1.24.4
	github.com/aws/aws-sdk-go-v2/service/evs v1.0.2
	github.com/aws/aws-sdk-go-v2/service/finspace v1.29.4
	github.com/aws/aws-sdk-go-v2/service/firehose v1.37.7
	github.com/aws/aws-sdk-go-v2/service/fis v1.33.4
	github.com/aws/aws-sdk-go-v2/service/fms v1.40.5
	github.com/aws/aws-sdk-go-v2/service/fsx v1.55.0
	github.com/aws/aws-sdk-go-v2/service/gamelift v1.42.1
	github.com/aws/aws-sdk-go-v2/service/glacier v1.27.5
	github.com/aws/aws-sdk-go-v2/service/globalaccelerator v1.30.4
	github.com/aws/aws-sdk-go-v2/service/glue v1.117.0
	github.com/aws/aws-sdk-go-v2/service/grafana v1.27.4
	github.com/aws/aws-sdk-go-v2/service/greengrass v1.28.4
	github.com/aws/aws-sdk-go-v2/service/groundstation v1.33.2
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.56.0
	github.com/aws/aws-sdk-go-v2/service/healthlake v1.30.5
	github.com/aws/aws-sdk-go-v2/service/iam v1.43.0
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.28.6
	github.com/aws/aws-sdk-go-v2/service/imagebuilder v1.42.3
	github.com/aws/aws-sdk-go-v2/service/inspector v1.26.4
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.38.1
	github.com/aws/aws-sdk-go-v2/service/internetmonitor v1.21.5
	github.com/aws/aws-sdk-go-v2/service/invoicing v1.2.2
	github.com/aws/aws-sdk-go-v2/service/iot v1.64.4
	github.com/aws/aws-sdk-go-v2/service/ivs v1.43.4
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.17.4
	github.com/aws/aws-sdk-go-v2/service/kafka v1.39.5
	github.com/aws/aws-sdk-go-v2/service/kafkaconnect v1.23.5
	github.com/aws/aws-sdk-go-v2/service/kendra v1.56.4
	github.com/aws/aws-sdk-go-v2/service/keyspaces v1.19.0
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.35.3
	github.com/aws/aws-sdk-go-v2/service/kinesisanalytics v1.26.6
	github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2 v1.32.6
	github.com/aws/aws-sdk-go-v2/service/kinesisvideo v1.28.4
	github.com/aws/aws-sdk-go-v2/service/kms v1.41.2
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.41.8
	github.com/aws/aws-sdk-go-v2/service/lambda v1.72.0
	github.com/aws/aws-sdk-go-v2/service/launchwizard v1.9.4
	github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice v1.29.4
	github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 v1.52.1
	github.com/aws/aws-sdk-go-v2/service/licensemanager v1.32.0
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.43.4
	github.com/aws/aws-sdk-go-v2/service/location v1.44.4
	github.com/aws/aws-sdk-go-v2/service/lookoutmetrics v1.32.4
	github.com/aws/aws-sdk-go-v2/service/m2 v1.21.2
	github.com/aws/aws-sdk-go-v2/service/macie2 v1.45.4
	github.com/aws/aws-sdk-go-v2/service/mediaconnect v1.40.2
	github.com/aws/aws-sdk-go-v2/service/mediaconvert v1.75.0
	github.com/aws/aws-sdk-go-v2/service/medialive v1.76.2
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.35.4
	github.com/aws/aws-sdk-go-v2/service/mediapackagev2 v1.24.0
	github.com/aws/aws-sdk-go-v2/service/mediapackagevod v1.35.4
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.25.4
	github.com/aws/aws-sdk-go-v2/service/memorydb v1.27.2
	github.com/aws/aws-sdk-go-v2/service/mgn v1.33.4
	github.com/aws/aws-sdk-go-v2/service/mq v1.29.2
	github.com/aws/aws-sdk-go-v2/service/mwaa v1.35.3
	github.com/aws/aws-sdk-go-v2/service/neptune v1.37.3
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.17.5
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.51.0
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.35.1
	github.com/aws/aws-sdk-go-v2/service/networkmonitor v1.8.4
	github.com/aws/aws-sdk-go-v2/service/notifications v1.2.5
	github.com/aws/aws-sdk-go-v2/service/notificationscontacts v1.1.4
	github.com/aws/aws-sdk-go-v2/service/oam v1.18.3
	github.com/aws/aws-sdk-go-v2/service/opensearch v1.46.6
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.19.6
	github.com/aws/aws-sdk-go-v2/service/organizations v1.39.0
	github.com/aws/aws-sdk-go-v2/service/osis v1.15.5
	github.com/aws/aws-sdk-go-v2/service/outposts v1.51.0
	github.com/aws/aws-sdk-go-v2/service/paymentcryptography v1.19.0
	github.com/aws/aws-sdk-go-v2/service/pcaconnectorad v1.11.3
	github.com/aws/aws-sdk-go-v2/service/pcs v1.6.2
	github.com/aws/aws-sdk-go-v2/service/pinpoint v1.35.4
	github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2 v1.20.3
	github.com/aws/aws-sdk-go-v2/service/pipes v1.19.5
	github.com/aws/aws-sdk-go-v2/service/polly v1.48.4
	github.com/aws/aws-sdk-go-v2/service/pricing v1.34.5
	github.com/aws/aws-sdk-go-v2/service/qbusiness v1.28.0
	github.com/aws/aws-sdk-go-v2/service/qldb v1.26.4
	github.com/aws/aws-sdk-go-v2/service/quicksight v1.87.0
	github.com/aws/aws-sdk-go-v2/service/ram v1.30.6
	github.com/aws/aws-sdk-go-v2/service/rbin v1.22.6
	github.com/aws/aws-sdk-go-v2/service/rds v1.99.1
	github.com/aws/aws-sdk-go-v2/service/redshift v1.54.6
	github.com/aws/aws-sdk-go-v2/service/redshiftdata v1.33.3
	github.com/aws/aws-sdk-go-v2/service/redshiftserverless v1.27.4
	github.com/aws/aws-sdk-go-v2/service/rekognition v1.47.2
	github.com/aws/aws-sdk-go-v2/service/resiliencehub v1.30.4
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.17.6
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.29.3
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.26.6
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.17.5
	github.com/aws/aws-sdk-go-v2/service/route53 v1.53.0
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.29.4
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.5.9
	github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig v1.27.3
	github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness v1.22.4
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.36.0
	github.com/aws/aws-sdk-go-v2/service/rum v1.24.4
	github.com/aws/aws-sdk-go-v2/service/s3 v1.83.0
	github.com/aws/aws-sdk-go-v2/service/s3control v1.60.0
	github.com/aws/aws-sdk-go-v2/service/s3outposts v1.29.4
	github.com/aws/aws-sdk-go-v2/service/s3tables v1.5.0
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.200.0
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.13.9
	github.com/aws/aws-sdk-go-v2/service/schemas v1.29.5
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.35.7
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.58.0
	github.com/aws/aws-sdk-go-v2/service/securitylake v1.20.5
	github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository v1.25.4
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.34.2
	github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry v1.31.4
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.35.7
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.28.3
	github.com/aws/aws-sdk-go-v2/service/ses v1.30.5
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.46.0
	github.com/aws/aws-sdk-go-v2/service/sfn v1.35.7
	github.com/aws/aws-sdk-go-v2/service/shield v1.30.4
	github.com/aws/aws-sdk-go-v2/service/signer v1.27.4
	github.com/aws/aws-sdk-go-v2/service/sns v1.34.7
	github.com/aws/aws-sdk-go-v2/service/sqs v1.38.8
	github.com/aws/aws-sdk-go-v2/service/ssm v1.60.0
	github.com/aws/aws-sdk-go-v2/service/ssmcontacts v1.27.4
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.35.4
	github.com/aws/aws-sdk-go-v2/service/ssmquicksetup v1.4.4
	github.com/aws/aws-sdk-go-v2/service/ssmsap v1.20.4
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.5
	github.com/aws/aws-sdk-go-v2/service/ssoadmin v1.31.2
	github.com/aws/aws-sdk-go-v2/service/storagegateway v1.38.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.34.0
	github.com/aws/aws-sdk-go-v2/service/swf v1.28.6
	github.com/aws/aws-sdk-go-v2/service/synthetics v1.35.3
	github.com/aws/aws-sdk-go-v2/service/taxsettings v1.12.2
	github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb v1.10.5
	github.com/aws/aws-sdk-go-v2/service/timestreamquery v1.31.2
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.31.2
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.47.0
	github.com/aws/aws-sdk-go-v2/service/transfer v1.61.0
	github.com/aws/aws-sdk-go-v2/service/verifiedpermissions v1.24.2
	github.com/aws/aws-sdk-go-v2/service/vpclattice v1.14.4
	github.com/aws/aws-sdk-go-v2/service/waf v1.26.4
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.26.4
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.63.1
	github.com/aws/aws-sdk-go-v2/service/wellarchitected v1.35.4
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.58.0
	github.com/aws/aws-sdk-go-v2/service/workspacesweb v1.27.4
	github.com/aws/aws-sdk-go-v2/service/xray v1.31.7
	github.com/aws/smithy-go v1.22.4
	github.com/beevik/etree v1.5.1
	github.com/cedar-policy/cedar-go v0.1.0
	github.com/davecgh/go-spew v1.1.1
	github.com/dlclark/regexp2 v1.11.5
	github.com/gertd/go-pluralize v0.2.1
	github.com/google/go-cmp v0.7.0
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.23.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.65
	github.com/hashicorp/awspolicyequivalence v1.7.0
	github.com/hashicorp/cli v1.1.7
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.5.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.7.0
	github.com/hashicorp/hcl/v2 v2.23.0
	github.com/hashicorp/terraform-json v0.25.0
	github.com/hashicorp/terraform-plugin-framework v1.15.0
	github.com/hashicorp/terraform-plugin-framework-jsontypes v0.2.0
	github.com/hashicorp/terraform-plugin-framework-timeouts v0.5.0
	github.com/hashicorp/terraform-plugin-framework-timetypes v0.5.0
	github.com/hashicorp/terraform-plugin-framework-validators v0.18.0
	github.com/hashicorp/terraform-plugin-go v0.28.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-mux v0.20.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.37.0
	github.com/hashicorp/terraform-plugin-testing v1.13.2
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20240118010651-0ba75a80ca38
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pquerna/otp v1.5.0
	github.com/shopspring/decimal v1.4.0
	golang.org/x/crypto v0.39.0
	golang.org/x/text v0.26.0
	golang.org/x/tools v0.34.0
	gopkg.in/dnaeon/go-vcr.v4 v4.0.4
	gopkg.in/yaml.v3 v3.0.1
	syreclabs.com/go/faker v1.2.3
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.3 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-plugin v1.6.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/hc-install v0.9.2 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.23.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.5 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/posener/complete v1.2.3 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.16.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.61.0 // indirect
	go.opentelemetry.io/otel v1.36.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/grpc v1.72.1 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/hashicorp/terraform-plugin-log => github.com/gdavison/terraform-plugin-log v0.0.0-20230928191232-6c653d8ef8fb

replace github.com/aws/aws-sdk-go-v2 => /Users/kewbank/Downloads/aws-sdk-go-v2

replace github.com/aws/aws-sdk-go-v2/config => /Users/kewbank/Downloads/aws-sdk-go-v2/config

replace github.com/aws/aws-sdk-go-v2/credentials => /Users/kewbank/Downloads/aws-sdk-go-v2/credentials

replace github.com/aws/aws-sdk-go-v2/feature/ec2/imds => /Users/kewbank/Downloads/aws-sdk-go-v2/feature/ec2/imds

replace github.com/aws/aws-sdk-go-v2/feature/s3/manager => /Users/kewbank/Downloads/aws-sdk-go-v2/feature/s3/manager

replace github.com/aws/aws-sdk-go-v2/service/accessanalyzer => /Users/kewbank/Downloads/aws-sdk-go-v2/service/accessanalyzer

replace github.com/aws/aws-sdk-go-v2/service/account => /Users/kewbank/Downloads/aws-sdk-go-v2/service/account

replace github.com/aws/aws-sdk-go-v2/service/acm => /Users/kewbank/Downloads/aws-sdk-go-v2/service/acm

replace github.com/aws/aws-sdk-go-v2/service/acmpca => /Users/kewbank/Downloads/aws-sdk-go-v2/service/acmpca

replace github.com/aws/aws-sdk-go-v2/service/amp => /Users/kewbank/Downloads/aws-sdk-go-v2/service/amp

replace github.com/aws/aws-sdk-go-v2/service/amplify => /Users/kewbank/Downloads/aws-sdk-go-v2/service/amplify

replace github.com/aws/aws-sdk-go-v2/service/apigateway => /Users/kewbank/Downloads/aws-sdk-go-v2/service/apigateway

replace github.com/aws/aws-sdk-go-v2/service/apigatewayv2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/apigatewayv2

replace github.com/aws/aws-sdk-go-v2/service/appconfig => /Users/kewbank/Downloads/aws-sdk-go-v2/service/appconfig

replace github.com/aws/aws-sdk-go-v2/service/appfabric => /Users/kewbank/Downloads/aws-sdk-go-v2/service/appfabric

replace github.com/aws/aws-sdk-go-v2/service/appflow => /Users/kewbank/Downloads/aws-sdk-go-v2/service/appflow

replace github.com/aws/aws-sdk-go-v2/service/appintegrations => /Users/kewbank/Downloads/aws-sdk-go-v2/service/appintegrations

replace github.com/aws/aws-sdk-go-v2/service/applicationautoscaling => /Users/kewbank/Downloads/aws-sdk-go-v2/service/applicationautoscaling

replace github.com/aws/aws-sdk-go-v2/service/applicationinsights => /Users/kewbank/Downloads/aws-sdk-go-v2/service/applicationinsights

replace github.com/aws/aws-sdk-go-v2/service/applicationsignals => /Users/kewbank/Downloads/aws-sdk-go-v2/service/applicationsignals

replace github.com/aws/aws-sdk-go-v2/service/appmesh => /Users/kewbank/Downloads/aws-sdk-go-v2/service/appmesh

replace github.com/aws/aws-sdk-go-v2/service/apprunner => /Users/kewbank/Downloads/aws-sdk-go-v2/service/apprunner

replace github.com/aws/aws-sdk-go-v2/service/appstream => /Users/kewbank/Downloads/aws-sdk-go-v2/service/appstream

replace github.com/aws/aws-sdk-go-v2/service/appsync => /Users/kewbank/Downloads/aws-sdk-go-v2/service/appsync

replace github.com/aws/aws-sdk-go-v2/service/athena => /Users/kewbank/Downloads/aws-sdk-go-v2/service/athena

replace github.com/aws/aws-sdk-go-v2/service/auditmanager => /Users/kewbank/Downloads/aws-sdk-go-v2/service/auditmanager

replace github.com/aws/aws-sdk-go-v2/service/autoscaling => /Users/kewbank/Downloads/aws-sdk-go-v2/service/autoscaling

replace github.com/aws/aws-sdk-go-v2/service/autoscalingplans => /Users/kewbank/Downloads/aws-sdk-go-v2/service/autoscalingplans

replace github.com/aws/aws-sdk-go-v2/service/backup => /Users/kewbank/Downloads/aws-sdk-go-v2/service/backup

replace github.com/aws/aws-sdk-go-v2/service/batch => /Users/kewbank/Downloads/aws-sdk-go-v2/service/batch

replace github.com/aws/aws-sdk-go-v2/service/bcmdataexports => /Users/kewbank/Downloads/aws-sdk-go-v2/service/bcmdataexports

replace github.com/aws/aws-sdk-go-v2/service/bedrock => /Users/kewbank/Downloads/aws-sdk-go-v2/service/bedrock

replace github.com/aws/aws-sdk-go-v2/service/bedrockagent => /Users/kewbank/Downloads/aws-sdk-go-v2/service/bedrockagent

replace github.com/aws/aws-sdk-go-v2/service/billing => /Users/kewbank/Downloads/aws-sdk-go-v2/service/billing

replace github.com/aws/aws-sdk-go-v2/service/budgets => /Users/kewbank/Downloads/aws-sdk-go-v2/service/budgets

replace github.com/aws/aws-sdk-go-v2/service/chatbot => /Users/kewbank/Downloads/aws-sdk-go-v2/service/chatbot

replace github.com/aws/aws-sdk-go-v2/service/chime => /Users/kewbank/Downloads/aws-sdk-go-v2/service/chime

replace github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines => /Users/kewbank/Downloads/aws-sdk-go-v2/service/chimesdkmediapipelines

replace github.com/aws/aws-sdk-go-v2/service/chimesdkvoice => /Users/kewbank/Downloads/aws-sdk-go-v2/service/chimesdkvoice

replace github.com/aws/aws-sdk-go-v2/service/cleanrooms => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cleanrooms

replace github.com/aws/aws-sdk-go-v2/service/cloud9 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloud9

replace github.com/aws/aws-sdk-go-v2/service/cloudcontrol => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloudcontrol

replace github.com/aws/aws-sdk-go-v2/service/cloudformation => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloudformation

replace github.com/aws/aws-sdk-go-v2/service/cloudfront => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloudfront

replace github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloudfrontkeyvaluestore

replace github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloudhsmv2

replace github.com/aws/aws-sdk-go-v2/service/cloudsearch => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloudsearch

replace github.com/aws/aws-sdk-go-v2/service/cloudtrail => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloudtrail

replace github.com/aws/aws-sdk-go-v2/service/cloudwatch => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloudwatch

replace github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cloudwatchlogs

replace github.com/aws/aws-sdk-go-v2/service/codeartifact => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codeartifact

replace github.com/aws/aws-sdk-go-v2/service/codebuild => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codebuild

replace github.com/aws/aws-sdk-go-v2/service/codecatalyst => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codecatalyst

replace github.com/aws/aws-sdk-go-v2/service/codecommit => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codecommit

replace github.com/aws/aws-sdk-go-v2/service/codeconnections => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codeconnections

replace github.com/aws/aws-sdk-go-v2/service/codedeploy => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codedeploy

replace github.com/aws/aws-sdk-go-v2/service/codeguruprofiler => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codeguruprofiler

replace github.com/aws/aws-sdk-go-v2/service/codegurureviewer => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codegurureviewer

replace github.com/aws/aws-sdk-go-v2/service/codepipeline => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codepipeline

replace github.com/aws/aws-sdk-go-v2/service/codestarconnections => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codestarconnections

replace github.com/aws/aws-sdk-go-v2/service/codestarnotifications => /Users/kewbank/Downloads/aws-sdk-go-v2/service/codestarnotifications

replace github.com/aws/aws-sdk-go-v2/service/cognitoidentity => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cognitoidentity

replace github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider => /Users/kewbank/Downloads/aws-sdk-go-v2/service/cognitoidentityprovider

replace github.com/aws/aws-sdk-go-v2/service/comprehend => /Users/kewbank/Downloads/aws-sdk-go-v2/service/comprehend

replace github.com/aws/aws-sdk-go-v2/service/computeoptimizer => /Users/kewbank/Downloads/aws-sdk-go-v2/service/computeoptimizer

replace github.com/aws/aws-sdk-go-v2/service/configservice => /Users/kewbank/Downloads/aws-sdk-go-v2/service/configservice

replace github.com/aws/aws-sdk-go-v2/service/connect => /Users/kewbank/Downloads/aws-sdk-go-v2/service/connect

replace github.com/aws/aws-sdk-go-v2/service/connectcases => /Users/kewbank/Downloads/aws-sdk-go-v2/service/connectcases

replace github.com/aws/aws-sdk-go-v2/service/controltower => /Users/kewbank/Downloads/aws-sdk-go-v2/service/controltower

replace github.com/aws/aws-sdk-go-v2/service/costandusagereportservice => /Users/kewbank/Downloads/aws-sdk-go-v2/service/costandusagereportservice

replace github.com/aws/aws-sdk-go-v2/service/costexplorer => /Users/kewbank/Downloads/aws-sdk-go-v2/service/costexplorer

replace github.com/aws/aws-sdk-go-v2/service/costoptimizationhub => /Users/kewbank/Downloads/aws-sdk-go-v2/service/costoptimizationhub

replace github.com/aws/aws-sdk-go-v2/service/customerprofiles => /Users/kewbank/Downloads/aws-sdk-go-v2/service/customerprofiles

replace github.com/aws/aws-sdk-go-v2/service/databasemigrationservice => /Users/kewbank/Downloads/aws-sdk-go-v2/service/databasemigrationservice

replace github.com/aws/aws-sdk-go-v2/service/databrew => /Users/kewbank/Downloads/aws-sdk-go-v2/service/databrew

replace github.com/aws/aws-sdk-go-v2/service/dataexchange => /Users/kewbank/Downloads/aws-sdk-go-v2/service/dataexchange

replace github.com/aws/aws-sdk-go-v2/service/datapipeline => /Users/kewbank/Downloads/aws-sdk-go-v2/service/datapipeline

replace github.com/aws/aws-sdk-go-v2/service/datasync => /Users/kewbank/Downloads/aws-sdk-go-v2/service/datasync

replace github.com/aws/aws-sdk-go-v2/service/datazone => /Users/kewbank/Downloads/aws-sdk-go-v2/service/datazone

replace github.com/aws/aws-sdk-go-v2/service/dax => /Users/kewbank/Downloads/aws-sdk-go-v2/service/dax

replace github.com/aws/aws-sdk-go-v2/service/detective => /Users/kewbank/Downloads/aws-sdk-go-v2/service/detective

replace github.com/aws/aws-sdk-go-v2/service/devicefarm => /Users/kewbank/Downloads/aws-sdk-go-v2/service/devicefarm

replace github.com/aws/aws-sdk-go-v2/service/devopsguru => /Users/kewbank/Downloads/aws-sdk-go-v2/service/devopsguru

replace github.com/aws/aws-sdk-go-v2/service/directconnect => /Users/kewbank/Downloads/aws-sdk-go-v2/service/directconnect

replace github.com/aws/aws-sdk-go-v2/service/directoryservice => /Users/kewbank/Downloads/aws-sdk-go-v2/service/directoryservice

replace github.com/aws/aws-sdk-go-v2/service/dlm => /Users/kewbank/Downloads/aws-sdk-go-v2/service/dlm

replace github.com/aws/aws-sdk-go-v2/service/docdb => /Users/kewbank/Downloads/aws-sdk-go-v2/service/docdb

replace github.com/aws/aws-sdk-go-v2/service/docdbelastic => /Users/kewbank/Downloads/aws-sdk-go-v2/service/docdbelastic

replace github.com/aws/aws-sdk-go-v2/service/drs => /Users/kewbank/Downloads/aws-sdk-go-v2/service/drs

replace github.com/aws/aws-sdk-go-v2/service/dsql => /Users/kewbank/Downloads/aws-sdk-go-v2/service/dsql

replace github.com/aws/aws-sdk-go-v2/service/dynamodb => /Users/kewbank/Downloads/aws-sdk-go-v2/service/dynamodb

replace github.com/aws/aws-sdk-go-v2/service/ec2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ec2

replace github.com/aws/aws-sdk-go-v2/service/ecr => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ecr

replace github.com/aws/aws-sdk-go-v2/service/ecrpublic => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ecrpublic

replace github.com/aws/aws-sdk-go-v2/service/ecs => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ecs

replace github.com/aws/aws-sdk-go-v2/service/efs => /Users/kewbank/Downloads/aws-sdk-go-v2/service/efs

replace github.com/aws/aws-sdk-go-v2/service/eks => /Users/kewbank/Downloads/aws-sdk-go-v2/service/eks

replace github.com/aws/aws-sdk-go-v2/service/elasticache => /Users/kewbank/Downloads/aws-sdk-go-v2/service/elasticache

replace github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk => /Users/kewbank/Downloads/aws-sdk-go-v2/service/elasticbeanstalk

replace github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing => /Users/kewbank/Downloads/aws-sdk-go-v2/service/elasticloadbalancing

replace github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/elasticloadbalancingv2

replace github.com/aws/aws-sdk-go-v2/service/elasticsearchservice => /Users/kewbank/Downloads/aws-sdk-go-v2/service/elasticsearchservice

replace github.com/aws/aws-sdk-go-v2/service/elastictranscoder => /Users/kewbank/Downloads/aws-sdk-go-v2/service/elastictranscoder

replace github.com/aws/aws-sdk-go-v2/service/emr => /Users/kewbank/Downloads/aws-sdk-go-v2/service/emr

replace github.com/aws/aws-sdk-go-v2/service/emrcontainers => /Users/kewbank/Downloads/aws-sdk-go-v2/service/emrcontainers

replace github.com/aws/aws-sdk-go-v2/service/emrserverless => /Users/kewbank/Downloads/aws-sdk-go-v2/service/emrserverless

replace github.com/aws/aws-sdk-go-v2/service/eventbridge => /Users/kewbank/Downloads/aws-sdk-go-v2/service/eventbridge

replace github.com/aws/aws-sdk-go-v2/service/evidently => /Users/kewbank/Downloads/aws-sdk-go-v2/service/evidently

replace github.com/aws/aws-sdk-go-v2/service/evs => /Users/kewbank/Downloads/aws-sdk-go-v2/service/evs

replace github.com/aws/aws-sdk-go-v2/service/finspace => /Users/kewbank/Downloads/aws-sdk-go-v2/service/finspace

replace github.com/aws/aws-sdk-go-v2/service/firehose => /Users/kewbank/Downloads/aws-sdk-go-v2/service/firehose

replace github.com/aws/aws-sdk-go-v2/service/fis => /Users/kewbank/Downloads/aws-sdk-go-v2/service/fis

replace github.com/aws/aws-sdk-go-v2/service/fms => /Users/kewbank/Downloads/aws-sdk-go-v2/service/fms

replace github.com/aws/aws-sdk-go-v2/service/fsx => /Users/kewbank/Downloads/aws-sdk-go-v2/service/fsx

replace github.com/aws/aws-sdk-go-v2/service/gamelift => /Users/kewbank/Downloads/aws-sdk-go-v2/service/gamelift

replace github.com/aws/aws-sdk-go-v2/service/glacier => /Users/kewbank/Downloads/aws-sdk-go-v2/service/glacier

replace github.com/aws/aws-sdk-go-v2/service/globalaccelerator => /Users/kewbank/Downloads/aws-sdk-go-v2/service/globalaccelerator

replace github.com/aws/aws-sdk-go-v2/service/glue => /Users/kewbank/Downloads/aws-sdk-go-v2/service/glue

replace github.com/aws/aws-sdk-go-v2/service/grafana => /Users/kewbank/Downloads/aws-sdk-go-v2/service/grafana

replace github.com/aws/aws-sdk-go-v2/service/greengrass => /Users/kewbank/Downloads/aws-sdk-go-v2/service/greengrass

replace github.com/aws/aws-sdk-go-v2/service/groundstation => /Users/kewbank/Downloads/aws-sdk-go-v2/service/groundstation

replace github.com/aws/aws-sdk-go-v2/service/guardduty => /Users/kewbank/Downloads/aws-sdk-go-v2/service/guardduty

replace github.com/aws/aws-sdk-go-v2/service/healthlake => /Users/kewbank/Downloads/aws-sdk-go-v2/service/healthlake

replace github.com/aws/aws-sdk-go-v2/service/iam => /Users/kewbank/Downloads/aws-sdk-go-v2/service/iam

replace github.com/aws/aws-sdk-go-v2/service/identitystore => /Users/kewbank/Downloads/aws-sdk-go-v2/service/identitystore

replace github.com/aws/aws-sdk-go-v2/service/imagebuilder => /Users/kewbank/Downloads/aws-sdk-go-v2/service/imagebuilder

replace github.com/aws/aws-sdk-go-v2/service/inspector => /Users/kewbank/Downloads/aws-sdk-go-v2/service/inspector

replace github.com/aws/aws-sdk-go-v2/service/inspector2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/inspector2

replace github.com/aws/aws-sdk-go-v2/service/internetmonitor => /Users/kewbank/Downloads/aws-sdk-go-v2/service/internetmonitor

replace github.com/aws/aws-sdk-go-v2/service/invoicing => /Users/kewbank/Downloads/aws-sdk-go-v2/service/invoicing

replace github.com/aws/aws-sdk-go-v2/service/iot => /Users/kewbank/Downloads/aws-sdk-go-v2/service/iot

replace github.com/aws/aws-sdk-go-v2/service/ivs => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ivs

replace github.com/aws/aws-sdk-go-v2/service/ivschat => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ivschat

replace github.com/aws/aws-sdk-go-v2/service/kafka => /Users/kewbank/Downloads/aws-sdk-go-v2/service/kafka

replace github.com/aws/aws-sdk-go-v2/service/kafkaconnect => /Users/kewbank/Downloads/aws-sdk-go-v2/service/kafkaconnect

replace github.com/aws/aws-sdk-go-v2/service/kendra => /Users/kewbank/Downloads/aws-sdk-go-v2/service/kendra

replace github.com/aws/aws-sdk-go-v2/service/keyspaces => /Users/kewbank/Downloads/aws-sdk-go-v2/service/keyspaces

replace github.com/aws/aws-sdk-go-v2/service/kinesis => /Users/kewbank/Downloads/aws-sdk-go-v2/service/kinesis

replace github.com/aws/aws-sdk-go-v2/service/kinesisanalytics => /Users/kewbank/Downloads/aws-sdk-go-v2/service/kinesisanalytics

replace github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/kinesisanalyticsv2

replace github.com/aws/aws-sdk-go-v2/service/kinesisvideo => /Users/kewbank/Downloads/aws-sdk-go-v2/service/kinesisvideo

replace github.com/aws/aws-sdk-go-v2/service/kms => /Users/kewbank/Downloads/aws-sdk-go-v2/service/kms

replace github.com/aws/aws-sdk-go-v2/service/lakeformation => /Users/kewbank/Downloads/aws-sdk-go-v2/service/lakeformation

replace github.com/aws/aws-sdk-go-v2/service/lambda => /Users/kewbank/Downloads/aws-sdk-go-v2/service/lambda

replace github.com/aws/aws-sdk-go-v2/service/launchwizard => /Users/kewbank/Downloads/aws-sdk-go-v2/service/launchwizard

replace github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice => /Users/kewbank/Downloads/aws-sdk-go-v2/service/lexmodelbuildingservice

replace github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/lexmodelsv2

replace github.com/aws/aws-sdk-go-v2/service/licensemanager => /Users/kewbank/Downloads/aws-sdk-go-v2/service/licensemanager

replace github.com/aws/aws-sdk-go-v2/service/lightsail => /Users/kewbank/Downloads/aws-sdk-go-v2/service/lightsail

replace github.com/aws/aws-sdk-go-v2/service/location => /Users/kewbank/Downloads/aws-sdk-go-v2/service/location

replace github.com/aws/aws-sdk-go-v2/service/lookoutmetrics => /Users/kewbank/Downloads/aws-sdk-go-v2/service/lookoutmetrics

replace github.com/aws/aws-sdk-go-v2/service/m2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/m2

replace github.com/aws/aws-sdk-go-v2/service/macie2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/macie2

replace github.com/aws/aws-sdk-go-v2/service/mediaconnect => /Users/kewbank/Downloads/aws-sdk-go-v2/service/mediaconnect

replace github.com/aws/aws-sdk-go-v2/service/mediaconvert => /Users/kewbank/Downloads/aws-sdk-go-v2/service/mediaconvert

replace github.com/aws/aws-sdk-go-v2/service/medialive => /Users/kewbank/Downloads/aws-sdk-go-v2/service/medialive

replace github.com/aws/aws-sdk-go-v2/service/mediapackage => /Users/kewbank/Downloads/aws-sdk-go-v2/service/mediapackage

replace github.com/aws/aws-sdk-go-v2/service/mediapackagev2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/mediapackagev2

replace github.com/aws/aws-sdk-go-v2/service/mediapackagevod => /Users/kewbank/Downloads/aws-sdk-go-v2/service/mediapackagevod

replace github.com/aws/aws-sdk-go-v2/service/mediastore => /Users/kewbank/Downloads/aws-sdk-go-v2/service/mediastore

replace github.com/aws/aws-sdk-go-v2/service/memorydb => /Users/kewbank/Downloads/aws-sdk-go-v2/service/memorydb

replace github.com/aws/aws-sdk-go-v2/service/mgn => /Users/kewbank/Downloads/aws-sdk-go-v2/service/mgn

replace github.com/aws/aws-sdk-go-v2/service/mq => /Users/kewbank/Downloads/aws-sdk-go-v2/service/mq

replace github.com/aws/aws-sdk-go-v2/service/mwaa => /Users/kewbank/Downloads/aws-sdk-go-v2/service/mwaa

replace github.com/aws/aws-sdk-go-v2/service/neptune => /Users/kewbank/Downloads/aws-sdk-go-v2/service/neptune

replace github.com/aws/aws-sdk-go-v2/service/neptunegraph => /Users/kewbank/Downloads/aws-sdk-go-v2/service/neptunegraph

replace github.com/aws/aws-sdk-go-v2/service/networkfirewall => /Users/kewbank/Downloads/aws-sdk-go-v2/service/networkfirewall

replace github.com/aws/aws-sdk-go-v2/service/networkmanager => /Users/kewbank/Downloads/aws-sdk-go-v2/service/networkmanager

replace github.com/aws/aws-sdk-go-v2/service/networkmonitor => /Users/kewbank/Downloads/aws-sdk-go-v2/service/networkmonitor

replace github.com/aws/aws-sdk-go-v2/service/notifications => /Users/kewbank/Downloads/aws-sdk-go-v2/service/notifications

replace github.com/aws/aws-sdk-go-v2/service/notificationscontacts => /Users/kewbank/Downloads/aws-sdk-go-v2/service/notificationscontacts

replace github.com/aws/aws-sdk-go-v2/service/oam => /Users/kewbank/Downloads/aws-sdk-go-v2/service/oam

replace github.com/aws/aws-sdk-go-v2/service/opensearch => /Users/kewbank/Downloads/aws-sdk-go-v2/service/opensearch

replace github.com/aws/aws-sdk-go-v2/service/opensearchserverless => /Users/kewbank/Downloads/aws-sdk-go-v2/service/opensearchserverless

replace github.com/aws/aws-sdk-go-v2/service/organizations => /Users/kewbank/Downloads/aws-sdk-go-v2/service/organizations

replace github.com/aws/aws-sdk-go-v2/service/osis => /Users/kewbank/Downloads/aws-sdk-go-v2/service/osis

replace github.com/aws/aws-sdk-go-v2/service/outposts => /Users/kewbank/Downloads/aws-sdk-go-v2/service/outposts

replace github.com/aws/aws-sdk-go-v2/service/paymentcryptography => /Users/kewbank/Downloads/aws-sdk-go-v2/service/paymentcryptography

replace github.com/aws/aws-sdk-go-v2/service/pcaconnectorad => /Users/kewbank/Downloads/aws-sdk-go-v2/service/pcaconnectorad

replace github.com/aws/aws-sdk-go-v2/service/pcs => /Users/kewbank/Downloads/aws-sdk-go-v2/service/pcs

replace github.com/aws/aws-sdk-go-v2/service/pinpoint => /Users/kewbank/Downloads/aws-sdk-go-v2/service/pinpoint

replace github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/pinpointsmsvoicev2

replace github.com/aws/aws-sdk-go-v2/service/pipes => /Users/kewbank/Downloads/aws-sdk-go-v2/service/pipes

replace github.com/aws/aws-sdk-go-v2/service/polly => /Users/kewbank/Downloads/aws-sdk-go-v2/service/polly

replace github.com/aws/aws-sdk-go-v2/service/pricing => /Users/kewbank/Downloads/aws-sdk-go-v2/service/pricing

replace github.com/aws/aws-sdk-go-v2/service/qbusiness => /Users/kewbank/Downloads/aws-sdk-go-v2/service/qbusiness

replace github.com/aws/aws-sdk-go-v2/service/qldb => /Users/kewbank/Downloads/aws-sdk-go-v2/service/qldb

replace github.com/aws/aws-sdk-go-v2/service/quicksight => /Users/kewbank/Downloads/aws-sdk-go-v2/service/quicksight

replace github.com/aws/aws-sdk-go-v2/service/ram => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ram

replace github.com/aws/aws-sdk-go-v2/service/rbin => /Users/kewbank/Downloads/aws-sdk-go-v2/service/rbin

replace github.com/aws/aws-sdk-go-v2/service/rds => /Users/kewbank/Downloads/aws-sdk-go-v2/service/rds

replace github.com/aws/aws-sdk-go-v2/service/redshift => /Users/kewbank/Downloads/aws-sdk-go-v2/service/redshift

replace github.com/aws/aws-sdk-go-v2/service/redshiftdata => /Users/kewbank/Downloads/aws-sdk-go-v2/service/redshiftdata

replace github.com/aws/aws-sdk-go-v2/service/redshiftserverless => /Users/kewbank/Downloads/aws-sdk-go-v2/service/redshiftserverless

replace github.com/aws/aws-sdk-go-v2/service/rekognition => /Users/kewbank/Downloads/aws-sdk-go-v2/service/rekognition

replace github.com/aws/aws-sdk-go-v2/service/resiliencehub => /Users/kewbank/Downloads/aws-sdk-go-v2/service/resiliencehub

replace github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/resourceexplorer2

replace github.com/aws/aws-sdk-go-v2/service/resourcegroups => /Users/kewbank/Downloads/aws-sdk-go-v2/service/resourcegroups

replace github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi => /Users/kewbank/Downloads/aws-sdk-go-v2/service/resourcegroupstaggingapi

replace github.com/aws/aws-sdk-go-v2/service/rolesanywhere => /Users/kewbank/Downloads/aws-sdk-go-v2/service/rolesanywhere

replace github.com/aws/aws-sdk-go-v2/service/route53 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/route53

replace github.com/aws/aws-sdk-go-v2/service/route53domains => /Users/kewbank/Downloads/aws-sdk-go-v2/service/route53domains

replace github.com/aws/aws-sdk-go-v2/service/route53profiles => /Users/kewbank/Downloads/aws-sdk-go-v2/service/route53profiles

replace github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig => /Users/kewbank/Downloads/aws-sdk-go-v2/service/route53recoverycontrolconfig

replace github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness => /Users/kewbank/Downloads/aws-sdk-go-v2/service/route53recoveryreadiness

replace github.com/aws/aws-sdk-go-v2/service/route53resolver => /Users/kewbank/Downloads/aws-sdk-go-v2/service/route53resolver

replace github.com/aws/aws-sdk-go-v2/service/rum => /Users/kewbank/Downloads/aws-sdk-go-v2/service/rum

replace github.com/aws/aws-sdk-go-v2/service/s3 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/s3

replace github.com/aws/aws-sdk-go-v2/service/s3control => /Users/kewbank/Downloads/aws-sdk-go-v2/service/s3control

replace github.com/aws/aws-sdk-go-v2/service/s3outposts => /Users/kewbank/Downloads/aws-sdk-go-v2/service/s3outposts

replace github.com/aws/aws-sdk-go-v2/service/s3tables => /Users/kewbank/Downloads/aws-sdk-go-v2/service/s3tables

replace github.com/aws/aws-sdk-go-v2/service/sagemaker => /Users/kewbank/Downloads/aws-sdk-go-v2/service/sagemaker

replace github.com/aws/aws-sdk-go-v2/service/scheduler => /Users/kewbank/Downloads/aws-sdk-go-v2/service/scheduler

replace github.com/aws/aws-sdk-go-v2/service/schemas => /Users/kewbank/Downloads/aws-sdk-go-v2/service/schemas

replace github.com/aws/aws-sdk-go-v2/service/secretsmanager => /Users/kewbank/Downloads/aws-sdk-go-v2/service/secretsmanager

replace github.com/aws/aws-sdk-go-v2/service/securityhub => /Users/kewbank/Downloads/aws-sdk-go-v2/service/securityhub

replace github.com/aws/aws-sdk-go-v2/service/securitylake => /Users/kewbank/Downloads/aws-sdk-go-v2/service/securitylake

replace github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository => /Users/kewbank/Downloads/aws-sdk-go-v2/service/serverlessapplicationrepository

replace github.com/aws/aws-sdk-go-v2/service/servicecatalog => /Users/kewbank/Downloads/aws-sdk-go-v2/service/servicecatalog

replace github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry => /Users/kewbank/Downloads/aws-sdk-go-v2/service/servicecatalogappregistry

replace github.com/aws/aws-sdk-go-v2/service/servicediscovery => /Users/kewbank/Downloads/aws-sdk-go-v2/service/servicediscovery

replace github.com/aws/aws-sdk-go-v2/service/servicequotas => /Users/kewbank/Downloads/aws-sdk-go-v2/service/servicequotas

replace github.com/aws/aws-sdk-go-v2/service/ses => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ses

replace github.com/aws/aws-sdk-go-v2/service/sesv2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/sesv2

replace github.com/aws/aws-sdk-go-v2/service/sfn => /Users/kewbank/Downloads/aws-sdk-go-v2/service/sfn

replace github.com/aws/aws-sdk-go-v2/service/shield => /Users/kewbank/Downloads/aws-sdk-go-v2/service/shield

replace github.com/aws/aws-sdk-go-v2/service/signer => /Users/kewbank/Downloads/aws-sdk-go-v2/service/signer

replace github.com/aws/aws-sdk-go-v2/service/sns => /Users/kewbank/Downloads/aws-sdk-go-v2/service/sns

replace github.com/aws/aws-sdk-go-v2/service/sqs => /Users/kewbank/Downloads/aws-sdk-go-v2/service/sqs

replace github.com/aws/aws-sdk-go-v2/service/ssm => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ssm

replace github.com/aws/aws-sdk-go-v2/service/ssmcontacts => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ssmcontacts

replace github.com/aws/aws-sdk-go-v2/service/ssmincidents => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ssmincidents

replace github.com/aws/aws-sdk-go-v2/service/ssmquicksetup => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ssmquicksetup

replace github.com/aws/aws-sdk-go-v2/service/ssmsap => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ssmsap

replace github.com/aws/aws-sdk-go-v2/service/sso => /Users/kewbank/Downloads/aws-sdk-go-v2/service/sso

replace github.com/aws/aws-sdk-go-v2/service/ssoadmin => /Users/kewbank/Downloads/aws-sdk-go-v2/service/ssoadmin

replace github.com/aws/aws-sdk-go-v2/service/storagegateway => /Users/kewbank/Downloads/aws-sdk-go-v2/service/storagegateway

replace github.com/aws/aws-sdk-go-v2/service/sts => /Users/kewbank/Downloads/aws-sdk-go-v2/service/sts

replace github.com/aws/aws-sdk-go-v2/service/swf => /Users/kewbank/Downloads/aws-sdk-go-v2/service/swf

replace github.com/aws/aws-sdk-go-v2/service/synthetics => /Users/kewbank/Downloads/aws-sdk-go-v2/service/synthetics

replace github.com/aws/aws-sdk-go-v2/service/taxsettings => /Users/kewbank/Downloads/aws-sdk-go-v2/service/taxsettings

replace github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb => /Users/kewbank/Downloads/aws-sdk-go-v2/service/timestreaminfluxdb

replace github.com/aws/aws-sdk-go-v2/service/timestreamquery => /Users/kewbank/Downloads/aws-sdk-go-v2/service/timestreamquery

replace github.com/aws/aws-sdk-go-v2/service/timestreamwrite => /Users/kewbank/Downloads/aws-sdk-go-v2/service/timestreamwrite

replace github.com/aws/aws-sdk-go-v2/service/transcribe => /Users/kewbank/Downloads/aws-sdk-go-v2/service/transcribe

replace github.com/aws/aws-sdk-go-v2/service/transfer => /Users/kewbank/Downloads/aws-sdk-go-v2/service/transfer

replace github.com/aws/aws-sdk-go-v2/service/verifiedpermissions => /Users/kewbank/Downloads/aws-sdk-go-v2/service/verifiedpermissions

replace github.com/aws/aws-sdk-go-v2/service/vpclattice => /Users/kewbank/Downloads/aws-sdk-go-v2/service/vpclattice

replace github.com/aws/aws-sdk-go-v2/service/waf => /Users/kewbank/Downloads/aws-sdk-go-v2/service/waf

replace github.com/aws/aws-sdk-go-v2/service/wafregional => /Users/kewbank/Downloads/aws-sdk-go-v2/service/wafregional

replace github.com/aws/aws-sdk-go-v2/service/wafv2 => /Users/kewbank/Downloads/aws-sdk-go-v2/service/wafv2

replace github.com/aws/aws-sdk-go-v2/service/wellarchitected => /Users/kewbank/Downloads/aws-sdk-go-v2/service/wellarchitected

replace github.com/aws/aws-sdk-go-v2/service/workspaces => /Users/kewbank/Downloads/aws-sdk-go-v2/service/workspaces

replace github.com/aws/aws-sdk-go-v2/service/workspacesweb => /Users/kewbank/Downloads/aws-sdk-go-v2/service/workspacesweb

replace github.com/aws/aws-sdk-go-v2/service/xray => /Users/kewbank/Downloads/aws-sdk-go-v2/service/xray
