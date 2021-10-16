//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTagsForStream -ListTagsInIDElem=StreamName -ServiceTagsSlice=yes -TagOp=AddTagsToStream -TagOpBatchSize=10 -TagInCustomVal=aws.StringMap(updatedTags.IgnoreAWS().Map()) -TagInIDElem=StreamName -UntagOp=RemoveTagsFromStream -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kinesis
