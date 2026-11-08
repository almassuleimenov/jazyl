// Wordmark: lowercase "jazyl" where the dot of the j is a checkmark (PRODUCT.md brand rule).
export function Logo({ tone = "light", className = "" }: { tone?: "light" | "dark"; className?: string }) {
  const ink = tone === "light" ? "text-white" : "text-brand";
  return (
    <span
      className={`inline-flex items-baseline font-display text-[22px] leading-none font-extrabold tracking-[-0.03em] ${ink} ${className}`}
      aria-label="Jazyl"
      role="img"
    >
      <span className="relative inline-block" aria-hidden>
        {/* dotless j + the checkmark in place of its dot */}
        ȷ
        <svg
          viewBox="0 0 12 10"
          className="absolute -top-[0.36em] left-[0.02em] h-[0.42em] w-[0.5em] text-action"
          fill="none"
          aria-hidden
        >
          <path d="M1.5 5.2 4.6 8.2 10.5 1.8" stroke="currentColor" strokeWidth="2.6" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </span>
      <span aria-hidden>azyl</span>
    </span>
  );
}
