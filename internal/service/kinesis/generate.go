//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTagsForStream -ListTagsInIDElem=StreamName -ServiceTagsSlice=yes -TagOp=AddTagsToStream -TagOpBatchSize=10 -TagInCustomVal=aws.StringMap(updatedTags.IgnoreAws().Map()) -TagInIDElem=StreamName -UntagOp=RemoveTagsFromStream -UpdateTags=yes

package kinesis
