// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=Arn -ListTagsOutTagsElem=ResourceTags.Tags -ServiceTagsMap -TagInIDElem=Arn -UpdateTags -KVTValues
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package mediaconvert
