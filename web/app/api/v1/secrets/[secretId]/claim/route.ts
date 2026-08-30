import { NextResponse } from "next/server";

const apiBaseURL = process.env.API_BASE_URL ?? "http://127.0.0.1:8080";

export async function POST(_request: Request, { params }: { params: Promise<{ secretId: string }> }) {
  const { secretId } = await params;
  if (!secretId || secretId.includes("/")) {
    return NextResponse.json({ error: "secret unavailable" }, { status: 404, headers: { "cache-control": "no-store" } });
  }

  try {
    const upstream = await fetch(`${apiBaseURL}/api/v1/secrets/${encodeURIComponent(secretId)}/claim`, {
      method: "POST",
      cache: "no-store",
    });
    return new NextResponse(upstream.body, {
      status: upstream.status,
      headers: { "content-type": "application/json", "cache-control": "no-store" },
    });
  } catch {
    return NextResponse.json({ error: "service unavailable" }, { status: 503, headers: { "cache-control": "no-store" } });
  }
}
