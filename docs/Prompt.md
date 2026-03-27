The initial codebase of this repo was created using a LLM. As several different
flavors were used, this document will summarize the conditions in order to
reproduce or judge the output.

## Tools

- OpenCode 1.3.2
- oh-my-opencode-slim: adfe7425a6cfb4332f4907baf7d0fa3ae1fc6fb1
- Github Copilot as providor

## Prompt

This repo is intended to hold the source code of a Go application called `hydrogen-index`. Both its intended purpose and overall design are outlines in docs/decisions. Please have a look at these files for context. `hydrogen-index` should use the Go library github.com/spf13/cobra for CLI argument parsing and feature both a version and help option. It have a command `scan`, which is also triggered when called it without any arguments. This should make it traverse the file system until it finds the root of the current git repository or fail and report an error in case it was not called in a git repo. Afterwards, it should traverse all folders recursively and scan XML files like those in res/hydrogen-artifacts. It should parse them and store the content deemed relevant in MADR 0002. It should also be able to handle all legacy formats, which can be found in res/hydrogen-artifacts/legacy* . Please note, that `.h2drumkit` files are tar archives. They have to be extracted to a temporary folder first in order to acces the contained `drumkit.xml` XML file containing the artifacts information. Once all files are parsed, it should create a `index.json` in the folder `hydrogen-index` was called in containing a summary of all encountered artifacts. The format of `index.json` must be similar to the one in res/references-index.json. The `scan` command should two additional options. `-d` restricts the scan to a particular folder provided as CLI argument. When present, `hydrogen-index` will not search for the root of a git repo but just operates in this folder. `-o` should allow the user to specifiy a different path and name for the `index.json` output file. Both options support absolute and relative paths and should work on all platforms supported the Go programming language. A second command called `validate` will check whether a provided input file has the expected format of the `index.json`. Please create a JSON specification based on the potential output of `hydrogen-index` to base this validation on. While implementing, please write comments to describe your intend, develop it in a test-driven way, and create a top-level README.md to describe the resulting. Please plan how to implement such a tool.


## Configs

### ~/.config/opencode/oh-my-opencode-slim.json

```
{
  "preset": "copilot",
  "presets": {
    "copilot": {
      "orchestrator": { "model": "github-copilot/claude-opus-4.6", "variant": "high", "skills": ["*"], "mcps": ["websearch"] },
      "oracle": { "model": "github-copilot/claude-opus-4.6", "variant": "high", "skills": [], "mcps": [] },
      "librarian": { "model": "github-copilot/gpt-5.1-codex-mini", "variant": "low", "skills": [], "mcps": ["websearch", "context7", "grep_app"] },
      "explorer": { "model": "github-copilot/gpt-5.1-codex-mini", "variant": "low", "skills": [], "mcps": [] },
      "designer": { "model": "github-copilot/gemini-3.1-pro-preview", "variant": "medium", "skills": ["agent-browser"], "mcps": [] },
      "fixer": { "model": "github-copilot/claude-sonnet-4.6", "variant": "low", "skills": [], "mcps": [] }
    }
  }
}
```

### ~/.config/opencode/opencode.json

```
{
  "$schema": "https://opencode.ai/config.json",
  "autoupdate": "notify",
  "agent": {
    "build": {
      "disable": true
    },
    "explore": {
      "disable": true
    },
    "general": {
      "disable": true
    },
    "plan": {
      "disable": true
    }
  },
  "command": {
    "mystatus": {
      "description": "Query quota usage for all AI accounts",
      "template": "Use the mystatus tool to query quota usage. Return the result as-is without modification."
    }
  },
  "formatter": false,
  "instructions": [
    ".config/opencode/instructions/*.md"
  ],
  "permission": {
    "bash": {
      "*": "ask",
      "ast-grep *": "allow",
      "awk *": "allow",
      "basename *": "allow",
      "biome *": "ask",
      "bun *": "ask",
      "cargo *": "ask",
      "cat *": "allow",
      "cloc *": "allow",
      "cp *": "allow",
      "date *": "allow",
      "diff *": "allow",
      "dirname *": "allow",
      "du *": "allow",
      "echo *": "allow",
      "fd *": "allow",
      "file *": "allow",
      "find *": "allow",
      "g++*": "allow",
      "gcc*": "allow",
      "git checkout*": "ask",
      "git commit*": "ask",
      "git push*": "ask",
      "git reset*": "ask",
      "grep *": "allow",
      "gunzip *": "allow",
      "gzip *": "allow",
      "head *": "allow",
      "hostname": "allow",
      "id": "allow",
      "llvm*": "allow",
      "ls *": "allow",
      "make *": "allow",
      "mkdir *": "allow",
      "mv *": "allow",
      "npm *": "ask",
      "open *": "allow",
      "printf *": "allow",
      "pwd": "allow",
      "readlink *": "allow",
      "realpath *": "allow",
      "rg *": "allow",
      "rm*": "ask",
      "sed *": "allow",
      "shellcheck *": "allow",
      "sort *": "allow",
      "stat *": "allow",
      "tail *": "allow",
      "tar *": "allow",
      "tee *": "allow",
      "touch *": "allow",
      "tr *": "allow",
      "tree *": "allow",
      "true": "allow",
      "tsc *": "allow",
      "uname *": "allow",
      "uniq *": "allow",
      "unzip *": "allow",
      "wc *": "allow",
      "which *": "allow",
      "whoami": "allow",
      "xargs *": "ask",
      "zip *": "allow"
    },
    "external_directory": {
      "*": "ask",
      "/tmp/*": "allow",
      "~/.config/opencode/**": "allow"
    },
    "lsp": "allow",
    "read": {
      "*": "allow",
      "*.env": "deny",
      "*.env.*": "deny",
      "*.env.example": "allow",
      "*.id_ed25519": "deny",
      "*.id_rsa": "deny",
      "*.key": "deny",
      "*.pem": "deny"
    },
    "webfetch": "allow"
  },
  "plugin": [
    "@mohak34/opencode-notifier@latest",
    "opencode-openai-codex-auth",
    "opencode-mystatus",
    "oh-my-opencode-slim"
  ],
  "tools": {
    "ast-grep": true,
    "webfetch": true
  }
}

```
