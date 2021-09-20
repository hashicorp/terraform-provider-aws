//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListQueueTags -ListTagsInIDElem=QueueUrl -ServiceTagsMap=yes -TagOp=TagQueue -TagInIDElem=QueueUrl -UntagOp=UntagQueue -UpdateTags=yes

package sqs
