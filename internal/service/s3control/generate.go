// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ServiceTagsSlice -TagResTypeElem=AccountId -UpdateTags
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ServiceTagsSlice -TagsFunc=tagsS3 -KeyValueTagsFunc=keyValueTagsS3 -GetTagsInFunc=getTagsInS3 -SetTagsOutFunc=setTagsOutS3 -SkipAWSServiceImp -TagType=S3Tag -- tagss3_gen.go
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package s3control
