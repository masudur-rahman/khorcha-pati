import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as api from '../api/endpoints'
import type { Transaction } from '../types'

export function useTransactions(params?: Record<string, string>) {
  return useQuery({
    queryKey: ['transactions', params],
    queryFn: () => api.listTransactions(params),
  })
}

export function useCreateTransaction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (txn: Partial<Transaction>) => api.createTransaction(txn),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['transactions'] }),
  })
}

export function useUpdateTransaction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...txn }: Partial<Transaction> & { id: number }) =>
      api.updateTransaction(id, txn),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['transactions'] }),
  })
}

export function useDeleteTransaction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.deleteTransaction(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['transactions'] }),
  })
}
