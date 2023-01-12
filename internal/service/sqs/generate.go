//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=ListQueueTags -ListTagsInIDElem=QueueUrl -ServiceTagsMap -TagOp=TagQueue -TagInIDElem=QueueUrl -UntagOp=UntagQueue -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package sqs
