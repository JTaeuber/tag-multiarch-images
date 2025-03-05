# tag-multiarch-images

`tag-multiarch-images` is a GitHub Action that tags all untagged versions created by pushing a multi architecture image to the GHCR.

## Quick start

Tagging all untagged versions for an organization package:

```yml
steps:
  - name: Tag
    uses: jtaeuber/tag-multiarch-images@v0.1.0
    with:
      gh_token: ${{ secrets.YOUR_TOKEN }}
      gh_user: your-org
      package_name: your-package
      tag: your-tag
```

## Permissions

This action uses `crane manifest` and `crane copy` commands in order to change the tags of your versions.

As a result, for this action to work, the token needs read:packages and write:packages permissions.

## Inputs

### gh_token

**Required** Secret access token with scopes `packages:read` and `packages:write`. See [Creating a personal access token
](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token) for more details about GitHub access tokens.

### gh_user

**Required** Name of the user or organization owning the package.

### package_name

**Required** Name of the package for which versions should be tagged.

### dry-run

**Optional** Boolean controlling whether to execute the action as a dry-run. When `true` the action will print out details of the version that will be tagged without actually tagging them. Defaults to `false`.
