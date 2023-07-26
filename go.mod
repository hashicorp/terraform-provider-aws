module github.com/hashicorp/terraform-provider-aws

go 1.20

require (
	github.com/ProtonMail/go-crypto v0.0.0-20230217124315-7d5c6f04bbb8
	github.com/aws/aws-sdk-go v1.44.305
	github.com/aws/aws-sdk-go-v2 v1.19.0
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.5
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.19.15
	github.com/aws/aws-sdk-go-v2/service/account v1.10.9
	github.com/aws/aws-sdk-go-v2/service/acm v1.17.14
	github.com/aws/aws-sdk-go-v2/service/appconfig v1.17.12
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.25.1
	github.com/aws/aws-sdk-go-v2/service/cleanrooms v1.2.1
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.11.15
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.22.1
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.24.5
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.24.3
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.17.4
	github.com/aws/aws-sdk-go-v2/service/docdbelastic v1.1.13
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.105.1
	github.com/aws/aws-sdk-go-v2/service/emrserverless v1.9.0
	github.com/aws/aws-sdk-go-v2/service/finspace v1.10.3
	github.com/aws/aws-sdk-go-v2/service/fis v1.14.13
	github.com/aws/aws-sdk-go-v2/service/glacier v1.14.14
	github.com/aws/aws-sdk-go-v2/service/healthlake v1.16.3
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.16.14
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.15.1
	github.com/aws/aws-sdk-go-v2/service/internetmonitor v1.3.1
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.4.8
	github.com/aws/aws-sdk-go-v2/service/kendra v1.41.1
	github.com/aws/aws-sdk-go-v2/service/keyspaces v1.3.3
	github.com/aws/aws-sdk-go-v2/service/lambda v1.37.1
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.27.2
	github.com/aws/aws-sdk-go-v2/service/medialive v1.32.1
	github.com/aws/aws-sdk-go-v2/service/oam v1.1.14
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.2.7
	github.com/aws/aws-sdk-go-v2/service/pipes v1.2.9
	github.com/aws/aws-sdk-go-v2/service/pricing v1.20.1
	github.com/aws/aws-sdk-go-v2/service/qldb v1.15.14
	github.com/aws/aws-sdk-go-v2/service/rbin v1.8.15
	github.com/aws/aws-sdk-go-v2/service/rds v1.46.2
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.2.16
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.2.3
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.15.1
	github.com/aws/aws-sdk-go-v2/service/s3control v1.31.9
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.1.14
	github.com/aws/aws-sdk-go-v2/service/securitylake v1.5.0
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.18.3
	github.com/aws/aws-sdk-go-v2/service/signer v1.15.3
	github.com/aws/aws-sdk-go-v2/service/ssm v1.36.8
	github.com/aws/aws-sdk-go-v2/service/ssmcontacts v1.15.8
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.21.7
	github.com/aws/aws-sdk-go-v2/service/swf v1.15.3
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.17.3
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.27.0
	github.com/aws/aws-sdk-go-v2/service/verifiedpermissions v1.0.5
	github.com/aws/aws-sdk-go-v2/service/vpclattice v1.0.8
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.28.16
	github.com/aws/aws-sdk-go-v2/service/xray v1.16.14
	github.com/beevik/etree v1.2.0
	github.com/google/go-cmp v0.5.9
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.21.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.32
	github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2 v2.0.0-beta.33
	github.com/hashicorp/awspolicyequivalence v1.6.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/hcl/v2 v2.17.0
	github.com/hashicorp/terraform-plugin-framework v1.3.3
	github.com/hashicorp/terraform-plugin-framework-timeouts v0.4.1
	github.com/hashicorp/terraform-plugin-framework-validators v0.10.0
	github.com/hashicorp/terraform-plugin-go v0.18.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-mux v0.11.2
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.27.0
	github.com/hashicorp/terraform-plugin-testing v1.3.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24
	github.com/mitchellh/cli v1.1.5
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pquerna/otp v1.4.0
	github.com/shopspring/decimal v1.3.1
	golang.org/x/crypto v0.11.0
	golang.org/x/exp v0.0.0-20230510235704-dd950f8aeaea
	golang.org/x/tools v0.6.0
	gopkg.in/dnaeon/go-vcr.v3 v3.1.2
	gopkg.in/yaml.v2 v2.4.0
	syreclabs.com/go/faker v1.2.3
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.1 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/armon/go-radix v0.0.0-20180808171621-7fddfc383310 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.18.28 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/service/iam v1.21.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.19.3 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-plugin v1.4.10 // indirect
	github.com/hashicorp/hc-install v0.5.2 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.18.1 // indirect
	github.com/hashicorp/terraform-json v0.17.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.1 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/posener/complete v1.1.1 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.13.2 // indirect
	go.opentelemetry.io/otel v1.16.0 // indirect
	go.opentelemetry.io/otel/trace v1.16.0 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/net v0.11.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/grpc v1.56.1 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
