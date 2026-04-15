import { useState } from 'react'
import { useBudgets, useBudgetAlerts, useSetBudget, useDeleteBudget } from '../hooks/useBudgets'
import { useQuery } from '@tanstack/react-query'
import { listCategories } from '../api/endpoints'
import BudgetGauge from '../components/charts/BudgetGauge'

export default function Budgets() {
  const { data: budgets, isLoading } = useBudgets()
  const { data: alerts } = useBudgetAlerts()
  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: listCategories })
  const [showAdd, setShowAdd] = useState(false)

  if (isLoading) return <p className="text-gray-500">Loading...</p>

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Budgets</h1>
        <button
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
          onClick={() => setShowAdd(true)}
        >
          + Set Budget
        </button>
      </div>

      {alerts && alerts.length > 0 && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <h2 className="text-sm font-semibold text-yellow-800 mb-2">Alerts</h2>
          {alerts.map(a => (
            <p key={a.categoryId} className="text-sm text-yellow-700">
              {a.categoryName}: {a.percent.toFixed(0)}% used ({a.spent.toFixed(2)} / {a.budgetAmount.toFixed(2)})
            </p>
          ))}
        </div>
      )}

      {(!budgets || budgets.length === 0) ? (
        <p className="text-gray-400 text-sm">No budgets set</p>
      ) : (
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
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
    <div className="bg-white rounded-lg shadow p-4">
      <div className="flex items-center justify-between mb-2">
        <h3 className="font-semibold text-sm">{budget.categoryName}</h3>
        <button
          className="text-xs text-red-500 hover:underline"
          onClick={() => del.mutate(budget.categoryId)}
        >
          Remove
        </button>
      </div>
      <BudgetGauge percent={budget.percent} />
      <div className="flex justify-between text-xs text-gray-500 mt-2">
        <span>Spent: {budget.spent.toFixed(2)}</span>
        <span>Budget: {budget.amount.toFixed(2)}</span>
      </div>
      <p className="text-sm font-medium mt-1">
        {budget.remaining >= 0 ? (
          <span className="text-green-600">{budget.remaining.toFixed(2)} remaining</span>
        ) : (
          <span className="text-red-600">{Math.abs(budget.remaining).toFixed(2)} over budget</span>
        )}
      </p>
      {budget.alertAt > 0 && (
        <p className="text-xs text-gray-400 mt-1">Alert at {budget.alertAt}%</p>
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
    if (!categoryId || !amount) return
    setBudget.mutate(
      { categoryId, amount: parseFloat(amount), alertAt: parseInt(alertAt) },
      { onSuccess: onClose },
    )
  }

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-lg shadow-lg p-6 w-full max-w-sm mx-4" onClick={e => e.stopPropagation()}>
        <h2 className="text-lg font-bold mb-4">Set Budget</h2>
        <div className="space-y-3">
          <label className="block text-sm">
            <span className="text-gray-600">Category</span>
            <select
              className="mt-1 block w-full border rounded px-3 py-2 text-sm"
              value={categoryId}
              onChange={e => setCategoryId(e.target.value)}
            >
              <option value="">Overall</option>
              {available.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </label>
          <label className="block text-sm">
            <span className="text-gray-600">Monthly Budget</span>
            <input
              type="number"
              className="mt-1 block w-full border rounded px-3 py-2 text-sm"
              value={amount}
              onChange={e => setAmount(e.target.value)}
            />
          </label>
          <label className="block text-sm">
            <span className="text-gray-600">Alert at (%)</span>
            <input
              type="number"
              className="mt-1 block w-full border rounded px-3 py-2 text-sm"
              value={alertAt}
              onChange={e => setAlertAt(e.target.value)}
            />
          </label>
          <div className="flex gap-2 justify-end pt-2">
            <button className="px-4 py-2 rounded text-sm bg-gray-100 hover:bg-gray-200" onClick={onClose}>Cancel</button>
            <button className="px-4 py-2 rounded text-sm bg-blue-600 text-white hover:bg-blue-700" onClick={handleSubmit}>Save</button>
          </div>
        </div>
      </div>
    </div>
  )
}
