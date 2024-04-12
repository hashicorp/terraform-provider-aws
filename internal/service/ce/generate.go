// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsOutTagsElem=ResourceTags -ServiceTagsSlice -TagInTagsElem=ResourceTags -UpdateTags -UntagInTagsElem=ResourceTagKeys -UntagInTagsElem=ResourceTagKeys -TagType=ResourceTag
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package ce
