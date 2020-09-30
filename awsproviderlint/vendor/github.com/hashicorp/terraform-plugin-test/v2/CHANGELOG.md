# 2.1.2 (September 15, 2020)

BUG FIXES:

 - Fix plan output to be in a human-friendly format ([#40](https://github.com/hashicorp/terraform-plugin-test/pull/40))

# 2.1.1 (September 9, 2020)

BUG FIXES:

 - Fix propagation of plugin reattach information ([#38](https://github.com/hashicorp/terraform-plugin-test/pull/38))

# 2.1.0 (September 2, 2020)

FEATURES:

  - Added the ability to create destroy plans. ([#37](https://github.com/hashicorp/terraform-plugin-test/pull/37))
  
ENHANCEMENTS:

  - Normalised internal Terraform CLI commands using github.com/hashicorp/terraform-exec module. ([#35](https://github.com/hashicorp/terraform-plugin-test/pull/35))

# 2.0.0 (August 10, 2020)

FEATURES:

 - Simplified API signatures to reflect no longer needing provider name ([#32](https://github.com/hashicorp/terraform-plugin-test/pull/32))
 - Implement SavedPlanStdout which captures a non-json stdout run of `terraform show` of a planfile ([#34](https://github.com/hashicorp/terraform-plugin-test/pull/34))

# 1.4.4 (July 10, 2020)

BUG FIXES:

 - Fix Windows bug in versions of Terraform below 0.13.0-beta2 ([#30](https://github.com/hashicorp/terraform-plugin-test/pull/30))

# 1.4.3 (July 7, 2020)

DEPENDENCIES:

 - `github.com/hashicorp/go-getter@v1.4.0` ([#29](https://github.com/hashicorp/terraform-plugin-test/pull/29))

# 1.4.2 (July 7, 2020)

DEPENDENCIES:

 - `github.com/hashicorp/terraform-exec@v0.1.1` ([#28](https://github.com/hashicorp/terraform-plugin-test/pull/28))

# 1.4.1 (July 7, 2020)

BUG FIXES:

 - Fix auto-install Terraform feature ([#26](https://github.com/hashicorp/terraform-plugin-test/pull/26))
