import React, { forwardRef } from 'react'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
}

const Input = forwardRef<HTMLInputElement, InputProps>(function Input({ label, error, style, type, inputMode, ...props }, ref) {
  // A stable, non-semantic name keeps mobile browsers/password managers from
  // guessing the field is a credit-card/credential and offering saved data.
  const fieldName = props.name ?? label.toLowerCase().replace(/[^a-z0-9]+/g, '-')
  // Chrome/Gboard force a card/key/location autofill strip onto text & number
  // inputs even with autocomplete=off — but exempt type="search". Render
  // non-sensitive fields as search while preserving the intended keypad via
  // inputMode (e.g. decimal for amounts).
  const resolvedType = type === undefined || type === 'text' || type === 'number' || type === 'email' ? 'search' : type
  const resolvedInputMode = inputMode ?? (type === 'number' ? 'decimal' : type === 'email' ? 'email' : undefined)
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 6, width: '100%', position: 'relative' }}>
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
        ref={ref}
        name={fieldName}
        type={resolvedType}
        inputMode={resolvedInputMode}
        autoComplete="off"
        autoCorrect="off"
        autoCapitalize="off"
        spellCheck={false}
        data-lpignore="true"
        data-1p-ignore="true"
        data-form-type="other"
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
          WebkitAppearance: 'none',
          appearance: 'none',
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
      {/* Absolute so showing/hiding it never changes the field height (modal won't resize). */}
      {error && (
        <span style={{
          position: 'absolute',
          top: '100%',
          left: 4,
          marginTop: 3,
          fontSize: 11,
          lineHeight: 1.2,
          color: 'var(--color-danger)',
          whiteSpace: 'nowrap',
        }}>{error}</span>
      )}
    </label>
  )
})

export default Input
