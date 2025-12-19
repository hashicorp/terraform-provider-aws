Please dont define your own constants for instance types. Use the ones already available in awstypes.
Dont hardcode values for which constants exist in awstypes.
Update test cases should have values for all fields before and after.
When defining schema, any field present in the CreateOutput structure but not in the CreateInput structure needs to be have its compute attribute as True.
Ensure there is a newline at the end of markdown files.
