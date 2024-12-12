# Tags Generator

This generator (`main.go`) generates files named `tags_gen`, such as `internal/service/ecs/tags_gen.go`. These files provide consistency in handling AWS resource tags. Initiate generating by calling `make gen` from the provider directory.

## Generator Directives

Control the code generated using flags of the directives that you include in a `generate.go` file for an individual service. For example, a file such as `internal/service/ecs/generate.go` may contain three directives (and a package declaration). This generator corresponds to the `../../generate/tags/main.go` directive. (The other directives are documented in their respective packages.)

```go
//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeCapacityProviders
//go:generate go run ../../generate/tagresource/main.go
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ServiceTagsSlice -UpdateTags

package ecs
```

**NOTE:** A `generate.go` file should _only_ contain generator directives and a package declaration. Do not include related Go functions in this file.

## Generator Directive Flags

Some flags control generation a certain section of code, such as whether the generator generates a certain function. Other flags determine how generated code will work. Do not include flags where you want the generator to use the default value.

| Flag | Default | Description | Example Use |
| --- | --- | --- | --- |
| `CreateTags` | `false` | Whether to generate `CreateTags` | `-CreateTags` |
| `CreateTagsFunc` | `createTags` | Name of the generated `CreateTags` function | `-CreateTagsFunc=createTags2` |
| `GetTag` | `false` | Whether to generate `GetTag` | `-GetTag` |
| `GetTagFunc` | `findTag` | Name of the generated `GetTag` function | `-GetTagFunc=findTag2` |
| `ListTags` | `false` | Whether to generate `ListTags` | `-ListTags` |
| `ListTagsFunc` | `listTags` | Name of the generated `ListTags` function | `-ListTagsFunc=listTags2` |
| `UpdateTags` | `false` | Whether to generate `UpdateTags` | `-UpdateTags` |
| `UpdateTagsFunc` | `updateTags` | Name of the generated `UpdateTags` function | `-UpdateTagsFunc=updateTags2` |
| `UpdateTagsNoIgnoreSystem` | `false` | Whether to ignore system tags in `UpdateTags` | `-UpdateTagsNoIgnoreSystem` |
| `ServiceTagsMap` | `false` | Whether to generate map service tags (use this or `ServiceTagsSlice`, not both) | `-ServiceTagsMap` |
| `KVTValues` | `false` | Whether map service tags have string pointer values | `-KVTValues` |
| `EmptyMap` | `false` | Whether map service tags should be empty for no tags | `-EmptyMap` |
| `ServiceTagsSlice` | `false` | Whether to generate slice service tags (use this or `ServiceTagsMap`, not both) | `-ServiceTagsSlice` |
| `KeyValueTagsFunc` | `KeyValueTags` | Name of the generated `KeyValueTags` function | `-KeyValueTagsFunc=keyValueTags2` |
| `TagsFunc` | `Tags` | Name of the generated `Tags` function | `-TagsFunc=tags2` |
| `GetTagsInFunc` | `getTagsIn` | Name of the generated `getTagsIn` function | `-GetTagsInFunc=getTagsIn2` |
| `SetTagsOutFunc` | `setTagsOut` | Name of the generated `setTagsOut` function | `-SetTagsOutFunc=setTagsOut2` |
| `Wait` | `false` | Whether to generate `waitTagsPropagated` | `-Wait` |
| `WaitFunc` | `waitTagsPropagated` | Name of the generated `waitTagsPropagated` function | `-WaitFunc=waitTagsPropagated2` |
| `WaitContinuousOccurence` | `0` | `ContinuousTargetOccurence` for `waitTagsPropagated` function | `-WaitContinuousOccurence=2` |
| `WaitFuncComparator` | `Equal` | Name of the function used for tags comparison during wait | `-WaitFuncComparator=ContainsAll` |
| `WaitDelay` | `0` | "Delay" for `waitTagsPropagated` function | `-WaitDelay=10s` |
| `WaitMinTimeout` | `0` | "MinTimeout" (minimum poll interval) for `waitTagsPropagated` function | `-WaitMinTimeout=1s` |
| `WaitPollInterval` | `0` | "PollInterval" for `waitTagsPropagated` function | `-WaitPollInterval=5s` |
| `WaitTimeout` | `0` | "Timeout" for `waitTagsPropagated` function | `-WaitTimeout=2m` |
| `ListTagsInFiltIDName` |  | List tags input filter identifier name | `-ListTagsInFiltIDName=resource-id` |
| `ListTagsInIDElem` | `ResourceArn` | List tags input identifier element | `-ListTagsInIDElem=ResourceARN` |
| `ListTagsInIDNeedValueSlice` | `false` | Whether list tags input identifier needs a slice | `-ListTagsInIDNeedSlice` |
| `ListTagsOp` | `ListTagsForResource` | List tags operation | `-ListTagsOp=ListTags` |
| `ListTagsOpPaginated` | `false` | Whether `ListTagsOp` is paginated | `-ListTagsOpPaginated` |
| `ListTagsOutTagsElem` | `Tags` | List tags output tags element | `-ListTagsOutTagsElem=TagList` |
| `TagInCustomVal` |  | Tag input custom value | `-TagInCustomVal=aws.StringMap(updatedTags.IgnoreAWS().Map())` |
| `TagInIDElem` | `ResourceArn` | Tag input identifier element | `-TagInIDElem=ResourceARN` |
| `TagInIDNeedValueSlice` | `false` | Tag input identifier needs a slice of values | `-TagInIDNeedValueSlice` |
| `TagInTagsElem` | Tags | Tag input tags element | `-TagInTagsElem=TagsList` |
| `TagKeyType` |  | Tag key type | `-TagKeyType=TagKeyOnly` |
| `TagOp` | `TagResource` | Tag operation | `-TagOp=AddTags` |
| `TagOpBatchSize` | `0` | Tag operation batch size | `-TagOpBatchSize=10` |
| `TagResTypeElem` |  | Tag resource type field | `-TagResTypeElem=ResourceType` |
| `TagResTypeElemType` |  | Tag resource type field type | `-TagResTypeElem=ResourceTypeForTagging` |
| `TagType` | `Tag` | Tag type | `-TagType=TagRef` |
| `TagType2` |  | Second tag type | `-TagType2=TagDescription` |
| `TagTypeAddBoolElem` |  | Tag type additional boolean element | `-TagTypeAddBoolElem=PropagateAtLaunch` |
| `TagTypeIDElem` |  | Tag type identifier field | `-TagTypeIDElem=ResourceId` |
| `TagTypeKeyElem` | `Key` | Tag type key element | `-TagTypeKeyElem=TagKey` |
| `TagTypeValElem` | `Value` | Tag type value element | `-TagTypeValElem=TagValue` |
| `UntagInCustomVal` |  | Untag input custom value | `-UntagInCustomVal="&cloudfront.TagKeys{Items: aws.StringSlice(removedTags.IgnoreAWS().Keys())}"` |
| `UntagInNeedTagKeyType` | `false` | Untag input needs tag key type | `-UntagInNeedTagKeyType` |
| `UntagInNeedTagType` | `false` | Untag input needs tag type | `-UntagInNeedTagType` |
| `UntagInTagsElem` | `TagKeys` | Untag input tags element | `-UntagInTagsElem=Tags` |
| `UntagOp` | `UntagResource` | Untag operation | `-UntagOp=DeleteTags` |
| `ParentNotFoundErrCode` |  | Parent _NotFound_ error code | `-ParentNotFoundErrCode=InvalidParameterException` |
| `ParentNotFoundErrMsg` |  | Parent _NotFound_ error Message | `"-ParentNotFoundErrMsg=The specified cluster is inactive. Specify an active cluster and try again."` |
