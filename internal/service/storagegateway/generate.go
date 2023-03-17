//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceARN -ServiceTagsSlice -TagOp=AddTagsToResource -TagInIDElem=ResourceARN -UntagOp=RemoveTagsFromResource -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package storagegateway
