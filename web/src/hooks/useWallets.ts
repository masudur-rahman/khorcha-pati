import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as api from '../api/endpoints'
import type { Wallet } from '../types'

export function useWallets() {
  return useQuery({
    queryKey: ['wallets'],
    queryFn: api.listWallets,
  })
}

export function useCreateWallet() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (wallet: Partial<Wallet>) => api.createWallet(wallet),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['wallets'] }),
  })
}
