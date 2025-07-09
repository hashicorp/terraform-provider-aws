//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=Arn -ListTagsOutTagsElem=ResourceTags -ServiceTagsMap -TagOp=TagResource -TagInIDElem=Arn -TagInTagsElem=Tags -UntagOp=UntagResource -UntagInTagsElem=ResourceTagKeys -UpdateTags -KVTValues
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package arcregionswitch
