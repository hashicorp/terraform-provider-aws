// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tags/main.go -KVTValues -ListTags -ListTagsInIDElem=ResourceArn -ServiceTagsMap -TagOp=TagResource -TagInIDElem=ResourceArn -UntagOp=UntagResource -UntagInTagsElem=TagKeys -UpdateTags
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package devopsagent
