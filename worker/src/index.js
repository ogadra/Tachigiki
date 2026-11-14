export default {
  async fetch(request, env) {
    if (!cameThroughAccess(request)) {
      return new Response("Forbidden", { status: 403 });
    }
    const { pathname } = new URL(request.url);
    if (pathname === "/data/days") {
      return listDates(env.BUCKET);
    }
    if (pathname.startsWith("/data/")) {
      return object(env.BUCKET, pathname.slice("/data/".length), request.headers);
    }
    return env.ASSETS.fetch(request);
  },
};

function cameThroughAccess(request) {
  return request.headers.has("Cf-Access-Jwt-Assertion");
}

async function listDates(bucket) {
  const topLevel = await bucket.list({ delimiter: "/" });
  const dates = topLevel.delimitedPrefixes.map((dateDir) => dateDir.replace(/\/$/, ""));
  return Response.json(dates.sort().reverse());
}

async function object(bucket, key, headers) {
  const obj = await bucket.get(key, { range: headers, onlyIf: headers });
  if (obj === null) {
    return new Response("Not Found", { status: 404 });
  }
  const res = new Headers({ etag: obj.httpEtag, "accept-ranges": "bytes" });
  obj.writeHttpMetadata(res);
  if (!("body" in obj)) {
    return new Response(null, { status: 304, headers: res });
  }
  if (!headers.has("range")) {
    return new Response(obj.body, { headers: res });
  }
  const { offset, length } = byteRange(obj.range, obj.size);
  res.set("content-range", `bytes ${offset}-${offset + length - 1}/${obj.size}`);
  res.set("content-length", String(length));
  return new Response(obj.body, { status: 206, headers: res });
}

export function byteRange(range, size) {
  if (range.suffix !== undefined) {
    return { offset: size - range.suffix, length: range.suffix };
  }
  const offset = range.offset ?? 0;
  return { offset, length: range.length ?? size - offset };
}
