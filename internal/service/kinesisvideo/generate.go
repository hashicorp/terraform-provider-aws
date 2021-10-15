//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTagsForStream -ListTagsInIDElem=StreamARN -ServiceTagsMap=yes -TagOp=TagStream -TagInIDElem=StreamARN -UntagOp=UntagStream -UntagInTagsElem=TagKeyList -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kinesisvideo
