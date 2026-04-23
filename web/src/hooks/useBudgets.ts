import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as api from '../api/endpoints'

export function useBudgets() {
  return useQuery({ queryKey: ['budgets'], queryFn: api.listBudgets })
}

export function useBudgetAlerts() {
  return useQuery({ queryKey: ['budgetAlerts'], queryFn: api.getBudgetAlerts })
}

export function useSetBudget() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ categoryId, amount, alertAt }: { categoryId: string; amount: number; alertAt: number }) =>
      api.setBudget(categoryId, amount, alertAt),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['budgets'] }),
  })
}

export function useDeleteBudget() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (categoryId: string) => api.deleteBudget(categoryId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['budgets'] }),
  })
}
