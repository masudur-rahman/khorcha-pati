import { useMemo, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getAdminStats, getAdminUsers, setAdminUserActive, type AdminUser } from '../api/endpoints'
import { getAccessToken } from '../api/client'
import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import MetricChip from '../components/ui/MetricChip'
import Modal from '../components/ui/Modal'
import Button from '../components/ui/Button'
import AICachePanel from '../components/admin/AICachePanel'

function decodeUserID(): number | null {
  const t = getAccessToken()
  if (!t) return null
  try {
    const payload = JSON.parse(atob(t.split('.')[1]))
    return typeof payload.uid === 'number' ? payload.uid : null
  } catch {
    return null
  }
}

function fullName(u: AdminUser): string {
  return [u.firstName, u.lastName].filter(Boolean).join(' ').trim()
}

// userLabel is used in compact contexts (confirmation modal, modal headers) — picks the
// strongest available identifier.
function userLabel(u: AdminUser): string {
  const name = fullName(u)
  if (name) return name
  if (u.username) return `@${u.username}`
  return `User #${u.id}`
}

// Field renders a single identity line, or a muted "not set" placeholder when value is missing,
// so admins can spot incomplete user records at a glance.
function Field({ value, missingLabel, style }: { value: string; missingLabel: string; style?: React.CSSProperties }) {
  if (value) {
    return <span style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', ...style }}>{value}</span>
  }
  return <span style={{ whiteSpace: 'nowrap', fontStyle: 'italic', opacity: 0.55, ...style }}>{missingLabel}</span>
}

function initials(u: AdminUser): string {
  const name = fullName(u)
  if (name) return name.split(/\s+/).map(p => p[0]).slice(0, 2).join('').toUpperCase()
  if (u.username) return u.username.slice(0, 2).toUpperCase()
  return String(u.id).slice(0, 2)
}

const avatarPalette: [string, string][] = [
  ['#6366f1', '#a78bfa'],
  ['#10b981', '#34d399'],
  ['#f59e0b', '#fbbf24'],
  ['#ef4444', '#fb7185'],
  ['#0ea5e9', '#38bdf8'],
  ['#14b8a6', '#5eead4'],
  ['#8b5cf6', '#c4b5fd'],
]

function avatarColors(id: number): [string, string] {
  return avatarPalette[id % avatarPalette.length]
}

function relativeTime(unix: number): string {
  if (!unix) return '—'
  const diff = Date.now() / 1000 - unix
  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  if (diff < 2592000) return `${Math.floor(diff / 86400)}d ago`
  if (diff < 31536000) return `${Math.floor(diff / 2592000)}mo ago`
  return `${Math.floor(diff / 31536000)}y ago`
}

