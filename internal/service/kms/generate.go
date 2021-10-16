//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListResourceTags -ListTagsInIDElem=KeyId -ServiceTagsSlice=yes -TagInIDElem=KeyId -TagTypeKeyElem=TagKey -TagTypeValElem=TagValue -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kms
