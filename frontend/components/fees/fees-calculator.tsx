"use client";

import Link from "next/link";
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type FormEvent,
  type ReactNode,
  type RefObject,
} from "react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast-provider";
import { ApiError } from "@/lib/api/client";
import { apiErrorToast } from "@/lib/auth/error-toast";
import {
  calculateFee,
  listAllAreas,
  listServices,
  type FeeCalculationResult,
  type FeeLevel,
  type PageMetadata,
  type Service,
} from "@/lib/fees/client";
import { formatCurrency } from "@/lib/fixed-costs/money";

type FeesCalculatorProps = {
  token: string;
  onUnauthorized: () => void;
};

const LEVELS: Array<{
  value: FeeLevel;
  riskLabel: string;
  riskHelp: string;
  complexityLabel: string;
  complexityHelp: string;
}> = [
  { value: "low", riskLabel: "Baixo", riskHelp: "Previsível", complexityLabel: "Baixa", complexityHelp: "Rotineira" },
  { value: "medium", riskLabel: "Médio", riskHelp: "Moderado", complexityLabel: "Média", complexityHelp: "Intermediária" },
  { value: "high", riskLabel: "Alto", riskHelp: "Elevado", complexityLabel: "Alta", complexityHelp: "Especializada" },
];

export function FeesCalculator({ token, onUnauthorized }: FeesCalculatorProps) {
  const { showToast } = useToast();
  const dropdownRef = useRef<HTMLDivElement>(null);
  const optionsRef = useRef<HTMLDivElement>(null);
  const loadMoreRef = useRef<HTMLDivElement>(null);
  const resultRef = useRef<HTMLElement>(null);
  const requestVersion = useRef(0);
  const loadingMoreRequest = useRef(false);

  const [areas, setAreas] = useState<Map<number, string>>(new Map());
  const [services, setServices] = useState<Service[]>([]);
  const [meta, setMeta] = useState<PageMetadata | null>(null);
  const [search, setSearch] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [queryRevision, setQueryRevision] = useState(0);
  const [selectedService, setSelectedService] = useState<Service | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const [loadingServices, setLoadingServices] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [loadMoreFailed, setLoadMoreFailed] = useState(false);
  const [estimatedHours, setEstimatedHours] = useState("");
  const [billableHours, setBillableHours] = useState("");
  const [risk, setRisk] = useState<FeeLevel>("medium");
  const [complexity, setComplexity] = useState<FeeLevel>("medium");
  const [calculating, setCalculating] = useState(false);
  const [result, setResult] = useState<FeeCalculationResult | null>(null);

  const areaName = useCallback(
    (areaId: number) => areas.get(areaId) ?? `Área ${areaId}`,
    [areas],
  );

  useEffect(() => {
    listAllAreas(token)
      .then((items) => setAreas(new Map(items.map((area) => [area.id, area.name]))))
      .catch((error: unknown) => {
        if (error instanceof ApiError && error.status === 401) {
          onUnauthorized();
          return;
        }
        showToast(apiErrorToast(error));
      });
  }, [onUnauthorized, showToast, token]);

  useEffect(() => {
    if (selectedService && search === selectedService.name) return;
    const timeout = window.setTimeout(() => {
      setDebouncedQuery(search.trim());
      setQueryRevision((revision) => revision + 1);
    }, 300);
    return () => window.clearTimeout(timeout);
  }, [search, selectedService]);

  useEffect(() => {
    const version = ++requestVersion.current;
    setLoadingServices(true);
    setLoadingMore(false);
    loadingMoreRequest.current = false;
    setLoadMoreFailed(false);
    listServices(token, 1, debouncedQuery)
      .then((page) => {
        if (version !== requestVersion.current) return;
        setServices(page.items);
        setMeta(page.meta);
      })
      .catch((error: unknown) => {
        if (version !== requestVersion.current) return;
        setServices([]);
        setMeta(null);
        if (error instanceof ApiError && error.status === 401) {
          onUnauthorized();
          return;
        }
        showToast(apiErrorToast(error));
      })
      .finally(() => {
        if (version === requestVersion.current) setLoadingServices(false);
      });
  }, [debouncedQuery, onUnauthorized, queryRevision, showToast, token]);

  const loadNextPage = useCallback(async () => {
    if (!meta || loadingServices || loadingMoreRequest.current || meta.page >= meta.totalPages) return;
    const version = requestVersion.current;
    loadingMoreRequest.current = true;
    setLoadingMore(true);
    setLoadMoreFailed(false);
    try {
      const page = await listServices(token, meta.page + 1, debouncedQuery);
      if (version !== requestVersion.current) return;
      setServices((current) => {
        const knownIds = new Set(current.map((service) => service.id));
        return [...current, ...page.items.filter((service) => !knownIds.has(service.id))];
      });
      setMeta(page.meta);
    } catch (error) {
      if (version !== requestVersion.current) return;
      if (error instanceof ApiError && error.status === 401) {
        onUnauthorized();
        return;
      }
      setLoadMoreFailed(true);
    } finally {
      loadingMoreRequest.current = false;
      if (version === requestVersion.current) setLoadingMore(false);
    }
  }, [debouncedQuery, loadingMore, loadingServices, meta, onUnauthorized, token]);

  useEffect(() => {
    if (!menuOpen || !loadMoreRef.current || !optionsRef.current) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) void loadNextPage();
      },
      { root: optionsRef.current, rootMargin: "80px" },
    );
    observer.observe(loadMoreRef.current);
    return () => observer.disconnect();
  }, [loadNextPage, menuOpen, services.length]);

  useEffect(() => {
    function handleDocumentClick(event: MouseEvent) {
      if (!dropdownRef.current?.contains(event.target as Node)) setMenuOpen(false);
    }
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") setMenuOpen(false);
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        setMenuOpen(true);
        dropdownRef.current?.querySelector("input")?.focus();
      }
    }
    document.addEventListener("click", handleDocumentClick);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("click", handleDocumentClick);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const hasMore = Boolean(meta && meta.page < meta.totalPages);

  function handleSearch(value: string) {
    requestVersion.current += 1;
    setSearch(value);
    setMenuOpen(true);
    if (selectedService && value !== selectedService.name) setSelectedService(null);
  }

  function handleSelect(service: Service) {
    setSelectedService(service);
    setSearch(service.name);
    setMenuOpen(false);
  }

  function handleReset() {
    setSelectedService(null);
    setSearch("");
    setDebouncedQuery("");
    setEstimatedHours("");
    setBillableHours("");
    setRisk("medium");
    setComplexity("medium");
    setResult(null);
    setMenuOpen(false);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const parsedEstimatedHours = Number(estimatedHours);
    const parsedBillableHours = Number(billableHours);
    if (!selectedService || parsedEstimatedHours <= 0 || parsedBillableHours <= 0) {
      showToast({
        tone: "error",
        title: "Dados incompletos",
        message: "Selecione um serviço e informe horas maiores que zero.",
      });
      return;
    }

    setCalculating(true);
    try {
      const calculation = await calculateFee(token, {
        serviceId: selectedService.id,
        estimatedHours: parsedEstimatedHours,
        billableHoursPerMonth: parsedBillableHours,
        complexity,
        risk,
      });
      setResult(calculation);
      if (window.matchMedia("(max-width: 68rem)").matches) {
        window.setTimeout(() => resultRef.current?.scrollIntoView({ behavior: "smooth", block: "start" }));
      }
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        onUnauthorized();
        return;
      }
      showToast(apiErrorToast(error));
    } finally {
      setCalculating(false);
    }
  }

  return (
    <div className="fees-layout">
      <form className="costs-form fees-form" onSubmit={handleSubmit} onReset={handleReset}>
        <section className="form-section" aria-labelledby="service-title">
          <SectionHeading id="service-title" tone="blue" icon={<SearchIcon />} title="Serviço jurídico" description="Digite para buscar ou selecione um serviço da lista." />
          <div className="service-picker">
            <label htmlFor="service-search">Qual serviço será prestado?</label>
            <div className={`service-dropdown${menuOpen ? " is-open" : ""}`} ref={dropdownRef}>
              <div className="service-combobox">
                <SearchIcon className="service-search-icon" />
                <input
                  id="service-search"
                  type="search"
                  role="combobox"
                  aria-expanded={menuOpen}
                  aria-controls="service-options"
                  aria-autocomplete="list"
                  autoComplete="off"
                  placeholder="Selecione ou busque um serviço..."
                  value={search}
                  onChange={(event) => handleSearch(event.target.value)}
                  onFocus={() => setMenuOpen(true)}
                />
                <button className="service-toggle" type="button" aria-label={menuOpen ? "Fechar lista de serviços" : "Abrir lista de serviços"} onClick={() => setMenuOpen((open) => !open)}>
                  <ChevronIcon />
                </button>
              </div>
              {menuOpen ? (
                <div className="service-menu">
                  <p className="service-count">
                    {loadingServices ? "Buscando serviços..." : `${meta?.total ?? 0} ${(meta?.total ?? 0) === 1 ? "serviço encontrado" : "serviços encontrados"}`}
                  </p>
                  <div className="service-options" id="service-options" role="listbox" aria-label="Serviços disponíveis" ref={optionsRef}>
                    {services.map((service) => {
                      const selected = selectedService?.id === service.id;
                      return (
                        <button key={service.id} className={`service-option${selected ? " is-selected" : ""}`} type="button" role="option" aria-selected={selected} onClick={() => handleSelect(service)}>
                          <span className="option-check">✓</span>
                          <span className="option-copy"><strong>{service.name}</strong><small>{areaName(service.areaId)}</small></span>
                        </button>
                      );
                    })}
                    {!loadingServices && services.length === 0 ? <p className="service-empty">Nenhum serviço encontrado. Tente outro termo.</p> : null}
                    {loadingMore ? <div className="service-page-loading" role="status"><span className="small-spinner" aria-hidden="true" /><span>Carregando mais serviços...</span></div> : null}
                    {loadMoreFailed ? <button className="service-retry" type="button" onClick={() => void loadNextPage()}>Tentar novamente</button> : null}
                    {hasMore && !loadingMore && !loadMoreFailed ? <div className="service-load-marker" ref={loadMoreRef} aria-hidden="true" /> : null}
                  </div>
                </div>
              ) : null}
            </div>
          </div>
        </section>

        <section className="form-section" aria-labelledby="work-title">
          <SectionHeading id="work-title" tone="teal" icon={<ClockIcon />} title="Tempo de trabalho" description="Informe a dedicação prevista e sua capacidade mensal faturável." />
          <div className="cost-fields-grid">
            <HoursField id="estimated-hours" label="Horas estimadas" help="Tempo previsto para concluir este serviço." value={estimatedHours} onChange={setEstimatedHours} />
            <HoursField id="billable-hours" label="Horas faturáveis por mês" help="Capacidade mensal usada para calcular o custo operacional da hora." value={billableHours} onChange={setBillableHours} />
          </div>
        </section>

        <section className="form-section" aria-labelledby="case-title">
          <SectionHeading id="case-title" tone="orange" icon={<RiskIcon />} title="Características do caso" description="Avalie os fatores que podem aumentar o esforço ou a responsabilidade." />
          <div className="factor-fields">
            <LevelField name="risk" title="Risco" description="Probabilidade de imprevistos ou resultado desfavorável." value={risk} onChange={setRisk} labelKey="riskLabel" helpKey="riskHelp" />
            <LevelField name="complexity" title="Complexidade" description="Nível de especialização e volume de trabalho exigidos." value={complexity} onChange={setComplexity} labelKey="complexityLabel" helpKey="complexityHelp" />
          </div>
        </section>

        <div className="form-actions fees-actions">
          <Button className="button-secondary" type="reset" disabled={calculating}>Limpar</Button>
          <Button className="button-save" type="submit" loading={calculating} loadingLabel="Calculando...">
            <CalculatorIcon /> Calcular honorários
          </Button>
        </div>
      </form>

      <EstimateCard cardRef={resultRef} result={result} selectedService={selectedService} areaName={areaName} />
    </div>
  );
}

