//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=DescribeCapacityProviders -Export=yes
//go:generate go run -tags generate ../../generate/tagresource/main.go
//go:generate go run -tags generate ../../generate/tags/main.go -GetTag=yes -ListTags=yes -ServiceTagsSlice=yes -UpdateTags=yes

package ecs
