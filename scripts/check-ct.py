#!/usr/bin/env python3
"""Run selected quality checks, inspect provider context, or serve diagnostics."""

import argparse
import asyncio
import hashlib
import json
import os
import sys
from contextlib import asynccontextmanager
from pathlib import Path

from integration_doctor import DOCTOR_INPUT_SCHEMA, doctor
from mcp import MCPError
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.server.subscriptions import (
    InMemorySubscriptionBus,
    ListenHandler,
    ResourcesListChanged,
    ResourceUpdated,
)
from mcp.types import (
    CallToolResult,
    ListResourcesResult,
    ListToolsResult,
    ReadResourceResult,
    Resource,
    TextContent,
    TextResourceContents,
    Tool,
    ToolAnnotations,
)
from provider_context import (
    current_file,
    emit_ct_diagnostics,
    native_configuration,
    provider_client,
    receiving_root,
    require_native_root,
    run_owned_process,
)

PROJECT_ID = "unifi-stubd"
ANNOTATIONS = f".agents/skills/{PROJECT_ID}-annotations/SKILL.md"
RESOURCE_URI = f"provider://{PROJECT_ID}/{ANNOTATIONS}"
SKILL_PATHS = tuple(
    f".agents/skills/{PROJECT_ID}{suffix}/SKILL.md"
    for suffix in ("", "-annotations", "-doctor", "-delivery", "-update")
)


def doctor_arguments(arguments):
    """Apply the published closed enum schema before any diagnostic effects."""
    properties = DOCTOR_INPUT_SCHEMA["properties"]
    if set(arguments) - set(properties) or any(
        not isinstance(value, str) or value not in properties[key]["enum"]
        for key, value in arguments.items()
    ):
        raise MCPError(-32602, "Use the published Doctor input schema")
    return arguments


