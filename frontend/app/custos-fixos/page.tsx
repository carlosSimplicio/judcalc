import type { Metadata } from "next";
import { FixedCostsScreen } from "@/components/fixed-costs/fixed-costs-screen";

export const metadata: Metadata = {
  title: "Custos fixos | JudCalc",
  description: "Cadastre suas despesas recorrentes no JudCalc.",
};

export default function FixedCostsPage() {
  return <FixedCostsScreen />;
}
