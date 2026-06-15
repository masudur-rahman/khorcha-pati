import { ReactNode } from 'react'

interface Props {
  title: string
  accent?: string
  action?: ReactNode
}

export default function SectionHeader({ title, accent = 'var(--color-primary)', action }: Props) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
      <h3
        style={{
          fontSize: 15,
          fontWeight: 700,
          color: 'var(--color-text-primary)',
          margin: 0,
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          fontFamily: 'var(--font-display)',
        }}
      >
        <span style={{ width: 3, height: 18, borderRadius: 2, background: accent }} />
        {title}
      </h3>
      {action}
    </div>
  )
}
