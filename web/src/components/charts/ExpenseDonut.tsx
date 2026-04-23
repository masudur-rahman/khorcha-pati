import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer } from 'recharts'
import type { CategorySpend } from '../../types'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16']

export default function ExpenseDonut({ data }: { data: CategorySpend[] }) {
  if (!data.length) return <p className="text-gray-400 text-sm text-center">No expenses</p>

  return (
    <ResponsiveContainer width="100%" height={250}>
      <PieChart>
        <Pie data={data} dataKey="amount" nameKey="categoryName" cx="50%" cy="50%" innerRadius={60} outerRadius={90} paddingAngle={2}>
          {data.map((_, i) => (
            <Cell key={i} fill={COLORS[i % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip formatter={(value) => Number(value).toFixed(2)} />
      </PieChart>
    </ResponsiveContainer>
  )
}
