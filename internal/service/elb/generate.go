// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=DescribeTags -ListTagsInIDElem=LoadBalancerNames -ListTagsInIDNeedValueSlice -ListTagsOutTagsElem=TagDescriptions[0].Tags -ServiceTagsSlice -TagOp=AddTags -TagInIDElem=LoadBalancerNames -TagInIDNeedValueSlice -TagKeyType=TagKeyOnly -UntagOp=RemoveTags -UntagInNeedTagKeyType -UntagInTagsElem=Tags -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elb
