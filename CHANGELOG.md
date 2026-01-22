# Changelog

## [1.2.0](https://github.com/rhwendt/enphase-exporter/compare/v1.1.0...v1.2.0) (2026-01-22)


### Features

* Add API call duration metrics and timing logs ([52dc896](https://github.com/rhwendt/enphase-exporter/commit/52dc896d1b590b74e4ac3fd9a1b5fadc08f7f369))
* Add API call duration metrics and timing logs ([#14](https://github.com/rhwendt/enphase-exporter/issues/14)) ([75bfd33](https://github.com/rhwendt/enphase-exporter/commit/75bfd3332380f3a662a9b584da1ace4c540a9341))

## [1.1.0](https://github.com/rhwendt/enphase-exporter/compare/v1.0.3...v1.1.0) (2026-01-21)


### Features

* Add energy export/import metrics ([#9](https://github.com/rhwendt/enphase-exporter/issues/9)) ([25a97f6](https://github.com/rhwendt/enphase-exporter/commit/25a97f6b812474f7b8fd3b62cdfe1d5f2cf7eaaa))

## [1.0.3](https://github.com/rhwendt/enphase-exporter/compare/v1.0.2...v1.0.3) (2026-01-19)


### Bug Fixes

* Use per-line values to fix split-phase consumption doubling ([ee84454](https://github.com/rhwendt/enphase-exporter/commit/ee844545f02b2bfe35a33638630c7657ff2a8f9f))

## [1.0.2](https://github.com/rhwendt/enphase-exporter/compare/v1.0.1...v1.0.2) (2026-01-19)


### Bug Fixes

* Add proactive session refresh to prevent data gaps ([#6](https://github.com/rhwendt/enphase-exporter/issues/6)) ([5281f5b](https://github.com/rhwendt/enphase-exporter/commit/5281f5b7cb1563de984ba2188615b0c0c6fd49c5))

## [1.0.1](https://github.com/rhwendt/enphase-exporter/compare/v1.0.0...v1.0.1) (2026-01-19)


### Bug Fixes

* Authenticate on startup to fix readiness probe deadlock ([#4](https://github.com/rhwendt/enphase-exporter/issues/4)) ([45d0d58](https://github.com/rhwendt/enphase-exporter/commit/45d0d58699c8b119dc2e657f91c7161bede6cfa1))
* Use TARGETARCH for proper cross-compilation in Docker ([#2](https://github.com/rhwendt/enphase-exporter/issues/2)) ([2aa26dd](https://github.com/rhwendt/enphase-exporter/commit/2aa26ddef77ac1f3353b34c727ffb64ae3d0a249))

## 1.0.0 (2026-01-18)


### Features

* Add authentication, API client, and K8s deployment ([b4c0b2f](https://github.com/rhwendt/enphase-exporter/commit/b4c0b2f5fdf8367e6f5cbded39862ef4ac7588e5))
* Add consumption metrics and improve documentation ([9b2ae3b](https://github.com/rhwendt/enphase-exporter/commit/9b2ae3bbccff4b2d088b30964f15abb0c92edab5))
* Initial project structure and scaffolding ([ad19615](https://github.com/rhwendt/enphase-exporter/commit/ad1961572aad8be788f7ec63be58ce90492103de))
