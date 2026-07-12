import { useEffect, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { createAICache, updateAICache, type AICacheEntry, type AICacheInput } from '../../api/endpoints'
import type { TxnType } from '../../types'
import Modal from '../ui/Modal'
import Button from '../ui/Button'
import Input from '../ui/Input'
import Select from '../ui/Select'
import SearchableSelect from '../ui/SearchableSelect'
import { notify } from '../../lib/notify'

export interface SubMeta { name: string; catName: string; types: TxnType[] }

export default function AICacheModal({ entry, subMeta, subOptions, onClose, onSaved }: {
  entry: AICacheEntry | 'new'
  subMeta: Map<string, SubMeta>
  subOptions: { value: string; label: string }[]
  onClose: () => void
  onSaved: () => void
}) {
  const isEdit = entry !== 'new'
  const existing = isEdit ? (entry as AICacheEntry) : null
  const [inputText, setInputText] = useState(existing?.inputText ?? '')
  const [subId, setSubId] = useState(existing?.subcategoryId ?? '')
  const [intent, setIntent] = useState(existing?.intent ?? '')
  const [pct, setPct] = useState(existing ? Math.round(existing.confidence * 100) : 100)
  const [error, setError] = useState('')

  const allowedIntents = (subId ? subMeta.get(subId)?.types : undefined) ?? []

  // Keep intent valid for the chosen subcategory; auto-pick when only one type applies.
  useEffect(() => {
    if (allowedIntents.length && !allowedIntents.includes(intent as TxnType)) setIntent(allowedIntents[0])
  }, [subId]) // eslint-disable-line react-hooks/exhaustive-deps

  const save = useMutation({
    mutationFn: () => {
      const body: AICacheInput = { subcategoryId: subId, intent: intent.toLowerCase(), confidence: clampPct(pct) / 100 }
      return isEdit ? updateAICache(existing!.id, body) : createAICache({ ...body, inputText: inputText.trim() })
    },
    onSuccess: () => { isEdit ? notify.updated('Cache entry') : notify.created('Cache entry'); onSaved() },
    onError: (e: Error) => setError(e.message || 'Failed to save'),
  })

  const canSave = !!subId && !!intent && (isEdit || !!inputText.trim()) && !save.isPending

  return (
    <Modal
      title={isEdit ? 'Edit cache entry' : 'Add cache entry'}
      onClose={onClose}
      width={520}
      onSubmit={() => { if (canSave) save.mutate() }}
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={save.isPending}>Cancel</Button>
          <Button onClick={() => save.mutate()} disabled={!canSave}>
            {save.isPending ? 'Saving…' : isEdit ? 'Save changes' : 'Add entry'}
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <Input
          label="Input text"
          value={inputText}
          disabled={isEdit}
          placeholder="e.g. lunch at cafe"
          onChange={e => setInputText(e.target.value)}
        />
        {isEdit && <Hint>Input text is the cache key and can’t be changed. Delete and re-add to fix a typo.</Hint>}

        <SearchableSelect
          label="Subcategory"
          value={subId}
          options={subOptions}
          onChange={setSubId}
          placeholder="Search subcategory…"
        />

        <Select
          label="Intent"
          value={intent}
          disabled={!subId}
          options={allowedIntents.length
            ? allowedIntents.map(t => ({ value: t, label: t }))
            : [{ value: '', label: 'Pick a subcategory first' }]}
          onChange={e => setIntent(e.target.value)}
        />

        <Input
          label="Confidence (%)"
          type="number"
          min={0}
          max={100}
          value={pct}
          onChange={e => setPct(Number(e.target.value))}
        />
        {/* Wrapper carries the hint's hug-margin and floats the server error below it (no modal resize). */}
        <div style={{ position: 'relative', marginTop: -8 }}>
          <p style={{ margin: '0 0 0 4px', fontSize: 12, color: 'var(--color-text-tertiary)' }}>Manually curated entries are usually 100%.</p>
          {error && <p style={{ position: 'absolute', top: '100%', left: 4, right: 4, margin: '6px 0 0', color: 'var(--color-danger)', fontSize: 13, lineHeight: 1.2 }}>{error}</p>}
        </div>
      </div>
    </Modal>
  )
}

function clampPct(n: number): number {
  return Math.max(0, Math.min(100, Number.isFinite(n) ? n : 100))
}

function Hint({ children }: { children: React.ReactNode }) {
  return <p style={{ margin: '-8px 0 0 4px', fontSize: 12, color: 'var(--color-text-tertiary)' }}>{children}</p>
}
