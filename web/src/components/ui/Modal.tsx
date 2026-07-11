import React, { useEffect } from 'react'

interface ModalProps {
  children: React.ReactNode
  onClose: () => void
  title?: string
  subtitle?: string
  /** Rendered in the header, to the left of the close button (e.g. edit/delete). */
  headerActions?: React.ReactNode
  /** Sticky action bar pinned to the bottom of the modal. */
  footer?: React.ReactNode
  width?: number | string
  /** Apply the standard body padding. Turn off when the body manages its own. */
  padded?: boolean
}

export default function Modal({
  children,
  onClose,
  title,
  subtitle,
  headerActions,
  footer,
  width = 560,
  padded = true,
}: ModalProps) {
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleEsc)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      window.removeEventListener('keydown', handleEsc)
      document.body.style.overflow = prevOverflow
    }
  }, [onClose])

  return (
    <div
      className="modal-overlay-resp"
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(9, 9, 11, 0.6)',
        backdropFilter: 'blur(4px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 100,
        padding: 20,
      }}
      onClick={onClose}
    >
      <div
        className="modal-container-resp"
        style={{
          background: 'var(--color-surface)',
          borderRadius: 24,
          boxShadow: '0 20px 40px rgba(0,0,0,0.18)',
          width: '100%',
          maxWidth: width,
          display: 'flex',
          flexDirection: 'column',
          position: 'relative',
          overflow: 'hidden',
          animation: 'modal-in 0.2s cubic-bezier(.4,0,.2,1)',
        }}
        onClick={e => e.stopPropagation()}
      >
        {(title || headerActions) && (
          <div className="modal-header-resp" style={{
            padding: '20px 32px',
            borderBottom: '1px solid var(--color-border)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: 12,
            flexShrink: 0,
          }}>
            <div style={{ minWidth: 0 }}>
              {title && (
                <h2 style={{ fontSize: 18, fontWeight: 700, color: 'var(--color-text-primary)', margin: 0, fontFamily: 'var(--font-display)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{title}</h2>
              )}
              {subtitle && (
                <p style={{ fontSize: 12, fontWeight: 600, color: 'var(--color-text-tertiary)', margin: '3px 0 0', textTransform: 'uppercase', letterSpacing: '0.06em', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{subtitle}</p>
              )}
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
              {headerActions}
              <button
                onClick={onClose}
                aria-label="Close"
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: 8,
                  border: 'none',
                  background: 'transparent',
                  cursor: 'pointer',
                  color: 'var(--color-text-tertiary)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  transition: 'all 0.15s',
                }}
                onMouseEnter={e => e.currentTarget.style.background = 'var(--color-hover)'}
                onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>
          </div>
        )}
        <div className={padded ? 'modal-body' : 'modal-body modal-body--flush'}>
          {children}
        </div>
        {footer && (
          <div className="modal-footer-resp" style={{
            padding: '16px 32px',
            borderTop: '1px solid var(--color-border)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-end',
            gap: 12,
            flexShrink: 0,
            background: 'var(--color-surface)',
          }}>
            {footer}
          </div>
        )}
      </div>
      <style>{`
        @keyframes modal-in {
          from { opacity: 0; transform: translateY(10px) scale(0.98); }
          to { opacity: 1; transform: translateY(0) scale(1); }
        }
      `}</style>
    </div>
  )
}
