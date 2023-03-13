//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeBrokerInstanceOptions -ContextOnly
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTags -ServiceTagsMap -TagOp=CreateTags -UntagOp=DeleteTags -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package mq
