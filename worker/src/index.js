// R2のファイルとWebプレイヤーを返す。認可はCloudflare Accessに任せ、Access経由のリクエストかだけを見る。
export default {
  async fetch(request, env) {
    // ponytail: ヘッダーの有無だけを見る。Access以外の経路 (workers.dev、プレビューURL) は wrangler.jsonc で塞いでいる。経路が増えたらJWTの署名とaudを検証する
    if (!request.headers.has("Cf-Access-Jwt-Assertion")) {
      return new Response("Forbidden", { status: 403 });
    }
    const { pathname } = new URL(request.url);
    if (pathname === "/data/days") {
      return days(env.BUCKET);
    }
    if (pathname.startsWith("/data/")) {
      return object(env.BUCKET, pathname.slice("/data/".length), request.headers);
    }
    return env.ASSETS.fetch(request);
  },
};

// R2のキーは {date}/{level}/... なので、先頭のprefixが日付の一覧になる。
async function days(bucket) {
  const list = await bucket.list({ delimiter: "/" });
  const dates = list.delimitedPrefixes
    .map((p) => p.slice(0, -1))
    .sort()
    .reverse();
  return Response.json(dates);
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
  console.log("RANGE", JSON.stringify(obj.range), typeof obj.range, obj.range && Object.keys(obj.range));
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
