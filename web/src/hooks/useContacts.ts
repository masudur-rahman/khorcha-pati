import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as api from '../api/endpoints'
import type { Contact } from '../types'

export function useContacts() {
  return useQuery({
    queryKey: ['contacts'],
    queryFn: api.listContacts,
  })
}

export function useCreateContact() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (contact: Partial<Contact>) => api.createContact(contact),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['contacts'] }),
  })
}