type SectionHeadingProps = { id: string; tone: string; icon: ReactNode; title: string; description: string };
function SectionHeading({ id, tone, icon, title, description }: SectionHeadingProps) {
  return <div className="section-heading"><span className={`section-icon section-icon-${tone}`}>{icon}</span><div><h2 id={id}>{title}</h2><p>{description}</p></div></div>;
}

function HoursField({ id, label, help, value, onChange }: { id: string; label: string; help: string; value: string; onChange: (value: string) => void }) {
  return <div className="cost-field"><label htmlFor={id}>{label}</label><div className="unit-input"><input id={id} type="number" min="0.5" step="0.5" required placeholder="0" value={value} onChange={(event) => onChange(event.target.value)} /><span>horas</span></div><small>{help}</small></div>;
}

type LevelFieldProps = {
  name: string;
  title: string;
  description: string;
  value: FeeLevel;
  onChange: (value: FeeLevel) => void;
  labelKey: "riskLabel" | "complexityLabel";
  helpKey: "riskHelp" | "complexityHelp";
};
function LevelField({ name, title, description, value, onChange, labelKey, helpKey }: LevelFieldProps) {
  return <fieldset className="factor-field"><legend>{title}</legend><p>{description}</p><div className="choice-group">{LEVELS.map((level) => <label key={level.value}><input type="radio" name={name} value={level.value} checked={value === level.value} onChange={() => onChange(level.value)} /><span><strong>{level[labelKey]}</strong><small>{level[helpKey]}</small></span></label>)}</div></fieldset>;
}

