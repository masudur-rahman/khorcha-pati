export const VALIDATION_RULES = {
  SHORT_NAME: /^[a-zA-Z0-9-_]+$/,
  DISPLAY_NAME: /^[a-zA-Z0-9-_]([a-zA-Z0-9-_ ]*[a-zA-Z0-9-_])?$/,
  // Wallet names additionally allow an internal apostrophe (e.g. "Masud's Savings").
  WALLET_NAME: /^[a-zA-Z0-9-_]([a-zA-Z0-9-_' ]*[a-zA-Z0-9-_])?$/,
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
}

export function validateShortName(name: string): string | null {
  if (!name) return 'Required'
  if (!VALIDATION_RULES.SHORT_NAME.test(name)) {
    return 'Only letters, numbers, - and _ (no spaces).'
  }
  return null
}

export function validateDisplayName(name: string, required = false): string | null {
  if (!name) {
    return required ? 'Required' : null
  }
  if (!VALIDATION_RULES.DISPLAY_NAME.test(name)) {
    return 'Only letters, numbers, spaces, - and _.'
  }
  return null
}

// validateWalletName is like validateDisplayName but allows an internal apostrophe.
export function validateWalletName(name: string, required = false): string | null {
  if (!name) {
    return required ? 'Required' : null
  }
  if (!VALIDATION_RULES.WALLET_NAME.test(name)) {
    return "Only letters, numbers, spaces, - _ and '."
  }
  return null
}

// validateEmail permits an empty value but rejects a malformed address.
export function validateEmail(email: string, required = false): string | null {
  if (!email) {
    return required ? 'Required' : null
  }
  if (!VALIDATION_RULES.EMAIL.test(email)) {
    return 'Enter a valid email (name@example.com).'
  }
  return null
}
