// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTags -ListTagsOpPaginated -ListTagsInIDElem=CertificateAuthorityArn -ServiceTagsSlice -TagOp=TagCertificateAuthority -TagInIDElem=CertificateAuthorityArn -UntagOp=UntagCertificateAuthority -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags -AWSSDKVersion=2
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package acmpca
