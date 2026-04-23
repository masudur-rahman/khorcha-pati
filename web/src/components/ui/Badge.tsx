import React from 'react'

interface BadgeProps {
  type: 'Income' | 'Expense' | 'Transfer'
  children?: React.ReactNode
}

export default function Badge({ type, children }: BadgeProps) {
  const colors = {
    Income: { bg: 'var(--color-success-subtle)', color: 'var(--color-success)' },
    Expense: { bg: 'var(--color-danger-subtle)', color: 'var(--color-danger)' },
    Transfer: { bg: 'var(--color-primary-subtle)', color: 'var(--color-primary)' },
  }
  
  const c = colors[type] || colors.Expense
  
  return (
    <span style={{
      display: 'inline-block',
      padding: '3px 10px',
      borderRadius: 8,
      fontSize: 11,
      fontWeight: 700,
      textTransform: 'uppercase',
      letterSpacing: '0.04em',
      background: c.bg,
      color: c.color,
    }}>
      {children || type}
    </span>
  )
}
