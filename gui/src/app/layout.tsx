import "./globals.css";

export const metadata = {
  title: "ntkpr",
  description: "CLI-managed notes database",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
