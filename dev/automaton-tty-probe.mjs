#!/usr/bin/env node
/**
 * Live Tessera -> Automaton -> mato-pty conformance.
 *
 * The token is read from the ignored runtime directory and is never printed.
 * This intentionally implements the small client half of RFC 6455 instead of
 * importing Automaton internals: the probe should catch a broken public wire,
 * not agree with the implementation by construction.
 */
import assert from "node:assert/strict";
import { randomBytes } from "node:crypto";
import { readFile } from "node:fs/promises";
import { connect } from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const TOKEN_PATH = process.env.TESSERA_AUTOMATON_TOKEN_FILE || path.join(ROOT, ".artifacts", "tok-ws0001.txt");
const BASE = new URL(process.env.AUTOMATON_URL || "http://127.0.0.1:8111");
const SESSION = process.env.AUTOMATON_TTY_SESSION || "tessera-first-tty";
const MARKER = "TESSERA_TTY_OK";

assert.equal(BASE.protocol, "http:", "the local conformance endpoint must use http on loopback");
assert.ok(["127.0.0.1", "localhost", "::1"].includes(BASE.hostname), "refusing to send a live seat token off loopback");

const token = (await readFile(TOKEN_PATH, "utf8")).trim();
assert.equal(token.split(".").length, 3, "the runtime token is not a compact JWS");

function ptyFrame(kind, payload) {
  const body = Buffer.isBuffer(payload) ? payload : Buffer.from(payload);
  const out = Buffer.allocUnsafe(5 + body.length);
  out[0] = kind;
  out.writeUInt32BE(body.length, 1);
  body.copy(out, 5);
  return out;
}

function maskedWebSocketFrame(opcode, payload) {
  const body = Buffer.isBuffer(payload) ? payload : Buffer.from(payload);
  assert.ok(body.length < 65_536, "probe frame unexpectedly large");
  const extended = body.length >= 126;
  const head = Buffer.allocUnsafe((extended ? 4 : 2) + 4);
  head[0] = 0x80 | opcode;
  head[1] = 0x80 | (extended ? 126 : body.length);
  let offset = 2;
  if (extended) {
    head.writeUInt16BE(body.length, offset);
    offset += 2;
  }
  const mask = randomBytes(4);
  mask.copy(head, offset);
  const encoded = Buffer.allocUnsafe(body.length);
  for (let i = 0; i < body.length; i += 1) encoded[i] = body[i] ^ mask[i % 4];
  return Buffer.concat([head, encoded]);
}

function pullWebSocketFrame(buffer) {
  if (buffer.length < 2) return null;
  const opcode = buffer[0] & 0x0f;
  const masked = Boolean(buffer[1] & 0x80);
  let length = buffer[1] & 0x7f;
  let offset = 2;
  if (length === 126) {
    if (buffer.length < 4) return null;
    length = buffer.readUInt16BE(2);
    offset = 4;
  } else if (length === 127) {
    if (buffer.length < 10) return null;
    const wide = buffer.readBigUInt64BE(2);
    assert.ok(wide <= BigInt(Number.MAX_SAFE_INTEGER), "server frame is too large");
    length = Number(wide);
    offset = 10;
  }
  assert.equal(masked, false, "a server WebSocket frame must not be masked");
  if (buffer.length < offset + length) return null;
  return { opcode, payload: buffer.subarray(offset, offset + length), consumed: offset + length };
}

function pullPtyFrame(buffer) {
  if (buffer.length < 5) return null;
  const length = buffer.readUInt32BE(1);
  assert.ok(length <= 1 << 20, "PTY frame exceeds its protocol ceiling");
  if (buffer.length < 5 + length) return null;
  return { kind: buffer[0], payload: buffer.subarray(5, 5 + length), consumed: 5 + length };
}

async function upgrade(cookie = "") {
  const socket = connect(Number(BASE.port || 80), BASE.hostname);
  await new Promise((resolve, reject) => socket.once("connect", resolve).once("error", reject));
  const key = randomBytes(16).toString("base64");
  const headers = [
    "GET /api/terminal HTTP/1.1",
    `Host: ${BASE.host}`,
    "Upgrade: websocket",
    "Connection: Upgrade",
    `Sec-WebSocket-Key: ${key}`,
    "Sec-WebSocket-Version: 13",
  ];
  if (cookie) headers.push(`Cookie: ${cookie}`);
  socket.write(`${headers.join("\r\n")}\r\n\r\n`);

  return new Promise((resolve, reject) => {
    let received = Buffer.alloc(0);
    const timer = setTimeout(() => reject(new Error("timed out waiting for WebSocket upgrade")), 5_000);
    const onData = (chunk) => {
      received = Buffer.concat([received, chunk]);
      const end = received.indexOf("\r\n\r\n");
      if (end === -1) return;
      clearTimeout(timer);
      socket.off("data", onData);
      const header = received.subarray(0, end).toString("utf8");
      const status = Number(header.match(/^HTTP\/1\.1 (\d{3})/)?.[1] || 0);
      resolve({ socket, status, remainder: received.subarray(end + 4) });
    };
    socket.on("data", onData);
    socket.once("error", reject);
  });
}

