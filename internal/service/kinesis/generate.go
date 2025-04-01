// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForStream -ListTagsInIDElem=StreamName -ServiceTagsSlice -TagOp=AddTagsToStream -TagOpBatchSize=10 -TagInCustomVal=updatedTags.IgnoreAWS().Map() -TagInIDElem=StreamName -UntagOp=RemoveTagsFromStream -UpdateTags -KeyValueTagsFunc=KeyValueTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kinesis
