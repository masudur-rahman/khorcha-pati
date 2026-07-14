import { ReactNode } from 'react'
import Eyebrow from './Eyebrow'
import { ICONS } from './Icons'

interface Props {
  label: string
  value: string
  accent: string
  icon?: ReactNode
  hint?: string
  /** Optional delta badge, e.g. "12%". Pair with trendUp for direction/color. */
  trend?: string
  trendUp?: boolean
}

// MetricChip is the single stat-card primitive: label + accent value on a surface
// with a left accent bar, plus an optional icon, delta badge and hint line.
export default function MetricChip({ label, value, accent, icon, hint, trend, trendUp }: Props) {
  return (
    <div
      className="metric-chip"
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
          <span
            className="metric-chip-icon"
            style={{
              width: 28,
              height: 28,
              borderRadius: 8,
              background: `color-mix(in srgb, ${accent} 14%, transparent)`,
              color: accent,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
            }}
          >
            {icon}
          </span>
        )}
      </div>
      <div
        className="metric-chip-value"
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
      {trend && (
        <span
          className="metric-chip-trend"
          style={{
            alignSelf: 'flex-start',
            display: 'inline-flex',
            alignItems: 'center',
            gap: 2,
            fontSize: 12,
            fontWeight: 600,
            color: trendUp ? 'var(--color-success)' : 'var(--color-danger)',
            background: trendUp ? 'var(--color-success-subtle)' : 'var(--color-danger-subtle)',
            padding: '2px 8px',
            borderRadius: 6,
          }}
        >
          {trendUp ? ICONS.arrowUp(12) : ICONS.arrowDown(12)}
          {trend}
        </span>
      )}
      {hint && (
        <span className="metric-chip-hint" style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 500 }}>{hint}</span>
      )}
    </div>
  )
}
