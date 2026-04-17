import { useState, useMemo } from 'react'
import { useTransactions, useCreateTransaction, useUpdateTransaction, useDeleteTransaction } from '../hooks/useTransactions'
import { useWallets } from '../hooks/useWallets'
import { useContacts } from '../hooks/useContacts'
import { useQuery } from '@tanstack/react-query'
import { listCategories, listSubcategories } from '../api/endpoints'
import type { Transaction, Wallet, Contact, TxnCategory } from '../types'
import { Plus, Edit2, Trash2 } from 'lucide-react'
import { fmt } from '../lib/formatter'

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

  const filtered = txns
    .filter(t => !filterType || t.type === filterType)
    .sort((a, b) => b.timestamp - a.timestamp)

  return (
    <div className="space-y-8 pb-8">
      <header className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
            <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Transactions</h1>
            <p className="text-gray-500 text-sm mt-1">Detailed history of your financial movements</p>
        </div>
        <button 
          onClick={() => setShowAdd(true)} 
          className="flex items-center justify-center gap-2 bg-blue-600 text-white px-6 py-3 rounded-2xl text-sm font-bold hover:bg-blue-700 transition-all shadow-lg shadow-blue-100 group cursor-pointer"
        >
          <Plus size={18} className="group-hover:rotate-90 transition-transform" />
          Add Transaction
        </button>
      </header>

      <div className="flex flex-wrap items-center gap-2 bg-white p-2 rounded-2xl border border-gray-100 shadow-sm w-fit">
        <FilterButton label="All" active={filterType === ''} onClick={() => setFilterType('')} />
        {typeOptions.map(t => (
          <FilterButton key={t} label={t} active={filterType === t} onClick={() => setFilterType(t)} />
        ))}
      </div>

      <div className="bg-white rounded-3xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-400 border-b border-gray-50 uppercase text-[10px] tracking-widest font-bold">
                <th className="px-6 py-5">Type</th>
                <th className="px-6 py-5">Category</th>
                <th className="px-6 py-5">Wallets / Contact</th>
                <th className="px-6 py-5">Amount</th>
                <th className="px-6 py-5">Date</th>
                <th className="px-6 py-5">Remarks</th>
                <th className="px-6 py-5 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {filtered.length === 0 ? (
                <tr><td colSpan={7} className="p-20 text-center text-gray-400 font-medium">No transactions found</td></tr>
              ) : filtered.map(t => (
                <tr key={t.id} className="hover:bg-gray-50/50 transition-colors group">
                  <td className="px-6 py-4">
                    <span className={`px-2.5 py-1 rounded-lg text-[10px] font-bold uppercase ${
                        t.type === 'Income' ? 'bg-green-100 text-green-700' : 
                        t.type === 'Transfer' ? 'bg-blue-100 text-blue-700' : 
                        'bg-red-100 text-red-700'
                    }`}>
                        {t.type}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="font-bold text-gray-900">{subcatMap.get(t.subcategoryId) || t.subcategoryId}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2 text-xs font-medium text-gray-500">
                      <span className="bg-gray-100 px-1.5 py-0.5 rounded uppercase tracking-tighter font-bold">{t.srcId || '-'}</span>
                      <span className="text-gray-300">→</span>
                      <span className="bg-gray-100 px-1.5 py-0.5 rounded uppercase tracking-tighter font-bold">{t.dstId || t.contactName || '-'}</span>
                    </div>
                  </td>
                  <td className={`px-6 py-4 font-bold text-base whitespace-nowrap ${
                    t.type === 'Income' ? 'text-green-600' : 
                    t.type === 'Transfer' ? 'text-blue-600' : 
                    'text-red-600'
                  }`}>
                    {t.type === 'Income' ? '+' : t.type === 'Transfer' ? '' : '-'}{fmt(t.amount)}
                  </td>
                  <td className="px-6 py-4 text-gray-400 font-bold text-xs uppercase">
                    {new Date(t.timestamp * 1000).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-gray-500 text-xs truncate max-w-[120px] font-medium italic">{t.remarks || 'No remarks'}</p>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button 
                            className="p-2 text-blue-600 hover:bg-blue-50 rounded-xl transition-colors cursor-pointer" 
                            onClick={() => setEditTxn(t)}
                            title="Edit"
                        >
                            <Edit2 size={16} />
                        </button>
                        <button 
                            className="p-2 text-red-600 hover:bg-red-50 rounded-xl transition-colors cursor-pointer" 
                            onClick={() => setDeleteTxn(t)}
                            title="Delete"
                        >
                            <Trash2 size={16} />
                        </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {(showAdd || editTxn) && (
        <TxnDialog
          txn={editTxn || undefined}
          wallets={wallets ?? []}
          contacts={contacts ?? []}
          categories={categories ?? []}
          subcategories={subcategories ?? []}
          onClose={() => { setShowAdd(false); setEditTxn(null); }}
        />
      )}
      {deleteTxn && <DeleteDialog txn={deleteTxn} onClose={() => setDeleteTxn(null)} />}
    </div>
  )
}

function FilterButton({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      className={`px-6 py-2 rounded-xl text-sm font-bold transition-all cursor-pointer ${
        active 
          ? 'bg-blue-600 text-white shadow-md shadow-blue-100' 
          : 'text-gray-400 hover:text-gray-600 hover:bg-gray-50'
      }`}
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
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-gray-900">{isEdit ? 'Edit' : 'Add'} Transaction</h2>
        <span className={`px-2 py-1 rounded-lg text-[10px] font-bold uppercase ${
            type === 'Income' ? 'bg-green-100 text-green-700' : 
            type === 'Transfer' ? 'bg-blue-100 text-blue-700' : 
            'bg-red-100 text-red-700'
        }`}>{type}</span>
      </div>
      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
            <Select label="Type" value={type} onChange={v => setType(v as TxnType)} options={typeOptions.map(t => ({ value: t, label: t }))} />
            <Input label="Amount" type="number" value={amount} onChange={setAmount} />
        </div>
        <div className="grid grid-cols-2 gap-4">
            <Select label="Category" value={catId} onChange={v => { setCatId(v); setSubcategoryId(''); }} options={categories.map(c => ({ value: c.id, label: c.name }))} />
            <Select label="Sub Category" value={subcategoryId} onChange={setSubcategoryId} options={filteredSubs.map(s => ({ value: s.id, label: s.name }))} />
        </div>
        <div className="grid grid-cols-2 gap-4">
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
        </div>
        {type !== 'Transfer' && (
          <Input label="Contact (Optional)" value={contactName} onChange={setContactName} />
        )}
        <Input label="Remarks" value={remarks} onChange={setRemarks} />
        
        <div className="flex gap-3 justify-end pt-4">
          <button className="px-6 py-3 rounded-2xl text-sm font-bold text-gray-400 hover:text-gray-600 hover:bg-gray-50 transition-colors cursor-pointer" onClick={onClose}>Cancel</button>
          <button 
            className="px-8 py-3 rounded-2xl text-sm font-bold bg-blue-600 text-white hover:bg-blue-700 transition-all shadow-lg shadow-blue-100 cursor-pointer" 
            onClick={handleSubmit}
            disabled={!amount || !subcategoryId}
          >
            {isEdit ? 'Update Changes' : 'Create Transaction'}
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
      <div className="text-center">
        <div className="w-16 h-16 bg-red-50 text-red-600 rounded-full flex items-center justify-center mx-auto mb-4 text-2xl">🗑️</div>
        <h2 className="text-xl font-bold text-gray-900 mb-2">Delete Transaction?</h2>
        <p className="text-sm text-gray-500 mb-8 leading-relaxed">
            Are you sure you want to delete this <span className="font-bold">{txn.type}</span> of <span className="font-bold text-gray-900 whitespace-nowrap">{fmt(txn.amount)}</span>? This action cannot be undone.
        </p>
        <div className="flex gap-3 justify-center">
            <button className="px-6 py-3 rounded-2xl text-sm font-bold text-gray-400 hover:text-gray-600 hover:bg-gray-50 transition-colors cursor-pointer" onClick={onClose}>Cancel</button>
            <button
            className="px-8 py-3 rounded-2xl text-sm font-bold bg-red-600 text-white hover:bg-red-700 transition-all shadow-lg shadow-red-100"
            onClick={() => del.mutate(txn.id, { onSuccess: onClose })}
            >
            Confirm Delete
            </button>
        </div>
      </div>
    </Overlay>
  )
}

function Overlay({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 bg-gray-900/60 backdrop-blur-sm flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-white rounded-[2rem] shadow-2xl p-8 w-full max-w-xl animate-in fade-in zoom-in duration-200" onClick={e => e.stopPropagation()}>
        {children}
      </div>
    </div>
  )
}

function Input({ label, value, onChange, type }: { label: string; value: string; onChange: (v: string) => void; type?: string }) {
  return (
    <label className="block space-y-1.5">
      <span className="text-[10px] font-bold uppercase tracking-widest text-gray-400 ml-1">{label}</span>
      <input 
        className="w-full bg-gray-50 border border-gray-100 rounded-2xl px-4 py-3 text-sm focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all outline-none font-medium" 
        type={type} 
        value={value} 
        onChange={e => onChange(e.target.value)} 
      />
    </label>
  )
}

function Select({ label, value, onChange, options }: { label: string; value: string; onChange: (v: string) => void; options: { value: string; label: string }[] }) {
  return (
    <label className="block space-y-1.5">
      <span className="text-[10px] font-bold uppercase tracking-widest text-gray-400 ml-1">{label}</span>
      <select 
        className="w-full bg-gray-50 border border-gray-100 rounded-2xl px-4 py-3 text-sm focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all outline-none font-medium appearance-none cursor-pointer" 
        value={value} 
        onChange={e => onChange(e.target.value)}
      >
        <option value="">Select...</option>
        {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </label>
  )
}
