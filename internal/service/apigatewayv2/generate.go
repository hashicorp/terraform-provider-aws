//go:generate go run ../../generate/listpages/main.go -ListOps=GetApis,GetDomainNames,GetVpcLinks -ContextOnly
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=GetTags -ServiceTagsMap -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package apigatewayv2
