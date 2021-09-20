//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=DescribeTags -ListTagsInIDElem=LoadBalancerNames -ListTagsInIDNeedSlice=yes -ListTagsOutTagsElem=TagDescriptions[0].Tags -ServiceTagsSlice=yes -TagOp=AddTags -TagInIDElem=LoadBalancerNames -TagInIDNeedSlice=yes -TagKeyType=TagKeyOnly -UntagOp=RemoveTags -UntagInNeedTagKeyType=yes -UntagInTagsElem=Tags -UpdateTags=yes

package elb
