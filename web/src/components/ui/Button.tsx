import React from 'react'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger'
  icon?: React.ReactNode
}

export default function Button({ 
  children, 
  onClick, 
  icon, 
  variant = 'primary',
  style: sx,
  ...props 
}: ButtonProps) {
  const getStyles = (): React.CSSProperties => {
    const base: React.CSSProperties = {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: 8,
      padding: '10px 20px',
      borderRadius: 12,
      fontSize: 13,
      fontWeight: 600,
      border: 'none',
      cursor: 'pointer',
      transition: 'all 0.15s',
      ...sx,
    }

    if (variant === 'primary') {
      return {
        ...base,
        background: 'var(--color-primary)',
        color: 'white',
        boxShadow: '0 4px 12px var(--color-primary-shadow)',
      }
    }
    
    if (variant === 'danger') {
        return {
          ...base,
          background: 'var(--color-danger)',
          color: 'white',
          boxShadow: '0 4px 12px var(--color-danger-subtle)',
        }
    }

    if (variant === 'secondary') {
      return {
        ...base,
        background: 'transparent',
        color: 'var(--color-text-secondary)',
        border: '1px solid var(--color-border)',
      }
    }

    return base
  }

  return (
    <button 
      onClick={onClick} 
      style={getStyles()}
      onMouseEnter={e => e.currentTarget.style.transform = 'translateY(-1px)'}
      onMouseLeave={e => e.currentTarget.style.transform = 'translateY(0)'}
      {...props}
    >
      {icon && <span style={{ display: 'flex' }}>{icon}</span>}
      {children}
    </button>
  )
}
