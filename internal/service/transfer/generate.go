//go:generate go run ../../generate/tagresource/main.go -IDAttribName=resource_arn
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsInIDElem=Arn -ServiceTagsSlice -TagInIDElem=Arn -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package transfer
