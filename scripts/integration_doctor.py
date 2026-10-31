"""Diagnose this terminal project's own provider setup without changing it."""

import json
import os
import tomllib
from pathlib import Path

from provider_context import (
    current_file,
    emit_ct_diagnostics,
    provider_client,
    receiving_root,
    run_owned_process,
)

DOCTOR_INPUT_SCHEMA = {
    "type": "object",
    "properties": {
        "mode": {
            "type": "string",
            "enum": ["inspect", "verify"],
            "default": "inspect",
        },
        "scope": {"type": "string", "enum": ["setup"], "default": "setup"},
        "consumer_role": {
            "type": "string",
            "enum": ["provider", "terminal", "unknown"],
            "default": "unknown",
        },
        "phase": {
            "type": "string",
            "enum": [
                "begin_setup",
                "after_change",
                "before_delivery",
                "before_completion",
                "unknown",
            ],
            "default": "unknown",
        },
    },
    "additionalProperties": False,
}


def finding(identifier, expected, observed, *, status="passed", evidence=()):
    """Keep observations separate from the actions needed to complete coverage."""
    return {
        "id": identifier,
        "owner": "unifi-stubd",
        "requirement": "required",
        "status": status,
        "expected": expected,
        "observed": observed,
        "evidence": list(evidence),
        "next_action": None
        if status == "passed"
        else {
            "owner": "unifi-stubd maintainer",
            "instruction": expected,
            "references": ["docs/en/ci-cd.md", "AGENTS.md"],
        },
    }


def read_anchor(root, relative, snapshots, findings):
    """Read only an explicitly owned anchor, retaining its consistency proof."""
    try:
        observed = current_file(root / relative)
    except (OSError, ValueError) as error:
        snapshots[relative] = error
        findings.append(
            finding(
                f"anchor:{relative}",
                f"Restore the readable regular project file {relative}.",
                str(error),
                status="failed",
                evidence=[relative],
            )
        )
        return None
    snapshots[relative] = observed
    return observed[0]


def anchor_unchanged(root, relative, observed):
    """Include unavailable anchors in the final consistency observation."""
    try:
        current = current_file(root / relative)
    except (OSError, ValueError) as error:
        return type(error) is type(observed) and error.args == observed.args
    return current == observed


def inspect_bindings(root, snapshots, findings):
    """Inspect saved transport data without starting or reconfiguring providers."""
    for provider in ("build-runner", "ct"):
        relative = f".consumer-data/connections/{provider}/management.json"
        raw = read_anchor(root, relative, snapshots, findings)
        if raw is None:
            continue
        try:
            configuration = json.loads(raw)
            command = Path(configuration["command"])
            matches = (
                command.is_absolute()
                and command.resolve(strict=True) == command
                and command.is_file()
                and os.access(command, os.X_OK)
                and command.name == "provider"
                and configuration["args"] == ["mcp"]
                and configuration["cwd"] == str(root)
                and configuration["env"]["PROVIDER_CONSUMER_ROOT"] == str(root)
                and configuration["env"]["PROVIDER_CONSUMER_CONTEXT"]
                == os.environ["PROVIDER_CONSUMER_CONTEXT"]
            )
        except (OSError, ValueError, KeyError, TypeError) as error:
            matches = False
            observed = str(error)
        else:
            observed = (
                "The saved executable, MCP arguments, cwd and inherited native "
                "context match this project. Fresh export and function behavior "
                "have not been checked by this observation."
                if matches
                else "The saved transport data conflicts with the native context."
            )
        findings.append(
            finding(
                f"binding:{provider}",
                "Use the provider-generated saved connection for this unchanged "
                "native project context.",
                observed,
                status="passed" if matches else "failed",
                evidence=[relative],
            )
        )


def inspect_formats(root, snapshots, findings):
    """Parse existing owned formats; their domain validators retain authority."""
    raw = read_anchor(root, "ct.toml", snapshots, findings)
    if raw is not None:
        try:
            if len(raw) > 1024 * 1024:
                raise ValueError("ct.toml exceeds the 1 MiB diagnostic read limit")
            configuration = tomllib.loads(raw.decode("utf-8"))
        except (ValueError, UnicodeError) as error:
            findings.append(
                finding(
                    "ct-configuration-format",
                    "Keep ct.toml readable as bounded UTF-8 TOML.",
                    str(error),
                    status="failed",
                    evidence=["ct.toml"],
                )
            )
        else:
            findings.append(
                finding(
                    "ct-configuration-format",
                    "Keep ct.toml readable as bounded UTF-8 TOML.",
                    f"Parsed profile {configuration.get('profile')!r}; domain "
                    "selection remains with Codetool's contract function.",
                    evidence=["ct.toml"],
                )
            )
    relative = "docs/schema/config.schema.json"
    raw = read_anchor(root, relative, snapshots, findings)
    if raw is None:
        return
    try:
        schema = json.loads(raw)
        if not isinstance(schema, dict):
            raise TypeError("The product configuration schema is not a JSON object")
    except (ValueError, TypeError, UnicodeError) as error:
        findings.append(
            finding(
                "product-schema-format",
                "Keep the product configuration schema readable as a JSON object.",
                str(error),
                status="failed",
                evidence=[relative],
            )
        )
        return
    findings.append(
        finding(
            "product-schema-format",
            "Keep the product configuration schema readable as a JSON object.",
            "Parsed the maintained schema; this does not validate packaged YAML "
            "or replace the product's Go configuration validator.",
            evidence=[relative],
        )
    )


