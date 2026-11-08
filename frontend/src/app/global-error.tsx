"use client";

import "./globals.css";

export default function GlobalError({ retry }: { error: unknown; retry: () => void }) {
  return (
    <html lang="ru">
      <body className="flex min-h-dvh items-center justify-center bg-milk px-4">
        <div className="max-w-md text-center">
          <h1 className="text-xl font-bold text-ink">Кабинет временно недоступен</h1>
          <p className="mt-2 text-sm text-ink-2">Попробуйте обновить страницу через минуту.</p>
          <button onClick={() => retry()} className="mt-5 h-10 rounded-[8px] border border-[#d3cdbf] bg-white px-4 text-sm">
            Обновить
          </button>
        </div>
      </body>
    </html>
  );
}
