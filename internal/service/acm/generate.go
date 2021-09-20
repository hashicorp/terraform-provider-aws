//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTagsForCertificate -ListTagsInIDElem=CertificateArn -ServiceTagsSlice=yes -TagOp=AddTagsToCertificate -TagInIDElem=CertificateArn -UntagOp=RemoveTagsFromCertificate -UntagInNeedTagType=yes -UntagInTagsElem=Tags -UpdateTags=yes

package acm
