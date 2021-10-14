//go:generate go run -tags generate ../../generate/tagresource/main.go
//go:generate go run -tags generate ../../generate/tags/main.go -GetTag=yes -ListTags=yes -ListTagsOp=ListTagsOfResource -ServiceTagsSlice=yes -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package dynamodb
