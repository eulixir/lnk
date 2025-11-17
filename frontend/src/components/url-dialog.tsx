import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@radix-ui/react-dialog";
import { CheckCircle2, Copy, QrCode } from "lucide-react";
import Image from "next/image";
import { useState } from "react";
import { Button } from "./ui/button";
import { DialogHeader } from "./ui/dialog";
import { Input } from "./ui/input";

export function UrlDialog() {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [shortUrl, setShortUrl] = useState("");
  const [url, setUrl] = useState("");
  const [copied, setCopied] = useState(false);

  return (
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
              onClick={() => setCopied(true)}
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
  );
}
