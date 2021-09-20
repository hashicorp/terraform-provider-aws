//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=ListApps -Export=yes
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ServiceTagsMap=yes -UpdateTags=yes

package amplify
