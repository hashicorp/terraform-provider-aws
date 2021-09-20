//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListTagsForDeliveryStream -ListTagsInIDElem=DeliveryStreamName -ServiceTagsSlice=yes -TagOp=TagDeliveryStream -TagInIDElem=DeliveryStreamName -UntagOp=UntagDeliveryStream -UpdateTags=yes

package firehose
