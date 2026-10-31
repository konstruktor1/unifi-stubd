# Maintainer backlog

## Completed work

| ID | Work and intended use | Status | Responsible owner | Completion criterion / observed outcome | Remarks and dependencies | Use cases |
| --- | --- | --- | --- | --- | --- | --- |
| CI-001 | Complete current-provider delivery and verify the GitHub-hosted CI run. | Complete | CT source owner and repository delivery owner | CT004/CT006 is published as 5c325ff on its actual default branch; the scoped consumer integration and CI corrections are published on codex/consolidate-unifi-stubd. GitHub run 41163533722 on d3fbe4d passed full make verify, all 13 required CT checks with no unresolved checks, coverage, SBOM, artifact uploads and both post-actions. | Replaced the fixed ten-second discovery helper with typed SDK discovery allowing cold setup; unified native/controller UID only on disposable GitHub-hosted Ubuntu; handed the product artifact directory to native Actions. Private 0700 checks, current-source integrity and the FreeBSD-only runner role remain enforced. Package/install jobs are branch-condition skips; no manual FreeBSD/controller or main merge was run. Source-review diagnostics remain visible; no private review record is published. Evidence: ci-delivery-20261031. | GitHub-hosted branch quality and native container/artifact integration. |
| CI-002 | Implement the clarified GitHub/container CI route and preserve the FreeBSD-only runner. | Complete | unifi-stubd CI integration owner | General jobs select GitHub-hosted Ubuntu and BR containers. Maintained setup prepares current private sources and SDK; actual public preflight derives optional CT_SOURCE through the saved CT connection. Source read token and native variables are configured. Full make verify including workflow-lint and all 13 CT checks passed; real SDK/source-bootstrap and BR-only independence passed. | FreeBSD workflow and runner labels are unchanged; obsolete docker-engine scheduling/policy removed. Actual mismatched remote CT is rejected by bytes/modes, fixtures are removed and staging is preserved. Remote delivery/execution remains separately blocked by CI-001; no push or real controller/VM operation is implied. Evidence: ci-runner-docker-20261030. | Correct local CI definitions and verified native integration. |
| CT-006 | Propagate the actual task use case to auxiliary source-review due checks. | Complete | CT source owner and unifi-stubd receiving owner | Current optional use_case is propagated across eight CLI/MCP auxiliaries, MCP annotations and requested CT Doctor. Five focused CLI cases and the real ten-function MCP regression passed; supplied cli stays current, while default/explicit library remains due. Local explicit-cli persistence inquiry has no false review diagnostic and the actual review bytes/time remain unchanged. | Fresh acknowledged current context, exact foreign adoption and own guidance/notices preserve the input duty onward. The returned Anti-Bloat verify command also retains the validated literal selector. Final receiving review uses that unchanged current procedure; original wrong-context outcomes stay in ct-precision-cleanup-20261030. No preference, ENV selector, diagnostic suppression or second store was added. | Accurate invocation-based review during selected CLI task completion. |
| CT-004 | Correct source-review binding across the verified native/BR namespace of one consumer. | Complete | CT source owner and unifi-stubd native receiving owner | CT schema 2 uses the existing native review label solely as scope discriminator and hashes criterion content/style/use case without navigation locators. Fresh host and real BR source-review both report current/due=false with the same digest and actual original 56-source timestamp. | CT owner: make check-quality exit 0, 58 tests passed and two optional PowerShell skips; seven required checks, CLI/MCP and final review passed. Local exact semantic/URL equality, single-store rebind and original bytes are recorded in ct-precision-cleanup-20261030. No legacy reader, clock rewrite, extra Root or duplicated authority. | Selected BR review-state application and repeated onward guidance. |
| CT-005 | Remove proven obsolete integration artifacts and current contradictions after CT-004. | Complete | unifi-stubd native integration owner | Removed 213 files in six abandoned source-only/empty operation directories, the empty hook and 12 retired Git-exclude patterns; replaced old Docker excludes, npm metadata and preflight conversion. Docker context now excludes private/generated roles; CT image preparation runs only for the selected BR environment with bounded child shutdown. | make verify exit 0 with all 13 required CT checks and unresolved_checks=[]; real SDK resources/notices/filters/invalidation/preservation and TERM-ignoring child cancellation passed; ct-quality manifest valid. Staged entries, protected offline/profile data, settings and useful evidence preserved. Original abandoned domain outcomes remain unknown; none replayed. Final task review and role layout use ct-precision-cleanup-20261030. | Exact receiving completion and one-time legacy cleanup. |
| CT-001 | Adopt the current CT receiving integration and exact foreign guidance. | Complete | unifi-stubd CT integration owner | Seven foreign CT skills plus consumer integration, instructions and glossary match current descriptors byte-for-byte and mode-for-mode; the marked AGENTS block is exact. Own update/setup/onward guidance and callable notices are current; successful-call diagnostics remain visible. | Current contract and complete shared package/dependency were read through acknowledged subscriptions. Public init saved only the four confirmed missing preferences, preserving prior choices/comments. Native source/configuration owners remain separate; The namespace gap is resolved by CT-004. | Current local CT integration and recursive delivery method. |
| CT-002 | Supply and verify the confirmed BR execution environment for CT checks. | Complete | unifi-stubd native runtime/connection owner | The native ARM64 ct-quality image, read-only current CT source, same physical consumer mapped to /workspace and contained mutable paths passed actual validation. Fresh CT contract, plan and all 13 checks exited 0; make verify exited 0. | Make quality-ct-image owns only the documented unsupported consumer-image build step; public BR owns execution. The saved source reference is public-exporter-derived. Canonical product Go scope remains GO_PACKAGES; no source copy, wrapper, host-native quality fallback or foreign edit. | Go/JS/Bash/Python checks through the selected environment. |
| CT-003 | Execute the confirmed initial rule-source review and native saved-state readback. | Complete | unifi-stubd calling agent and native review writer | All 56 current URLs have actual dispositions: 55 reviewed and one C-specific reference excluded with a reason. The exact nine-field result was saved only after review and passing checks; fresh native source-review reports current and due=false for the selected seven-day cadence. | GNU direct retrieval failed but substantive indexed official content was reviewed with the limitation recorded. No timer/job was added. CT-004 subsequently verified BR readback of that same authoritative result while preserving its actual time and observations. | Current selected criteria and invocation-based review. |
| BR-005 | Adopt current BR provider-update, protocol/capability, storage-placement and own-notice duties. | Complete | unifi-stubd integration owner | Native BR/CT clients select the bound catalog protocol and verify SDK/server capabilities, exact catalog bytes/modes and active acknowledged subscriptions. Five own skills, canonical callable notices/filters, the update Doctor anchor and byte/mode invalidation/reconnect passed the real SDK verifier. make verify and all four core manifest validations exited 0. | Replaced fixed protocol literals and superseded own four-skill wording; current shared package/dependency and requirements read and reconciled. Both direct BR/CT bindings remain necessary. No obsolete peer tree, new placement or executable copy was introduced. Final task review and role-layout verification use private br-update-20261030 evidence; unrelated CI/CT work retains its owner. | Explicit local BR update and recursive receiving guidance. |
| HOST-001 | Restore dependable local Docker/storage operation for BR verification. | Complete | Authorized local Docker host owner | Docker Desktop restarted successfully and its live builder confirms explicit Max Used Space 20 GiB / Reserved Space 10 GiB. The cache inventory is about 20 GB; native Engine settings were preserved except the GC policy, with the original backed up in private task inputs. | The initial recovery to a healthy daemon and 23 GiB free preceded this continuation. The narrow old/private/regular prune reported 0 B itself; do not attribute the separate 46.7 GB to 20 GB inventory reduction solely to that command. No containers, images, volumes or VM disks were broadly deleted. | Authorized Docker/storage repair. |
| BR-003 | Complete full BR product verification, local manifest validation and backend cleanup. | Complete | unifi-stubd integration owner | make verify exited 0: JS format/lint, Go format/vet/lint, policy, configuration/profile validation, tests, govulncheck and all 7 native Bash/Python checks passed. All 4 core instance/testhost manifests validated with exit 0. | The earlier failed run was terminal and inactive; public consumer-scoped clean with warm caches preserved removed its remaining network. Final inventory has no task containers or BR networks. Tier3 readiness and controller lifecycle were not executed. Evidence: private task results/verify-recovery.* and recovery-manifest-*. | Complete local BR integration verification. |
| BR-004 | Finish scoped review and prescribed evidence layout. | Complete | unifi-stubd integration owner | Reviewed current authorized source changes and actual check results; maintained verifier is scripts/check-provider-integration.py, tests/.work is empty and task evidence is .consumer-data/evidence/unifi-stubd/br-integration-20261030/. | Initial failure evidence is retained; authorized recovery closes BR-003 with complete local verification and manifest validation. Remote CI and controller acceptance retain their separate owners. | Engineering review and preserved evidence. |
| BR-001 | Reconcile current BR context and affected native declarations. | Complete | unifi-stubd integration owner | Read 21 unique required BR resources with exact SHA-256/modes and acknowledged active catalog/annotations/list subscriptions; saved native export matches; current CT selection passed after removing the obsolete snapshot exclusion. | Local tests directory mode 0700 satisfies the current verification owner; no foreign materialization or incoming Git index was changed. | Current-file BR integration; portable native setup. |
| BR-002 | Implement own Doctor, current-context client and four own functional skills. | Complete | unifi-stubd integration owner | Real SDK discovery, four exact skills, notice filter, invalid inputs, Doctor inspect/verify, settings preservation and disposable byte/mode invalidation with sticky rejection/reconnect passed. | Native Bash/Python gate: seven required checks passed, unresolved_checks empty. Direct BR CLI retained; native special identity route is provider --json without project options. | Local diagnosis; native verification; requested onward guidance. |
| INT-20261028 | Replace obsolete registration/history assumptions with native current-file CT and BR use. | Complete | unifi-stubd integration owner | The dated observation below records actual local implementation, native checks, bounded SDK probes and preserved data/index. | Historical result applies to its observed inputs; newer BR/CT source content requires current reconciliation. Remote CI was unverified. | Earlier receiving integration baseline. |

