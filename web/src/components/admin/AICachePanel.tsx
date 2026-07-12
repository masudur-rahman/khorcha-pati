import { useEffect, useMemo, useState, useRef } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  listAICache, deleteAICache, listCategories, listSubcategories, type AICacheEntry,
} from '../../api/endpoints'
import { useSearch } from '../../context/SearchContext'
import { notify } from '../../lib/notify'
import Card from '../ui/Card'
import ConfirmDialog from '../ui/ConfirmDialog'
import Button from '../ui/Button'
import ActionButton from '../ui/ActionButton'
import { ICONS } from '../ui/Icons'
import { useViewportPageSize } from '../../hooks/useViewportPageSize'
import AICacheModal, { type SubMeta } from './AICacheModal'

const INTENT_COLOR: Record<string, string> = {
  Income: 'var(--color-success)',
  Expense: 'var(--color-danger)',
  Transfer: 'var(--color-primary)',
}

// useDebounced delays propagating a rapidly-changing value (search box) to avoid a request per keystroke.
function useDebounced<T>(value: T, delay = 300): T {
  const [v, setV] = useState(value)
  useEffect(() => {
    const t = setTimeout(() => setV(value), delay)
    return () => clearTimeout(t)
  }, [value, delay])
  return v
}

function fmtDate(unix: number): string {
  if (!unix) return '—'
  return new Date(unix * 1000).toLocaleDateString(undefined, { day: '2-digit', month: 'short', year: 'numeric' })
}

function confidenceColor(c: number): string {
  if (c >= 0.8) return 'var(--color-success)'
  if (c >= 0.5) return 'var(--color-warning)'
  return 'var(--color-danger)'
}

