//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForResource -ListTagsInIDElem=ResourceArn -ListTagsOutTagsElem=TagList -ServiceTagsSlice -TagOp=TagResource -TagInTagsElem=TagList -TagInIDElem=ResourceArn -UpdateTags -TagType=Tag -ContextOnly

// ONLY generate directives and package declaration! Do not add anything else to this file.

package fms
