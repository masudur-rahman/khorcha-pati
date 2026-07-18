import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  getAccessSettings, updateAccessSettings, addAllowedUser, removeAllowedUser, restoreAllowedUser,
  type AccessSettings, type AllowedUser, type AdminUser,
} from '../../api/endpoints'
import { notify } from '../../lib/notify'
import Card from '../ui/Card'
import Button from '../ui/Button'
import Input from '../ui/Input'

function relativeDate(unix: number): string {
  if (!unix) return '—'
  return new Date(unix * 1000).toLocaleDateString()
}

function entryFor(settings: AccessSettings | undefined, user: AdminUser): AllowedUser | undefined {
  return settings?.allowedUsers.find(e =>
    (e.telegramId !== 0 && e.telegramId === user.telegramId) ||
    (e.username !== '' && user.username !== '' && e.username.toLowerCase() === user.username.toLowerCase()),
  )
}

/* Single shared query (includes revoked tombstones) — panel and all row
   buttons cost one fetch. */
const accessQuery = { queryKey: ['accessSettings'], queryFn: () => getAccessSettings(true) }

/* Allow/Revoke pill for the registered-users table. */
export function AllowButton({ user }: { user: AdminUser }) {
  const queryClient = useQueryClient()
  const { data: settings } = useQuery(accessQuery)
  const entry = entryFor(settings, user)
  const label = user.username ? `@${user.username}` : String(user.telegramId)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['accessSettings'] })
  const allow = useMutation({
    mutationFn: () =>
      entry ? restoreAllowedUser(entry.id) : addAllowedUser({ username: user.username, telegramId: user.telegramId }),
    onSuccess: () => { invalidate(); notify.success(`${label} allowed.`) },
    onError: err => notify.error(err, 'allow user'),
  })
  const revoke = useMutation({
    mutationFn: (id: number) => removeAllowedUser(id),
    onSuccess: () => { invalidate(); notify.success(`${label} revoked.`) },
    onError: err => notify.error(err, 'revoke user'),
  })

  if (!settings) return null
  const active = entry && !entry.revoked
  const pending = allow.isPending || revoke.isPending
  // Flag pending users prominently only when the gate is actually on.
  const highlight = settings.restricted && !active

  return (
    <button
      disabled={pending}
      onClick={() => (active ? revoke.mutate(entry.id) : allow.mutate())}
      title={active ? 'Click to revoke access' : 'Click to allow access'}
      style={{
        padding: '4px 10px', borderRadius: 6, fontSize: 11, fontWeight: 600,
        border: '1px solid var(--color-border)',
        background: active
          ? 'var(--color-success-subtle, var(--color-hover))'
          : highlight ? 'var(--color-warning-subtle, var(--color-hover))' : 'transparent',
        color: active
          ? 'var(--color-success, var(--color-text))'
          : highlight ? 'var(--color-warning, var(--color-text))' : 'var(--color-text-secondary)',
        cursor: 'pointer',
        marginLeft: 8,
      }}
    >
      {active ? 'Allowed' : 'Allow'}
    </button>
  )
}

