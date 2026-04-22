interface Segment {
  value: number
  color: string
}

interface MiniDonutProps {
  segments: Segment[]
  size?: number
  totalValue?: number
}

export default function MiniDonut({ segments, size = 120, totalValue }: MiniDonutProps) {
  const r = 42
  const cx = 50
  const cy = 50
  const circumference = 2 * Math.PI * r
  let offset = 0
  
  const total = totalValue ?? segments.reduce((s, seg) => s + seg.value, 0)
  
  const formatCurrency = (amount: number) => {
    const abs = Math.abs(amount)
    return '৳' + abs.toLocaleString('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 0 })
  }

  return (
    <svg width={size} height={size} viewBox="0 0 100 100">
      {segments.map((seg, i) => {
        const pct = total > 0 ? seg.value / total : 0
        const dash = circumference * pct
        const gap = circumference - dash
        const el = (
          <circle 
            key={i} cx={cx} cy={cy} r={r}
            fill="none" stroke={seg.color} strokeWidth="12"
            strokeDasharray={`${dash} ${gap}`}
            strokeDashoffset={-offset}
            strokeLinecap="round"
            transform={`rotate(-90 ${cx} ${cy})`}
            style={{ transition: 'all 0.6s ease' }}
          />
        )
        offset += dash
        return el
      })}
      <text 
        x={cx} y={cy - 4} textAnchor="middle" 
        style={{ fontSize: 14, fontWeight: 700, fill: 'var(--color-text-primary)' }}
      >
        {total > 0 ? formatCurrency(total) : '—'}
      </text>
      <text 
        x={cx} y={cy + 10} textAnchor="middle" 
        style={{ fontSize: 7, fill: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em' }}
      >
        total
      </text>
    </svg>
  )
}
