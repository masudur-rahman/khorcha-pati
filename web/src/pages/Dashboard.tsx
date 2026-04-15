import {useMemo, useState} from 'react'
import {useQuery} from '@tanstack/react-query'
import {downloadReport, getChartData, listCategories, listSubcategories} from '../api/endpoints'
import ExpenseDonut from '../components/charts/ExpenseDonut'
import IncomeVsExpense from '../components/charts/IncomeVsExpense'
import BudgetGauge from '../components/charts/BudgetGauge'
import {useTransactions} from '../hooks/useTransactions'
import {Download, FileText} from 'lucide-react'

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
    const recentTxns = txns.slice(0, 10)

    const handleDownloadReport = async (duration: string) => {
        try {
            setIsDownloading(true)
            const blob = await downloadReport(duration)
            const url = window.URL.createObjectURL(blob as any)
            const a = document.createElement('a')
            a.href = url
            a.download = `financial_transaction_report_${duration}_${new Date().toISOString().split('T')[0]}.pdf`
            document.body.appendChild(a)
            a.click()
            window.URL.revokeObjectURL(url)
            document.body.removeChild(a)
            setShowReportModal(false)
        } catch (err) {
            alert('Failed to download report: ' + err)
        } finally {
            setIsDownloading(false)
        }
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold">Dashboard</h1>
                <button
                    onClick={() => setShowReportModal(true)}
                    className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors shadow-sm"
                >
                    <Download size={18}/>
                    Download Report
                </button>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <Card label="Total Balance" value={fmt(overview.totalBalance)} color="text-blue-600"/>
                <Card label="Month Income" value={fmt(overview.monthIncome)} color="text-green-600"/>
                <Card label="Month Expense" value={fmt(overview.monthExpense)} color="text-red-600"/>
                <div className="bg-white rounded-lg shadow p-4">
                    <p className="text-xs text-gray-500 mb-1">Budget Usage</p>
                    <BudgetGauge percent={overview.budgetUsage}/>
                    <p className="text-sm font-semibold mt-1">{overview.budgetUsage.toFixed(0)}%</p>
                </div>
            </div>

            <div className="grid md:grid-cols-2 gap-6">
                <div className="bg-white rounded-lg shadow p-4">
                    <h2 className="text-sm font-semibold mb-2">Expense by Category</h2>
                    <ExpenseDonut data={chartCategories}/>
                </div>
                <div className="bg-white rounded-lg shadow p-4">
                    <h2 className="text-sm font-semibold mb-2">Income vs Expense</h2>
                    <IncomeVsExpense data={comparison ?? []}/>
                </div>
            </div>

            <div className="bg-white rounded-lg shadow p-4">
                <h2 className="text-sm font-semibold mb-3">Recent Transactions</h2>
                {recentTxns.length === 0 ? (
                    <p className="text-gray-400 text-sm">No recent transactions</p>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                            <tr className="text-left text-gray-500 border-b">
                                <th className="pb-2">Type</th>
                                <th className="pb-2">Amount</th>
                                <th className="pb-2">Sub Category</th>
                                <th className="pb-2">Date</th>
                            </tr>
                            </thead>
                            <tbody>
                            {recentTxns.map(t => (
                                <tr key={t.id} className="border-b last:border-0">
                                    <td className="py-2">{t.type}</td>
                                    <td className={`py-2 ${t.type === 'Income' ? 'text-green-600' : 'text-red-600'}`}>
                                        {fmt(t.amount)}
                                    </td>
                                    <td className="py-2">{subcatMap.get(t.subcategoryId) || t.subcategoryId}</td>
                                    <td className="py-2 text-gray-500">{new Date(t.timestamp * 1000).toLocaleDateString()}</td>
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>

            {showReportModal && (
                <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4"
                     onClick={() => setShowReportModal(false)}>
                    <div className="bg-white rounded-xl shadow-xl w-full max-w-sm overflow-hidden"
                         onClick={e => e.stopPropagation()}>
                        <div className="bg-blue-600 p-4 text-white flex items-center gap-3">
                            <FileText size={24}/>
                            <h2 className="text-lg font-bold">Generate PDF Report</h2>
                        </div>
                        <div className="p-4 space-y-2">
                            <p className="text-sm text-gray-500 mb-4">Select the duration for your transaction
                                report.</p>
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
                                className="w-full mt-4 py-2 text-sm text-gray-500 hover:text-gray-700"
                            >
                                Cancel
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}

function Card({label, value, color}: { label: string; value: string; color: string }) {
    return (
        <div className="bg-white rounded-lg shadow p-4 border-l-4 border-blue-500">
            <p className="text-xs text-gray-500 font-medium">{label}</p>
            <p className={`text-xl font-bold mt-1 ${color}`}>{value}</p>
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
            className="w-full flex items-center justify-between p-3 rounded-lg border border-gray-100 hover:bg-blue-50 hover:border-blue-200 transition-all text-sm font-medium group"
        >
            <span>{label}</span>
            <Download size={16} className="text-gray-400 group-hover:text-blue-500"/>
        </button>
    )
}

function fmt(n: number) {
    return (n || 0).toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})
}
