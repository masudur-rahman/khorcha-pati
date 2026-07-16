import { useState, useRef, useEffect, useLayoutEffect, useCallback } from 'react'
import { createPortal } from 'react-dom'
import Input from './Input'

interface Option {
  value: string
  label: string
}

interface SearchableSelectProps {
  label: string
  value: string
  onChange: (value: string) => void
  options: Option[]
  placeholder?: string
  error?: string
}

interface Position {
  left: number
  width: number
  top?: number
  bottom?: number
  maxHeight: number
}

const GAP = 4
const DESIRED_HEIGHT = 240
const VIEWPORT_MARGIN = 8

// SearchableSelect is a filterable combobox. The options list renders in a
// portal with fixed positioning, so it floats over everything without
// belonging to any scroll container — nothing behind it shifts or grows
// scrollable. It flips above the field when there isn't room below, so it
// never runs off-screen or covers a modal footer. Same on mobile and desktop.
export default function SearchableSelect({ label, value, onChange, options, placeholder, error }: SearchableSelectProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [focusedIndex, setFocusedIndex] = useState(-1)
  const [pos, setPos] = useState<Position | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const listboxRef = useRef<HTMLDivElement>(null)

  const selectedOption = options.find(o => o.value === value)
  const displayValue = isOpen ? search : (selectedOption ? selectedOption.label : value)

  const filtered = options.filter(o =>
    o.label.toLowerCase().includes(search.toLowerCase()) ||
    o.value.toLowerCase().includes(search.toLowerCase())
  )

  // Anchor the list to the field, flipping above it when space below is tight.
  const updatePosition = useCallback(() => {
    const el = containerRef.current
    if (!el) return
    const rect = el.getBoundingClientRect()
    const vh = window.innerHeight
    const spaceBelow = vh - rect.bottom
    const spaceAbove = rect.top
    const openUp = spaceBelow < Math.min(DESIRED_HEIGHT, 200) && spaceAbove > spaceBelow
    if (openUp) {
      setPos({
        left: rect.left, width: rect.width,
        bottom: vh - rect.top + GAP,
        maxHeight: Math.min(DESIRED_HEIGHT, spaceAbove - GAP - VIEWPORT_MARGIN),
      })
    } else {
      setPos({
        left: rect.left, width: rect.width,
        top: rect.bottom + GAP,
        maxHeight: Math.min(DESIRED_HEIGHT, spaceBelow - GAP - VIEWPORT_MARGIN),
      })
    }
  }, [])

  // Reposition on open and keep it pinned while any ancestor scrolls / on resize.
  useLayoutEffect(() => {
    if (!isOpen) return
    updatePosition()
    window.addEventListener('scroll', updatePosition, true)
    window.addEventListener('resize', updatePosition)
    return () => {
      window.removeEventListener('scroll', updatePosition, true)
      window.removeEventListener('resize', updatePosition)
    }
  }, [isOpen, updatePosition])

  // Close when clicking outside both the field and the portaled list.
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const t = e.target as Node
      if (containerRef.current?.contains(t)) return
      if (listboxRef.current?.contains(t)) return
      setIsOpen(false)
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Focus the selected option (or first) when opening / filtering.
  useEffect(() => {
    if (!isOpen) return
    const idx = filtered.findIndex(o => o.value === value)
    setFocusedIndex(idx >= 0 ? idx : 0)
  }, [isOpen, search, value, options]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (isOpen && listboxRef.current && focusedIndex >= 0) {
      const el = listboxRef.current.children[focusedIndex] as HTMLElement
      el?.scrollIntoView({ block: 'nearest' })
    }
  }, [focusedIndex, isOpen])

  const commit = (v: string) => {
    onChange(v)
    setIsOpen(false)
    setSearch('')
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (!isOpen) {
      if (e.key === 'ArrowDown' || e.key === 'Enter') {
        setIsOpen(true)
        e.preventDefault()
      }
      return
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setFocusedIndex(prev => (prev < filtered.length - 1 ? prev + 1 : prev))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setFocusedIndex(prev => (prev > 0 ? prev - 1 : prev))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (filtered[focusedIndex]) commit(filtered[focusedIndex].value)
    } else if (e.key === 'Escape') {
      e.preventDefault()
      setIsOpen(false)
    }
  }

  return (
    <div ref={containerRef} style={{ position: 'relative', width: '100%' }}>
      <Input
        label={label}
        type="search"
        value={displayValue}
        onChange={e => {
          setSearch(e.target.value)
          if (!isOpen) setIsOpen(true)
        }}
        onFocus={() => {
          setSearch('')
          setIsOpen(true)
        }}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        error={error}
      />

      {isOpen && pos && createPortal(
        <div
          ref={listboxRef}
          style={{
            position: 'fixed',
            left: pos.left, width: pos.width,
            top: pos.top, bottom: pos.bottom,
            zIndex: 1000,
            background: 'var(--color-surface)',
            border: '1px solid var(--color-border)',
            borderRadius: 12,
            boxShadow: '0 12px 28px rgba(0,0,0,0.16)',
            maxHeight: pos.maxHeight,
            overflowY: 'auto',
            padding: 6,
            display: 'flex', flexDirection: 'column', gap: 2,
          }}
        >
          {filtered.length === 0 ? (
            <div style={{ padding: '12px 14px', fontSize: 13, color: 'var(--color-text-tertiary)' }}>No results found</div>
          ) : (
            filtered.map((o, idx) => {
              const isSelected = o.value === value
              const isFocused = idx === focusedIndex
              return (
                <button
                  key={o.value}
                  type="button"
                  onMouseDown={e => { e.preventDefault(); commit(o.value) }}
                  onMouseMove={() => setFocusedIndex(idx)}
                  style={{
                    padding: '10px 12px',
                    background: isFocused ? 'var(--color-primary-subtle)' : (isSelected ? 'var(--color-hover)' : 'transparent'),
                    border: 'none', borderRadius: 8, cursor: 'pointer',
                    textAlign: 'left', fontSize: 13, fontWeight: 600, lineHeight: 1.4,
                    color: (isSelected || isFocused) ? 'var(--color-primary)' : 'var(--color-text-primary)',
                    whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis',
                    flexShrink: 0,
                  }}
                >
                  {o.label}
                </button>
              )
            })
          )}
        </div>,
        document.body
      )}
    </div>
  )
}
