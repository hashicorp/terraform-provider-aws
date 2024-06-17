module github.com/hashicorp/terraform-provider-aws/skaff

go 1.22.2

require (
	github.com/YakDriver/regexache v0.23.0
	github.com/hashicorp/terraform-provider-aws v1.60.1-0.20220322001452-8f7a597d0c24
	github.com/spf13/cobra v1.8.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/hashicorp/terraform-provider-aws => ../
