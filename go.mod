module github.com/hashicorp/terraform-provider-aws

go 1.20

require (
	github.com/ProtonMail/go-crypto v0.0.0-20230923063757-afb1ddc0824c
	github.com/YakDriver/regexache v0.23.0
	github.com/aws/aws-sdk-go v1.47.7
	github.com/aws/aws-sdk-go-v2 v1.22.2
	github.com/aws/aws-sdk-go-v2/config v1.20.0
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.3
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.13.2
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.23.1
	github.com/aws/aws-sdk-go-v2/service/account v1.13.1
	github.com/aws/aws-sdk-go-v2/service/acm v1.21.1
	github.com/aws/aws-sdk-go-v2/service/appconfig v1.25.1
	github.com/aws/aws-sdk-go-v2/service/athena v1.34.1
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.29.1
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.3.1
	github.com/aws/aws-sdk-go-v2/service/cleanrooms v1.6.1
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.14.2
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.26.1
	github.com/aws/aws-sdk-go-v2/service/codecatalyst v1.7.1
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.17.1
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.19.1
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.27.1
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.29.1
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.21.1
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.5.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.130.1
	github.com/aws/aws-sdk-go-v2/service/eks v1.32.1
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.13.1
	github.com/aws/aws-sdk-go-v2/service/finspace v1.16.1
	github.com/aws/aws-sdk-go-v2/service/fis v1.19.1
	github.com/aws/aws-sdk-go-v2/service/glacier v1.18.1
	github.com/aws/aws-sdk-go-v2/service/healthlake v1.19.1
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.20.1
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.19.1
	github.com/aws/aws-sdk-go-v2/service/internetmonitor v1.8.1
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.8.1
	github.com/aws/aws-sdk-go-v2/service/kafka v1.26.1
	github.com/aws/aws-sdk-go-v2/service/kendra v1.46.1
	github.com/aws/aws-sdk-go-v2/service/keyspaces v1.6.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.44.1
	github.com/aws/aws-sdk-go-v2/service/lexmodelsv2 v1.36.1
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.31.1
	github.com/aws/aws-sdk-go-v2/service/mediaconnect v1.23.1
	github.com/aws/aws-sdk-go-v2/service/medialive v1.40.1
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.26.1
	github.com/aws/aws-sdk-go-v2/service/oam v1.6.1
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.8.1
	github.com/aws/aws-sdk-go-v2/service/pipes v1.6.1
	github.com/aws/aws-sdk-go-v2/service/pricing v1.23.1
	github.com/aws/aws-sdk-go-v2/service/qldb v1.18.1
	github.com/aws/aws-sdk-go-v2/service/rbin v1.12.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.62.2
	github.com/aws/aws-sdk-go-v2/service/redshiftdata v1.22.1
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.6.2
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.18.1
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.18.1
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.5.1
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.19.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.42.1
	github.com/aws/aws-sdk-go-v2/service/s3control v1.36.1
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.5.1
	github.com/aws/aws-sdk-go-v2/service/securitylake v1.9.1
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.18.1
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.23.1
	github.com/aws/aws-sdk-go-v2/service/signer v1.18.1
	github.com/aws/aws-sdk-go-v2/service/sns v1.25.1
	github.com/aws/aws-sdk-go-v2/service/sqs v1.28.0
	github.com/aws/aws-sdk-go-v2/service/ssm v1.42.1
	github.com/aws/aws-sdk-go-v2/service/ssmcontacts v1.19.1
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.25.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.25.1
	github.com/aws/aws-sdk-go-v2/service/swf v1.19.1
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.22.1
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.31.1
	github.com/aws/aws-sdk-go-v2/service/verifiedpermissions v1.5.1
	github.com/aws/aws-sdk-go-v2/service/vpclattice v1.4.1
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.33.1
	github.com/aws/aws-sdk-go-v2/service/xray v1.22.1
	github.com/beevik/etree v1.2.0
	github.com/davecgh/go-spew v1.1.1
	github.com/gertd/go-pluralize v0.2.1
	github.com/google/go-cmp v0.6.0
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.21.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.38
	github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2 v2.0.0-beta.39
	github.com/hashicorp/awspolicyequivalence v1.6.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200723130312-85980079f637
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/hcl/v2 v2.18.0
	github.com/hashicorp/terraform-plugin-framework v1.4.2
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
	golang.org/x/crypto v0.15.0
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
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.25.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/iam v1.26.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.8.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.16.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.17.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.19.1 // indirect
	github.com/aws/smithy-go v1.16.0 // indirect
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
	github.com/hashicorp/go-plugin v1.5.2 // indirect
	github.com/hashicorp/hc-install v0.6.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.19.0 // indirect
	github.com/hashicorp/terraform-json v0.17.1 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.2 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
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
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230920204549-e6e6cdab5c13 // indirect
	google.golang.org/grpc v1.58.2 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/hashicorp/terraform-plugin-log => github.com/gdavison/terraform-plugin-log v0.0.0-20230928191232-6c653d8ef8fb

exclude ( // Contains INI parsing regression
	github.com/aws/aws-sdk-go-v2/config v1.21.0
	github.com/aws/aws-sdk-go-v2/config v1.22.0
	github.com/aws/aws-sdk-go-v2/config v1.22.1
	github.com/aws/aws-sdk-go-v2/internal/ini v1.5.0
	github.com/aws/aws-sdk-go-v2/internal/ini v1.5.1
)

replace github.com/hashicorp/aws-sdk-go-base/v2 => github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.38.0.20231106231929-ee4c6bb6c5c8
