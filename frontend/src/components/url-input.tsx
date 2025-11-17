import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { LinkIcon } from "lucide-react";
import { Button } from "./ui/button";
import { Input } from "./ui/input";

interface UrlInputProps {
  url: string;
  setUrl: (url: string) => void;
  requestShorten: () => void;
}

interface FormData {
  url: string;
}

export function UrlInput({ url, setUrl, requestShorten }: UrlInputProps) {
  const {
    register,
    handleSubmit,
    watch,
    formState: { isValid },
  } = useForm<FormData>({
    mode: "onChange",
    defaultValues: {
      url,
    },
    values: {
      url,
    },
  });

  const watchedUrl = watch("url");

  useEffect(() => {
    setUrl(watchedUrl);
  }, [watchedUrl, setUrl]);

  const onSubmit = (data: FormData) => {
    if (isValid) {
      setUrl(data.url);
      requestShorten();
    }
  };

  const validateUrl = (value: string) => {
    if (!value) {
      return "URL is required";
    }
    try {
      new URL(value);
      return true;
    } catch {
      return "Please enter a valid URL starting with http:// or https://";
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div className="flex items-center w-[659px] mx-auto h-[76px] border-border rounded-4xl border-4">
        <LinkIcon className="size-4 text-muted-foreground shrink-0 ml-6" />
        <Input
          {...register("url", {
            required: "URL is required",
            validate: validateUrl,
          })}
          className="bg-transparent h-full dark:bg-transparent ml-4 outline-none border-none focus-visible:outline-none focus-visible:border-none focus-visible:ring-0 focus-visible:ring-offset-0"
          type="url"
          placeholder="Enter the link here"
        />
        <Button
          type="submit"
          disabled={!isValid}
          className="bg-[#30b6db] text-primary-foreground rounded-4xl w-[178px] h-[60px] cursor-pointer hover:scale-105 hover:shadow-lg hover:bg-[#30b6db]/90 transition-all duration-300 mr-0.5 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
        >
          Shorten
        </Button>
      </div>
    </form>
  );
}
