# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

def sdk_modules:
  [.Require[] | select(.Path == "github.com/aws/aws-sdk-go-v2"
    or (.Path | startswith("github.com/aws/aws-sdk-go-v2/")))] | INDEX(.Path);

def module_tag:
  if .Path == "github.com/aws/aws-sdk-go-v2" then .Version
  else (.Path | ltrimstr("github.com/aws/aws-sdk-go-v2/") | sub("/v([2-9]|[1-9][0-9]+)$"; "")) + "/" + .Version
  end;

def version_key:
  if type == "string" then
    (capture("^v(?<major>[0-9]+)\\.(?<minor>[0-9]+)\\.(?<patch>[0-9]+)$")
      | [.major, .minor, .patch] | map(tonumber)) // null
  else null end;

def release_key:
  ltrimstr("refs/tags/release-") | ltrimstr("release-")
  | split(".") | [.[0], (.[1] // "0" | tonumber)];

# A generic release requirement uses the newest release represented by each manifest.
# If any module tag is unknown, do not guess that manifest's release boundary.
def manifest_release($modules; $index; $releases):
  [$modules[] | module_tag | $index["refs/tags/" + .] as $module
    | if $module == null then null
      else ([$releases[] | select(.sha == $module.sha) | .ref | release_key] | max)
      end]
  | if length > 0 and all(.[]; . != null) then max else null end;

def module_requirements:
  [capture(
    "(?<path>github\\.com/aws/aws-sdk-go-v2(?:/[A-Za-z0-9._/-]+)?)[`\\]) \\t:,(=*-]{0,20}(?<version>v[0-9][0-9A-Za-z.+-]*)";
    "g")
    | .version |= rtrimstr(".")];

def newly_satisfied($requirements; $old; $new):
  all($requirements[];
    . as $r | ($r.version | version_key) as $need | ($new[$r.path].Version | version_key) as $installed
    | $need != null and $installed != null and $need <= $installed)
  and any($requirements[];
    . as $r | ($r.version | version_key) as $need | ($old[$r.path].Version | version_key) as $installed
    | $need != null and ($old[$r.path] == null or ($installed != null and $installed < $need)));

def release_requirements($releases):
  [
    (capture("https://github\\.com/aws/aws-sdk-go-v2/releases/tag/(?<tag>release-[0-9]{4}-[0-9]{2}-[0-9]{2}(?:\\.[0-9]+)?)"; "g")
      | .tag | release_key),
    (capture("AWS\\s+SDK\\s+for\\s+Go\\s+v2\\s*\\[?Release[\\s-]+(?<date>[0-9]{4}-[0-9]{2}-[0-9]{2}(?:\\.[0-9]+)?)"; "ig")
      | .date | release_key),
    (capture("https://github\\.com/aws/aws-sdk-go-v2/commit/(?<sha>[0-9a-fA-F]{7,40})"; "g")
      | (.sha | ascii_downcase) as $sha
      | [$releases[] | select(.sha | startswith($sha))] as $matches
      | select(($matches | unique_by(.sha) | length) == 1)
      | $matches[] | .ref | release_key)
  ] | unique;

($before[0] | sdk_modules) as $old
| ($after[0] | sdk_modules) as $new
| ($tags[0] | INDEX(.ref)) as $index
| [$tags[0][] | select(.ref | test("^refs/tags/release-[0-9]{4}-[0-9]{2}-[0-9]{2}(\\.[0-9]+)?$"))] as $releases
| manifest_release($old; $index; $releases) as $start
| manifest_release($new; $index; $releases) as $end
| [
    .[] | select(.state == "open" and .pull_request == null)
    | ((.title // "") + "\n" + (.body // "")) as $text
    | ($text | module_requirements) as $requirements
    # Explicit module requirements take precedence over a less precise SDK-wide release date.
    | select(if ($requirements | length) > 0 then newly_satisfied($requirements; $old; $new)
      else ($text | release_requirements($releases) | max) as $need
        | $start != null and $end != null and $need != null and $start < $need and $need <= $end
      end)
    | .number
  ]
| unique
| if length == 0 then "No linked issue."
  else map("Relates #\(.).") | join("\n")
  end
