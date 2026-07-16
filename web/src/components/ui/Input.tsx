import React, { forwardRef, useId, useState } from 'react'

import { IS_ANDROID, autofillSafeType, keypadFor } from '../../lib/platform'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
}

/* Desktop Chrome ignores autocomplete="off" and heuristically matches fields
   to saved cards/addresses using name/id/label text — e.g. "full-name" hits
   the cardholder-name regex. Layered defenses, all needed:
   1. Random name/id per mount — field names never match autofill regexes.
   2. Unrecognized autocomplete token — recognized tokens invite autofill,
      bare "off" is ignored; unknown values give Chrome nothing to map.
   3. readOnly until first pointer/focus — autofill skips readonly fields,
      so no suggestion popup can attach before the user interacts.
   On Android none of this hides Chrome's keyboard autofill bar — only
   type="search" does (see lib/platform.ts), so fields render as search there.
   Desktop keeps real types; email/tel get their real token so the browser
   offers the designated data, not cards. */
const suppressedAutoCompleteFor = (type: string, uid: string) => {
  if (IS_ANDROID) return `khp-${uid}`
  if (type === 'email') return 'email'
  if (type === 'tel') return 'tel'
  return `khp-${uid}`
}

const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { label, error, style, type = 'text', inputMode, autoComplete, name, id, onFocus, onBlur, onPointerDown, ...props },
  ref,
) {
  const uid = useId().replace(/[^a-zA-Z0-9]/g, '')
  const [armed, setArmed] = useState(false)
  const resolvedType = autofillSafeType(type)
  const resolvedInputMode = inputMode ?? keypadFor(type)
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
        name={name ?? `f-${uid}`}
        id={id ?? `f-${uid}`}
        type={resolvedType}
        inputMode={resolvedInputMode}
        autoComplete={autoComplete ?? suppressedAutoCompleteFor(type, uid)}
        autoCorrect="off"
        autoCapitalize="off"
        spellCheck={false}
        data-lpignore="true"
        data-1p-ignore="true"
        data-form-type="other"
        readOnly={props.readOnly ?? !armed}
        onPointerDown={e => {
          setArmed(true)
          onPointerDown?.(e)
        }}
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
          setArmed(true)
          e.currentTarget.style.border = '1px solid var(--color-primary)'
          e.currentTarget.style.boxShadow = '0 0 0 4px var(--color-primary-subtle)'
          onFocus?.(e)
        }}
        onBlur={e => {
          e.currentTarget.style.border = error ? '1px solid var(--color-danger)' : '1px solid var(--color-border)'
          e.currentTarget.style.boxShadow = 'none'
          onBlur?.(e)
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
