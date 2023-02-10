//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=ResourceName -ListTagsOutTagsElem=TagList -ServiceTagsSlice -TagOp=AddTagsToResource -TagInIDElem=ResourceName -UntagOp=RemoveTagsFromResource -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package docdb
