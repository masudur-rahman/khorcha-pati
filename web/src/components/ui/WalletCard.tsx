import { CSSProperties } from 'react'
import { fmt } from '../../lib/formatter'

export type WalletCardVariant = 'Bank' | 'Cash' | 'Mobile' | 'Credit'

interface Props {
  variant: WalletCardVariant
  name: string
  shortName: string
  balance: number
  trend?: { amount: number; days?: number }
  onClick?: () => void
  style?: CSSProperties
}

const gradients: Record<WalletCardVariant, string> = {
  Bank: 'linear-gradient(135deg, #003D9B 0%, #0052CC 55%, #00B8D9 100%)',
  Cash: 'linear-gradient(135deg, #00875A 0%, #36B37E 100%)',
  Mobile: 'linear-gradient(135deg, #E2136E 0%, #FF5630 100%)',
  Credit: 'linear-gradient(135deg, #172B4D 0%, #0A1628 100%)',
}

const typeLabel: Record<WalletCardVariant, string> = {
  Bank: 'BANK',
  Cash: 'CASH',
  Mobile: 'MOBILE',
  Credit: 'CREDIT',
}

export default function WalletCard({ variant, name, shortName, balance, trend, onClick, style }: Props) {
  const trendPositive = trend && trend.amount >= 0
  return (
    <button
      onClick={onClick}
      style={{
        aspectRatio: '1.586 / 1',
        borderRadius: 16,
        padding: '20px 22px',
        color: 'white',
        position: 'relative',
        overflow: 'hidden',
        background: gradients[variant],
        boxShadow: '0 10px 30px rgba(23,43,77,0.12)',
        border: 'none',
        cursor: onClick ? 'pointer' : 'default',
        textAlign: 'left',
        fontFamily: 'inherit',
        transition: 'transform var(--transition-fast), box-shadow var(--transition-fast)',
        width: '100%',
        display: 'flex',
        flexDirection: 'column',
        ...style,
      }}
      onMouseEnter={e => {
        if (!onClick) return
        e.currentTarget.style.transform = 'translateY(-4px)'
        e.currentTarget.style.boxShadow = '0 18px 40px rgba(23,43,77,0.22)'
      }}
      onMouseLeave={e => {
        if (!onClick) return
        e.currentTarget.style.transform = 'translateY(0)'
        e.currentTarget.style.boxShadow = '0 10px 30px rgba(23,43,77,0.12)'
      }}
    >
      <span aria-hidden style={{ position: 'absolute', top: -40, right: -40, width: 160, height: 160, borderRadius: '50%', background: 'rgba(255,255,255,0.08)' }} />
      <span aria-hidden style={{ position: 'absolute', bottom: -60, right: 30, width: 140, height: 140, borderRadius: '50%', background: 'rgba(255,255,255,0.05)' }} />
      <span aria-hidden style={{ position: 'absolute', inset: 0, boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.15)', borderRadius: 16, pointerEvents: 'none' }} />

      <div style={{ position: 'relative', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <span style={{ fontFamily: 'var(--font-display)', fontWeight: 700, fontSize: 14, letterSpacing: '0.18em' }}>HISAB</span>
        <span style={{
          padding: '3px 10px', borderRadius: 999, fontSize: 10, fontWeight: 700,
          letterSpacing: '0.08em', background: 'rgba(255,255,255,0.18)', color: 'white',
        }}>{typeLabel[variant]}</span>
      </div>

      <div style={{ position: 'relative', marginTop: 'auto', display: 'flex', flexDirection: 'column', gap: 8 }}>
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, letterSpacing: '0.24em', opacity: 0.9 }}>
          {shortName}
        </span>
        <span style={{ fontFamily: 'var(--font-display)', fontSize: 24, fontWeight: 700, letterSpacing: '-0.02em' }}>
          {fmt(balance)}
        </span>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', fontSize: 11, fontWeight: 600, opacity: 0.92 }}>
          <span style={{ textTransform: 'uppercase', letterSpacing: '0.06em' }}>{name}</span>
          {trend && (
            <span>
              {trendPositive ? '↑' : '↓'} {trendPositive ? '+' : '−'}{fmt(Math.abs(trend.amount))}
              {trend.days ? ` · ${trend.days}d` : ''}
            </span>
          )}
        </div>
      </div>
    </button>
  )
}

export function WalletCardGhost({ onClick }: { onClick?: () => void }) {
  return (
    <button
      onClick={onClick}
      style={{
        aspectRatio: '1.586 / 1',
        borderRadius: 16,
        border: '2px dashed var(--color-border)',
        background: 'var(--color-surface)',
        color: 'var(--color-text-tertiary)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: 14,
        fontWeight: 600,
        cursor: onClick ? 'pointer' : 'default',
        width: '100%',
        fontFamily: 'inherit',
        transition: 'all var(--transition-fast)',
      }}
      onMouseEnter={e => {
        if (!onClick) return
        e.currentTarget.style.borderColor = 'var(--color-primary)'
        e.currentTarget.style.color = 'var(--color-primary)'
      }}
      onMouseLeave={e => {
        if (!onClick) return
        e.currentTarget.style.borderColor = 'var(--color-border)'
        e.currentTarget.style.color = 'var(--color-text-tertiary)'
      }}
    >
      + Add Wallet
    </button>
  )
}
