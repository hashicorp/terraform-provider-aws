//go:generate go run ../../generate/tagresource/main.go -WithContext=false
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsOp=ListTagsOfResource -ServiceTagsSlice -UpdateTags -ParentNotFoundErrCode=ResourceNotFoundException -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package dynamodb
