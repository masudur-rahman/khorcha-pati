import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  importAICache,
  type AICacheExportEntry, type AICacheImportMode, type AICacheImportSummary,
} from '../../api/endpoints'
import { notify } from '../../lib/notify'
import Modal from '../ui/Modal'
import Button from '../ui/Button'

const MODES: { value: AICacheImportMode; label: string; desc: string }[] = [
  { value: 'skip', label: 'Skip duplicates', desc: 'Keep existing entries; only add new input texts.' },
  { value: 'overwrite', label: 'Overwrite duplicates', desc: 'Imported entries replace the existing classification.' },
  { value: 'confidence', label: 'Higher confidence wins', desc: 'Replace only when the imported entry is more confident.' },
]

// AICacheImportDialog lets the admin pick a conflict mode before importing a parsed
// cache file, then reports the per-entry outcome.
export default function AICacheImportDialog({ entries, onClose }: {
  entries: AICacheExportEntry[]
  onClose: () => void
}) {
  const qc = useQueryClient()
  const [mode, setMode] = useState<AICacheImportMode>('skip')

  const imp = useMutation({
    mutationFn: () => importAICache(mode, entries),
    onSuccess: (s: AICacheImportSummary) => {
      notify.success(
        `Imported ${s.imported}, overwrote ${s.overwritten}, skipped ${s.skipped}` +
        (s.invalid ? `, ${s.invalid} invalid` : '') + '.'
      )
      qc.invalidateQueries({ queryKey: ['aiCache'] })
      onClose()
    },
    onError: (e) => notify.error(e, 'import cache'),
  })

  return (
    <Modal
      title="Import AI cache"
      subtitle={`${entries.length} ${entries.length === 1 ? 'entry' : 'entries'} in file`}
      onClose={onClose}
      width={460}
      onSubmit={() => { if (!imp.isPending) imp.mutate() }}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => imp.mutate()} disabled={imp.isPending || entries.length === 0}>
            {imp.isPending ? 'Importing…' : 'Import'}
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <p style={{ fontSize: 13, color: 'var(--color-text-tertiary)', margin: 0 }}>On conflicting input text:</p>
        {MODES.map(m => {
          const active = mode === m.value
          return (
            <button
              key={m.value}
              type="button"
              onClick={() => setMode(m.value)}
              style={{
                textAlign: 'left', cursor: 'pointer', fontFamily: 'inherit',
                display: 'flex', flexDirection: 'column', gap: 3, padding: '12px 14px', borderRadius: 10,
                border: `1.5px solid ${active ? 'var(--color-primary)' : 'var(--color-border)'}`,
                background: active ? 'var(--color-primary-subtle)' : 'var(--color-surface)',
              }}
            >
              <span style={{ fontSize: 14, fontWeight: 600, color: active ? 'var(--color-primary)' : 'var(--color-text-primary)' }}>{m.label}</span>
              <span style={{ fontSize: 12, color: 'var(--color-text-tertiary)' }}>{m.desc}</span>
            </button>
          )
        })}
      </div>
    </Modal>
  )
}
