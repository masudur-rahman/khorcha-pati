interface LogoProps {
  size?: number
  collapsed?: boolean
}

export default function Logo({ size = 32, collapsed = false }: LogoProps) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: collapsed ? 0 : 12 }}>
      <img
        src="/logo-short.svg"
        alt="Hisab"
        style={{ height: size, width: size, borderRadius: size * 0.22 }}
      />
      {!collapsed && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 2, lineHeight: 1 }}>
          <span style={{
            fontSize: size * 0.55,
            fontWeight: 700,
            letterSpacing: '-0.03em',
            color: 'var(--color-text-primary)',
            fontFamily: 'var(--font-display)',
          }}>
            Hisab
          </span>
          <span style={{
            fontSize: 10,
            fontWeight: 600,
            letterSpacing: '0.05em',
            textTransform: 'uppercase',
            color: 'var(--color-text-tertiary)',
          }}>
            Every taka, accounted for
          </span>
        </div>
      )}
    </div>
  )
}
