//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsInIDElem=ResourceARN -ServiceTagsSlice=yes -TagOp=AddTagsToResource -TagInIDElem=ResourceARN -UntagOp=RemoveTagsFromResource -UpdateTags=yes

package storagegateway
