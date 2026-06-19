import { ReactNode } from 'react'
import Eyebrow from './Eyebrow'

interface Props {
  label: string
  value: string
  accent: string
  icon?: ReactNode
  hint?: string
}

export default function MetricChip({ label, value, accent, icon, hint }: Props) {
  return (
    <div
      style={{
        background: 'var(--color-surface)',
        border: '1px solid var(--color-border)',
        borderLeft: `4px solid ${accent}`,
        borderRadius: 'var(--radius-md)',
        padding: '14px 18px',
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
        minWidth: 0,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Eyebrow>{label}</Eyebrow>
        {icon && (
          <span style={{ display: 'flex', color: accent, opacity: 0.85 }}>{icon}</span>
        )}
      </div>
      <div
        style={{
          fontSize: 22,
          fontWeight: 700,
          letterSpacing: '-0.02em',
          color: accent,
          fontFamily: 'var(--font-display)',
          lineHeight: 1.1,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
        }}
      >
        {value}
      </div>
      {hint && (
        <span style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 500 }}>{hint}</span>
      )}
    </div>
  )
}
