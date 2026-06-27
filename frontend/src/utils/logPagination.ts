export function normalizeNumber(value: number, min: number, max: number) {
  const next = Number.isFinite(value) ? Math.trunc(value) : min;
  if (next < min) return min;
  if (next > max) return max;
  return next;
}

export function totalPages(total: number, pageSize: number) {
  const safeTotal = Math.max(0, Number(total) || 0);
  const safePageSize = normalizeNumber(Number(pageSize) || 20, 1, 100);
  return Math.max(1, Math.ceil(safeTotal / safePageSize));
}

export function detailTotalPages(total: number, pageSize: number) {
  const safeTotal = Math.max(0, Number(total) || 0);
  const safePageSize = normalizeNumber(Number(pageSize) || 10, 1, 100);
  return Math.max(1, Math.ceil(safeTotal / safePageSize));
}