type EstimateCardProps = {
  cardRef: RefObject<HTMLElement | null>;
  result: FeeCalculationResult | null;
  selectedService: Service | null;
  areaName: (areaId: number) => string;
};
function EstimateCard({ cardRef, result, selectedService, areaName }: EstimateCardProps) {
  const warning = result?.warnings.find((item) => item.code === "fixed_costs_unavailable");
  return (
    <aside className="estimate-card" aria-labelledby="estimate-title" aria-live="polite" ref={cardRef}>
      <div className="estimate-heading"><span className="estimate-icon"><CalculatorIcon /></span><div><p>{result ? "Cálculo concluído" : "Sua estimativa"}</p><h2 id="estimate-title">{result ? "Resultado dos honorários" : "Resumo do cálculo"}</h2></div></div>
      {!result ? <div><div className="estimate-empty"><span><DocumentIcon /></span><strong>Pronto para calcular</strong><p>Preencha os dados ao lado para obter uma sugestão de honorários.</p></div><div className="selected-service"><small>Serviço selecionado</small><strong>{selectedService?.name ?? "Nenhum serviço selecionado"}</strong><span>{selectedService ? areaName(selectedService.areaId) : "Escolha um serviço para ver os detalhes."}</span></div><p className="estimate-note">ⓘ A estimativa considera seus custos, o tempo informado e as características do caso.</p></div> : (
        <div className="estimate-result">
          {result.calculation.technicalEstimateCents !== null ? <div className="result-value"><small>Estimativa técnica</small><strong>{formatCurrency(result.calculation.technicalEstimateCents)}</strong><span>Valor total calculado</span></div> : <div className="estimate-warning"><strong>Estimativa indisponível</strong><p>{warning?.message ?? "Cadastre custos fixos para concluir o cálculo."}</p><Link href="/custos-fixos">Cadastrar custos fixos</Link></div>}
          <div className="result-range"><span>Referência da tabela OAB</span><strong>{formatOabReference(result)}</strong></div>
          <dl className="result-breakdown">
            <ResultRow label="Custo operacional por hora" value={formatNullableCurrency(result.calculation.operationalHourCostCents)} />
            <ResultRow label="Custo mínimo sustentável" value={formatNullableCurrency(result.calculation.minimumSustainableCostCents)} />
            <ResultRow label="Custos fixos mensais" value={formatCurrency(result.calculation.monthlyFixedCostsCents)} />
            <ResultRow label="Ajuste de risco" value={formatAdjustment(result.inputs.riskFactor)} />
            <ResultRow label="Ajuste de complexidade" value={formatAdjustment(result.inputs.complexityFactor)} />
          </dl>
          <div className="selected-service"><small>Serviço calculado</small><strong>{result.service.name}</strong><span>{areaName(result.service.areaId)}</span></div>
          <p className="estimate-note">ⓘ A referência da OAB é apresentada separadamente da estimativa técnica.</p>
        </div>
      )}
    </aside>
  );
}

