# Product backlog

Updated: 2026-10-31. This page owns product outcomes and dependencies. Current
provider integration, CI and laboratory maintenance belong to the
[engineering backlog](../engineering/backlog.md).

| ID | Outcome | Current status and completion criterion |
| --- | --- | --- |
| STUB-01 | Identify a reproducible US8/Linux candidate. | The complete project gate and Linux package build pass on `main`. Bind the next controller-test binary to its exact source and inputs; an earlier candidate does not qualify a new binary. |
| STUB-02 | Maintain the controller-backed laboratory route. | Modernization in progress. Core owns product-test orchestration; Labs owns controller preparation and authoritative machine data. Use assigned machines and current native interfaces. See the [controller lab](docker-lab.md). |
| STUB-03 | Prove safety boundaries and metric failure behavior. | Open. Verify bounded adoption persistence, read-only observation, missing sources, link loss, first samples and counter resets. Controller restart, upgrade, shell and provisioning commands must not mutate the host. |
| STUB-04 | Accept the same US8 candidate on current Stable and RC. | After STUB-01–03. Resolve both channels at run start; prove adoption, connected state and visible real speed and bidirectional traffic from the named source. API values alone do not establish visible acceptance. |
| STUB-05 | Accept installation on a supported Linux host. | After STUB-03–04. CI builds all four Linux formats and passes a real Debian install/version/configuration/service-file smoke test. Product acceptance still needs the installed service with a named bridge and real uplink. |
| STUB-06 | Complete the English and German operator path. | After STUB-05. Validate installation, configuration, adoption, visible metrics, limitations and recovery against the accepted release. |
| STUB-07 | Deliver verified Linux releases. | Main CI, package build and Debian installation smoke pass. Publication requires the accepted product and matching source/binary evidence. |
| STUB-08 | Add another representation for a concrete need. | Deferred until Linux acceptance. Name its operator outcome, read-only source, safety boundary and independent Stable/RC criteria. FreeBSD/OPNsense remains a separate optional target. |

The next product proof is STUB-03 followed by combined Stable/RC acceptance.
Historical Stable 11.0.97 laboratory results concern their recorded candidate
and controller instance; RC remains unqualified. They do not establish current
combined acceptance. The authoritative contract and durable-result requirements
are in [requirements](requirements.md) and [compatibility](compatibility.md).

Use current native BR/CT bindings for execution. Historical provider revisions,
retired snapshots and earlier runner readiness are not execution selectors.
