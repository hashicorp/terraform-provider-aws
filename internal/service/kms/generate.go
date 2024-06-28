// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsOp=ListResourceTags -ListTagsOpPaginated -ListTagsInIDElem=KeyId -ServiceTagsSlice -TagInIDElem=KeyId -TagTypeKeyElem=TagKey -TagTypeValElem=TagValue -UpdateTags -Wait -WaitContinuousOccurence 5 -WaitMinTimeout 1s -WaitTimeout 10m -ParentNotFoundErrCode=NotFoundException
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kms
