---
status: proposed
date: 2026-05-22
deciders: phil (theGreatWhiteShark)
---

# AD: Permalink URL construction

## Context and Problem Statement

`hydrogen-index` scans a directory of Hydrogen artifacts and produces an
`index.json` file. Each artifact entry contains a `url` field that is supposed
to be a permalink allowing the `hydrogen` application to download the artifact
over the internet.

Currently the `url` field is a stub — it simply holds the relative path from
the scan root (e.g. `v2.0.0.h2drumkit`). This is not a usable download URL.

Users will host their artifacts in git repositories and use `hydrogen-index`
(typically via CI) to build the index. The tool needs to construct proper
permalinks from the repository information and the artifact's relative path.

Two usage patterns need to be supported:

1. **Major providers** (GitHub, GitLab) — the user specifies the provider and
   repository, the tool constructs the permalink using the provider's known
   raw-file URL pattern.
2. **Self-hosted / arbitrary hosting** — the user specifies a base URL and the
   tool appends the artifact's relative path.

## Decision Drivers

- Minimal friction for the common case (GitHub/GitLab hosting).
- Full flexibility for self-hosted or custom hosting setups.
- No network access required at index-build time (all URL construction is
  local and deterministic).
- The solution should not require the user to manually maintain URLs.

## Considered Options

1. **Relative path only** (current) — `url` is the relative path. Requires
   external tooling or manual editing to produce real URLs.
2. **URL template** — user provides a template string with placeholders
   (e.g. `https://{host}/{owner}/{repo}/raw/{branch}/{path}`). Maximum
   flexibility but high cognitive load and error-prone.
3. **Base URL** — user provides `--base-url` and the tool appends the
   relative path. Simple and universal, but requires the user to know the
   raw-file URL prefix.
4. **Provider presets + base URL** — known providers (GitHub, GitLab) get
   convenience flags that construct the base URL automatically. Everything
   else falls back to `--base-url`.

## Decision Outcome

Option 4 is chosen. It combines the convenience of provider shortcuts for the
common case with the flexibility of `--base-url` for everything else.

### URL Construction

The URL of each artifact is constructed as:

```
{base-url}/{relative-path}
```

Where `{base-url}` is derived from one of three sources (in priority order):

1. **`--base-url` flag** — used directly, no transformation.
2. **Provider preset** — `--provider github --repo owner/repo --branch main`
   is expanded to the provider's raw-file URL pattern.
3. **Default** — empty base URL. Resulting `url` is the relative path only
   (current behavior, useful for local testing).

### Provider Presets

| Provider | Flags | Base URL pattern |
|---|---|---|
| GitHub | `--provider github --repo owner/repo --branch main` | `https://raw.githubusercontent.com/{owner}/{repo}/{branch}` |
| GitLab | `--provider gitlab --repo owner/repo --branch main` | `https://gitlab.com/{owner}/{repo}/-/raw/{branch}` |

New providers can be added incrementally as they become relevant.

### Flag Design

```
hydrogen-index scan -d /path/to/artifacts \
  --provider github --repo hydrogen-music/my-artifacts --branch main

# Equivalent to:
hydrogen-index scan -d /path/to/artifacts \
  --base-url https://raw.githubusercontent.com/hydrogen-music/my-artifacts/main
```

The `--base-url` flag takes precedence over `--provider`/`--repo`/`--branch`
if both are specified.

### Example

Given a repo with:
```
drumkits/
  v2.0.0.h2drumkit
patterns/
  boom.h2pattern
```

And flags `--provider github --repo user/artifacts --branch main`, the
resulting `url` fields would be:

```json
"url": "https://raw.githubusercontent.com/user/artifacts/main/drumkits/v2.0.0.h2drumkit"
"url": "https://raw.githubusercontent.com/user/artifacts/main/patterns/boom.h2pattern"
```

## Consequences

- Users can generate usable `index.json` files with a single command.
- CI workflows for GitHub and GitLab become trivial — just set the provider
  and repo flags.
- Self-hosted solutions use `--base-url` for full control.
- The default behavior (relative path only) remains for local testing and
  backward compatibility.
- The `url` field in `index.json` is no longer a stub — it contains a real
  permalink.
