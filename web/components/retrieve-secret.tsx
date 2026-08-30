"use client";

import { useEffect, useState } from "react";
import { decryptSecret, SECRET_VERSION } from "@/lib/crypto";

type ViewState = "loading" | "secret" | "missing-key" | "unavailable" | "invalid" | "error";

export function RetrieveSecret({ secretId }: { secretId: string }) {
  const [state, setState] = useState<ViewState>("loading");
  const [secret, setSecret] = useState("");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let cancelled = false;
    async function claim() {
      const key = window.location.hash.slice(1);
      if (!key) { setState("missing-key"); return; }
      try {
        const response = await fetch(`/api/v1/secrets/${encodeURIComponent(secretId)}/claim`, { method: "POST", cache: "no-store" });
        if (cancelled) return;
        if (response.status === 404) { setState("unavailable"); return; }
        if (!response.ok) throw new Error("claim failed");
        const payload = await response.json() as { version: number; ciphertext: string; nonce: string };
        if (payload.version !== SECRET_VERSION) throw new Error("unsupported version");
        const plaintext = await decryptSecret(payload.ciphertext, payload.nonce, key, secretId);
        if (!cancelled) { setSecret(plaintext); setState("secret"); window.history.replaceState(null, "", window.location.pathname); }
      } catch {
        if (!cancelled) setState("invalid");
      }
    }
    void claim();
    return () => { cancelled = true; };
  }, [secretId]);

  async function copySecret() { await navigator.clipboard.writeText(secret); setCopied(true); }

  return <section className="content-column retrieve-column" aria-live="polite">
    {state === "loading" && <StateBlock title="Opening your secret…" detail="The link is being claimed securely." />}
    {state === "secret" && <div className="result-block"><h1>Your secret</h1><textarea className="revealed-secret" readOnly value={secret} aria-label="Decrypted secret" /><div className="actions"><button className="button-primary" type="button" onClick={copySecret}>{copied ? "Copied" : "Copy secret"}</button><a className="text-button" href="/">Make a link</a></div><p className="form-note">This link has now been used. Treat copied text as sensitive.</p></div>}
    {state === "missing-key" && <StateBlock title="This link is incomplete" detail="The decryption key is missing from the URL, so the secret cannot be opened." />}
    {state === "unavailable" && <StateBlock title="This secret is no longer available" detail="It may have already been opened, expired, or the link may be incorrect." />}
    {state === "invalid" && <StateBlock title="This secret could not be verified" detail="The encrypted content was modified or the link is not valid." />}
    {state === "error" && <StateBlock title="Something went wrong" detail="Check your connection and try again later." />}
  </section>;
}

function StateBlock({ title, detail }: { title: string; detail: string }) { return <div className="state-block"><h1>{title}</h1><p>{detail}</p><a className="button-secondary" href="/">Create a new link</a></div>; }
