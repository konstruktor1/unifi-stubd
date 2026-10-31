"""Read current native provider instructions under an active SDK subscription."""

import asyncio
import base64
import hashlib
import json
import os
import signal
import stat
import subprocess
import sys
from contextlib import asynccontextmanager
from pathlib import Path

from mcp import Client, ClientSession, StdioServerParameters
from mcp.types import DiscoverResult
from mcp.types.version import MODERN_PROTOCOL_VERSIONS


def require_native_root(root):
    """Confirm the one native directory is physically readable and searchable."""
    if (
        not root.is_absolute()
        or Path(os.environ["PROVIDER_CONSUMER_ROOT"]) != root
        or root.resolve(strict=True) != root
        or Path.cwd() != root
    ):
        raise ValueError("Use the unchanged native receiving root and cwd")
    descriptor = os.open(root, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        opened = os.fstat(descriptor)
        searched = os.stat(".", dir_fd=descriptor, follow_symlinks=False)
        if not os.path.samestat(opened, searched) or not os.path.samestat(
            opened, root.lstat()
        ):
            raise ValueError("The native receiving directory changed during inspection")
    finally:
        os.close(descriptor)


def receiving_root():
    """Require inherited native context to match this physical script owner."""
    root = Path(os.environ["PROVIDER_CONSUMER_ROOT"])
    require_native_root(root)
    script = Path(__file__).absolute()
    if script.resolve(strict=True) != script or script.parents[1] != root:
        raise ValueError("Use the unchanged native receiving root and cwd")
    return root


async def run_owned_process(root: Path, argv: list[str]) -> dict[str, int | str]:
    """Run a contained child and close its process group before cancellation exits."""
    require_native_root(root)
    child = await asyncio.create_subprocess_exec(
        *argv,
        cwd=root,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE,
        start_new_session=True,
    )
    try:
        stdout, stderr = await child.communicate()
    except BaseException as error:
        # Cleanup also covers I/O failure; a child may ignore TERM or retain pipes.
        for shutdown_signal in (signal.SIGTERM, signal.SIGKILL):
            try:
                os.killpg(child.pid, shutdown_signal)
            except ProcessLookupError:
                pass
            try:
                async with asyncio.timeout(5):
                    await child.communicate()
            except TimeoutError:
                continue
            try:
                os.killpg(child.pid, 0)
            except ProcessLookupError:
                raise error
        raise RuntimeError(
            "Child shutdown is unconfirmed; reconcile its outcome before retrying"
        ) from error
    return {
        "exit_code": child.returncode,
        "stdout": stdout.decode("utf-8", errors="replace"),
        "stderr": stderr.decode("utf-8", errors="replace"),
    }


def current_file(path):
    """Read one physical regular file without accepting links or concurrent writes."""
    if path.resolve(strict=True) != path:
        raise ValueError("Current integration source must have its physical owned path")
    with os.fdopen(os.open(path, os.O_RDONLY | os.O_NOFOLLOW), "rb") as stream:
        before = os.fstat(stream.fileno())
        content = stream.read()
        after = os.fstat(stream.fileno())
    if (
        not stat.S_ISREG(after.st_mode)
        or after.st_nlink != 1
        or not os.path.samestat(before, after)
        or not os.path.samestat(after, path.lstat())
        or any(
            getattr(before, field) != getattr(after, field)
            for field in (
                "st_mode",
                "st_size",
                "st_mtime_ns",
                "st_ctime_ns",
                "st_nlink",
            )
        )
        or len(content) != after.st_size
    ):
        raise ValueError(
            "Current integration source is linked or changed while reading"
        )
    return content, stat.S_IMODE(after.st_mode)


def native_configuration(root, provider):
    """Validate the existing binding against its public native connection owner."""
    if provider not in {"ct", "build-runner"}:
        raise ValueError("Use this project's independently selected BR or CT binding")
    require_native_root(root)
    data, _mode = current_file(
        root / f".consumer-data/connections/{provider}/management.json"
    )
    configuration = json.loads(data)
    command = Path(configuration["command"])
    if (
        command.resolve(strict=True) != command
        or command.name != "provider"
        or command.parent.name != "bin"
        or configuration["args"] != ["mcp"]
        or configuration["cwd"] != str(root)
        or configuration["env"]["PROVIDER_CONSUMER_ROOT"] != str(root)
        or configuration["env"]["PROVIDER_CONSUMER_CONTEXT"]
        != os.environ["PROVIDER_CONSUMER_CONTEXT"]
    ):
        raise ValueError("Regenerate the native connection from its current context")
    fresh = subprocess.run(
        [str(command), "mcp", "--connection"],
        cwd=root,
        env=os.environ | configuration["env"],
        capture_output=True,
        text=True,
        check=True,
    )
    if json.loads(fresh.stdout) != configuration:
        raise ValueError("Native connection changed; regenerate its saved settings")
    return configuration


def native_catalog(configuration, provider):
    """Read the current catalog through the already validated native source owner."""
    source = Path(configuration["command"]).parents[1]
    data, _mode = current_file(source / "provider-contract.json")
    contract = json.loads(data)
    if contract["provider"] != {"id": provider}:
        raise RuntimeError("Native source contract identifies another provider")
    requirement = next(
        row
        for row in contract["requirements"]
        if row["id"] == "provider-provider.catalog"
    )
    relative = Path(requirement["materialized_path"])
    if (
        relative.is_absolute()
        or relative.as_posix() != requirement["materialized_path"]
        or ".." in relative.parts
    ):
        raise ValueError("The catalog must retain its source-relative physical path")
    data, mode = current_file(source / relative)
    if hashlib.sha256(data).hexdigest() != requirement["source"]["sha256"]:
        raise RuntimeError("Current catalog differs from its native source contract")
    return relative.as_posix(), data, mode


def protocol_options(catalog):
    """Require the catalog's protocol and SDK support without choosing a fallback."""
    required = catalog["required_domain_provider_interface"]["mcp"]["protocol_version"]
    missing = [
        name
        for owner, name in (
            (Client, "list_tools"),
            (Client, "list_resources"),
            (Client, "read_resource"),
            (Client, "call_tool"),
            (Client, "listen"),
            (ClientSession, "send_discover"),
            (ClientSession, "adopt"),
        )
        if not callable(getattr(owner, name, None))
    ]
    if required not in MODERN_PROTOCOL_VERSIONS:
        missing.append(f"protocol {required}")
    if missing:
        raise RuntimeError(
            f"The native MCP client owner must update SDK support for {', '.join(missing)}; "
            "no protocol downgrade is permitted"
        )
    return {"mode": required}


async def require_server_capabilities(client, required):
    """Confirm actual discovery supports the selected common MCP behavior."""
    discovered = DiscoverResult.model_validate(
        await client.session.send_discover(required)
    )
    capabilities = discovered.capabilities
    if (
        client.protocol_version != required
        or required not in discovered.supported_versions
        or capabilities.tools is None
        or capabilities.resources is None
        or not capabilities.resources.subscribe
        or not capabilities.resources.list_changed
    ):
        raise RuntimeError(
            f"The native MCP server/client owner must implement catalog protocol {required} "
            "with tools, resources, subscriptions and resource-list changes; no downgrade"
        )
    client.session.adopt(
        discovered.model_copy(update={"supported_versions": [required]})
    )


async def stop_on_change(subscription):
    async for _event in subscription:
        raise RuntimeError("Provider changed; reconcile outcomes, reconnect and reread")
    raise RuntimeError("Provider subscription closed; reconnect before further calls")


def emit_ct_diagnostics(reply):
    """Retain CT's successful-call review work on the existing diagnostic channel."""
    if reply.is_error or reply.structured_content is None:
        return
    diagnostics = reply.structured_content.get("result", {}).get("diagnostics")
    if diagnostics:
        print(diagnostics, file=sys.stderr)


async def read_instructions(client, provider, root):
    """Read complete current instructions and verify each resource descriptor."""
    logical_call = {
        "provider": provider,
        "name": "get_instructions",
        "arguments": {"operation": "work", "consumer_role": "terminal"},
    }
    assignment = await client.call_tool(logical_call["name"], logical_call["arguments"])
    if assignment.is_error or assignment.structured_content["provider"] != {
        "id": provider
    }:
        raise RuntimeError("Provider did not deliver its current function context")
    context = assignment.structured_content
    resources = {
        row["resource"]["uri"]: row["resource"]
        for row in context["reading"]["required"]
    }
    catalog_path = next(
        row["materialized_path"]
        for row in context["requirements"]
        if row["id"] == "provider-provider.catalog"
    )
    descriptors = {row["name"]: row for row in context["resources"]}
    pending = list(resources.values())
    for item in pending:
        data = await read_exact_resource(client, item)
        name = item["name"]
        if provider == "ct" and (
            (name.startswith(".agents/skills/ct") and name.endswith("/SKILL.md"))
            or name
            in {
                "requirements/ct/consumer-integration.md",
                "requirements/ct/provider-instructions.md",
                "requirements/ct/glossary.md",
            }
        ):
            retained, mode = current_file(root / name)
            if retained != data or mode != item["mode"]:
                raise RuntimeError(
                    f"Retained CT artifact {name} differs from the bound current "
                    "source; its integration/source owners must reconcile "
                    "provider-update before this project can use that binding"
                )
        if item["name"] == catalog_path:
            package = json.loads(data)["required_domain_provider_interface"]["mcp"][
                "package"
            ]
            # Keep the complete shared package and its dependency in current context;
            # these read-only resources never become local executable alternatives.
            for name in package.values():
                descriptor = descriptors[name]
                if descriptor["uri"] not in resources:
                    resources[descriptor["uri"]] = descriptor
                    pending.append(descriptor)
    for row in context["reading"]["deferred"]:
        if row.get("requirement_id") == f"{provider}.implementation":
            descriptor = row["resource"]
            if descriptor["uri"] not in resources:
                # Reuse the same exact-resource check for the implementation guidance.
                await read_exact_resource(client, descriptor)


async def read_exact_resource(client, item):
    """Verify an additional declared artifact without copying it into execution."""
    reply = await client.read_resource(item["uri"])
    if len(reply.contents) != 1:
        raise RuntimeError("Expected one exact instruction resource")
    content = reply.contents[0]
    data = (
        content.text.encode()
        if hasattr(content, "text")
        else base64.b64decode(content.blob, validate=True)
    )
    if (
        str(content.uri) != item["uri"]
        or len(data) != item["size"]
        or hashlib.sha256(data).hexdigest() != item["sha256"]
        or type(item["mode"]) is not int
        or content.meta is None
        or content.meta.get("mode") != item["mode"]
        or content.meta.get("sha256") != item["sha256"]
    ):
        raise RuntimeError("Current instruction resource integrity failed")
    return data


@asynccontextmanager
async def provider_client(root, provider):
    """Keep this receiving owner's acknowledged change stream active during use."""
    configuration = native_configuration(root, provider)
    catalog_path, catalog_data, catalog_mode = native_catalog(configuration, provider)
    options = protocol_options(json.loads(catalog_data))
    async with Client(
        StdioServerParameters(**configuration), **options, cache=None
    ) as client:
        await require_server_capabilities(client, options["mode"])
        resources = await client.list_resources(cache_mode="bypass")
        names = {item.name: str(item.uri) for item in resources.resources}
        declared = json.loads(
            (await client.read_resource(names["provider-contract.json"]))
            .contents[0]
            .text
        )
        if declared["provider"] != {"id": provider}:
            raise RuntimeError("Native binding selected another provider")
        catalog = next(
            item["materialized_path"]
            for item in declared["requirements"]
            if item["id"] == "provider-provider.catalog"
        )
        if catalog != catalog_path:
            raise RuntimeError(
                "Discovered catalog differs from the native source binding"
            )
        await read_exact_resource(
            client,
            {
                "uri": names[catalog],
                "size": len(catalog_data),
                "sha256": hashlib.sha256(catalog_data).hexdigest(),
                "mode": catalog_mode,
            },
        )
        uris = [
            names[catalog],
            names[f".agents/skills/{provider}-annotations/SKILL.md"],
        ]
        async with client.listen(
            resource_subscriptions=uris, resources_list_changed=True
        ) as subscription:
            if (
                not subscription.subscription_id
                or not subscription.honored.resources_list_changed
                or set(subscription.honored.resource_subscriptions) != set(uris)
            ):
                raise RuntimeError(
                    "Provider did not acknowledge the required subscription"
                )
            async with asyncio.TaskGroup() as tasks:
                watcher = tasks.create_task(stop_on_change(subscription))
                try:
                    await client.list_tools(cache_mode="bypass")
                    await read_instructions(client, provider, root)
                    yield client
                finally:
                    watcher.cancel()
