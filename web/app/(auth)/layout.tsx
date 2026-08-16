import { SiteHeader } from "@/components/layout/site-header";

export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <>
      <SiteHeader variant="minimal" />
      {children}
    </>
  );
}
