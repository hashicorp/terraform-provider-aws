// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsOp=ListTagsForDomain -ListTagsInIDElem=DomainName -ListTagsOutTagsElem=TagList -ServiceTagsSlice -UpdateTags -UntagOp=DeleteTagsForDomain -UntagInTagsElem=TagsToDelete -TagOp=UpdateTagsForDomain -TagInTagsElem=TagsToUpdate -TagInIDElem=DomainName
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package route53domains
