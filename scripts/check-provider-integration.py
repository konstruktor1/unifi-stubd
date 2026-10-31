"""Exercise the real own MCP boundary and disposable source invalidation."""

import asyncio
import hashlib
import json
import os
import shutil
import stat
import tempfile
from contextlib import asynccontextmanager
from pathlib import Path

from mcp import Client, MCPError, StdioServerParameters
from mcp.shared.subscriptions import ResourcesListChanged, ResourceUpdated
from provider_context import (
    native_catalog,
    native_configuration,
    protocol_options,
    require_server_capabilities,
    run_owned_process,
)


def require(condition, message):
    if not condition:
        raise AssertionError(message)


def snapshot(path):
    if not path.exists():
        return None
    return path.read_bytes(), stat.S_IMODE(path.stat().st_mode)


async def rejected(client, name, arguments, code):
    try:
        await client.call_tool(name, arguments)
    except MCPError as error:
        require(error.code == code, f"Unexpected {name} rejection: {error}")
        return
    raise AssertionError(f"{name} unexpectedly accepted {arguments!r}")


@asynccontextmanager
async def endpoint(configuration, lifecycle, options, *, changing=False):
    lifecycle["closed"] = False
    caught = None
    async with Client(
        StdioServerParameters(**configuration), **options, cache=None
    ) as client:
        try:
            await require_server_capabilities(client, options["mode"])
            resources = await client.list_resources(cache_mode="bypass")
            annotation = next(
                str(row.uri)
                for row in resources.resources
                if "-annotations/" in row.name
            )
            async with client.listen(
                resource_subscriptions=[annotation], resources_list_changed=True
            ) as stream:
                require(
                    stream.subscription_id
                    and stream.honored.resources_list_changed
                    and stream.honored.resource_subscriptions == [annotation],
                    "Own subscription was not acknowledged",
                )
                events = asyncio.Queue()

                async def observe():
                    async for event in stream:
                        if not changing:
                            raise RuntimeError("Own source changed during verification")
                        await events.put(event)
                    raise RuntimeError("Own subscription closed unexpectedly")

                async with asyncio.TaskGroup() as tasks:
                    watcher = tasks.create_task(observe())
                    try:
                        yield client, resources, events
                    finally:
                        watcher.cancel()
        except BaseException as error:
            caught = error
    lifecycle["closed"] = True
    if caught is not None:
        raise caught


