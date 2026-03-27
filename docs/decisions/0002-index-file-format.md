---
status: accepted
date: 2026-03-25
deciders: phil (theGreatWhiteShark)
---

# AD: Format of the resulting index file

## Context and Problem Statement

As discussed in MADR 0000 the CLI tool `hydrogen-index` should create an index
file containing permalinks to all Hydrogen assets in order to allow their online
import in `hydrogen`. This file will be called `index.SUFFIX`.

Which overall file type/format should we use and how do be structure the actual
content?

## Decision Drivers

- The file should be both human readable and allow for manual editing.
- It should scale reasonable to large repo sizes (but we do not need to
  over-engineer it either).

## Considered Options

1. We use an `XML` file - like in all other Hydrogen files up till now.
2. We use `JSON`.

## Decision Outcome

It is tempting to use option 1. since we do not need any new parsing routines in
`hydrogen` itself. But it's 2026. People might laugh and point fingers if we use
`XML` for a new format of data we send via the internet. But seriously, people
use `JSON` over `XML` for a reason. There are other choices too. But `JSON` is
currently the most frequently used one and is open invitation for contributors
to play with it without the need to learn new tricks.

The content itself will consist of a number of top-level nodes and lists of
per-hydrogen-artifact blocks.

### Top-level nodes

#### version

Version of `hydrogen-index` used to create the file. This will be the version of
the index file format as well as both are designed to evolve together.

#### created

Timestamp of the file creation.

#### patternCount

Number of pattern blocks in the **patterns** list. Might be utilized for
consistency checks and/or to display a progress bar while parsing the index file
in `hydrogen`.

#### songCount

Number of song blocks in the **song** list. Might be utilized for
consistency checks and/or to display a progress bar while parsing the index file
in `hydrogen`.

#### drumkitCount

Number of drumkit blocks in the **drumkits** list. Might be utilized for
consistency checks and/or to display a progress bar while parsing the index file
in `hydrogen`.

#### patterns

List of pattern blocks.

#### songs

List of song blocks.

#### drumkits

List of drumkit blocks.

#### hash

sha256sum of the index file with this **hash** node removed. `hydrogen` will
treat this node as optional and shows a warning dialog in case the checksum does
not match the file. This will allow

### Shared artifact block nodes

All artifact blocks feature a set of common blocks.

#### type

Specifies which list the block resides in. In the first implementation in
`hydrogen` I would load the whole index file into RAM and parse it. There this
node is redundant. But in case index files get really big, we might have to
resort to a streaming-based approach in which we parse blocks of index files as
they arrive in the network interface. Than this node will be crucial.

#### name

Unique string representing the artifact in `hydrogen`'s side of the download
dialog. `hydrogen-index` will ensure this value is indeed unique but the user
might edit the resulting index file by hand. Thus, `hydrogen` should not rely on
it being unique.

#### url

Permalink to the artifact.

#### hash

sha256sum of the artifact accessible via the URL above.

### author

Name or pseudonym of the person who did create the artifact. This might coincide
with the holder of the copyright but does not have to. In case it does not
match, the copyright holder has to be referenced in the **license** node string.

### description

Free text description of the artifact. Initially we do not introduce any size
constraints in here. But we might do at a latter point in case it causes
problems.

#### version

Integer value set by the user to indicate the revision of the artifact.

#### formatVersion

Integer value set by `hydrogen` to indicate the revision of the overall format
of the artifact.

#### tags

List of categories set by the user to describe the artifact. Will allow for
filtering artifacts and is key when scaling to large databases.

#### size

File size of the artifact in bytes.

#### license

License identifier of the artifact.

### Specific artifact block nodes

Some artifact blocks will contain additional nodes.

#### pattern block

- **notes**: number of notes in the pattern.
- **instrumentTypes**: full list of all instrument types associated with the
  notes in the pattern (these will be used to map the notes to a particular
  drumkit).

#### song block

- **patterns**: number of patterns available in the song (not to be confused
  with the length of the songs in patterns).

#### drumkit block

- **instruments**: number of instruments in the drumkit.
- **components**: number of instrument components in the drumkit.
- **samples**: number of samples in the drumkit.
- **instrumentTypes**: full list of all instrument types associated with the
  instruments (these will be used to map the instruments to the notes within
  patterns).

Note that no nodes regarding the drumkit image or image license are contained.
This feature is almost never used in drumkits I encountered and I do not see how
it would add valuable information. Plus, we would have to make room displaying
the image for in the download dialog in `hydrogen`.

## Consequences

* `hydrogen-index` will be produce a file called `index.json`, which has to
  comply with the specification above.
