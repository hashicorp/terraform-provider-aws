//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ListTags -ListTagsOp=ListTagsForResource -ListTagsInIDElem=ResourceArn  -ServiceTagsSlice -TagOp=TagResource -TagInIDElem=ResourceArn  -UntagOp=UntagResource -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package rbin
