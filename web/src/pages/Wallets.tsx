import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listWallets } from '../api/endpoints'
import { useCreateWallet } from '../hooks/useWallets'
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
import DrawerPanel from '../components/ui/DrawerPanel'
import { ICONS } from '../components/ui/Icons'

function variantOf(w: Wallet) {
  return inferVariant(w.type, w.name, w.shortName)
}

export default function Wallets() {
  const { searchTerm } = useSearch()
  const { data: wallets, isLoading } = useQuery({ queryKey: ['wallets'], queryFn: listWallets })
  const [showAddWallet, setShowAddWallet] = useState(false)
  const [activeWalletId, setActiveWalletId] = useState<number | null>(null)

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
      {activeWallet && (
        <WalletDrawer wallet={activeWallet} onClose={() => setActiveWalletId(null)} />
      )}
    </div>
  )
}

function WalletDrawer({ wallet, onClose }: { wallet: Wallet; onClose: () => void }) {
  const { data: resp } = useTransactions()
  const txns = (resp?.data ?? [])
    .filter(t => t.srcId === wallet.shortName || t.dstId === wallet.shortName)
    .sort((a, b) => b.timestamp - a.timestamp)
    .slice(0, 10)

  return (
    <DrawerPanel
      title={wallet.name}
      subtitle={wallet.shortName.toUpperCase()}
      onClose={onClose}
      width={520}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <WalletCard
          variant={variantOf(wallet)}
          name={wallet.name}
          shortName={wallet.shortName}
          balance={wallet.balance}
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
                      <Badge type={t.type as any} />
                      <span style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 600 }}>
                        {new Date(t.timestamp * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                      </span>
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
    </DrawerPanel>
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
