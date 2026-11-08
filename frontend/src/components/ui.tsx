import Link from "next/link";
import type { ComponentProps, ReactNode } from "react";

import type { Tone } from "@/lib/labels";

export function cx(...c: (string | false | null | undefined)[]) {
  return c.filter(Boolean).join(" ");
}

// ---------- buttons ----------

type Variant = "primary" | "brand" | "secondary" | "ghost" | "danger";

const variants: Record<Variant, string> = {
  // Amber carries graphite text only (white on amber fails contrast and the brand rule).
  primary: "bg-action text-ink hover:bg-action-hover border border-[#d98d22]/40 shadow-panel",
  brand: "bg-brand text-white hover:bg-brand-strong border border-brand-strong shadow-panel",
  secondary: "bg-surface text-ink hover:bg-sunk border border-line-strong shadow-panel",
  ghost: "text-ink-2 hover:text-ink hover:bg-sunk border border-transparent",
  danger: "bg-surface text-bad hover:bg-bad-soft border border-line-strong",
};

const sizes = {
  sm: "h-8 px-3 text-[13px] gap-1.5",
  md: "h-10 px-4 text-sm gap-2",
};

export function buttonClass(variant: Variant = "secondary", size: keyof typeof sizes = "md", extra = "") {
  return cx(
    "inline-flex shrink-0 items-center justify-center rounded-control font-medium whitespace-nowrap select-none",
    "transition-[background-color,color,box-shadow,transform] duration-150 ease-out active:translate-y-px",
    "disabled:cursor-not-allowed disabled:opacity-55 disabled:active:translate-y-0",
    variants[variant],
    sizes[size],
    extra,
  );
}

export function Button({
  variant = "secondary",
  size = "md",
  className = "",
  ...props
}: ComponentProps<"button"> & { variant?: Variant; size?: keyof typeof sizes }) {
  return <button {...props} className={buttonClass(variant, size, className)} />;
}

export function ButtonLink({
  variant = "secondary",
  size = "md",
  className = "",
  ...props
}: ComponentProps<typeof Link> & { variant?: Variant; size?: keyof typeof sizes }) {
  return <Link {...props} className={buttonClass(variant, size, className)} />;
}

// ---------- layout ----------

export function PageHeader({
  title,
  description,
  actions,
}: {
  title: string;
  description?: ReactNode;
  actions?: ReactNode;
}) {
  return (
    <header className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        <h1 className="text-[26px] leading-tight font-bold text-ink">{title}</h1>
        {description ? <p className="mt-1.5 text-sm text-ink-2">{description}</p> : null}
      </div>
      {actions ? <div className="flex flex-wrap items-center gap-2">{actions}</div> : null}
    </header>
  );
}

export function Panel({
  title,
  description,
  actions,
  children,
  className = "",
  flush = false,
}: {
  title?: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
  flush?: boolean;
}) {
  return (
    <section className={cx("rounded-panel border border-line bg-surface shadow-panel", className)}>
      {title || actions ? (
        <div className="flex flex-wrap items-start justify-between gap-3 border-b border-line px-5 py-4">
          <div className="min-w-0">
            {title ? <h2 className="text-[15px] font-semibold text-ink">{title}</h2> : null}
            {description ? <p className="mt-0.5 text-[13px] text-ink-2">{description}</p> : null}
          </div>
          {actions ? <div className="flex flex-wrap items-center gap-2">{actions}</div> : null}
        </div>
      ) : null}
      <div className={flush ? "" : "p-5"}>{children}</div>
    </section>
  );
}

// ---------- numbers ----------

export type Stat = { label: string; value: string; hint?: ReactNode; emphasis?: boolean };

/** One strip of figures with hairline dividers (not a grid of cards). */
export function StatStrip({ stats }: { stats: Stat[] }) {
  return (
    <dl
      className="grid overflow-hidden rounded-panel border border-line bg-line shadow-panel [grid-template-columns:repeat(auto-fit,minmax(150px,1fr))]"
      style={{ gap: "1px" }}
    >
      {stats.map((s) => (
        <div key={s.label} className="bg-surface px-5 py-4">
          <dt className="text-[13px] text-ink-2">{s.label}</dt>
          <dd
            className={cx(
              "num mt-1 font-display text-[26px] leading-tight font-bold tracking-[-0.02em]",
              s.emphasis ? "text-brand" : "text-ink",
            )}
          >
            {s.value}
          </dd>
          {s.hint ? <dd className="mt-0.5 text-xs text-ink-3">{s.hint}</dd> : null}
        </div>
      ))}
    </dl>
  );
}

// ---------- chips ----------

