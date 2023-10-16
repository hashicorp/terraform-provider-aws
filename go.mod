module github.com/hashicorp/terraform-provider-aws

go 1.20

require (
	github.com/ProtonMail/go-crypto v0.0.0-20230717121422-5aa5874ade95
	github.com/YakDriver/regexache v0.23.0
	github.com/aws/aws-sdk-go v1.45.25
	github.com/aws/aws-sdk-go-v2 v1.22.0-zeta.3351ef76d077
	github.com/aws/aws-sdk-go-v2/config v1.18.45
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.13
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.90
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.21.2
	github.com/aws/aws-sdk-go-v2/service/account v1.11.7
	github.com/aws/aws-sdk-go-v2/service/acm v1.19.2
	github.com/aws/aws-sdk-go-v2/service/appconfig v1.21.2
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.27.0
	github.com/aws/aws-sdk-go-v2/service/cleanrooms v1.4.2
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.12.7
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.24.2
	github.com/aws/aws-sdk-go-v2/service/codecatalyst v1.5.7
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.15.8
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.16.2
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.25.7
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.27.2
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.18.7
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.3.2
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.125.0
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.11.2
	github.com/aws/aws-sdk-go-v2/service/finspace v1.12.2
	github.com/aws/aws-sdk-go-v2/service/fis v1.16.2
	github.com/aws/aws-sdk-go-v2/service/glacier v1.16.2
	github.com/aws/aws-sdk-go-v2/service/healthlake v1.17.7
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.18.4
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.17.0
	github.com/aws/aws-sdk-go-v2/service/internetmonitor v1.6.2
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.6.2
	github.com/aws/aws-sdk-go-v2/service/kafka v1.22.8
	github.com/aws/aws-sdk-go-v2/service/kendra v1.43.2
	github.com/aws/aws-sdk-go-v2/service/keyspaces v1.4.7
	github.com/aws/aws-sdk-go-v2/service/lambda v1.40.0
	github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 v1.32.7
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.28.7
	github.com/aws/aws-sdk-go-v2/service/mediaconnect v1.20.7
	github.com/aws/aws-sdk-go-v2/service/medialive v1.37.2
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.23.5
	github.com/aws/aws-sdk-go-v2/service/oam v1.4.2
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.5.2
	github.com/aws/aws-sdk-go-v2/service/pipes v1.4.3
	github.com/aws/aws-sdk-go-v2/service/pricing v1.21.9
	github.com/aws/aws-sdk-go-v2/service/qldb v1.16.7
	github.com/aws/aws-sdk-go-v2/service/rbin v1.10.2
	github.com/aws/aws-sdk-go-v2/service/rds v1.56.0
	github.com/aws/aws-sdk-go-v2/service/redshiftdata v1.20.7
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.4.3
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.16.2
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.16.2
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.3.8
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.17.5
	github.com/aws/aws-sdk-go-v2/service/s3 v1.41.0-zeta.3351ef76d077
	github.com/aws/aws-sdk-go-v2/service/s3control v1.33.2
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.3.2
	github.com/aws/aws-sdk-go-v2/service/securitylake v1.7.2
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.16.2
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.20.3
	github.com/aws/aws-sdk-go-v2/service/signer v1.16.7
	github.com/aws/aws-sdk-go-v2/service/ssm v1.38.2
	github.com/aws/aws-sdk-go-v2/service/ssmcontacts v1.17.2
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.23.2
	github.com/aws/aws-sdk-go-v2/service/swf v1.17.5
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.19.2
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.29.0
	github.com/aws/aws-sdk-go-v2/service/verifiedpermissions v1.2.3
	github.com/aws/aws-sdk-go-v2/service/vpclattice v1.2.2
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.31.2
	github.com/aws/aws-sdk-go-v2/service/xray v1.18.2
	github.com/beevik/etree v1.2.0
	github.com/google/go-cmp v0.6.0
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.21.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.37
	github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2 v2.0.0-beta.38
	github.com/hashicorp/awspolicyequivalence v1.6.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/hcl/v2 v2.18.0
	github.com/hashicorp/terraform-plugin-framework v1.4.1
	github.com/hashicorp/terraform-plugin-framework-timeouts v0.4.1
	github.com/hashicorp/terraform-plugin-framework-validators v0.12.0
	github.com/hashicorp/terraform-plugin-go v0.19.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-mux v0.12.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.29.0
	github.com/hashicorp/terraform-plugin-testing v1.5.1
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20230413205102-771768614e91
	github.com/mitchellh/cli v1.1.5
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pquerna/otp v1.4.0
	github.com/shopspring/decimal v1.3.1
	golang.org/x/crypto v0.14.0
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	golang.org/x/tools v0.14.0
	gopkg.in/dnaeon/go-vcr.v3 v3.1.2
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
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.43 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.45 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.1.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.22.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/iam v1.22.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.0-zeta.3351ef76d077 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.37 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.15.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.24.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.23.2 // indirect
	github.com/aws/smithy-go v1.15.0 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/bufbuild/protocompile v0.6.0 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/frankban/quicktest v1.14.6 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-test/deep v1.1.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-plugin v1.5.1 // indirect
	github.com/hashicorp/hc-install v0.6.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.19.0 // indirect
	github.com/hashicorp/terraform-json v0.17.1 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.2 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/posener/complete v1.2.3 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.14.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.45.0 // indirect
	go.opentelemetry.io/otel v1.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.19.0 // indirect
	go.opentelemetry.io/otel/trace v1.19.0 // indirect
	golang.org/x/mod v0.13.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230807174057-1744710a1577 // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/aws/aws-sdk-go-v2 => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2

replace github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream

replace github.com/aws/aws-sdk-go-v2/config => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/config

replace github.com/aws/aws-sdk-go-v2/credentials => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/credentials

replace github.com/aws/aws-sdk-go-v2/feature/ec2/imds => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/feature/ec2/imds

replace github.com/aws/aws-sdk-go-v2/feature/s3/manager => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/feature/s3/manager

replace github.com/aws/aws-sdk-go-v2/internal/configsources => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/internal/configsources

replace github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/internal/endpoints/v2

replace github.com/aws/aws-sdk-go-v2/internal/ini => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/internal/ini

replace github.com/aws/aws-sdk-go-v2/internal/v4a => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/internal/v4a

replace github.com/aws/aws-sdk-go-v2/service/accessanalyzer => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/accessanalyzer

replace github.com/aws/aws-sdk-go-v2/service/account => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/account

replace github.com/aws/aws-sdk-go-v2/service/acm => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/acm

replace github.com/aws/aws-sdk-go-v2/service/appconfig => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/appconfig

replace github.com/aws/aws-sdk-go-v2/service/auditmanager => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/auditmanager

replace github.com/aws/aws-sdk-go-v2/service/cleanrooms => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/cleanrooms

replace github.com/aws/aws-sdk-go-v2/service/cloudcontrol => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/cloudcontrol

replace github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs

replace github.com/aws/aws-sdk-go-v2/service/codecatalyst => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/codecatalyst

replace github.com/aws/aws-sdk-go-v2/service/codestarconnections => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/codestarconnections

replace github.com/aws/aws-sdk-go-v2/service/codestarnotifications => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/codestarnotifications

replace github.com/aws/aws-sdk-go-v2/service/comprehend => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/comprehend

replace github.com/aws/aws-sdk-go-v2/service/computeoptimizer => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/computeoptimizer

replace github.com/aws/aws-sdk-go-v2/service/directoryservice => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/directoryservice

replace github.com/aws/aws-sdk-go-v2/service/docdbelastic => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/docdbelastic

replace github.com/aws/aws-sdk-go-v2/service/dynamodb => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/dynamodb

replace github.com/aws/aws-sdk-go-v2/service/ec2 => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/ec2

replace github.com/aws/aws-sdk-go-v2/service/emrserverless => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/emrserverless

replace github.com/aws/aws-sdk-go-v2/service/finspace => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/finspace

replace github.com/aws/aws-sdk-go-v2/service/fis => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/fis

replace github.com/aws/aws-sdk-go-v2/service/glacier => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/glacier

replace github.com/aws/aws-sdk-go-v2/service/healthlake => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/healthlake

replace github.com/aws/aws-sdk-go-v2/service/iam => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/iam

replace github.com/aws/aws-sdk-go-v2/service/identitystore => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/identitystore

replace github.com/aws/aws-sdk-go-v2/service/inspector2 => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/inspector2

replace github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding

replace github.com/aws/aws-sdk-go-v2/service/internal/checksum => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/internal/checksum

replace github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery

replace github.com/aws/aws-sdk-go-v2/service/internal/presigned-url => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/internal/presigned-url

replace github.com/aws/aws-sdk-go-v2/service/internal/s3shared => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/internal/s3shared

replace github.com/aws/aws-sdk-go-v2/service/internetmonitor => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/internetmonitor

replace github.com/aws/aws-sdk-go-v2/service/ivschat => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/ivschat

replace github.com/aws/aws-sdk-go-v2/service/kafka => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/kafka

replace github.com/aws/aws-sdk-go-v2/service/kendra => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/kendra

replace github.com/aws/aws-sdk-go-v2/service/keyspaces => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/keyspaces

replace github.com/aws/aws-sdk-go-v2/service/lambda => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/lambda

replace github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/lexmodelsv2

replace github.com/aws/aws-sdk-go-v2/service/lightsail => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/lightsail

replace github.com/aws/aws-sdk-go-v2/service/mediaconnect => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/mediaconnect

replace github.com/aws/aws-sdk-go-v2/service/medialive => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/medialive

replace github.com/aws/aws-sdk-go-v2/service/mediapackage => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/mediapackage

replace github.com/aws/aws-sdk-go-v2/service/oam => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/oam

replace github.com/aws/aws-sdk-go-v2/service/opensearchserverless => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/opensearchserverless

replace github.com/aws/aws-sdk-go-v2/service/pipes => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/pipes

replace github.com/aws/aws-sdk-go-v2/service/pricing => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/pricing

replace github.com/aws/aws-sdk-go-v2/service/qldb => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/qldb

replace github.com/aws/aws-sdk-go-v2/service/rbin => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/rbin

replace github.com/aws/aws-sdk-go-v2/service/rds => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/rds

replace github.com/aws/aws-sdk-go-v2/service/redshiftdata => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/redshiftdata

replace github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/resourceexplorer2

replace github.com/aws/aws-sdk-go-v2/service/resourcegroups => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/resourcegroups

replace github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi

replace github.com/aws/aws-sdk-go-v2/service/rolesanywhere => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/rolesanywhere

replace github.com/aws/aws-sdk-go-v2/service/route53domains => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/route53domains

replace github.com/aws/aws-sdk-go-v2/service/s3 => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/s3

replace github.com/aws/aws-sdk-go-v2/service/s3control => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/s3control

replace github.com/aws/aws-sdk-go-v2/service/scheduler => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/scheduler

replace github.com/aws/aws-sdk-go-v2/service/securitylake => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/securitylake

replace github.com/aws/aws-sdk-go-v2/service/servicequotas => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/servicequotas

replace github.com/aws/aws-sdk-go-v2/service/sesv2 => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/sesv2

replace github.com/aws/aws-sdk-go-v2/service/signer => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/signer

replace github.com/aws/aws-sdk-go-v2/service/sqs => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/sqs

replace github.com/aws/aws-sdk-go-v2/service/ssm => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/ssm

replace github.com/aws/aws-sdk-go-v2/service/ssmcontacts => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/ssmcontacts

replace github.com/aws/aws-sdk-go-v2/service/ssmincidents => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/ssmincidents

replace github.com/aws/aws-sdk-go-v2/service/sso => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/sso

replace github.com/aws/aws-sdk-go-v2/service/ssooidc => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/ssooidc

replace github.com/aws/aws-sdk-go-v2/service/sts => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/sts

replace github.com/aws/aws-sdk-go-v2/service/swf => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/swf

replace github.com/aws/aws-sdk-go-v2/service/timestreamwrite => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/timestreamwrite

replace github.com/aws/aws-sdk-go-v2/service/transcribe => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/transcribe

replace github.com/aws/aws-sdk-go-v2/service/verifiedpermissions => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/verifiedpermissions

replace github.com/aws/aws-sdk-go-v2/service/vpclattice => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/vpclattice

replace github.com/aws/aws-sdk-go-v2/service/workspaces => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/workspaces

replace github.com/aws/aws-sdk-go-v2/service/xray => /Users/ewbankkit/src/github.com/aws/aws-sdk-go-v2/service/xray
