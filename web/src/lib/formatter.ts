// Khorcha-Pati money formatter: Bangladeshi Taka (৳), lakh/crore grouping
// (1,00,000), no decimals, real minus sign (−) for negatives.
// Mirrors models.FormatMoneyValue on the Go side so bot, dashboard and both
// PDF paths render amounts identically.
export const fmt = (n: number) => {
  const v = Math.round(n || 0)
  return `${v < 0 ? '−' : ''}৳${groupBD(Math.abs(v))}`
}

// groupBD inserts Bangladeshi separators: last 3 digits, then groups of 2.
function groupBD(n: number): string {
  const s = String(n)
  if (s.length <= 3) return s
  const head = s.slice(0, -3)
  const tail = s.slice(-3)
  const parts: string[] = []
  let h = head
  while (h.length > 2) {
    parts.unshift(h.slice(-2))
    h = h.slice(0, -2)
  }
	parts.unshift(h)
	return `${parts.join(',')},${tail}`
}

// Format a date with an explicit timezone
export const formatDate = (
  date: Date | string | number,
  opts: Intl.DateTimeFormatOptions = {},
  timeZone = 'UTC'
) => {
  const d = new Date(date)
  return d.toLocaleDateString('en-US', { ...opts, timeZone: timeZone || 'UTC' })
}

// Convert a UTC timestamp into a timezone-adjusted local string (YYYY-MM-DD)
export const toTZString = (date: Date, timeZone: string) => {
  const d = new Date(date.toLocaleString('en-US', { timeZone: timeZone || 'UTC' }))
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
