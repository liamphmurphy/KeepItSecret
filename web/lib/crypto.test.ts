import { describe, expect, it } from "vitest";
import { base64URLToBytes, bytesToBase64URL, decryptSecret, encryptSecret } from "./crypto";

describe("crypto helpers", () => {
  it("round trips a secret with URL-safe encoding", async () => {
    const encrypted = await encryptSecret("a private note", "secret-id");
    await expect(decryptSecret(encrypted.ciphertext, encrypted.nonce, encrypted.key, "secret-id")).resolves.toBe("a private note");
    expect(bytesToBase64URL(base64URLToBytes(encrypted.key))).toBe(encrypted.key);
  });

  it("rejects altered ciphertext and the wrong secret id", async () => {
    const encrypted = await encryptSecret("do not share", "secret-id");
    const altered = `${encrypted.ciphertext[0] === "A" ? "B" : "A"}${encrypted.ciphertext.slice(1)}`;
    await expect(decryptSecret(altered, encrypted.nonce, encrypted.key, "secret-id")).rejects.toThrow();
    await expect(decryptSecret(encrypted.ciphertext, encrypted.nonce, encrypted.key, "other-id")).rejects.toThrow();
  });

  it("rejects malformed base64url values", () => {
    expect(() => base64URLToBytes("not valid!" )).toThrow();
  });
});
