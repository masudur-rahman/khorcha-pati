import { CSSProperties } from 'react'
import { fmt } from '../../lib/formatter'

export type WalletCardVariant = 'Bank' | 'Cash' | 'Mobile' | 'Credit'

interface Props {
  variant: WalletCardVariant
  name: string
  shortName: string
  balance: number
  paletteIndex?: number
  trend?: { amount: number; days?: number }
  onClick?: () => void
  style?: CSSProperties
}

const palettes: Record<WalletCardVariant, string[]> = {
  Bank: [
    'linear-gradient(135deg, #003D9B 0%, #0052CC 55%, #00B8D9 100%)',
    'linear-gradient(135deg, #1E1A78 0%, #4C3FB6 55%, #7C4DFF 100%)',
    'linear-gradient(135deg, #0A4D68 0%, #088395 55%, #05BFDB 100%)',
    'linear-gradient(135deg, #1E3A5F 0%, #2C5282 55%, #3182CE 100%)',
  ],
  Cash: [
    'linear-gradient(135deg, #00875A 0%, #36B37E 100%)',
    'linear-gradient(135deg, #006844 0%, #2BAE66 55%, #7BC47F 100%)',
    'linear-gradient(135deg, #1B5E20 0%, #43A047 55%, #C8E6C9 100%)',
  ],
  Mobile: [
    'linear-gradient(135deg, #E2136E 0%, #FF5630 100%)',
    'linear-gradient(135deg, #FF6F00 0%, #FF9100 55%, #FFC400 100%)',
    'linear-gradient(135deg, #6A1B9A 0%, #C2185B 55%, #FF5630 100%)',
  ],
  Credit: [
    'linear-gradient(135deg, #172B4D 0%, #0A1628 100%)',
    'linear-gradient(135deg, #1A1A2E 0%, #16213E 55%, #0F3460 100%)',
  ],
}

const typeLabel: Record<WalletCardVariant, string> = {
  Bank: 'BANK',
  Cash: 'CASH',
  Mobile: 'MOBILE',
  Credit: 'CREDIT',
}

const PATTERNS = [
  '',
  'radial-gradient(circle at 88% 18%, rgba(255,255,255,0.14) 0 12%, transparent 13%), radial-gradient(circle at 92% 30%, rgba(255,255,255,0.10) 0 8%, transparent 9%)',
  'linear-gradient(115deg, transparent 0 55%, rgba(255,255,255,0.10) 55% 60%, transparent 60% 70%, rgba(255,255,255,0.07) 70% 73%, transparent 73%)',
  'repeating-linear-gradient(45deg, transparent 0 14px, rgba(255,255,255,0.05) 14px 15px)',
]

export function inferVariant(type: string, name: string, shortName: string): WalletCardVariant {
  const hay = `${name} ${shortName}`.toLowerCase()
  if (/bkash|nagad|rocket|upay|mcash|tap|tap pay|sure ?cash/.test(hay)) return 'Mobile'
  if (/credit|amex|visa|master|card/.test(hay)) return 'Credit'
  if (type === 'Bank') return 'Bank'
  return 'Cash'
}

