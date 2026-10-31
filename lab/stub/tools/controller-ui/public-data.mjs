import { Buffer } from "node:buffer";
import { TextDecoder } from "node:util";

export const publicFingerprintURL =
  "https://static.ui.com/fingerprint/ui/public.json";
const maximumBytes = 4 * 1024 * 1024;

/** Fetch public fingerprint JSON without controller credentials or redirects. */
export async function fetchPublicFingerprint() {
  const response = await fetch(publicFingerprintURL, {
    headers: { accept: "application/json" },
    credentials: "omit",
    redirect: "error",
    signal: AbortSignal.timeout(15_000),
  });
  const reader = response.body?.getReader();
  try {
    if (
      response.status !== 200 ||
      !/^application\/json(?:\s*;|$)/i.test(
        response.headers.get("content-type") ?? "",
      ) ||
      !reader
    )
      throw new Error("public fingerprint response is not JSON HTTP 200");
    const chunks = [];
    let length = 0;
    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;
      if (value.byteLength > maximumBytes - length)
        throw new Error("public fingerprint response exceeds 4 MiB");
      length += value.byteLength;
      chunks.push(value);
    }
    const body = Buffer.concat(chunks, length);
    JSON.parse(new TextDecoder("utf-8", { fatal: true }).decode(body));
    return body;
  } catch (error) {
    try {
      await reader?.cancel(error);
    } catch (cleanupError) {
      throw new AggregateError(
        [error, cleanupError],
        "public fingerprint response rejected and cleanup failed",
        { cause: cleanupError },
      );
    }
    throw error;
  } finally {
    reader?.releaseLock();
  }
}
