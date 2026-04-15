import { useState, useMemo } from 'react'
import { useTransactions, useCreateTransaction, useUpdateTransaction, useDeleteTransaction } from '../hooks/useTransactions'
import { useWallets } from '../hooks/useWallets'
import { useContacts } from '../hooks/useContacts'
import { useQuery } from '@tanstack/react-query'
import { listCategories, listSubcategories } from '../api/endpoints'
import type { Transaction, Wallet, Contact, TxnCategory } from '../types'

type TxnType = 'Expense' | 'Income' | 'Transfer'
const typeOptions: TxnType[] = ['Expense', 'Income', 'Transfer']

export default function Transactions() {
  const { data: resp, isLoading } = useTransactions()
  const txns = resp?.data ?? []
  const { data: wallets } = useWallets()
  const { data: contacts } = useContacts()
  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: listCategories })
  const { data: subcategories } = useQuery({ queryKey: ['subcategories'], queryFn: () => listSubcategories() })

  const [filterType, setFilterType] = useState<string>('')
  const [showAdd, setShowAdd] = useState(false)
  const [editTxn, setEditTxn] = useState<Transaction | null>(null)
  const [deleteTxn, setDeleteTxn] = useState<Transaction | null>(null)

  const subcatMap = useMemo(() => {
    const m = new Map<string, string>()
    subcategories?.forEach(s => m.set(s.id, s.name))
    return m
  }, [subcategories])

  if (isLoading) return <p className="text-gray-500">Loading...</p>

  const filtered = txns.filter(t => !filterType || t.type === filterType)

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h1 className="text-2xl font-bold">Transactions</h1>
        <div className="flex gap-2">
          <button onClick={() => setShowAdd(true)} className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700">+ Add Txn</button>
        </div>
      </div>

      <div className="flex gap-2">
        <FilterButton label="All" active={filterType === ''} onClick={() => setFilterType('')} />
        {typeOptions.map(t => (
          <FilterButton key={t} label={t} active={filterType === t} onClick={() => setFilterType(t)} />
        ))}
      </div>

      <div className="bg-white rounded-lg shadow overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-gray-500 border-b">
              <th className="p-3">Type</th>
              <th className="p-3">Amount</th>
              <th className="p-3">Sub Category</th>
              <th className="p-3">From</th>
              <th className="p-3">To</th>
              <th className="p-3">Date</th>
              <th className="p-3">Remarks</th>
              <th className="p-3">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr><td colSpan={8} className="p-4 text-center text-gray-400">No transactions</td></tr>
            ) : filtered.map(t => (
              <tr key={t.id} className="border-b last:border-0 hover:bg-gray-50">
                <td className="p-3">{t.type}</td>
                <td className={`p-3 font-medium ${t.type === 'Income' ? 'text-green-600' : 'text-red-600'}`}>
                  {t.amount.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                </td>
                <td className="p-3">{subcatMap.get(t.subcategoryId) || t.subcategoryId}</td>
                <td className="p-3">{t.srcId}</td>
                <td className="p-3">{t.dstId}</td>
                <td className="p-3 text-gray-500">{new Date(t.timestamp * 1000).toLocaleDateString()}</td>
                <td className="p-3 text-gray-500 truncate max-w-32">{t.remarks}</td>
                <td className="p-3 flex gap-1">
                  <button className="text-blue-600 hover:underline text-xs" onClick={() => setEditTxn(t)}>Edit</button>
                  <button className="text-red-600 hover:underline text-xs" onClick={() => setDeleteTxn(t)}>Del</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showAdd && (
        <TxnDialog
          wallets={wallets ?? []}
          contacts={contacts ?? []}
          categories={categories ?? []}
          subcategories={subcategories ?? []}
          onClose={() => setShowAdd(false)}
        />
      )}
      {editTxn && (
        <TxnDialog
          txn={editTxn}
          wallets={wallets ?? []}
          contacts={contacts ?? []}
          categories={categories ?? []}
          subcategories={subcategories ?? []}
          onClose={() => setEditTxn(null)}
        />
      )}
      {deleteTxn && <DeleteDialog txn={deleteTxn} onClose={() => setDeleteTxn(null)} />}
    </div>
  )
}

function FilterButton({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      className={`px-3 py-1 rounded text-sm ${active ? 'bg-blue-600 text-white' : 'bg-gray-100 hover:bg-gray-200'}`}
      onClick={onClick}
    >
      {label}
    </button>
  )
}

interface TxnDialogProps {
  txn?: Transaction
  wallets: Wallet[]
  contacts: Contact[]
  categories: TxnCategory[]
  subcategories: { id: string; name: string; catId: string }[]
  onClose: () => void
}

