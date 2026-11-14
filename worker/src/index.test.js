import assert from "node:assert/strict";
import { test } from "node:test";
import worker from "./index.js";

const access = { "Cf-Access-Jwt-Assertion": "x" };

function bucket(obj) {
  return { get: async () => obj };
}

test("rejects requests that did not come through Access", async () => {
  const res = await worker.fetch(new Request("https://t/data/a"), { BUCKET: bucket(null) });
  assert.equal(res.status, 403);
});

test("answers a Range request with 206 and Content-Range", async () => {
  const obj = {
    size: 100,
    range: { offset: 10 },
    body: "rest",
    httpEtag: '"e"',
    writeHttpMetadata: () => {},
  };
  const req = new Request("https://t/data/2026-11-14/a2/audio.m4a", { headers: { ...access, range: "bytes=10-" } });
  const res = await worker.fetch(req, { BUCKET: bucket(obj) });
  assert.equal(res.status, 206);
  assert.equal(res.headers.get("content-range"), "bytes 10-99/100");
});