function ResultRow({ label, value }: { label: string; value: string }) {
  return <div><dt>{label}</dt><dd>{value}</dd></div>;
}

function formatNullableCurrency(value: number | null): string {
  return value === null ? "Indisponível" : formatCurrency(value);
}

function formatOabReference(result: FeeCalculationResult): string {
  if (result.oabReference.amountCents !== null) return formatCurrency(result.oabReference.amountCents);
  if (result.oabReference.percentageMin !== null && result.oabReference.percentageMax !== null) {
    return `${formatNumber(result.oabReference.percentageMin)}% a ${formatNumber(result.oabReference.percentageMax)}%`;
  }
  return "Não informada";
}

function formatAdjustment(factor: number): string {
  const percentage = (factor - 1) * 100;
  return `${percentage > 0 ? "+" : ""}${formatNumber(percentage)}%`;
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 2 }).format(value);
}

function SearchIcon({ className }: { className?: string }) { return <svg className={className} width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><circle cx="11" cy="11" r="7"/><path d="m20 20-4-4"/></svg>; }
function ChevronIcon() { return <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><path d="m7 10 5 5 5-5" strokeLinecap="round" strokeLinejoin="round"/></svg>; }
function ClockIcon() { return <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></svg>; }
function RiskIcon() { return <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><path d="M12 3 3.5 19h17L12 3Z"/><path d="M12 9v4M12 16h.01"/></svg>; }
function CalculatorIcon() { return <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true"><path d="M4 3h16v18H4zM8 7h8M8 12h2M14 12h2M8 16h2M14 16h2"/></svg>; }
function DocumentIcon() { return <svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" aria-hidden="true"><path d="M7 3h10l3 3v15H4V3h3Z"/><path d="M8 10h8M8 14h8M8 18h5"/></svg>; }