async def real_endpoint(root, configuration, options):
    protected = [root / "ct.toml", root / "build-runner/targets.toml"]
    protected.extend((root / ".consumer-data/connections").glob("*/management.json"))
    before = {path: snapshot(path) for path in protected}
    async with endpoint(configuration, {}, options) as (client, resources, _events):
        tools = {row.name for row in (await client.list_tools()).tools}
        require(
            {"annotations", "doctor"} <= tools, "Own Doctor/annotations not callable"
        )
        for row in resources.resources:
            response = (await client.read_resource(str(row.uri))).contents[0]
            content, mode = snapshot(root / row.name)
            require(
                response.text.encode() == content
                and response.meta["mode"] == mode
                and response.meta["sha256"] == hashlib.sha256(content).hexdigest(),
                f"Skill resource differs from its actual source: {row.name}",
            )
        require(
            {row.name for row in resources.resources}
            == {
                path.relative_to(root).as_posix()
                for path in (root / ".agents/skills").glob("unifi-stubd*/SKILL.md")
            },
            "Installed own skills were not completely discoverable",
        )
        all_notices = (await client.call_tool("annotations", {})).structured_content
        canonical = json.loads(
            (root / ".agents/skills/unifi-stubd-annotations/SKILL.md")
            .read_text()
            .split("```json\n", 1)[1]
            .split("\n```", 1)[0]
        )
        require(
            all_notices == {"provider": {"id": "unifi-stubd"}, **canonical},
            "Callable annotations differ from the current canonical own notices",
        )
        filtered = (
            await client.call_tool("annotations", {"function": "doctor"})
        ).structured_content
        require(
            filtered["notices"]
            == [row for row in all_notices["notices"] if row["function"] == "doctor"],
            "Annotations filter changed notice selection",
        )
        for name, arguments in (
            ("annotations", {"function": ""}),
            ("annotations", {"root": str(root)}),
            ("doctor", {"mode": None}),
            ("doctor", {"mode": "repair"}),
            ("doctor", {"argv": ["true"]}),
        ):
            await rejected(client, name, arguments, -32602)
        for mode in ("inspect", "verify"):
            reply = await client.call_tool(
                "doctor", {} if mode == "inspect" else {"mode": mode}
            )
            require(not reply.is_error, f"Doctor {mode} returned a tool error")
            report = reply.structured_content["result"]
            require(
                report["scope"]["mode"] == mode, "Doctor changed the requested mode"
            )
            require(
                report["coverage"] == "incomplete", "Unexecuted product work was hidden"
            )
            require(
                any(
                    row["requirement"] == "required" and row["status"] == "not_checked"
                    for row in report["findings"]
                ),
                "Doctor omitted required unexecuted behavior",
            )
            if mode == "verify":
                checked = {row["id"]: row for row in report["findings"]}
                for name in (
                    "project-policy",
                    "build-runner-native-identity",
                    "ct-configuration-selection",
                ):
                    require(
                        checked[name]["status"] == "passed"
                        and checked[name]["evidence"],
                        f"Doctor lacks successful actual {name} evidence",
                    )
    require(
        before == {path: snapshot(path) for path in protected},
        "Diagnosis changed native settings",
    )


async def invalidation(root, configuration, options):
    work = root / "tests/.work"
    work.mkdir(mode=0o700, exist_ok=True)
    require(
        work.resolve() == work and stat.S_IMODE(work.stat().st_mode) == 0o700,
        "Disposable verification owner is not private and physical",
    )
    fixture = Path(tempfile.mkdtemp(prefix="own-mcp-", dir=work))
    lifecycle = {"closed": True}
    try:
        owned = [
            "scripts/check-ct.py",
            "scripts/provider_context.py",
            "scripts/integration_doctor.py",
            "requirements/unifi-stubd/mcp.txt",
        ]
        owned.extend(
            str(path.relative_to(root))
            for path in (root / ".agents/skills").glob("unifi-stubd*/SKILL.md")
        )
        for relative in owned:
            target = fixture / relative
            target.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
            shutil.copy2(root / relative, target)
        context = dict(
            configuration["env"],
            PROVIDER_CONSUMER_ROOT=str(fixture),
            PROVIDER_CONSUMER_CONTEXT=json.dumps(
                {
                    "consumer_id": "own-mcp-smoke",
                    "persistence": ".consumer-data",
                    "resources": [],
                }
            ),
        )
        native = dict(
            configuration,
            cwd=str(fixture),
            env=context,
            args=["-B", str(fixture / "scripts/check-ct.py"), "--mcp-annotations"],
        )
        target = fixture / ".agents/skills/unifi-stubd-annotations/SKILL.md"
        content, mode = snapshot(target)
        for change in ("bytes", "mode"):
            async with endpoint(native, lifecycle, options, changing=True) as (
                client,
                _resources,
                events,
            ):
                try:
                    if change == "bytes":
                        target.write_bytes(content + b"\n")
                    else:
                        target.chmod(mode ^ stat.S_IXUSR)
                    await rejected(client, "annotations", {}, -32603)
                    hints = [await asyncio.wait_for(events.get(), 5) for _ in range(2)]
                    require(
                        any(isinstance(row, ResourceUpdated) for row in hints)
                        and any(isinstance(row, ResourcesListChanged) for row in hints),
                        f"{change} invalidation did not publish both hints",
                    )
                finally:
                    target.write_bytes(content)
                    target.chmod(mode)
                await rejected(client, "annotations", {}, -32603)
            async with endpoint(native, lifecycle, options) as (
                client,
                _resources,
                _events,
            ):
                require(
                    not (await client.call_tool("annotations", {})).is_error,
                    "Reconnection did not recover restored current content",
                )
    finally:
        if lifecycle["closed"]:
            shutil.rmtree(fixture)
        else:
            print(f"Unconfirmed SDK shutdown; preserve fixture at {fixture}")


