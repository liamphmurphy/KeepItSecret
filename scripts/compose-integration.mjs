const webURL = process.env.TEST_WEB_URL ?? "http://127.0.0.1:13000";
const encoder = new TextEncoder();

function base64URL(bytes) {
  return Buffer.from(bytes).toString("base64url");
}

async function createPayload(secretId, plaintext) {
  const key = await crypto.subtle.generateKey({ name: "AES-GCM", length: 256 }, true, ["encrypt", "decrypt"]);
  const nonce = crypto.getRandomValues(new Uint8Array(12));
  const additionalData = encoder.encode(`keepitsecret:1:${secretId}`);
  const ciphertext = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: nonce, additionalData, tagLength: 128 },
    key,
    encoder.encode(plaintext),
  );
  const exportedKey = await crypto.subtle.exportKey("raw", key);

  return {
    key: await crypto.subtle.importKey("raw", exportedKey, { name: "AES-GCM" }, false, ["decrypt"]),
    payload: {
      secretId,
      version: 1,
      ciphertext: base64URL(new Uint8Array(ciphertext)),
      nonce: base64URL(nonce),
      expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
    },
  };
}

async function expectStatus(response, want, operation) {
  if (response.status !== want) {
    const body = await response.text();
    throw new Error(`${operation} returned ${response.status}, want ${want}: ${body}`);
  }
}

async function main() {
  const secretId = base64URL(crypto.getRandomValues(new Uint8Array(16)));
  const plaintext = "compose integration secret";
  const { key, payload } = await createPayload(secretId, plaintext);

  const createResponse = await fetch(`${webURL}/api/v1/secrets`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(payload),
  });
  await expectStatus(createResponse, 201, "create secret");

  const claimURL = `${webURL}/api/v1/secrets/${encodeURIComponent(secretId)}/claim`;
  const claimResponse = await fetch(claimURL, { method: "POST" });
  await expectStatus(claimResponse, 200, "claim secret");
  const claimed = await claimResponse.json();
  const decrypted = await crypto.subtle.decrypt(
    {
      name: "AES-GCM",
      iv: Buffer.from(claimed.nonce, "base64url"),
      additionalData: encoder.encode(`keepitsecret:1:${secretId}`),
      tagLength: 128,
    },
    key,
    Buffer.from(claimed.ciphertext, "base64url"),
  );
  if (new TextDecoder().decode(decrypted) !== plaintext) {
    throw new Error("decrypted secret did not match the created secret");
  }

  const secondClaimResponse = await fetch(claimURL, { method: "POST" });
  await expectStatus(secondClaimResponse, 404, "second claim");
  console.log("Compose integration smoke test passed");
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
