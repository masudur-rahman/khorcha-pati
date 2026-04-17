import {useMemo, useState} from 'react'
import { Link } from 'react-router-dom'
import {useQuery} from '@tanstack/react-query'
import {downloadReport, getChartData, listCategories, listSubcategories} from '../api/endpoints'
import ExpenseDonut from '../components/charts/ExpenseDonut'
import IncomeVsExpense from '../components/charts/IncomeVsExpense'
import BudgetGauge from '../components/charts/BudgetGauge'
import {useTransactions} from '../hooks/useTransactions'
import {Download, FileText} from 'lucide-react'
import { fmt } from '../lib/formatter'

export default function Dashboard() {
    const {data: charts, isLoading: isChartsLoading} = useQuery({
        queryKey: ['chartData'],
        queryFn: () => getChartData(),
    })
    const {data: resp} = useTransactions()
    const txns = resp?.data ?? []
    const {data: subcategories, isLoading: isSubsLoading} = useQuery({queryKey: ['subcategories'], queryFn: () => listSubcategories()})
    const {data: allCategories, isLoading: isCatsLoading} = useQuery({queryKey: ['categories'], queryFn: () => listCategories()})
    const [showReportModal, setShowReportModal] = useState(false)
    const [isDownloading, setIsDownloading] = useState(false)

    const isLoading = isChartsLoading || isSubsLoading || isCatsLoading

    const subcatMap = useMemo(() => {
        const m = new Map<string, string>()
        subcategories?.forEach(s => m.set(s.id, s.name))
        return m
    }, [subcategories])

    const catMap = useMemo(() => {
        const m = new Map<string, string>()
        allCategories?.forEach(c => m.set(c.id, c.name))
        return m
    }, [allCategories])

    const chartCategories = useMemo(() => {
        const categorySpends = charts?.categories || []
        return categorySpends.map(c => ({
            ...c,
            categoryName: catMap.get(c.categoryId) || c.categoryName || c.categoryId
        }))
    }, [charts?.categories, catMap])

    if (isLoading) return <p className="text-gray-500">Loading...</p>
    if (!charts) return null

    const overview = charts.overview
    const comparison = charts.comparison || []
    const recentTxns = [...txns].sort((a, b) => b.timestamp - a.timestamp).slice(0, 10)

    const handleDownloadReport = async (duration: string) => {
        try {
            setIsDownloading(true)
            const blob = await downloadReport(duration)
            const url = window.URL.createObjectURL(blob as any)
            window.open(url, '_blank')
            setShowReportModal(false)
        } catch (err) {
            alert('Failed to generate statement: ' + err)
        } finally {
            setIsDownloading(false)
        }
    }

    return (
        <div className="space-y-8 pb-8">
            <header className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Dashboard</h1>
                    <p className="text-gray-500 text-sm mt-1 font-medium">Summary of your financial activity</p>
                </div>
                <button
                    onClick={() => setShowReportModal(true)}
                    className="flex items-center justify-center gap-2 bg-blue-600 text-white px-6 py-3 rounded-2xl text-sm font-bold hover:bg-blue-700 transition-all shadow-lg shadow-blue-100 group cursor-pointer"
                >
                    <Download size={18} className="group-hover:translate-y-0.5 transition-transform"/>
                    Generate Statement
                </button>
            </header>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                <Card label="Total Balance" value={fmt(overview.totalBalance)} color="text-blue-600" icon="💰" />
                <Card label="Month Income" value={fmt(overview.monthIncome)} color="text-green-600" icon="📈" />
                <Card label="Month Expense" value={fmt(overview.monthExpense)} color="text-red-600" icon="📉" />
                <div className="bg-white rounded-2xl shadow-sm p-6 border border-gray-100 flex flex-col justify-between">
                    <div>
                        <div className="flex items-center justify-between mb-2">
                            <p className="text-xs text-gray-400 font-bold uppercase tracking-wider">Budget Usage</p>
                            <span className="text-xl">🎯</span>
                        </div>
                        <BudgetGauge percent={overview.budgetUsage}/>
                    </div>
                    <div className="mt-4 flex items-end justify-between">
                        <p className="text-2xl font-bold text-gray-900">{overview.budgetUsage.toFixed(0)}%</p>
                        <p className="text-xs text-gray-400 font-medium pb-1">of monthly limit</p>
                    </div>
                </div>
            </div>

            <div className="grid lg:grid-cols-2 gap-8">
                <section className="bg-white rounded-2xl shadow-sm p-6 border border-gray-100">
                    <h2 className="text-lg font-bold text-gray-900 mb-6 flex items-center gap-2">
                        <span className="w-2 h-6 bg-blue-600 rounded-full"></span>
                        Expense by Category
                    </h2>
                    <div className="h-64 flex items-center justify-center">
                        {chartCategories.length > 0 ? (
                            <ExpenseDonut data={chartCategories}/>
                        ) : (
                            <p className="text-gray-400 text-sm">No categorical data available</p>
                        )}
                    </div>
                </section>
                <section className="bg-white rounded-2xl shadow-sm p-6 border border-gray-100">
                    <h2 className="text-lg font-bold text-gray-900 mb-6 flex items-center gap-2">
                        <span className="w-2 h-6 bg-green-500 rounded-full"></span>
                        Income vs Expense
                    </h2>
                    <div className="h-64 flex items-center justify-center">
                        {comparison.length > 0 ? (
                            <IncomeVsExpense data={comparison}/>
                        ) : (
                            <p className="text-gray-400 text-sm">No comparison data available</p>
                        )}
                    </div>
                </section>
            </div>

            <section className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
                <div className="p-6 border-b border-gray-50 flex items-center justify-between">
                    <h2 className="text-lg font-bold text-gray-900">Recent Transactions</h2>
                    <Link to="/transactions" className="text-blue-600 text-sm font-bold hover:underline cursor-pointer">View All</Link>
                </div>
                {recentTxns.length === 0 ? (
                    <div className="p-12 text-center">
                        <div className="text-4xl mb-4">📝</div>
                        <p className="text-gray-400 text-sm">No recent transactions found</p>
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                                <tr className="text-left text-gray-400 border-b border-gray-50 uppercase text-[10px] tracking-widest font-bold">
                                    <th className="px-6 py-4">Type</th>
                                    <th className="px-6 py-4">Category</th>
                                    <th className="px-6 py-4">Amount</th>
                                    <th className="px-6 py-4">Date</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-50">
                                {recentTxns.map(t => (
                                    <tr key={t.id} className="hover:bg-gray-50 transition-colors group">
                                        <td className="px-6 py-4">
                                            <span className={`px-2 py-1 rounded-md text-[10px] font-bold uppercase ${
                                                t.type === 'Income' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                                            }`}>
                                                {t.type}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 font-medium text-gray-700">
                                            {subcatMap.get(t.subcategoryId) || t.subcategoryId}
                                        </td>
                                        <td className={`px-6 py-4 font-bold whitespace-nowrap ${t.type === 'Income' ? 'text-green-600' : 'text-red-600'}`}>
                                            {t.type === 'Income' ? '+' : '-'}{fmt(t.amount)}
                                        </td>
                                        <td className="px-6 py-4 text-gray-400 font-medium">
                                            {new Date(t.timestamp * 1000).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </section>

            {showReportModal && (
                <div className="fixed inset-0 bg-gray-900/60 backdrop-blur-sm flex items-center justify-center z-50 p-4"
                     onClick={() => setShowReportModal(false)}>
                    <div className="bg-white rounded-2xl shadow-2xl w-full max-w-sm overflow-hidden animate-in fade-in zoom-in duration-200"
                         onClick={e => e.stopPropagation()}>
                        <div className="bg-blue-600 p-6 text-white text-center">
                            <div className="w-12 h-12 bg-white/20 rounded-full flex items-center justify-center mx-auto mb-3">
                                <FileText size={24}/>
                            </div>
                            <h2 className="text-xl font-bold">Generate Statement</h2>
                            <p className="text-blue-100 text-xs mt-1">Select a timeframe for your PDF statement</p>
                        </div>
                        <div className="p-6 space-y-3">
                            <ReportOption label="This Month" duration="this_month" onSelect={handleDownloadReport}
                                          disabled={isDownloading}/>
                            <ReportOption label="Last 30 Days" duration="one_month" onSelect={handleDownloadReport}
                                          disabled={isDownloading}/>
                            <ReportOption label="Last 7 Days" duration="one_week" onSelect={handleDownloadReport}
                                          disabled={isDownloading}/>
                            <ReportOption label="Last 6 Months" duration="half_year" onSelect={handleDownloadReport}
                                          disabled={isDownloading}/>
                            <ReportOption label="This Year" duration="this_year" onSelect={handleDownloadReport}
                                          disabled={isDownloading}/>
                            <button
                                onClick={() => setShowReportModal(false)}
                                className="w-full mt-4 py-2 text-sm font-bold text-gray-400 hover:text-gray-600 transition-colors cursor-pointer"
                            >
                                Close
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}

function Card({label, value, color, icon}: { label: string; value: string; color: string, icon: string }) {
    return (
        <div className="bg-white rounded-2xl shadow-sm p-6 border border-gray-100 flex flex-col justify-between hover:border-blue-100 transition-colors group">
            <div className="flex items-center justify-between mb-4">
                <p className="text-xs text-gray-400 font-bold uppercase tracking-wider">{label}</p>
                <span className="text-xl group-hover:scale-110 transition-transform">{icon}</span>
            </div>
            <p className={`text-2xl font-bold ${color}`}>{value}</p>
        </div>
    )
}

function ReportOption({label, duration, onSelect, disabled}: {
    label: string;
    duration: string;
    onSelect: (d: string) => void;
    disabled: boolean
}) {
    return (
        <button
            onClick={() => onSelect(duration)}
            disabled={disabled}
            className="w-full flex items-center justify-between p-4 rounded-xl border border-gray-100 hover:bg-blue-50 hover:border-blue-200 transition-all text-sm font-bold group cursor-pointer"
        >
            <span>{label}</span>
            <Download size={16} className="text-gray-400 group-hover:text-blue-500"/>
        </button>
    )
}
