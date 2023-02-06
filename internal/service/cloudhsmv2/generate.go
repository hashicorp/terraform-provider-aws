//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTags -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=TagList -ServiceTagsSlice -TagInIDElem=ResourceId -TagInTagsElem=TagList -UntagInTagsElem=TagKeyList -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cloudhsmv2
