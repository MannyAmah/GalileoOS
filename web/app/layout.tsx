import type { ReactNode } from "react";

export const metadata = {
  title: "Galileo OS",
  description: "Galileo OS — the brain of businesses.",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
