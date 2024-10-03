module github.com/hashicorp/terraform-provider-aws

go 1.23.1

// Disable experimental post-quantum key exchange mechanism X25519Kyber768Draft00
// This was causing errors with AWS Network Firewall
godebug tlskyber=0

require (
	github.com/ProtonMail/go-crypto v1.1.0-beta.0-proton
	github.com/YakDriver/go-version v0.1.0
	github.com/YakDriver/regexache v0.24.0
	github.com/aws/aws-sdk-go v1.55.5
	github.com/aws/aws-sdk-go-v2 v1.31.0
	github.com/aws/aws-sdk-go-v2/config v1.27.39
	github.com/aws/aws-sdk-go-v2/credentials v1.17.37
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.14
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.26
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.33.3
	github.com/aws/aws-sdk-go-v2/service/account v1.20.3
	github.com/aws/aws-sdk-go-v2/service/acm v1.29.3
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.36.3
	github.com/aws/aws-sdk-go-v2/service/amp v1.28.3
	github.com/aws/aws-sdk-go-v2/service/amplify v1.25.3
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.26.3
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.23.3
	github.com/aws/aws-sdk-go-v2/service/appconfig v1.33.3
	github.com/aws/aws-sdk-go-v2/service/appfabric v1.10.3
	github.com/aws/aws-sdk-go-v2/service/appflow v1.44.3
	github.com/aws/aws-sdk-go-v2/service/appintegrations v1.29.3
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.32.3
	github.com/aws/aws-sdk-go-v2/service/applicationinsights v1.27.4
	github.com/aws/aws-sdk-go-v2/service/applicationsignals v1.5.3
	github.com/aws/aws-sdk-go-v2/service/appmesh v1.28.3
	github.com/aws/aws-sdk-go-v2/service/apprunner v1.31.3
	github.com/aws/aws-sdk-go-v2/service/appstream v1.40.0
	github.com/aws/aws-sdk-go-v2/service/appsync v1.37.3
	github.com/aws/aws-sdk-go-v2/service/athena v1.46.2
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.36.3
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.44.3
	github.com/aws/aws-sdk-go-v2/service/autoscalingplans v1.23.3
	github.com/aws/aws-sdk-go-v2/service/backup v1.38.3
	github.com/aws/aws-sdk-go-v2/service/batch v1.45.3
	github.com/aws/aws-sdk-go-v2/service/bcmdataexports v1.6.3
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.19.3
	github.com/aws/aws-sdk-go-v2/service/bedrockagent v1.22.0
	github.com/aws/aws-sdk-go-v2/service/budgets v1.27.2
	github.com/aws/aws-sdk-go-v2/service/chatbot v1.7.1
	github.com/aws/aws-sdk-go-v2/service/chime v1.33.3
	github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines v1.19.3
	github.com/aws/aws-sdk-go-v2/service/chimesdkvoice v1.18.3
	github.com/aws/aws-sdk-go-v2/service/cleanrooms v1.17.3
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.27.3
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.21.3
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.54.3
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.39.3
	github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore v1.7.3
	github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 v1.26.3
	github.com/aws/aws-sdk-go-v2/service/cloudsearch v1.25.3
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.43.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.41.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.40.3
	github.com/aws/aws-sdk-go-v2/service/codeartifact v1.32.0
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.44.3
	github.com/aws/aws-sdk-go-v2/service/codecatalyst v1.16.3
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.26.3
	github.com/aws/aws-sdk-go-v2/service/codeconnections v1.4.3
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.28.3
	github.com/aws/aws-sdk-go-v2/service/codeguruprofiler v1.23.3
	github.com/aws/aws-sdk-go-v2/service/codegurureviewer v1.28.3
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.32.3
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.28.3
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.25.3
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.26.3
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.45.3
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.34.3
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.38.3
	github.com/aws/aws-sdk-go-v2/service/configservice v1.49.3
	github.com/aws/aws-sdk-go-v2/service/connect v1.111.0
	github.com/aws/aws-sdk-go-v2/service/connectcases v1.20.3
	github.com/aws/aws-sdk-go-v2/service/controltower v1.17.3
	github.com/aws/aws-sdk-go-v2/service/costandusagereportservice v1.27.3
	github.com/aws/aws-sdk-go-v2/service/costexplorer v1.42.3
	github.com/aws/aws-sdk-go-v2/service/costoptimizationhub v1.9.3
	github.com/aws/aws-sdk-go-v2/service/customerprofiles v1.41.0
	github.com/aws/aws-sdk-go-v2/service/databasemigrationservice v1.41.3
	github.com/aws/aws-sdk-go-v2/service/databrew v1.32.3
	github.com/aws/aws-sdk-go-v2/service/dataexchange v1.31.3
	github.com/aws/aws-sdk-go-v2/service/datapipeline v1.24.3
	github.com/aws/aws-sdk-go-v2/service/datasync v1.41.3
	github.com/aws/aws-sdk-go-v2/service/datazone v1.21.3
	github.com/aws/aws-sdk-go-v2/service/dax v1.22.3
	github.com/aws/aws-sdk-go-v2/service/detective v1.30.3
	github.com/aws/aws-sdk-go-v2/service/devicefarm v1.27.3
	github.com/aws/aws-sdk-go-v2/service/devopsguru v1.33.3
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.28.3
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.29.3
	github.com/aws/aws-sdk-go-v2/service/dlm v1.27.3
	github.com/aws/aws-sdk-go-v2/service/docdb v1.38.3
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.12.3
	github.com/aws/aws-sdk-go-v2/service/drs v1.29.3
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.35.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.179.2
	github.com/aws/aws-sdk-go-v2/service/ecr v1.35.3
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.26.3
	github.com/aws/aws-sdk-go-v2/service/ecs v1.46.3
	github.com/aws/aws-sdk-go-v2/service/efs v1.32.3
	github.com/aws/aws-sdk-go-v2/service/eks v1.49.3
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.41.3
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.27.3
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.27.3
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.38.3
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.31.3
	github.com/aws/aws-sdk-go-v2/service/elastictranscoder v1.26.3
	github.com/aws/aws-sdk-go-v2/service/emr v1.44.3
	github.com/aws/aws-sdk-go-v2/service/emrcontainers v1.32.3
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.25.2
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.34.3
	github.com/aws/aws-sdk-go-v2/service/evidently v1.22.3
	github.com/aws/aws-sdk-go-v2/service/finspace v1.27.3
	github.com/aws/aws-sdk-go-v2/service/firehose v1.33.4
	github.com/aws/aws-sdk-go-v2/service/fis v1.29.3
	github.com/aws/aws-sdk-go-v2/service/fms v1.36.3
	github.com/aws/aws-sdk-go-v2/service/fsx v1.48.3
	github.com/aws/aws-sdk-go-v2/service/gamelift v1.35.3
	github.com/aws/aws-sdk-go-v2/service/glacier v1.25.3
	github.com/aws/aws-sdk-go-v2/service/globalaccelerator v1.28.3
	github.com/aws/aws-sdk-go-v2/service/glue v1.99.2
	github.com/aws/aws-sdk-go-v2/service/grafana v1.25.3
	github.com/aws/aws-sdk-go-v2/service/greengrass v1.26.3
	github.com/aws/aws-sdk-go-v2/service/groundstation v1.30.3
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.48.3
	github.com/aws/aws-sdk-go-v2/service/healthlake v1.27.4
	github.com/aws/aws-sdk-go-v2/service/iam v1.36.3
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.26.3
	github.com/aws/aws-sdk-go-v2/service/inspector v1.24.3
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.31.3
	github.com/aws/aws-sdk-go-v2/service/internetmonitor v1.18.3
	github.com/aws/aws-sdk-go-v2/service/iot v1.57.3
	github.com/aws/aws-sdk-go-v2/service/iotanalytics v1.25.3
	github.com/aws/aws-sdk-go-v2/service/iotevents v1.26.3
	github.com/aws/aws-sdk-go-v2/service/ivs v1.39.3
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.15.3
	github.com/aws/aws-sdk-go-v2/service/kafka v1.37.3
	github.com/aws/aws-sdk-go-v2/service/kafkaconnect v1.20.3
	github.com/aws/aws-sdk-go-v2/service/kendra v1.53.3
	github.com/aws/aws-sdk-go-v2/service/keyspaces v1.13.3
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.31.2
	github.com/aws/aws-sdk-go-v2/service/kinesisanalytics v1.24.3
	github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2 v1.30.3
	github.com/aws/aws-sdk-go-v2/service/kinesisvideo v1.26.3
	github.com/aws/aws-sdk-go-v2/service/kms v1.36.3
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.36.3
	github.com/aws/aws-sdk-go-v2/service/lambda v1.62.1
	github.com/aws/aws-sdk-go-v2/service/launchwizard v1.7.3
	github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice v1.27.3
	github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 v1.48.3
	github.com/aws/aws-sdk-go-v2/service/licensemanager v1.28.3
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.41.3
	github.com/aws/aws-sdk-go-v2/service/location v1.41.4
	github.com/aws/aws-sdk-go-v2/service/lookoutmetrics v1.30.3
	github.com/aws/aws-sdk-go-v2/service/m2 v1.16.3
	github.com/aws/aws-sdk-go-v2/service/macie2 v1.42.3
	github.com/aws/aws-sdk-go-v2/service/mediaconnect v1.34.3
	github.com/aws/aws-sdk-go-v2/service/mediaconvert v1.60.3
	github.com/aws/aws-sdk-go-v2/service/medialive v1.61.3
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.33.3
	github.com/aws/aws-sdk-go-v2/service/mediapackagev2 v1.16.3
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.23.3
	github.com/aws/aws-sdk-go-v2/service/memorydb v1.22.4
	github.com/aws/aws-sdk-go-v2/service/mq v1.26.3
	github.com/aws/aws-sdk-go-v2/service/mwaa v1.30.3
	github.com/aws/aws-sdk-go-v2/service/neptune v1.34.3
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.12.3
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.42.3
	github.com/aws/aws-sdk-go-v2/service/networkmanager v1.30.3
	github.com/aws/aws-sdk-go-v2/service/networkmonitor v1.6.3
	github.com/aws/aws-sdk-go-v2/service/oam v1.14.3
	github.com/aws/aws-sdk-go-v2/service/opensearch v1.40.3
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.15.4
	github.com/aws/aws-sdk-go-v2/service/opsworks v1.25.3
	github.com/aws/aws-sdk-go-v2/service/organizations v1.33.1
	github.com/aws/aws-sdk-go-v2/service/osis v1.13.3
	github.com/aws/aws-sdk-go-v2/service/outposts v1.43.3
	github.com/aws/aws-sdk-go-v2/service/paymentcryptography v1.13.3
	github.com/aws/aws-sdk-go-v2/service/pcaconnectorad v1.8.3
	github.com/aws/aws-sdk-go-v2/service/pcs v1.1.4
	github.com/aws/aws-sdk-go-v2/service/pinpoint v1.33.3
	github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2 v1.14.2
	github.com/aws/aws-sdk-go-v2/service/pipes v1.16.3
	github.com/aws/aws-sdk-go-v2/service/polly v1.44.3
	github.com/aws/aws-sdk-go-v2/service/pricing v1.31.4
	github.com/aws/aws-sdk-go-v2/service/qbusiness v1.12.3
	github.com/aws/aws-sdk-go-v2/service/qldb v1.24.3
	github.com/aws/aws-sdk-go-v2/service/quicksight v1.74.0
	github.com/aws/aws-sdk-go-v2/service/ram v1.28.3
	github.com/aws/aws-sdk-go-v2/service/rbin v1.19.3
	github.com/aws/aws-sdk-go-v2/service/rds v1.86.0
	github.com/aws/aws-sdk-go-v2/service/redshift v1.47.3
	github.com/aws/aws-sdk-go-v2/service/redshiftdata v1.29.3
	github.com/aws/aws-sdk-go-v2/service/redshiftserverless v1.22.3
	github.com/aws/aws-sdk-go-v2/service/rekognition v1.44.3
	github.com/aws/aws-sdk-go-v2/service/resiliencehub v1.25.3
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.14.2
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.26.0
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.24.3
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.15.3
	github.com/aws/aws-sdk-go-v2/service/route53 v1.44.3
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.26.3
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.3.3
	github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig v1.24.3
	github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness v1.20.3
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.31.3
	github.com/aws/aws-sdk-go-v2/service/rum v1.20.3
	github.com/aws/aws-sdk-go-v2/service/s3 v1.64.0
	github.com/aws/aws-sdk-go-v2/service/s3control v1.48.3
	github.com/aws/aws-sdk-go-v2/service/s3outposts v1.27.3
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.162.0
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.11.3
	github.com/aws/aws-sdk-go-v2/service/schemas v1.27.3
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.33.3
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.53.3
	github.com/aws/aws-sdk-go-v2/service/securitylake v1.17.3
	github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository v1.23.3
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.31.3
	github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry v1.29.3
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.32.3
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.24.3
	github.com/aws/aws-sdk-go-v2/service/ses v1.27.3
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.35.0
	github.com/aws/aws-sdk-go-v2/service/sfn v1.32.3
	github.com/aws/aws-sdk-go-v2/service/shield v1.28.3
	github.com/aws/aws-sdk-go-v2/service/signer v1.25.3
	github.com/aws/aws-sdk-go-v2/service/sns v1.32.3
	github.com/aws/aws-sdk-go-v2/service/sqs v1.35.3
	github.com/aws/aws-sdk-go-v2/service/ssm v1.54.3
	github.com/aws/aws-sdk-go-v2/service/ssmcontacts v1.25.3
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.33.3
	github.com/aws/aws-sdk-go-v2/service/ssmsap v1.17.3
	github.com/aws/aws-sdk-go-v2/service/sso v1.23.3
	github.com/aws/aws-sdk-go-v2/service/ssoadmin v1.28.3
	github.com/aws/aws-sdk-go-v2/service/storagegateway v1.33.3
	github.com/aws/aws-sdk-go-v2/service/sts v1.31.3
	github.com/aws/aws-sdk-go-v2/service/swf v1.26.3
	github.com/aws/aws-sdk-go-v2/service/synthetics v1.28.3
	github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb v1.5.0
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.28.3
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.40.3
	github.com/aws/aws-sdk-go-v2/service/transfer v1.51.3
	github.com/aws/aws-sdk-go-v2/service/verifiedpermissions v1.18.4
	github.com/aws/aws-sdk-go-v2/service/vpclattice v1.11.3
	github.com/aws/aws-sdk-go-v2/service/waf v1.24.3
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.24.3
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.53.3
	github.com/aws/aws-sdk-go-v2/service/wellarchitected v1.33.3
	github.com/aws/aws-sdk-go-v2/service/worklink v1.23.2
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.47.4
	github.com/aws/aws-sdk-go-v2/service/workspacesweb v1.23.3
	github.com/aws/aws-sdk-go-v2/service/xray v1.28.3
	github.com/aws/smithy-go v1.21.0
	github.com/beevik/etree v1.4.1
	github.com/cedar-policy/cedar-go v0.1.0
	github.com/davecgh/go-spew v1.1.1
	github.com/dlclark/regexp2 v1.11.4
	github.com/gertd/go-pluralize v0.2.1
	github.com/google/go-cmp v0.6.0
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.23.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.58
	github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2 v2.0.0-beta.59
	github.com/hashicorp/awspolicyequivalence v1.6.0
	github.com/hashicorp/cli v1.1.6
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200723130312-85980079f637
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.7.0
	github.com/hashicorp/hcl/v2 v2.22.0
	github.com/hashicorp/terraform-json v0.22.1
	github.com/hashicorp/terraform-plugin-framework v1.12.0
	github.com/hashicorp/terraform-plugin-framework-jsontypes v0.2.0
	github.com/hashicorp/terraform-plugin-framework-timeouts v0.4.1
	github.com/hashicorp/terraform-plugin-framework-timetypes v0.5.0
	github.com/hashicorp/terraform-plugin-framework-validators v0.13.0
	github.com/hashicorp/terraform-plugin-go v0.24.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-mux v0.16.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.34.0
	github.com/hashicorp/terraform-plugin-testing v1.10.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20240118010651-0ba75a80ca38
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pquerna/otp v1.4.0
	github.com/shopspring/decimal v1.4.0
	golang.org/x/crypto v0.27.0
	golang.org/x/mod v0.21.0
	golang.org/x/text v0.18.0
	golang.org/x/tools v0.25.0
	gopkg.in/dnaeon/go-vcr.v3 v3.2.1
	gopkg.in/yaml.v2 v2.4.0
	syreclabs.com/go/faker v1.2.3
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.0 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.27.3 // indirect
	github.com/bgentry/speakeasy v0.2.0 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/bufbuild/protocompile v0.6.0 // indirect
	github.com/cloudflare/circl v1.4.0 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-test/deep v1.1.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-plugin v1.6.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/hc-install v0.8.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.21.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.3 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/posener/complete v1.2.3 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.15.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.55.0 // indirect
	go.opentelemetry.io/otel v1.30.0 // indirect
	go.opentelemetry.io/otel/metric v1.30.0 // indirect
	go.opentelemetry.io/otel/trace v1.30.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240827150818-7e3bb234dfed // indirect
	google.golang.org/grpc v1.66.2 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/hashicorp/terraform-plugin-log => github.com/gdavison/terraform-plugin-log v0.0.0-20230928191232-6c653d8ef8fb
