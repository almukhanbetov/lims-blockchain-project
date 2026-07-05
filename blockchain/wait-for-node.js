// Checks whether the local Hardhat JSON-RPC endpoint is accepting connections.
// Default: retries every second until it succeeds (used before deploying).
// --once: a single attempt, exits 0/1 immediately (used as the Docker HEALTHCHECK).
const http = require("http");

const once = process.argv.includes("--once");

function check() {
  const req = http.request(
    {
      host: "127.0.0.1",
      port: 8545,
      method: "POST",
      headers: { "Content-Type": "application/json" },
    },
    (res) => {
      res.on("data", () => {});
      res.on("end", () => process.exit(0));
    }
  );
  req.on("error", () => {
    if (once) {
      process.exit(1);
    } else {
      setTimeout(check, 1000);
    }
  });
  req.write(JSON.stringify({ jsonrpc: "2.0", method: "eth_chainId", params: [], id: 1 }));
  req.end();
}

check();
