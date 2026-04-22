import { useState } from 'react'
import { ICONS } from '../ui/Icons'
import { useSearch } from '../../context/SearchContext'

interface TopBarProps {
  title: string
  subtitle?: string
}

export default function TopBar({ title, subtitle }: TopBarProps) {
  const { searchTerm, setSearchTerm } = useSearch()
  const [isSearchExpanded, setIsSearchExpanded] = useState(false)

  return (
    <div style={{
      display: 'flex', alignItems: 'center', justifyContent: 'space-between',
      padding: '20px 0', marginBottom: 8,
    }}>
      <div>
        <h1 style={{ 
          fontSize: 26, 
          fontWeight: 700, 
          color: 'var(--color-text-primary)', 
          letterSpacing: '-0.02em', 
          margin: 0 
        }}>
          {title}
        </h1>
        {subtitle && (
          <p style={{ 
            fontSize: 13, 
            color: 'var(--color-text-tertiary)', 
            marginTop: 4 
          }}>
            {subtitle}
          </p>
        )}
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <div style={{
          display: 'flex', alignItems: 'center',
          background: 'var(--color-surface)',
          border: '1px solid var(--color-border)',
          borderRadius: 12,
          padding: '0 12px',
          height: 40,
          width: isSearchExpanded ? 240 : 40,
          transition: 'width 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          overflow: 'hidden',
          cursor: isSearchExpanded ? 'default' : 'pointer',
        }} onClick={() => !isSearchExpanded && setIsSearchExpanded(true)}>
          <span style={{ color: 'var(--color-text-secondary)', display: 'flex', flexShrink: 0 }}>
            {ICONS.search(18)}
          </span>
          {isSearchExpanded && (
            <input
              autoFocus
              value={searchTerm}
              onChange={e => setSearchTerm(e.target.value)}
              onBlur={() => !searchTerm && setIsSearchExpanded(false)}
              placeholder="Search anything..."
              style={{
                background: 'transparent',
                border: 'none',
                outline: 'none',
                fontSize: 13,
                color: 'var(--color-text-primary)',
                marginLeft: 10,
                width: '100%',
              }}
            />
          )}
        </div>
        <button style={{
          width: 40, height: 40, borderRadius: 12, border: '1px solid var(--color-border)',
          background: 'var(--color-surface)', display: 'flex', alignItems: 'center', justifyContent: 'center',
          cursor: 'pointer', color: 'var(--color-text-secondary)', position: 'relative',
        }}>
          {ICONS.bell(18)}
          <span style={{
            position: 'absolute', top: 8, right: 8, width: 7, height: 7, borderRadius: '50%',
            background: 'var(--color-danger)', border: '2px solid var(--color-surface)',
          }} />
        </button>
        <div style={{
          width: 40, height: 40, borderRadius: 12, background: 'var(--color-primary)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          color: 'white', fontWeight: 700, fontSize: 14, cursor: 'pointer',
        }}>
          MR
        </div>
      </div>
    </div>
  )
}
