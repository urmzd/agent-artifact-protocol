# Changelog

## 0.3.0 (2026-04-01)

### Features

- add format-aware section markers module ([1a5aae9](https://github.com/urmzd/artifact-protocol/commit/1a5aae96bd4ce62a6af25344119666057607e975))

### Documentation

- update contributing guide and tool skill doc ([b5b56a8](https://github.com/urmzd/artifact-protocol/commit/b5b56a817658a340bddbfdfe25fb91c1c7fd9ee3))
- update README with envelope resolution focus ([e893093](https://github.com/urmzd/artifact-protocol/commit/e893093641dbbc9fb23503b51c0cf3cb42f2437c))
- add JSON Pointer diff example ([1d586eb](https://github.com/urmzd/artifact-protocol/commit/1d586ebb3d5e5862109312700f11ff1bafd28fde))

### Refactoring

- delete rendering-hints and update core schemas ([20b00df](https://github.com/urmzd/artifact-protocol/commit/20b00dfff7410c726affc9171742e94f9d45da18))
- **tools**: update Python tooling for markers and remove render ([052ed05](https://github.com/urmzd/artifact-protocol/commit/052ed05a7857405ea766638a5e6acae7f5066e85))
- **spec**: remove rendering layer and update markers section ([c6105d7](https://github.com/urmzd/artifact-protocol/commit/c6105d797d1b9e7164f18a6ebd26461ef8e65516))
- update source modules for markers and envelope handling ([62b6297](https://github.com/urmzd/artifact-protocol/commit/62b62972025802454943c6bbf26a8c603daa118f))
- **main**: replace render thread with envelope resolution ([f2d0d10](https://github.com/urmzd/artifact-protocol/commit/f2d0d10c7eda1fa7af875aa70297e8826d4634f2))
- **apply**: integrate markers and add JSON Pointer support ([b8a781f](https://github.com/urmzd/artifact-protocol/commit/b8a781fc1d36553db5c568bb8d5e2bece27fead4))
- remove PDF rendering dependencies and PDF renderer ([ddae166](https://github.com/urmzd/artifact-protocol/commit/ddae166d53d86092fe7ef364b8480db104241ce9))

### Miscellaneous

- add git hooks configuration ([9be8c37](https://github.com/urmzd/artifact-protocol/commit/9be8c3799afb708aeea829b1df2ed8c4f7e14194))
- update CI/CD and justfile for envelope resolution ([97172e3](https://github.com/urmzd/artifact-protocol/commit/97172e31d91737b0d44d777030368bad85bef91e))

[Full Changelog](https://github.com/urmzd/artifact-protocol/compare/v0.2.0...v0.3.0)


## 0.2.0 (2026-03-30)

### Features

- add rendering layer, entity state, and SSE transport binding to AAP spec ([a23b6ce](https://github.com/urmzd/artifact-protocol/commit/a23b6ceeb3ea49c7a625385205e4ad079a3c6545))
- rename Artifact Protocol to Agent-Artifact Protocol (AAP) ([fb7824e](https://github.com/urmzd/artifact-protocol/commit/fb7824ec4acf53763920fb54f25a47485642b2c3))

### Refactoring

- rename project from artifact-generator to aap ([238e1d4](https://github.com/urmzd/artifact-protocol/commit/238e1d45a40425a99ed0a9cca02513e3f498dc15))

### Miscellaneous

- update sr action from v2 to v3 ([92ebb82](https://github.com/urmzd/artifact-protocol/commit/92ebb82ad121fe9b03e26dd984d049cdb5f3f257))

[Full Changelog](https://github.com/urmzd/artifact-protocol/compare/v0.1.0...v0.2.0)


## 0.1.0 (2026-03-29)

### Features

- add OTEL tracing and metrics; rename python/ to tools/ ([4341bf3](https://github.com/urmzd/artifact-generator/commit/4341bf38d12f501b080499823f07308e330e6407))
- replace axum web server with headless Chrome PDF renderer ([20027cd](https://github.com/urmzd/artifact-generator/commit/20027cd6c720e4f84badecdbea67eb735ccdbf1b))

### Bug Fixes

- use Ubuntu 24.04 compatible package names for Chrome deps ([26ac3b1](https://github.com/urmzd/artifact-generator/commit/26ac3b13963915f104a0033ed654b503dae41df5))
- install Chrome shared libs and poll for PDF instead of fixed sleep ([bdb9c33](https://github.com/urmzd/artifact-generator/commit/bdb9c33c8b2e2165e33dff995c97947da3b18243))
- disable Chrome sandbox in CI and install chromium for PDF test ([e76497f](https://github.com/urmzd/artifact-generator/commit/e76497ffd0e77af33d5811ac3396d734519055c8))

### Documentation

- add agent skill following agentskills.io spec ([ba278bd](https://github.com/urmzd/artifact-generator/commit/ba278bde85fd412e83f47c201ac2d8790716e5a7))
- update README for headless Chrome PDF workflow ([13134d9](https://github.com/urmzd/artifact-generator/commit/13134d91a6f689e7d0ac6647c52114788e9e95b0))
- update repo URL to https://github.com/urmzd/artifact-generator ([d88cacd](https://github.com/urmzd/artifact-generator/commit/d88cacd4238ee65303a91f26c1c933fbfb0a4a5b))

### Refactoring

- restructure Python scripts into proper package ([157b31e](https://github.com/urmzd/artifact-generator/commit/157b31ef72de0d3d97ec500a9c385a45b194c9f6))

### Miscellaneous

- standardize CI/CD — add sr.yaml, workflow_call trigger, release workflow ([cd924d7](https://github.com/urmzd/artifact-generator/commit/cd924d7a4b7288687ba4e77f0c43176b80ac1fdb))
- update justfile and CI for PDF-based workflow ([be91725](https://github.com/urmzd/artifact-generator/commit/be917256ae1e4872a43cbbd421cd4c6491598d42))
- add GitHub Actions workflow (Rust build/test + Python benchmarks) ([22ff1b9](https://github.com/urmzd/artifact-generator/commit/22ff1b9486b3358d1834f07eae5310324113c75b))
