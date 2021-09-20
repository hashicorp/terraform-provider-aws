//go:generate go run -tags generate ../../generate/tagresource/main.go
//go:generate go run -tags generate ../../generate/tags/main.go -GetTag=yes -ListTags=yes -ListTagsOp=ListTagsOfResource -ServiceTagsSlice=yes -UpdateTags=yes

package dynamodb
