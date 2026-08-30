"use client";

import { FormEvent, useState } from "react";
import { bytesToBase64URL, encryptSecret, SECRET_VERSION } from "@/lib/crypto";

const expirationOptions = [
  { label: "5 minutes", minutes: 5 },
  { label: "1 hour", minutes: 60 },
  { label: "24 hours", minutes: 24 * 60 },
];

export function CreateSecret() {
  const [message, setMessage] = useState("");
  const [expiration, setExpiration] = useState(60);
  const [shareURL, setShareURL] = useState("");
  const [copied, setCopied] = useState(false);
  const [status, setStatus] = useState<"idle" | "working" | "error">("idle");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!message.trim()) return;
    setStatus("working");
    setShareURL("");
    setCopied(false);
    try {
      const secretId = bytesToBase64URL(crypto.getRandomValues(new Uint8Array(16)));
      const encrypted = await encryptSecret(message, secretId);
      const expiresAt = new Date(Date.now() + expiration * 60_000).toISOString();
      const response = await fetch("/api/v1/secrets", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ secretId, version: SECRET_VERSION, ...encrypted, expiresAt }),
      });
      if (!response.ok) throw new Error("create failed");
      setShareURL(`${window.location.origin}/s/${secretId}#${encrypted.key}`);
      setMessage("");
      setStatus("idle");
    } catch {
      setStatus("error");
    }
  }

  async function copyLink() {
    await navigator.clipboard.writeText(shareURL);
    setCopied(true);
  }

  return <section className="content-column" aria-labelledby="create-title">
    <div className="page-intro"><h1 id="create-title">Share something once.</h1><p>Write a note, make a link, and it disappears after the first successful read.</p></div>
    {!shareURL ? <form className="secret-form" onSubmit={submit}>
      <label htmlFor="secret-message">Your secret</label>
      <textarea id="secret-message" value={message} onChange={(event) => setMessage(event.target.value)} placeholder="Paste a password, private note, or anything else…" rows={9} maxLength={1_000_000} autoFocus />
      <div className="form-row"><div><label htmlFor="expiration">Expires after</label><select id="expiration" value={expiration} onChange={(event) => setExpiration(Number(event.target.value))}>{expirationOptions.map((option) => <option key={option.minutes} value={option.minutes}>{option.label}</option>)}</select></div><button className="button-primary" type="submit" disabled={!message.trim() || status === "working"}>{status === "working" ? "Encrypting…" : "Create private link"}</button></div>
      {status === "error" && <p className="form-error" role="alert">That did not work. Check your connection and try again.</p>}
      <p className="form-note">Encryption happens in your browser. The server receives only the encrypted text.</p>
    </form> : <div className="result-block" role="status"><h2>Your link is ready</h2><p>Anyone with this link can open the secret once. Keep it private.</p><div className="link-row"><input aria-label="Share link" readOnly value={shareURL} onFocus={(event) => event.currentTarget.select()} /><button className="button-primary" type="button" onClick={copyLink}>{copied ? "Copied" : "Copy link"}</button></div><button className="text-button" type="button" onClick={() => setShareURL("")}>Create another</button></div>}
  </section>;
}
