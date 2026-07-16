/* Android Chrome shows a non-dismissable autofill bar (key/card/location)
   above the keyboard for any field it classifies as autofillable — attributes
   like autocomplete cannot hide it (crbug 40856139, WontFix by design). The
   only exempt widget is type="search". Detect Android so fields can render
   as search there (with the real keypad via inputmode) while keeping their
   semantic type everywhere else. */
export const IS_ANDROID = typeof navigator !== 'undefined' && /Android/i.test(navigator.userAgent)

/* Autofill-exempt rendering: on Android map autofillable types to search,
   preserving the intended keypad through inputmode. */
export function autofillSafeType(type: string): string {
  if (!IS_ANDROID) return type
  return type === 'text' || type === 'number' || type === 'email' || type === 'tel' ? 'search' : type
}

/* Keypad hint matching the semantic type (works with any rendered type). */
export function keypadFor(type: string): 'decimal' | 'email' | 'tel' | undefined {
  if (type === 'number') return 'decimal'
  if (type === 'email') return 'email'
  if (type === 'tel') return 'tel'
  return undefined
}
