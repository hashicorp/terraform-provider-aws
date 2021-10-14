module github.com/hashicorp/terraform-provider-aws/tools

go 1.15

require (
	github.com/apparentlymart/go-dump v0.0.0-20190214190832-042adf3cf4a0 // indirect
	github.com/bflad/tfproviderdocs v0.9.1
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.42.1
	github.com/hashicorp/go-changelog v0.0.0-20201005170154-56335215ce3a
	github.com/hashicorp/terraform-exec v0.14.0 // indirect
	github.com/katbyte/terrafmt v0.3.0
	github.com/pavius/impi v0.0.3
	github.com/terraform-linters/tflint v0.32.1
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
)

replace github.com/katbyte/terrafmt => github.com/gdavison/terrafmt v0.3.1-0.20210204054728-84242796be99

replace github.com/hashicorp/go-changelog => github.com/breathingdust/go-changelog v0.0.0-20210127001721-f985d5709c15

// v1.5.1 was incorrectly built
exclude github.com/hashicorp/go-getter v1.5.1
