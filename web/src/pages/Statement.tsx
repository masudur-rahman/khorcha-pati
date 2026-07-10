import { useEffect, useState } from 'react'
import { fetchReportData } from '../api/endpoints'
import type { StatementReport, StatementTransaction, FieldCost } from '../types'
import { fmt } from '../lib/formatter'

const ACCENT = '#0052CC'

function fmtDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('en-GB', { day: '2-digit', month: '2-digit' })
}
function fmtDateRange(start: string, end: string) {
  const opts: Intl.DateTimeFormatOptions = { month: 'long', day: 'numeric', year: 'numeric' }
  return `${new Date(start).toLocaleDateString('en-US', opts)} — ${new Date(end).toLocaleDateString('en-US', opts)}`
}
function fmtFooterTime(serverTime?: string) {
  const d = serverTime ? new Date(serverTime) : new Date()
  return d.toLocaleString('en-US', { year: 'numeric', month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit', hour12: true })
}
function fmtShortDate(dateStr: string) {
  const d = new Date(dateStr)
  const dd = String(d.getDate()).padStart(2, '0')
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const yy = String(d.getFullYear()).slice(-2)
  return `${dd}-${mm}-${yy}`
}

const typeBg: Record<string, string> = { Income: '#E3FCEF', Transfer: '#DEEBFF', Expense: '#FFEBE6' }
const typeText: Record<string, string> = { Income: '#00875A', Transfer: '#0052CC', Expense: '#DE350B' }
const amountColor: Record<string, string> = { Income: '#00875A', Transfer: '#0052CC', Expense: '#DE350B' }

function TypeBadge({ type }: { type: string }) {
  return (
    <span style={{ padding: '2px 8px', borderRadius: 4, fontSize: 9, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.04em', background: typeBg[type] ?? '#F4F5F7', color: typeText[type] ?? '#505F79', whiteSpace: 'nowrap' }}>
      {type}
    </span>
  )
}

const th: React.CSSProperties = { padding: '8px 10px', textAlign: 'left', fontWeight: 700, fontSize: 10, letterSpacing: '0.02em' }
const td: React.CSSProperties = { padding: '6px 10px', fontSize: 10 }
const SMALL_TABLE_WIDTH = '55%'
const SMALL_TABLE_MIN_WIDTH = '280px'

function RunningBalanceTable({ transactions }: { transactions: StatementTransaction[] }) {
  const sorted = [...transactions].sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
  let balance = 0
  const rows = sorted.map(t => {
    if (typeof t.runningBalance === 'number') {
      return { ...t, runningBalance: t.runningBalance }
    }
    if (t.type === 'Income') balance += t.amount
    else if (t.type === 'Expense') balance -= t.amount
    return { ...t, runningBalance: balance }
  })
  const totalAmount = sorted.reduce((sum, t) => sum + t.amount, 0)

  return (
    <table style={{ width: '100%', tableLayout: 'fixed', borderCollapse: 'collapse', fontSize: 10 }}>
      <colgroup>
        <col style={{ width: '5%' }} /><col style={{ width: '9%' }} /><col style={{ width: '11%' }} />
        <col style={{ width: '10%' }} /><col style={{ width: '10%' }} /><col style={{ width: '10%' }} />
        <col style={{ width: '10%' }} /><col style={{ width: '10%' }} /><col style={{ width: '11%' }} />
        <col style={{ width: '14%' }} />
      </colgroup>
      <thead>
        <tr style={{ background: ACCENT, color: 'white' }}>
          <th style={th}>Date</th><th style={th}>Type</th><th style={{ ...th, textAlign: 'right' }}>Amount</th>
          <th style={th}>Source</th><th style={th}>Dest</th><th style={th}>Person</th>
          <th style={th}>Category</th><th style={th}>Subcategory</th><th style={{ ...th, textAlign: 'right' }}>Balance</th>
          <th style={th}>Remarks</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((t, i) => (
          <tr key={i} style={{ background: i % 2 === 0 ? '#F4F5F7' : 'white', borderBottom: '1px solid #DFE1E6' }}>
            <td style={{ ...td, whiteSpace: 'nowrap' }}>{fmtDate(t.date)}</td>
            <td style={td}><TypeBadge type={t.type} /></td>
            <td style={{ ...td, textAlign: 'right', fontWeight: 600, color: amountColor[t.type] ?? '#505F79', whiteSpace: 'nowrap' }}>{fmt(t.amount)}</td>
            <td style={{ ...td, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.source}</td>
            <td style={{ ...td, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.destination}</td>
            <td style={{ ...td, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.person}</td>
            <td style={{ ...td, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.category}</td>
            <td style={{ ...td, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.subcategory}</td>
            <td style={{ ...td, textAlign: 'right', fontWeight: 600, color: t.runningBalance >= 0 ? ACCENT : '#DE350B', whiteSpace: 'nowrap' }}>{fmt(t.runningBalance)}</td>
            <td style={{ ...td, color: '#6B778C', overflow: 'hidden', textOverflow: 'ellipsis' }}>{t.remarks}</td>
          </tr>
        ))}
        <tr style={{ borderTop: `2px solid ${ACCENT}`, background: '#DEEBFF' }}>
          <td style={{ ...td, fontWeight: 700 }} colSpan={2}>Total</td>
          <td style={{ ...td, textAlign: 'right', fontWeight: 700, color: ACCENT, whiteSpace: 'nowrap' }}>{fmt(totalAmount)}</td>
          <td colSpan={7}></td>
        </tr>
      </tbody>
    </table>
  )
}

function SummaryTable({ items, title, showType = false }: { items: FieldCost[], title: string, showType?: boolean }) {
  if (!items || items.length === 0) return null
  return (
    <div>
      <h3 style={{ fontSize: 11, fontWeight: 700, color: '#505F79', textTransform: 'uppercase', letterSpacing: '0.06em', margin: '0 0 6px', breakAfter: 'avoid' }}>{title}</h3>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11 }}>
        <thead>
          <tr style={{ background: ACCENT, color: 'white' }}>
            <th style={th}>Name</th>
            {showType && <th style={th}>Type</th>}
            <th style={{ ...th, textAlign: 'right' }}>Amount</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item, i) => (
            <tr key={i} style={{ borderBottom: '1px solid #F4F5F7', background: i % 2 === 0 ? '#FAFBFC' : 'white' }}>
              <td style={td}>{item.name}</td>
              {showType && <td style={td}><TypeBadge type={item.type ?? ''} /></td>}
              <td style={{ ...td, textAlign: 'right', fontWeight: 600, color: amountColor[item.type ?? ''] ?? '#505F79', whiteSpace: 'nowrap' }}>{fmt(item.amount)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

const printStyles = `
  @media print {
    body { margin: 0; background: white !important; font-size: 85%; }
    #print-controls { display: none !important; }
    #statement-root { padding: 0 !important; }
    .page-content { padding: 0 !important; box-shadow: none !important; max-width: none !important; }
    .print-layout { width: 100%; border-collapse: collapse; }
    .print-layout > thead > tr > td, .print-layout > tbody > tr > td { padding: 0; vertical-align: top; }
    tr { break-inside: avoid; page-break-inside: avoid; }
  }
  @page { size: A4 portrait; margin: 6mm 6mm 12mm; }
  @page {
    @bottom-right { content: counter(page) " / " counter(pages); font-size: 8px; color: #6B778C; font-family: var(--font-body); }
  }
  * { -webkit-print-color-adjust: exact !important; print-color-adjust: exact !important; }
  body { font-family: var(--font-body); }
`

export default function Statement() {
  const [report, setReport] = useState<StatementReport | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const duration = params.get('duration') || undefined
    const start = params.get('start') || undefined
    const end = params.get('end') || undefined

    fetchReportData(duration, start, end).then(data => {
      setReport(data)
      document.title = `Khorcha-Pati Statement (${fmtShortDate(data.startDate)} — ${fmtShortDate(data.endDate)})`
    }).catch(err => setError('Failed to load statement: ' + err))
  }, [])

  if (error) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', flexDirection: 'column', gap: 16 }}>
        <p style={{ color: '#DE350B', fontSize: 16 }}>{error}</p>
        <button onClick={() => window.close()} style={{ padding: '8px 16px', background: ACCENT, color: 'white', border: 'none', borderRadius: 8, cursor: 'pointer', fontFamily: 'inherit' }}>Close</button>
      </div>
    )
  }
  if (!report) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh' }}>
        <p style={{ color: '#6B778C' }}>Loading statement...</p>
      </div>
    )
  }

  const netColor = report.netBalance >= 0 ? '#00875A' : '#DE350B'

  return (
    <div id="statement-root" style={{ background: '#F4F5F7', minHeight: '100vh' }}>
      <style>{printStyles}</style>

      <div id="print-controls" style={{
        position: 'fixed', top: 0, left: 0, right: 0, zIndex: 100,
        background: ACCENT, padding: '10px 24px',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        boxShadow: '0 2px 8px rgba(0,0,0,0.2)',
      }}>
        <span style={{ color: 'white', fontWeight: 700, fontSize: 13, fontFamily: "var(--font-display)" }}>Khorcha-Pati — Statement Preview</span>
        <div style={{ display: 'flex', gap: 10 }}>
          <button onClick={() => window.print()} style={{ padding: '8px 20px', background: 'white', color: ACCENT, border: 'none', borderRadius: 8, fontWeight: 700, cursor: 'pointer', fontSize: 12, fontFamily: 'inherit' }}>Save as PDF / Print</button>
          <button onClick={() => window.close()} style={{ padding: '8px 16px', background: 'transparent', color: 'white', border: '1px solid rgba(255,255,255,0.4)', borderRadius: 8, cursor: 'pointer', fontSize: 12, fontFamily: 'inherit' }}>Close</button>
        </div>
      </div>

      <div className="page-content" style={{ maxWidth: 900, margin: '0 auto', padding: '70px 20px 40px', background: 'white', minHeight: '100vh', boxShadow: '0 0 30px rgba(0,0,0,0.06)' }}>
        <table className="print-layout" style={{ width: '100%', borderCollapse: 'collapse', border: 'none' }}>
          <thead>
            <tr><td style={{ padding: 0 }}>
              <div className="statement-header" style={{ textAlign: 'center', marginBottom: 28, borderBottom: `3px solid ${ACCENT}`, paddingBottom: 20 }}>
                <h1 style={{ fontSize: 22, fontWeight: 800, color: ACCENT, margin: '0 0 4px', fontFamily: "var(--font-display)" }}>Khorcha-Pati Statement</h1>
                <p style={{ fontSize: 13, color: '#505F79', margin: '0 0 2px', fontWeight: 600 }}>{report.name}</p>
                <p style={{ fontSize: 11, color: '#6B778C', margin: 0 }}>{fmtDateRange(report.startDate, report.endDate)}</p>
              </div>
            </td></tr>
          </thead>
          <tbody>
            <tr><td style={{ padding: 0 }}>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12, marginBottom: 28 }}>
          <StatCard label="Total Amount" value={fmt(report.totalAmount)} bg="#E3FCEF" border="#ABF5D1" color="#00875A" />
          <StatCard label="Net Balance" value={fmt(report.netBalance)} bg="#FFEBE6" border="#FFBDAD" color={netColor} />
          <StatCard label="Transactions" value={String(report.transactions.length)} bg="#DEEBFF" border="#B3D4FF" color="#0052CC" />
        </div>

        <section style={{ marginBottom: 36 }}>
          <SectionTitle>Transaction Details</SectionTitle>
          <div style={{ overflowX: 'auto', borderRadius: 8, border: '1px solid #DFE1E6' }}>
            <RunningBalanceTable transactions={report.transactions} />
          </div>
        </section>

        <section style={{ marginBottom: 36 }}>
          <SectionTitle>Summary</SectionTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {(report.typeSummary ?? []).length > 0 && (
              <div style={{ border: '1px solid #DFE1E6', borderRadius: 8, padding: 12, maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
                <SummaryTable items={report.typeSummary ?? []} title="By Type" />
              </div>
            )}
            {(report.categorySummary ?? []).length > 0 && (
              <div style={{ border: '1px solid #DFE1E6', borderRadius: 8, padding: 12, maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
                <SummaryTable items={report.categorySummary ?? []} title="By Category" showType />
              </div>
            )}
            {(report.subcategorySummary ?? []).length > 0 && (
              <div style={{ border: '1px solid #DFE1E6', borderRadius: 8, padding: 12, maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
                <SummaryTable items={report.subcategorySummary ?? []} title="By Subcategory" showType />
              </div>
            )}
          </div>
        </section>

        {report.wallets && report.wallets.length > 0 && (
          <section style={{ marginBottom: 36, breakInside: 'avoid' }}>
            <SectionTitle>Wallet Balances</SectionTitle>
            <div style={{ borderRadius: 8, border: '1px solid #DFE1E6', overflow: 'hidden', maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11 }}>
                <thead><tr style={{ background: ACCENT, color: 'white' }}><th style={th}>Name</th><th style={{ ...th, textAlign: 'right' }}>Balance</th></tr></thead>
                <tbody>
                  {report.wallets.map((w, i) => (
                    <tr key={i} style={{ borderBottom: '1px solid #DFE1E6', background: i % 2 === 0 ? '#F4F5F7' : 'white' }}>
                      <td style={td}>{w.name} <span style={{ color: '#6B778C', fontSize: 9 }}>({w.shortName})</span></td>
                      <td style={{ ...td, textAlign: 'right', fontWeight: 600, color: w.balance >= 0 ? ACCENT : '#DE350B', whiteSpace: 'nowrap' }}>{fmt(w.balance)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        )}

        {report.contacts && report.contacts.length > 0 && (
          <section style={{ marginBottom: 36, breakInside: 'avoid' }}>
            <SectionTitle>Contacts</SectionTitle>
            <div style={{ borderRadius: 8, border: '1px solid #DFE1E6', overflow: 'hidden', maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11 }}>
                <thead><tr style={{ background: ACCENT, color: 'white' }}><th style={th}>Name</th><th style={{ ...th, textAlign: 'right' }}>Net Balance</th></tr></thead>
                <tbody>
                  {report.contacts.map((c, i) => (
                    <tr key={i} style={{ borderBottom: '1px solid #DFE1E6', background: i % 2 === 0 ? '#F4F5F7' : 'white' }}>
                      <td style={td}>{c.fullName || c.nickName}</td>
                      <td style={{ ...td, textAlign: 'right', fontWeight: 600, color: c.netBalance >= 0 ? ACCENT : '#DE350B', whiteSpace: 'nowrap' }}>{fmt(c.netBalance)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        )}

        <div style={{ borderTop: '1px solid #DFE1E6', paddingTop: 12, textAlign: 'center', color: '#6B778C', fontSize: 10 }}>
          Generated by Khorcha-Pati &nbsp;·&nbsp; {fmtFooterTime(report.generatedAt)}
        </div>

            </td></tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}

function StatCard({ label, value, bg, border, color }: { label: string; value: string; bg: string; border: string; color: string }) {
  return (
    <div style={{ background: bg, border: `1px solid ${border}`, borderRadius: 12, padding: 16, textAlign: 'center' }}>
      <p style={{ fontSize: 10, color: '#6B778C', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em', margin: '0 0 4px' }}>{label}</p>
      <p style={{ fontSize: 20, fontWeight: 800, color, margin: 0, fontFamily: "var(--font-display)" }}>{value}</p>
    </div>
  )
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2 style={{ fontSize: 13, fontWeight: 700, color: ACCENT, borderLeft: `4px solid ${ACCENT}`, paddingLeft: 10, margin: '0 0 12px', breakAfter: 'avoid', fontFamily: "var(--font-display)" }}>
      {children}
    </h2>
  )
}
