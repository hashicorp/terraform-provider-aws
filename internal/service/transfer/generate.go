//go:generate go run ../../generate/tagresource/main.go -IDAttribName=resource_arn -UpdateTagsFunc=UpdateTagsNoIgnoreSystem
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsInIDElem=Arn -ServiceTagsSlice -TagInIDElem=Arn -UpdateTags
//go:generate go run ../../generate/tags/main.go "-TagInCustomVal=Tags(updatedTags)" -TagInIDElem=Arn "-UntagInCustomVal=aws.StringSlice(removedTags.Keys())" -UpdateTags -UpdateTagsFunc=UpdateTagsNoIgnoreSystem -- update_tags_no_system_ignore_gen.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package transfer