export default function AICachePanel() {
  const qc = useQueryClient()
  const [search, setSearch] = useState('')
  const q = useDebounced(search.trim())
  const [page, setPage] = useState(0)
  const [editing, setEditing] = useState<AICacheEntry | 'new' | null>(null)
  const [deleting, setDeleting] = useState<AICacheEntry | null>(null)
  const firstRowRef = useRef<HTMLTableRowElement>(null)
  const firstCardRef = useRef<HTMLDivElement>(null)
  const pageSize = useViewportPageSize(firstRowRef, firstCardRef)

  const { data: rawEntries, isLoading } = useQuery({ queryKey: ['aiCache', q], queryFn: () => listAICache(q) })

  const entries = useMemo(() => {
    if (!rawEntries) return undefined
    return rawEntries.map(e => ({
      ...e,
      intent: e.intent ? e.intent.charAt(0).toUpperCase() + e.intent.slice(1).toLowerCase() : e.intent,
    }))
  }, [rawEntries])

  const { searchTerm } = useSearch()
  const all = entries ?? []

  const filteredAll = useMemo(() => {
    if (!searchTerm.trim()) return all
    const term = searchTerm.toLowerCase()
    return all.filter(e =>
      (e.inputText || '').toLowerCase().includes(term) ||
      (e.intent || '').toLowerCase().includes(term) ||
      (e.subcategoryId || '').toLowerCase().includes(term)
    )
  }, [all, searchTerm])

  const totalPages = Math.max(1, Math.ceil(filteredAll.length / pageSize))
  const paged = filteredAll.slice(page * pageSize, (page + 1) * pageSize)

  useEffect(() => { setPage(0) }, [q, searchTerm])                            // reset paging on new search
  useEffect(() => { if (page >= totalPages) setPage(totalPages - 1) }, [page, totalPages]) // clamp after deletes
  const { data: subs } = useQuery({ queryKey: ['subcategories'], queryFn: () => listSubcategories() })
  const { data: cats } = useQuery({ queryKey: ['categories'], queryFn: () => listCategories() })

  const subMeta = useMemo(() => {
    const catName = new Map((cats ?? []).map(c => [c.id, c.name]))
    const m = new Map<string, SubMeta>()
    for (const s of subs ?? []) m.set(s.id, { name: s.name, catName: catName.get(s.catId) ?? s.catId, types: s.types })
    return m
  }, [subs, cats])

  const subOptions = useMemo(() => {
    const opts = [...subMeta.entries()].map(([id, s]) => ({ value: id, label: `${s.catName} › ${s.name}` }))
    opts.sort((a, b) => a.label.localeCompare(b.label))
    return opts
  }, [subMeta])

  const delMut = useMutation({
    mutationFn: (id: number) => deleteAICache(id),
    onSuccess: () => { notify.deleted('Cache entry'); qc.invalidateQueries({ queryKey: ['aiCache'] }); setDeleting(null) },
    onError: (err) => notify.error(err, 'delete cache entry'),
  })

  const onSaved = () => { qc.invalidateQueries({ queryKey: ['aiCache'] }); setEditing(null) }

  return (
    <Card padding={0} style={{ width: '100%', overflow: 'hidden' }}>
      <div className="aicache-header" style={{ padding: '16px 20px', borderBottom: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
        <div style={{ flex: 1, minWidth: 180 }}>
          <h3 style={{ fontSize: 16, fontWeight: 600, color: 'var(--color-text)' }}>AI Classification Cache</h3>
          <p className="aicache-desc" style={{ fontSize: 12, color: 'var(--color-text-tertiary)', marginTop: 2 }}>
            Text → category mappings the classifier reuses. Edits apply live, no restart needed.
          </p>
        </div>
        <div className="aicache-actions" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <input
            type="search"
            name="ai-cache-search"
            autoComplete="off"
            autoCorrect="off"
            autoCapitalize="off"
            spellCheck={false}
            data-lpignore="true"
            data-1p-ignore="true"
            data-form-type="other"
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="Search input text…"
            style={{
              background: 'var(--color-bg)', border: '1px solid var(--color-border)', borderRadius: 10,
              padding: '9px 14px', fontSize: 13, color: 'var(--color-text-primary)', outline: 'none', minWidth: 160,
            }}
          />
          <Button icon={<PlusIcon />} onClick={() => setEditing('new')}>Add entry</Button>
        </div>
      </div>

      {/* Desktop: zebra table */}
      <div className="zebra-table-wrap hidden md:block">
        <table className="zebra-table" style={{ width: '100%', minWidth: 820, tableLayout: 'fixed', borderCollapse: 'collapse', fontSize: 14 }}>
          <thead>
            <tr>
              {[
                { h: 'Input Text', w: undefined },
                { h: 'Category', w: undefined },
                { h: 'Intent', w: 120 },
                { h: 'Confidence', w: 150 },
                { h: 'Added', w: 110 },
                { h: '', w: 96 },
              ].map(({ h, w }, i) => (
                <th key={h || i} style={{
                  width: w, padding: '12px 16px', textAlign: i === 5 ? 'right' : 'left', fontWeight: 600,
                  color: 'var(--color-text-secondary)', fontSize: 12, textTransform: 'uppercase', letterSpacing: '0.05em',
                }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {paged.map((e, i) => {
              const meta = subMeta.get(e.subcategoryId)
              return (
                <tr key={e.id} ref={i === 0 ? firstRowRef : undefined} style={{ borderBottom: '1px solid var(--color-border)' }}>
                  <td style={{ padding: '12px 16px', fontWeight: 500, color: 'var(--color-text-primary)', maxWidth: 320 }}>
                    <div style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }} title={e.inputText}>{e.inputText}</div>
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <div style={{ color: 'var(--color-text-primary)' }}>{meta ? `${meta.catName} › ${meta.name}` : e.subcategoryId}</div>
                    <div style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontFamily: 'var(--font-mono, monospace)' }}>{e.subcategoryId}</div>
                  </td>
                  <td style={{ padding: '12px 16px' }}><IntentPill intent={e.intent} /></td>
                  <td style={{ padding: '12px 16px' }}><ConfidenceBar value={e.confidence} /></td>
                  <td style={{ padding: '12px 16px', color: 'var(--color-text-tertiary)', fontSize: 13, whiteSpace: 'nowrap' }}>{fmtDate(e.createdAt)}</td>
                  <td style={{ padding: '12px 16px', textAlign: 'right', whiteSpace: 'nowrap' }}>
                    <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                      <ActionButton actionType="edit" icon={ICONS.edit(14)} onClick={() => setEditing(e)} title="Edit" />
                      <ActionButton actionType="delete" icon={ICONS.trash(14)} onClick={() => setDeleting(e)} title="Delete" />
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {/* Mobile: stacked cards */}
      <div className="txn-card-list flex flex-col md:hidden">
        {paged.map((e, i) => {
          const meta = subMeta.get(e.subcategoryId)
          return (
            <div key={e.id} ref={i === 0 ? firstCardRef : undefined} className="txn-card" style={{ cursor: 'default' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10 }}>
                <span style={{ fontWeight: 700, fontSize: 14, color: 'var(--color-text-primary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={e.inputText}>{e.inputText}</span>
                <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                  <ActionButton actionType="edit" icon={ICONS.edit(14)} onClick={() => setEditing(e)} title="Edit" />
                  <ActionButton actionType="delete" icon={ICONS.trash(14)} onClick={() => setDeleting(e)} title="Delete" />
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10 }}>
                <span style={{ fontSize: 13, color: 'var(--color-text-secondary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {meta ? `${meta.catName} › ${meta.name}` : e.subcategoryId}
                </span>
                <IntentPill intent={e.intent} />
              </div>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10 }}>
                <div style={{ flex: 1, minWidth: 0 }}><ConfidenceBar value={e.confidence} /></div>
                <span style={{ fontSize: 12, color: 'var(--color-text-tertiary)', whiteSpace: 'nowrap', flexShrink: 0 }}>{fmtDate(e.createdAt)}</span>
              </div>
            </div>
          )
        })}
      </div>

      {isLoading && <p style={{ padding: 24, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>Loading…</p>}
      {!isLoading && filteredAll.length === 0 && (
        <p style={{ padding: 24, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
          {q || searchTerm ? `No entries matching your search query.` : 'No cache entries yet.'}
        </p>
      )}

      {totalPages > 1 && (
        <div style={{ padding: '16px 20px', borderTop: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', fontWeight: 500 }}>
            Showing {page * pageSize + 1}–{Math.min((page + 1) * pageSize, filteredAll.length)} of {filteredAll.length}
          </p>
          <div style={{ display: 'flex', gap: 6 }}>
            <PageButton disabled={page === 0} onClick={() => setPage(p => p - 1)}>Previous</PageButton>
            <PageButton disabled={page >= totalPages - 1} onClick={() => setPage(p => p + 1)}>Next</PageButton>
          </div>
        </div>
      )}

      {editing && (
        <AICacheModal entry={editing} subMeta={subMeta} subOptions={subOptions} onClose={() => setEditing(null)} onSaved={onSaved} />
      )}
      {deleting && (
        <ConfirmDialog
          title="Delete cache entry?"
          type="danger"
          confirmText="Delete"
          onConfirm={() => delMut.mutate(deleting.id)}
          onClose={() => setDeleting(null)}
          message={
            <>Remove the mapping for <strong style={{ color: 'var(--color-text-primary)' }}>“{deleting.inputText}”</strong>? The classifier will re-learn it next time it sees that text.</>
          }
        />
      )}
    </Card>
  )
}

function PageButton({ children, disabled, onClick }: { children: React.ReactNode; disabled: boolean; onClick: () => void }) {
  return (
    <button
      disabled={disabled}
      onClick={onClick}
      style={{
        padding: '8px 14px', borderRadius: 'var(--radius-sm)', fontSize: 12, fontWeight: 600,
        border: '1px solid var(--color-border)', background: 'var(--color-surface)',
        cursor: disabled ? 'not-allowed' : 'pointer', color: 'var(--color-text-secondary)',
        opacity: disabled ? 0.5 : 1, fontFamily: 'inherit',
      }}
    >{children}</button>
  )
}

function IntentPill({ intent }: { intent: string }) {
  const color = INTENT_COLOR[intent] ?? 'var(--color-text-tertiary)'
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 6, padding: '3px 10px', borderRadius: 999,
      fontSize: 11, fontWeight: 700, background: `color-mix(in srgb, ${color} 14%, transparent)`, color,
    }}>
      <span style={{ width: 6, height: 6, borderRadius: '50%', background: color }} />{intent}
    </span>
  )
}

function ConfidenceBar({ value }: { value: number }) {
  const pct = Math.round(Math.max(0, Math.min(1, value)) * 100)
  const color = confidenceColor(value)
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <div style={{ flex: 1, height: 6, borderRadius: 999, background: 'var(--color-hover)', overflow: 'hidden', minWidth: 48 }}>
        <div style={{ width: `${pct}%`, height: '100%', background: color, borderRadius: 999 }} />
      </div>
      <span style={{ fontSize: 12, fontWeight: 600, color, minWidth: 32, textAlign: 'right' }}>{pct}%</span>
    </div>
  )
}

const iconProps = { width: 16, height: 16, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', strokeWidth: 2, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const }

function PlusIcon() { return <svg {...iconProps}><line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" /></svg> }
