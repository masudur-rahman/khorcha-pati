import toast from 'react-hot-toast'

// Standard toast templates so every mutation reports the same way. Success is a
// terse past-tense sentence ("Wallet \"Cash\" created successfully."); the entity
// name is shown when available (named entities on delete, etc.). Errors surface
// the server message, falling back to "Couldn't <action>." and linger 4s so they
// can be read.
const label = (entity: string, name?: string) => (name ? `${entity} "${name}"` : entity)

export const notify = {
  created: (entity: string, name?: string) => toast.success(`${label(entity, name)} created successfully.`),
  updated: (entity: string, name?: string) => toast.success(`${label(entity, name)} updated successfully.`),
  deleted: (entity: string, name?: string) => toast.success(`${label(entity, name)} deleted successfully.`),
  saved: (entity: string, name?: string) => toast.success(`${label(entity, name)} saved successfully.`),
  success: (message: string) => toast.success(message),
  error: (err: unknown, action: string) =>
    toast.error((err as { message?: string })?.message || `Couldn't ${action}.`, { duration: 4000 }),
}
