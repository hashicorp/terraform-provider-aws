//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTagsForResource -ListTagsInIDElem=ResourceArn -ListTagsOutTagsElem=TagList -ServiceTagsSlice=yes -TagOp=TagResource -TagInTagsElem=TagList -TagInIDElem=ResourceArn -UpdateTags=yes -TagType=Tag

// ONLY generate directives and package declaration! Do not add anything else to this file.

package fms
