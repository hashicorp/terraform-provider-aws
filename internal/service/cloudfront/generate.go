// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsInIDElem=Resource -ListTagsOutTagsElem=Tags.Items -ServiceTagsSlice "-TagInCustomVal=&awstypes.Tags{Items: Tags(updatedTags)}" -TagInIDElem=Resource "-UntagInCustomVal=&awstypes.TagKeys{Items: removedTags.Keys()}" -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cloudfront
