//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsInIDElem=Resource -ListTagsOutTagsElem=Tags.Items -ServiceTagsSlice "-TagInCustomVal=&cloudfront.Tags{Items: Tags(updatedTags.IgnoreAWS())}" -TagInIDElem=Resource "-UntagInCustomVal=&cloudfront.TagKeys{Items: aws.StringSlice(removedTags.IgnoreAWS().Keys())}" -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cloudfront
