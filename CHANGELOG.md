# Changelog

## [1.0.2](https://github.com/apideck-libraries/cli/compare/v1.0.1...v1.0.2) (2026-03-13)


### Bug Fixes

* restore permissions block in release-please workflow ([50d1ac1](https://github.com/apideck-libraries/cli/commit/50d1ac1350952d1923b422700e275b4fee556c50))
* use HOMEBREW_TAP_TOKEN instead of separate release token ([4db38f6](https://github.com/apideck-libraries/cli/commit/4db38f62579c4a37e8bad6a62afbba244ce46a10))
* use PAT in Release Please to trigger release workflow ([8b05152](https://github.com/apideck-libraries/cli/commit/8b05152764a1124037bfd6d263faf55b28dd1668))

## [1.0.1](https://github.com/apideck-libraries/cli/compare/v1.0.0...v1.0.1) (2026-03-13)


### Bug Fixes

* add token to GoReleaser brew config for cross-repo push ([30f9445](https://github.com/apideck-libraries/cli/commit/30f9445cfb83b5354d3f971db6facf53500bb0e5))
* add token to GoReleaser brew config for cross-repo push ([a8f8a0f](https://github.com/apideck-libraries/cli/commit/a8f8a0f93b0918312f18571e7b370d9a9c6939f2))

## 1.0.0 (2026-03-13)


### Features

* Add blog post ([59d64da](https://github.com/apideck-libraries/cli/commit/59d64da9298c1f76ea50fd9c30a3b067b2e47bf8))
* add Dockerfile with distroless base image ([51c7e7f](https://github.com/apideck-libraries/cli/commit/51c7e7fe0181ee287811d58b43bb39d77a9739d3))
* add goreleaser config for multi-platform distribution ([8c9d893](https://github.com/apideck-libraries/cli/commit/8c9d893d25590e35c49de0ee0221449da3f71246))
* add history, permissions, and skill install commands ([765ba5e](https://github.com/apideck-libraries/cli/commit/765ba5ef7ca6f84de7241fe932fed579b7cfe700))
* add huh confirmation dialog and --data [@file](https://github.com/file).json support ([a454ed3](https://github.com/apideck-libraries/cli/commit/a454ed3ec1147f01c6c1b8652080f5ebcc53fb8b))
* add internal model types for API spec, operations, permissions ([b460237](https://github.com/apideck-libraries/cli/commit/b460237765a0266138c639811d969dbc497362fe))
* add shared UI styles, brand colors, and message helpers ([e95bb9e](https://github.com/apideck-libraries/cli/commit/e95bb9e39abc45696573024126ab93f973ee0e8d))
* add User-Agent header, filter global flags, and project setup ([447a352](https://github.com/apideck-libraries/cli/commit/447a352d7c09a41c238dec702dd547e225ae8e4f))
* auto-create default permissions config during setup ([1e12785](https://github.com/apideck-libraries/cli/commit/1e12785b695281b6e723458c7d3a9ba684c37e63))
* embed baseline Apideck OpenAPI spec ([1bdc0b0](https://github.com/apideck-libraries/cli/commit/1bdc0b000d0c715f46044660b76da41b36c4cfe2))
* implement auth manager with env var and config file resolution ([1eec7e0](https://github.com/apideck-libraries/cli/commit/1eec7e0f485da5ca5a211b297ff73d9139a62ec8))
* implement dynamic command router from OpenAPI spec ([613d4a9](https://github.com/apideck-libraries/cli/commit/613d4a9ee447f16097bf42d2b77c45e19b5842f9))
* implement HTTP client with retryablehttp and response normalization ([3c71576](https://github.com/apideck-libraries/cli/commit/3c715769af8b0f9f122acd29323fdaa5d38d7a2c))
* implement interactive auth setup wizard with huh ([0ec06c3](https://github.com/apideck-libraries/cli/commit/0ec06c3931fa0b1fd044d9687a15ec785ca854b5))
* implement OpenAPI parser with libopenapi ([6928adb](https://github.com/apideck-libraries/cli/commit/6928adb6df154a76c2c6c25ca9314570701ac8a1))
* implement output formatters (JSON, YAML, CSV, table) ([4c0251b](https://github.com/apideck-libraries/cli/commit/4c0251b50852116d7ca9c9ae5af5213457dbf76c))
* implement permission engine with override support ([faedc6e](https://github.com/apideck-libraries/cli/commit/faedc6ea489f8eb58bace19e73c1435d20f8a250))
* implement spec cache with gob encoding and atomic writes ([f7e2108](https://github.com/apideck-libraries/cli/commit/f7e21085e7cacba405898aa9bf938d87fb9fbf40))
* implement token-optimized agent prompt generation ([03b4d03](https://github.com/apideck-libraries/cli/commit/03b4d035f47be06550a3e09d2e30cb2041f6b96e))
* implement TUI explorer with two-panel layout ([f536847](https://github.com/apideck-libraries/cli/commit/f53684730baaf2ec432e1e4a47e6b90950da8645))
* initialize Go project with Cobra CLI skeleton ([5adbad0](https://github.com/apideck-libraries/cli/commit/5adbad020965d3904b029a4fe0878013f663ab40))
* wire history logging into API executor ([b7e3688](https://github.com/apideck-libraries/cli/commit/b7e36889f1cad3e0a3121fbc706e2a0dd3558c9d))
* wire up CLI with all components and static commands ([8ae26bc](https://github.com/apideck-libraries/cli/commit/8ae26bcef835419fe3389749b09798048fcc6dbf))


### Bug Fixes

* serialize maps and nil as JSON in table and CSV output ([60a4868](https://github.com/apideck-libraries/cli/commit/60a4868ece65de8c42d41869b396924f83c05742))
