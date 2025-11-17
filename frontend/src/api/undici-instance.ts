const baseURL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const getFetch = async (): Promise<typeof fetch> => {
  if (typeof window === "undefined") {
    try {
      const { fetch: undiciFetch } = await import("undici");
      return undiciFetch as unknown as typeof fetch;
    } catch {
      return globalThis.fetch;
    }
  } else {
    return window.fetch;
  }
};

export const customInstance = async <T>(
  url: string,
  options?: RequestInit,
): Promise<{
  data: T;
  status: number;
  statusText: string;
  headers: Headers;
}> => {
  const fullUrl = `${baseURL}${url}`;

  const fetchImpl = await getFetch();
  const response = await fetchImpl(fullUrl, {
    ...options,
    method: options?.method || "GET",
  });

  let responseData: T;
  const contentType = response.headers.get("content-type");
  if (contentType?.includes("application/json")) {
    responseData = await response.json();
  } else {
    const text = await response.text();
    try {
      responseData = JSON.parse(text) as T;
    } catch {
      responseData = text as unknown as T;
    }
  }

  return {
    data: responseData,
    status: response.status,
    statusText: response.statusText,
    headers: response.headers,
  };
};
