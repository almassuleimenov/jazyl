"use client";

import { useState } from "react";
import { Copy, Check, RefreshCw } from "lucide-react";

import { Button, inputClass, cx } from "./ui";

// No look-alike characters (0/O, 1/l/I) so the password survives being read aloud or retyped.
const alphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789";

function generate(): string {
  const bytes = new Uint32Array(12);
  crypto.getRandomValues(bytes);
  const chars = Array.from(bytes, (b) => alphabet[b % alphabet.length]);
  return [chars.slice(0, 4), chars.slice(4, 8), chars.slice(8, 12)].map((g) => g.join("")).join("-");
}

/** Temporary password chosen by Jazyl/owner and handed over in person; the user replaces it at first login. */
export function TempPassword({ id, name = "password" }: { id: string; name?: string }) {
  const [value, setValue] = useState("");
  const [copied, setCopied] = useState(false);
  return (
    <div className="flex gap-2">
      <input
        id={id}
        name={name}
        type="text"
        required
        minLength={10}
        maxLength={72}
        autoComplete="off"
        spellCheck={false}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        placeholder="Нажмите «Создать»"
        className={cx(inputClass, "num font-mono text-[13px] placeholder:font-sans placeholder:text-sm")}
      />
      <Button type="button" onClick={() => setValue(generate())} aria-label="Создать пароль">
        <RefreshCw className="size-4" aria-hidden />
        Создать
      </Button>
      {value ? (
        <Button
          type="button"
          onClick={async () => {
            await navigator.clipboard.writeText(value);
            setCopied(true);
            setTimeout(() => setCopied(false), 1600);
          }}
          aria-label="Скопировать пароль"
        >
          {copied ? <Check className="size-4 text-ok" aria-hidden /> : <Copy className="size-4" aria-hidden />}
        </Button>
      ) : null}
    </div>
  );
}
