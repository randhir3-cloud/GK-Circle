import http from "http";
import { exec } from "child_process";
import path from "path";

// Start a local dummy server returning 404.
const server = http.createServer((req, res) => {
  res.writeHead(404, { "Content-Type": "application/json" });
  res.end(JSON.stringify({ status: "error", message: "Not Found" }));
});

server.listen(0, "127.0.0.1", () => {
  const address = server.address();
  const baseUrl = `http://${address.address}:${address.port}`;
  console.log(`Dummy server running at ${baseUrl}`);

  // Run the smoke test pointing to our dummy server.
  const scriptPath = path.join("tests", "e2e", "railway-smoke-verify.mjs");
  const proc = exec(`node ${scriptPath}`, {
    env: { ...process.env, BASE_URL: baseUrl },
  });

  let output = "";
  proc.stdout.on("data", (data) => {
    output += data;
  });
  proc.stderr.on("data", (data) => {
    output += data;
  });

  proc.on("close", (code) => {
    server.close(() => {
      console.log("--- Dummy Server Output ---");
      console.log(output);
      console.log(`Exit code: ${code}`);
      if (code === 1) {
        console.log("TEST_PASSED: Script exited with 1 on 404 server");
        process.exit(0);
      } else {
        console.error(
          "TEST_FAILED: Script exited with 0 or non-1 on 404 server"
        );
        process.exit(1);
      }
    });
  });
});