function absoluteDate(unix: number): string {
  if (!unix) return '—'
  const d = new Date(unix * 1000)
  return d.toLocaleString(undefined, { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function Avatar({ user, size = 40, showStatus = true }: { user: AdminUser; size?: number; showStatus?: boolean }) {
  const [c1, c2] = avatarColors(user.id)
  return (
    <div style={{
      width: size, height: size, borderRadius: '50%',
      background: `linear-gradient(135deg, ${c1}, ${c2})`,
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      fontWeight: 700, fontSize: size * 0.4, color: '#fff', letterSpacing: '-0.02em',
      flexShrink: 0, position: 'relative',
    }}>
      {initials(user)}
      {showStatus && (
        <span style={{
          position: 'absolute', bottom: 0, right: 0,
          width: Math.max(8, size * 0.28), height: Math.max(8, size * 0.28),
          borderRadius: '50%',
          background: user.isActive ? '#10b981' : '#ef4444',
          boxShadow: '0 0 0 2px var(--color-surface)',
        }} />
      )}
    </div>
  )
}

export default function Admin() {
  const queryClient = useQueryClient()
  const currentUserID = useMemo(() => decodeUserID(), [])
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['adminStats'],
    queryFn: getAdminStats,
  })
  const { data: users, isLoading: usersLoading } = useQuery({
    queryKey: ['adminUsers'],
    queryFn: getAdminUsers,
  })
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null)
  const [confirmUser, setConfirmUser] = useState<AdminUser | null>(null)

  const toggleActive = useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) => setAdminUserActive(id, isActive),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminUsers'] })
      setConfirmUser(null)
    },
  })

  const isLoading = statsLoading || usersLoading

  if (isLoading) return <p style={{ color: 'var(--color-text-tertiary)', padding: 40 }}>Loading...</p>

  const statCards = [
    { label: 'Total Users', value: stats?.userCount ?? 0, color: 'var(--color-primary)' },
    { label: 'Total Transactions', value: stats?.txnCount ?? 0, color: 'var(--color-success)' },
    { label: 'Total Wallets', value: stats?.walletCount ?? 0, color: 'var(--color-danger)' },
    { label: 'Database', value: stats?.databaseType ?? 'unknown', color: 'var(--color-text-secondary)' },
  ]

  // refresh selectedUser data from query cache when it updates
  const liveSelected = selectedUser ? (users?.find(u => u.id === selectedUser.id) ?? selectedUser) : null

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Admin Dashboard" subtitle="System overview and user management" />

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: 16 }}>
        {statCards.map(card => (
          <MetricChip key={card.label} label={card.label} value={String(card.value)} accent={card.color} />
        ))}
      </div>

      <Card padding={0}>
        <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--color-border)' }}>
          <h3 style={{ fontSize: 16, fontWeight: 600, color: 'var(--color-text)' }}>Registered Users</h3>
        </div>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                {['User', 'Registered', 'Last txn', 'Wallets', 'Txns', 'Contacts', 'Role', 'Status'].map(h => (
                  <th key={h} style={{
                    padding: '12px 16px', textAlign: 'left', fontWeight: 600,
                    color: 'var(--color-text-secondary)', fontSize: 12,
                    textTransform: 'uppercase', letterSpacing: '0.05em',
                  }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {(users ?? []).map(u => (
                <tr
                  key={u.id}
                  onClick={() => setSelectedUser(selectedUser?.id === u.id ? null : u)}
                  style={{
                    borderBottom: '1px solid var(--color-border)',
                    cursor: 'pointer',
                    background: selectedUser?.id === u.id ? 'var(--color-primary-subtle)' : 'transparent',
                    transition: 'background var(--transition-fast)',
                  }}
                  onMouseEnter={e => {
                    if (selectedUser?.id !== u.id) e.currentTarget.style.background = 'var(--color-hover)'
                  }}
                  onMouseLeave={e => {
                    if (selectedUser?.id !== u.id) e.currentTarget.style.background = 'transparent'
                  }}
                >
                  <td style={{ padding: '12px 16px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12, minWidth: 0 }}>
                      <Avatar user={u} size={40} />
                      <div style={{ minWidth: 0, overflow: 'hidden', display: 'flex', flexDirection: 'column', gap: 1 }}>
                        <Field
                          value={fullName(u)}
                          missingLabel="Name not set"
                          style={{ fontSize: 14, fontWeight: 600, color: 'var(--color-text-primary)' }}
                        />
                        <Field
                          value={u.username ? `@${u.username}` : ''}
                          missingLabel="@ username not set"
                          style={{ fontSize: 12, color: 'var(--color-text-tertiary)' }}
                        />
                        <Field
                          value={u.telegramId ? `Telegram · ${u.telegramId}` : ''}
                          missingLabel="Telegram ID missing"
                          style={{ fontSize: 12, color: 'var(--color-text-tertiary)' }}
                        />
                      </div>
                    </div>
                  </td>
                  <td style={{ padding: '12px 16px', color: 'var(--color-text-tertiary)', fontSize: 13 }}>{relativeTime(u.createdAt)}</td>
                  <td style={{ padding: '12px 16px', color: 'var(--color-text-tertiary)', fontSize: 13 }}>{relativeTime(u.lastTxnAt)}</td>
                  <td style={{ padding: '12px 16px' }}>{u.walletCount}</td>
                  <td style={{ padding: '12px 16px' }}>{u.txnCount}</td>
                  <td style={{ padding: '12px 16px' }}>{u.contactCount}</td>
                  <td style={{ padding: '12px 16px' }}>
                    <span style={{
                      display: 'inline-block', padding: '2px 8px', borderRadius: 6, fontSize: 11, fontWeight: 600,
                      background: u.isAdmin ? 'var(--color-primary-subtle)' : 'var(--color-hover)',
                      color: u.isAdmin ? 'var(--color-primary)' : 'var(--color-text-tertiary)',
                    }}>
                      {u.isAdmin ? 'Admin' : 'User'}
                    </span>
                  </td>
                  <td style={{ padding: '12px 16px' }} onClick={e => e.stopPropagation()}>
                    <button
                      disabled={u.id === currentUserID || toggleActive.isPending}
                      onClick={() => setConfirmUser(u)}
                      style={{
                        padding: '4px 10px', borderRadius: 6, fontSize: 11, fontWeight: 600,
                        border: '1px solid var(--color-border)',
                        background: u.isActive ? 'var(--color-success-subtle, var(--color-hover))' : 'var(--color-danger-subtle, var(--color-hover))',
                        color: u.isActive ? 'var(--color-success, var(--color-text))' : 'var(--color-danger, var(--color-text))',
                        cursor: u.id === currentUserID ? 'not-allowed' : 'pointer',
                        opacity: u.id === currentUserID ? 0.5 : 1,
                      }}
                    >
                      {u.isActive ? 'Active' : 'Disabled'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {(!users || users.length === 0) && (
          <p style={{ padding: 24, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>No users found.</p>
        )}
      </Card>

      {liveSelected && <UserDetailCard
        user={liveSelected}
        isSelf={liveSelected.id === currentUserID}
        onClose={() => setSelectedUser(null)}
        onToggle={() => setConfirmUser(liveSelected)}
      />}

      <AICachePanel />

      {confirmUser && (
        <Modal title={confirmUser.isActive ? 'Disable user?' : 'Enable user?'} onClose={() => setConfirmUser(null)} width={460}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
            <Avatar user={confirmUser} size={44} />
            <div>
              <div style={{ fontWeight: 600, color: 'var(--color-text-primary)' }}>{userLabel(confirmUser)}</div>
              <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)' }}>Telegram · {confirmUser.telegramId}</div>
            </div>
          </div>
          <p style={{ margin: 0, color: 'var(--color-text-secondary)', lineHeight: 1.6 }}>
            {confirmUser.isActive
              ? 'Web login and bot interaction will be blocked until re-enabled.'
              : 'Web login and bot interaction will be restored.'}
          </p>
          <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 24 }}>
            <Button variant="secondary" onClick={() => setConfirmUser(null)} disabled={toggleActive.isPending}>
              Cancel
            </Button>
            <Button
              onClick={() => toggleActive.mutate({ id: confirmUser.id, isActive: !confirmUser.isActive })}
              disabled={toggleActive.isPending}
            >
              {toggleActive.isPending ? 'Working…' : confirmUser.isActive ? 'Disable' : 'Enable'}
            </Button>
          </div>
        </Modal>
      )}
    </div>
  )
}

function UserDetailCard({ user, isSelf, onClose, onToggle }: { user: AdminUser; isSelf: boolean; onClose: () => void; onToggle: () => void }) {
  return (
    <Card padding={0}>
      <div style={{
        padding: '24px 28px', borderBottom: '1px solid var(--color-border)',
        display: 'flex', alignItems: 'center', gap: 20, flexWrap: 'wrap',
      }}>
        <Avatar user={user} size={64} showStatus={false} />
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap', minWidth: 0, marginBottom: 6 }}>
            <Field
              value={fullName(user)}
              missingLabel="Name not set"
              style={{
                fontSize: 22, fontWeight: 700, color: 'var(--color-text-primary)',
                letterSpacing: '-0.02em',
              }}
            />
            <Pill kind={user.isActive ? 'success' : 'danger'}>{user.isActive ? 'Active' : 'Disabled'}</Pill>
            {user.isAdmin && <Pill kind="primary">Admin</Pill>}
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
            <Field
              value={user.username ? `@${user.username}` : ''}
              missingLabel="@ username not set"
              style={{ fontSize: 13, color: 'var(--color-text-secondary)', fontWeight: 500 }}
            />
            <span style={{ color: 'var(--color-text-tertiary)', opacity: 0.6 }}>·</span>
            <Field
              value={user.telegramId ? `Telegram · ${user.telegramId}` : ''}
              missingLabel="Telegram ID missing"
              style={{ fontSize: 13, color: 'var(--color-text-tertiary)' }}
            />
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <Button
            variant="secondary"
            disabled={isSelf}
            onClick={onToggle}
            title={isSelf ? "Can't change your own status" : undefined}
          >
            {user.isActive ? 'Disable' : 'Enable'}
          </Button>
          <button
            onClick={onClose}
            style={{
              padding: '10px 12px', borderRadius: 12,
              border: '1px solid var(--color-border)', background: 'transparent',
              cursor: 'pointer', color: 'var(--color-text-tertiary)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              transition: 'all 0.15s',
            }}
            onMouseEnter={e => e.currentTarget.style.transform = 'translateY(-1px)'}
            onMouseLeave={e => e.currentTarget.style.transform = 'translateY(0)'}
            aria-label="Close"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      </div>

      <Section title="Activity">
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16 }}>
          <Tile
            icon={<TileIcon path="M3 7l2-2 7 4 7-4 2 2v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z" />}
            label="Wallets"
            value={String(user.walletCount)}
            accent="#0ea5e9"
          />
          <Tile
            icon={<TileIcon path="M3 3v18h18 M7 14l4-4 4 4 5-7" />}
            label="Transactions"
            value={String(user.txnCount)}
            accent="#10b981"
          />
          <Tile
            icon={<TileIcon path="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2 M9 7a4 4 0 1 1 0 8 4 4 0 0 1 0-8z M23 21v-2a4 4 0 0 0-3-3.87 M16 3.13a4 4 0 0 1 0 7.75" />}
            label="Contacts"
            value={String(user.contactCount)}
            accent="#f59e0b"
          />
        </div>
      </Section>

      <Section title="Timeline">
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
          <DetailRow
            label="Registered"
            primary={absoluteDate(user.createdAt)}
            secondary={user.createdAt ? relativeTime(user.createdAt) : undefined}
          />
          <DetailRow
            label="Last transaction"
            primary={absoluteDate(user.lastTxnAt)}
            secondary={user.lastTxnAt ? relativeTime(user.lastTxnAt) : undefined}
          />
        </div>
      </Section>

      <Section title="Account">
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 16 }}>
          <DetailRow label="User ID" primary={String(user.id)} />
          <DetailRow label="Username" primary={user.username ? `@${user.username}` : '—'} />
          <DetailRow label="Full name" primary={fullName(user) || '—'} />
          <DetailRow label="Telegram ID" primary={String(user.telegramId)} />
        </div>
      </Section>
    </Card>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div style={{ padding: '20px 28px', borderBottom: '1px solid var(--color-border)' }}>
      <h4 style={{
        fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)',
        textTransform: 'uppercase', letterSpacing: '0.08em', margin: '0 0 14px',
      }}>
        {title}
      </h4>
      {children}
    </div>
  )
}

function Tile({ icon, label, value, accent }: { icon: React.ReactNode; label: string; value: string; accent: string }) {
  return (
    <div style={{
      padding: 16, borderRadius: 12, border: '1px solid var(--color-border)',
      background: 'var(--color-surface-elevated, var(--color-surface))',
      display: 'flex', alignItems: 'center', gap: 14,
    }}>
      <div style={{
        width: 40, height: 40, borderRadius: 10,
        background: `${accent}1a`, color: accent,
        display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
      }}>
        {icon}
      </div>
      <div>
        <div style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em' }}>{label}</div>
        <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--color-text-primary)', letterSpacing: '-0.02em', lineHeight: 1.1 }}>{value}</div>
      </div>
    </div>
  )
}

function TileIcon({ path }: { path: string }) {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d={path} />
    </svg>
  )
}

function DetailRow({ label, primary, secondary }: { label: string; primary: string; secondary?: string }) {
  return (
    <div>
      <div style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 4 }}>
        {label}
      </div>
      <div style={{ fontSize: 14, fontWeight: 500, color: 'var(--color-text-primary)' }}>{primary}</div>
      {secondary && (
        <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', marginTop: 2 }}>{secondary}</div>
      )}
    </div>
  )
}

function Pill({ kind, children }: { kind: 'success' | 'danger' | 'primary'; children: React.ReactNode }) {
  const palette = {
    success: { bg: '#10b9811a', fg: '#10b981' },
    danger: { bg: '#ef44441a', fg: '#ef4444' },
    primary: { bg: 'var(--color-primary-subtle)', fg: 'var(--color-primary)' },
  }[kind]
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 6,
      padding: '4px 10px', borderRadius: 999, fontSize: 11, fontWeight: 700,
      background: palette.bg, color: palette.fg,
      textTransform: 'uppercase', letterSpacing: '0.06em',
    }}>
      <span style={{ width: 6, height: 6, borderRadius: '50%', background: palette.fg }} />
      {children}
    </span>
  )
}