## Historical observation — 2026-10-28

The following record preserves the earlier task's observed inputs and outcomes.
It does not certify the current upstream source or newer local implementation.

Replaced obsolete provider registration and historical execution assumptions
while preserving the actual incoming staged/unstaged work, policy choices,
manifests, controller/lab boundaries, domain data and Git index. Native Make/CI
dispatch inherits one receiving ENV. The existing check owner is now
`scripts/check-ct.py`, uses official SDK/direct discovered schemas and selects
only ID `ct`, `bin/provider`, `bin/ct` and `connections/ct/`. The configuration was
renamed to `ct.toml` without replacing the existing choices. Go/JavaScript retain
their existing Build Runner owners.

Own function/annotations skills pass current-file, implementation, subscription
and onward duties. The existing check owner also offers narrow read-only
annotations, exact own skill resources and SDK subscriptions. Its dependency is
`requirements/unifi-stubd/mcp.txt`; the terminal is not a full domain provider.

Removed two obsolete provider.json files, 719 inactive installed source/entry/
registration files and a further 2,920 obsolete source/contract/catalog captures
after finite role/hash inventories and writer, handle, lock and mount checks.
Removed obsolete native aliases without a source archive. Useful execution/
closure evidence, current SDK/runtime locks and Go/Node/controller data remain.
A final direct inventory found no `runtime/release-*` tree.

