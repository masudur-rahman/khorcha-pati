import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  listAICache, deleteAICache, listCategories, listSubcategories, type AICacheEntry,
} from '../../api/endpoints'
import Card from '../ui/Card'
import Modal from '../ui/Modal'
import Button from '../ui/Button'
import AICacheModal, { type SubMeta } from './AICacheModal'

const INTENT_COLOR: Record<string, string> = {
  Income: 'var(--color-success)',
  Expense: 'var(--color-danger)',
  Transfer: 'var(--color-primary)',
}

const PAGE_SIZE = 8

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

  const { data: entries, isLoading } = useQuery({ queryKey: ['aiCache', q], queryFn: () => listAICache(q) })

  const all = entries ?? []
  const totalPages = Math.max(1, Math.ceil(all.length / PAGE_SIZE))
  const paged = all.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE)

  useEffect(() => { setPage(0) }, [q])                                        // reset paging on new search
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
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['aiCache'] }); setDeleting(null) },
  })

  const onSaved = () => { qc.invalidateQueries({ queryKey: ['aiCache'] }); setEditing(null) }

  return (
    <Card padding={0}>
      <div style={{ padding: '16px 20px', borderBottom: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
        <div style={{ flex: 1, minWidth: 200 }}>
          <h3 style={{ fontSize: 16, fontWeight: 600, color: 'var(--color-text)' }}>AI Classification Cache</h3>
          <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', marginTop: 2 }}>
            Text → category mappings the classifier reuses. Edits apply live, no restart needed.
          </p>
        </div>
        <input
          value={search}
          onChange={e => setSearch(e.target.value)}
          placeholder="Search input text…"
          style={{
            background: 'var(--color-bg)', border: '1px solid var(--color-border)', borderRadius: 10,
            padding: '9px 14px', fontSize: 13, color: 'var(--color-text-primary)', outline: 'none', minWidth: 200,
          }}
        />
        <Button icon={<PlusIcon />} onClick={() => setEditing('new')}>Add entry</Button>
      </div>

      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', tableLayout: 'fixed', borderCollapse: 'collapse', fontSize: 14 }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
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
            {paged.map(e => {
              const meta = subMeta.get(e.subcategoryId)
              return (
                <tr key={e.id} style={{ borderBottom: '1px solid var(--color-border)' }}>
                  <td style={{ padding: '12px 16px', fontWeight: 500, color: 'var(--color-text-primary)', maxWidth: 320 }}>
                    <div style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }} title={e.inputText}>{e.inputText}</div>
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <div style={{ color: 'var(--color-text-primary)' }}>{meta ? `${meta.catName} › ${meta.name}` : e.subcategoryId}</div>
                    <div style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontFamily: 'var(--font-mono, monospace)' }}>{e.subcategoryId}</div>
                  </td>
                  <td style={{ padding: '12px 16px' }}><IntentPill intent={e.intent} /></td>
                  <td style={{ padding: '12px 16px', minWidth: 120 }}><ConfidenceBar value={e.confidence} /></td>
                  <td style={{ padding: '12px 16px', color: 'var(--color-text-tertiary)', fontSize: 13, whiteSpace: 'nowrap' }}>{fmtDate(e.createdAt)}</td>
                  <td style={{ padding: '12px 16px', textAlign: 'right', whiteSpace: 'nowrap' }}>
                    <IconButton label="Edit" onClick={() => setEditing(e)}><EditIcon /></IconButton>
                    <IconButton label="Delete" danger onClick={() => setDeleting(e)}><TrashIcon /></IconButton>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {isLoading && <p style={{ padding: 24, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>Loading…</p>}
      {!isLoading && all.length === 0 && (
        <p style={{ padding: 24, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
          {q ? `No entries matching “${q}”.` : 'No cache entries yet.'}
        </p>
      )}

      {totalPages > 1 && (
        <div style={{ padding: '16px 20px', borderTop: '1px solid var(--color-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', fontWeight: 500 }}>
            Showing {page * PAGE_SIZE + 1}–{Math.min((page + 1) * PAGE_SIZE, all.length)} of {all.length}
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
        <Modal title="Delete cache entry?" onClose={() => setDeleting(null)} width={460}>
          <p style={{ margin: 0, color: 'var(--color-text-secondary)', lineHeight: 1.6 }}>
            Remove the mapping for <strong style={{ color: 'var(--color-text-primary)' }}>“{deleting.inputText}”</strong>? The classifier will re-learn it next time it sees that text.
          </p>
          <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 24 }}>
            <Button variant="secondary" onClick={() => setDeleting(null)} disabled={delMut.isPending}>Cancel</Button>
            <Button variant="danger" onClick={() => delMut.mutate(deleting.id)} disabled={delMut.isPending}>
              {delMut.isPending ? 'Deleting…' : 'Delete'}
            </Button>
          </div>
        </Modal>
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

function IconButton({ children, label, onClick, danger }: { children: React.ReactNode; label: string; onClick: () => void; danger?: boolean }) {
  return (
    <button
      onClick={onClick}
      aria-label={label}
      title={label}
      style={{
        padding: 7, marginLeft: 4, borderRadius: 8, border: '1px solid var(--color-border)', background: 'transparent',
        cursor: 'pointer', color: danger ? 'var(--color-danger)' : 'var(--color-text-tertiary)',
        transition: 'all 0.15s',
      }}
      onMouseEnter={e => { e.currentTarget.style.background = 'var(--color-hover)' }}
      onMouseLeave={e => { e.currentTarget.style.background = 'transparent' }}
    >
      {children}
    </button>
  )
}

const iconProps = { width: 15, height: 15, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', strokeWidth: 2, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const }

function PlusIcon() { return <svg {...iconProps} width={16} height={16}><line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" /></svg> }
function EditIcon() { return <svg {...iconProps}><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" /><path d="M18.5 2.5a2.12 2.12 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" /></svg> }
function TrashIcon() { return <svg {...iconProps}><polyline points="3 6 5 6 21 6" /><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /></svg> }
