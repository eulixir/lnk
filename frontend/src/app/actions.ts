"use server";

// Simple URL shortener using a hash function
// In production, you'd want to use a database to store mappings
function generateShortCode(url: string): string {
  let hash = 0;
  const timestamp = Date.now().toString(36);

  for (let i = 0; i < url.length; i++) {
    const char = url.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash;
  }

  const shortCode =
    Math.abs(hash).toString(36).substring(0, 6) + timestamp.substring(0, 3);
  return shortCode;
}

export async function shortenUrl(url: string) {
  // Validate URL
  try {
    new URL(url);
  } catch {
    throw new Error("Invalid URL");
  }

  // Generate short code
  const shortCode = generateShortCode(url);

  // In production, you would save this to a database
  // For now, we'll just return a mock shortened URL
  const baseUrl = process.env.NEXT_PUBLIC_BASE_URL || "https://lnk.sh";
  const shortUrl = `${baseUrl}/${shortCode}`;

  // Simulate network delay
  await new Promise((resolve) => setTimeout(resolve, 500));

  return {
    shortUrl,
    originalUrl: url,
    shortCode,
  };
}
