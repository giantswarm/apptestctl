# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.22.1] - 2024-10-28

### Changed

- Update dependencies.

## [0.22.0] - 2024-10-24

### Added

- Add `ClusterPolicies` CRDs.

### Changed

- Run `sync-crds` script.

## [0.21.0] - 2024-09-20

### Added

- Add `PrometheusRules` CRDs.

## [0.20.0] - 2024-09-03

### Added

- Add `Prometheuses` CRDs.
- Add `RemoteWrites` CRDs.

## [0.19.1] - 2024-06-06

- Dependency upgrades

## [0.19.0] - 2024-05-14

### Changed

- Synchronize CRDs.

## [0.18.1] - 2024-05-08

### Changed

- Promote Kyverno `PolicyException` CRD to `v2beta1`.

## [0.18.0] - 2023-10-09

### Added

- Install `VerticalPodAutoscalers` CRDs.
- Add `PolicyException` CRDs.

### Changed

- AppCatalogEntries: Rename `nginx-ingress-controller-app` to `ingress-nginx`. ([#320](https://github.com/giantswarm/apptestctl/pull/320))

## [0.17.0] - 2023-06-07

### Added

- Install `ServiceMonitors` and `PodMonitors` CRDs
  as part of test bootstrap.

## [0.16.0] - 2023-05-11

### Changed

- Update base container image to alpine:3.18.0.
- Added support for kubernetes 1.25
- Upgrade app-operator to version 6.7.0
- Upgrade chart-operator to version 2.35.0
- Upgrade chartmuseum chart to version 3.9.3

## [0.15.0] - 2023-04-28

### Added

- Install `CiliumNetworkPolicies` and `CiliumClusterwideNetworkPolicies` CRDs
  as part of test bootstrap.

### Changed

- Update App and Chart CRDs.

## [0.14.5] - 2022-07-28

### fixed

- Catalog CR is fixed to include "Repositories" section

## [0.14.4] - 2022-07-07

## [0.14.3] - 2022-07-01

- Refresh App CRD to the latest version

## [0.14.2] - 2022-06-30

### Changed

- Update application.giantswarm.io/v1alpha1 Catalog CRD to the latest

## [0.14.1] - 2022-03-02

### Fixed

- Use latest AppCatalogEntry CRD with hardcoded providers removed.

## [0.14.0] - 2022-01-05

### Changed

- Update app-operator to v5.4.0.
- Update chart-operator to v2.20.0 with Helm v3.6.3.

## [0.13.1] - 2021-12-15

### Fixed

- Fix behaviour of `--install-operators` flag.

## [0.13.0] - 2021-12-15

### Added

- Add `log-level` flag and default to `error` level.

### Changed

- Use apiextensions-application to remove CAPI dependency.
- Upgrade helm to v3.6.3.
- Sync CRDs with `apiextensions-application`.

### Fixed

- Fix `--install-operators` flag so it doesn't wait for `chartmuseum`.

## [0.12.0] - 2021-10-04

### Changed

- Embed `application.giantswarm.io` CRDs to avoid hitting GitHub API rate limits.

## [0.11.1] - 2021-09-28

### Fixed

- Set `isManagementCluster` to true so we `ClusterFirst` for chart-operator
DNS settings.

## [0.11.0] - 2021-09-28

### Changed

- Update app-operator to v5.2.0.
- Update chart-operator to v2.19.0.

## [0.10.0] - 2021-08-09

### Added

- Add `--github-token` as a flag.

### Changed

- Update app-operator to v5.1.1.
- Update chart-operator to v2.18.1.

## [0.9.0] - 2021-06-14

### Changed

- Ensuring `Catalog` CRD when bootstrapping the cluster.

## [0.8.0] - 2021-04-06

### Changed

- Update app-operator to v4.3.2 for Kubernetes 1.20 support.
- Update chart-operator to v2.13.1 for Kubernetes 1.20 support.

## [0.7.0] - 2021-02-24

### Changed

- Update app-operator to v4.0.0.
- Update chart-operator to v2.9.0.

## [0.6.1] - 2021-01-11

### Fixed

- Wait for ready chartmuseum pod and configure readiness probe.

## [0.6.0] - 2021-01-07

### Added

- Pretty print errors.
- Update app-operator to v3.0.0.

### Changed

- Remove helmclient.MergeValue functions usage.

### Fixed

- Use shorter resync period and reduce finalizer TTL in app-operator so
resources are deleted on subsequent test runs.

## [0.5.2] - 2020-12-01

### Changed

- Chart-museum is now deployed with "allow-overwrite" option, so the same chart may be uploaded multiple times.

### Fixed

- "Version" in `version` command is printed correctly now

## [0.5.1] - 2020-11-30

### Added

- Update apiextensions to v3.9.0 to add printer columns for app and chart CRDs.

## [0.5.0] - 2020-11-27

### Added

- Add kubeconfig-path flag and support for KUBECONFIG env var.
- Add wait flag for whether to wait for all components to be ready.

## [0.4.1] - 2020-11-03

### Added

- Patch dnsConfig and dnsPolicy of chart-operator-unique deployment.

## [0.4.0] - 2020-10-30

### Added

- Ensure PodSecurityPolicy for chartmuseum.
- Enable ServiceAccount creation for chartmuseum.
- Enable API of chartmuseum.
- Container build through CircleCI.
- NetworkPolicy to enable communication with chartmuseum.

### Fixed

- Use apiextensions v3 and include replace for Giant Swarm CAPI fork.
- Add replace for moby v20.10.0-beta1 to fix build issue on darwin.
- Optimize apps wait interval as app-operator has a status webhook.
- Use new catalog URL for Helm Stable.

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

[Unreleased]: https://github.com/giantswarm/apptestctl/compare/v0.22.1...HEAD
[0.22.1]: https://github.com/giantswarm/apptestctl/compare/v0.22.0...v0.22.1
[0.22.0]: https://github.com/giantswarm/apptestctl/compare/v0.21.0...v0.22.0
[0.21.0]: https://github.com/giantswarm/apptestctl/compare/v0.20.0...v0.21.0
[0.20.0]: https://github.com/giantswarm/apptestctl/compare/v0.19.1...v0.20.0
[0.19.1]: https://github.com/giantswarm/apptestctl/compare/v0.19.0...v0.19.1
[0.19.0]: https://github.com/giantswarm/apptestctl/compare/v0.18.1...v0.19.0
[0.18.1]: https://github.com/giantswarm/apptestctl/compare/v0.18.0...v0.18.1
[0.18.0]: https://github.com/giantswarm/apptestctl/compare/v0.17.0...v0.18.0
[0.17.0]: https://github.com/giantswarm/apptestctl/compare/v0.16.0...v0.17.0
[0.16.0]: https://github.com/giantswarm/apptestctl/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/giantswarm/apptestctl/compare/v0.14.5...v0.15.0
[0.14.5]: https://github.com/giantswarm/apptestctl/compare/v0.14.4...v0.14.5
[0.14.4]: https://github.com/giantswarm/apptestctl/compare/v0.14.3...v0.14.4
[0.14.3]: https://github.com/giantswarm/apptestctl/compare/v0.14.2...v0.14.3
[0.14.2]: https://github.com/giantswarm/apptestctl/compare/v0.14.1...v0.14.2
[0.14.1]: https://github.com/giantswarm/apptestctl/compare/v0.14.0...v0.14.1
[0.14.0]: https://github.com/giantswarm/apptestctl/compare/v0.13.1...v0.14.0
[0.13.1]: https://github.com/giantswarm/apptestctl/compare/v0.13.0...v0.13.1
[0.13.0]: https://github.com/giantswarm/apptestctl/compare/v0.12.0...v0.13.0
[0.12.0]: https://github.com/giantswarm/apptestctl/compare/v0.11.1...v0.12.0
[0.11.1]: https://github.com/giantswarm/apptestctl/compare/v0.11.0...v0.11.1
[0.11.0]: https://github.com/giantswarm/apptestctl/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/giantswarm/apptestctl/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/giantswarm/apptestctl/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/giantswarm/apptestctl/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/giantswarm/apptestctl/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/giantswarm/apptestctl/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/giantswarm/apptestctl/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/giantswarm/apptestctl/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/giantswarm/apptestctl/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/giantswarm/apptestctl/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/giantswarm/apptestctl/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/giantswarm/apptestctl/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/giantswarm/apptestctl/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/apptestctl/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/apptestctl/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/apptestctl/releases/tag/v0.1.0
