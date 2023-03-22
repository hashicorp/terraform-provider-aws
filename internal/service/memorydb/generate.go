//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeACLs,DescribeClusters,DescribeParameterGroups,DescribeSnapshots,DescribeSubnetGroups,DescribeUsers -ContextOnly
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTags -ListTagsOutTagsElem=TagList -ServiceTagsSlice -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package memorydb