const anonymous = await upgrade();
assert.equal(anonymous.status, 401, "an unauthenticated terminal upgrade must be 401");
anonymous.socket.destroy();
console.log("  ✓ anonymous terminal upgrade refused with 401");

const auth = await fetch(new URL("/auth/session", BASE), {
  method: "POST",
  headers: { "content-type": "application/json" },
  body: JSON.stringify({ token }),
});
const session = await auth.json();
assert.equal(
  auth.status,
  200,
  `Automaton session exchange answered ${auth.status} (${session.reason || session.error || "unknown refusal"})`,
);
assert.equal(session.ok, true);
assert.ok(session.expiresIn > 0 && session.expiresIn <= 900, "cookie lifetime must be positive and no longer than the seat token");
assert.ok(session.scopes.includes("hosting:active"));
assert.ok(session.scopes.includes("terminal:advanced"));

const setCookie = auth.headers.get("set-cookie") || "";
assert.match(setCookie, /^automaton_token=/);
assert.match(setCookie, /; HttpOnly(?:;|$)/i);
assert.match(setCookie, /; SameSite=Lax(?:;|$)/i);
const cookie = setCookie.split(";", 1)[0];
console.log(`  ✓ Tessera token became an HTTP-only Automaton session (${session.expiresIn}s maximum)`);

const caps = await fetch(new URL("/api/caps", BASE), { headers: { cookie } });
assert.equal(caps.status, 200, `authenticated capability request answered ${caps.status}`);
console.log("  ✓ the authenticated Automaton API accepted the same session");

const ws = await upgrade(cookie);
assert.equal(ws.status, 101, `authenticated terminal upgrade answered ${ws.status}`);
console.log("  ✓ entitled terminal upgrade accepted with 101");

let websocketBuffer = ws.remainder;
let terminalBuffer = Buffer.alloc(0);
let hello = null;
let output = Buffer.alloc(0);
let commandSent = false;

const attach = ptyFrame(0x01, JSON.stringify({ session: SESSION, cols: 100, rows: 30 }));
ws.socket.write(maskedWebSocketFrame(0x02, attach));

await new Promise((resolve, reject) => {
  const timer = setTimeout(() => reject(new Error("timed out waiting for live PTY output")), 15_000);

  const inspect = () => {
    for (;;) {
      const frame = pullWebSocketFrame(websocketBuffer);
      if (!frame) break;
      websocketBuffer = websocketBuffer.subarray(frame.consumed);
      if (frame.opcode === 0x08) return reject(new Error("terminal WebSocket closed before the marker arrived"));
      if (frame.opcode !== 0x02) continue;
      terminalBuffer = Buffer.concat([terminalBuffer, frame.payload]);
      for (;;) {
        const message = pullPtyFrame(terminalBuffer);
        if (!message) break;
        terminalBuffer = terminalBuffer.subarray(message.consumed);
        if (message.kind === 0x81) {
          hello = JSON.parse(message.payload.toString("utf8"));
          if (!commandSent) {
            commandSent = true;
            const input = ptyFrame(0x02, Buffer.from(`printf '${MARKER}\\n'\r`));
            ws.socket.write(maskedWebSocketFrame(0x02, input));
          }
        } else if (message.kind === 0x82) {
          output = Buffer.concat([output, message.payload]);
          if (output.includes(Buffer.from(MARKER))) {
            clearTimeout(timer);
            resolve();
            return;
          }
        }
      }
    }
  };

  ws.socket.on("data", (chunk) => {
    websocketBuffer = Buffer.concat([websocketBuffer, chunk]);
    inspect();
  });
  ws.socket.once("error", reject);
  inspect();
});

assert.equal(hello?.session, SESSION);
assert.equal(typeof hello?.created, "boolean");
ws.socket.destroy();
console.log(`  ✓ attached ${SESSION} (created=${hello.created}) and observed live command output`);
console.log("\nTessera → Automaton session/TTY conformance passed");
