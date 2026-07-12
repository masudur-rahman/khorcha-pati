import { formatDate } from '../lib/formatter'
import { useMemo, useState } from 'react'
import { notify } from '../lib/notify'
import { useQuery } from '@tanstack/react-query'
import { listWallets, listCategories, listSubcategories , getProfile } from '../api/endpoints'
import { useCreateWallet, useUpdateWallet, useDeleteWallet } from '../hooks/useWallets'
import { useTransactions } from '../hooks/useTransactions'
import { useSearch } from '../context/SearchContext'
import { fmt } from '../lib/formatter'
import type { Wallet } from '../types'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Modal from '../components/ui/Modal'
import Input from '../components/ui/Input'
import Select from '../components/ui/Select'
import SectionHeader from '../components/ui/SectionHeader'
import Eyebrow from '../components/ui/Eyebrow'
import Badge from '../components/ui/Badge'
import WalletCard, { WalletCardGhost, inferVariant } from '../components/ui/WalletCard'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import { ICONS } from '../components/ui/Icons'
import { validateWalletName, validateShortName } from '../utils/validators'

function variantOf(w: Wallet) {
  return inferVariant(w.type, w.name, w.shortName)
}

export default function Wallets() {
  const { searchTerm } = useSearch()
  const { data: wallets, isLoading } = useQuery({ queryKey: ['wallets'], queryFn: listWallets })
  const [showAddWallet, setShowAddWallet] = useState(false)
  const [showEditWallet, setShowEditWallet] = useState<Wallet | null>(null)
  const [showDeleteWallet, setShowDeleteWallet] = useState<Wallet | null>(null)
  const [activeWalletId, setActiveWalletId] = useState<number | null>(null)
  const del = useDeleteWallet()

  const filteredWallets = useMemo(() =>
    (wallets ?? []).filter(w =>
      !searchTerm ||
      w.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      w.shortName.toLowerCase().includes(searchTerm.toLowerCase())
    ),
    [wallets, searchTerm])

  const activeWallet = useMemo(
    () => (wallets ?? []).find(w => w.id === activeWalletId) ?? null,
    [wallets, activeWalletId]
  )

  const handleDeleteConfirm = (w: Wallet) => {
    del.mutate(w.shortName, {
      onSuccess: () => {
        notify.deleted('Wallet', w.name)
        setShowDeleteWallet(null)
      },
      onError: (err: any) => {
        notify.error(err, 'delete wallet')
        setShowDeleteWallet(null)
      }
    })
  }

  if (isLoading) return <p style={{ color: 'var(--color-text-tertiary)', padding: 40 }}>Loading...</p>

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 28 }}>
      <TopBar title="Wallets" subtitle="Manage your accounts and cash" />

      <section>
        <SectionHeader
          title="My Wallets"
          action={<Button onClick={() => setShowAddWallet(true)} icon={ICONS.addCircle(16)}>Add Wallet</Button>}
        />

        {filteredWallets.length === 0 ? (
          <Card style={{ padding: 48, textAlign: 'center', border: '2px dashed var(--color-border)' }}>
            <p style={{ color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
              {searchTerm ? 'No wallets match your search' : 'No wallets found. Add one to get started.'}
            </p>
          </Card>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 20 }}>
            {(() => {
              const counts: Record<string, number> = {}
              return filteredWallets.map(w => {
                const variant = variantOf(w)
                const idx = counts[variant] ?? 0
                counts[variant] = idx + 1
                return (
                  <WalletCard
                    key={w.id}
                    variant={variant}
                    paletteIndex={idx}
                    name={w.name}
                    shortName={w.shortName}
                    balance={w.balance}
                    onClick={() => setActiveWalletId(w.id)}
                  />
                )
              })
            })()}
            <WalletCardGhost onClick={() => setShowAddWallet(true)} />
          </div>
        )}
      </section>

      {showAddWallet && <AddWalletDialog onClose={() => setShowAddWallet(false)} />}
      {showEditWallet && (
        <EditWalletDialog 
          wallet={showEditWallet} 
          onClose={() => setShowEditWallet(null)} 
        />
      )}
      {showDeleteWallet && (
        <ConfirmDialog
          title="Delete Wallet"
          message={<>
            Delete <strong>"{showDeleteWallet.name}"</strong>?
            <br />
            <span style={{ fontSize: 13, opacity: 0.7 }}>This action cannot be undone.</span>
          </>}
          confirmText="Delete"
          type="danger"
          onConfirm={() => handleDeleteConfirm(showDeleteWallet)}
          onClose={() => setShowDeleteWallet(null)}
        />
      )}
      {activeWallet && (
        <WalletDrawer
          wallet={activeWallet}
          wallets={wallets ?? []}
          onEdit={(w) => {
            setShowEditWallet(w)
            setActiveWalletId(null)
          }}
          onDelete={(w) => {
            setShowDeleteWallet(w)
            setActiveWalletId(null)
          }}
          onClose={() => setActiveWalletId(null)}
        />
      )}
    </div>
  )
}

