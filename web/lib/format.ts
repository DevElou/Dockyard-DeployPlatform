// fr-FR locale is intentional — this is a personal homelab tool used in France.
const dateFormatter = new Intl.DateTimeFormat("fr-FR", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

const relativeFormatter = new Intl.RelativeTimeFormat("fr-FR", {
  numeric: "auto",
});

export function formatDate(iso: string): string {
  return dateFormatter.format(new Date(iso));
}

export function formatRelative(iso: string): string {
  const diff = (new Date(iso).getTime() - Date.now()) / 1000;
  const absDiff = Math.abs(diff);

  if (absDiff < 60) return relativeFormatter.format(Math.round(diff), "second");
  if (absDiff < 3600) return relativeFormatter.format(Math.round(diff / 60), "minute");
  if (absDiff < 86400) return relativeFormatter.format(Math.round(diff / 3600), "hour");
  return relativeFormatter.format(Math.round(diff / 86400), "day");
}

export function truncate(str: string, max: number): string {
  return str.length > max ? `${str.slice(0, max)}…` : str;
}
