// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tagresource/main.go -UpdateTagsFunc=updateTagsNoIgnoreSystem
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsInIDElem=Arn -ServiceTagsSlice -TagInIDElem=Arn -UpdateTags
//go:generate go run ../../generate/tags/main.go -TagInIDElem=Arn -UpdateTags -UpdateTagsFunc=updateTagsNoIgnoreSystem -UpdateTagsNoIgnoreSystem -SkipNamesImp -- update_tags_no_system_ignore_gen.go
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package transfer
