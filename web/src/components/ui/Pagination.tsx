import { forwardRef, ReactNode } from 'react'
import Button from './Button'

interface PaginationProps {
  /** e.g. "Showing 1–20 of 145". */
  rangeText: ReactNode
  canPrev: boolean
  canNext: boolean
  onPrev: () => void
  onNext: () => void
}

// Pagination renders the shared "Showing X–Y of Z" bar with Previous/Next controls.
const Pagination = forwardRef<HTMLDivElement, PaginationProps>(function Pagination(
  { rangeText, canPrev, canNext, onPrev, onNext }, ref
) {
  return (
    <div
      ref={ref}
      style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        gap: 16, flexWrap: 'wrap',
        padding: '12px 24px', borderTop: '1px solid var(--color-border)',
      }}
    >
      <span style={{ fontSize: 13, color: 'var(--color-text-tertiary)', fontWeight: 500 }}>{rangeText}</span>
      <div style={{ display: 'flex', gap: 8 }}>
        <Button variant="secondary" disabled={!canPrev} onClick={onPrev} style={{ padding: '6px 12px', fontSize: 13 }}>
          Previous
        </Button>
        <Button variant="secondary" disabled={!canNext} onClick={onNext} style={{ padding: '6px 12px', fontSize: 13 }}>
          Next
        </Button>
      </div>
    </div>
  )
})

export default Pagination
