// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -TagType=ResourceTag -GetTag -ListTags -ListTagsOutTagsElem=ResourceTags -ServiceTagsSlice -TagOp=TagResource -TagInTagsElem=ResourceTags -UpdateTags -UntagInTagsElem=ResourceTagKeys -CreateTags -ParentNotFoundErrCode=InvalidParameterException "-ParentNotFoundErrMsg=The specified cluster is inactive. Specify an active cluster and try again."
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package billing
