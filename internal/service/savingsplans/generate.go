// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -KVTValues -ListTags -ListTagsInIDElem=ResourceArn -ServiceTagsMap -TagInIDElem=ResourceArn -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY://go:generate go run ../../generate/tagstests/main.go

package savingsplans
