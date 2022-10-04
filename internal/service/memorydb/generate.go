//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeACLs,DescribeClusters,DescribeParameterGroups,DescribeSnapshots,DescribeSubnetGroups,DescribeUsers
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTags -ListTagsOutTagsElem=TagList -ServiceTagsSlice -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package memorydb
