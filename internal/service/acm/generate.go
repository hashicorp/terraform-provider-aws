//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForCertificate -ListTagsInIDElem=CertificateArn -ServiceTagsSlice -TagOp=AddTagsToCertificate -TagInIDElem=CertificateArn -UntagOp=RemoveTagsFromCertificate -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package acm
