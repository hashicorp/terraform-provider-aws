module github.com/hashicorp/terraform-provider-aws

go 1.19

require (
	github.com/ProtonMail/go-crypto v0.0.0-20230201104953-d1d05f4e2bfb
	github.com/aws/aws-sdk-go v1.44.201
	github.com/aws/aws-sdk-go-v2 v1.17.4
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.22
	github.com/aws/aws-sdk-go-v2/service/auditmanager v1.23.1
	github.com/aws/aws-sdk-go-v2/service/cloudcontrol v1.11.2
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.20.2
	github.com/aws/aws-sdk-go-v2/service/comprehend v1.21.1
	github.com/aws/aws-sdk-go-v2/service/computeoptimizer v1.21.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.86.0
	github.com/aws/aws-sdk-go-v2/service/fis v1.14.1
	github.com/aws/aws-sdk-go-v2/service/identitystore v1.16.1
	github.com/aws/aws-sdk-go-v2/service/inspector2 v1.11.2
	github.com/aws/aws-sdk-go-v2/service/ivschat v1.3.1
	github.com/aws/aws-sdk-go-v2/service/kendra v1.38.2
	github.com/aws/aws-sdk-go-v2/service/medialive v1.29.1
	github.com/aws/aws-sdk-go-v2/service/opensearchserverless v1.1.3
	github.com/aws/aws-sdk-go-v2/service/pipes v1.1.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.40.2
	github.com/aws/aws-sdk-go-v2/service/resourceexplorer2 v1.2.2
	github.com/aws/aws-sdk-go-v2/service/rolesanywhere v1.1.1
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.14.1
	github.com/aws/aws-sdk-go-v2/service/s3control v1.29.2
	github.com/aws/aws-sdk-go-v2/service/scheduler v1.1.1
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.16.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.35.2
	github.com/aws/aws-sdk-go-v2/service/ssmincidents v1.20.1
	github.com/aws/aws-sdk-go-v2/service/transcribe v1.25.1
	github.com/aws/smithy-go v1.13.5
	github.com/beevik/etree v1.1.0
	github.com/google/go-cmp v0.5.9
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.20.0
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.23
	github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2 v2.0.0-beta.24
	github.com/hashicorp/awspolicyequivalence v1.6.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/hcl/v2 v2.15.0
	github.com/hashicorp/terraform-plugin-framework v1.1.1
	github.com/hashicorp/terraform-plugin-framework-timeouts v0.3.1
	github.com/hashicorp/terraform-plugin-framework-validators v0.10.0
	github.com/hashicorp/terraform-plugin-go v0.14.3
	github.com/hashicorp/terraform-plugin-log v0.8.0
	github.com/hashicorp/terraform-plugin-mux v0.9.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.24.1
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24
	github.com/mitchellh/cli v1.1.5
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/pquerna/otp v1.4.0
	github.com/shopspring/decimal v1.3.1
	golang.org/x/crypto v0.6.0
	golang.org/x/exp v0.0.0-20230206171751-46f607a40771
	golang.org/x/tools v0.2.0
	gopkg.in/dnaeon/go-vcr.v3 v3.1.2
	gopkg.in/yaml.v2 v2.4.0
	syreclabs.com/go/faker v1.2.3
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.1 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/armon/go-radix v0.0.0-20180808171621-7fddfc383310 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.18.12 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/iam v1.19.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.3 // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cloudflare/circl v1.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fatih/color v1.14.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-hclog v1.4.0 // indirect
	github.com/hashicorp/go-plugin v1.4.8 // indirect
	github.com/hashicorp/hc-install v0.5.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.17.3 // indirect
	github.com/hashicorp/terraform-json v0.14.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.1.0 // indirect
	github.com/hashicorp/terraform-svchost v0.1.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/posener/complete v1.1.1 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.12.1 // indirect
	go.opentelemetry.io/otel v1.13.0 // indirect
	go.opentelemetry.io/otel/trace v1.13.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.6.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230202175211-008b39050e57 // indirect
	google.golang.org/grpc v1.53.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