async def cancelled_owned_process(root):
    """Prove subscription cancellation cannot leave the image-build client alive."""
    work = root / "tests/.work"
    work.mkdir(mode=0o700, exist_ok=True)
    require(
        work.resolve() == work and stat.S_IMODE(work.stat().st_mode) == 0o700,
        "Disposable verification owner is not private and physical",
    )
    fixture = Path(tempfile.mkdtemp(prefix="owned-process-", dir=work))
    pid_file = fixture / "pid"
    child = asyncio.create_task(
        run_owned_process(
            root,
            [
                "sh",
                "-c",
                "trap '' TERM; printf '%s\\n' \"$$\" > \"$1\"; exec sleep 30",
                "owned-process",
                str(pid_file),
            ],
        )
    )
    closed = False
    try:
        async with asyncio.timeout(5):
            # The independent child signals through its contained PID file.
            for _attempt in range(500):
                if pid_file.exists() and pid_file.stat().st_size:
                    break
                await asyncio.sleep(0.01)
            else:
                raise TimeoutError("Owned child did not report readiness")
        pid = int(pid_file.read_text())
        require(os.getpgid(pid) == pid, "Child lacks its owned process group")
        child.cancel()
        try:
            async with asyncio.timeout(12):
                await child
        except asyncio.CancelledError:
            closed = True
        else:
            raise AssertionError("Owned child ignored cancellation")
        try:
            os.killpg(pid, 0)
        except ProcessLookupError:
            pass
        else:
            raise AssertionError("Cancelled child process group remains alive")
    finally:
        if not child.done():
            child.cancel()
            try:
                await child
            except asyncio.CancelledError:
                closed = True
        if closed:
            shutil.rmtree(fixture)
        else:
            print(f"Unconfirmed child shutdown; preserve fixture at {fixture}")


async def check(root, configuration):
    native = native_configuration(root, "build-runner")
    _path, data, _mode = native_catalog(native, "build-runner")
    catalog = json.loads(data)
    options = protocol_options(catalog)
    # An unsupported catalog requirement used to be ignored by the fixed mode.
    catalog["required_domain_provider_interface"]["mcp"]["protocol_version"] = (
        "unsupported"
    )
    try:
        protocol_options(catalog)
    except RuntimeError as error:
        require("no protocol downgrade" in str(error), "Missing SDK-owner action")
    else:
        raise AssertionError("Unsupported catalog protocol silently fell back")
    await real_endpoint(root, configuration, options)
    await invalidation(root, configuration, options)
    await cancelled_owned_process(root)
    print(
        "Own SDK discovery, skills, Doctor, preservation, isolated invalidation "
        "and child cancellation passed"
    )


def main():
    root = Path(os.environ["PROVIDER_CONSUMER_ROOT"])
    require(
        root.resolve(strict=True)
        == root
        == Path.cwd()
        == Path(__file__).resolve().parents[1],
        "Use the unchanged native context of this test owner",
    )
    configuration = json.loads(
        (root / ".consumer-data/connections/unifi-stubd/annotations.json").read_bytes()
    )
    require(
        configuration["cwd"] == str(root)
        and configuration["env"]["PROVIDER_CONSUMER_ROOT"] == str(root),
        "Saved own connection belongs to another native context",
    )
    asyncio.run(check(root, configuration))


if __name__ == "__main__":
    main()
