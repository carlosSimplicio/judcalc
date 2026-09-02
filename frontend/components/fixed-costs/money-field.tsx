import type { CostFieldDefinition } from "./fields";
import { normalizeMoneyInput } from "@/lib/fixed-costs/money";

export type MoneyFieldProps = {
  field: CostFieldDefinition;
  value: string;
  disabled: boolean;
  onChange: (value: string) => void;
};

export function MoneyField({
  field,
  value,
  disabled,
  onChange,
}: MoneyFieldProps) {
  const id = `fixed-cost-${field.key.replaceAll("_", "-")}`;
  const helpId = field.help ? `${id}-help` : undefined;

  return (
    <div className={`cost-field${field.wide ? " cost-field-wide" : ""}`}>
      <label htmlFor={id}>{field.label}</label>
      <div className="money-input">
        <span aria-hidden="true">R$</span>
        <input
          id={id}
          name={field.key}
          type="text"
          inputMode="decimal"
          value={value}
          disabled={disabled}
          aria-describedby={helpId}
          onChange={(event) => onChange(normalizeMoneyInput(event.target.value))}
        />
      </div>
      {field.help ? <small id={helpId}>{field.help}</small> : null}
    </div>
  );
}
