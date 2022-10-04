//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeQueryDefinitions
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsLogGroup -ListTagsInIDElem=LogGroupName -ServiceTagsMap -TagOp=TagLogGroup -TagInIDElem=LogGroupName -UntagOp=UntagLogGroup -UntagInTagsElem=Tags -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package logs
