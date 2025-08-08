// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -CreateTags -ListTags -ListTagsInIDElem=ResourceName -ListTagsOutTagsElem=TagList -ServiceTagsSlice -TagOp=AddTagsToResource -TagInIDElem=ResourceName -UntagOp=RemoveTagsFromResource -UpdateTags -RetryTagOps -RetryTagsListTagsType=ListTagsForResourceOutput -RetryErrorCode=awstypes.InvalidReplicationGroupStateFault "-RetryErrorMessage=not in available state" -RetryTimeout=15m
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elasticache