def checked_process(identifier, expected, result, references):
    """Preserve actual process output and distinguish unresolved execution."""
    exit_code = result["exit_code"]
    if exit_code == 125 or exit_code < 0:
        raise RuntimeError(
            f"{identifier} has an unresolved execution or cleanup outcome: "
            + json.dumps(result)
        )
    return finding(
        identifier,
        expected,
        f"exit_code={exit_code}\nstdout:\n{result['stdout']}\nstderr:\n"
        f"{result['stderr']}",
        status="passed" if exit_code == 0 else "failed",
        evidence=references,
    )


async def verify_setup(root, findings):
    """Run existing local policy and public identity/configuration procedures."""
    policy = await run_owned_process(root, ["sh", "scripts/check-policy.sh"])
    findings.append(
        checked_process(
            "project-policy",
            "Pass the existing repository policy check.",
            policy,
            ["scripts/check-policy.sh"],
        )
    )
    async with provider_client(root, "build-runner") as client:
        logical = {
            "provider": "build-runner",
            "name": "runner",
            "arguments": {
                "runner": "runner",
                "argv": ["provider", "--json"],
            },
        }
        reply = await client.call_tool(logical["name"], logical["arguments"])
        if reply.is_error or reply.structured_content["provider"] != {
            "id": "build-runner"
        }:
            raise RuntimeError(f"Build Runner identity call failed: {reply}")
        findings.append(
            checked_process(
                "build-runner-native-identity",
                "Run the current bound provider's public native identity route.",
                reply.structured_content["result"],
                [".consumer-data/connections/build-runner/management.json"],
            )
        )
    async with provider_client(root, "ct") as client:
        logical = {
            "provider": "ct",
            "name": "contract",
            "arguments": {"use_case": "cli"},
        }
        reply = await client.call_tool(logical["name"], logical["arguments"])
        if reply.is_error or reply.structured_content["provider"] != {"id": "ct"}:
            raise RuntimeError(f"Codetool contract call failed: {reply}")
        emit_ct_diagnostics(reply)
        result = reply.structured_content["result"]
        exit_code = result["exit_code"]
        if exit_code == 125 or exit_code < 0:
            raise RuntimeError(f"Codetool contract outcome unresolved: {reply}")
        matches = False
        if exit_code == 0:
            work = result["result"]["work_instruction"]
            matches = (
                work["consumer_boundary"]["project"] == str(root)
                and work["binding"]["used_service"] == {"id": "ct"}
                and work["binding"]["configuration_path"] == str(root / "ct.toml")
            )
        findings.append(
            finding(
                "ct-configuration-selection",
                "Resolve the existing ct.toml through the bound Codetool contract.",
                f"contract exit_code={exit_code}; "
                + (
                    "selected this project's current CLI configuration; no code "
                    "checks run."
                    if matches
                    else json.dumps(result)
                ),
                status="passed" if matches else "failed",
                evidence=["ct.toml", ".consumer-data/connections/ct/management.json"],
            )
        )


async def doctor(
    root,
    *,
    mode="inspect",
    scope="setup",
    consumer_role="unknown",
    phase="unknown",
):
    """Return own setup observations after the caller validates common inputs.

    Inspection reads known project files and saved bindings only. Explicit
    verification additionally starts contained provider support and runs the
    existing policy, native BR identity and CT contract procedures. It creates
    no settings or report and performs no controller or workload operation.
    """
    if receiving_root() != root:
        raise ValueError("Use this Doctor's unchanged native receiving context")
    findings = []
    snapshots = {}
    for relative in (
        "AGENTS.md",
        "Makefile",
        "scripts/check-policy.sh",
        ".agents/skills/unifi-stubd/SKILL.md",
        ".agents/skills/unifi-stubd-annotations/SKILL.md",
        ".agents/skills/unifi-stubd-doctor/SKILL.md",
        ".agents/skills/unifi-stubd-delivery/SKILL.md",
        ".agents/skills/unifi-stubd-update/SKILL.md",
        "packaging/linux/etc/unifi-stubd/config.yaml",
        "packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml",
    ):
        raw = read_anchor(root, relative, snapshots, findings)
        if raw is not None:
            findings.append(
                finding(
                    f"anchor:{relative}",
                    f"Keep {relative} as a safely readable regular project file.",
                    f"Read {len(raw)} bytes with mode {snapshots[relative][1]:04o}.",
                    evidence=[relative],
                )
            )
    inspect_formats(root, snapshots, findings)
    inspect_bindings(root, snapshots, findings)
    if mode == "verify":
        await verify_setup(root, findings)
    else:
        findings.append(
            finding(
                "active-setup-verification",
                "Explicitly request verify for project policy and native provider "
                "identity/configuration checks.",
                "Inspection did not execute project or provider commands.",
                status="not_checked",
            )
        )
    findings.append(
        finding(
            "product-runtime-validation",
            "Run make verify and the BR Go make validate-config entry under "
            "the selected runtime when full product validation is requested.",
            "This setup diagnosis does not run Go/JavaScript checks, packaged "
            "YAML semantic validation, controller acceptance or deployment. "
            "No persistent Operator target was requested or created.",
            status="not_checked",
        )
    )
    if receiving_root() != root or any(
        not anchor_unchanged(root, relative, observed)
        for relative, observed in snapshots.items()
    ):
        raise ValueError("Diagnosis inputs changed; read current context again")
    failed = any(row["status"] == "failed" for row in findings)
    return {
        "scope": {
            "name": scope,
            "mode": mode,
            "consumer_role": consumer_role,
            "phase": phase,
            "context": "unifi-stubd native terminal project; local setup only",
        },
        "findings": findings,
        "coverage": "incomplete",
        "outcome": "failed" if failed else "incomplete",
    }
