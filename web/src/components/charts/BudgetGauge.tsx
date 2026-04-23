interface BudgetGaugeProps {
  percent: number
  size?: number
}

export default function BudgetGauge({ percent, size = 100 }: BudgetGaugeProps) {
  const r = 38
  const circumference = Math.PI * r // semicircle
  const p = Math.min(percent, 100) / 100
  
  const color = percent > 90 
    ? 'var(--color-danger)' 
    : percent > 70 
      ? '#f59e0b' 
      : 'var(--color-primary)'

  return (
    <svg width={size} height={size * 0.65} viewBox="0 0 100 65">
      <path 
        d={`M 12 58 A ${r} ${r} 0 0 1 88 58`} 
        fill="none" stroke="var(--color-border)" strokeWidth="8" strokeLinecap="round" 
      />
      <path 
        d={`M 12 58 A ${r} ${r} 0 0 1 88 58`} 
        fill="none" stroke={color} strokeWidth="8" strokeLinecap="round"
        strokeDasharray={`${circumference * p} ${circumference}`}
        style={{ transition: 'all 0.8s ease' }} 
      />
      <text 
        x="50" y="52" textAnchor="middle" 
        style={{ fontSize: 16, fontWeight: 700, fill: 'var(--color-text-primary)' }}
      >
        {percent.toFixed(0)}%
      </text>
    </svg>
  )
}
