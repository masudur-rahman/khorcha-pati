import { formatDate } from '../lib/formatter'
import { useEffect, useMemo, useState } from 'react'
import toast from 'react-hot-toast'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { listContacts , getProfile } from '../api/endpoints'
import { useCreateContact, useUpdateContact, useDeleteContact } from '../hooks/useContacts'
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
import DrawerPanel from '../components/ui/DrawerPanel'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import ActionButton from '../components/ui/ActionButton'
import TxnDialog, { TxnType } from '../components/ui/TxnDialog'
import { ICONS } from '../components/ui/Icons'
import { validateDisplayName, validateShortName } from '../utils/validators'

export default function Contacts() {
  const { searchTerm } = useSearch()
  const { data: contacts, isLoading } = useQuery({ queryKey: ['contacts'], queryFn: listContacts })
  const [showAddContact, setShowAddContact] = useState(false)
  const [showEditContact, setShowEditContact] = useState<Contact | null>(null)
  const [showDeleteContact, setShowDeleteContact] = useState<Contact | null>(null)
  const [activeId, setActiveId] = useState<number | null>(null)
  const [searchParams] = useSearchParams()
  const showParam = searchParams.get('show')
  const del = useDeleteContact()

  useEffect(() => {
    if (showParam) {
      const id = parseInt(showParam, 10)
      if (!isNaN(id)) {
        setActiveId(id)
      }
    }
  }, [showParam])

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

  const handleDeleteConfirm = (c: Contact) => {
    del.mutate(c.id, {
      onSuccess: () => {
        toast.success('Contact deleted successfully')
        setShowDeleteContact(null)
      },
      onError: (err: any) => {
        toast.error(
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            <strong style={{ fontSize: 14 }}>Failed to Delete Contact</strong>
            <span style={{ fontSize: 13, opacity: 0.9 }}>{err.message || 'An unknown error occurred.'}</span>
          </div>,
          { duration: 4000 }
        )
        setShowDeleteContact(null)
      }
    })
  }

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
          icon={ICONS.trendingUp(18)}
        />
        <CirclePanel
          label="You owe others"
          amount={oweOthers}
          accent="var(--color-danger)"
          icon={ICONS.trendingDown(18)}
        />
        <Card style={{ display: 'flex', flexDirection: 'column', justifyContent: 'space-between', gap: 12, padding: 18 }}>
          <Eyebrow>Reminders</Eyebrow>
          {/* Reminder API not wired yet — disabled for now. */}
          <Button disabled onClick={() => { }} style={{ width: '100%' }}>Send Reminders</Button>
        </Card>
      </div>

      {/* Contacts list */}
      <section>
        <SectionHeader
          title="Contacts"
          action={<Button onClick={() => setShowAddContact(true)} icon={ICONS.personAdd(16)}>Add Contact</Button>}
        />

        {filtered.length === 0 ? (
          <Card style={{ padding: 48, textAlign: 'center', border: '2px dashed var(--color-border)' }}>
            <p style={{ color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
              {searchTerm ? 'No contacts match your search' : 'No contacts yet. Add one to get started.'}
            </p>
          </Card>
        ) : (
          <Card padding={0}>
            <div className="contact-list">
              {filtered.map(c => (
                <ContactRow
                  key={c.id}
                  contact={c}
                  onClick={() => setActiveId(c.id)}
                />
              ))}
            </div>
          </Card>
        )}
      </section>

      {showAddContact && <AddContactDialog onClose={() => setShowAddContact(false)} />}
      {showEditContact && (
        <EditContactDialog 
          contact={showEditContact} 
          onClose={() => setShowEditContact(null)} 
        />
      )}
      {showDeleteContact && (
        <ConfirmDialog
          title="Delete Contact"
          message={<>
            Delete <strong>"{showDeleteContact.fullName || showDeleteContact.nickName}"</strong>?
            <br />
            <span style={{ fontSize: 13, opacity: 0.7 }}>This action cannot be undone.</span>
          </>}
          confirmText="Delete"
          type="danger"
          onConfirm={() => handleDeleteConfirm(showDeleteContact)}
          onClose={() => setShowDeleteContact(null)}
        />
      )}
      {active && (
        <ContactDrawer
          contact={active}
          onEdit={(c) => {
            setShowEditContact(c)
            setActiveId(null)
          }}
          onDelete={(c) => {
            setShowDeleteContact(c)
            setActiveId(null)
          }}
          onClose={() => setActiveId(null)}
        />
      )}
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
  const { data: profile } = useQuery({ queryKey: ['profile'], queryFn: getProfile })
  const owesYou = contact.netBalance > 0
  const settled = contact.netBalance === 0
  const color = settled ? 'var(--color-text-tertiary)' : owesYou ? 'var(--color-success)' : 'var(--color-danger)'
  return (
    <button
      onClick={onClick}
      className="hover-row"
      style={{
        display: 'flex', alignItems: 'center', gap: 14,
        padding: '14px 20px',
        background: 'transparent',
        border: 'none',
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
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 4, flexShrink: 0 }}>
        <span style={{
          padding: '2px 8px', borderRadius: 999, fontSize: 9, fontWeight: 800,
          letterSpacing: '0.06em', textTransform: 'uppercase',
          background: settled ? 'var(--color-bg)' : owesYou ? 'var(--color-success-subtle)' : 'var(--color-danger-subtle)',
          color,
          whiteSpace: 'nowrap',
        }}>
          {settled ? 'Settled' : owesYou ? 'Owes you' : 'You owe'}
        </span>
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700, color }}>
          {settled ? fmt(0) : `${owesYou ? '+' : '−'}${fmt(Math.abs(contact.netBalance))}`}
        </span>
        <span style={{ fontSize: 10, color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
          {contact.lastTxnTimestamp
            ? formatDate(contact.lastTxnTimestamp * 1000, { month: 'short', day: 'numeric' }, profile?.timezone)
            : 'No activity'}
        </span>
      </div>
    </button>
  )
}

function ContactDrawer({ contact, onEdit, onDelete, onClose }: { contact: Contact; onEdit: (c: Contact) => void; onDelete: (c: Contact) => void; onClose: () => void }) {
  const { data: profile } = useQuery({ queryKey: ['profile'], queryFn: getProfile })
  const [add, setAdd] = useState<{ type: TxnType; sub: string } | null>(null)
  const { data: resp } = useTransactions()

  const nick = contact.nickName.toLowerCase()
  const txns = (resp?.data ?? [])
    .filter(t =>
      (t.contactName ?? '').toLowerCase() === nick ||
      (t.srcId ?? '').toLowerCase() === nick ||
      (t.dstId ?? '').toLowerCase() === nick
    )
    .sort((a, b) => b.timestamp - a.timestamp)
    .slice(0, 20)

  const owesYou = contact.netBalance > 0
  const settled = contact.netBalance === 0
  const color = settled ? 'var(--color-text-tertiary)' : owesYou ? 'var(--color-success)' : 'var(--color-danger)'

  const goAdd = (type: 'Expense' | 'Income') => {
    const sub = type === 'Expense'
      ? (contact.netBalance < 0 ? 'fin-return' : 'fin-lend')
      : (contact.netBalance > 0 ? 'fin-recover' : 'fin-borrow')
    setAdd({ type, sub })
  }

  return (
    <DrawerPanel
      title={contact.fullName || contact.nickName}
      subtitle={contact.nickName.toUpperCase()}
      onClose={onClose}
      width={480}
      headerActions={
        <>
          <ActionButton
            actionType="edit"
            icon={ICONS.edit(15)}
            onClick={() => onEdit(contact)}
            title="Edit Contact"
          />
          <ActionButton
            actionType="delete"
            icon={ICONS.trash(15)}
            onClick={() => onDelete(contact)}
            title="Delete Contact"
          />
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <Card style={{ display: 'flex', flexDirection: 'column', gap: 8, padding: 18, borderLeft: `4px solid ${color}`, borderRadius: 'var(--radius-md)' }}>
          <Eyebrow>{settled ? 'Settled' : owesYou ? 'Owes you' : 'You owe'}</Eyebrow>
          <span style={{ fontSize: 28, fontWeight: 700, color, fontFamily: 'var(--font-display)', letterSpacing: '-0.02em' }}>
            {fmt(Math.abs(contact.netBalance))}
          </span>
        </Card>

        {/* Action buttons */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 8 }}>
          <ActionBtn icon={ICONS.trendingDown(18)} label="Pay them" accent="var(--color-danger)" onClick={() => goAdd('Expense')} />
          <ActionBtn icon={ICONS.trendingUp(18)} label="Got from them" accent="var(--color-success)" onClick={() => goAdd('Income')} />
          <ActionBtn icon={ICONS.bell(18)} label="Remind" accent="var(--color-primary)" disabled onClick={() => { }} />
        </div>

        {contact.email && (
          <div>
            <Eyebrow>Email</Eyebrow>
            <p style={{ fontSize: 14, fontWeight: 500, color: 'var(--color-text-primary)', margin: '6px 0 0' }}>{contact.email}</p>
          </div>
        )}

        <div>
          <SectionHeader title="History with them" />
          {txns.length === 0 ? (
            <p style={{ color: 'var(--color-text-tertiary)', fontSize: 13, fontWeight: 500 }}>No transactions with this contact yet.</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {txns.map(t => {
                // Reframe from the contact relationship POV:
                //   Income to user = "Received from them" (money came in)
                //   Expense to user = "Paid to them"     (money went out)
                //   Transfer = literal transfer
                const isReceived = t.type === 'Income'
                const isPaid = t.type === 'Expense'
                const label = isReceived ? 'Received from them' : isPaid ? 'Paid to them' : 'Transfer'
                const sign = isReceived ? '+' : isPaid ? '−' : ''
                const txnColor = isReceived ? 'var(--color-success)' : isPaid ? 'var(--color-danger)' : 'var(--color-primary)'
                const subtleBg = isReceived ? 'var(--color-success-subtle)' : isPaid ? 'var(--color-danger-subtle)' : 'var(--color-primary-subtle)'
                return (
                  <div key={t.id} style={{
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                    padding: '12px 14px', background: 'var(--color-bg)',
                    borderRadius: 'var(--radius-md)', gap: 12,
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12, minWidth: 0 }}>
                      <div style={{
                        width: 32, height: 32, borderRadius: '50%',
                        background: subtleBg, color: txnColor,
                        display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
                      }}>
                        {isReceived ? ICONS.trendingUp(16) : isPaid ? ICONS.trendingDown(16) : ICONS.swapHoriz(16)}
                      </div>
                      <div style={{ display: 'flex', flexDirection: 'column', minWidth: 0, gap: 2 }}>
                        <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--color-text-primary)' }}>{label}</span>
                        <span style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
                          {formatDate(t.timestamp * 1000, { month: 'short', day: 'numeric', year: 'numeric' }, profile?.timezone)}
                          {t.remarks ? ` · ${t.remarks}` : ''}
                        </span>
                      </div>
                    </div>
                    <span style={{
                      fontFamily: 'var(--font-mono)', fontWeight: 700, fontSize: 14, whiteSpace: 'nowrap',
                      color: txnColor,
                    }}>
                      {sign}{fmt(t.amount)}
                    </span>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
      {add && (
        <TxnDialog
          initialType={add.type}
          initialSubcategory={add.sub}
          initialContact={contact.nickName}
          onClose={() => setAdd(null)}
        />
      )}
    </DrawerPanel>
  )
}

function ActionBtn({ icon, label, accent, onClick, disabled }: { icon: React.ReactNode; label: string; accent: string; onClick: () => void; disabled?: boolean }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      style={{
        display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6,
        padding: '12px 8px', borderRadius: 'var(--radius-md)',
        background: 'var(--color-surface)', border: `1px solid var(--color-border)`,
        color: accent, cursor: disabled ? 'not-allowed' : 'pointer', fontFamily: 'inherit',
        fontSize: 11, fontWeight: 700, transition: 'all var(--transition-fast)',
        textTransform: 'uppercase', letterSpacing: '0.04em',
        opacity: disabled ? 0.45 : 1,
      }}
      onMouseEnter={e => { if (disabled) return; e.currentTarget.style.borderColor = accent; e.currentTarget.style.background = `color-mix(in srgb, ${accent} 8%, var(--color-surface))` }}
      onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--color-border)'; e.currentTarget.style.background = 'var(--color-surface)' }}
    >
      <span style={{ display: 'flex' }}>{icon}</span>
      {label}
    </button>
  )
}

function AddContactDialog({ onClose }: { onClose: () => void }) {
  const create = useCreateContact()
  const [nickName, setNickName] = useState('')
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')

  const nickNameError = nickName ? validateShortName(nickName) : null
  const fullNameError = fullName ? validateDisplayName(fullName) : null

  const handleSubmit = () => {
    if (nickNameError || fullNameError) return
    create.mutate({ nickName, fullName, email }, { 
      onSuccess: () => {
        toast.success('Contact created successfully')
        onClose()
      },
      onError: (err: any) => {
        toast.error(err.message || 'Failed to create contact.')
      }
    })
  }

  return (
    <Modal title="Add New Contact" onClose={onClose} width={460}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Input label="Nick Name (Unique)" placeholder="e.g. karim" value={nickName} onChange={e => setNickName(e.target.value)} error={nickNameError || undefined} />
        <Input label="Full Name" placeholder="e.g. Abdul Karim" value={fullName} onChange={e => setFullName(e.target.value)} error={fullNameError || undefined} />
        <Input label="Email Address" type="email" placeholder="karim@example.com" value={email} onChange={e => setEmail(e.target.value)} />
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 12 }}>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button
            onClick={handleSubmit}
            disabled={!nickName || !!nickNameError || !!fullNameError || create.isPending}
            style={{ padding: '12px 32px' }}
          >
            Create Contact
          </Button>
        </div>
      </div>
    </Modal>
  )
}

function EditContactDialog({ contact, onClose }: { contact: Contact; onClose: () => void }) {
  const update = useUpdateContact()
  const [nickName, setNickName] = useState(contact.nickName)
  const [fullName, setFullName] = useState(contact.fullName || '')
  const [email, setEmail] = useState(contact.email || '')

  const nickNameError = nickName ? validateShortName(nickName) : null
  const fullNameError = fullName ? validateDisplayName(fullName) : null

  const handleSubmit = () => {
    if (nickNameError || fullNameError) return
    update.mutate(
      { id: contact.id, contact: { nickName, fullName, email } },
      { 
        onSuccess: () => {
          toast.success('Contact updated successfully')
          onClose()
        },
        onError: (err: any) => {
          toast.error(err.message || 'Failed to update contact.')
        }
      }
    )
  }

  return (
    <Modal title="Edit Contact" onClose={onClose} width={460}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Input label="Nick Name (Unique)" placeholder="e.g. karim" value={nickName} onChange={e => setNickName(e.target.value)} error={nickNameError || undefined} />
        <Input label="Full Name" placeholder="e.g. Abdul Karim" value={fullName} onChange={e => setFullName(e.target.value)} error={fullNameError || undefined} />
        <Input label="Email Address" type="email" placeholder="karim@example.com" value={email} onChange={e => setEmail(e.target.value)} />
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 12 }}>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={!nickName || !!nickNameError || !!fullNameError || update.isPending} style={{ padding: '12px 32px' }}>
            Save Changes
          </Button>
        </div>
      </div>
    </Modal>
  )
}
