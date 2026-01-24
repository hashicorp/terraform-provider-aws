// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ServiceTagsSlice -TagResTypeIsAccountID -TagResTypeElem=AccountId -UpdateTags
//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -TagsFunc=svcS3Tags -KeyValueTagsFunc=keyValueTagsFromS3Tags -GetTagsInFunc=getS3TagsIn -SetTagsOutFunc=setS3TagsOut -TagType=S3Tag -- s3_tags_gen.go
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/identitytests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package s3control
