import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listWallets, listContacts } from '../api/endpoints'
import { useCreateWallet } from '../hooks/useWallets'
import { useCreateContact } from '../hooks/useContacts'
import { Plus, CreditCard, Banknote } from 'lucide-react'
import { fmt } from '../lib/formatter'

export default function Wallets() {
  const { data: wallets, isLoading: wLoading } = useQuery({ queryKey: ['wallets'], queryFn: listWallets })
  const { data: contacts, isLoading: cLoading } = useQuery({ queryKey: ['contacts'], queryFn: listContacts })
  
  const [showAddWallet, setShowAddWallet] = useState(false)
  const [showAddContact, setShowAddContact] = useState(false)

  if (wLoading || cLoading) return <p className="text-gray-500">Loading...</p>

  return (
    <div className="space-y-12 pb-8">
      <section>
        <header className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Wallets</h1>
            <p className="text-gray-500 text-sm mt-1">Manage your bank accounts and cash</p>
          </div>
          <button
            onClick={() => setShowAddWallet(true)}
            className="flex items-center justify-center gap-2 bg-blue-600 text-white px-6 py-3 rounded-2xl text-sm font-bold hover:bg-blue-700 transition-all shadow-lg shadow-blue-100 group cursor-pointer"
          >
            <Plus size={18} className="group-hover:rotate-90 transition-transform" />
            Add Wallet
          </button>
        </header>

        {(!wallets || wallets.length === 0) ? (
          <div className="bg-white rounded-3xl p-12 text-center border border-dashed border-gray-200">
            <div className="text-4xl mb-4">💳</div>
            <p className="text-gray-400 font-medium">No wallets found. Add one to get started.</p>
          </div>
        ) : (
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {wallets.map(w => (
              <div key={w.id} className="bg-white rounded-[2rem] shadow-sm p-8 border border-gray-100 hover:border-blue-200 transition-all group relative overflow-hidden cursor-pointer">
                <div className="absolute top-0 right-0 p-6 opacity-5 group-hover:opacity-10 transition-opacity">
                    {w.type === 'Bank' ? <CreditCard size={80} /> : <Banknote size={80} />}
                </div>
                <div className="flex items-center justify-between mb-6">
                  <div className={`p-3 rounded-2xl ${w.type === 'Bank' ? 'bg-blue-50 text-blue-600' : 'bg-emerald-50 text-emerald-600'}`}>
                    {w.type === 'Bank' ? <CreditCard size={20} /> : <Banknote size={20} />}
                  </div>
                  <span className="text-[10px] bg-gray-50 px-2.5 py-1 rounded-lg font-bold text-gray-400 uppercase tracking-widest">{w.type}</span>
                </div>
                <h3 className="text-lg font-bold text-gray-900 mb-1">{w.name}</h3>
                <p className="text-xs text-gray-400 font-bold uppercase tracking-widest mb-4">{w.shortName}</p>
                <p className={`text-3xl font-bold whitespace-nowrap ${w.balance >= 0 ? 'text-gray-900' : 'text-red-600'}`}>
                  {fmt(w.balance)}
                </p>
              </div>
            ))}
          </div>
        )}
      </section>

      <section>
        <header className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
          <div>
            <h2 className="text-2xl font-bold text-gray-900 tracking-tight">Contacts</h2>
            <p className="text-gray-500 text-sm mt-1">People you transact with frequently</p>
          </div>
          <button
            onClick={() => setShowAddContact(true)}
            className="flex items-center justify-center gap-2 bg-gray-900 text-white px-6 py-3 rounded-2xl text-sm font-bold hover:bg-black transition-all shadow-lg shadow-gray-200 group cursor-pointer"
          >
            <Plus size={18} className="group-hover:rotate-90 transition-transform" />
            Add Contact
          </button>
        </header>

        {(!contacts || contacts.length === 0) ? (
          <div className="bg-white rounded-3xl p-12 text-center border border-dashed border-gray-200">
            <div className="text-4xl mb-4">👥</div>
            <p className="text-gray-400 font-medium">No contacts found.</p>
          </div>
        ) : (
          <div className="bg-white rounded-3xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-gray-400 border-b border-gray-50 uppercase text-[10px] tracking-widest font-bold">
                    <th className="px-8 py-5">Contact</th>
                    <th className="px-8 py-5">Email</th>
                    <th className="px-8 py-5">Net Balance</th>
                    <th className="px-8 py-5">Last Transaction</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-50">
                  {contacts.map(c => (
                    <tr key={c.id} className="hover:bg-gray-50/50 transition-colors group cursor-pointer">
                      <td className="px-8 py-5">
                        <div className="flex items-center gap-4">
                            <div className="w-10 h-10 bg-gray-100 rounded-full flex items-center justify-center text-gray-400 font-bold uppercase">
                                {c.nickName.slice(0, 2)}
                            </div>
                            <div>
                                <div className="font-bold text-gray-900">{c.fullName || c.nickName}</div>
                                <div className="text-[10px] text-gray-400 font-bold uppercase tracking-widest">{c.nickName}</div>
                            </div>
                        </div>
                      </td>
                      <td className="px-8 py-5 text-gray-500 font-medium">{c.email || '—'}</td>
                      <td className={`px-8 py-5 font-bold text-base whitespace-nowrap ${c.netBalance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                        {c.netBalance >= 0 ? '+' : ''}{fmt(c.netBalance)}
                      </td>
                      <td className="px-8 py-5 text-gray-400 font-bold text-xs uppercase">
                        {c.lastTxnTimestamp ? new Date(c.lastTxnTimestamp * 1000).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) : 'Never'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
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

  const handleSubmit = () => {
    create.mutate({ type: type as any, shortName, name, balance: parseFloat(balance) }, { onSuccess: onClose })
  }

  return (
    <Overlay onClose={onClose}>
      <h2 className="text-xl font-bold text-gray-900 mb-6">Add New Wallet</h2>
      <div className="space-y-4">
        <Select label="Type" value={type} onChange={setType} options={[{ value: 'Bank', label: 'Bank Account' }, { value: 'Cash', label: 'Cash / Other' }]} />
        <Input label="Short Name (e.g. brac, cash)" value={shortName} onChange={setShortName} />
        <Input label="Display Name" value={name} onChange={setName} />
        <Input label="Initial Balance" type="number" value={balance} onChange={setBalance} />
        <div className="flex gap-3 justify-end pt-4">
          <button className="px-6 py-3 rounded-2xl text-sm font-bold text-gray-400 hover:text-gray-600 hover:bg-gray-50 transition-colors cursor-pointer" onClick={onClose}>Cancel</button>
          <button 
            className="px-8 py-3 rounded-2xl text-sm font-bold bg-blue-600 text-white hover:bg-blue-700 transition-all shadow-lg shadow-blue-100 cursor-pointer" 
            onClick={handleSubmit}
            disabled={!name || !shortName || !balance}
          >
            Create Wallet
          </button>
        </div>
      </div>
    </Overlay>
  )
}

function AddContactDialog({ onClose }: { onClose: () => void }) {
  const create = useCreateContact()
  const [nickName, setNickName] = useState('')
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')

  const handleSubmit = () => {
    create.mutate({ nickName, fullName, email }, { onSuccess: onClose })
  }

  return (
    <Overlay onClose={onClose}>
      <h2 className="text-xl font-bold text-gray-900 mb-6">Add New Contact</h2>
      <div className="space-y-4">
        <Input label="Nick Name (Unique)" value={nickName} onChange={setNickName} />
        <Input label="Full Name" value={fullName} onChange={setFullName} />
        <Input label="Email Address" value={email} onChange={setEmail} />
        <div className="flex gap-3 justify-end pt-4">
          <button className="px-6 py-3 rounded-2xl text-sm font-bold text-gray-400 hover:text-gray-600 hover:bg-gray-50 transition-colors cursor-pointer" onClick={onClose}>Cancel</button>
          <button 
            className="px-8 py-3 rounded-2xl text-sm font-bold bg-gray-900 text-white hover:bg-black transition-all shadow-lg shadow-gray-200 cursor-pointer" 
            onClick={handleSubmit}
            disabled={!nickName}
          >
            Create Contact
          </button>
        </div>
      </div>
    </Overlay>
  )
}

function Overlay({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 bg-gray-900/60 backdrop-blur-sm flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-white rounded-[2rem] shadow-2xl p-8 w-full max-w-md animate-in fade-in zoom-in duration-200" onClick={e => e.stopPropagation()}>
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
