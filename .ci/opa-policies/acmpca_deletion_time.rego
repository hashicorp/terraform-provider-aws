# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

package tflint

import rego.v1

deny_acmpca_deletion_time contains issue if {
	resources := terraform.resources("aws_acmpca_certificate_authority", {"permanent_deletion_time_in_days": "number"}, {})
	r := resources[_]
	attr := r.config.permanent_deletion_time_in_days
	attr.value > 7

	issue := tflint.issue("permanent_deletion_time_in_days should be 7 in acceptance tests", attr.range)
}

deny_acmpca_deletion_time contains issue if {
	resources := terraform.resources("aws_acmpca_certificate_authority", {"permanent_deletion_time_in_days": "number"}, {})
	r := resources[_]
	not r.config.permanent_deletion_time_in_days

	issue := tflint.issue("permanent_deletion_time_in_days should be explicitly set to 7 in acceptance tests", r.decl_range)
}
