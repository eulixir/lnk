"use client";

import { CheckCircle2, Copy, Link2, Loader2, QrCode } from "lucide-react";
import Image from "next/image";
import { useState } from "react";
import { toast } from "sonner";
import { shortenUrl } from "@/app/actions";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

export function UrlShortener() {
  const [url, setUrl] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      new URL(url);
    } catch {
      toast.error("Invalid URL", {
        description:
          "Please enter a valid URL starting with http:// or https://",
      });
      return;
    }

    setIsLoading(true);

    try {
      const result = await shortenUrl(url);
      setShortUrl(result.shortUrl);
      setDialogOpen(true);
      setCopied(false);
    } catch {
      toast.error("Error", {
        description: "Failed to shorten URL. Please try again.",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopy = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      toast.success("Copied!", {
        description: "Short URL copied to clipboard",
      });
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Copy failed", {
        description: "Please copy the URL manually",
      });
    }
  };

  return (
    <div className="w-full max-w-4xl mx-auto space-y-8">
      <div className="bg-card rounded-2xl border border-border p-8 shadow-lg">
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="flex gap-2">
            <div className="relative flex-1">
              <Link2 className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground size-5" />
              <Input
                type="url"
                placeholder="Enter the link here"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                className="pl-12 h-14 text-base bg-background border-2"
                disabled={isLoading}
                required
              />
            </div>
            <Button
              type="submit"
              size="lg"
              className="h-14 px-8 bg-blue-600 hover:bg-blue-700 text-white font-semibold"
              disabled={isLoading || !url}
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 size-5 animate-spin" />
                  Shortening...
                </>
              ) : (
                "Shorten Now!"
              )}
            </Button>
          </div>
        </form>
      </div>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <CheckCircle2 className="size-6 text-green-500" />
              Your Short URL is Ready!
            </DialogTitle>
            <DialogDescription>
              Share this shortened link anywhere
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4 mt-4">
            <div className="flex items-center gap-2">
              <Input readOnly value={shortUrl} className="font-mono bg-muted" />
              <Button
                size="icon"
                variant="outline"
                onClick={() => handleCopy(shortUrl)}
                className="shrink-0"
              >
                {copied ? (
                  <CheckCircle2 className="size-4 text-green-500" />
                ) : (
                  <Copy className="size-4" />
                )}
              </Button>
            </div>

            <div className="bg-muted/50 rounded-lg p-4 border border-border">
              <div className="flex items-center gap-2 mb-3">
                <QrCode className="size-4 text-muted-foreground" />
                <div className="text-xs font-medium text-muted-foreground">
                  QR Code
                </div>
              </div>
              <div className="flex justify-center bg-background rounded-lg p-4">
                <Image
                  src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(shortUrl)}`}
                  alt="QR Code"
                  className="size-[200px]"
                  width={200}
                  height={200}
                />
              </div>
            </div>

            <div className="bg-muted/50 rounded-lg p-4 border border-border">
              <div className="text-xs text-muted-foreground mb-1">
                Original URL
              </div>
              <div className="text-sm break-all">{url}</div>
            </div>

            <Button
              onClick={() => {
                setDialogOpen(false);
                setUrl("");
                setShortUrl("");
              }}
              className="w-full bg-blue-600 hover:bg-blue-700 text-white"
            >
              Shorten Another URL
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
