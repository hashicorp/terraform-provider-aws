//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=DescribeTags -ListTagsInIDElem=LoadBalancerNames -ListTagsInIDNeedSlice=yes -ListTagsOutTagsElem=TagDescriptions[0].Tags -ServiceTagsSlice -TagOp=AddTags -TagInIDElem=LoadBalancerNames -TagInIDNeedSlice=yes -TagKeyType=TagKeyOnly -UntagOp=RemoveTags -UntagInNeedTagKeyType=yes -UntagInTagsElem=Tags -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elb
