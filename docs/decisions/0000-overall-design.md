---
status: accepted
date: 2026-03-23
deciders: phil (theGreatWhiteShark)
---

# AD: Overall design of the online import in Hydrogen v2

## Context and Problem Statement

In `hydrogen` we provide a couple of custom XML file formats for data exchange:
`.h2song`, `.h2pattern`, and `.h2drumkit` (tar archive containing `drumkit.xml`
as well as samples). As of now - prior to version 2.0 of Hydrogen - only
`.h2drumkit` files are commonly shared via the internet. To ease exchange, there
is a custom widget in `hydrogen` serving as a frontend for a curated list of
drumkits we - the Hydrogen dev team - do host at SourceForge.

If a users wishes to share her own kits with other people, she has to host an
XML index file using the same structure as our
https://github.com/hydrogen-music/hydrogen-music/blob/main/feeds/drumkit_list.php
(yes, an XML file bearing a `.php` suffix...), which holds a list of permalinks
and some additional meta data per link. The URL of this file has then to be
included by other users in the online import dialog. This is doable but not very
accessible.

The main problems with the current approach are:

- The index files has to be manually maintained.
- There is no notion of versioning. So, it does not support updates of existing
  content with newer versions.
- It is focused on solely on `.h2drumkit`s.

The focus on drumkits was due to former architectural constraints. Prior to
version 2.0 all patterns in Hydrogen were associated with a single drumkit.
Switching to a different drumkit did in general corrupt the pattern. Therefore,
sharing `.h2patterns` and `.h2song` - which do heavily build on patterns - could
be only done for a particular drum kit and did not scale at all. From version
2.0 onward, the picture has changed. Now, `.h2pattern`s - and by extension
`.h2song`s - are no longer associate with a particular drumkit but linked to
their instruments using more general string-based identifiers instead (see
https://github.com/hydrogen-music/hydrogen/blob/main/docs/proposals/0002-drumkit-independent-patterns_v2.md
for details).

So, now it makes perfect sense for users to share both patterns and songs and we
have to come up with an accessible and scalable way for them to do so.

## Decision Drivers

- Hydrogen's online database must be decentralized. We must not and do not want
  to act as gatekeepers.
- Only minimal programming and administrative skills should be required. We want
  artists to share their creative content. Not primarily software developers.
- The solution should be scalable to large databases.

## Considered Options

1. We could extend the existing XML-based format / workflow to support song,
   patterns, etc. as well.
2. We introduce a CLI tool traverse a folder recursively and add all encountered
   Hydrogen artifacts to the index file automatically.
3. We teach the `hydrogen` or `h2cli` binary to do the traversal and index
   creation.

## Decision Outcome

Option 1. is off the table since it is neither accessible nor does it scale to
larger collections.

Option 3. is tempting. But I think it is not accessible enough either since the
user still needs a particular version of our binaries.

Option 2. could be done with a small and portable application or script which
can evolve at its own pace. It will be called `hydrogen-index`, resides in this
repo, and will come with templates for Gitlab CI, GitHub Action, ...
integration. All an user needs to do is to add, is as add is as a git submodule
and activate a particular CI workflow or to just fork another repo and add their
own content. We should provide a template repo too.

I'll go for option 2.

While at it, I will the previous XML-based format and come up with a more
fitting JSON-based one.

### Consequences

* A CLI tool called `hydrogen-index` has to be written in a portable language,
  which creates an index file `hydrogen` can consume for online import.
* The JSON format of the index file has to be specified and mock data has to be
  provided for unit tests in the `hydrogen` repo.
* The `hydrogen` application itself has to be patched to support both the new
  format and the new features.
