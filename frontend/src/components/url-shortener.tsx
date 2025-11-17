"use client";

import { useState } from "react";
import { UrlDialog } from "./url-dialog";
import { UrlInput } from "./url-input";

export function UrlShortener() {
  const [url, setUrl] = useState("");

  const requestShorten = async () => {
    // const response = await shortenUrl(url);
    console.log("banana");
  };

  return (
    <div className="w-full max-w-4xl mx-auto space-y-8">
      <UrlInput url={url} setUrl={setUrl} requestShorten={requestShorten} />

      <UrlDialog />
    </div>
  );
}
