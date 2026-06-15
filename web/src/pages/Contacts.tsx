import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listContacts } from '../api/endpoints'
import { useCreateContact } from '../hooks/useContacts'
import { useTransactions } from '../hooks/useTransactions'
import { useSearch } from '../context/SearchContext'
import { fmt } from '../lib/formatter'
import type { Contact } from '../types'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Modal from '../components/ui/Modal'
import Input from '../components/ui/Input'
import Eyebrow from '../components/ui/Eyebrow'
import SectionHeader from '../components/ui/SectionHeader'
import Badge from '../components/ui/Badge'
import DrawerPanel from '../components/ui/DrawerPanel'
import { ICONS } from '../components/ui/Icons'

export default function Contacts() {
  const { searchTerm } = useSearch()
  const { data: contacts, isLoading } = useQuery({ queryKey: ['contacts'], queryFn: listContacts })
  const [showAddContact, setShowAddContact] = useState(false)
  const [activeId, setActiveId] = useState<number | null>(null)

  const filtered = useMemo(() =>
    (contacts ?? []).filter(c =>
      !searchTerm ||
      c.nickName.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (c.fullName && c.fullName.toLowerCase().includes(searchTerm.toLowerCase()))
    ),
    [contacts, searchTerm])

  const { owedToYou, oweOthers } = useMemo(() => {
    const list = contacts ?? []
    return {
      owedToYou: list.filter(c => c.netBalance > 0).reduce((s, c) => s + c.netBalance, 0),
      oweOthers: list.filter(c => c.netBalance < 0).reduce((s, c) => s + Math.abs(c.netBalance), 0),
    }
  }, [contacts])

  const active = useMemo(() => (contacts ?? []).find(c => c.id === activeId) ?? null, [contacts, activeId])

  if (isLoading) return <p style={{ color: 'var(--color-text-tertiary)', padding: 40 }}>Loading...</p>

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Contacts" subtitle="People you transact with" />

      {/* Financial Circle widget */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16 }}>
        <CirclePanel
          label="Others owe you"
          amount={owedToYou}
          accent="var(--color-success)"
          icon={ICONS.arrowDown(18)}
        />
        <CirclePanel
          label="You owe others"
          amount={oweOthers}
          accent="var(--color-danger)"
          icon={ICONS.arrowUp(18)}
        />
        <Card style={{ display: 'flex', flexDirection: 'column', justifyContent: 'space-between', gap: 12, padding: 18 }}>
          <Eyebrow>Reminders</Eyebrow>
          <Button onClick={() => {/* TODO: wire reminder API */ }} style={{ width: '100%' }}>Send Reminders</Button>
        </Card>
      </div>

      {/* Contacts list */}
      <section>
        <SectionHeader
          title="Contacts"
          action={<Button onClick={() => setShowAddContact(true)} icon={ICONS.plus(16)}>Add Contact</Button>}
        />

        {filtered.length === 0 ? (
          <Card style={{ padding: 48, textAlign: 'center', border: '2px dashed var(--color-border)' }}>
            <p style={{ color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
              {searchTerm ? 'No contacts match your search' : 'No contacts yet. Add one to get started.'}
            </p>
          </Card>
        ) : (
          <Card padding={0}>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {filtered.map(c => (
                <ContactRow key={c.id} contact={c} onClick={() => setActiveId(c.id)} />
              ))}
            </div>
          </Card>
        )}
      </section>

      {showAddContact && <AddContactDialog onClose={() => setShowAddContact(false)} />}
      {active && <ContactDrawer contact={active} onClose={() => setActiveId(null)} />}
    </div>
  )
}

