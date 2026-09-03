import type { Metadata } from "next";
import { HomeScreen } from "@/components/home/home-screen";

export const metadata: Metadata = {
  title: "JudCalc | Honorários mais seguros",
  description:
    "Organize seus custos e calcule honorários jurídicos com mais segurança.",
};

export default function HomePage() {
  return <HomeScreen />;
}
