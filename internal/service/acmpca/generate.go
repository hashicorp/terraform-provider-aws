//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTags -ListTagsInIDElem=CertificateAuthorityArn -ServiceTagsSlice=yes -TagOp=TagCertificateAuthority -TagInIDElem=CertificateAuthorityArn -UntagOp=UntagCertificateAuthority -UntagInNeedTagType=yes -UntagInTagsElem=Tags -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package acmpca
