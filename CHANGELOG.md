# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

### Added

- Ensure PodSecurityPolicy for chartmuseum.
- Enable ServiceAccount creation for chartmuseum.
- Enable API of chartmuseum.
- Container build through CircleCI.
- NetworkPolicy to enable communication with chartmuseum.
- Patching dnsConfig and dnsPolicy of chart-operator-unique deployment.

### Fixed

- Use apiextensions v3 and include replace for Giant Swarm CAPI fork. 
- Add replace for moby v20.10.0-beta1 to fix build issue on darwin.

## [0.3.1] - 2020-10-08

### Added

- Enable persistence in chartmuseum via its app CR.

## [0.3.0] - 2020-10-06

### Added

- Add install operators flag.
- Create AppCatalogEntry CRD.

## [0.2.0] - 2020-10-01

### Added

- Use apptest library to create chartmuseum app CR.
- Update app-operator to v2.3.2.
- Update chart-operator to v2.3.3.

## [0.1.0] - 2020-09-18

### Added

- Add initial version that bootstraps app-operator, chart-operator and chartmuseum.

[Unreleased]: https://github.com/giantswarm/apptestctl/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/giantswarm/apptestctl/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/apptestctl/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/apptestctl/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/apptestctl/releases/tag/v0.1.0
