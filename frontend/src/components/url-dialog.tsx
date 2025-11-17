import { CheckCircle2, Copy, QrCode } from "lucide-react";
import Image from "next/image";
import { useState } from "react";
import { Button } from "./ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Input } from "./ui/input";
import { toast } from "sonner";

interface UrlDialogProps {
  shortUrl: string;
  originalUrl: string;
  open: boolean;
  setUrl: (url: string) => void;
  onOpenChange: (open: boolean) => void;
}

export function UrlDialog({
  shortUrl,
  originalUrl,
  open,
  setUrl,
  onOpenChange,
}: UrlDialogProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    if (!shortUrl) return;
    try {
      await navigator.clipboard.writeText(shortUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
      toast.success("Copied to clipboard");
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md bg-gray-800 border-0">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-[#30b6db]">
            <CheckCircle2 className="size-6 text-[#30b6db]" />
            Your Short URL is Ready!
          </DialogTitle>
          <DialogDescription className="text-muted-foreground">
            Share this shortened link anywhere
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4 mt-4">
          <div className="flex items-center gap-2">
            <Input
              readOnly
              value={shortUrl}
              className="font-mono bg-gray-800 border-[#30b6db]/30 text-foreground"
            />
            <Button
              size="icon"
              variant="outline"
              onClick={handleCopy}
              className="shrink-0 border-[#30b6db]/30 hover:bg-[#30b6db]/10 hover:border-[#30b6db]"
            >
              {copied ? (
                <CheckCircle2 className="size-4 text-[#30b6db]" />
              ) : (
                <Copy className="size-4 text-[#30b6db]" />
              )}
            </Button>
          </div>

          <div className="bg-gray-800 rounded-lg p-4 border border-[#30b6db]/20">
            <div className="flex items-center gap-2 mb-3">
              <QrCode className="size-4 text-[#30b6db]" />
              <div className="text-xs font-medium text-[#30b6db]">QR Code</div>
            </div>
            <div className="flex justify-center bg-gray-900 rounded-lg p-4">
              <Image
                src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(shortUrl)}`}
                alt="QR Code"
                className="size-[200px]"
                width={200}
                height={200}
              />
            </div>
          </div>

          <div className="bg-gray-800 rounded-lg p-4 border border-[#30b6db]/20">
            <div className="text-xs text-[#30b6db] mb-1">Original URL</div>
            <div className="text-sm break-all text-foreground">
              {originalUrl}
            </div>
          </div>

          <Button
            onClick={() => {
              setUrl("");
              onOpenChange(false);
            }}
            className="w-full bg-[#30b6db] hover:bg-[#30b6db]/90 text-white"
          >
            Shorten a new URL
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