function TxnDialog({ txn, wallets, contacts, categories, subcategories, onClose }: TxnDialogProps) {
  const create = useCreateTransaction()
  const update = useUpdateTransaction()
  const isEdit = !!txn

  const [type, setType] = useState<TxnType>(txn?.type ?? 'Expense')
  const [amount, setAmount] = useState(txn?.amount?.toString() ?? '')
  const [catId, setCatId] = useState(() => {
    if (txn?.subcategoryId) {
      return subcategories.find(s => s.id === txn.subcategoryId)?.catId || ''
    }
    return ''
  })
  const [subcategoryId, setSubcategoryId] = useState(txn?.subcategoryId ?? '')
  const [srcId, setSrcId] = useState(txn?.srcId ?? '')
  const [dstId, setDstId] = useState(txn?.dstId ?? '')
  const [contactName, setContactName] = useState(txn?.contactName ?? '')
  const [remarks, setRemarks] = useState(txn?.remarks ?? '')

  const filteredSubs = useMemo(() => {
    return subcategories.filter(s => !catId || s.catId === catId)
  }, [subcategories, catId])

  const handleSubmit = () => {
    const payload: Partial<Transaction> = {
      type,
      amount: parseFloat(amount),
      subcategoryId,
      srcId,
      dstId,
      contactName,
      remarks,
    }
    if (isEdit) {
      update.mutate({ id: txn.id, ...payload }, { onSuccess: onClose })
    } else {
      create.mutate(payload, { onSuccess: onClose })
    }
  }

  return (
    <Overlay onClose={onClose}>
      <h2 className="text-lg font-bold mb-4">{isEdit ? 'Edit' : 'Add'} Transaction</h2>
      <div className="space-y-3">
        <Select label="Type" value={type} onChange={v => setType(v as TxnType)} options={typeOptions.map(t => ({ value: t, label: t }))} />
        <Input label="Amount" type="number" value={amount} onChange={setAmount} />
        <Select label="Category" value={catId} onChange={v => { setCatId(v); setSubcategoryId(''); }} options={categories.map(c => ({ value: c.id, label: c.name }))} />
        <Select label="Sub Category" value={subcategoryId} onChange={setSubcategoryId} options={filteredSubs.map(s => ({ value: s.id, label: s.name }))} />
        <Select label="From (Wallet)" value={srcId} onChange={setSrcId} options={wallets.map(w => ({ value: w.shortName, label: w.name }))} />
        <Select
          label={type === 'Transfer' ? 'To (Wallet)' : 'To'}
          value={dstId}
          onChange={setDstId}
          options={
            type === 'Transfer'
              ? wallets.map(w => ({ value: w.shortName, label: w.name }))
              : contacts.map(c => ({ value: c.nickName, label: c.fullName || c.nickName }))
          }
        />
        {type !== 'Transfer' && (
          <Input label="Contact" value={contactName} onChange={setContactName} />
        )}
        <Input label="Remarks" value={remarks} onChange={setRemarks} />
        <div className="flex gap-2 justify-end pt-2">
          <button className="px-4 py-2 rounded text-sm bg-gray-100 hover:bg-gray-200" onClick={onClose}>Cancel</button>
          <button className="px-4 py-2 rounded text-sm bg-blue-600 text-white hover:bg-blue-700" onClick={handleSubmit}>
            {isEdit ? 'Update' : 'Create'}
          </button>
        </div>
      </div>
    </Overlay>
  )
}

function DeleteDialog({ txn, onClose }: { txn: Transaction; onClose: () => void }) {
  const del = useDeleteTransaction()
  return (
    <Overlay onClose={onClose}>
      <h2 className="text-lg font-bold mb-2">Delete Transaction?</h2>
      <p className="text-sm text-gray-500 mb-4">
        {txn.type} of {txn.amount.toLocaleString(undefined, { minimumFractionDigits: 2 })}
      </p>
      <div className="flex gap-2 justify-end">
        <button className="px-4 py-2 rounded text-sm bg-gray-100 hover:bg-gray-200" onClick={onClose}>Cancel</button>
        <button
          className="px-4 py-2 rounded text-sm bg-red-600 text-white hover:bg-red-700"
          onClick={() => del.mutate(txn.id, { onSuccess: onClose })}
        >
          Delete
        </button>
      </div>
    </Overlay>
  )
}

function Overlay({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-lg shadow-lg p-6 w-full max-w-md mx-4" onClick={e => e.stopPropagation()}>
        {children}
      </div>
    </div>
  )
}

function Input({ label, value, onChange, type }: { label: string; value: string; onChange: (v: string) => void; type?: string }) {
  return (
    <label className="block text-sm">
      <span className="text-gray-600">{label}</span>
      <input className="mt-1 block w-full border rounded px-3 py-2 text-sm" type={type} value={value} onChange={e => onChange(e.target.value)} />
    </label>
  )
}

function Select({ label, value, onChange, options }: { label: string; value: string; onChange: (v: string) => void; options: { value: string; label: string }[] }) {
  return (
    <label className="block text-sm">
      <span className="text-gray-600">{label}</span>
      <select className="mt-1 block w-full border rounded px-3 py-2 text-sm" value={value} onChange={e => onChange(e.target.value)}>
        <option value="">-- select --</option>
        {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </label>
  )
}
