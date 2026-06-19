import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listCategories, listSubcategories } from '../../api/endpoints'
import { useCreateTransaction, useUpdateTransaction } from '../../hooks/useTransactions'
import { useWallets } from '../../hooks/useWallets'
import { useContacts } from '../../hooks/useContacts'
import type { Transaction, TxnType } from '../../types'

import Modal from './Modal'
import Button from './Button'
import Input from './Input'
import Select from './Select'

export type { TxnType }
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

  const [type, setType] = useState<TxnType>(txn?.type as any ?? initialType ?? 'Expense')
  const [amount, setAmount] = useState(txn?.amount?.toString() ?? '')
  const [catId, setCatId] = useState('')
  const [subcategoryId, setSubcategoryId] = useState(txn?.subcategoryId ?? '')
  const [srcId, setSrcId] = useState(txn?.srcId ?? '')
  const [dstId, setDstId] = useState(txn?.dstId ?? initialContact ?? '')
  const [contactName, setContactName] = useState(txn?.contactName ?? initialContact ?? '')
  const [remarks, setRemarks] = useState(txn?.remarks ?? '')

  const { data: categories = [] } = useQuery({
    queryKey: ['categories', type],
    queryFn: () => listCategories(type),
  })
  const { data: subcategories = [] } = useQuery({
    queryKey: ['subcategories', catId, type],
    queryFn: () => listSubcategories(catId || undefined, type),
    enabled: !!catId,
  })

  // Edit-mode: restore catId by looking up the subcategory in the full list.
  const { data: allSubs = [] } = useQuery({
    queryKey: ['subcategories', 'all'],
    queryFn: () => listSubcategories(),
    enabled: isEdit && !catId && !!txn?.subcategoryId,
  })
  useEffect(() => {
    if (!isEdit || catId || !txn?.subcategoryId) return
    const cat = allSubs.find(s => s.id === txn.subcategoryId)?.catId
    if (cat) setCatId(cat)
  }, [allSubs, isEdit, txn?.subcategoryId, catId])

  // When type changes, reset cat/subcat if current selection no longer fits.
  useEffect(() => {
    if (isEdit) return
    if (catId && !categories.find(c => c.id === catId)) {
      setCatId('')
      setSubcategoryId('')
    }
  }, [categories, catId, isEdit])

  useEffect(() => {
    if (isEdit) return
    if (subcategoryId && !subcategories.find(s => s.id === subcategoryId)) {
      setSubcategoryId('')
    }
  }, [subcategories, subcategoryId, isEdit])

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
          <Select label="Sub Category" value={subcategoryId} onChange={e => setSubcategoryId(e.target.value)} options={subcategories.map(s => ({ value: s.id, label: s.name }))} />
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
