import assert from "node:assert/strict";
import { Buffer } from "node:buffer";
import { once } from "node:events";
import { createServer } from "node:http";
import { ReadableStream } from "node:stream/web";
import test from "node:test";
import { fetchPublicFingerprint } from "../../lab/stub/tools/controller-ui/public-data.mjs";

function streamedResponse(chunks, options = {}) {
  const state = { pulled: 0, cancelled: false };
  const stream = new ReadableStream(
    {
      pull(controller) {
        if (state.pulled === chunks.length) {
          controller.close();
          return;
        }
        controller.enqueue(chunks[state.pulled++]);
      },
      cancel() {
        state.cancelled = true;
      },
    },
    { highWaterMark: 0 },
  );
  const response = new Response(stream, {
    status: 200,
    headers: { "content-type": "application/json" },
    ...options,
  });
  return { response, state };
}

test("public JSON accepts the exact 4 MiB boundary and releases its reader", async (t) => {
  const body = Buffer.from(`"${"x".repeat(4 * 1024 * 1024 - 2)}"`);
  const { response, state } = streamedResponse(
    [body.subarray(0, 1024), body.subarray(1024)],
    { headers: { "content-type": "Application/JSON; charset=utf-8" } },
  );
  t.mock.method(globalThis, "fetch", async (url, options) => {
    assert.equal(url, "https://static.ui.com/fingerprint/ui/public.json");
    assert.equal(options.credentials, "omit");
    assert.ok(options.signal instanceof AbortSignal);
    return response;
  });
  assert.deepEqual(await fetchPublicFingerprint(), body);
  assert.equal(state.pulled, 2);
  assert.equal(response.body.locked, false);
});

test("oversized public JSON is cancelled before the remaining body is read", async (t) => {
  const { response, state } = streamedResponse([
    Buffer.alloc(2 * 1024 * 1024, 32),
    Buffer.alloc(2 * 1024 * 1024, 32),
    Buffer.from(" "),
    Buffer.alloc(4 * 1024 * 1024, 32),
  ]);
  t.mock.method(globalThis, "fetch", async () => response);
  await assert.rejects(fetchPublicFingerprint(), /exceeds 4 MiB/);
  assert.equal(state.pulled, 3);
  assert.equal(state.cancelled, true);
  assert.equal(response.body.locked, false);
});

test("HTTP and media-type failures cancel unread public bodies", async (t) => {
  for (const options of [
    { status: 503 },
    { headers: { "content-type": "text/html" } },
  ]) {
    await t.test(JSON.stringify(options), async (t) => {
      const { response, state } = streamedResponse(
        [Buffer.from("{}")],
        options,
      );
      t.mock.method(globalThis, "fetch", async () => response);
      await assert.rejects(fetchPublicFingerprint(), /not JSON HTTP 200/);
      assert.equal(state.pulled, 0);
      assert.equal(state.cancelled, true);
      assert.equal(response.body.locked, false);
    });
  }
});

test("invalid JSON and invalid UTF-8 are rejected with released readers", async (t) => {
  for (const body of [Buffer.from("{"), Buffer.from([0x22, 0xff, 0x22])]) {
    await t.test(body.toString("hex"), async (t) => {
      const { response } = streamedResponse([body]);
      t.mock.method(globalThis, "fetch", async () => response);
      await assert.rejects(fetchPublicFingerprint());
      assert.equal(response.body.locked, false);
    });
  }
});

test("body read failures retain the original error and release the reader", async (t) => {
  const failure = new Error("public stream interrupted");
  const response = new Response(
    new ReadableStream({
      pull(controller) {
        controller.error(failure);
      },
    }),
    { headers: { "content-type": "application/json" } },
  );
  t.mock.method(globalThis, "fetch", async () => response);
  await assert.rejects(
    fetchPublicFingerprint(),
    (error) =>
      error === failure ||
      (error instanceof AggregateError && error.errors.includes(failure)),
  );
  assert.equal(response.body.locked, false);
});

test("the native public request sends no controller credentials and refuses redirects", async (t) => {
  const requests = [];
  const server = createServer((request, response) => {
    requests.push({ url: request.url, headers: request.headers });
    if (request.url === "/") {
      response.writeHead(302, { location: "/redirected" });
      response.end();
      return;
    }
    response.writeHead(200, { "content-type": "application/json" });
    response.end("{}");
  });
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  t.after(async () => {
    const closed = once(server, "close");
    server.close();
    server.closeAllConnections();
    await closed;
  });
  const nativeFetch = globalThis.fetch;
  t.mock.method(globalThis, "fetch", async (url, options) => {
    assert.equal(url, "https://static.ui.com/fingerprint/ui/public.json");
    return nativeFetch(`http://127.0.0.1:${server.address().port}/`, options);
  });
  await assert.rejects(fetchPublicFingerprint(), TypeError);
  assert.equal(requests.length, 1);
  assert.equal(requests[0].url, "/");
  assert.equal(requests[0].headers.accept, "application/json");
  for (const header of ["cookie", "authorization", "x-csrf-token"])
    assert.equal(requests[0].headers[header], undefined);
});