export default function WalletCard({ variant, name, shortName, balance, paletteIndex = 0, trend, onClick, style }: Props) {
  const palette = palettes[variant]
  const bg = palette[paletteIndex % palette.length]
  const pattern = PATTERNS[paletteIndex % PATTERNS.length]
  const trendPositive = trend && trend.amount >= 0

  return (
    <button
      onClick={onClick}
      className="khp-wallet-card"
      style={{
        aspectRatio: '1.586 / 1',
        borderRadius: 16,
        padding: 'clamp(14px, 4.5cqi, 22px)',
        color: 'white',
        position: 'relative',
        overflow: 'hidden',
        containerType: 'inline-size',
        background: bg,
        boxShadow: '0 10px 30px rgba(23,43,77,0.12)',
        border: 'none',
        cursor: onClick ? 'pointer' : 'default',
        textAlign: 'left',
        fontFamily: 'inherit',
        transition: 'transform var(--transition-fast), box-shadow var(--transition-fast)',
        width: '100%',
        display: 'flex',
        flexDirection: 'column',
        minWidth: 0,
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
      <span aria-hidden style={{ position: 'absolute', inset: 0, background: pattern, pointerEvents: 'none' }} />
      <span aria-hidden style={{ position: 'absolute', top: -40, right: -40, width: 160, height: 160, borderRadius: '50%', background: 'rgba(255,255,255,0.10)' }} />
      <span aria-hidden style={{ position: 'absolute', bottom: -60, right: 30, width: 140, height: 140, borderRadius: '50%', background: 'rgba(255,255,255,0.06)' }} />
      <span aria-hidden style={{
        position: 'absolute', inset: 0, borderRadius: 16, pointerEvents: 'none',
        background: 'linear-gradient(135deg, rgba(255,255,255,0.18) 0%, rgba(255,255,255,0) 38%, rgba(255,255,255,0) 62%, rgba(255,255,255,0.10) 100%)',
      }} />
      <span aria-hidden style={{ position: 'absolute', inset: 0, boxShadow: 'inset 0 1px 0 rgba(255,255,255,0.25), inset 0 -1px 0 rgba(0,0,0,0.10)', borderRadius: 16, pointerEvents: 'none' }} />

      <div style={{ position: 'relative', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8, minWidth: 0 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'clamp(3px, 1cqi, 5px)', minWidth: 0, overflow: 'hidden' }}>
          <span style={{ fontFamily: 'var(--font-display)', fontWeight: 700, fontSize: 'clamp(12px, 3.6cqi, 15px)', letterSpacing: '0.04em', lineHeight: 1.05, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>Khorcha-Pati</span>
          <span style={{ fontSize: 'clamp(8px, 2.2cqi, 10px)', fontWeight: 500, fontStyle: 'italic', letterSpacing: '0.01em', opacity: 0.85, lineHeight: 1.2, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
            Keep your khorcha on track
          </span>
        </div>
        <span style={{
          padding: 'clamp(2px, 0.6cqi, 3px) clamp(7px, 2.2cqi, 10px)',
          borderRadius: 999, fontSize: 'clamp(8px, 2.4cqi, 10px)', fontWeight: 700,
          letterSpacing: '0.08em',
          background: 'rgba(255,255,255,0.22)',
          color: 'white',
          backdropFilter: 'blur(4px)',
          border: '1px solid rgba(255,255,255,0.25)',
          flexShrink: 0,
          whiteSpace: 'nowrap',
        }}>{typeLabel[variant]}</span>
      </div>

      {/* Frosted-glass chip + contactless — palette-agnostic */}
      <div aria-hidden style={{ position: 'relative', display: 'flex', alignItems: 'center', gap: 'clamp(8px, 2.4cqi, 12px)', marginTop: 'clamp(16px, 5cqi, 26px)' }}>
        <span style={{
          width: 'clamp(28px, 8cqi, 36px)', height: 'clamp(20px, 5.8cqi, 26px)', borderRadius: 5,
          background: 'linear-gradient(135deg, rgba(255,255,255,0.45) 0%, rgba(255,255,255,0.18) 55%, rgba(255,255,255,0.32) 100%)',
          boxShadow: 'inset 0 0 0 1px rgba(255,255,255,0.40), inset 0 1px 0 rgba(255,255,255,0.55), 0 1px 3px rgba(0,0,0,0.18)',
          backdropFilter: 'blur(2px)',
          display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gridTemplateRows: 'repeat(3, 1fr)', gap: 1.5, padding: 'clamp(3px, 0.9cqi, 4px)',
          flexShrink: 0,
        }}>
          {[...Array(9)].map((_, i) => (
            <span key={i} style={{ background: 'rgba(255,255,255,0.28)', borderRadius: 1, boxShadow: 'inset 0 0 0 0.5px rgba(255,255,255,0.40)' }} />
          ))}
        </span>
        <svg width="clamp(16px, 4.5cqi, 20px)" height="clamp(16px, 4.5cqi, 20px)" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" style={{ opacity: 0.85, transform: 'rotate(90deg)', flexShrink: 0 }}>
          <path d="M5 12a7 7 0 0 1 7-7" />
          <path d="M5 17a12 12 0 0 1 12-12" />
          <path d="M5 7a17 17 0 0 1 17 -2" />
        </svg>
      </div>

      <div style={{ position: 'relative', marginTop: 'auto', display: 'flex', flexDirection: 'column', gap: 'clamp(6px, 1.8cqi, 10px)', minWidth: 0 }}>
        <span style={{
          fontFamily: 'var(--font-mono)', fontSize: 'clamp(10px, 3cqi, 13px)',
          letterSpacing: '0.28em', opacity: 0.92, textShadow: '0 1px 2px rgba(0,0,0,0.15)',
          lineHeight: 1.2, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis',
        }}>
          {shortName}
        </span>
        <span style={{
          fontFamily: 'var(--font-display)', fontSize: 'clamp(18px, 6cqi, 26px)', fontWeight: 700,
          letterSpacing: '-0.02em', textShadow: '0 1px 4px rgba(0,0,0,0.18)',
          lineHeight: 1.05, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis',
        }}>
          {fmt(balance)}
        </span>
        <div style={{
          display: 'flex', justifyContent: 'space-between', alignItems: 'baseline',
          fontSize: 'clamp(9px, 2.6cqi, 11px)', fontWeight: 700, opacity: 0.94,
          gap: 8, minWidth: 0,
        }}>
          <span style={{
            textTransform: 'uppercase', letterSpacing: '0.08em',
            whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', minWidth: 0,
          }}>{name}</span>
          {trend && (
            <span style={{ whiteSpace: 'nowrap', flexShrink: 0 }}>
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
        background:
          'linear-gradient(135deg, var(--color-surface) 0%, var(--color-bg) 100%)',
        color: 'var(--color-text-tertiary)',
        position: 'relative',
        overflow: 'hidden',
        padding: '20px 22px',
        cursor: onClick ? 'pointer' : 'default',
        width: '100%',
        fontFamily: 'inherit',
        textAlign: 'left',
        transition: 'all var(--transition-fast)',
        display: 'flex',
        flexDirection: 'column',
      }}
      onMouseEnter={e => {
        if (!onClick) return
        e.currentTarget.style.borderColor = 'var(--color-primary)'
        e.currentTarget.style.color = 'var(--color-primary)'
        e.currentTarget.style.background =
          'linear-gradient(135deg, var(--color-primary-subtle) 0%, var(--color-surface) 100%)'
        e.currentTarget.style.transform = 'translateY(-4px)'
        e.currentTarget.style.boxShadow = '0 14px 32px rgba(0, 82, 204, 0.16)'
      }}
      onMouseLeave={e => {
        if (!onClick) return
        e.currentTarget.style.borderColor = 'var(--color-border)'
        e.currentTarget.style.color = 'var(--color-text-tertiary)'
        e.currentTarget.style.background =
          'linear-gradient(135deg, var(--color-surface) 0%, var(--color-bg) 100%)'
        e.currentTarget.style.transform = 'translateY(0)'
        e.currentTarget.style.boxShadow = 'none'
      }}
    >
      {/* Faint card-skeleton silhouette */}
      <span aria-hidden style={{
        position: 'absolute', top: -28, right: -28, width: 120, height: 120,
        borderRadius: '50%', background: 'currentColor', opacity: 0.04,
      }} />
      <span aria-hidden style={{
        position: 'absolute', bottom: -36, left: -16, width: 100, height: 100,
        borderRadius: '50%', background: 'currentColor', opacity: 0.03,
      }} />

      {/* Top row — placeholder brand line + type chip outline */}
      <div style={{ position: 'relative', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <span style={{
          fontFamily: 'var(--font-display)', fontWeight: 700, fontSize: 13, letterSpacing: '0.08em', opacity: 0.55,
        }}>Khorcha-Pati</span>
        <span style={{
          padding: '3px 10px', borderRadius: 999, fontSize: 9, fontWeight: 700,
          letterSpacing: '0.08em', border: '1px dashed currentColor', opacity: 0.55,
        }}>NEW</span>
      </div>

      {/* Center plus medallion */}
      <div style={{ position: 'relative', flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <span style={{
          width: 52, height: 52, borderRadius: '50%',
          border: '2px dashed currentColor',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: 'color-mix(in srgb, var(--color-surface) 70%, transparent)',
          backdropFilter: 'blur(2px)',
          transition: 'all var(--transition-fast)',
        }}>
          <svg width={22} height={22} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.5} strokeLinecap="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
        </span>
      </div>

      {/* Bottom — label + subtitle */}
      <div style={{ position: 'relative', display: 'flex', flexDirection: 'column', gap: 4 }}>
        <span style={{ fontSize: 14, fontWeight: 700, letterSpacing: '-0.01em' }}>Add Wallet</span>
        <span style={{ fontSize: 10, fontWeight: 600, letterSpacing: '0.06em', textTransform: 'uppercase', opacity: 0.7 }}>
          Track a new account
        </span>
      </div>
    </button>
  )
}
