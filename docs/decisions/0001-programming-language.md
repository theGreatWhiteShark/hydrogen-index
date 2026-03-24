---
status: accepted
date: 2026-03-24
deciders: phil (theGreatWhiteShark)
---

# AD: Programming language to implement `hydrogen-index`

## Context and Problem Statement

As discussed in MADR 0000 we want to have a CLI tool called `hydrogen-index`. It
has to traverse a directory recursively and create a summary index file for all
Hydrogen artifacts encountered in it.

Which language should we write it in?

## Decision Drivers

- Code should be portable and platform-independent.
- If it requires a runtime and/or toolchain, that one should be easy to install
  and should be available in most common CI/CD infrastructures.
- It should allow for convenient maintaining.

## Considered Options

1. Native script language, like `bash` or `Windows Powershell`.
2. Interpreted language, like `Lua` or `Python`.
3. Compiled language, like `C`, `C++`, `Rust`, or `Go`.


## Decision Outcome

Option 1. is not portable. The CLI should be able to used on Linux, macOS,
Windows, and others. In addition, it does not come with a builtin support for
unit tests, which are essential for good stability and maintaining.

Option 2. is feasible and probably implementation will probably be done faster.
But the absence of a compiler (or at least a transpiler) has a couple of
drawbacks. Especially when it comes to long-time stability.

This leaves option 3. But which compiled language?

`C++` is a natural candidate since the `hydrogen` application itself is written
in it. But for a CLI tool - which is basically just a nice CLI argument parser,
a handful of XML parsing routines, and a couple of unit tests - this would be a
pain because of the absence of a well-supported package manager (yes, there are
solutions like `conan`. But the resulting toolchain would be over the top).

`TypeScript` as a transpiled and very common language would be a good candidate
too. But especially in its ecosystem I have rather bad experiences with
maintenance burdens introduced by the supply chain - the whole package stack
required for even the simplest features. Since `hydrogen-index` should be around
with minimal changes for years to come, this one wouldn't be a good choice.

`Rust` and `Go` are both very good candidates. Although I like `Rust` more from
both the conceptional as well as from tooling perspective, I think I will with
go with `Go` in here. It's more easy to understand and will more likely receive
external contributions and its toolchain is more easy to install.

### Consequences

* `hydrogen-index` will be implemented in `Go`.