The ready CT source was bound through its public `mcp --connection` owner.
Acknowledged active catalog/annotations/list subscriptions, all 11 required
terminal resources, ID-only annotations and the exact-root contract passed.
Exactly four current CT skills and their companion retain original bytes/modes;
no code/catalog/contract cache remains. The initial observation stopped on a
task-local response-envelope mistake after reading; its corrected observation
passed.

The real native CT Bash/Python gate passed all seven required checks. Own SDK
discovery/resource/filter/invalid-input checks, empty server version,
byte/mode/missing-file invalidation, sticky restore rejection and reconnection
passed. The concrete unreadable-file fault was reproduced before the narrow
OSError correction; both affected SDK probes and three CT Python checks then
passed. Local BR and CT preflight, existing policy checks, scoped Ruff, shell
syntax/lint/format, YAML parsing and own skill validation passed. Actionlint is
unavailable. No controller, hardware, remote CI or production workload was run.

The existing root owner now confirms directory read/search access through a
safe directory descriptor. Isolated SDK cases for root modes 000/0400 fail
without hints and permit the same session after restore; an unreadable required
skill remains sticky with hints and needs reconnect. The current four-file
fixtures were retired after client exit and handle checks. The affected CT
Python checks passed again; no domain tests were added or repeated.

The ready current BR delivery was read after its source owner's receiving-root
fix: all 88 declared resources, including 17 terminal required resources, verified
with an acknowledged live stream and actual annotations. Its public native
connection matches the saved binding; no domain workload ran.
The final ready CT delivery was read with its acknowledged stream and all 11
required resources. Four native CT skills, their companion and the declared
AGENTS section retain exact supplied bytes/modes; the project prefix is preserved.
Its new root-availability notice matches the verified own root implementation.
The same bounded review and its CT binding/coverage result use the existing
work order and selected own files; see the task's review/results records.

The actual GitHub observation found zero repository variables, one online
FreeBSD worker without `docker-engine`, and published CI still on
`ubuntu-latest`. That worker explicitly forbids Docker daemon access. The
then-proposed route required a Docker-capable self-hosted worker and runner-local
`INITIAL_MANAGEMENT_LAUNCHER` / `CT_INITIAL_MANAGEMENT_LAUNCHER` settings.
Existing administration documents
name `docker` / VM 400 / Compose deployment `unifi-stubd-gh-runner`, but record
unknown host/application access and no MCP target channel. The local CT source
path supplies no Linux CI access. No remote setting or runner was changed;
that observation did not establish complete self-hosted CI success.

The user clarified the current role split on 2026-10-30: the existing
self-hosted runner serves only FreeBSD packaging; general checks use
GitHub-hosted workers and BR containers. The earlier requirement for a separate
Docker-enabled self-hosted CI worker was overbroad. Current CI-001 tracks
private-source access and current-file delivery instead; the historical
observation above grants no runner conversion or remote change authority.

Thin inventories, actual outcomes and preservation records are in
`.consumer-data/evidence/unifi-stubd/current-file-integration-20261028/`.
No commit, staging or push is authorized in this receiving task.
