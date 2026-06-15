import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listWallets, listContacts } from '../api/endpoints'
import { useCreateWallet } from '../hooks/useWallets'
import { useCreateContact } from '../hooks/useContacts'
import { useSearch } from '../context/SearchContext'
import { fmt } from '../lib/formatter'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Modal from '../components/ui/Modal'
import Input from '../components/ui/Input'
import Select from '../components/ui/Select'
import { ICONS } from '../components/ui/Icons'

export default function Wallets() {
  const { searchTerm } = useSearch()
  const { data: wallets, isLoading: wLoading } = useQuery({ queryKey: ['wallets'], queryFn: listWallets })
  const { data: contacts, isLoading: cLoading } = useQuery({ queryKey: ['contacts'], queryFn: listContacts })
  const [showAddWallet, setShowAddWallet] = useState(false)
  const [showAddContact, setShowAddContact] = useState(false)

  const filteredWallets = useMemo(() =>
    (wallets ?? []).filter(w => !searchTerm || w.name.toLowerCase().includes(searchTerm.toLowerCase()) || w.shortName.toLowerCase().includes(searchTerm.toLowerCase())),
    [wallets, searchTerm])

  const filteredContacts = useMemo(() =>
    (contacts ?? []).filter(c => !searchTerm || c.nickName.toLowerCase().includes(searchTerm.toLowerCase()) || (c.fullName && c.fullName.toLowerCase().includes(searchTerm.toLowerCase()))),
    [contacts, searchTerm])

  if (wLoading || cLoading) return <p style={{ color: 'var(--color-text-tertiary)', padding: 40 }}>Loading...</p>

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Wallets" subtitle="Manage your bank accounts and contacts" />

      {/* Wallets Section */}
      <section>
        <header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 }}>
          <h2 style={{ fontSize: 18, fontWeight: 700, color: 'var(--color-text-primary)' }}>My Wallets</h2>
          <Button onClick={() => setShowAddWallet(true)} icon={ICONS.plus(16)}>Add Wallet</Button>
        </header>

        {filteredWallets.length === 0 ? (
          <Card style={{ padding: 48, textAlign: 'center', border: '2px dashed var(--color-border)' }}>
            <p style={{ color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
              {searchTerm ? 'No wallets match your search' : 'No wallets found. Add one to get started.'}
            </p>
          </Card>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: 20 }}>
            {filteredWallets.map(w => {
              const accent = w.type === 'Bank' ? 'var(--color-primary)' : 'var(--color-success)'
              return (
                <Card key={w.id} padding={0} style={{ overflow: 'hidden', position: 'relative', transition: 'all 0.2s ease', display: 'flex', flexDirection: 'column' }}
                  onMouseEnter={(e: React.MouseEvent<HTMLDivElement>) => { e.currentTarget.style.transform = 'translateY(-4px)'; e.currentTarget.style.boxShadow = '0 12px 32px rgba(0,0,0,0.06)' }}
                  onMouseLeave={(e: React.MouseEvent<HTMLDivElement>) => { e.currentTarget.style.transform = 'translateY(0)'; e.currentTarget.style.boxShadow = 'none' }}>
                  {/* Accent top bar */}
                  <div style={{ height: 4, background: accent }} />

                  <div style={{ padding: '24px 28px', flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 }}>
                      <div style={{
                        width: 44, height: 44, borderRadius: 'var(--radius-md)',
                        background: accent + '12', color: accent,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                      }}>
                        {w.type === 'Bank' ? ICONS.creditCard(22) : ICONS.banknote(22)}
                      </div>
                      <span style={{
                        fontSize: 10, fontWeight: 700, color: 'var(--color-text-tertiary)',
                        background: 'var(--color-bg)', padding: '4px 12px', borderRadius: 'var(--radius-sm)',
                        textTransform: 'uppercase', letterSpacing: '0.08em',
                      }}>{w.type}</span>
                    </div>

                    <h3 style={{ fontSize: 17, fontWeight: 700, color: 'var(--color-text-primary)', marginBottom: 2 }}>{w.name}</h3>
                    <p style={{ fontSize: 11, fontWeight: 600, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 20 }}>{w.shortName}</p>

                    <div style={{ fontSize: 28, fontWeight: 800, color: 'var(--color-text-primary)', letterSpacing: '-0.02em', display: 'flex', alignItems: 'baseline', gap: 4, fontFamily: "var(--font-display)" }}>
                      <span style={{ fontSize: 18, fontWeight: 600, color: 'var(--color-text-tertiary)' }}>৳</span>
                      {Math.abs(w.balance).toLocaleString('en-US', { minimumFractionDigits: 2 })}
                      {w.balance < 0 && <span style={{ fontSize: 12, color: 'var(--color-danger)', marginLeft: 4, fontWeight: 700 }}>(DR)</span>}
                    </div>
                  </div>
                </Card>
              )
            })}
          </div>
        )}
      </section>

      {/* Contacts Section */}
      <section>
        <header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8, gap: 12, flexWrap: 'wrap' }}>
          <div style={{ minWidth: 0 }}>
            <h2 style={{ fontSize: 18, fontWeight: 700, color: 'var(--color-text-primary)', margin: 0 }}>Frequent Contacts</h2>
            <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', margin: '4px 0 0', fontWeight: 500 }}>
              <span style={{ color: 'var(--color-success)', fontWeight: 700 }}>+</span> they owe you
              <span style={{ margin: '0 8px', color: 'var(--color-border)' }}>·</span>
              <span style={{ color: 'var(--color-danger)', fontWeight: 700 }}>−</span> you owe them
            </p>
          </div>
          <Button onClick={() => setShowAddContact(true)} icon={ICONS.plus(16)}>Add Contact</Button>
        </header>

        {filteredContacts.length === 0 ? (
          <Card style={{ padding: 48, textAlign: 'center', border: '2px dashed var(--color-border)' }}>
            <p style={{ color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
              {searchTerm ? 'No contacts match your search' : 'No contacts found.'}
            </p>
          </Card>
        ) : (
          <Card padding={0}>
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                    {['Contact', 'Email', 'Net Balance', 'Last Transaction'].map(h => (
                      <th
                        key={h}
                        title={h === 'Net Balance' ? 'Positive: they owe you · Negative: you owe them' : undefined}
                        style={{ padding: '14px 24px', textAlign: 'left', fontSize: 10, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em' }}
                      >
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {filteredContacts.map(c => (
                    <tr key={c.id} style={{ borderBottom: '1px solid var(--color-border)' }} className="hover-row transition-colors">
                      <td style={{ padding: '16px 24px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                          <div style={{ width: 36, height: 36, background: 'var(--color-primary-subtle)', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--color-primary)', fontWeight: 700, fontSize: 12 }}>
                            {c.nickName.slice(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <div style={{ fontWeight: 700, color: 'var(--color-text-primary)' }}>{c.fullName || c.nickName}</div>
                            <div style={{ fontSize: 10, color: 'var(--color-text-tertiary)', fontWeight: 600, textTransform: 'uppercase' }}>{c.nickName}</div>
                          </div>
                        </div>
                      </td>
                      <td style={{ padding: '16px 24px', color: 'var(--color-text-secondary)', fontWeight: 500 }}>{c.email || '—'}</td>
                      <td style={{ padding: '16px 24px', fontWeight: 700, fontSize: 14, color: c.netBalance >= 0 ? 'var(--color-success)' : 'var(--color-danger)' }}>
                        {c.netBalance >= 0 ? '+' : ''}{fmt(c.netBalance)}
                      </td>
                      <td style={{ padding: '16px 24px', color: 'var(--color-text-tertiary)', fontSize: 12, fontWeight: 600 }}>
                        {c.lastTxnTimestamp ? new Date(c.lastTxnTimestamp * 1000).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) : 'Never'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>
        )}
      </section>

      {showAddWallet && <AddWalletDialog onClose={() => setShowAddWallet(false)} />}
      {showAddContact && <AddContactDialog onClose={() => setShowAddContact(false)} />}
    </div>
  )
}

function AddWalletDialog({ onClose }: { onClose: () => void }) {
  const create = useCreateWallet()
  const [type, setType] = useState('Bank')
  const [shortName, setShortName] = useState('')
  const [name, setName] = useState('')
  const [balance, setBalance] = useState('')

  return (
    <Modal title="Add New Wallet" onClose={onClose} width={460}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Select label="Type" value={type} onChange={e => setType(e.target.value)} options={[{ value: 'Bank', label: 'Bank Account' }, { value: 'Cash', label: 'Cash / Other' }]} />
        <Input label="Short Name" placeholder="e.g. brac, cash" value={shortName} onChange={e => setShortName(e.target.value)} />
        <Input label="Display Name" placeholder="e.g. Personal Savings" value={name} onChange={e => setName(e.target.value)} />
        <Input label="Initial Balance" type="number" placeholder="0.00" value={balance} onChange={e => setBalance(e.target.value)} />
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 12 }}>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => create.mutate({ type: type as any, shortName, name, balance: parseFloat(balance) }, { onSuccess: onClose })} disabled={!name || !shortName || !balance} style={{ padding: '12px 32px' }}>Create Wallet</Button>
        </div>
      </div>
    </Modal>
  )
}

function AddContactDialog({ onClose }: { onClose: () => void }) {
  const create = useCreateContact()
  const [nickName, setNickName] = useState('')
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')

  return (
    <Modal title="Add New Contact" onClose={onClose} width={460}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Input label="Nick Name (Unique)" placeholder="e.g. karim" value={nickName} onChange={e => setNickName(e.target.value)} />
        <Input label="Full Name" placeholder="e.g. Abdul Karim" value={fullName} onChange={e => setFullName(e.target.value)} />
        <Input label="Email Address" type="email" placeholder="karim@example.com" value={email} onChange={e => setEmail(e.target.value)} />
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 12 }}>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => create.mutate({ nickName, fullName, email }, { onSuccess: onClose })} disabled={!nickName} style={{ padding: '12px 32px' }}>Create Contact</Button>
        </div>
      </div>
    </Modal>
  )
}
