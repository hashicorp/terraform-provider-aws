// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tags/main.go -ListTags -ServiceTagsSlice -TagType=ResourceTag -UntagInTagsElem=ResourceTagKeys -UpdateTags -ListTagsOutTagsElem=ResourceTags -TagInTagsElem=ResourceTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package invoicing
