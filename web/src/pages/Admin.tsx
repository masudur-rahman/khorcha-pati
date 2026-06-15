import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getAdminStats, getAdminUsers, type AdminUser } from '../api/endpoints'
import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'

export default function Admin() {
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['adminStats'],
    queryFn: getAdminStats,
  })
  const { data: users, isLoading: usersLoading } = useQuery({
    queryKey: ['adminUsers'],
    queryFn: getAdminUsers,
  })
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null)

  const isLoading = statsLoading || usersLoading

  if (isLoading) return <p style={{ color: 'var(--color-text-tertiary)', padding: 40 }}>Loading...</p>

  const statCards = [
    { label: 'Total Users', value: stats?.userCount ?? 0, color: 'var(--color-primary)' },
    { label: 'Total Transactions', value: stats?.txnCount ?? 0, color: 'var(--color-success)' },
    { label: 'Total Wallets', value: stats?.walletCount ?? 0, color: 'var(--color-danger)' },
    { label: 'Database', value: stats?.databaseType ?? 'unknown', color: 'var(--color-text-secondary)' },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Admin Dashboard" subtitle="System overview and user management" />

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 20 }}>
        {statCards.map(card => (
          <Card key={card.label} style={{ borderLeft: `4px solid ${card.color}` }}>
            <p style={{ fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.06em', margin: '0 0 8px' }}>{card.label}</p>
            <p style={{ fontSize: 26, fontWeight: 700, color: card.color, margin: 0, fontFamily: "var(--font-display)" }}>{card.value}</p>
          </Card>
        ))}
      </div>

      {/* Users Table */}
      <Card padding={0}>
          <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--color-border)' }}>
            <h3 style={{ fontSize: 16, fontWeight: 600, color: 'var(--color-text)' }}>Registered Users</h3>
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                  {['Username', 'Name', 'Wallets', 'Transactions', 'Contacts', 'Role'].map(h => (
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
                    <td style={{ padding: '12px 16px', fontWeight: 500 }}>@{u.username}</td>
                    <td style={{ padding: '12px 16px', color: 'var(--color-text-secondary)' }}>
                      {[u.firstName, u.lastName].filter(Boolean).join(' ') || '-'}
                    </td>
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
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {(!users || users.length === 0) && (
            <p style={{ padding: 24, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>No users found.</p>
          )}
        </Card>

      {/* User Detail Panel */}
      {selectedUser && (
        <Card padding={20}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
              <h3 style={{ fontSize: 16, fontWeight: 600 }}>User Detail</h3>
              <button
                onClick={() => setSelectedUser(null)}
                style={{
                  background: 'none', border: 'none', cursor: 'pointer',
                  color: 'var(--color-text-tertiary)', fontSize: 18,
                }}
              >
                &times;
              </button>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 12, fontSize: 14 }}>
              <Detail label="Username" value={`@${selectedUser.username}`} />
              <Detail label="Name" value={[selectedUser.firstName, selectedUser.lastName].filter(Boolean).join(' ') || '-'} />
              <Detail label="Telegram ID" value={String(selectedUser.telegramId)} />
              <Detail label="Wallets" value={String(selectedUser.walletCount)} />
              <Detail label="Transactions" value={String(selectedUser.txnCount)} />
              <Detail label="Contacts" value={String(selectedUser.contactCount)} />
              <Detail label="Role" value={selectedUser.isAdmin ? 'Admin' : 'User'} />
            </div>
        </Card>
      )}
    </div>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p style={{ fontSize: 11, color: 'var(--color-text-tertiary)', marginBottom: 2, textTransform: 'uppercase', letterSpacing: '0.05em' }}>{label}</p>
      <p style={{ fontWeight: 500, color: 'var(--color-text)' }}>{value}</p>
    </div>
  )
}
