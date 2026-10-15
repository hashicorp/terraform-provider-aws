// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

// ruleid: aws-in-func-name
func AWSThingCreate() {}

// ruleid: aws-in-func-name
func awsThingRead() {}

// ruleid: aws-in-func-name
func CreateAwsWidget() {}

// ok: aws-in-func-name
func ThingCreate() {}

// ok: aws-in-func-name
func ReadWidget() {}

type resourceData struct{}

// ok: aws-in-func-name
func (d *resourceData) GetRawState() cty.Value {
	return cty.Value{}
}
