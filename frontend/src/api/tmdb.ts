// TMDB 图片 URL 工具
const TMDB_IMAGE_BASE = "https://image.tmdb.org/t/p";

function svgDataUri(svg: string): string {
  return `data:image/svg+xml,${encodeURIComponent(svg)}`;
}

// 占位图（无海报/头像时显示）
const PLACEHOLDER_POSTER = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" width="185" height="278" viewBox="0 0 185 278">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="#1e293b"/>
      <stop offset="0.58" stop-color="#0f172a"/>
      <stop offset="1" stop-color="#0f766e"/>
    </linearGradient>
  </defs>
  <rect width="185" height="278" rx="14" fill="url(#bg)"/>
  <rect x="18" y="22" width="149" height="234" rx="10" fill="none" stroke="rgba(255,255,255,.18)"/>
  <path d="M63 118h59a9 9 0 0 1 9 9v37a9 9 0 0 1-9 9H63a9 9 0 0 1-9-9v-37a9 9 0 0 1 9-9Z" fill="rgba(255,255,255,.08)" stroke="rgba(255,255,255,.42)" stroke-width="2"/>
  <path d="m65 160 19-21 16 16 10-11 17 16" fill="none" stroke="#90cea1" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
  <circle cx="113" cy="135" r="7" fill="#01b4e4"/>
  <text x="92.5" y="204" text-anchor="middle" fill="rgba(255,255,255,.78)" font-family="PingFang SC, Microsoft YaHei, sans-serif" font-size="15" font-weight="600">暂无图片</text>
</svg>`);

const PLACEHOLDER_BACKDROP = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" width="500" height="281" viewBox="0 0 500 281">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="#1f2937"/>
      <stop offset="0.62" stop-color="#111827"/>
      <stop offset="1" stop-color="#0e7490"/>
    </linearGradient>
  </defs>
  <rect width="500" height="281" fill="url(#bg)"/>
  <path d="M0 210 108 142l70 45 84-78 238 142v30H0Z" fill="rgba(144,206,161,.16)"/>
  <path d="M0 228 128 160l68 44 82-76 222 128" fill="none" stroke="rgba(255,255,255,.24)" stroke-width="3"/>
  <circle cx="388" cy="72" r="22" fill="rgba(1,180,228,.72)"/>
  <text x="250" y="150" text-anchor="middle" fill="rgba(255,255,255,.78)" font-family="PingFang SC, Microsoft YaHei, sans-serif" font-size="24" font-weight="600">暂无背景图</text>
</svg>`);

const PLACEHOLDER_PROFILE = svgDataUri(`
<svg xmlns="http://www.w3.org/2000/svg" width="185" height="278" viewBox="0 0 185 278">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="#243244"/>
      <stop offset="0.62" stop-color="#111827"/>
      <stop offset="1" stop-color="#0d9488"/>
    </linearGradient>
  </defs>
  <rect width="185" height="278" rx="14" fill="url(#bg)"/>
  <circle cx="92.5" cy="114" r="32" fill="rgba(255,255,255,.16)" stroke="rgba(255,255,255,.42)" stroke-width="2"/>
  <path d="M42 206c7-35 27-53 50.5-53S136 171 143 206" fill="rgba(255,255,255,.1)" stroke="#90cea1" stroke-width="4" stroke-linecap="round"/>
  <text x="92.5" y="232" text-anchor="middle" fill="rgba(255,255,255,.78)" font-family="PingFang SC, Microsoft YaHei, sans-serif" font-size="15" font-weight="600">暂无头像</text>
</svg>`);

export type ImageSize = "w92" | "w154" | "w185" | "w342" | "w500" | "w780" | "original";

function isDirectImagePath(path: string): boolean {
  return (
    path.startsWith("http://") ||
    path.startsWith("https://") ||
    path.startsWith("data:") ||
    path.startsWith("blob:") ||
    path.startsWith("/uploads/")
  );
}

/**
 * 生成 TMDB 图片完整 URL
 * @param path   poster_path / backdrop_path / profile_path
 * @param size   尺寸，默认 w342
 */
export function tmdbImg(path: string | null | undefined, size: ImageSize = "w342"): string {
  if (!path) return size === "original" ? PLACEHOLDER_BACKDROP : PLACEHOLDER_POSTER;
  if (isDirectImagePath(path)) return path;
  return `${TMDB_IMAGE_BASE}/${size}${path}`;
}

/**
 * 人物头像 URL（fallback 不同）
 */
export function profileImg(path: string | null | undefined, size: ImageSize = "w185"): string {
  if (!path) return PLACEHOLDER_PROFILE;
  if (isDirectImagePath(path)) return path;
  return `${TMDB_IMAGE_BASE}/${size}${path}`;
}
