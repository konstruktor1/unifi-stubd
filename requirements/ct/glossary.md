# Codetool glossary

This supplement specializes Codetool terms alongside the unchanged canonical
dictionary at
`provider-provider.catalog.required_domain_provider_interface.glossary` and any
applicable retained glossary requirements. It does not extend inherited meanings.

- **Codetool**: The function owner for mandatory code-quality checks and
  optional Anti-Bloat and internationalization policies in a consumer project.
  Its short name is CT and its provider ID is `ct`. C means consumer; CC means
  consumer of a consumer. Neither C nor CC names this provider.
- **Service contract**: The effective policies, catalog/module selections,
  native settings and actual implementation identity returned by
  `bin/ct contract . --json`. The provider contract separately declares
  functional commands, exact artifacts and owned boundaries.
- **Current artifact integrity**: Provider identity contains `id` only.
  The native-bound file set supplies execution and resources; actual SHA-256
  bytes and modes establish equality. Requirement provenance contains only
  `provider_id`, `path` and `sha256`. Git, commit, tag, approval or acceptance
  history selects no implementation. Schema/protocol versions and functional
  toolchain constraints retain their technical meanings.
- **Portable source paths**: Owner-relative artifact paths, external native
  configuration and validated physical-root/source derivation in owned code and
  generated executable content. Hard-coded machine/project absolute paths and
  cwd/PATH fallback for provider or project roots are absent. Configured or
  derived canonical absolute runtime paths still require physical checks; exact
  foreign/inherited bytes remain under their original owner.
- **Canonical invocation**: Management uses the catalog's
  `provider_contract.management` verbatim. Public entries resolve from the
  configured durable provider root and execute at the declared consumer cwd with
  required native context. Callers use the generated native MCP connection,
  exact tool/schema names and literal domain argv, without management aliases,
  public startup wrappers, internal-script bypasses or cwd/PATH root fallback.
  Every onward owner enforces and passes this same method, including to terminals.
- **Function input ownership**: PP defines mandatory common interfaces and
  inputs; each existing function owner implements its own commands and determines
  necessary own required and optional arguments. The existing contract, own skill
  and callable MCP schema declare them; callers supply applicable common and
  function-required inputs through existing validators. Every onward owner
  preserves and passes this same boundary without weakening the common interface.
- **Policy catalog**: The packaged language, paradigm, architecture, pattern and
  policy data under `src/ct_provider/catalog`. The service contract
  selects applicable rules without creating another declaration owner.
- **Work instruction**: The service contract's `work_instruction` with selected
  rule IDs and full-text locators, source/configuration binding, scope, required
  checks and completion duties. The consumer applies it; delivery certifies no
  semantic work.
- **Mandatory check**: Applicable code-quality verification executed through
  `bin/ct check . --require-tools --json` or MCP `check`.
- **Anti-Bloat**: Optional rules applied during ordinary implementation and one
  final review of the authorized task diff. Full scope is explicit. It creates
  no cadence, separate reviewer, scheduled job or exclusive completion gate.
- **Work order**: The Anti-Bloat instruction for the finished task diff, binding
  its consumer, scope, current content and policy and supplying the compact result
  path and finalization command. Local corrections remain in that domain order.
- **Anti-Bloat result**: The record of reviewed paths, findings and actual check
  references. Finalization validates binding, freshness and coverage, not semantic
  correctness or execution of referenced checks.
- **Internationalization plan**: Selected locale, language and platform work
  returned by `i18n plan . --json`; the consumer implements and verifies it.
- **Project configuration**: Consumer-owned `ct.toml`. Its deliberate
  settings and native Python choices remain data owned by that consumer.
- **Codetool persistence instructions**: The read-only common-format response
  for a declared CT function/purpose, describing native consumer preferences or
  selected Anti-Bloat review-result storage and actual later use. Existing native
  owners retain writes, fresh reads, values, reset and unknown outcomes.
- **Codetool delivery skill**: The own selectable `ct-delivery` product skill
  for authorized receiving/onward setup. It carries current guidance, persistence
  inquiry/use, the common PP backlog format and observed verification; filled
  engineering records and personal values remain local.
- **Codetool storage**: Its specialization of PP's sole physical project
  layout. Evidence, dependencies and connection exports use their inherited
  `.consumer-data` roles; contained runtime stays in `.provider-provider/runtime`
  and disposable local verification uses `tests/.work`. Native data never grants
  filesystem authority.
- **Stable service launcher**: The supplied current bound source's `bin/ct`
  entry, resolving actual current implementation without integration acceptance
  or a registered consumer installation.
- **Native Python workload**: The actual interpreter selected by the existing
  Codetool runtime owner from request options, saved `[python]` and the
  observed interpreter. Fixed shared MCP SDK setup is a separate owner.
- **Management MCP**: The stable `bin/provider mcp` transport for current
  instructions, exact resources and callable domain functions. The unchanged PP
  package owns its contained runtime and integrity; Codetool owns its adapter.
- **Codetool annotations**: Current necessary function-change notices from
  `.agents/skills/ct-annotations/SKILL.md`, also projected through
  callable `annotations` with provider ID only and an optional function filter.
- **Codetool implementation supplement**: The transitive source/operation map
  at `requirements/ct/implementation.md`, binding actual source and data
  owners under PP's unchanged implementation definition.
- **Codetool integration requirement**: This provider's transitive function
  use, completion and onward-delivery duty at
  `requirements/ct/consumer-integration.md`. One-time inventories and
  execution records remain in the receiving project's backlog/evidence. Each
  receiver implements and passes current-file execution, portable paths,
  canonical invocation and function input ownership through its own contract,
  skills and work instructions, including to terminal consumers; actual
  violations are corrected in that authorized task.

Receiving providers retain applicable foreign glossary sources with original
owners, definitions, source paths, bytes and modes. Returning requirements stop at
their source owner. Add an owned English specialization beside inherited
originals and expose both through current MCP. Proposed meaning changes belong
to the affected source owner.
