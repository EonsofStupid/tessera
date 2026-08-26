#!/usr/bin/env python3
"""Grok/Claude PreToolUse gate for Nomen process rules. Fail-open on parse errors."""
import json
import os
import re
import sys

DENY = {
    "decision": "deny",
    "reason": "Nomen process: do not write live secrets, ZITADEL/Authentik product language, or nomen_vault_/nomen_mesh_ tables in this IAM repo. First account is deploy-time env only. web/ uses pnpm.",
}


def allow():
    print(json.dumps({"decision": "allow"}))
    return 0


def deny(reason=None):
    payload = dict(DENY)
    if reason:
        payload["reason"] = reason
    print(json.dumps(payload))
    return 0


def main():
    raw = sys.stdin.read()
    try:
        event = json.loads(raw) if raw.strip() else {}
    except json.JSONDecodeError:
        return allow()
    tool_input = event.get("toolInput") or event.get("tool_input") or {}
    blob = json.dumps(tool_input)
    lower = blob.lower()
    path = str(
        tool_input.get("target_file")
        or tool_input.get("file_path")
        or tool_input.get("path")
        or tool_input.get("filePath")
        or ""
    )
    command = str(tool_input.get("command") or "")
    cwd = event.get("cwd") or event.get("workspaceRoot") or os.getcwd()

    if "jesse@" in lower and "nomen.sh" in lower:
        return deny("Refusing hardcoded operator identity. First account is created at deploy time.")
    if "founder.password" in lower.replace("_", "."):
        return deny("Refusing hardcoded operator identity. First account is created at deploy time.")
    if "password: password1!" in lower:
        return deny("Refusing a password in git. Supply the first-human password at deploy time through the environment.")
    if "firstinstance_org_human_password=" in lower:
        return deny("Refusing a password assignment in git. Supply the first-human password at deploy time through the environment.")
    if "begin openssh" in lower:
        return deny("Refusing a private key in the workspace.")
    if re.search(r"nomen_vault_|nomen_mesh_", lower) and re.search(r"create table|table ", lower):
        return deny("Vault and Mesh stay sibling instances. No nomen_vault_* or nomen_mesh_* tables in this IAM database.")
    if "github.com/zitadel/" in lower:
        return deny("Do not reintroduce github.com/zitadel libraries. Authored packages live under github.com/shippinAI/nomen/.")
    web_scope = "/web/" in path.replace("\\", "/") or path.replace("\\", "/").endswith("/web") or "/web/" in (cwd + "/").replace("\\", "/")
    if web_scope and re.search(r"\b(npm|npx)\b", command):
        return deny("web/ uses pnpm, not npm or npx.")
    return allow()


if __name__ == "__main__":
    sys.exit(main())
