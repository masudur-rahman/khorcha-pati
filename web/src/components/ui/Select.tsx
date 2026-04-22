import React from 'react'

interface Option {
  value: string
  label: string
}

interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label: string
  options: Option[]
  error?: string
}

export default function Select({ label, options, error, style, ...props }: SelectProps) {
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 6, width: '100%' }}>
      <span style={{ 
        fontSize: 10, 
        fontWeight: 700, 
        textTransform: 'uppercase', 
        letterSpacing: '0.08em', 
        color: 'var(--color-text-tertiary)',
        marginLeft: 4,
      }}>
        {label}
      </span>
      <div style={{ position: 'relative' }}>
        <select 
          style={{
            width: '100%',
            background: 'var(--color-bg)',
            border: error ? '1px solid var(--color-danger)' : '1px solid var(--color-border)',
            borderRadius: 12,
            padding: '12px 16px',
            fontSize: 14,
            fontWeight: 500,
            color: 'var(--color-text-primary)',
            outline: 'none',
            transition: 'all 0.15s ease',
            appearance: 'none',
            cursor: 'pointer',
            ...style,
          }}
          onFocus={e => {
              e.currentTarget.style.border = '1px solid var(--color-primary)'
              e.currentTarget.style.boxShadow = '0 0 0 4px var(--color-primary-subtle)'
          }}
          onBlur={e => {
              e.currentTarget.style.border = error ? '1px solid var(--color-danger)' : '1px solid var(--color-border)'
              e.currentTarget.style.boxShadow = 'none'
          }}
          {...props}
        >
          <option value="">Select...</option>
          {options.map(o => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
        <div style={{
          position: 'absolute',
          right: 16,
          top: '50%',
          transform: 'translateY(-50%)',
          pointerEvents: 'none',
          color: 'var(--color-text-tertiary)',
        }}>
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </div>
      </div>
      {error && <span style={{ fontSize: 11, color: 'var(--color-danger)', marginLeft: 4 }}>{error}</span>}
    </label>
  )
}
