"use client";

import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { searchCompanies } from "@/client/sdk.gen";
import type { SearchResult } from "@/client/types.gen";

interface Props {
  onSelect: (result: SearchResult) => void;
}

export function SearchBox({ onSelect }: Props) {
  const [input, setInput] = useState("");
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const { data, isFetching } = useQuery({
    queryKey: ["search", input],
    queryFn: () =>
      searchCompanies({
        query: { q: input, country: "uk" },
        throwOnError: true,
      }).then((r) => r.data ?? []),
    enabled: input.trim().length >= 2,
    staleTime: 60_000,
  });

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      )
        setOpen(false);
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  function handleSelect(result: SearchResult) {
    setInput(result.name);
    setOpen(false);
    onSelect(result);
  }

  const results = data ?? [];
  const showDropdown = open && input.trim().length >= 2;

  return (
    <div ref={containerRef} className="relative w-full max-w-xl">
      <div className="relative">
        <input
          type="text"
          value={input}
          onChange={(e) => {
            setInput(e.target.value);
            setOpen(true);
          }}
          onFocus={() => setOpen(true)}
          placeholder="Search a UK company…"
          className="w-full rounded-lg border border-[--color-border] bg-[--color-manila] px-4 py-3 pr-10 text-sm text-[--color-navy] shadow-sm outline-none placeholder:text-stone-400 focus:border-[--color-string] focus:ring-2 focus:ring-[--color-string]/20"
          autoComplete="off"
          spellCheck={false}
        />
        {isFetching && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2">
            <div className="h-4 w-4 animate-spin rounded-full border-2 border-[--color-string] border-t-transparent" />
          </div>
        )}
      </div>

      {showDropdown && results.length > 0 && (
        <ul
          className="absolute z-50 mt-1 w-full overflow-y-auto rounded-lg border border-stone-200 bg-white shadow-xl"
          style={{ maxHeight: 320 }}
        >
          {results.map((r) => (
            <li key={r.company_number}>
              <button
                type="button"
                onMouseDown={() => handleSelect(r)}
                className="flex w-full flex-col px-4 py-3 text-left hover:bg-stone-50"
              >
                <span className="flex items-center gap-2">
                  <span className="text-sm font-medium text-[--color-navy]">
                    {r.name}
                  </span>
                  <StatusBadge status={r.status} />
                </span>
                <span className="mt-0.5 text-xs text-stone-500">
                  {r.company_number}
                  {r.address_snippet ? ` · ${r.address_snippet}` : ""}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}

      {showDropdown && !isFetching && results.length === 0 && (
        <div className="absolute z-50 mt-1 w-full rounded-lg border border-stone-200 bg-white px-4 py-3 text-sm text-stone-500 shadow-xl">
          No companies found.
        </div>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const active = status?.toLowerCase() === "active";
  return (
    <span
      className={`rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide ${
        active
          ? "bg-emerald-100 text-emerald-700"
          : "bg-stone-200 text-stone-500"
      }`}
    >
      {status}
    </span>
  );
}
