// Color mapping for note card backgrounds
// Tailwind requires explicit class names for static analysis

export const noteColorMap: Record<string, string> = {
  zinc: 'bg-zinc-500/20 text-zinc-950 dark:text-zinc-50',
  rose: 'bg-rose-500/20 text-rose-950 dark:text-rose-50',
  amber: 'bg-amber-500/20 text-amber-950 dark:text-amber-50',
  emerald: 'bg-emerald-500/20 text-emerald-950 dark:text-emerald-50',
  violet: 'bg-violet-500/20 text-violet-950 dark:text-violet-50',
  sky: 'bg-sky-500/20 text-sky-950 dark:text-sky-50',
  red: 'bg-red-500/20 text-red-950 dark:text-red-50',
  orange: 'bg-orange-500/20 text-orange-950 dark:text-orange-50',
  yellow: 'bg-yellow-500/20 text-yellow-950 dark:text-yellow-50',
  lime: 'bg-lime-500/20 text-lime-950 dark:text-lime-50',
  green: 'bg-green-500/20 text-green-950 dark:text-green-50',
  teal: 'bg-teal-500/20 text-teal-950 dark:text-teal-50',
  cyan: 'bg-cyan-500/20 text-cyan-950 dark:text-cyan-50',
  blue: 'bg-blue-500/20 text-blue-950 dark:text-blue-50',
  indigo: 'bg-indigo-500/20 text-indigo-950 dark:text-indigo-50',
  purple: 'bg-purple-500/20 text-purple-950 dark:text-purple-50',
  fuchsia: 'bg-fuchsia-500/20 text-fuchsia-950 dark:text-fuchsia-50',
  pink: 'bg-pink-500/20 text-pink-950 dark:text-pink-50',
}

export function getNoteColorClass(color?: string): string {
  if (!color) return 'bg-muted'
  return noteColorMap[color] || 'bg-muted'
}
