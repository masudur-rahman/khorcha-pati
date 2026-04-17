import { useState } from 'react'
import { useBudgets, useBudgetAlerts, useSetBudget, useDeleteBudget } from '../hooks/useBudgets'
import { useQuery } from '@tanstack/react-query'
import { listCategories } from '../api/endpoints'
import BudgetGauge from '../components/charts/BudgetGauge'
import { Plus, Trash2, AlertCircle } from 'lucide-react'
import { fmt } from '../lib/formatter'

export default function Budgets() {
  const { data: budgets, isLoading } = useBudgets()
  const { data: alerts } = useBudgetAlerts()
  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: listCategories })
  const [showAdd, setShowAdd] = useState(false)

  if (isLoading) return <p className="text-gray-500">Loading...</p>

  return (
    <div className="space-y-8 pb-8">
      <header className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Budgets</h1>
          <p className="text-gray-500 text-sm mt-1">Plan your monthly spending limits</p>
        </div>
        <button
          className="flex items-center justify-center gap-2 bg-blue-600 text-white px-6 py-3 rounded-2xl text-sm font-bold hover:bg-blue-700 transition-all shadow-lg shadow-blue-100 group cursor-pointer"
          onClick={() => setShowAdd(true)}
        >
          <Plus size={18} className="group-hover:rotate-90 transition-transform" />
          Set Budget
        </button>
      </header>

      {alerts && alerts.length > 0 && (
        <div className="bg-red-50 border border-red-100 rounded-3xl p-6 flex items-start gap-4">
          <div className="p-2 bg-red-100 text-red-600 rounded-xl">
            <AlertCircle size={24} />
          </div>
          <div className="flex-1">
            <h2 className="text-sm font-bold text-red-900 uppercase tracking-widest mb-2">Spending Alerts</h2>
            <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {alerts.map(a => (
                <div key={a.categoryId} className="bg-white/50 p-3 rounded-xl border border-red-200/50">
                  <p className="text-xs font-bold text-red-800">{a.categoryName}</p>
                  <p className="text-[10px] text-red-600 font-bold uppercase tracking-tighter mt-0.5">
                    {a.percent.toFixed(0)}% used • {(a.budgetAmount - a.spent) < 0 ? <span>Over by <span className="whitespace-nowrap">{fmt(Math.abs(a.budgetAmount - a.spent), 0)}</span></span> : <span><span className="whitespace-nowrap">{fmt(a.budgetAmount - a.spent, 0)}</span> left</span>}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {(!budgets || budgets.length === 0) ? (
        <div className="bg-white rounded-3xl p-12 text-center border border-dashed border-gray-200">
          <div className="text-4xl mb-4">🎯</div>
          <p className="text-gray-400 font-medium">No budgets set. Create one to start tracking your goals.</p>
        </div>
      ) : (
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {budgets.map(b => (
            <BudgetCard key={b.categoryId} budget={b} />
          ))}
        </div>
      )}

      {showAdd && (
        <SetBudgetDialog
          categories={categories ?? []}
          existingIds={(budgets ?? []).map(b => b.categoryId)}
          onClose={() => setShowAdd(false)}
        />
      )}
    </div>
  )
}

function BudgetCard({ budget }: { budget: import('../types').BudgetStatus }) {
  const del = useDeleteBudget()
  return (
    <div className="bg-white rounded-[2rem] shadow-sm p-8 border border-gray-100 hover:border-blue-100 transition-all group cursor-pointer">
      <div className="flex items-center justify-between mb-8">
        <h3 className="font-bold text-gray-900 tracking-tight">{budget.categoryName}</h3>
        <button
          className="p-2 text-gray-300 hover:text-red-500 hover:bg-red-50 rounded-xl transition-all"
          onClick={() => del.mutate(budget.categoryId)}
          title="Remove Budget"
        >
          <Trash2 size={16} />
        </button>
      </div>
      
      <div className="mb-8">
        <BudgetGauge percent={budget.percent} />
      </div>

      <div className="space-y-4">
        <div className="flex justify-between items-end">
            <div>
                <p className="text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Spent</p>
                <p className="text-lg font-bold text-gray-900 whitespace-nowrap">{fmt(budget.spent, 0)}</p>
            </div>
            <div className="text-right">
                <p className="text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Limit</p>
                <p className="text-lg font-bold text-gray-400 whitespace-nowrap">{fmt(budget.amount, 0)}</p>
            </div>
        </div>

        <div className={`p-4 rounded-2xl flex items-center justify-between ${budget.remaining >= 0 ? 'bg-emerald-50' : 'bg-red-50'}`}>
            <span className={`text-xs font-bold ${budget.remaining >= 0 ? 'text-emerald-700' : 'text-red-700'}`}>
                {budget.remaining >= 0 ? 'Remaining' : 'Over Budget'}
            </span>
            <span className={`text-sm font-bold whitespace-nowrap ${budget.remaining >= 0 ? 'text-emerald-600' : 'text-red-600'}`}>
                {fmt(Math.abs(budget.remaining))}
            </span>
        </div>
      </div>

      {budget.alertAt > 0 && (
        <p className="text-[10px] text-gray-400 font-bold uppercase tracking-widest mt-6 text-center">
            Alert configured at {budget.alertAt}%
        </p>
      )}
    </div>
  )
}

interface SetBudgetDialogProps {
  categories: import('../types').TxnCategory[]
  existingIds: string[]
  onClose: () => void
}

function SetBudgetDialog({ categories, existingIds, onClose }: SetBudgetDialogProps) {
  const setBudget = useSetBudget()
  const [categoryId, setCategoryId] = useState('')
  const [amount, setAmount] = useState('')
  const [alertAt, setAlertAt] = useState('80')

  const available = categories.filter(c => !existingIds.includes(c.id))

  const handleSubmit = () => {
    if (!amount) return
    setBudget.mutate(
      { categoryId: categoryId || 'overall', amount: parseFloat(amount), alertAt: parseInt(alertAt) },
      { onSuccess: onClose },
    )
  }

  return (
    <Overlay onClose={onClose}>
      <h2 className="text-xl font-bold text-gray-900 mb-6">Set New Budget</h2>
      <div className="space-y-4">
        <label className="block space-y-1.5">
          <span className="text-[10px] font-bold uppercase tracking-widest text-gray-400 ml-1">Category</span>
          <select
            className="w-full bg-gray-50 border border-gray-100 rounded-2xl px-4 py-3 text-sm focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all outline-none font-medium appearance-none cursor-pointer"
            value={categoryId}
            onChange={e => setCategoryId(e.target.value)}
          >
            <option value="">Overall Budget</option>
            {available.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
        </label>
        
        <Input label="Monthly Budget Amount" type="number" value={amount} onChange={setAmount} />
        
        <Input label="Alert Threshold (%)" type="number" value={alertAt} onChange={setAlertAt} />

        <div className="flex gap-3 justify-end pt-4">
          <button className="px-6 py-3 rounded-2xl text-sm font-bold text-gray-400 hover:text-gray-600 hover:bg-gray-50 transition-colors cursor-pointer" onClick={onClose}>Cancel</button>
          <button 
            className="px-8 py-3 rounded-2xl text-sm font-bold bg-blue-600 text-white hover:bg-blue-700 transition-all shadow-lg shadow-blue-100 cursor-pointer" 
            onClick={handleSubmit}
            disabled={!amount}
          >
            Save Budget
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
