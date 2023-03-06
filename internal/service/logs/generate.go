//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeQueryDefinitions,DescribeResourcePolicies -ContextOnly
//go:generate go run ../../generate/tags/main.go -ListTags -ServiceTagsMap -UpdateTags -ContextOnly
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsLogGroup -ListTagsInIDElem=LogGroupName -ListTagsFunc=ListLogGroupTags -TagOp=TagLogGroup -TagInIDElem=LogGroupName -UntagOp=UntagLogGroup -UntagInTagsElem=Tags -UpdateTags -UpdateTagsFunc=UpdateLogGroupTags -ContextOnly -- log_group_tags_gen.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package logs
