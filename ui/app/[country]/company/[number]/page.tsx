import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { Suspense } from "react";
import { createClient, createConfig } from "@/client/client";
import { getCompanyGraph } from "@/client/sdk.gen";
import { Home } from "@/components/Home";

const VALID_COUNTRIES = new Set(["uk"]);

type Props = { params: Promise<{ country: string; number: string }> };

function serverClient() {
  return createClient(
    createConfig({
      baseUrl: process.env.API_URL ?? "http://localhost:8080",
      next: { revalidate: 3600 },
    }),
  );
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { country, number } = await params;
  if (!VALID_COUNTRIES.has(country)) return {};

  const { data } = await getCompanyGraph({
    client: serverClient(),
    path: { company_number: number },
    query: { country },
  });

  const companyNode = data?.nodes?.find((n) => n.type === "company");
  const name = companyNode?.label ?? number;

  return {
    title: `${name} — Diktyon`,
    description: `Officers, PSCs and corporate network for ${name} (${number}).`,
  };
}

export default async function CompanyPage({ params }: Props) {
  const { country, number } = await params;
  if (!VALID_COUNTRIES.has(country)) notFound();

  return (
    <Suspense>
      <Home rootCompany={number} />
    </Suspense>
  );
}
