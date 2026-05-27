// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForCertificate -ListTagsInIDElem=CertificateArn -ServiceTagsSlice -TagOp=AddTagsToCertificate -TagInIDElem=CertificateArn -UntagOp=RemoveTagsFromCertificate -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
//go:generate go run ../../generate/identitytests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package acm
