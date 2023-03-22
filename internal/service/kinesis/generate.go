//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForStream -ListTagsInIDElem=StreamName -ServiceTagsSlice -TagOp=AddTagsToStream -TagOpBatchSize=10 -TagInCustomVal=aws.StringMap(updatedTags.IgnoreAWS().Map()) -TagInIDElem=StreamName -UntagOp=RemoveTagsFromStream -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kinesis
