import { useState, useRef, useEffect } from 'react'
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
}

export default function SearchableSelect({ label, value, onChange, options, placeholder }: SearchableSelectProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [focusedIndex, setFocusedIndex] = useState(-1)
  const containerRef = useRef<HTMLDivElement>(null)
  const listboxRef = useRef<HTMLDivElement>(null)

  // Find the selected option's label to display when not searching
  const selectedOption = options.find(o => o.value === value)
  const displayValue = isOpen ? search : (selectedOption ? selectedOption.label : value)

  const filtered = options.filter(o => o.label.toLowerCase().includes(search.toLowerCase()) || o.value.toLowerCase().includes(search.toLowerCase()))

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Auto-focus the currently selected value when opening
  useEffect(() => {
    if (isOpen) {
      const idx = filtered.findIndex(o => o.value === value)
      setFocusedIndex(idx >= 0 ? idx : 0)
    }
  }, [isOpen, search, value, options])

  // Scroll focused item into view
  useEffect(() => {
    if (isOpen && listboxRef.current && focusedIndex >= 0) {
      const el = listboxRef.current.children[focusedIndex] as HTMLElement
      if (el) {
        el.scrollIntoView({ block: 'nearest' })
      }
    }
  }, [focusedIndex, isOpen])

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
      if (filtered[focusedIndex]) {
        onChange(filtered[focusedIndex].value)
        setIsOpen(false)
        setSearch('')
      }
    } else if (e.key === 'Escape') {
      e.preventDefault()
      setIsOpen(false)
    }
  }

  return (
    <div ref={containerRef} style={{ position: 'relative', width: '100%', maxWidth: 300 }}>
      <Input
        label={label}
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
      />
      
      {isOpen && (
        <div 
          ref={listboxRef}
          style={{
            position: 'absolute', bottom: 0, left: 'calc(100% + 12px)',
            width: 300,
            background: 'var(--color-surface)',
            border: '1px solid var(--color-border)', borderRadius: 12,
            boxShadow: '0 10px 24px rgba(0,0,0,0.1)',
            maxHeight: 280, overflowY: 'auto', zIndex: 50,
            display: 'flex', flexDirection: 'column', padding: 8,
          }}
        >
          {filtered.length === 0 ? (
            <div style={{ padding: '12px 16px', fontSize: 13, color: 'var(--color-text-tertiary)' }}>No results found</div>
          ) : (
            filtered.map((o, idx) => {
              const isSelected = o.value === value
              const isFocused = idx === focusedIndex
              return (
                <button
                  key={o.value}
                  onClick={() => {
                    onChange(o.value)
                    setIsOpen(false)
                    setSearch('')
                  }}
                  style={{
                    padding: '10px 12px', 
                    background: isFocused ? 'var(--color-primary-subtle)' : (isSelected ? 'var(--color-hover)' : 'transparent'),
                    border: 'none', borderRadius: 8, cursor: 'pointer',
                    textAlign: 'left', fontSize: 13, fontWeight: 600,
                    color: (isSelected || isFocused) ? 'var(--color-primary)' : 'var(--color-text-primary)',
                    transition: 'none', // Removed transition to make keyboard nav feel snappy
                  }}
                  onMouseEnter={() => setFocusedIndex(idx)}
                >
                  {o.label}
                </button>
              )
            })
          )}
        </div>
      )}
    </div>
  )
}
