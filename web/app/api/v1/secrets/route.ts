import { NextResponse } from "next/server";

const apiBaseURL = process.env.API_BASE_URL ?? "http://127.0.0.1:8080";

export async function POST(request: Request) {
  try {
    const body = await request.text();
    const upstream = await fetch(`${apiBaseURL}/api/v1/secrets`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body,
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
