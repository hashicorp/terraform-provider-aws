// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// ONLY generate directives and package declaration!
//
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceId -ServiceTagsSlice -TagInIDElem=ResourceId -UntagInTagsElem=TagKeys -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/identitytests/main.go
//go:generate go run ../../generate/tagstests/main.go
package s3files
