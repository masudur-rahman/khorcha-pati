export const VALIDATION_RULES = {
  SHORT_NAME: /^[a-zA-Z0-9-_]+$/,
  DISPLAY_NAME: /^[a-zA-Z0-9-_]([a-zA-Z0-9-_ ]*[a-zA-Z0-9-_])?$/,
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
}

export function validateShortName(name: string): string | null {
  if (!name) return 'Required'
  if (!VALIDATION_RULES.SHORT_NAME.test(name)) {
    return 'No spaces or special characters allowed (only letters, numbers, dashes, and underscores).'
  }
  return null
}

export function validateDisplayName(name: string, required = false): string | null {
  if (!name) {
    return required ? 'Required' : null
  }
  if (!VALIDATION_RULES.DISPLAY_NAME.test(name)) {
    return 'Cannot have leading/trailing spaces or special characters (only letters, numbers, dashes, underscores, and spaces).'
  }
  return null
}

// validateEmail permits an empty value but rejects a malformed address.
export function validateEmail(email: string, required = false): string | null {
  if (!email) {
    return required ? 'Required' : null
  }
  if (!VALIDATION_RULES.EMAIL.test(email)) {
    return 'Enter a valid email address (e.g. name@example.com).'
  }
  return null
}
