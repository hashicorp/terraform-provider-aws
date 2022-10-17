module github.com/hashicorp/terraform-provider-aws/skaff

go 1.17

require (
	github.com/hashicorp/terraform-provider-aws v1.60.1-0.20220322001452-8f7a597d0c24
	github.com/spf13/cobra v1.6.0
)

require (
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/hashicorp/terraform-provider-aws => ../