async def serve_annotations(root):
    paths = (
        Path(__file__).relative_to(root).as_posix(),
        "scripts/provider_context.py",
        "scripts/integration_doctor.py",
        *SKILL_PATHS,
        f"requirements/{PROJECT_ID}/mcp.txt",
    )
    initial = {path: current_file(root / path) for path in paths}
    skill = initial[ANNOTATIONS][0]
    exposed = {
        f"provider://{PROJECT_ID}/{path}": (
            path,
            data,
            {"sha256": hashlib.sha256(data).hexdigest(), "mode": mode},
        )
        for path in SKILL_PATHS
        for data, mode in (initial[path],)
    }
    bus = InMemorySubscriptionBus()
    listener = ListenHandler(bus)
    stale = False

    async def require_current():
        nonlocal stale
        try:
            require_native_root(root)
        except (OSError, ValueError) as error:
            raise MCPError(
                -32603, "The native receiving context is unavailable"
            ) from error
        if stale:
            raise MCPError(
                -32603, "Current integration files changed; reconnect and read again"
            )
        try:
            changed = any(
                current_file(root / path) != value for path, value in initial.items()
            )
        except OSError, ValueError:
            try:
                require_native_root(root)
            except (OSError, ValueError) as error:
                raise MCPError(
                    -32603, "The native receiving context is unavailable"
                ) from error
            changed = True
        if changed:
            stale = True
            await bus.publish(ResourceUpdated(RESOURCE_URI))
            await bus.publish(ResourcesListChanged())
            raise MCPError(
                -32603, "Current integration files changed; reconnect and read again"
            )

    async def list_tools(_context, _params):
        await require_current()
        return ListToolsResult(
            tools=[
                Tool(
                    name="annotations",
                    description="Read this project's current integration notices.",
                    input_schema={
                        "type": "object",
                        "properties": {"function": {"type": "string", "minLength": 1}},
                        "additionalProperties": False,
                    },
                    annotations=ToolAnnotations(
                        read_only_hint=True,
                        destructive_hint=False,
                        idempotent_hint=True,
                        open_world_hint=False,
                    ),
                ),
                Tool(
                    name="doctor",
                    description="Inspect this project's setup; explicit verify runs local plausibility checks.",
                    input_schema=DOCTOR_INPUT_SCHEMA,
                    annotations=ToolAnnotations(
                        read_only_hint=False,
                        destructive_hint=True,
                        idempotent_hint=False,
                        open_world_hint=True,
                    ),
                ),
            ]
        )

    async def call_tool(_context, params):
        await require_current()
        arguments = params.arguments or {}
        if params.name == "doctor":
            report = await doctor(root, **doctor_arguments(arguments))
            await require_current()
            payload = {"provider": {"id": PROJECT_ID}, "result": report}
            return CallToolResult(
                content=[TextContent(type="text", text=json.dumps(payload))],
                structured_content=payload,
            )
        function = arguments.get("function")
        if (
            params.name != "annotations"
            or set(arguments) - {"function"}
            or (
                "function" in arguments
                and (not isinstance(function, str) or not function)
            )
        ):
            raise MCPError(
                -32602,
                "Use annotations with only its optional nonempty function filter",
            )
        notices = json.loads(
            skill.decode().split("```json\n", 1)[1].split("\n```", 1)[0]
        )["notices"]
        if function is not None:
            notices = [notice for notice in notices if notice["function"] == function]
        payload = {"provider": {"id": PROJECT_ID}, "notices": notices}
        await require_current()
        return CallToolResult(
            content=[TextContent(type="text", text=json.dumps(payload))],
            structured_content=payload,
        )

    async def list_resources(_context, _params):
        await require_current()
        return ListResourcesResult(
            resources=[
                Resource(
                    uri=uri,
                    name=path,
                    mime_type="text/markdown",
                    size=len(data),
                    meta=metadata,
                )
                for uri, (path, data, metadata) in exposed.items()
            ]
        )

    async def read_resource(_context, params):
        await require_current()
        if str(params.uri) not in exposed:
            raise MCPError(-32602, "Read this project's current own skills")
        _path, data, metadata = exposed[str(params.uri)]
        return ReadResourceResult(
            contents=[
                TextResourceContents(
                    uri=str(params.uri),
                    mime_type="text/markdown",
                    text=data.decode(),
                    meta=metadata,
                )
            ]
        )

    async def listen(context, params):
        await require_current()
        if set(params.notifications.resource_subscriptions or ()) - set(exposed):
            raise MCPError(
                -32602, "Subscribe only to this project's own skill resources"
            )
        return await listener(context, params)

    @asynccontextmanager
    async def lifespan(_server):
        try:
            yield {}
        finally:
            listener.close()

    server = Server(
        PROJECT_ID,
        version="",
        lifespan=lifespan,
        on_list_tools=list_tools,
        on_call_tool=call_tool,
        on_list_resources=list_resources,
        on_read_resource=read_resource,
        on_subscriptions_listen=listen,
    )
    async with stdio_server() as streams:
        await server.run(*streams, server.create_initialization_options())


async def ct_call(client, name, arguments):
    logical_call = {"provider": "ct", "name": name, "arguments": arguments}
    reply = await client.call_tool(logical_call["name"], logical_call["arguments"])
    emit_ct_diagnostics(reply)
    return reply


async def check_build_runner(root):
    """Read the declared BR notices and optional persistence guidance without saving."""
    async with provider_client(root, "build-runner") as client:
        for name, arguments in (
            ("annotations", {}),
            (
                "persistence-instructions",
                {"function": "runner", "purpose": "operator-targets"},
            ),
        ):
            logical = {"provider": "build-runner", "name": name, "arguments": arguments}
            reply = await client.call_tool(logical["name"], logical["arguments"])
            if reply.is_error:
                raise RuntimeError(f"Build Runner {name} failed: {reply}")
            print(json.dumps(reply.structured_content, indent=2))
    return 0


