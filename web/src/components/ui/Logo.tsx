interface LogoProps {
  size?: number
  collapsed?: boolean
}

export default function Logo({ size = 32, collapsed = false }: LogoProps) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: collapsed ? 0 : 12 }}>
      <img
        src="/logo-short.svg"
        alt="Khorcha-Pati"
        style={{ height: size, width: size, borderRadius: size * 0.22 }}
      />
      {!collapsed && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4, lineHeight: 1 }}>
          <span style={{
            fontSize: size * 0.62,
            fontWeight: 800,
            letterSpacing: '-0.025em',
            color: 'var(--color-text-primary)',
            fontFamily: 'var(--font-display)',
            lineHeight: 1,
          }}>
            Khorcha<span style={{ color: 'var(--color-primary)' }}>-Pati</span>
          </span>
          <span style={{
            fontSize: 11,
            fontWeight: 500,
            color: 'var(--color-text-tertiary)',
            letterSpacing: '0.01em',
            lineHeight: 1.2,
          }}>
            Keep your <span style={{ color: 'var(--color-primary)', fontWeight: 600 }}>khorcha</span> on track.
          </span>
        </div>
      )}
    </div>
  )
}
