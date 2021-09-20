//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=GetApis,GetDomainNames -Export=yes
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=GetTags -ServiceTagsMap=yes -UpdateTags=yes

package apigatewayv2
