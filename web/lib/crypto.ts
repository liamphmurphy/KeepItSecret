const encoder = new TextEncoder();
const decoder = new TextDecoder();

function asBufferSource(value: Uint8Array): BufferSource {
  return value as unknown as BufferSource;
}

export const SECRET_VERSION = 1;

export function bytesToBase64URL(bytes: Uint8Array): string {
  let binary = "";
  bytes.forEach((byte) => { binary += String.fromCharCode(byte); });
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

export function base64URLToBytes(value: string): Uint8Array {
  if (!/^[A-Za-z0-9_-]+$/.test(value)) throw new Error("invalid base64url value");
  const padded = value.replace(/-/g, "+").replace(/_/g, "/") + "=".repeat((4 - value.length % 4) % 4);
  const binary = atob(padded);
  return Uint8Array.from(binary, (character) => character.charCodeAt(0));
}

function additionalData(secretId: string): Uint8Array {
  return encoder.encode(`keepitsecret:${SECRET_VERSION}:${secretId}`);
}

export async function encryptSecret(plaintext: string, secretId: string) {
  const key = await crypto.subtle.generateKey({ name: "AES-GCM", length: 256 }, true, ["encrypt", "decrypt"]);
  const nonce = crypto.getRandomValues(new Uint8Array(12));
  const encrypted = await crypto.subtle.encrypt({ name: "AES-GCM", iv: asBufferSource(nonce), additionalData: asBufferSource(additionalData(secretId)), tagLength: 128 }, key, asBufferSource(encoder.encode(plaintext)));
  const exportedKey = await crypto.subtle.exportKey("raw", key);
  return { key: bytesToBase64URL(new Uint8Array(exportedKey)), nonce: bytesToBase64URL(nonce), ciphertext: bytesToBase64URL(new Uint8Array(encrypted)) };
}

export async function decryptSecret(ciphertext: string, nonce: string, keyValue: string, secretId: string): Promise<string> {
  const key = await crypto.subtle.importKey("raw", asBufferSource(base64URLToBytes(keyValue)), { name: "AES-GCM" }, false, ["decrypt"]);
  const decrypted = await crypto.subtle.decrypt({ name: "AES-GCM", iv: asBufferSource(base64URLToBytes(nonce)), additionalData: asBufferSource(additionalData(secretId)), tagLength: 128 }, key, asBufferSource(base64URLToBytes(ciphertext)));
  return decoder.decode(decrypted);
}
