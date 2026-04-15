import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { listWallets, listContacts } from '../api/endpoints'
import { useCreateWallet } from '../hooks/useWallets'
import { useCreateContact } from '../hooks/useContacts'
import { Plus } from 'lucide-react'

export default function Wallets() {
  const { data: wallets, isLoading: wLoading } = useQuery({ queryKey: ['wallets'], queryFn: listWallets })
  const { data: contacts, isLoading: cLoading } = useQuery({ queryKey: ['contacts'], queryFn: listContacts })
  
  const [showAddWallet, setShowAddWallet] = useState(false)
  const [showAddContact, setShowAddContact] = useState(false)

  if (wLoading || cLoading) return <p className="text-gray-500">Loading...</p>

  return (
    <div className="space-y-8">
      <section>
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-2xl font-bold">Wallets</h1>
          <button
            onClick={() => setShowAddWallet(true)}
            className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors shadow-sm"
          >
            <Plus size={18} />
            Add Wallet
          </button>
        </div>
        {(!wallets || wallets.length === 0) ? (
          <p className="text-gray-400 text-sm">No wallets</p>
        ) : (
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {wallets.map(w => (
              <div key={w.id} className="bg-white rounded-lg shadow p-4 border-l-4 border-blue-500">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-semibold">{w.name}</h3>
                  <span className="text-xs bg-gray-100 px-2 py-0.5 rounded font-medium text-gray-600">{w.type}</span>
                </div>
                <p className={`text-xl font-bold ${w.balance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                  {w.balance.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                </p>
                <p className="text-xs text-gray-400 mt-1 font-mono">{w.shortName}</p>
              </div>
            ))}
          </div>
        )}
      </section>

      <section>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold">Contacts</h2>
          <button
            onClick={() => setShowAddContact(true)}
            className="flex items-center gap-2 bg-gray-100 text-gray-700 px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-200 transition-colors"
          >
            <Plus size={18} />
            Add Contact
          </button>
        </div>
        {(!contacts || contacts.length === 0) ? (
          <p className="text-gray-400 text-sm">No contacts</p>
        ) : (
          <div className="bg-white rounded-lg shadow overflow-hidden border border-gray-100">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-gray-500 bg-gray-50 border-b">
                    <th className="p-4 font-semibold">Nickname</th>
                    <th className="p-4 font-semibold">Full Name</th>
                    <th className="p-4 font-semibold">Email</th>
                    <th className="p-4 font-semibold">Net Balance</th>
                    <th className="p-4 font-semibold">Last Txn</th>
                  </tr>
                </thead>
                <tbody>
                  {contacts.map(c => (
                    <tr key={c.id} className="border-b last:border-0 hover:bg-gray-50 transition-colors">
                      <td className="p-4 font-medium">{c.nickName}</td>
                      <td className="p-4">{c.fullName}</td>
                      <td className="p-4 text-gray-500">{c.email || '-'}</td>
                      <td className={`p-4 font-bold ${c.netBalance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                        {c.netBalance.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-4 text-gray-500">
                        {c.lastTxnTimestamp ? new Date(c.lastTxnTimestamp * 1000).toLocaleDateString() : '-'}
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
      <h2 className="text-lg font-bold mb-4">Add Wallet</h2>
      <div className="space-y-3">
        <Select label="Type" value={type} onChange={setType} options={[{ value: 'Bank', label: 'Bank' }, { value: 'Cash', label: 'Cash' }]} />
        <Input label="Short Name (e.g. brac)" value={shortName} onChange={setShortName} />
        <Input label="Full Name" value={name} onChange={setName} />
        <Input label="Initial Balance" type="number" value={balance} onChange={setBalance} />
        <div className="flex gap-2 justify-end pt-2">
          <button className="px-4 py-2 rounded text-sm bg-gray-100 hover:bg-gray-200" onClick={onClose}>Cancel</button>
          <button className="px-4 py-2 rounded text-sm bg-blue-600 text-white hover:bg-blue-700" onClick={handleSubmit}>Create</button>
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
      <h2 className="text-lg font-bold mb-4">Add Contact</h2>
      <div className="space-y-3">
        <Input label="Nick Name" value={nickName} onChange={setNickName} />
        <Input label="Full Name" value={fullName} onChange={setFullName} />
        <Input label="Email" value={email} onChange={setEmail} />
        <div className="flex gap-2 justify-end pt-2">
          <button className="px-4 py-2 rounded text-sm bg-gray-100 hover:bg-gray-200" onClick={onClose}>Cancel</button>
          <button className="px-4 py-2 rounded text-sm bg-blue-600 text-white hover:bg-blue-700" onClick={handleSubmit}>Create</button>
        </div>
      </div>
    </Overlay>
  )
}

function Overlay({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-white rounded-xl shadow-xl p-6 w-full max-w-md" onClick={e => e.stopPropagation()}>
        {children}
      </div>
    </div>
  )
}

function Input({ label, value, onChange, type }: { label: string; value: string; onChange: (v: string) => void; type?: string }) {
  return (
    <label className="block text-sm">
      <span className="text-gray-600 font-medium">{label}</span>
      <input 
        className="mt-1 block w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all" 
        type={type} 
        value={value} 
        onChange={e => onChange(e.target.value)} 
      />
    </label>
  )
}

function Select({ label, value, onChange, options }: { label: string; value: string; onChange: (v: string) => void; options: { value: string; label: string }[] }) {
  return (
    <label className="block text-sm">
      <span className="text-gray-600 font-medium">{label}</span>
      <select 
        className="mt-1 block w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all" 
        value={value} 
        onChange={e => onChange(e.target.value)}
      >
        <option value="">-- select --</option>
        {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </label>
  )
}
