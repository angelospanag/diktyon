import { type NextRequest, NextResponse } from "next/server";

const BACKEND = process.env.API_URL ?? "http://localhost:8080";

async function proxy(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> },
) {
  const { path } = await params;
  const url = `${BACKEND}/api/${path.join("/")}${request.nextUrl.search}`;
  try {
    const upstream = await fetch(url);
    const data = await upstream.json();
    return NextResponse.json(data, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "upstream unreachable" },
      { status: 502 },
    );
  }
}

export { proxy as GET };
