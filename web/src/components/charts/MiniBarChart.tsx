interface BarData {
  label: string
  income: number
  expense: number
}

interface MiniBarChartProps {
  data: BarData[]
  size?: { w: number, h: number }
}

export default function MiniBarChart({ data, size = { w: 260, h: 120 } }: MiniBarChartProps) {
  const max = Math.max(...data.map(d => Math.max(d.income, d.expense)), 1)
  const barW = 12
  const gap = (size.w - data.length * barW * 2.5) / (data.length + 1)
  
  return (
    <svg width={size.w} height={size.h + 24} viewBox={`0 0 ${size.w} ${size.h + 24}`}>
      {data.map((d, i) => {
        const x = gap + i * (barW * 2.5 + gap)
        const hI = (d.income / max) * (size.h - 10)
        const hE = (d.expense / max) * (size.h - 10)
        return (
          <g key={i}>
            <rect 
              x={x} y={size.h - hI} 
              width={barW} height={hI} rx={4} 
              fill="var(--color-success)" opacity={0.7} 
            />
            <rect 
              x={x + barW + 3} y={size.h - hE} 
              width={barW} height={hE} rx={4} 
              fill="var(--color-danger)" opacity={0.7} 
            />
            <text 
              x={x + barW} y={size.h + 16} textAnchor="middle" 
              style={{ fontSize: 8, fill: 'var(--color-text-tertiary)' }}
            >
              {d.label}
            </text>
          </g>
        )
      })}
    </svg>
  )
}
