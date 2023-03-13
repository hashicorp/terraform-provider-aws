//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTags -ListTagsInIDElem=CertificateAuthorityArn -ServiceTagsSlice -TagOp=TagCertificateAuthority -TagInIDElem=CertificateAuthorityArn -UntagOp=UntagCertificateAuthority -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package acmpca
