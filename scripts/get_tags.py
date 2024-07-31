from collections import defaultdict
import json
import re
import os
import subprocess
import tempfile
from typing import TypedDict
import git
import shutil
import argparse

AWS_PROVIDER = "https://github.com/hashicorp/terraform-provider-aws"
PLUGIN_SDK = "github.com/hashicorp/terraform-plugin-sdk/v2"
PATCHED_SDK = "github.com/cloutierMat/terraform-plugin-sdk/v2"
PATCH_TAG_FILE = "patched_tags.json"
TAGS_REGEX = r"v\d+\.\d+\.\d+"
a = defaultdict(lambda: 0)


class Tag(TypedDict):
    tag: str
    sdk_version: str | None
    sdk_patched: bool
    terraform_patched: bool


def find_all_tags(repo: git.Repo) -> list[Tag]:
    return [
        Tag(
            tag=str(tag),
            sdk_version=None,
            sdk_patched=False,
            terraform_patched=False,
        )
        for tag in sorted(repo.tags, key=lambda t: t.commit.committed_datetime)
    ]


def extract_correct_plugin_version() -> str:
    tag = os.getenv("GITHUB_REF")
    if not tag:
        print("No tag found from 'GITHUB_REF'")
        os._exit(1)
    
    found = re.match(TAGS_REGEX, tag.split()[-1])
    if not found:
        print(f"Invalid tag {tag}")
        os._exit(1)
    
    print(f"Using ref {tag}")

    patched_tags = {}
    if os.path.exists(PATCH_TAG_FILE):
        with open(PATCH_TAG_FILE, "r") as file:
            patched_tags = json.load(file) or {}

    if not patched_tags:
        # os._exit(1)
        print(f"Unable to load in tag document {PATCH_TAG_FILE}")
        os._exit(1)

    tag_info = patched_tags.get(tag)
    if not tag_info:
        print(f"No SDK and patch information available at {tag}")
        os._exit(1)
    
    current_sdk_version = tag_info.get("sdk_version")

    patched_versions = get_available_versions(PATCHED_SDK)
    if current_sdk_version not in patched_versions:
        print(f"No patched sdk version available for {current_sdk_version} in {PATCHED_SDK}. Available tags: {patched_versions}")

    print(f"There is an available patch for the SDK in '{PATCHED_SDK}'")
    os._exit(1)


def get_available_versions(sdk : str) -> list[str]:
    process = subprocess.Popen(
        f"go list -m -versions {sdk}".split(),
        stdout=subprocess.PIPE,
    )

    plugin_sdk, _ = process.communicate()
    plugin_sdk = plugin_sdk.decode()

    return plugin_sdk.split()[1:]

def enrich_tag(tag: Tag, repo: git.Repo, valid_tags: list[str]=[]) -> Tag:
    print(f"checking_out {tag['tag']}")
    repo.git.checkout(f"tags/{tag['tag']}", force=True)
    process = subprocess.Popen(
        f"go list -m {PLUGIN_SDK}".split(),
        cwd=repo.common_dir,
        stdout=subprocess.PIPE,
    )
    plugin_sdk, _ = process.communicate()
    plugin_sdk = plugin_sdk.decode()

    if not plugin_sdk:
        return tag

    if PLUGIN_SDK in plugin_sdk:
        found = re.search(TAGS_REGEX, plugin_sdk) or [None]
        tag["sdk_version"] = found[0]
        if (tag := found[0]):
            tag["sdk_patched"] = tag in valid_tags

    return tag

def is_git_repo(path):
    try:
        _ = git.Repo(path).git_dir
        return True
    except git.exc.InvalidGitRepositoryError:
        return False

def fetch_available_tags(sdk: str,repo: git.Repo):
    tags = get_available_versions(sdk)
    
    refspecs = [f'refs/tags/{tag}:refs/tags/{tag}' for tag in tags]
    
    repo.remotes.origin.fetch(refspec=refspecs)
    print(f"Fetched tags: {', '.join(tags)}")
    return tags

# if __name__ == "__main__":
#     plugin_tags = get_available_versions(PLUGIN_SDK)
#     patched_tags = get_available_versions(PATCHED_SDK)
    
#     common = [tag for tag in plugin_tags if tag in patched_tags]
#     print("Available tags in patched repo:",common)


if __name__ == "__main__":
    # Create the main parser
    parser = argparse.ArgumentParser(description='Python operators to manage the release workflow')

    # Create sub-parsers
    subparsers = parser.add_subparsers(dest='operation', help='Available operations')
    
    add_parser = subparsers.add_parser('generate_patch_tags', help='generate tags file')
    subtract_parser = subparsers.add_parser('extract_sdk_version', help='extract sdk version')


    # Parse the command-line arguments
    args = parser.parse_args()

    if args.operation == 'generate_patch_tags':
        result = sum(args.numbers)
    elif args.operation == 'extract_sdk_version':
        extract_correct_plugin_version()
    

def xxx():
    patched_tag = {}
    if os.path.exists(PATCH_TAG_FILE):
        with open(PATCH_TAG_FILE, "r") as file:
            patched_tag = json.load(file) or {}
    else:
        with open(PATCH_TAG_FILE, "w") as file:
            json.dump({}, file)

    temp_dir = tempfile.mkdtemp(prefix="temp_aws_provider_")
    try:
        if is_git_repo(temp_dir):
            aws_repo = git.Repo(temp_dir)
        else:
            print(f"Cloning into {temp_dir}...")
            aws_repo = git.Repo.clone_from(url=AWS_PROVIDER, to_path=temp_dir, depth=1)
            print(f"Finished cloning {temp_dir}.")

        aws_repo.remotes.origin.fetch(refspec="refs/tags/*:refs/tags/*")
        
        patched_tags = get_available_versions(PATCHED_SDK)
        
        unlisted = [
            tag for tag in find_all_tags(aws_repo) if tag["tag"] not in patched_tag
        ]
        added = {tag["tag"]: enrich_tag(tag, aws_repo, patched_tags) for tag in unlisted}
        patched_tag.update(added)
    finally:
        pass
        shutil.rmtree(temp_dir)

    with open(PATCH_TAG_FILE, "+r") as file:
        json.dump({k: patched_tag[k] for k in sorted(patched_tag)}, file, indent=4)