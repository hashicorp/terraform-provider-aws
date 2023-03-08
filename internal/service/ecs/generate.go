//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeCapacityProviders -ContextOnly
//go:generate go run ../../generate/tagresource/main.go  -WithContext=false
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ServiceTagsSlice -UpdateTags -ParentNotFoundErrCode=InvalidParameterException "-ParentNotFoundErrMsg=The specified cluster is inactive. Specify an active cluster and try again." -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package ecs
