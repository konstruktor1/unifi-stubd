// The facade is deliberately tiny and synchronous outside of request-body
// collection; UniFi Core only expects JSON responses from the observed calls.
function readBody(req) {
  return new Promise((resolve, reject) => {
    let data = "";
    req.setEncoding("utf8");
    req.on("data", (chunk) => {
      data += chunk;
    });
    req.on("end", () => resolve(data));
    req.on("error", reject);
  });
}

// Always include content-length because UniFi Core's proxying path is easier to
// inspect when each synthetic response is a complete, deterministic JSON frame.
function sendJson(res, status, payload) {
  const body = JSON.stringify(payload);
  res.writeHead(status, {
    "content-type": "application/json",
    "content-length": Buffer.byteLength(body),
  });
  res.end(body);
}

module.exports = {
  readBody,
  sendJson,
};
