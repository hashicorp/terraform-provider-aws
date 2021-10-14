//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=ListQueueTags -ListTagsInIDElem=QueueUrl -ServiceTagsMap=yes -TagOp=TagQueue -TagInIDElem=QueueUrl -UntagOp=UntagQueue -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package sqs
