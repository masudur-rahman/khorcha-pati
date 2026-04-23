interface LogoProps {
  size?: number
  collapsed?: boolean
}

export default function Logo({ size = 32, collapsed = false }: LogoProps) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: collapsed ? 0 : 12 }}>
      <svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        {/* Shield-like shape with upward arrow — expense tracking + protection */}
        <rect 
          x="4" y="4" width="32" height="32" rx="10" 
          fill="currentColor" 
          style={{ color: 'var(--color-primary)' }} 
        />
        {/* Abstract upward chart line */}
        <path 
          d="M12 26 L18 20 L23 23 L28 14" 
          stroke="white" strokeWidth="2.5" 
          strokeLinecap="round" strokeLinejoin="round" 
          fill="none" 
        />
        <circle cx="28" cy="14" r="2.5" fill="white" />
        {/* Small dollar hint */}
        <path 
          d="M12 28 L28 28" 
          stroke="white" strokeWidth="1.5" 
          strokeLinecap="round" 
          opacity="0.4" 
        />
      </svg>
      {!collapsed && (
        <span style={{ 
          fontSize: size * 0.5, 
          fontWeight: 700, 
          letterSpacing: '-0.03em', 
          color: 'var(--color-text-primary)', 
          lineHeight: 1 
        }}>
          Expense<span style={{ color: 'var(--color-primary)' }}> Tracker</span>
        </span>
      )}
    </div>
  )
}
