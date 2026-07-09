import { Transaction } from '../../types'
import { fmt } from '../../lib/formatter'
import { useDeleteTransaction } from '../../hooks/useTransactions'
import ConfirmDialog from './ConfirmDialog'

export default function DeleteTxnDialog({ txn, onClose }: { txn: Transaction; onClose: () => void }) {
  const del = useDeleteTransaction()
  
  return (
    <ConfirmDialog
      title="Delete Transaction?"
      type="danger"
      message={
        <>
          Are you sure you want to delete this <strong>{txn.type}</strong> for <strong>{fmt(txn.amount)}</strong>?
        </>
      }
      confirmText="Delete"
      onConfirm={() => del.mutate(txn.id)}
      onClose={onClose}
    />
  )
}
