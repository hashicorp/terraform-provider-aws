//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=DescribeQueryDefinitions
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTagsLogGroup -ListTagsInIDElem=LogGroupName -ServiceTagsMap=yes -TagOp=TagLogGroup -TagInIDElem=LogGroupName -UntagOp=UntagLogGroup -UntagInTagsElem=Tags -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cloudwatchlogs
