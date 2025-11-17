import { LinkIcon } from "lucide-react";
import { useForm } from "react-hook-form";
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
    formState: { isValid },
  } = useForm<FormData>({
    mode: "onChange",
    defaultValues: {
      url,
    },
  });

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

  const { onChange, ...registerProps } = register("url", {
    required: "URL is required",
    validate: validateUrl,
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="w-full px-4 sm:px-6">
      <div className=" bg-gray-800/50 flex flex-col sm:flex-row items-stretch sm:items-center w-full max-w-2xl lg:max-w-3xl mx-auto min-h-[60px] sm:h-[76px]  rounded-2xl sm:rounded-4xl gap-2 sm:gap-0 overflow-hidden">
        <div className="flex items-center flex-1 h-full min-h-[56px] sm:min-h-0 bg-transparent">
          <LinkIcon className="size-4 text-foreground shrink-0 ml-4 sm:ml-6" />
          <Input
            {...registerProps}
            onChange={(e) => {
              onChange(e);
              setUrl(e.target.value);
            }}
            className="bg-transparent dark:bg-transparent h-full text-foreground ml-3 sm:ml-4 outline-none border-none focus-visible:outline-none focus-visible:border-none focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:bg-transparent selection:bg-[#30b6db]/30 selection:text-foreground text-sm sm:text-base py-3 sm:py-0"
            type="url"
            placeholder="Enter the link here"
            autoComplete="off"
          />
        </div>
        <Button
          type="submit"
          disabled={!isValid}
          className="bg-[#30b6db] text-primary-foreground sm:rounded-4xl w-full sm:w-[178px] h-full sm:h-[60px] cursor-pointer hover:shadow-lg hover:bg-[#30b6db]/90 transition-all duration-300 sm:mr-0.5 disabled:opacity-50 disabled:cursor-not-allowed shrink-0"
        >
          Shorten
        </Button>
      </div>
    </form>
  );
}
