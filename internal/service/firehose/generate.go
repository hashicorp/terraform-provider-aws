//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListTagsForDeliveryStream -ListTagsInIDElem=DeliveryStreamName -ServiceTagsSlice -TagOp=TagDeliveryStream -TagInIDElem=DeliveryStreamName -UntagOp=UntagDeliveryStream -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package firehose