export default function AccessControlPanel() {
  const queryClient = useQueryClient()
  const { data: settings, isLoading } = useQuery(accessQuery)

  const [replyText, setReplyText] = useState('')
  const [newEntry, setNewEntry] = useState('')
  const [showRevoked, setShowRevoked] = useState(false)

  useEffect(() => {
    if (settings) setReplyText(settings.replyText)
  }, [settings?.replyText])

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['accessSettings'] })
  const update = useMutation({
    mutationFn: updateAccessSettings,
    onSuccess: () => { invalidate(); notify.updated('Access settings') },
    onError: err => notify.error(err, 'update access settings'),
  })
  const add = useMutation({
    mutationFn: () => {
      const raw = newEntry.trim().replace(/^@/, '')
      return /^\d+$/.test(raw) ? addAllowedUser({ telegramId: Number(raw) }) : addAllowedUser({ username: raw })
    },
    onSuccess: () => { setNewEntry(''); invalidate(); notify.success('Allowlist entry added.') },
    onError: err => notify.error(err, 'add allowed user'),
  })
  const revoke = useMutation({
    mutationFn: (id: number) => removeAllowedUser(id),
    onSuccess: invalidate,
    onError: err => notify.error(err, 'revoke allowed user'),
  })
  const restore = useMutation({
    mutationFn: (id: number) => restoreAllowedUser(id),
    onSuccess: invalidate,
    onError: err => notify.error(err, 'restore allowed user'),
  })

  if (isLoading || !settings) return null

  const activeEntries = settings.allowedUsers.filter(e => !e.revoked)
  const visibleEntries = showRevoked ? settings.allowedUsers : activeEntries
  const revokedCount = settings.allowedUsers.length - activeEntries.length

  return (
    <Card padding={0} style={{ width: '100%', overflow: 'hidden' }}>
      <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
        <div style={{ flex: 1, minWidth: 180 }}>
          <h3 style={{ fontSize: 16, fontWeight: 600, color: 'var(--color-text)' }}>Access Control</h3>
          <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', marginTop: 2 }}>
            Restrict this instance to allowed users. Everyone else can only /start and gets the redirect reply. The bot owner is always allowed.
          </p>
        </div>
        <label style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer', fontSize: 13, fontWeight: 600, color: settings.restricted ? 'var(--color-danger)' : 'var(--color-text-secondary)' }}>
          <input
            type="checkbox"
            checked={settings.restricted}
            disabled={update.isPending}
            onChange={e => update.mutate({ restricted: e.target.checked })}
            style={{ width: 16, height: 16, cursor: 'pointer' }}
          />
          Allowed users only
        </label>
      </div>

      <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 20 }}>
        {/* Redirect reply text */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <span style={{ fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--color-text-tertiary)' }}>
            Redirect reply for non-allowed users
          </span>
          <textarea
            value={replyText}
            onChange={e => setReplyText(e.target.value)}
            rows={3}
            style={{
              width: '100%', background: 'var(--color-bg)', border: '1px solid var(--color-border)',
              borderRadius: 12, padding: '10px 14px', fontSize: 13, color: 'var(--color-text-primary)',
              outline: 'none', resize: 'vertical', fontFamily: 'inherit',
            }}
          />
          <div>
            <Button
              onClick={() => update.mutate({ replyText })}
              disabled={update.isPending || replyText === settings.replyText}
              style={{ padding: '8px 16px', borderRadius: 10 }}
            >
              Save Text
            </Button>
          </div>
        </div>

        {/* Allowlist */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span style={{ fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--color-text-tertiary)' }}>
              Allowed users ({activeEntries.length})
            </span>
            {revokedCount > 0 && (
              <button
                onClick={() => setShowRevoked(v => !v)}
                style={{
                  border: 'none', background: 'none', padding: 0, fontSize: 11, fontWeight: 600,
                  color: 'var(--color-primary)', cursor: 'pointer',
                }}
              >
                {showRevoked ? 'Hide revoked' : `Show all (+${revokedCount} revoked)`}
              </button>
            )}
          </div>
          {visibleEntries.length === 0 ? (
            <p style={{ fontSize: 13, color: 'var(--color-text-tertiary)', margin: 0 }}>
              No entries yet. Add one below or use “Allow” in the users table.
            </p>
          ) : (
            <div style={{ border: '1px solid var(--color-border)', borderRadius: 12, overflow: 'hidden' }}>
              {visibleEntries.map(e => (
                <div key={e.id} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '8px 14px', borderBottom: '1px solid var(--color-border)', fontSize: 13, opacity: e.revoked ? 0.55 : 1 }}>
                  <span style={{ fontWeight: 600, color: 'var(--color-text-primary)', minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {e.username ? `@${e.username}` : '—'}
                  </span>
                  <span style={{ color: 'var(--color-text-tertiary)', fontFamily: 'var(--font-mono, monospace)', fontSize: 12 }}>
                    {e.telegramId !== 0 ? e.telegramId : 'id pending'}
                  </span>
                  {e.revoked && (
                    <span title={`Revoked ${relativeDate(e.revokedAt)}`} style={{ padding: '1px 8px', borderRadius: 6, fontSize: 10, fontWeight: 700, background: 'var(--color-danger-subtle, var(--color-hover))', color: 'var(--color-danger)' }}>
                      Revoked
                    </span>
                  )}
                  <span style={{ color: 'var(--color-text-tertiary)', fontSize: 12, marginLeft: 'auto' }}>
                    {e.revoked ? `revoked ${relativeDate(e.revokedAt)}` : relativeDate(e.createdAt)}
                  </span>
                  <button
                    onClick={() => (e.revoked ? restore.mutate(e.id) : revoke.mutate(e.id))}
                    disabled={revoke.isPending || restore.isPending}
                    style={{
                      padding: '3px 10px', borderRadius: 6, fontSize: 11, fontWeight: 600,
                      border: '1px solid var(--color-border)', background: 'transparent',
                      color: e.revoked ? 'var(--color-success)' : 'var(--color-danger)', cursor: 'pointer',
                    }}
                  >
                    {e.revoked ? 'Restore' : 'Revoke'}
                  </button>
                </div>
              ))}
            </div>
          )}
          <div style={{ display: 'flex', gap: 10, alignItems: 'flex-end', maxWidth: 420 }}>
            <Input
              label="Add by username or Telegram ID"
              value={newEntry}
              placeholder="e.g. karim or 123456789"
              onChange={e => setNewEntry(e.target.value)}
              onKeyDown={e => { if (e.key === 'Enter' && newEntry.trim() && !add.isPending) add.mutate() }}
            />
            <Button onClick={() => add.mutate()} disabled={!newEntry.trim() || add.isPending} style={{ padding: '11px 18px', borderRadius: 10 }}>
              Add
            </Button>
          </div>
        </div>
      </div>
    </Card>
  )
}