async def check_quality(root):
    async with provider_client(root, "ct") as client:
        selected = await ct_call(client, "contract", {"use_case": "cli"})
        if selected.is_error or selected.structured_content["result"]["exit_code"]:
            raise RuntimeError("Codetool contract failed")
        work = selected.structured_content["result"]["result"]["work_instruction"]
        if work["consumer_boundary"]["project"] != str(root) or work["binding"][
            "used_service"
        ] != {"id": "ct"}:
            raise RuntimeError("Codetool selected another source or receiving boundary")
        if (
            selected.structured_content["result"]["result"]["configuration"].get(
                "check_environment"
            )
            == "build-runner"
        ):
            return await check_runner(root)
        result = await ct_call(
            client,
            "check",
            {"use_case": "cli", "require_tools": True},
        )
        payload = result.structured_content
        print(json.dumps(payload, indent=2))
        if result.is_error or payload["provider"] != {"id": "ct"}:
            raise RuntimeError("Codetool returned a protocol/function error")
        outcome = payload["result"]
        status = outcome["exit_code"]
        rows = outcome["result"].get("results", [])
        if not status and (
            not rows
            or outcome["result"]["semantic_work"]["unresolved_checks"]
            or any(row.get("required") and row["status"] != "passed" for row in rows)
        ):
            raise RuntimeError("Required native checks were empty or unresolved")
        return status


async def check_runner(root):
    """Project the confirmed BR choice through its existing native function owner."""
    ct = native_configuration(root, "ct")
    runner = native_configuration(root, "build-runner")
    if runner["env"].get("BUILD_RUNNER_HOST_PATH_CT_SOURCE") != str(
        Path(ct["command"]).parents[1]
    ):
        raise ValueError(
            "The native connection owner must rebind the current CT source"
        )
    async with provider_client(root, "build-runner") as client:
        prepared = await run_owned_process(root, ["make", "quality-ct-image"])
        print(prepared["stdout"], file=sys.stderr, end="")
        print(prepared["stderr"], file=sys.stderr, end="")
        if prepared["exit_code"]:
            raise RuntimeError("The selected Build Runner CT image preparation failed")
        logical = {
            "provider": "build-runner",
            "name": "runner",
            "arguments": {
                "runner": "build-runner",
                "argv": [
                    "--project-name",
                    PROJECT_ID,
                    "instance",
                    "run",
                    "--manifest",
                    "build-runner/instances/ct-quality.yaml",
                    "--",
                    "env",
                    "PROVIDER_CONSUMER_CONTEXT="
                    + os.environ["PROVIDER_CONSUMER_CONTEXT"],
                    "/opt/codetool/bin/ct",
                    "check",
                    ".",
                    "--use-case",
                    "cli",
                    "--require-tools",
                    "--json",
                ],
            },
        }
        reply = await client.call_tool(logical["name"], logical["arguments"])
        if reply.is_error or reply.structured_content["provider"] != {
            "id": "build-runner"
        }:
            raise RuntimeError(f"Build Runner CT check failed: {reply}")
        result = reply.structured_content["result"]
        print(result["stderr"], file=sys.stderr, end="")
        print(result["stdout"], end="")
        checked = json.loads(result["stdout"])
        if (
            result["exit_code"] != 0
            or not checked.get("results")
            or checked["semantic_work"]["unresolved_checks"]
            or any(
                row.get("required") and row["status"] != "passed"
                for row in checked["results"]
            )
        ):
            raise RuntimeError("The selected Build Runner CT checks remain unresolved")
    return 0


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    action = parser.add_mutually_exclusive_group()
    action.add_argument("--mcp-annotations", action="store_true")
    action.add_argument("--doctor", action="store_true")
    action.add_argument("--check-build-runner", action="store_true")
    for key, description in DOCTOR_INPUT_SCHEMA["properties"].items():
        parser.add_argument(f"--{key.replace('_', '-')}", choices=description["enum"])
    args = parser.parse_args()
    root = receiving_root()
    os.environ["PYTHONDONTWRITEBYTECODE"] = "1"
    arguments = {
        key: getattr(args, key)
        for key in DOCTOR_INPUT_SCHEMA["properties"]
        if getattr(args, key) is not None
    }
    if arguments and not args.doctor:
        parser.error("Doctor inputs require --doctor")
    if args.mcp_annotations:
        asyncio.run(serve_annotations(root))
        return 0
    if args.doctor:
        report = asyncio.run(doctor(root, **doctor_arguments(arguments)))
        print(json.dumps({"provider": {"id": PROJECT_ID}, "result": report}, indent=2))
        return 0 if report["outcome"] == "passed" else 1
    if args.check_build_runner:
        return asyncio.run(check_build_runner(root))
    return asyncio.run(check_quality(root))


if __name__ == "__main__":
    raise SystemExit(main())
