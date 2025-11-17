"use client";

import { UrlDialog } from "./url-dialog";
import { UrlInput } from "./url-input";

export function UrlShortener() {
  return (
    <div className="w-full max-w-4xl mx-auto space-y-8">
      <UrlInput />

      <UrlDialog />
    </div>
  );
}
