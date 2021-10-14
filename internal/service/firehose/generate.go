//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTagsForDeliveryStream -ListTagsInIDElem=DeliveryStreamName -ServiceTagsSlice=yes -TagOp=TagDeliveryStream -TagInIDElem=DeliveryStreamName -UntagOp=UntagDeliveryStream -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package firehose
