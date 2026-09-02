"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast-provider";
import { ApiError } from "@/lib/api/client";
import { apiErrorToast } from "@/lib/auth/error-toast";
import {
  type FixedCostKey,
  type FixedCosts,
  type FixedCostsValues,
  saveFixedCosts,
} from "@/lib/fixed-costs/client";
import {
  formatCentsInput,
  formatCurrency,
  monthlyTotal,
  parseMoneyInput,
} from "@/lib/fixed-costs/money";
import { COST_SECTIONS } from "./fields";
import { MoneyField } from "./money-field";

type FixedCostsFormProps = {
  token: string;
  initialCosts: FixedCosts | null;
  loading?: boolean;
  onUnauthorized: () => void;
};

const FIXED_COST_KEYS = COST_SECTIONS.flatMap((section) =>
  section.fields.map((field) => field.key),
);

function emptyInputs(): Record<FixedCostKey, string> {
  return Object.fromEntries(FIXED_COST_KEYS.map((key) => [key, ""])) as Record<
    FixedCostKey,
    string
  >;
}

function valuesToInputs(
  values: FixedCostsValues,
): Record<FixedCostKey, string> {
  return Object.fromEntries(
    FIXED_COST_KEYS.map((key) => [key, formatCentsInput(values[key])]),
  ) as Record<FixedCostKey, string>;
}

function inputsToValues(inputs: Record<FixedCostKey, string>): FixedCostsValues {
  return Object.fromEntries(
    FIXED_COST_KEYS.map((key) => [key, parseMoneyInput(inputs[key])]),
  ) as FixedCostsValues;
}

export function FixedCostsForm({
  token,
  initialCosts,
  loading = false,
  onUnauthorized,
}: FixedCostsFormProps) {
  const { showToast } = useToast();
  const [savedValues, setSavedValues] = useState<FixedCostsValues | null>(
    initialCosts?.values ?? null,
  );
  const [inputs, setInputs] = useState<Record<FixedCostKey, string>>(() =>
    initialCosts ? valuesToInputs(initialCosts.values) : emptyInputs(),
  );
  const [saving, setSaving] = useState(false);
  const isLoading = loading || initialCosts === null;

  useEffect(() => {
    if (!initialCosts) {
      setSavedValues(null);
      setInputs(emptyInputs());
      return;
    }

    setSavedValues(initialCosts.values);
    setInputs(valuesToInputs(initialCosts.values));
  }, [initialCosts]);

  const displayedInputs = initialCosts ? inputs : emptyInputs();
  const currentValues = useMemo(
    () => inputsToValues(displayedInputs),
    [displayedInputs],
  );
  const formDisabled = isLoading || saving;
  const total = initialCosts ? formatCurrency(monthlyTotal(currentValues)) : "—";

  function handleInputChange(key: FixedCostKey, value: string) {
    setInputs((currentInputs) => ({ ...currentInputs, [key]: value }));
  }

  function handleRestore() {
    if (savedValues) {
      setInputs(valuesToInputs(savedValues));
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (formDisabled || !savedValues) {
      return;
    }

    setSaving(true);
    try {
      const result = await saveFixedCosts(token, currentValues);
      setSavedValues(result.values);
      setInputs(valuesToInputs(result.values));
      showToast({
        tone: "success",
        title: "Custos salvos",
        message: "Seus custos fixos foram atualizados.",
      });
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        onUnauthorized();
        return;
      }
      showToast(apiErrorToast(error));
    } finally {
      setSaving(false);
    }
  }

  return (
    <>
      <section className="cost-summary" aria-labelledby="summary-title">
        <div className="summary-icon">
          <svg
            width="30"
            height="30"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <path
              d="M6 2v3M18 2v3M3 9h18M5 4h14a2 2 0 0 1 2 2v14H3V6a2 2 0 0 1 2-2Z"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path d="M8 14h8M8 17h5" strokeLinecap="round" />
          </svg>
        </div>
        <div>
          <p id="summary-title">Total mensal estimado</p>
          <strong>{total}</strong>
          <small>Inclui a média mensal da anuidade da OAB</small>
        </div>
        <span className="summary-badge">Resumo mensal</span>
      </section>

      <form
        className="costs-form"
        aria-busy={isLoading || saving}
        onSubmit={handleSubmit}
      >
        {COST_SECTIONS.map((section) => (
          <section
            key={section.id}
            className="form-section"
            aria-labelledby={section.id}
          >
            <div className="section-heading">
              <span className={`section-icon section-icon-${section.tone}`}>
                {section.icon}
              </span>
              <div>
                <h2 id={section.id}>{section.title}</h2>
                <p>{section.description}</p>
              </div>
            </div>
            <div className="cost-fields-grid">
              {section.fields.map((field) => (
                <MoneyField
                  key={field.key}
                  field={field}
                  value={displayedInputs[field.key]}
                  disabled={formDisabled}
                  onChange={(value) => handleInputChange(field.key, value)}
                />
              ))}
            </div>
          </section>
        ))}

        <div className="form-actions">
          <Button
            className="button-secondary"
            type="button"
            disabled={formDisabled}
            onClick={handleRestore}
          >
            Restaurar
          </Button>
          <Button
            className="button-save"
            type="submit"
            disabled={isLoading}
            loading={saving}
            loadingLabel="Salvando..."
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              aria-hidden="true"
            >
              <path
                d="M5 3h12l2 2v16H5zM8 3v6h8V3M8 21v-7h8v7"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
            Salvar custos
          </Button>
        </div>
      </form>
    </>
  );
}
