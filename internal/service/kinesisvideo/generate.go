// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -ListOps=ListTagsForStream
//go:generate go run ../../generate/tags/main.go -KVTValues -ListTags -ListTagsOpPaginated -ListTagsOpPaginatorCustom -ListTagsOp=ListTagsForStream -ListTagsInIDElem=StreamARN -ServiceTagsMap -TagOp=TagStream -TagInIDElem=StreamARN -UntagOp=UntagStream -UntagInTagsElem=TagKeyList -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kinesisvideo
