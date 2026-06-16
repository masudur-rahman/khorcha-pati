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
            fontSize: 11,
            fontWeight: 500,
            fontStyle: 'italic',
            color: 'var(--color-text-tertiary)',
            letterSpacing: 0,
          }}>
            Every taka, accounted for.
          </span>
        </div>
      )}
    </div>
  )
}
