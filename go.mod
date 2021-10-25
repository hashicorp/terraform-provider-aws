module github.com/hashicorp/terraform-provider-aws

go 1.17

require (
	github.com/aws/aws-sdk-go v1.41.9
	github.com/beevik/etree v1.1.0
	github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go v0.14.0
	github.com/hashicorp/aws-sdk-go-base v1.0.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.8.0
	github.com/jen20/awspolicyequivalence v1.1.0
	github.com/keybase/go-crypto v0.0.0-20161004153544-93f5b35093ba
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1
	github.com/pquerna/otp v1.3.0
	github.com/shopspring/decimal v1.3.1
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go v0.81.0 // indirect
	cloud.google.com/go/storage v1.10.0 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/apparentlymart/go-cidr v1.0.1 // indirect
	github.com/apparentlymart/go-textseg v1.0.0 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fatih/color v1.12.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-getter v1.5.3 // indirect
	github.com/hashicorp/go-hclog v0.15.0 // indirect
	github.com/hashicorp/go-plugin v1.4.1 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/hcl/v2 v2.3.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.14.0 // indirect
	github.com/hashicorp/terraform-json v0.12.0 // indirect
	github.com/hashicorp/terraform-plugin-go v0.3.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/jhump/protoreflect v1.6.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/klauspost/compress v1.11.2 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/ulikunitz/xz v0.5.8 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.8.4 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/tools v0.1.5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.44.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/grpc v1.38.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/hashicorp/terraform-plugin-sdk/v2 => github.com/gdavison/terraform-plugin-sdk/v2 v2.7.1-0.20210913224932-c7c2dbd9e010
