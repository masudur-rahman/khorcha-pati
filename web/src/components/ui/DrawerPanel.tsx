import { ReactNode, useEffect } from 'react'
import { ICONS } from './Icons'

interface Props {
  title: string
  subtitle?: string
  onClose: () => void
  width?: number
  children: ReactNode
  footer?: ReactNode
}

export default function DrawerPanel({ title, subtitle, onClose, width = 480, children, footer }: Props) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', onKey)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = prevOverflow
    }
  }, [onClose])

  return (
    <>
      <div
        onClick={onClose}
        style={{
          position: 'fixed', inset: 0, background: 'rgba(15,23,42,0.45)',
          backdropFilter: 'blur(2px)', zIndex: 300,
          animation: 'drawerFade 0.18s ease-out',
        }}
      />
      <aside
        role="dialog"
        aria-modal="true"
        aria-label={title}
        style={{
          position: 'fixed', top: 0, right: 0, height: '100%',
          width: `min(${width}px, 100vw)`,
          background: 'var(--color-surface)',
          boxShadow: '-16px 0 48px rgba(0,0,0,0.18)',
          zIndex: 301,
          display: 'flex', flexDirection: 'column',
          animation: 'drawerIn 0.22s cubic-bezier(0.4,0,0.2,1)',
        }}
      >
        <header
          style={{
            display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            padding: '18px 24px', borderBottom: '1px solid var(--color-border)',
            flexShrink: 0,
          }}
        >
          <div style={{ minWidth: 0 }}>
            <h2 style={{ margin: 0, fontSize: 16, fontWeight: 700, color: 'var(--color-text-primary)', fontFamily: 'var(--font-display)' }}>{title}</h2>
            {subtitle && (
              <p style={{ margin: '2px 0 0', fontSize: 12, color: 'var(--color-text-tertiary)' }}>{subtitle}</p>
            )}
          </div>
          <button
            onClick={onClose}
            aria-label="Close"
            style={{
              width: 36, height: 36, borderRadius: 'var(--radius-sm)',
              border: '1px solid var(--color-border)', background: 'var(--color-surface)',
              color: 'var(--color-text-secondary)', cursor: 'pointer',
              display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
            }}
          >
            {ICONS.x(16)}
          </button>
        </header>
        <div style={{ flex: 1, overflowY: 'auto', padding: '20px 24px' }}>{children}</div>
        {footer && (
          <footer
            style={{
              padding: '14px 24px',
              borderTop: '1px solid var(--color-border)',
              background: 'var(--color-bg)',
              flexShrink: 0,
            }}
          >
            {footer}
          </footer>
        )}
      </aside>
      <style>{`
        @keyframes drawerFade { from { opacity: 0; } to { opacity: 1; } }
        @keyframes drawerIn { from { transform: translateX(100%); } to { transform: translateX(0); } }
        @media (max-width: 640px) {
          aside[role="dialog"] { width: 100vw !important; }
        }
      `}</style>
    </>
  )
}
