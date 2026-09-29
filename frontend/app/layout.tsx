import type { Metadata } from "next";
import { Bricolage_Grotesque, Figtree, Noto_Sans_Thai } from "next/font/google";
import { Suspense } from "react";

import { AppShell } from "@/components/layout/app-shell";
import { Toaster } from "@/components/ui/toaster";

import "./globals.css";

const bricolage = Bricolage_Grotesque({ variable: "--font-bricolage", subsets: ["latin"], axes: ["opsz"] });
const figtree = Figtree({ variable: "--font-figtree", subsets: ["latin"] });
const notoThai = Noto_Sans_Thai({ variable: "--font-noto-thai", subsets: ["thai"] });

export const metadata: Metadata = {
  title: { default: "TableTalk", template: "%s — TableTalk" },
  description: "Find a table you'll love: browse restaurants, read reviews and reserve a table.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en" className={`${bricolage.variable} ${figtree.variable} ${notoThai.variable} h-full`}>
      <body className="flex min-h-full flex-col">
        <AppShell>{children}</AppShell>
        <Suspense>
          <Toaster />
        </Suspense>
      </body>
    </html>
  );
}