const tones: Record<Tone, string> = {
  neutral: "bg-sunk text-ink-2 ring-line-strong/70",
  teal: "bg-brand-soft text-brand ring-brand-line",
  amber: "bg-action-soft text-action-ink ring-[#efd2a3]",
  red: "bg-bad-soft text-bad ring-[#f3c5bf]",
  green: "bg-ok-soft text-ok ring-[#bfdfcc]",
};

export function Chip({ tone = "neutral", children }: { tone?: Tone; children: ReactNode }) {
  return (
    <span
      className={cx(
        "inline-flex h-6 items-center gap-1.5 rounded-full px-2.5 text-xs font-medium whitespace-nowrap ring-1 ring-inset",
        tones[tone],
      )}
    >
      <span className="size-1.5 rounded-full bg-current opacity-80" aria-hidden />
      {children}
    </span>
  );
}

// ---------- tables ----------

export function Table({ children, label }: { children: ReactNode; label: string }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[640px] border-collapse text-left text-sm" aria-label={label}>
        {children}
      </table>
    </div>
  );
}

export function Th({ children, className = "" }: { children?: ReactNode; className?: string }) {
  return (
    <th
      scope="col"
      className={cx(
        "border-b border-line bg-[#fbfaf7] px-4 py-2.5 text-xs font-medium whitespace-nowrap text-ink-2 first:pl-5 last:pr-5",
        className,
      )}
    >
      {children}
    </th>
  );
}

export function Td({ children, className = "" }: { children?: ReactNode; className?: string }) {
  return <td className={cx("border-b border-line px-4 py-3 align-middle first:pl-5 last:pr-5", className)}>{children}</td>;
}

export function Tr({ children, className = "" }: { children: ReactNode; className?: string }) {
  return <tr className={cx("transition-colors last:[&>td]:border-b-0 hover:bg-[#fbfaf7]", className)}>{children}</tr>;
}

// ---------- empty & notices ----------

export function Empty({ title, children, action }: { title: string; children?: ReactNode; action?: ReactNode }) {
  return (
    <div className="flex flex-col items-start gap-2 px-5 py-10 sm:items-center sm:text-center">
      <p className="font-display text-[15px] font-semibold text-ink">{title}</p>
      {children ? <div className="max-w-md text-sm text-ink-2">{children}</div> : null}
      {action ? <div className="mt-2">{action}</div> : null}
    </div>
  );
}

export function Notice({ tone = "neutral", children }: { tone?: "neutral" | "amber" | "red" | "green"; children: ReactNode }) {
  const t = {
    neutral: "border-line bg-sunk text-ink-2",
    amber: "border-[#efd2a3] bg-action-soft text-action-ink",
    red: "border-[#f3c5bf] bg-bad-soft text-bad",
    green: "border-[#bfdfcc] bg-ok-soft text-ok",
  }[tone];
  return <div className={cx("rounded-control border px-4 py-3 text-sm", t)}>{children}</div>;
}

// ---------- form fields (server-safe) ----------

export function Field({
  label,
  hint,
  htmlFor,
  children,
  className = "",
}: {
  label: string;
  hint?: ReactNode;
  htmlFor: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <div className={cx("flex flex-col gap-1.5", className)}>
      <label htmlFor={htmlFor} className="text-[13px] font-medium text-ink">
        {label}
      </label>
      {children}
      {hint ? <p className="text-xs text-ink-3">{hint}</p> : null}
    </div>
  );
}

export const inputClass = cx(
  "h-10 w-full rounded-control border border-line-strong bg-surface px-3 text-sm text-ink shadow-[inset_0_1px_0_rgb(31_35_40/0.03)]",
  "transition-[border-color,box-shadow] duration-150 outline-none",
  "hover:border-[#bfb8a8] focus:border-brand focus:ring-3 focus:ring-brand/15",
  "disabled:cursor-not-allowed disabled:bg-sunk disabled:text-ink-3",
  "aria-invalid:border-bad aria-invalid:ring-bad/15",
);

export function Input(props: ComponentProps<"input">) {
  return <input {...props} className={cx(inputClass, props.className)} />;
}

export function Select(props: ComponentProps<"select">) {
  return <select {...props} className={cx(inputClass, "pr-8", props.className)} />;
}

export function Textarea(props: ComponentProps<"textarea">) {
  return <textarea {...props} className={cx(inputClass, "h-auto min-h-24 py-2.5 leading-relaxed", props.className)} />;
}

export function Checkbox({ label, hint, ...props }: ComponentProps<"input"> & { label: string; hint?: string }) {
  return (
    <label className="flex cursor-pointer items-start gap-3 py-1.5">
      <input type="checkbox" {...props} className="mt-0.5 size-4 shrink-0 cursor-pointer accent-brand" />
      <span className="text-sm">
        <span className="text-ink">{label}</span>
        {hint ? <span className="mt-0.5 block text-xs text-ink-3">{hint}</span> : null}
      </span>
    </label>
  );
}
