/** Loading placeholder that keeps the page's shape (header, figures, table) to avoid layout jumps. */
export function PageSkeleton({ stats = false, rows = 6 }: { stats?: boolean; rows?: number }) {
  const bar = "animate-pulse rounded-[6px] bg-line/70";
  return (
    <div aria-busy="true" aria-label="Загрузка">
      <div className="mb-6">
        <div className={`${bar} h-7 w-48`} />
        <div className={`${bar} mt-3 h-4 w-80 max-w-full`} />
      </div>
      {stats ? (
        <div className="mb-6 grid gap-px overflow-hidden rounded-panel border border-line bg-line [grid-template-columns:repeat(auto-fit,minmax(150px,1fr))]">
          {Array.from({ length: 4 }, (_, i) => (
            <div key={i} className="bg-surface px-5 py-4">
              <div className={`${bar} h-3.5 w-24`} />
              <div className={`${bar} mt-3 h-7 w-20`} />
            </div>
          ))}
        </div>
      ) : null}
      <div className="rounded-panel border border-line bg-surface">
        {Array.from({ length: rows }, (_, i) => (
          <div key={i} className="flex items-center gap-4 border-b border-line px-5 py-4 last:border-b-0">
            <div className={`${bar} h-4 w-14`} />
            <div className={`${bar} h-4 flex-1`} />
            <div className={`${bar} h-6 w-24 rounded-full`} />
          </div>
        ))}
      </div>
    </div>
  );
}