export function WalletDrawer({ wallet, wallets, onEdit, onDelete, onClose }: { wallet: Wallet; wallets: Wallet[]; onEdit: (w: Wallet) => void; onDelete: (w: Wallet) => void; onClose: () => void }) {
  const { data: profile } = useQuery({ queryKey: ['profile'], queryFn: getProfile })
  const { data: resp } = useTransactions()
  const { data: subcategories } = useQuery({ queryKey: ['subcategories'], queryFn: () => listSubcategories() })
  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: () => listCategories() })

  const subcatMap = useMemo(() => {
    const m = new Map<string, string>()
    subcategories?.forEach(s => m.set(s.id, s.name))
    return m
  }, [subcategories])

  const catNameMap = useMemo(() => {
    const m = new Map<string, string>()
    categories?.forEach(c => m.set(c.id, c.name))
    return m
  }, [categories])

  const subToCatMap = useMemo(() => {
    const m = new Map<string, string>()
    subcategories?.forEach(s => {
      const cName = catNameMap.get(s.catId) || s.catId
      m.set(s.id, cName)
    })
    return m
  }, [subcategories, catNameMap])

  // Compute the same paletteIndex that the list view uses
  const paletteIndex = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const w of wallets) {
      const v = variantOf(w)
      const idx = counts[v] ?? 0
      counts[v] = idx + 1
      if (w.id === wallet.id) return idx
    }
    return 0
  }, [wallets, wallet.id])

  const txns = (resp?.data ?? [])
    .filter(t => t.srcId === wallet.shortName || t.dstId === wallet.shortName)
    .sort((a, b) => b.timestamp - a.timestamp)
    .slice(0, 10)

  return (
    <Modal
      title={wallet.name}
      subtitle={wallet.shortName.toUpperCase()}
      onClose={onClose}
      width={520}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <WalletCard
          variant={variantOf(wallet)}
          paletteIndex={paletteIndex}
          name={wallet.name}
          shortName={wallet.shortName}
          balance={wallet.balance}
          onEdit={() => onEdit(wallet)}
          onDelete={() => onDelete(wallet)}
        />

        <div>
          <Eyebrow>Type</Eyebrow>
          <p style={{ fontSize: 14, fontWeight: 600, color: 'var(--color-text-primary)', margin: '6px 0 0' }}>{wallet.type}</p>
        </div>

        <div>
          <SectionHeader title="Recent Activity" />
          {txns.length === 0 ? (
            <p style={{ color: 'var(--color-text-tertiary)', fontSize: 13, fontWeight: 500 }}>No transactions yet.</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {txns.map(t => {
                const outgoing = t.srcId === wallet.shortName && t.type !== 'Income'
                const sign = outgoing ? '−' : '+'
                const color = outgoing ? 'var(--color-danger)' : 'var(--color-success)'
                return (
                  <div key={t.id} style={{
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                    padding: '12px 14px', background: 'var(--color-bg)',
                    borderRadius: 'var(--radius-md)', gap: 12,
                  }}>
                    <div style={{ display: 'flex', flexDirection: 'column', minWidth: 0, gap: 4 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <Badge type={t.type as any} />
                        <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--color-text-primary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {t.remarks || subcatMap.get(t.subcategoryId) || t.subcategoryId}
                        </span>
                      </div>
                      <div style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 500, marginLeft: 4, display: 'flex', gap: 6, alignItems: 'center' }}>
                        <span>{subToCatMap.get(t.subcategoryId) || 'Miscellaneous'} › {subcatMap.get(t.subcategoryId) || t.subcategoryId}</span>
                        <span>•</span>
                        <span>{formatDate(t.timestamp * 1000, { month: 'short', day: 'numeric', year: 'numeric' }, profile?.timezone)}</span>
                      </div>
                    </div>
                    <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 700, color, fontSize: 14, whiteSpace: 'nowrap' }}>
                      {sign}{fmt(t.amount)}
                    </span>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </Modal>
  )
}

function AddWalletDialog({ onClose }: { onClose: () => void }) {
  const create = useCreateWallet()
  const [type, setType] = useState('Bank')
  const [shortName, setShortName] = useState('')
  const [name, setName] = useState('')
  const [balance, setBalance] = useState('')

  const shortNameError = shortName ? validateShortName(shortName) : null
  const nameError = name ? validateWalletName(name) : null

  const handleSubmit = () => {
    if (shortNameError || nameError) return
    create.mutate({
      type: type as any,
      shortName,
      name,
      balance: balance === '' ? 0 : parseFloat(balance)
    }, {
      onSuccess: () => {
        notify.created('Wallet', name)
        onClose()
      },
      onError: (err: any) => {
        notify.error(err, 'create wallet')
      }
    })
  }

  return (
    <Modal
      title="Add New Wallet"
      onClose={onClose}
      width={460}
      onSubmit={() => { if (name && shortName && !shortNameError && !nameError && !create.isPending) handleSubmit() }}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={!name || !shortName || !!shortNameError || !!nameError || create.isPending}>
            Create Wallet
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Select label="Type" value={type} onChange={e => setType(e.target.value)} options={[{ value: 'Bank', label: 'Bank Account' }, { value: 'Cash', label: 'Cash / Other' }]} />
        <Input label="Short Name" placeholder="e.g. brac, cash" value={shortName} onChange={e => setShortName(e.target.value)} error={shortNameError || undefined} />
        <Input label="Display Name" placeholder="e.g. Personal Savings" value={name} onChange={e => setName(e.target.value)} error={nameError || undefined} />
        <Input label="Initial Balance" type="number" placeholder="0.00" value={balance} onChange={e => setBalance(e.target.value)} />
      </div>
    </Modal>
  )
}

export function EditWalletDialog({ wallet, onClose }: { wallet: Wallet; onClose: () => void }) {
  const update = useUpdateWallet()
  const [name, setName] = useState(wallet.name)
  const [shortName, setShortName] = useState(wallet.shortName)

  const shortNameError = shortName ? validateShortName(shortName) : null
  const nameError = name ? validateWalletName(name) : null

  const handleSubmit = () => {
    if (shortNameError || nameError) return
    update.mutate(
      { id: wallet.id, wallet: { name, shortName } },
      { 
        onSuccess: () => {
          notify.updated('Wallet', name)
          onClose()
        },
        onError: (err: any) => {
          notify.error(err, 'update wallet')
        }
      }
    )
  }

  return (
    <Modal
      title="Edit Wallet"
      onClose={onClose}
      width={460}
      onSubmit={() => { if (name && shortName && !shortNameError && !nameError && !update.isPending) handleSubmit() }}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={!name || !shortName || !!shortNameError || !!nameError || update.isPending}>
            Save Changes
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Input label="Short Name" placeholder="e.g. brac, cash" value={shortName} onChange={e => setShortName(e.target.value)} error={shortNameError || undefined} />
        <Input label="Display Name" placeholder="e.g. Personal Savings" value={name} onChange={e => setName(e.target.value)} error={nameError || undefined} />
      </div>
    </Modal>
  )
}
