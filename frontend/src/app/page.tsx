import { UrlShortener } from "@/components/url-shortener";

export default function Home() {
  return (
    <div className="min-h-screen">
      <header className="border-b border-border">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-xl font-bold text-[#30b6db]">lnk</span>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-12 md:py-20 mt-32">
        <div className="text-center mb-12 max-w-4xl mx-auto">
          <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold mb-6 ">
            <span className="text-[#30b6db]">
              Shorten Your Loooong Links :)
            </span>
          </h1>
          <p className="text-lg text-muted-foreground text-pretty max-w-2xl mx-auto">
            lnk is an efficient and easy-to-use URL shortening service that
            streamlines your online experience.
          </p>
        </div>

        <UrlShortener />
      </main>
    </div>
  );
}
