import React, { useState } from 'react'

export interface ActionButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  actionType: 'edit' | 'delete'
  icon: React.ReactNode
  variant?: 'default' | 'glass'
}

export default function ActionButton({ actionType, icon, variant = 'default', style, disabled, children, ...props }: ActionButtonProps) {
  const [isHovered, setIsHovered] = useState(false)
  const isDelete = actionType === 'delete'
  const accent = isDelete ? 'var(--color-danger)' : 'var(--color-primary)'
  
  const isGlass = variant === 'glass'
  
  // Use CSS color-mix to create very light and slightly darker backgrounds
  const baseBg = isGlass ? 'rgba(255,255,255,0.2)' : `color-mix(in srgb, ${accent} 12%, transparent)`
  const hoverBg = isGlass ? 'rgba(255,255,255,0.35)' : `color-mix(in srgb, ${accent} 24%, transparent)`
  
  // Base Color
  const color = isGlass ? 'white' : accent

  return (
    <button
      {...props}
      disabled={disabled}
      onMouseEnter={(e) => {
        setIsHovered(true)
        if (props.onMouseEnter) props.onMouseEnter(e)
      }}
      onMouseLeave={(e) => {
        setIsHovered(false)
        if (props.onMouseLeave) props.onMouseLeave(e)
      }}
      style={{
        width: children ? 'auto' : 32,
        height: 32,
        padding: children ? '0 12px' : 0,
        borderRadius: 'var(--radius-sm)',
        flexShrink: 0,
        background: isHovered && !disabled ? hoverBg : baseBg,
        color: color,
        border: 'none',
        cursor: disabled ? 'not-allowed' : 'pointer',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 6,
        transition: 'all var(--transition-fast)',
        opacity: disabled ? 0.5 : 1,
        fontFamily: 'inherit',
        fontWeight: 600,
        ...style
      }}
    >
      {icon}
      {children}
    </button>
  )
}
