import React from 'react'

interface CardProps {
  children: React.ReactNode
  style?: React.CSSProperties
  padding?: number | string
  onClick?: () => void
}

export default function Card({ children, style, padding = 24, onClick }: CardProps) {
  return (
    <div 
      onClick={onClick}
      style={{
        background: 'var(--color-surface)',
        borderRadius: 16,
        border: '1px solid var(--color-border)',
        padding,
        cursor: onClick ? 'pointer' : 'default',
        ...style,
      }}
    >
      {children}
    </div>
  )
}
