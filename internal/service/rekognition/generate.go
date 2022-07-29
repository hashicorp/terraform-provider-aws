//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForResource -ListTagsInIDElem=ResourceArn -ServiceTagsMap -TagOp=TagResource -TagInIDElem=ResourceArn -UntagOp=UntagResource -UntagInTagsElem=TagKeys -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package rekognition
