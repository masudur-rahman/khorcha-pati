import { useEffect, useMemo, useState, useRef, type KeyboardEvent } from 'react'
import { notify } from '../../lib/notify'
import { useQuery } from '@tanstack/react-query'
import { listCategories, listSubcategories } from '../../api/endpoints'
import { useCreateTransaction, useUpdateTransaction } from '../../hooks/useTransactions'
import { useWallets } from '../../hooks/useWallets'
import { useContacts } from '../../hooks/useContacts'
import type { Transaction, TxnType, TxnSubcategory } from '../../types'

import Modal from './Modal'
import Button from './Button'
import Input from './Input'
import Select from './Select'
import SearchableSelect from './SearchableSelect'
import ContactCombobox from './ContactCombobox'

export type { TxnType }
export const TXN_TYPE_OPTIONS: TxnType[] = ['Expense', 'Income', 'Transfer']

// Subcategories that imply a fixed counterpart wallet, mirroring the NL parser
// (modules/transaction/parser.go isVerbKeyword). from/to are wallet shortNames.
const SUBCAT_PREFILL: Record<string, { from?: string; to?: string }> = {
  'fin-with': { to: 'cash' },     // withdraw: money out of a bank into cash
  'fin-deposit': { from: 'cash' }, // deposit: cash into a bank
}

// Subcategories that settle a debt with a *person* — contact is mandatory. Bank
// loan/repayment are with an institution, so they're excluded. Mirrors
// isDebtSubcategory in services/transaction/transaction.go.
const CONTACT_REQUIRED_SUBS = new Set([
  'fin-lend', 'fin-recover', 'fin-borrow', 'fin-return',
])

interface Props {
  txn?: Transaction
  initialType?: TxnType
  initialContact?: string
  initialSubcategory?: string
  onClose: () => void
}

// Rank subcategories against a free-text query over name + keywords. Substring
// hits on the name rank highest, then token coverage across name+keywords.
function rankSubcategories(query: string, subs: TxnSubcategory[]): TxnSubcategory[] {
  const q = query.trim().toLowerCase()
  if (!q) return []
  const tokens = q.split(/\s+/)
  const scored = subs.map(s => {
    const name = s.name.toLowerCase()
    const hay = `${name} ${(s.keywords ?? '').toLowerCase()}`
    let score = 0
    if (name === q) score += 100
    else if (name.startsWith(q)) score += 60
    else if (name.includes(q)) score += 40
    for (const t of tokens) if (hay.includes(t)) score += 10
    return { s, score }
  })
  return scored.filter(x => x.score > 0).sort((a, b) => b.score - a.score).slice(0, 6).map(x => x.s)
}

