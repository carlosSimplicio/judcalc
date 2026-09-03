import type { Metadata } from "next";
import { FeesScreen } from "@/components/fees/fees-screen";

export const metadata: Metadata = {
  title: "Calcular honorários | JudCalc",
  description: "Calcule uma estimativa técnica de honorários jurídicos.",
};

export default function FeesPage() {
  return <FeesScreen />;
}
