// Phase 2's "done when", run against the real consumer rather than a fake one.
// The real consumer, imported rather than reimplemented. A verifier written
// here to match would only ever prove that we agree with ourselves.
import { readFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

const identityModule = process.env.AUTOMATON_IDENTITY_MODULE
  ?? new URL("../../automaton/engine/serve/identity.mjs", import.meta.url).pathname;
const { createIdentity, TokenError } = await import(pathToFileURL(identityModule));

const ISSUER = "http://localhost:8088";
const tok = (n) => readFileSync(new URL(`../.artifacts/${n}.txt`, import.meta.url), "utf8").trim();

// This workspace is ws-0001 and answers to that audience, nothing else.
const id = createIdentity({ issuer: ISSUER, audience: "automaton:ws-0001" });

let failures = 0;
const check = async (name, fn) => {
  try { await fn(); console.log(`  ✓ ${name}`); }
  catch (e) { failures++; console.log(`  ✗ ${name}\n      ${e.message}`); }
};

console.log("\nAutomaton (engine/serve/identity.mjs) verifying Nomen-minted tokens\n");

await check("accepts a token minted for this workspace", async () => {
  const c = await id.verify(tok("tok-ws0001"));
  const want = {
    schema: "shippin.seat-token.v1",
    workspace_id: "ws-0001",
    occupant: "agent",
    basis: "subscription",
  };
  for (const [k, v] of Object.entries(want)) {
    if (c[k] !== v) throw new Error(`claim ${k} = ${JSON.stringify(c[k])}, want ${JSON.stringify(v)}`);
  }
  const scopes = c.authorization?.scopes ?? [];
  for (const s of ["hosting:active", "terminal:advanced", "chat:unified"]) {
    if (!scopes.includes(s)) throw new Error(`missing scope ${s} in ${JSON.stringify(scopes)}`);
  }
  if (c.authorization?.policy_version !== "pol_2026_08_17") throw new Error("policy_version missing");
});

await check("refuses a token minted for a different workspace", async () => {
  try {
    await id.verify(tok("tok-ws0002"));
  } catch (e) {
    if (e instanceof TokenError && e.reason === "wrong_audience") return;
    throw new Error(`refused, but for the wrong reason: ${e.reason} — ${e.message}`);
  }
  throw new Error("ACCEPTED a token minted for ws-0002 — the tenant boundary is not holding");
});

await check("refuses a tampered payload", async () => {
  const [h, p, s] = tok("tok-ws0001").split(".");
  const claims = JSON.parse(Buffer.from(p, "base64url").toString());
  claims.basis = "subscription";
  claims.authorization.scopes.push("hosting:unlimited");
  const forged = `${h}.${Buffer.from(JSON.stringify(claims)).toString("base64url")}.${s}`;
  try { await id.verify(forged); } catch (e) {
    if (e.reason === "bad_signature") return;
    throw new Error(`refused for ${e.reason}, want bad_signature`);
  }
  throw new Error("ACCEPTED a forged token");
});

await check("refuses alg:none over the same claims", async () => {
  const [, p] = tok("tok-ws0001").split(".");
  const header = Buffer.from(JSON.stringify({ alg: "none", typ: "JWT" })).toString("base64url");
  try { await id.verify(`${header}.${p}.`); } catch (e) {
    if (e.reason === "bad_algorithm") return;
    throw new Error(`refused for ${e.reason}, want bad_algorithm`);
  }
  throw new Error("ACCEPTED alg:none");
});

await check("discovery and JWKS are reachable and asymmetric", async () => {
  const d = await id.discovery();
  await id.refresh();
  const st = id.state();
  if (!st.keys) throw new Error("no keys published");
  console.log(`      jwks_uri=${d.jwks_uri} keys=${st.keys} algs=${id.algorithms.join(",")}`);
});

console.log(failures ? `\n${failures} FAILED\n` : "\nall green\n");
process.exit(failures ? 1 : 0);
