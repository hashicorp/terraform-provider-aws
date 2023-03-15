//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForStream -ListTagsInIDElem=StreamARN -ServiceTagsMap -TagOp=TagStream -TagInIDElem=StreamARN -UntagOp=UntagStream -UntagInTagsElem=TagKeyList -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kinesisvideo
