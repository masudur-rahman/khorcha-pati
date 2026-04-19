import { useEffect, useState } from 'react'
import { fetchReportData } from '../api/endpoints'
import type { StatementReport, StatementTransaction, FieldCost } from '../types'

const CURRENCY = '৳'

function fmt(amount: number) {
  return `${CURRENCY}${amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function fmtDate(dateStr: string) {
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: '2-digit' })
}

function fmtDateRange(start: string, end: string) {
  const s = new Date(start)
  const e = new Date(end)
  const opts: Intl.DateTimeFormatOptions = { month: 'long', day: 'numeric', year: 'numeric' }
  return `${s.toLocaleDateString('en-US', opts)} — ${e.toLocaleDateString('en-US', opts)}`
}

function fmtFooterTime() {
  return new Date().toLocaleString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: 'numeric', minute: '2-digit', hour12: true,
  })
}

const TEAL = '#0f766e'

const typeBg: Record<string, string> = {
  Income: '#dcfce7',
  Transfer: '#dbeafe',
  Expense: '#fee2e2',
}
const typeText: Record<string, string> = {
  Income: '#15803d',
  Transfer: '#1d4ed8',
  Expense: '#b91c1c',
}
const amountColor: Record<string, string> = {
  Income: '#15803d',
  Transfer: '#1d4ed8',
  Expense: '#b91c1c',
}

function TypeBadge({ type }: { type: string }) {
  return (
    <span style={{
      padding: '2px 7px',
      borderRadius: '4px',
      fontSize: '9px',
      fontWeight: 700,
      textTransform: 'uppercase',
      letterSpacing: '0.04em',
      background: typeBg[type] ?? '#f1f5f9',
      color: typeText[type] ?? '#374151',
      whiteSpace: 'nowrap',
    }}>
      {type}
    </span>
  )
}

function RunningBalanceTable({ transactions }: { transactions: StatementTransaction[] }) {
  const sorted = [...transactions].sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
  let balance = 0
  const rows = sorted.map(t => {
    if (t.type === 'Income') balance += t.amount
    else if (t.type === 'Expense') balance -= t.amount
    return { ...t, runningBalance: balance }
  })
  const totalAmount = sorted.reduce((sum, t) => sum + t.amount, 0)

  return (
    <table style={{ width: '100%', tableLayout: 'fixed', borderCollapse: 'collapse', fontSize: '10px' }}>
      <colgroup>
        <col style={{ width: '5%' }} />   {/* Date  DD/MM       */}
        <col style={{ width: '9%' }} />   {/* Type  Transfer    */}
        <col style={{ width: '11%' }} />  {/* Amount            */}
        <col style={{ width: '11%' }} />  {/* Source            */}
        <col style={{ width: '11%' }} />  {/* Dest              */}
        <col style={{ width: '8%' }} />   {/* Person            */}
        <col style={{ width: '10%' }} />  {/* Category          */}
        <col style={{ width: '11%' }} />  {/* Subcategory       */}
        <col style={{ width: '11%' }} />  {/* Balance           */}
        <col style={{ width: '13%' }} />  {/* Remarks           */}
      </colgroup>
      <thead>
        <tr style={{ background: TEAL, color: 'white' }}>
          <th style={th}>Date</th>
          <th style={th}>Type</th>
          <th style={{ ...th, textAlign: 'right' }}>Amount</th>
          <th style={th}>Source</th>
          <th style={th}>Dest</th>
          <th style={th}>Person</th>
          <th style={th}>Category</th>
          <th style={th}>Subcategory</th>
          <th style={{ ...th, textAlign: 'right' }}>Balance</th>
          <th style={th}>Remarks</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((t, i) => (
          <tr key={i} style={{ background: i % 2 === 0 ? '#f8fafc' : 'white', borderBottom: '1px solid #e2e8f0' }}>
            <td style={{ ...td, whiteSpace: 'nowrap' }}>{fmtDate(t.date)}</td>
            <td style={td}><TypeBadge type={t.type} /></td>
            <td style={{ ...td, textAlign: 'right', fontWeight: 600, color: amountColor[t.type] ?? '#374151', whiteSpace: 'nowrap' }}>
              {fmt(t.amount)}
            </td>
            <td style={{ ...td, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{t.source}</td>
            <td style={{ ...td, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{t.destination}</td>
            <td style={{ ...td, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.person}</td>
            <td style={{ ...td, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.category}</td>
            <td style={{ ...td, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.subcategory}</td>
            <td style={{ ...td, textAlign: 'right', fontWeight: 600, color: t.runningBalance >= 0 ? TEAL : '#b91c1c', whiteSpace: 'nowrap' }}>
              {fmt(t.runningBalance)}
            </td>
            <td style={{ ...td, color: '#6b7280', overflow: 'hidden', textOverflow: 'ellipsis' }}>{t.remarks}</td>
          </tr>
        ))}
        <tr style={{ borderTop: `2px solid ${TEAL}`, background: '#f0fdf4' }}>
          <td style={{ ...td, fontWeight: 700 }} colSpan={2}>Total</td>
          <td style={{ ...td, textAlign: 'right', fontWeight: 700, color: TEAL, whiteSpace: 'nowrap' }}>{fmt(totalAmount)}</td>
          <td colSpan={7} />
        </tr>
      </tbody>
    </table>
  )
}

function SummaryTable({ items, title, showType = false }: { items: FieldCost[], title: string, showType?: boolean }) {
  if (!items || items.length === 0) return null
  return (
    <div>
      <h3 style={{ fontSize: '11px', fontWeight: 700, color: '#374151', textTransform: 'uppercase', letterSpacing: '0.06em', margin: '0 0 6px', breakAfter: 'avoid', pageBreakAfter: 'avoid' }}>{title}</h3>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '11px' }}>
        <thead>
          <tr style={{ background: '#f1f5f9' }}>
            <th style={{ ...th, color: '#374151', fontWeight: 700 }}>Name</th>
            {showType && <th style={{ ...th, color: '#374151', fontWeight: 700 }}>Type</th>}
            <th style={{ ...th, textAlign: 'right', color: '#374151', fontWeight: 700 }}>Amount</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item, i) => (
            <tr key={i} style={{ borderBottom: '1px solid #f1f5f9', background: i % 2 === 0 ? '#fafafa' : 'white' }}>
              <td style={td}>{item.name}</td>
              {showType && <td style={td}><TypeBadge type={item.type ?? ''} /></td>}
              <td style={{
                ...td, textAlign: 'right', fontWeight: 600,
                color: item.type === 'Income' ? '#15803d' : item.type === 'Transfer' ? '#1d4ed8' : '#b91c1c',
                whiteSpace: 'nowrap',
              }}>
                {fmt(item.amount)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

const th: React.CSSProperties = { padding: '7px 8px', textAlign: 'left', fontWeight: 700, fontSize: '10px', letterSpacing: '0.02em' }
const td: React.CSSProperties = { padding: '5px 8px', fontSize: '10px' }

// In print: .statement-header becomes position:fixed at top of every page.
// @page margin-top (30mm ≈ 113px) must be >= header height so content on
// every page starts below the fixed header — no padding-top trick needed.
const printStyles = `
  @media print {
    body { margin: 0; background: white !important; }
    #print-controls { display: none !important; }
    #statement-root { padding: 0 !important; }
    .page-content { padding: 80px 0 0 !important; box-shadow: none !important; }
    thead { display: table-row-group; }
    .statement-header {
      position: fixed;
      top: 0; left: 0; right: 0;
      background: white;
      z-index: 999;
      margin: 0 !important;
      padding-bottom: 14px !important;
    }
    .screen-footer { display: none !important; }
    tr { break-inside: avoid; page-break-inside: avoid; }
  }
  @page {
    size: A4 portrait;
    margin: 10mm 6mm 18mm;
  }
  @page {
    @bottom-right {
      content: "Page " counter(page) " of " counter(pages);
      font-size: 9px;
      color: #9ca3af;
      font-family: 'Segoe UI', 'Helvetica Neue', Arial, sans-serif;
    }
    @bottom-left {
      content: "Generated by Expense Tracker Bot";
      font-size: 9px;
      color: #9ca3af;
      font-family: 'Segoe UI', 'Helvetica Neue', Arial, sans-serif;
    }
  }
  * { -webkit-print-color-adjust: exact !important; print-color-adjust: exact !important; color-adjust: exact !important; }
  body { font-family: 'Segoe UI', 'Helvetica Neue', Arial, sans-serif; }
`

// Width applied to all single-section tables (summary, wallets, contacts)
const SMALL_TABLE_WIDTH = '55%'
const SMALL_TABLE_MIN_WIDTH = '280px'

export default function Statement() {
  const [report, setReport] = useState<StatementReport | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const duration = params.get('duration') || 'this_month'
    fetchReportData(duration)
      .then(data => setReport(data))
      .catch(err => setError('Failed to load statement: ' + err))
  }, [])

  if (error) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', flexDirection: 'column', gap: '16px' }}>
        <p style={{ color: '#b91c1c', fontSize: '16px' }}>{error}</p>
        <button onClick={() => window.close()} style={{ padding: '8px 16px', background: TEAL, color: 'white', border: 'none', borderRadius: '8px', cursor: 'pointer' }}>Close</button>
      </div>
    )
  }

  if (!report) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh' }}>
        <p style={{ color: '#6b7280' }}>Loading statement...</p>
      </div>
    )
  }

  const netColor = report.netBalance >= 0 ? '#15803d' : '#b91c1c'

  return (
    <div id="statement-root" style={{ background: '#f1f5f9', minHeight: '100vh' }}>
      <style>{printStyles}</style>

      {/* Print controls — hidden on print */}
      <div id="print-controls" style={{
        position: 'fixed', top: 0, left: 0, right: 0, zIndex: 100,
        background: TEAL, padding: '10px 20px',
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        boxShadow: '0 2px 8px rgba(0,0,0,0.2)',
      }}>
        <span style={{ color: 'white', fontWeight: 700, fontSize: '13px' }}>Expense Tracker — Statement Preview</span>
        <div style={{ display: 'flex', gap: '10px' }}>
          <button
            onClick={() => window.print()}
            style={{ padding: '7px 18px', background: 'white', color: TEAL, border: 'none', borderRadius: '6px', fontWeight: 700, cursor: 'pointer', fontSize: '12px' }}
          >
            Save as PDF / Print
          </button>
          <button
            onClick={() => window.close()}
            style={{ padding: '7px 14px', background: 'transparent', color: 'white', border: '1px solid rgba(255,255,255,0.45)', borderRadius: '6px', cursor: 'pointer', fontSize: '12px' }}
          >
            Close
          </button>
        </div>
      </div>

      {/* Statement content */}
      <div className="page-content" style={{ maxWidth: '900px', margin: '0 auto', padding: '70px 20px 40px', background: 'white', minHeight: '100vh', boxShadow: '0 0 30px rgba(0,0,0,0.08)' }}>

        {/* Header — becomes position:fixed in print, repeats on every page.
            @page margin-top: 30mm ensures content on every page starts below this header. */}
        <div className="statement-header" style={{ textAlign: 'center', marginBottom: '28px', borderBottom: `3px solid ${TEAL}`, paddingBottom: '20px' }}>
          <h1 style={{ fontSize: '20px', fontWeight: 800, color: TEAL, margin: '0 0 4px' }}>Expense Tracker Statement</h1>
          <p style={{ fontSize: '13px', color: '#374151', margin: '0 0 2px', fontWeight: 600 }}>{report.name}</p>
          <p style={{ fontSize: '11px', color: '#6b7280', margin: 0 }}>{fmtDateRange(report.startDate, report.endDate)}</p>
        </div>

        {/* Overview cards */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px', marginBottom: '28px' }}>
          <StatCard label="Total Amount" value={fmt(report.totalAmount)} bg="#f0fdf4" border="#bbf7d0" color="#15803d" />
          <StatCard label="Net Balance" value={fmt(report.netBalance)} bg="#fef2f2" border="#fecaca" color={netColor} />
          <StatCard label="Transactions" value={String(report.transactions.length)} bg="#eff6ff" border="#bfdbfe" color="#1d4ed8" />
        </div>

        {/* Transaction table */}
        <section style={{ marginBottom: '36px' }}>
          <SectionTitle>Transaction Details</SectionTitle>
          <div style={{ overflowX: 'auto', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
            <RunningBalanceTable transactions={report.transactions} />
          </div>
        </section>

        {/* Summary — stacked, all at the same width */}
        <section style={{ marginBottom: '36px' }}>
          <SectionTitle>Summary</SectionTitle>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {(report.typeSummary ?? []).length > 0 && (
              <div style={{ border: '1px solid #e2e8f0', borderRadius: '8px', padding: '12px', maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
                <SummaryTable items={report.typeSummary ?? []} title="By Type" />
              </div>
            )}
            {(report.categorySummary ?? []).length > 0 && (
              <div style={{ border: '1px solid #e2e8f0', borderRadius: '8px', padding: '12px', maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
                <SummaryTable items={report.categorySummary ?? []} title="By Category" showType />
              </div>
            )}
            {(report.subcategorySummary ?? []).length > 0 && (
              <div style={{ border: '1px solid #e2e8f0', borderRadius: '8px', padding: '12px', maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
                <SummaryTable items={report.subcategorySummary ?? []} title="By Subcategory" showType />
              </div>
            )}
          </div>
        </section>

        {/* Wallet Balances */}
        {report.wallets && report.wallets.length > 0 && (
          <section style={{ marginBottom: '36px', breakInside: 'avoid', pageBreakInside: 'avoid' }}>
            <SectionTitle>Wallet Balances</SectionTitle>
            <div style={{ borderRadius: '8px', border: '1px solid #e2e8f0', overflow: 'hidden', maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '11px' }}>
                <thead>
                  <tr style={{ background: TEAL, color: 'white' }}>
                    <th style={th}>Name</th>
                    <th style={{ ...th, textAlign: 'right' }}>Balance</th>
                  </tr>
                </thead>
                <tbody>
                  {report.wallets.map((w, i) => (
                    <tr key={i} style={{ borderBottom: '1px solid #e2e8f0', background: i % 2 === 0 ? '#f8fafc' : 'white' }}>
                      <td style={td}>{w.name} <span style={{ color: '#9ca3af', fontSize: '9px' }}>({w.shortName})</span></td>
                      <td style={{ ...td, textAlign: 'right', fontWeight: 600, color: w.balance >= 0 ? TEAL : '#b91c1c', whiteSpace: 'nowrap' }}>{fmt(w.balance)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        )}

        {/* Contacts */}
        {report.contacts && report.contacts.length > 0 && (
          <section style={{ marginBottom: '36px', breakInside: 'avoid', pageBreakInside: 'avoid' }}>
            <SectionTitle>Contacts</SectionTitle>
            <div style={{ borderRadius: '8px', border: '1px solid #e2e8f0', overflow: 'hidden', maxWidth: SMALL_TABLE_WIDTH, minWidth: SMALL_TABLE_MIN_WIDTH }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '11px' }}>
                <thead>
                  <tr style={{ background: TEAL, color: 'white' }}>
                    <th style={th}>Name</th>
                    <th style={{ ...th, textAlign: 'right' }}>Net Balance</th>
                  </tr>
                </thead>
                <tbody>
                  {report.contacts.map((c, i) => (
                    <tr key={i} style={{ borderBottom: '1px solid #e2e8f0', background: i % 2 === 0 ? '#f8fafc' : 'white' }}>
                      <td style={td}>{c.fullName || c.nickName}</td>
                      <td style={{ ...td, textAlign: 'right', fontWeight: 600, color: c.netBalance >= 0 ? TEAL : '#b91c1c', whiteSpace: 'nowrap' }}>{fmt(c.netBalance)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        )}

        {/* Screen-only footer */}
        <div className="screen-footer" style={{ borderTop: '1px solid #e2e8f0', paddingTop: '12px', textAlign: 'center', color: '#9ca3af', fontSize: '10px' }}>
          Generated by Expense Tracker Bot &nbsp;·&nbsp; {fmtFooterTime()}
        </div>
      </div>
    </div>
  )
}

function StatCard({ label, value, bg, border, color }: { label: string; value: string; bg: string; border: string; color: string }) {
  return (
    <div style={{ background: bg, border: `1px solid ${border}`, borderRadius: '10px', padding: '14px', textAlign: 'center' }}>
      <p style={{ fontSize: '10px', color: '#6b7280', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em', margin: '0 0 4px' }}>{label}</p>
      <p style={{ fontSize: '18px', fontWeight: 800, color, margin: 0 }}>{value}</p>
    </div>
  )
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <h2 style={{ fontSize: '13px', fontWeight: 700, color: TEAL, borderLeft: `4px solid ${TEAL}`, paddingLeft: '10px', margin: '0 0 10px', breakAfter: 'avoid', pageBreakAfter: 'avoid' }}>
      {children}
    </h2>
  )
}
