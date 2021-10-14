//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsInIDElem=Resource -ListTagsOutTagsElem=Tags.Items -ServiceTagsSlice=yes "-TagInCustomVal=&cloudfront.Tags{Items: Tags(updatedTags.IgnoreAWS())}" -TagInIDElem=Resource "-UntagInCustomVal=&cloudfront.TagKeys{Items: aws.StringSlice(removedTags.IgnoreAWS().Keys())}" -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cloudfront
