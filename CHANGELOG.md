# Changelog

## 0.5.0 (2026-04-03)

### Features

- **cli**: add provider/fallback options to LLM commands ([72954ae](https://github.com/urmzd/agent-artifact-protocol/commit/72954ae51beb8fb59618e7b7de4592e81905bf6d))
- **agents**: add fallback LLM model support ([b9f7776](https://github.com/urmzd/agent-artifact-protocol/commit/b9f7776bc66fb793357b0f94dd373d2e73392d13))

### Miscellaneous

- **evals**: add HTML dashboard ecommerce evaluation output ([125bf00](https://github.com/urmzd/agent-artifact-protocol/commit/125bf005b97993286bee141faa4474537a74c986))

[Full Changelog](https://github.com/urmzd/agent-artifact-protocol/compare/v0.4.1...v0.5.0)


## 0.4.1 (2026-04-03)

### Documentation

- remove benchmark generation instruction from README ([a060430](https://github.com/urmzd/agent-artifact-protocol/commit/a06043094cc8ff4a485684c283a6327de1c99bb5))
- **spec**: fix cross-reference link in error recovery section ([08ec641](https://github.com/urmzd/agent-artifact-protocol/commit/08ec6414880b501aafb9b0f471adcf66192bac27))
- **readme**: restructure documentation around apply engine library ([64a0578](https://github.com/urmzd/agent-artifact-protocol/commit/64a0578669c863488d3c8633de828a0a98c45be6))
- **CONTRIBUTING**: simplify contribution guidelines ([bf503ac](https://github.com/urmzd/agent-artifact-protocol/commit/bf503ac14fa6da8f5d010f48dafff2ac6918b521))

### Refactoring

- **evals/cli**: remove experiment command and add provider option ([764d898](https://github.com/urmzd/agent-artifact-protocol/commit/764d898275447fe8cffadf206769beeffadf8865))
- **evals/agents**: simplify to model factory and artifact generation ([2acabbd](https://github.com/urmzd/agent-artifact-protocol/commit/2acabbd1c12c76da9ef56f183e40460bc76ef29e))

### Miscellaneous

- add eval dataset for apply-engine operations ([44e57f4](https://github.com/urmzd/agent-artifact-protocol/commit/44e57f424181108ce1b9a902f37d2eb3f125814d))
- update project metadata for agent-artifact-protocol ([d5c84cb](https://github.com/urmzd/agent-artifact-protocol/commit/d5c84cb92b61189343cf7bc28b7a95172e63fd4a))
- **evals**: add evaluation test cases for apply engine ([4ae007a](https://github.com/urmzd/agent-artifact-protocol/commit/4ae007a2ff4c1b44bba8cfd2b4bc2677568b1f6d))
- regenerate binary and dependency lock files ([feccecd](https://github.com/urmzd/agent-artifact-protocol/commit/feccecd1244ade1432ab608bc5ab07fe69a828bf))
- **justfile**: update recipes for experiment removal and provider option ([473ab22](https://github.com/urmzd/agent-artifact-protocol/commit/473ab22d13749fe6ca25d8c7ecbadd4a1f6bf6a1))
- **evals**: add google provider support to pydantic-ai-slim ([3695121](https://github.com/urmzd/agent-artifact-protocol/commit/369512159ea3b6650d9a5275215b47b5e44218b0))
- **experiments**: add html-dashboard-ecommerce output turn 2 ([ea95050](https://github.com/urmzd/agent-artifact-protocol/commit/ea95050bebff2db7f68304a2748f8b1b702face0))
- **apply-engine**: add test case 0020 (energy consumption dashboard) ([1dadf26](https://github.com/urmzd/agent-artifact-protocol/commit/1dadf2679eb72bdacbc48c27a6c358bfbfbf70e1))
- **apply-engine**: add test case 0019 (server monitoring dashboard) ([5fcf2d5](https://github.com/urmzd/agent-artifact-protocol/commit/5fcf2d5cd8027ba2c095a1e378228cdf000c06ec))
- **evals**: add experiment 001 HTML dashboard ecommerce ([1d7becd](https://github.com/urmzd/agent-artifact-protocol/commit/1d7becd3d17a67e8e4e6cc3984fa613a84f29058))
- **evals**: update case 0018 to social media analytics ([20d7a34](https://github.com/urmzd/agent-artifact-protocol/commit/20d7a344f1201d42d15cf6ad53cff352d73c3655))

[Full Changelog](https://github.com/urmzd/agent-artifact-protocol/compare/v0.4.0...v0.4.1)


## 0.4.0 (2026-04-03)

### Breaking Changes

- **markers**: use universal XML markers for all text formats ([bfe24e4](https://github.com/urmzd/agent-artifact-protocol/commit/bfe24e4b826bafbe754f5dedda919dfecb9c447d))

### Features

- **cli**: add conversation benchmark experiment runner ([22e012d](https://github.com/urmzd/agent-artifact-protocol/commit/22e012d87cdbff64ae4a3fd26a94f536e6a80af1))
- **store**: add versioned artifact store with control-plane envelopes ([ec2878f](https://github.com/urmzd/agent-artifact-protocol/commit/ec2878f3c94042ae8fc27962eb6763a1fc9a1f92))
- **aap**: add envelope types for control-plane operations ([830d70e](https://github.com/urmzd/agent-artifact-protocol/commit/830d70e00772113385dacf2f417dc908960cc04a))
- **spec**: standardize section markers to universal XML format ([c93bd0c](https://github.com/urmzd/agent-artifact-protocol/commit/c93bd0c1377462e1ea10d0becd4f224bda99b560))
- add deterministic envelope generation ([0f15547](https://github.com/urmzd/agent-artifact-protocol/commit/0f155470624466f915266b1325e6787db70eb7fd))
- more benchmarks ([2dbd24d](https://github.com/urmzd/agent-artifact-protocol/commit/2dbd24d56fbb9c0af7e19e6ff3b2047594b04b05))
- **tools**: add Python utilities for prompt management ([95c7013](https://github.com/urmzd/agent-artifact-protocol/commit/95c7013d1ccecdbf3825bd56f3d5d4f8f134f3a0))
- **skills**: add generate-artifact skill definition ([337a2eb](https://github.com/urmzd/agent-artifact-protocol/commit/337a2eb132cfb7604b03dc33ec643b38a84673db))
- **benches**: implement envelope apply engine in Go ([09d047d](https://github.com/urmzd/agent-artifact-protocol/commit/09d047d578b6c0bc16643301d1567c65b6edc1f8))

### Bug Fixes

- **main**: update logging field from mode to name ([5fe69f0](https://github.com/urmzd/agent-artifact-protocol/commit/5fe69f07afc86491a1c1f51e0d4918635bd1ec88))

### Documentation

- **project**: update structure to reflect tools → evals reorganization ([ed7a164](https://github.com/urmzd/agent-artifact-protocol/commit/ed7a164aa4a9bedb4da0c70ca81af8f6b5dda7f9))
- deprecate tools package in favor of evals framework ([62e03f0](https://github.com/urmzd/agent-artifact-protocol/commit/62e03f0cfcd7c9bff8a1a6055db82eea8248c2fc))
- **spec**: clarify cost model variable notation in aap ([1fd01f6](https://github.com/urmzd/agent-artifact-protocol/commit/1fd01f635097243bb7281478c7f6ee29bc5eb0db))
- **spec**: clarify cost model and output token reduction ([09b2af5](https://github.com/urmzd/agent-artifact-protocol/commit/09b2af58a8b4966103593e6fde982cb51f8dd76a))
- **readme**: expand cost model section with detailed analysis ([07b16f8](https://github.com/urmzd/agent-artifact-protocol/commit/07b16f80f88787614447e5a46ee70f369bbfa5f7))
- **evals**: update maintain context terminology across all experiments ([89d5e27](https://github.com/urmzd/agent-artifact-protocol/commit/89d5e27fcce7f26002cf8d98ae094abd41c8a175))
- **evals**: add framework documentation and benchmarks ([7f3944b](https://github.com/urmzd/agent-artifact-protocol/commit/7f3944bbdb9ba8c998c50e523b80f5737d298542))
- **spec**: refactor context terminology and cost model ([732ea33](https://github.com/urmzd/agent-artifact-protocol/commit/732ea33b6325546d69eca07de2dca5fe5c2fb82e))
- **evals**: add AAP section marker instructions to turn-0 prompts ([7be70e7](https://github.com/urmzd/agent-artifact-protocol/commit/7be70e74ef4e41e6ad66b9923a05c964dd21d73c))
- **evals**: simplify AAP maintain-agent instructions ([834ae1b](https://github.com/urmzd/agent-artifact-protocol/commit/834ae1bc88b41688f1239b3f9f0aed877b74626f))
- **evals**: update AAP section marker syntax in init-system prompts ([7e0ef79](https://github.com/urmzd/agent-artifact-protocol/commit/7e0ef791ed8aa5956ff752510468f9f478ff8af1))
- **evals**: update experiment documentation with new overhead metrics ([0c00fb7](https://github.com/urmzd/agent-artifact-protocol/commit/0c00fb7ea52c13b3ec6e937d153d0b56f8242e66))
- **evals**: update all prompts to use universal aap section markers ([9f8dfb1](https://github.com/urmzd/agent-artifact-protocol/commit/9f8dfb1f3f233701ab506d737a18be900c5cca60))
- **readme**: add v0 stability warning ([cef8a75](https://github.com/urmzd/agent-artifact-protocol/commit/cef8a75bb8a8c172464dffb37e5bbb339ef71b95))
- **spec**: downgrade AAP protocol to v0.1 ([7bfa484](https://github.com/urmzd/agent-artifact-protocol/commit/7bfa4845b13336dbd5d37c0679dcb7480ae7ce7a))
- **spec**: update AAP specification to v2.0.0-draft ([85de1cc](https://github.com/urmzd/agent-artifact-protocol/commit/85de1cc6ec8c0d51b5431a3191856bed34e01c5a))
- **benches**: update protocol benchmark results ([55bd8fb](https://github.com/urmzd/agent-artifact-protocol/commit/55bd8fb3a89bb96dc9fe68de9cc6ed46033be2d5))

### Refactoring

- **evals**: clarify context terminology in evaluation framework ([83028d5](https://github.com/urmzd/agent-artifact-protocol/commit/83028d5e9e44f8aff4fe8f1f6f63b896e9e64a60))
- **evals**: restructure evaluation framework ([b2962dd](https://github.com/urmzd/agent-artifact-protocol/commit/b2962dd9a08cad82632097a586b3fa67221901e6))
- **ffi**: update Python bindings for new apply signature ([e36ccfa](https://github.com/urmzd/agent-artifact-protocol/commit/e36ccfa322560f4e569ccf2a8dd52e524e17d5c9))
- **apply**: redesign as stateless function ([13d86d0](https://github.com/urmzd/agent-artifact-protocol/commit/13d86d09141a58a0e050b2024b5581092f1bbc15))
- **evals**: remove outdated marker instructions from inputs ([c38a853](https://github.com/urmzd/agent-artifact-protocol/commit/c38a85349aae6e520f6897cb5ce98c9a83e67ba8))
- **build**: replace legacy benchmark script with new CLI ([9bbeec1](https://github.com/urmzd/agent-artifact-protocol/commit/9bbeec1d5a9531ef22195bda72864cede60e8808))
- **evals**: update apply module for marker system ([13f16cc](https://github.com/urmzd/agent-artifact-protocol/commit/13f16cc4558a20da800983554caf0e1e90a88d98))
- **evals**: update CLI for deterministic corpus generation ([234a6ca](https://github.com/urmzd/agent-artifact-protocol/commit/234a6ca47964c53fcbff8bac6bcfedbda64f02cb))
- **benches**: migrate to real fixtures and universal markers ([64c08b1](https://github.com/urmzd/agent-artifact-protocol/commit/64c08b10ca867dac6015a2539f0725894c46d75b))
- **evals**: migrate from Go benchmarks to Python CLI ([05ba8d9](https://github.com/urmzd/agent-artifact-protocol/commit/05ba8d99ad676d9bd273a4d28d1f83bc98f458ea))
- **ollama**: replace async aiohttp with pydantic-ai ([3d012f5](https://github.com/urmzd/agent-artifact-protocol/commit/3d012f5c997170400ce005715b5d7d56b6d16f75))
- replace Go benchmarks with Python eval framework ([c3c0283](https://github.com/urmzd/agent-artifact-protocol/commit/c3c02834b406d04a32cbf0adf5a38da433e2960b))
- **benches**: improve provider error handling ([41d0428](https://github.com/urmzd/agent-artifact-protocol/commit/41d0428c3fbd38163a3852229a33cb6f540d48d4))
- **core**: restructure envelope data model for v0.1 ([0d51ba2](https://github.com/urmzd/agent-artifact-protocol/commit/0d51ba228ca1711e1307ee0c2f255a583fe4247d))
- **tools**: remove legacy aap implementation ([66fe8a8](https://github.com/urmzd/agent-artifact-protocol/commit/66fe8a813881f9c01c092d3d458ab6a0029bb513))

### Miscellaneous

- **data**: add html dashboard ecommerce experiment outputs ([585e265](https://github.com/urmzd/agent-artifact-protocol/commit/585e26533e8ffb9941a5726a61da6543eb3c0e4c))
- **data**: add apply-engine conversation benchmark corpus (0018) ([786c51b](https://github.com/urmzd/agent-artifact-protocol/commit/786c51b114bfb2c3133adf56a1a5ea50bf6a76f9))
- **data**: update apply-engine conversation benchmark corpus (0017) ([5672382](https://github.com/urmzd/agent-artifact-protocol/commit/5672382c95c467a756d3886c6c2cb75432b139c8))
- **data**: add apply-engine conversation benchmark corpus (0017) ([7d2a380](https://github.com/urmzd/agent-artifact-protocol/commit/7d2a380ca282fe7fb7df28b2632eb78abbad436b))
- **justfile**: add run command for conversation benchmarks ([85ee166](https://github.com/urmzd/agent-artifact-protocol/commit/85ee166a3b8d18344a24ebac171ac07943962f85))
- **evals**: add apply-engine test case 0016 ([34162a7](https://github.com/urmzd/agent-artifact-protocol/commit/34162a753ff14b6b73290a4f0a02b7945697cf38))
- **evals**: add apply-engine test case 0015 ([91ab6a3](https://github.com/urmzd/agent-artifact-protocol/commit/91ab6a30f51e4990e6681c084c6cf3a1c0a10b96))
- **evals**: add apply-engine test case 0014 ([77c25ed](https://github.com/urmzd/agent-artifact-protocol/commit/77c25ed6d395e454a6407a3f1e137dcf6d4b1c5a))
- **evals**: add apply-engine case 0013 (html-dashboard) ([6ea1720](https://github.com/urmzd/agent-artifact-protocol/commit/6ea172084913a1f1fe147f00a2e13d9173581d85))
- **evals**: add apply-engine test cases 0002-0012 ([55e8037](https://github.com/urmzd/agent-artifact-protocol/commit/55e80378d056c50e33055e8758338b651a20591d))
- **deps**: add sha2 for artifact checksums ([b0b4f14](https://github.com/urmzd/agent-artifact-protocol/commit/b0b4f14c703190ab4c73b124877e3b03c8998b0e))
- **evals**: remove old apply-engine evaluation artifacts ([e3344d4](https://github.com/urmzd/agent-artifact-protocol/commit/e3344d49374398d769f6daea6367942219ed6d1a))
- **evals**: add apply-engine benchmark corpus fixtures ([fc8951f](https://github.com/urmzd/agent-artifact-protocol/commit/fc8951f7ef403d0af54ef4346a24cc0fb18b3e88))
- **apply**: update test fixtures to universal XML markers ([65c692a](https://github.com/urmzd/agent-artifact-protocol/commit/65c692afa4485f95a4f0e9147799e6559e9ca5a8))
- **evals**: compile extension and update dependencies ([d303b09](https://github.com/urmzd/agent-artifact-protocol/commit/d303b098f6967f73c187a3fa461a8cf3130a5a9e))
- remove deprecated AAP skill definitions ([a8da8f4](https://github.com/urmzd/agent-artifact-protocol/commit/a8da8f4cd73242b997c835655cc81304f98c8202))
- **benches**: add AAP protocol benchmark dataset ([c6db0f0](https://github.com/urmzd/agent-artifact-protocol/commit/c6db0f06d07de7a78d5a30b06452563ff7f25970))
- **benches**: add AAP protocol benchmark infrastructure ([f9dbf93](https://github.com/urmzd/agent-artifact-protocol/commit/f9dbf93ab54e0ddc4ebd2667ab6fa064d0ecda45))
- **benches**: add comprehensive benchmark dataset (88 experiments) ([96ead37](https://github.com/urmzd/agent-artifact-protocol/commit/96ead3733124836d9c2e347909ed901b759f7b65))
- **justfile**: simplify development recipes ([1898c00](https://github.com/urmzd/agent-artifact-protocol/commit/1898c001890e31f89fa466a8219bf1f4fcad2977))
- **gitignore**: exclude generated asset inputs ([574cc0f](https://github.com/urmzd/agent-artifact-protocol/commit/574cc0f2adba7485afc1889afb1f14defc2c3f46))
- **cargo**: add package metadata for crates.io ([285e4e7](https://github.com/urmzd/agent-artifact-protocol/commit/285e4e7025e4bd508cd3a3d7c8c76a07c8f9e5c7))
- **license**: add dual licensing structure ([e4915d3](https://github.com/urmzd/agent-artifact-protocol/commit/e4915d35a37f6eb20feabe34d1c31fe37ea8b200))
- **release**: add automated crates.io publish workflow ([45d720d](https://github.com/urmzd/agent-artifact-protocol/commit/45d720db418792c0110b6344a2dac969a6df2c8a))
- **python**: rename package from artifact-generator to aap ([677e7af](https://github.com/urmzd/agent-artifact-protocol/commit/677e7aff22f78fea14d1cda8adad8c81db37f6d5))
- **justfile**: add bench-all convenience recipe ([911ea1b](https://github.com/urmzd/agent-artifact-protocol/commit/911ea1be4951aa4200c97af5c1f601f61ac7cf55))

[Full Changelog](https://github.com/urmzd/agent-artifact-protocol/compare/v0.3.0...v0.4.0)


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
