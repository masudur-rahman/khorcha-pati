import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listCategories, listSubcategories } from '../../api/endpoints'
import { useCreateTransaction, useUpdateTransaction } from '../../hooks/useTransactions'
import { useWallets } from '../../hooks/useWallets'
import { useContacts } from '../../hooks/useContacts'
import type { Transaction } from '../../types'

import Modal from './Modal'
import Button from './Button'
import Input from './Input'
import Select from './Select'

export type TxnType = 'Expense' | 'Income' | 'Transfer'
export const TXN_TYPE_OPTIONS: TxnType[] = ['Expense', 'Income', 'Transfer']

interface Props {
  txn?: Transaction
  initialType?: TxnType
  initialContact?: string
  onClose: () => void
}

export default function TxnDialog({ txn, initialType, initialContact, onClose }: Props) {
  const create = useCreateTransaction()
  const update = useUpdateTransaction()
  const isEdit = !!txn

  const { data: wallets = [] } = useWallets()
  const { data: contacts = [] } = useContacts()
  const { data: categories = [] } = useQuery({ queryKey: ['categories'], queryFn: listCategories })
  const { data: subcategories = [] } = useQuery({ queryKey: ['subcategories'], queryFn: () => listSubcategories() })

  const [type, setType] = useState<TxnType>(txn?.type as any ?? initialType ?? 'Expense')
  const [amount, setAmount] = useState(txn?.amount?.toString() ?? '')
  const [catId, setCatId] = useState(() => {
    if (txn?.subcategoryId) return subcategories.find(s => s.id === txn.subcategoryId)?.catId || ''
    return ''
  })
  const [subcategoryId, setSubcategoryId] = useState(txn?.subcategoryId ?? '')
  const [srcId, setSrcId] = useState(txn?.srcId ?? '')
  const [dstId, setDstId] = useState(txn?.dstId ?? initialContact ?? '')
  const [contactName, setContactName] = useState(txn?.contactName ?? initialContact ?? '')
  const [remarks, setRemarks] = useState(txn?.remarks ?? '')

  // Restore catId once subcategories load (handles edit-mode race).
  useEffect(() => {
    if (txn?.subcategoryId && !catId) {
      const cat = subcategories.find(s => s.id === txn.subcategoryId)?.catId
      if (cat) setCatId(cat)
    }
  }, [subcategories, txn?.subcategoryId, catId])

  useEffect(() => {
    if (isEdit) return
    if (type === 'Transfer') {
      const finCat = categories.find(c => c.name.toLowerCase() === 'financial')
      if (finCat) {
        setCatId(finCat.id)
        const transSub = subcategories.find(s => s.catId === finCat.id && s.name.toLowerCase().includes('transfer'))
        if (transSub) setSubcategoryId(transSub.id)
      }
    } else {
      const finCat = categories.find(c => c.name.toLowerCase() === 'financial')
      if (catId === finCat?.id) { setCatId(''); setSubcategoryId('') }
    }
  }, [type, categories, subcategories, isEdit])

  const filteredSubs = useMemo(() => {
    let subs = subcategories
    if (catId) subs = subs.filter(s => s.catId === catId)
    if (type === 'Transfer') subs = subs.filter(s => s.name.toLowerCase().includes('transfer') || s.name.toLowerCase().includes('withdraw') || s.name.toLowerCase().includes('deposit'))
    return subs
  }, [subcategories, catId, type])

  useEffect(() => {
    if (isEdit) return
    const sub = subcategories.find(s => s.id === subcategoryId)
    if (sub?.name.toLowerCase() === 'withdraw') {
      const cashWallet = wallets.find(w => w.type === 'Cash' || w.shortName.toLowerCase() === 'cash')
      if (cashWallet) setDstId(cashWallet.shortName)
    }
  }, [subcategoryId, subcategories, wallets, isEdit])

  const handleSubmit = () => {
    const payload: Partial<Transaction> = { type, amount: parseFloat(amount), subcategoryId, srcId, dstId, contactName, remarks }
    if (isEdit) update.mutate({ id: txn!.id, ...payload }, { onSuccess: onClose })
    else create.mutate(payload, { onSuccess: onClose })
  }

  return (
    <Modal title={isEdit ? 'Edit Transaction' : 'Add Transaction'} onClose={onClose}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
          <Select label="Type" value={type} onChange={e => setType(e.target.value as TxnType)} options={TXN_TYPE_OPTIONS.map(t => ({ value: t, label: t }))} />
          <Input label="Amount" type="number" value={amount} onChange={e => setAmount(e.target.value)} placeholder="0.00" />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
          <Select label="Category" value={catId} onChange={e => { setCatId(e.target.value); setSubcategoryId('') }} options={categories.map(c => ({ value: c.id, label: c.name }))} />
          <Select label="Sub Category" value={subcategoryId} onChange={e => setSubcategoryId(e.target.value)} options={filteredSubs.map(s => ({ value: s.id, label: s.name }))} />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
          <Select label="From (Wallet)" value={srcId} onChange={e => setSrcId(e.target.value)} options={wallets.map(w => ({ value: w.shortName, label: w.name }))} />
          <Select label={type === 'Transfer' ? 'To (Wallet)' : 'To'} value={dstId} onChange={e => setDstId(e.target.value)}
            options={type === 'Transfer' ? wallets.map(w => ({ value: w.shortName, label: w.name })) : [{ value: '', label: 'Select Contact' }, ...contacts.map(c => ({ value: c.nickName, label: c.fullName || c.nickName }))]} />
        </div>
        {type !== 'Transfer' && <Input label="Contact (Optional)" value={contactName} onChange={e => setContactName(e.target.value)} placeholder="Who is this with?" />}
        <Input label="Remarks" value={remarks} onChange={e => setRemarks(e.target.value)} placeholder="Any notes..." />
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 12 }}>
          <Button variant="secondary" onClick={onClose} style={{ padding: '12px 24px' }}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={!amount || !subcategoryId} style={{ padding: '12px 32px' }}>
            {isEdit ? 'Update Changes' : 'Create Transaction'}
          </Button>
        </div>
      </div>
    </Modal>
  )
}
