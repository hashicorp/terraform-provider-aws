//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=GetTags -ListTagsInIDElem=Arn -ServiceTagsMap=yes -TagOp=Tag -TagInIDElem=Arn -UntagOp=Untag -UntagInTagsElem=Keys -UpdateTags=yes

package resourcegroups
