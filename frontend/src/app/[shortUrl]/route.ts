import { NextResponse } from "next/server";
import { getShortUrl } from "@/api/lnk";

export async function GET(
  _request: Request,
  context: { params: Promise<{ shortUrl: string }> | { shortUrl: string } },
) {
  const params = context.params;
  const resolvedParams = params instanceof Promise ? await params : params;
  const shortUrl = resolvedParams?.shortUrl;

  if (!shortUrl) {
    return NextResponse.json(
      { error: "Short URL parameter is required" },
      { status: 400 },
    );
  }

  try {
    const response = await getShortUrl(shortUrl);

    if (response.status === 308) {
      const location = response.headers.get("Location");
      if (location) {
        return NextResponse.redirect(location, 308);
      }

      const data = response.data as { [key: string]: string };
      const originalUrl =
        data.original_url || data.url || Object.values(data)[0];
      if (originalUrl) {
        return NextResponse.redirect(originalUrl, 308);
      }
    }

    if (response.status === 404) {
      return NextResponse.json(
        { error: "Short URL not found" },
        { status: 404 },
      );
    }

    return NextResponse.json(
      { error: "Failed to redirect", status: response.status },
      { status: response.status },
    );
  } catch (error) {
    console.error("Error fetching short URL:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 },
    );
  }
}