export default function TxnDialog({ txn, initialType, initialContact, initialSubcategory, onClose }: Props) {
  const create = useCreateTransaction()
  const update = useUpdateTransaction()
  const isEdit = !!txn

  const amountRef = useRef<HTMLInputElement>(null)

  const { data: wallets = [] } = useWallets()
  const { data: contacts = [] } = useContacts()

  const [type, setType] = useState<TxnType>(txn?.type as any ?? initialType ?? 'Expense')
  const [amount, setAmount] = useState(txn?.amount?.toString() ?? '')
  const [catId, setCatId] = useState('')
  const [subcategoryId, setSubcategoryId] = useState(txn?.subcategoryId ?? initialSubcategory ?? '')
  const [srcId, setSrcId] = useState(txn?.srcId ?? '')
  const [dstId, setDstId] = useState(txn?.dstId ?? '')
  const [contactName, setContactName] = useState(txn?.contactName ?? initialContact ?? '')
  const [remarks, setRemarks] = useState(txn?.remarks ?? '')
  const [search, setSearch] = useState('')
  const [showResults, setShowResults] = useState(false)
  const [highlight, setHighlight] = useState(0)
  const [attempted, setAttempted] = useState(false)

  const { data: categories = [], isFetching: catFetching } = useQuery({
    queryKey: ['categories', type],
    queryFn: () => listCategories(type),
  })
  const { data: subcategories = [], isFetching: subFetching } = useQuery({
    queryKey: ['subcategories', catId, type],
    queryFn: () => listSubcategories(catId || undefined, type),
    enabled: !!catId,
  })

  // Full taxonomy for fuzzy search and edit-mode category restore.
  const { data: allCats = [] } = useQuery({ queryKey: ['categories', 'all'], queryFn: () => listCategories() })
  const { data: allSubs = [] } = useQuery({ queryKey: ['subcategories', 'all'], queryFn: () => listSubcategories() })

  const catNameById = useMemo(() => {
    const m: Record<string, string> = {}
    allCats.forEach(c => { m[c.id] = c.name })
    return m
  }, [allCats])

  const results = useMemo(() => rankSubcategories(search, allSubs), [search, allSubs])

  // Restore catId for a pre-set subcategory (edit mode or an initial prefill) by
  // looking it up in the full taxonomy, so the Sub Category select shows it.
  const presetSub = txn?.subcategoryId ?? initialSubcategory
  useEffect(() => {
    if (catId || !presetSub) return
    const cat = allSubs.find(s => s.id === presetSub)?.catId
    if (cat) setCatId(cat)
  }, [allSubs, presetSub, catId])

  // When type changes, drop the wallet field that no longer applies. Expense uses
  // srcId (debited), Income uses dstId (credited), Transfer uses both.
  useEffect(() => {
    if (isEdit) return
    if (type === 'Expense') setDstId('')
    else if (type === 'Income') setSrcId('')
  }, [type, isEdit])

  // Default the active wallet to Cash for Expense/Income when unset.
  useEffect(() => {
    if (isEdit || type === 'Transfer') return
    const cash = wallets.find(w => w.type === 'Cash')?.shortName ?? 'cash'
    if (type === 'Income') setDstId(prev => prev || cash)
    else setSrcId(prev => prev || cash)
  }, [type, wallets, isEdit])

  // When type changes, reset cat/subcat if the current selection no longer fits.
  // Guarded by list presence so a mid-flight refetch never clears a valid pick.
  useEffect(() => {
    if (isEdit || catFetching) return
    if (catId && categories.length && !categories.find(c => c.id === catId)) {
      setCatId('')
      setSubcategoryId('')
    }
  }, [categories, catId, isEdit, catFetching])

  useEffect(() => {
    if (isEdit || subFetching) return
    if (subcategoryId && subcategories.length && !subcategories.find(s => s.id === subcategoryId)) {
      setSubcategoryId('')
    }
  }, [subcategories, subcategoryId, isEdit, subFetching])

  // Apply a fuzzy-picked subcategory: set cat/subcat, derive type, prefill wallets.
  const applySubcategory = (s: TxnSubcategory) => {
    setCatId(s.catId)
    setSubcategoryId(s.id)
    const nextType = s.types.includes(type) ? type : s.types[0]
    if (nextType) setType(nextType)
    // A prefill fixes one wallet and leaves the other undecided — clear the other
    // side so a leftover default (e.g. Cash) doesn't collide with it.
    const prefill = SUBCAT_PREFILL[s.id]
    if (prefill) {
      setSrcId(prefill.from ?? '')
      setDstId(prefill.to ?? '')
    }
    setSearch('')
    setShowResults(false)
    setHighlight(0)
    setTimeout(() => {
      amountRef.current?.focus()
    }, 50)
  }

  // Keyboard traversal of the Quick Add suggestions.
  const onSearchKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (!showResults || results.length === 0) return
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setHighlight(h => Math.min(h + 1, results.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setHighlight(h => Math.max(h - 1, 0))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      const pick = results[Math.min(highlight, results.length - 1)]
      if (pick) applySubcategory(pick)
    } else if (e.key === 'Escape') {
      setShowResults(false)
    }
  }

  const walletVal = type === 'Income' ? dstId : srcId
  const setWalletVal = (v: string) => (type === 'Income' ? setDstId(v) : setSrcId(v))
  const contactRequired = CONTACT_REQUIRED_SUBS.has(subcategoryId)

  // Drop a stale contact when the subcategory no longer involves one.
  useEffect(() => {
    if (isEdit) return
    if (!contactRequired && contactName) setContactName('')
  }, [contactRequired, isEdit])

  // Validate and expose per-field errors only after a submit attempt.
  const validate = () => {
    const e: { amount?: string; sub?: string; wallet?: string; contact?: string } = {}
    const amt = parseFloat(amount)
    if (!amount || isNaN(amt) || amt <= 0) e.amount = 'Amount must be greater than zero'
    if (!subcategoryId) e.sub = 'Pick a subcategory'
    if (type === 'Transfer') {
      if (!srcId || !dstId) e.wallet = 'Select both wallets'
      else if (srcId === dstId) e.wallet = 'Source and destination must differ'
    }
    if (contactRequired && !contactName.trim()) e.contact = 'Contact is required'
    return e
  }
  const errors = attempted ? validate() : {}

  const handleSubmit = () => {
    setAttempted(true)
    if (Object.keys(validate()).length > 0) return
    const payload: Partial<Transaction> = {
      type,
      amount: parseFloat(amount),
      subcategoryId,
      remarks,
      // Sanitize per type so unused fields never carry stale values.
      srcId: type === 'Income' ? '' : srcId,
      dstId: type === 'Expense' ? '' : dstId,
      contactName: type === 'Transfer' ? '' : contactName,
    }
    if (isEdit) {
      update.mutate({ id: txn!.id, ...payload }, {
        onSuccess: () => {
          notify.updated('Transaction')
          onClose()
        },
        onError: (err) => notify.error(err, 'update transaction'),
      })
    } else {
      create.mutate(payload, {
        onSuccess: () => {
          notify.created('Transaction')
          onClose()
        },
        onError: (err) => notify.error(err, 'record transaction'),
      })
    }
  }

  const mutationError = create.isError || update.isError
  const walletOptions = wallets.map(w => ({ value: w.shortName, label: w.name }))

  return (
    <Modal
      title={isEdit ? 'Edit Transaction' : 'Add Transaction'}
      onClose={onClose}
      onSubmit={() => { if (!create.isPending && !update.isPending) handleSubmit() }}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={create.isPending || update.isPending}>
            {isEdit ? 'Update Changes' : 'Create Transaction'}
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        {/* Fuzzy smart-search */}
        <div style={{ position: 'relative' }}>
          <Input
            label="Smart Search"
            type="search"
            value={search}
            onChange={e => { setSearch(e.target.value); setShowResults(true); setHighlight(0) }}
            onKeyDown={onSearchKeyDown}
            placeholder="Search e.g. withdraw, salary, lunch, repay…"
          />
          {showResults && results.length > 0 && (
            <div style={{
              position: 'absolute', zIndex: 20, top: '100%', left: 0, right: 0, marginTop: 4,
              background: 'var(--color-surface)', border: '1px solid var(--color-border)',
              borderRadius: 12, boxShadow: '0 8px 24px rgba(0,0,0,0.12)', overflow: 'hidden',
            }}>
              {results.map((s, i) => (
                <button
                  key={s.id}
                  type="button"
                  onMouseDown={e => { e.preventDefault(); applySubcategory(s) }}
                  onMouseMove={() => setHighlight(i)}
                  style={{
                    display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: 2,
                    width: '100%', padding: '10px 16px', border: 'none',
                    background: i === highlight ? 'var(--color-hover)' : 'transparent',
                    cursor: 'pointer', textAlign: 'left',
                  }}
                >
                  <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--color-text-primary)' }}>
                    {catNameById[s.catId] ?? s.catId} › {s.name}
                  </span>
                  <span style={{ fontSize: 11, color: 'var(--color-text-tertiary)' }}>
                    {s.types.join(' / ')}{s.keywords ? ` · ${s.keywords}` : ''}
                  </span>
                </button>
              ))}
            </div>
          )}
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
          <Select label="Type" value={type} onChange={e => setType(e.target.value as TxnType)} options={TXN_TYPE_OPTIONS.map(t => ({ value: t, label: t }))} />
          <Input ref={amountRef} label="Amount" type="number" inputMode="decimal" value={amount} onChange={e => setAmount(e.target.value)} placeholder="0.00" error={errors.amount} />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
          <SearchableSelect label="Category" value={catId} onChange={v => { setCatId(v); setSubcategoryId('') }} options={categories.map(c => ({ value: c.id, label: c.name }))} placeholder="Search category…" />
          <SearchableSelect label="Sub Category" value={subcategoryId} onChange={setSubcategoryId} options={subcategories.map(s => ({ value: s.id, label: s.name }))} placeholder="Search sub category…" error={errors.sub} />
        </div>

        {type === 'Transfer' ? (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <Select label="From (Wallet)" value={srcId} onChange={e => setSrcId(e.target.value)} options={walletOptions.filter(o => o.value !== dstId)} error={errors.wallet} />
            <Select label="To (Wallet)" value={dstId} onChange={e => setDstId(e.target.value)} options={walletOptions.filter(o => o.value !== srcId)} />
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: contactRequired ? '1fr 1fr' : '1fr', gap: 16 }}>
            <Select label={type === 'Income' ? 'To (Wallet)' : 'From (Wallet)'} value={walletVal} onChange={e => setWalletVal(e.target.value)} options={walletOptions} />
            {/* Contact only matters for personal debt subcategories (lend/borrow/etc). */}
            {contactRequired && (
              <ContactCombobox
                label="Contact"
                contacts={contacts}
                value={contactName}
                onChange={setContactName}
                error={errors.contact}
              />
            )}
          </div>
        )}

        {/* Relative wrapper so the server error can float below without resizing the modal. */}
        <div style={{ position: 'relative' }}>
          <Input label="Remarks" value={remarks} onChange={e => setRemarks(e.target.value)} placeholder="Any notes..." />
          {mutationError && (
            <span style={{ position: 'absolute', top: '100%', left: 0, marginTop: 6, fontSize: 12, lineHeight: 1.2, color: 'var(--color-danger)', whiteSpace: 'nowrap' }}>
              Couldn't save the transaction. Please try again.
            </span>
          )}
        </div>
      </div>
    </Modal>
  )
}
