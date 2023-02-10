//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeDirectoryConfigs,DescribeFleets,DescribeImageBuilders,DescribeStacks,DescribeUsers,ListAssociatedStacks -ContextOnly
//go:generate go run ../../generate/tags/main.go -ListTags -ServiceTagsMap -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package appstream
