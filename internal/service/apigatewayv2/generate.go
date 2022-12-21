//go:generate go run ../../generate/listpages/main.go -ListOps=GetApis,GetDomainNames,GetApiMappings,GetStages
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=GetTags -ServiceTagsMap -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package apigatewayv2
