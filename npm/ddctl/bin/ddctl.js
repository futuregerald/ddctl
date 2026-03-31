#!/usr/bin/env node

const { execFileSync } = require("child_process");
const path = require("path");

const PLATFORMS = {
  "darwin-arm64": "@ddctl/darwin-arm64",
  "darwin-x64": "@ddctl/darwin-x64",
  "linux-arm64": "@ddctl/linux-arm64",
  "linux-x64": "@ddctl/linux-x64",
  "win32-arm64": "@ddctl/win32-arm64",
  "win32-x64": "@ddctl/win32-x64",
};

const key = `${process.platform}-${process.arch}`;
const pkg = PLATFORMS[key];

if (!pkg) {
  console.error(`ddctl: unsupported platform ${key}`);
  process.exit(1);
}

const bin = process.platform === "win32" ? "ddctl.exe" : "ddctl";

let binPath;
try {
  binPath = path.join(path.dirname(require.resolve(`${pkg}/package.json`)), "bin", bin);
} catch {
  console.error(
    `ddctl: could not find package ${pkg}.\n` +
    `Make sure your package manager installed the platform-specific dependency.\n` +
    `Try: npm install ddctl`
  );
  process.exit(1);
}

try {
  execFileSync(binPath, process.argv.slice(2), { stdio: "inherit" });
} catch (e) {
  if (e.status != null) {
    process.exit(e.status);
  }
  if (e.signal) {
    process.kill(process.pid, e.signal);
  }
  throw e;
}
