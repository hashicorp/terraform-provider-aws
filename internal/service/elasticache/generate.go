// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceName -ListTagsOutTagsElem=TagList -ServiceTagsSlice -TagOp=AddTagsToResource -TagInIDElem=ResourceName -UntagOp=RemoveTagsFromResource -UpdateTags -CreateTags -RetryTagsListTagsType=TagListMessage -RetryTagsErrorCodes=elasticache.ErrCodeInvalidReplicationGroupStateFault "-RetryTagsErrorMessages=not in available state" -RetryTagsTimeout=15m
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -TagsFunc=TagsV2 -KeyValueTagsFunc=keyValueTagsV2 -GetTagsInFunc=getTagsInV2 -SetTagsOutFunc=setTagsOutV2 -SkipAWSServiceImp -KVTValues -ServiceTagsSlice -- tagsv2_gen.go
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elasticache
