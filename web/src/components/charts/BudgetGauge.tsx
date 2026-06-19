interface BudgetGaugeProps {
  percent: number
  size?: number
  color?: string
  trackColor?: string
  textColor?: string
}

export default function BudgetGauge({
  percent,
  size = 100,
  color,
  trackColor,
  textColor,
}: BudgetGaugeProps) {
  const r = 30
  const circumference = 2 * Math.PI * r
  const clamped = Math.max(0, Math.min(100, percent))
  const dash = (clamped / 100) * circumference

  const resolvedColor = color ?? (
    percent > 90 ? 'var(--color-danger)'
      : percent > 70 ? '#f59e0b'
      : 'var(--color-primary)'
  )
  const resolvedTrack = trackColor ?? 'var(--color-border)'
  const resolvedText = textColor ?? 'var(--color-text-primary)'

  return (
    <svg width={size} height={size} viewBox="0 0 80 80" style={{ display: 'block' }}>
      <circle
        cx={40} cy={40} r={r}
        fill="none" stroke={resolvedTrack} strokeWidth={8}
      />
      <circle
        cx={40} cy={40} r={r}
        fill="none" stroke={resolvedColor} strokeWidth={8}
        strokeLinecap="round"
        strokeDasharray={`${dash} ${circumference}`}
        transform="rotate(-90 40 40)"
        style={{ transition: 'stroke-dasharray 0.8s ease' }}
      />
      <text
        x="40" y="46" textAnchor="middle"
        style={{ fontSize: 16, fontWeight: 700, fill: resolvedText, fontFamily: 'var(--font-display)' }}
      >
        {clamped.toFixed(0)}%
      </text>
    </svg>
  )
}
