import { LinkIcon } from "lucide-react";
import { Input } from "./ui/input";
import { Button } from "./ui/button";

export function UrlInput() {
  return (
    <div className="flex items-center w-[659px] mx-auto h-[76px] border-border rounded-4xl border-4">
      <LinkIcon className="size-4 text-muted-foreground shrink-0 ml-6" />
      <Input
        className="bg-transparent h-full dark:bg-transparent ml-4 outline-none border-none focus-visible:outline-none focus-visible:border-none focus-visible:ring-0 focus-visible:ring-offset-0"
        type="url"
        placeholder="Enter the link here"
      />
      <Button className="bg-[#30b6db] text-primary-foreground rounded-4xl w-[178px] h-[60px] cursor-pointer hover:scale-105 hover:shadow-lg hover:bg-[#30b6db]/90 transition-all duration-300 mr-0.5">
        Shorten
      </Button>
    </div>
  );
}
