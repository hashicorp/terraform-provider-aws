module github.com/hashicorp/terraform-provider-aws

go 1.22.2

require (
	github.com/ProtonMail/go-crypto v1.1.0-alpha.2
	github.com/YakDriver/go-version v0.1.0
	github.com/YakDriver/regexache v0.23.0
	github.com/aws/aws-sdk-go v1.53.1
	github.com/aws/aws-sdk-go-v2 v1.26.1
	github.com/aws/aws-sdk-go-v2/config v1.27.13
	github.com/aws/aws-sdk-go-v2/credentials v1.17.13
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.1
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.16.17
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.29.3
	github.com/aws/aws-sdk-go-v2/service/account v1.16.5
	github.com/aws/aws-sdk-go-v2/service/acm v1.25.5
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.29.5
	github.com/aws/aws-sdk-go-v2/service/amp v1.25.5
	github.com/aws/aws-sdk-go-v2/service/amplify v1.21.6
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.23.7
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.20.5
	github.com/aws/aws-sdk-go-v2/service/appconfig v1.29.3
	github.com/aws/aws-sdk-go-v2/service/appfabric v1.7.5
	github.com/aws/aws-sdk-go-v2/service/appflow v1.41.5
	github.com/aws/aws-sdk-go-v2/service/appintegrations v1.25.5
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.27.5
	github.com/aws/aws-sdk-go-v2/service/apprunner v1.28.5
	github.com/aws/aws-sdk-go-v2/service/athena v1.40.5
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.32.5
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.40.6
	github.com/aws/aws-sdk-go-v2/service/autoscalingplans v1.20.6
	github.com/aws/aws-sdk-go-v2/service/batch v1.37.1
	github.com/aws/aws-sdk-go-v2/service/bcmdataexports v1.3.5
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.8.2
	github.com/aws/aws-sdk-go-v2/service/bedrockagent v1.10.1
	github.com/aws/aws-sdk-go-v2/service/budgets v1.23.1
	github.com/aws/aws-sdk-go-v2/service/chatbot v1.1.6
	github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines v1.15.6
	github.com/aws/aws-sdk-go-v2/service/chimesdkvoice v1.15.1
	github.com/aws/aws-sdk-go-v2/service/cleanrooms v1.12.1
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.24.5
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.18.5
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.50.1
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.36.1
	github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore v1.4.5
	github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 v1.21.5
	github.com/aws/aws-sdk-go-v2/service/cloudsearch v1.22.5
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.39.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.38.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.35.2
	github.com/aws/aws-sdk-go-v2/service/codeartifact v1.27.1
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.34.2
	github.com/aws/aws-sdk-go-v2/service/codecatalyst v1.13.2
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.22.5
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.25.5
	github.com/aws/aws-sdk-go-v2/service/codeguruprofiler v1.20.5
	github.com/aws/aws-sdk-go-v2/service/codegurureviewer v1.25.5
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.27.1
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.25.3
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.22.5
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.23.8
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.31.5
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.34.2
	github.com/aws/aws-sdk-go-v2/service/configservice v1.46.6
	github.com/aws/aws-sdk-go-v2/service/connectcases v1.17.1
	github.com/aws/aws-sdk-go-v2/service/controltower v1.13.7
	github.com/aws/aws-sdk-go-v2/service/costandusagereportservice v1.23.5
	github.com/aws/aws-sdk-go-v2/service/costexplorer v1.38.1
	github.com/aws/aws-sdk-go-v2/service/costoptimizationhub v1.4.5
	github.com/aws/aws-sdk-go-v2/service/customerprofiles v1.36.5
	github.com/aws/aws-sdk-go-v2/service/datasync v1.37.2
	github.com/aws/aws-sdk-go-v2/service/datazone v1.8.1
	github.com/aws/aws-sdk-go-v2/service/dax v1.19.5
	github.com/aws/aws-sdk-go-v2/service/devopsguru v1.30.5
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.24.5
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.9.4
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.32.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.161.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.28.0
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.23.5
	github.com/aws/aws-sdk-go-v2/service/ecs v1.41.8
	github.com/aws/aws-sdk-go-v2/service/eks v1.42.2
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.38.2
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.23.5
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.30.6
	github.com/aws/aws-sdk-go-v2/service/emr v1.39.6
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.19.2
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.31.0
	github.com/aws/aws-sdk-go-v2/service/evidently v1.19.5
	github.com/aws/aws-sdk-go-v2/service/finspace v1.24.2
	github.com/aws/aws-sdk-go-v2/service/firehose v1.28.7
	github.com/aws/aws-sdk-go-v2/service/fis v1.24.3
	github.com/aws/aws-sdk-go-v2/service/fms v1.33.2
	github.com/aws/aws-sdk-go-v2/service/glacier v1.22.5
	github.com/aws/aws-sdk-go-v2/service/globalaccelerator v1.23.2
	github.com/aws/aws-sdk-go-v2/service/groundstation v1.27.1
	github.com/aws/aws-sdk-go-v2/service/healthlake v1.24.1
	github.com/aws/aws-sdk-go-v2/service/iam v1.32.1
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.23.6
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.26.1
	github.com/aws/aws-sdk-go-v2/service/internetmonitor v1.14.1
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.12.6
	github.com/aws/aws-sdk-go-v2/service/kafka v1.31.4
	github.com/aws/aws-sdk-go-v2/service/kendra v1.50.2
	github.com/aws/aws-sdk-go-v2/service/keyspaces v1.10.5
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.27.5
	github.com/aws/aws-sdk-go-v2/service/kms v1.31.1
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.32.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.54.1
	github.com/aws/aws-sdk-go-v2/service/launchwizard v1.3.5
	github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 v1.43.5
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.37.1
	github.com/aws/aws-sdk-go-v2/service/lookoutmetrics v1.27.5
	github.com/aws/aws-sdk-go-v2/service/m2 v1.13.1
	github.com/aws/aws-sdk-go-v2/service/mediaconnect v1.28.5
	github.com/aws/aws-sdk-go-v2/service/mediaconvert v1.53.2
	github.com/aws/aws-sdk-go-v2/service/medialive v1.52.1
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.30.6
	github.com/aws/aws-sdk-go-v2/service/mediapackagev2 v1.11.1
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.20.5
	github.com/aws/aws-sdk-go-v2/service/mq v1.22.5
	github.com/aws/aws-sdk-go-v2/service/mwaa v1.26.6
	github.com/aws/aws-sdk-go-v2/service/neptunegraph v1.8.2
	github.com/aws/aws-sdk-go-v2/service/oam v1.11.1
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.11.7
	github.com/aws/aws-sdk-go-v2/service/osis v1.8.5
	github.com/aws/aws-sdk-go-v2/service/paymentcryptography v1.10.1
	github.com/aws/aws-sdk-go-v2/service/pcaconnectorad v1.5.5
	github.com/aws/aws-sdk-go-v2/service/pipes v1.11.5
	github.com/aws/aws-sdk-go-v2/service/polly v1.40.0
	github.com/aws/aws-sdk-go-v2/service/pricing v1.28.2
	github.com/aws/aws-sdk-go-v2/service/qbusiness v1.6.1
	github.com/aws/aws-sdk-go-v2/service/qldb v1.21.5
	github.com/aws/aws-sdk-go-v2/service/ram v1.25.5
	github.com/aws/aws-sdk-go-v2/service/rbin v1.16.5
	github.com/aws/aws-sdk-go-v2/service/rds v1.78.1
	github.com/aws/aws-sdk-go-v2/service/redshift v1.44.1
	github.com/aws/aws-sdk-go-v2/service/redshiftdata v1.25.5
	github.com/aws/aws-sdk-go-v2/service/redshiftserverless v1.18.1
	github.com/aws/aws-sdk-go-v2/service/rekognition v1.40.1
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.10.6
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.22.1
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.21.5
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.11.1
	github.com/aws/aws-sdk-go-v2/service/route53 v1.40.5
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.23.5
	github.com/aws/aws-sdk-go-v2/service/route53profiles v1.0.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.53.2
	github.com/aws/aws-sdk-go-v2/service/s3control v1.44.7
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.8.5
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.28.7
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.48.1
	github.com/aws/aws-sdk-go-v2/service/securitylake v1.13.4
	github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry v1.26.5
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.21.5
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.29.1
	github.com/aws/aws-sdk-go-v2/service/shield v1.25.5
	github.com/aws/aws-sdk-go-v2/service/signer v1.22.8
	github.com/aws/aws-sdk-go-v2/service/sns v1.29.5
	github.com/aws/aws-sdk-go-v2/service/sqs v1.32.0
	github.com/aws/aws-sdk-go-v2/service/ssm v1.50.1
	github.com/aws/aws-sdk-go-v2/service/ssmcontacts v1.22.5
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.30.5
	github.com/aws/aws-sdk-go-v2/service/ssmsap v1.13.0
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.6
	github.com/aws/aws-sdk-go-v2/service/ssoadmin v1.25.6
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.7
	github.com/aws/aws-sdk-go-v2/service/swf v1.22.5
	github.com/aws/aws-sdk-go-v2/service/synthetics v1.24.5
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.25.6
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.37.1
	github.com/aws/aws-sdk-go-v2/service/transfer v1.47.3
	github.com/aws/aws-sdk-go-v2/service/verifiedpermissions v1.14.0
	github.com/aws/aws-sdk-go-v2/service/vpclattice v1.8.0
	github.com/aws/aws-sdk-go-v2/service/waf v1.20.5
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.21.5
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.48.4
	github.com/aws/aws-sdk-go-v2/service/wellarchitected v1.30.1
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.39.1
	github.com/aws/aws-sdk-go-v2/service/workspacesweb v1.18.1
	github.com/aws/aws-sdk-go-v2/service/xray v1.25.5
	github.com/aws/smithy-go v1.20.2
	github.com/beevik/etree v1.4.0
	github.com/cedar-policy/cedar-go v0.0.0-20240318205125-470d1fe984bb
	github.com/davecgh/go-spew v1.1.1
	github.com/gertd/go-pluralize v0.2.1
	github.com/google/go-cmp v0.6.0
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.22.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.53
	github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2 v2.0.0-beta.54
	github.com/hashicorp/awspolicyequivalence v1.6.0
	github.com/hashicorp/cli v1.1.6
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200723130312-85980079f637
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/hcl/v2 v2.20.0
	github.com/hashicorp/terraform-plugin-framework v1.8.0
	github.com/hashicorp/terraform-plugin-framework-jsontypes v0.1.0
	github.com/hashicorp/terraform-plugin-framework-timeouts v0.4.1
	github.com/hashicorp/terraform-plugin-framework-timetypes v0.3.0
	github.com/hashicorp/terraform-plugin-framework-validators v0.12.0
	github.com/hashicorp/terraform-plugin-go v0.23.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-mux v0.16.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.33.0
	github.com/hashicorp/terraform-plugin-testing v1.7.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20230413205102-771768614e91
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pquerna/otp v1.4.0
	github.com/shopspring/decimal v1.4.0
	golang.org/x/crypto v0.23.0
	golang.org/x/text v0.15.0
	golang.org/x/tools v0.18.0
	gopkg.in/dnaeon/go-vcr.v3 v3.2.0
	gopkg.in/yaml.v2 v2.4.0
	syreclabs.com/go/faker v1.2.3
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.24.0 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/bufbuild/protocompile v0.6.0 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/frankban/quicktest v1.14.6 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-test/deep v1.1.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-plugin v1.6.0 // indirect
	github.com/hashicorp/hc-install v0.6.3 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.20.0 // indirect
	github.com/hashicorp/terraform-json v0.21.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.3 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/posener/complete v1.2.3 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.14.3 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.51.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/mod v0.15.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.34.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/hashicorp/terraform-plugin-log => github.com/gdavison/terraform-plugin-log v0.0.0-20230928191232-6c653d8ef8fb