function CirclePanel({ label, amount, accent, icon }: { label: string; amount: number; accent: string; icon: React.ReactNode }) {
  return (
    <Card style={{ display: 'flex', flexDirection: 'column', gap: 8, padding: 18, borderLeft: `4px solid ${accent}`, borderRadius: 'var(--radius-md)' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Eyebrow>{label}</Eyebrow>
        <span style={{ color: accent, display: 'flex' }}>{icon}</span>
      </div>
      <span style={{ fontSize: 24, fontWeight: 700, color: accent, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>
        {fmt(amount)}
      </span>
    </Card>
  )
}

function ContactRow({ contact, onClick }: { contact: Contact; onClick: () => void }) {
  const owesYou = contact.netBalance > 0
  const settled = contact.netBalance === 0
  const color = settled ? 'var(--color-text-tertiary)' : owesYou ? 'var(--color-success)' : 'var(--color-danger)'
  return (
    <button
      onClick={onClick}
      className="hover-row"
      style={{
        display: 'flex', alignItems: 'center', gap: 14,
        padding: '14px 20px', borderBottom: '1px solid var(--color-border)',
        background: 'transparent', border: 'none', borderBottomColor: 'var(--color-border)',
        cursor: 'pointer', fontFamily: 'inherit', textAlign: 'left', width: '100%',
      }}
    >
      <div style={{
        width: 42, height: 42, borderRadius: '50%',
        background: 'var(--color-primary-subtle)', color: 'var(--color-primary)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        fontWeight: 700, fontSize: 13, flexShrink: 0,
      }}>
        {contact.nickName.slice(0, 2).toUpperCase()}
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--color-text-primary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {contact.fullName || contact.nickName}
        </div>
        <div style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
          {contact.nickName}{contact.email ? ` · ${contact.email}` : ''}
        </div>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 2, flexShrink: 0 }}>
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700, color }}>
          {settled ? fmt(0) : `${owesYou ? '+' : '−'}${fmt(Math.abs(contact.netBalance))}`}
        </span>
        <span style={{ fontSize: 10, color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
          {contact.lastTxnTimestamp
            ? new Date(contact.lastTxnTimestamp * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
            : 'No activity'}
        </span>
      </div>
    </button>
  )
}

function ContactDrawer({ contact, onClose }: { contact: Contact; onClose: () => void }) {
  const { data: resp } = useTransactions()
  const txns = (resp?.data ?? [])
    .filter(t => t.contactName === contact.nickName)
    .sort((a, b) => b.timestamp - a.timestamp)
    .slice(0, 20)

  const owesYou = contact.netBalance > 0
  const settled = contact.netBalance === 0
  const color = settled ? 'var(--color-text-tertiary)' : owesYou ? 'var(--color-success)' : 'var(--color-danger)'

  return (
    <DrawerPanel
      title={contact.fullName || contact.nickName}
      subtitle={contact.nickName.toUpperCase()}
      onClose={onClose}
      width={480}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <Card style={{ display: 'flex', flexDirection: 'column', gap: 8, padding: 18, borderLeft: `4px solid ${color}`, borderRadius: 'var(--radius-md)' }}>
          <Eyebrow>{settled ? 'Settled' : owesYou ? 'Owes you' : 'You owe'}</Eyebrow>
          <span style={{ fontSize: 28, fontWeight: 700, color, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>
            {fmt(Math.abs(contact.netBalance))}
          </span>
        </Card>

        {contact.email && (
          <div>
            <Eyebrow>Email</Eyebrow>
            <p style={{ fontSize: 14, fontWeight: 500, color: 'var(--color-text-primary)', margin: '6px 0 0' }}>{contact.email}</p>
          </div>
        )}

        <div>
          <SectionHeader title="Recent Activity" />
          {txns.length === 0 ? (
            <p style={{ color: 'var(--color-text-tertiary)', fontSize: 13, fontWeight: 500 }}>No transactions with this contact yet.</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {txns.map(t => (
                <div key={t.id} style={{
                  display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                  padding: '12px 14px', background: 'var(--color-bg)',
                  borderRadius: 'var(--radius-md)', gap: 12,
                }}>
                  <div style={{ display: 'flex', flexDirection: 'column', minWidth: 0, gap: 4 }}>
                    <Badge type={t.type as any} />
                    <span style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
                      {new Date(t.timestamp * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                    </span>
                  </div>
                  <span style={{
                    fontFamily: 'var(--font-mono)', fontWeight: 700, fontSize: 14, whiteSpace: 'nowrap',
                    color: t.type === 'Income' ? 'var(--color-success)' : t.type === 'Transfer' ? 'var(--color-primary)' : 'var(--color-danger)',
                  }}>
                    {t.type === 'Income' ? '+' : t.type === 'Transfer' ? '' : '−'}{fmt(t.amount)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </DrawerPanel>
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
          <Button
            onClick={() => create.mutate({ nickName, fullName, email }, { onSuccess: onClose })}
            disabled={!nickName}
            style={{ padding: '12px 32px' }}
          >
            Create Contact
          </Button>
        </div>
      </div>
    </Modal>
  )
}
