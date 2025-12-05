// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -TagType=ResourceTag -ListTags -UpdateTags -ListTagsOutTagsElem=ResourceTags -UntagInTagsElem=ResourceTagKeys -TagInTagsElem=ResourceTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package billing
