//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsOp=DescribeTags -ListTagsInFiltIDName=auto-scaling-group -ServiceTagsSlice -TagOp=CreateOrUpdateTags -TagResTypeElem=ResourceType -TagType2=TagDescription -TagTypeAddBoolElem=PropagateAtLaunch -TagTypeIDElem=ResourceId -UntagOp=DeleteTags -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package autoscaling
