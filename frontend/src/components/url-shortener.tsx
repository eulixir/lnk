"use client";

import { postShorten } from "@/api/lnk";
import { UrlDialog } from "./url-dialog";
import { UrlInput } from "./url-input";
import { useState } from "react";
import { toast } from "sonner";

export function UrlShortener() {
  const [url, setUrl] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);

  const requestShorten = async () => {
    const response = await postShorten({ url });
    if (response.status === 200) {
      setShortUrl(response.data.short_url ?? "");
      setDialogOpen(true);
    } else {
      toast.error("Failed to shorten URL");
    }
  };

  return (
    <div className="w-full max-w-4xl mx-auto space-y-8">
      <UrlInput url={url} setUrl={setUrl} requestShorten={requestShorten} />

      <UrlDialog
        shortUrl={shortUrl}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
      />
    </div>
  );
}
