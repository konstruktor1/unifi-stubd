# Credits and Research Sources

`unifi-stubd` stands on public documentation, older reverse-engineering notes,
and lab validation. Thank you to the people who published their notes and code;
without that trail this project would be much slower and much fuzzier.

No source code from the research repositories listed here has been copied into
`unifi-stubd`. They were used as protocol documentation, historical context,
and sanity checks for independent Go code in this repository.

## Project License Decision

`unifi-stubd` is licensed under AGPL-3.0-or-later.

AGPL-3.0-or-later is a good fit because:

- The project is a small lab tool with original implementation code.
- The direct runtime dependencies are permissively licensed.
- The reverse-engineering sources were used for ideas and protocol facts, not
  copied implementation.
- The project should remain free software when redistributed, embedded in
  packages, or offered as a modified network-accessible service.
- The license is GPLv3-compatible for combined works covered by GPLv3.

If future work copies source code from a project with a different license, the
license decision must be revisited before merging that code.

## Research and Idea Sources

| Source | Who | What helped | Used in this repo | License/status noted |
| --- | --- | --- | --- | --- |
| [Ubiquiti Required Ports Reference](https://help.ui.com/hc/en-us/articles/218506997-UniFi-Network-Required-Ports-Reference) | Ubiquiti | Official port roles for discovery, inform, STUN, UI/API, and related services | Lab docs, service assumptions, safety boundaries | Official documentation |
| [Ubiquiti Remote Adoption / Layer 3](https://help.ui.com/hc/en-us/articles/204909754-Remote-Adoption-Layer-3) | Ubiquiti | Official remote adoption framing and inform URL workflow | Adoption docs and command model | Official documentation |
| [Ubiquiti UniFi Pro XG 48 Tech Specs](https://techspecs.ui.com/unifi/switching/usw-pro-xg-48?subcategory=all-switching) | Ubiquiti | Official product identifier and mixed 16x 2.5G RJ45, 32x 10G RJ45, 4x 25G SFP28 port layout | `usw-pro-xg-48` profile and port group tests | Official documentation |
| [The unofficial guide to UniFi](https://jrjparks.github.io/unofficial-unifi-guide/) | jrjparks | Discovery, inform, adoption flow, flags, AES-CBC/AES-GCM, zlib/snappy notes | `internal/discovery`, `internal/inform`, protocol docs | Apache-2.0 |
| [jeffreykog/unifi-inform-protocol](https://github.com/jeffreykog/unifi-inform-protocol) | Jeffrey Kog, historically also referenced as jk-5 | Discovery TLVs, inform packet header, flags, SSH adoption commands | Discovery TLV layout, inform packet layout, adoption SSH command expectations | No GitHub license detected; no code copied |
| [fxkr/unifi-protocol-reverse-engineering](https://github.com/fxkr/unifi-protocol-reverse-engineering) | fxkr | Early reverse-engineered discovery and inform notes; broadcast/multicast addresses; TLV framing | Discovery broadcast/multicast sanity checks and protocol docs | No GitHub license detected; no code copied |
| [mcrute/ubntmfi inform_protocol.md](https://github.com/mcrute/ubntmfi/blob/master/inform_protocol.md) | mcrute | Raw inform packet structure, CBC/zlib framing, pull-based inform/provisioning model | Inform encoder/decoder mental model and protocol notes | BSD-3-Clause |
| [wvengen/unifi-controllable-switch](https://github.com/wvengen/unifi-controllable-switch) | wvengen | Earlier TOUGHswitch-to-UniFi switch integration idea; status payload examples; `syswrapper.sh set-adopt` behavior | Project concept, fake switch payload shape, adoption SSH shim direction | No GitHub license detected; no code copied |
| [stephanlascar/unifi-gateway](https://github.com/stephanlascar/unifi-gateway) | stephanlascar | pfSense/UGW emulation proof of concept and gateway payload direction | Roadmap context for later gateway/DPI work | No GitHub license detected; no code copied |
| [jda/pixiedust](https://github.com/jda/pixiedust) | jda | PCAP-based inform analysis, key extraction, `setparam` examples, AES-GCM field observations | Adoption state parsing and lab comparison approach | MIT |
| [ZAP-Quebec/unifi-inform](https://github.com/ZAP-Quebec/unifi-inform) | ZAP Québec | Older Go implementation used as a cross-check for packet/header concepts | Inform packet structure sanity checks | MIT |
| [Tamarack: Reverse Engineering the UniFi Inform Protocol](https://tamarack.cloud/blog/reverse-engineering-unifi-inform-protocol) | Tamarack | Narrative explanation of inform protocol reverse engineering | Research context and protocol docs | Article/source reference only |
| [dmke/inform-inspect](https://github.com/dmke/inform-inspect) | dmke | Independent packet inspection model and upstream thanks to Mike Crute and Jeffrey Kog | Cross-check for parser/decrypt/decompress separation | MIT |

## Build and Release Sources

| Source | What helped |
| --- | --- |
| [Go Toolchains](https://go.dev/doc/toolchain) | Go version policy and automatic toolchain behavior considered when keeping repository files unpinned to a patch toolchain |
| [Go Workspaces](https://go.dev/doc/tutorial/workspaces) | Repository-local `go.work` setup |
| [Go Modules Reference](https://go.dev/ref/mod) | `tool` directives for tracked build tools |
| [nFPM documentation](https://nfpm.goreleaser.com/docs/configuration/) | Debian, RPM, and Arch Linux package generation |
| [GitHub Actions checkout](https://github.com/actions/checkout), [setup-go](https://github.com/actions/setup-go), and [upload-artifact](https://github.com/actions/upload-artifact) | CI workflow and package artifact publishing |
| [GNU Affero General Public License v3](https://www.gnu.org/licenses/agpl-3.0.txt) | Copyleft license text used by the project |
| [SPDX AGPL-3.0-or-later](https://spdx.org/licenses/AGPL-3.0-or-later.html) | Package and documentation license identifier |

## Agent and LLM Documentation Sources

| Source | What helped |
| --- | --- |
| [AGENTS.md](https://github.com/agentsmd/agents.md) | Root agent instruction file for coding agents |
| [GitHub Copilot repository custom instructions](https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/add-custom-instructions/add-repository-instructions) | `.github/copilot-instructions.md` and path-specific `.github/instructions/*.instructions.md` setup |
| [llms.txt proposal](https://llmstxt.org/) | Root `llms.txt` public project index |
| [Claude Code memory documentation](https://code.claude.com/docs/en/memory) | `CLAUDE.md` project instruction bridge |
| [Gemini CLI GEMINI.md documentation](https://google-gemini.github.io/gemini-cli/docs/cli/gemini-md.html) | `GEMINI.md` project instruction bridge and import syntax |
| [Cursor rules documentation](https://docs.cursor.com/en/context) | `.cursor/rules/*.mdc` project rules |
| [Windsurf memories and rules documentation](https://docs.windsurf.com/plugins/cascade/memories) | `.windsurf/rules/*.md` workspace rules |
| [Cline rules documentation](https://docs.cline.bot/customization/cline-rules) | `.clinerules/` workspace rules |
| [Roo Code documentation](https://docs.roocode.com/) | Legacy `.roo/rules/` compatibility note |
| [Aider conventions documentation](https://aider.chat/docs/usage/conventions.html) and [Aider config documentation](https://aider.chat/docs/config/aider_conf.html) | `CONVENTIONS.md` and `.aider.conf.yml` read-only context setup |

## Runtime Dependencies

Only the command runtime dependencies are considered part of the distributed
binary:

| Module | License |
| --- | --- |
| `golang.org/x/crypto` | BSD-3-Clause |
| `gopkg.in/yaml.v3` | MIT and Apache-2.0, with upstream NOTICE |

`golangci-lint` and `nFPM` are tracked as Go tools for reproducible builds.
They are not linked into the `unifi-stubd` binary. If their binaries are
redistributed separately, their licenses must be followed separately.

## Trademarks and Affiliation

UniFi, Ubiquiti, and related product names are trademarks of their respective
owners. This project is independent, unofficial, and not endorsed by Ubiquiti.
