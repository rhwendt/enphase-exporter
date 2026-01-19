# Changelog

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
