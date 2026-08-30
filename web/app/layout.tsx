import type { Metadata } from "next";
import { ThemeToggle } from "@/components/theme-toggle";
import "./globals.css";

export const metadata: Metadata = {
  title: "KeepItSecret",
  description: "Share text once, then it disappears.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <div className="site-shell">
          <header className="site-header">
            <a className="wordmark" href="/" aria-label="KeepItSecret home">
              <span className="wordmark-mark" aria-hidden="true">K</span>
              <span>KeepItSecret</span>
            </a>
            <nav className="site-nav" aria-label="Primary navigation">
              <a href="/">Create</a>
              <ThemeToggle />
            </nav>
          </header>
          <main>{children}</main>
          <footer className="site-footer">Private by design. Nothing is kept after it is opened.</footer>
        </div>
      </body>
    </html>
  );
}
