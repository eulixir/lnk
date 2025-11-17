import { CheckCircle2, Copy } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "./ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Input } from "./ui/input";

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
  const frontEndUrl = process.env.NEXT_PUBLIC_FRONT_URL;
  const fullUrl = `${frontEndUrl}/${shortUrl}`;

  const handleCopy = async () => {
    if (!fullUrl) return;
    try {
      await navigator.clipboard.writeText(fullUrl);
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
              value={fullUrl}
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
