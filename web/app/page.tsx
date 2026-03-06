import Link from "next/link";

import { Button } from "@/components/ui/button";
import { DOCS_URL } from "@/lib/config";

export default function Home() {
  return (
    <div className="min-h-screen">
      <main className="mx-auto w-full max-w-6xl px-4 py-12 sm:px-6">
        <section className="mb-8 rounded-2xl border border-border/60 bg-card/70 p-8">
          <h1 className="mb-4 max-w-3xl text-4xl font-semibold tracking-tight sm:text-5xl">
            Authentication made easy.
          </h1>
          {/* TODO don't render these if signed in already, maybe show an account button */}
          <div className="mt-6 flex flex-wrap gap-3">
            <Button asChild>
              <Link className="text-black!" href="/login">Sign in</Link>
            </Button>
            <Button asChild variant="outline">
              <Link href="/register">Create account</Link>
            </Button>
          </div>

          <p className="mt-6 text-sm text-muted-foreground">
            Developer?{" "}
            <Link href={DOCS_URL} className="text-foreground underline underline-offset-4">
              Get started
            </Link>
          </p>
        </section>
      </main>
    </div>
  );
}
