"use client";

import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useState,
} from "react";
import { toast } from "sonner";
import { postShorten } from "@/api/lnk";

interface UrlShortenerContextType {
  url: string;
  setUrl: (url: string) => void;
  shortUrl: string;
  originalUrl: string;
  dialogOpen: boolean;
  setDialogOpen: (open: boolean) => void;
  requestShorten: () => Promise<void>;
  clearUrl: () => void;
}

const UrlShortenerContext = createContext<UrlShortenerContextType | undefined>(
  undefined,
);

export function UrlShortenerProvider({ children }: { children: ReactNode }) {
  const [url, setUrl] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [originalUrl, setOriginalUrl] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);

  const requestShorten = useCallback(async () => {
    if (!url) return;

    try {
      const response = await postShorten({ url });
      if (response.status !== 200) {
        toast.error("Failed to shorten URL");
        return;
      }

      setOriginalUrl(url);
      setShortUrl(response.data.short_url ?? "");
      setDialogOpen(true);
      setUrl("");
      toast.success("URL shortened successfully!");
      return;
    } catch (_error) {
      toast.error("Failed to shorten URL");
    }
  }, [url]);

  const clearUrl = useCallback(() => {
    setUrl("");
  }, []);

  return (
    <UrlShortenerContext.Provider
      value={{
        url,
        setUrl,
        shortUrl,
        originalUrl,
        dialogOpen,
        setDialogOpen,
        requestShorten,
        clearUrl,
      }}
    >
      {children}
    </UrlShortenerContext.Provider>
  );
}

export function useUrlShortener() {
  const context = useContext(UrlShortenerContext);
  if (context === undefined) {
    throw new Error("useUrlShortener must be used within UrlShortenerProvider");
  }
  return context;
}
