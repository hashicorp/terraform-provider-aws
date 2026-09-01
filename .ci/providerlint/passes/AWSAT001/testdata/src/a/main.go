// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package a

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const resourceName = `aws_example_thing.test`

var testRegexp = regexache.MustCompile(`.*`)

func f() {
	/* Passing cases */

	resource.TestMatchResourceAttr(resourceName, "not_matching", testRegexp)

	/* Comment ignored cases */

	//lintignore:AWSAT001
	resource.TestMatchResourceAttr(resourceName, "arn", testRegexp)

	resource.TestMatchResourceAttr(resourceName, "some_arn", testRegexp) //lintignore:AWSAT001

	/* Failing cases */

	resource.TestMatchResourceAttr(resourceName, "arn", testRegexp)                     // want "prefer resource.TestCheckResourceAttrPair\\(\\) or ARN check functions"
	resource.TestMatchResourceAttr(resourceName, "some_arn", testRegexp)                // want "prefer resource.TestCheckResourceAttrPair\\(\\) or ARN check functions"
	resource.TestMatchResourceAttr(resourceName, "config_block.0.arn", testRegexp)      // want "prefer resource.TestCheckResourceAttrPair\\(\\) or ARN check functions"
	resource.TestMatchResourceAttr(resourceName, "config_block.0.some_arn", testRegexp) // want "prefer resource.TestCheckResourceAttrPair\\(\\) or ARN check functions"
}
