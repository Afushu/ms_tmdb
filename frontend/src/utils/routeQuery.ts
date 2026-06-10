export function readQueryString(value: unknown): string {
  if (Array.isArray(value)) return String(value[0] ?? "").trim();
  return String(value ?? "").trim();
}

export function queryValue(value: unknown): string {
  return Array.isArray(value) ? String(value[0] ?? "") : String(value ?? "");
}

export function isSameQuery(currentQuery: Record<string, unknown>, nextQuery: Record<string, unknown>): boolean {
  const keys = new Set([...Object.keys(currentQuery), ...Object.keys(nextQuery)]);
  for (const key of keys) {
    if (queryValue(currentQuery[key]) !== queryValue(nextQuery[key])) return false;
  }
  return true;
}
