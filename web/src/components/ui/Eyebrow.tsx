import { CSSProperties, ReactNode } from 'react'

export default function Eyebrow({ children, color, style }: { children: ReactNode; color?: string; style?: CSSProperties }) {
  return (
    <span
      style={{
        fontSize: 11,
        fontWeight: 700,
        letterSpacing: '0.06em',
        textTransform: 'uppercase',
        color: color ?? 'var(--color-text-tertiary)',
        ...style,
      }}
    >
      {children}
    </span>
  )
}
