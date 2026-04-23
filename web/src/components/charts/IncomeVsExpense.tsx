import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import type { MonthlyComparison } from '../../types'

export default function IncomeVsExpense({ data }: { data: MonthlyComparison[] }) {
  if (!data.length) return <p className="text-gray-400 text-sm text-center">No data</p>

  return (
    <ResponsiveContainer width="100%" height={250}>
      <BarChart data={data}>
        <XAxis dataKey="month" tick={{ fontSize: 12 }} />
        <YAxis tick={{ fontSize: 12 }} />
        <Tooltip />
        <Legend />
        <Bar dataKey="income" fill="#10b981" name="Income" />
        <Bar dataKey="expense" fill="#ef4444" name="Expense" />
      </BarChart>
    </ResponsiveContainer>
  )
}
