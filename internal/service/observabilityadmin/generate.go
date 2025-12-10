// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tags/main.go -ServiceTagsMap -KVTValues -ListTags -ListTagsInIDElem=ResourceARN -UpdateTags -TagOp=TagResource -TagInIDElem=ResourceARN -UntagOp=UntagResource -UntagInTagsElem=TagKeys
// ONLY generate directives and package declaration! Do not add anything else to this file.

package observabilityadmin
