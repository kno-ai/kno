# Changelog

## [0.9.0](https://github.com/kno-ai/kno/compare/v0.8.1...v0.9.0) (2026-03-08)


### Features

* add /kno entry point, refine skills voice, awareness-first docs ([517e47d](https://github.com/kno-ai/kno/commit/517e47d7ba35a3cb672376773b13be4a45c3a86f))


### Bug Fixes

* disable staticcheck cache to prevent tar corruption in CI ([c7d23db](https://github.com/kno-ai/kno/commit/c7d23db5567fe8442cb6b1dbb91715aec919518e))
* reject unexpected positional args in setup command ([f882f9a](https://github.com/kno-ai/kno/commit/f882f9a8e5f490e9eb9ce9c6d3308cb6c5d920df))
* remove unused nilIfEmpty function flagged by staticcheck ([dc0d919](https://github.com/kno-ai/kno/commit/dc0d9191883578bc61076278239b375b119f3d3f))
* skip staticcheck Go install to avoid cache corruption ([584c6a1](https://github.com/kno-ai/kno/commit/584c6a1626d679e7953bf1c06dcf8aa1884d4796))
* use correct input name for staticcheck cache disable ([f21f29c](https://github.com/kno-ai/kno/commit/f21f29c53babae3b37dcf81208615e86860acb4f))

## [0.8.1](https://github.com/kno-ai/kno/compare/v0.8.0...v0.8.1) (2026-03-08)


### Bug Fixes

* gofmt formatting in setup.go and config.go ([ab33640](https://github.com/kno-ai/kno/commit/ab33640ed214c57392e015b70fbeffc44f1e7c2b))

## [0.8.0](https://github.com/kno-ai/kno/compare/v0.7.0...v0.8.0) (2026-03-08)


### Features

* add publish command for exporting pages to markdown with frontmatter ([ada89f5](https://github.com/kno-ai/kno/commit/ada89f540f8ea7b95f36a4f455025a45584fbcd3))

## [0.7.0](https://github.com/kno-ai/kno/compare/v0.6.0...v0.7.0) (2026-03-08)


### Features

* add active awareness, refine voice and docs for all audiences ([7fbba6e](https://github.com/kno-ai/kno/commit/7fbba6e2be20bb4efa3b8832f442f6253d3391c3))

## [0.6.0](https://github.com/kno-ai/kno/compare/v0.5.0...v0.6.0) (2026-03-07)


### Features

* enforce content size limits, deduplicate metadata, consolidate auto-removal ([3fa393a](https://github.com/kno-ai/kno/commit/3fa393a10aecfafdb199e5d0316fa94f8ab36a40))

## [0.5.0](https://github.com/kno-ai/kno/compare/v0.4.0...v0.5.0) (2026-03-07)


### ⚠ BREAKING CHANGES

* Metadata fields renamed (distilled_at → curated_at, distilled_into → curated_into, last_distilled_at → last_curated_at). Config section renamed ([distill] → [curate]). Skills renamed (save → capture, distill → curate).

### Features

* add page rename with file and reference updates ([993201b](https://github.com/kno-ai/kno/commit/993201b8f34b5d15b5ba0822a6d7548b0f016705))
* flatten page storage for Obsidian browsability ([e5250d7](https://github.com/kno-ai/kno/commit/e5250d7dd3e3f581e1a6e718b9fc514e829870b8))
* readable note IDs, doc consistency pass, remove personal references ([dbb059c](https://github.com/kno-ai/kno/commit/dbb059c341e0d08959d517fa8aaba805202e857e))
* rename knowledge loop to capture/curate/load, flatten CLI, add delete tools ([61e8528](https://github.com/kno-ai/kno/commit/61e852821edd3fc8c2b942db24785853ef0d6917))


### Bug Fixes

* move directory field out of repository block in goreleaser config ([68d1632](https://github.com/kno-ai/kno/commit/68d1632048bd786ecdbb2841125aa0fe78b2aa15))
* place Homebrew formula in Formula/ directory ([338e579](https://github.com/kno-ai/kno/commit/338e579575ac7a826b25c178bd1c2f9e5cdd245b))
* use manifest mode for release-please to respect bump-minor-pre-major ([13eb73c](https://github.com/kno-ai/kno/commit/13eb73c87decad3ea4f45509a4ccbf7127e1ef6e))

## [0.6.0](https://github.com/kno-ai/kno/compare/v0.5.0...v0.6.0) (2026-03-07)


### ⚠ BREAKING CHANGES

* Metadata fields renamed (distilled_at → curated_at, distilled_into → curated_into, last_distilled_at → last_curated_at). Config section renamed ([distill] → [curate]). Skills renamed (save → capture, distill → curate).

### Features

* rename knowledge loop to capture/curate/load, flatten CLI, add delete tools ([61e8528](https://github.com/kno-ai/kno/commit/61e852821edd3fc8c2b942db24785853ef0d6917))


### Bug Fixes

* use manifest mode for release-please to respect bump-minor-pre-major ([13eb73c](https://github.com/kno-ai/kno/commit/13eb73c87decad3ea4f45509a4ccbf7127e1ef6e))

## [0.5.0](https://github.com/kno-ai/kno/compare/v0.4.0...v0.5.0) (2026-03-07)


### Features

* add page rename with file and reference updates ([993201b](https://github.com/kno-ai/kno/commit/993201b8f34b5d15b5ba0822a6d7548b0f016705))
* flatten page storage for Obsidian browsability ([e5250d7](https://github.com/kno-ai/kno/commit/e5250d7dd3e3f581e1a6e718b9fc514e829870b8))
* readable note IDs, doc consistency pass, remove personal references ([dbb059c](https://github.com/kno-ai/kno/commit/dbb059c341e0d08959d517fa8aaba805202e857e))


### Bug Fixes

* place Homebrew formula in Formula/ directory ([338e579](https://github.com/kno-ai/kno/commit/338e579575ac7a826b25c178bd1c2f9e5cdd245b))

## [0.4.0](https://github.com/kno-ai/kno/compare/v0.3.0...v0.4.0) (2026-03-07)


### Features

* show setup hint for new users ([cc12fb9](https://github.com/kno-ai/kno/commit/cc12fb938984af80446a8b6301703faad64dd606))

## [0.3.0](https://github.com/kno-ai/kno/compare/v0.2.0...v0.3.0) (2026-03-07)


### Features

* add update check and release-please automation ([468653a](https://github.com/kno-ai/kno/commit/468653a22aa9cad6009bdeb603e63b5f27e6bd59))
