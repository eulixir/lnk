"use client";

import { useState } from "react";
import { postShorten } from "@/api/lnk";
import { UrlDialog } from "./url-dialog";
import { UrlInput } from "./url-input";

export function UrlShortener() {
  const [url, setUrl] = useState("");

  const requestShorten = async () => {
    const response = await postShorten({ url });
    console.log("response", response);
    if (response.status === 200) {
      console.log("response.data", response.data);
    }
  };

  return (
    <div className="w-full max-w-4xl mx-auto space-y-8">
      <UrlInput url={url} setUrl={setUrl} requestShorten={requestShorten} />

      <UrlDialog />
    </div>
  );
}
