// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeListenerCertificates -InputPaginator=Marker -OutputPaginator=NextMarker -- list_listener_certificates_pages_gen.go
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceArns -ListTagsInIDNeedValueSlice -ListTagsOutTagsElem=TagDescriptions[0].Tags -ServiceTagsSlice -TagOp=AddTags -TagInIDElem=ResourceArns -TagInIDNeedValueSlice -UntagOp=RemoveTags -UpdateTags -CreateTags -KVTValues
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elbv2
