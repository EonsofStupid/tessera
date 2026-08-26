# Public deploy

`public.compose.yaml` is the stranger-bootable public edition: PostgreSQL plus
Nomen, Redis off, `NOMEN_EDITION=public`.

Secrets stay out of git. Set `NOMEN_MASTERKEY` (32 characters) in the shell.
Pass a first-human password through the process environment if you want a
password owner; otherwise omit it and enroll a passkey owner with
`NOMEN_BOOTSTRAP_AUTHORITY`. Do not write that password into this file.

Hosted demo uses the same image with `NOMEN_DEMO_CAPS=true`.
Enterprise is the same image with `NOMEN_EDITION=enterprise`.
