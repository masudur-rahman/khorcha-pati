import React from 'react'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
}

export default function Input({ label, error, style, ...props }: InputProps) {
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
      <input 
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
      />
      {error && <span style={{ fontSize: 11, color: 'var(--color-danger)', marginLeft: 4 }}>{error}</span>}
    </label>
  )
}
