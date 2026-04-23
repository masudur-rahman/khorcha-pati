import React from 'react'
import Card from './Card'
import { ICONS } from './Icons'

interface StatCardProps {
  label: string
  value: string | number
  trend?: string
  trendUp?: boolean
  icon: React.ReactNode
  accentColor: string
}

export default function StatCard({ 
  label, 
  value, 
  trend, 
  trendUp, 
  icon, 
  accentColor 
}: StatCardProps) {
  return (
    <Card style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span style={{ 
          fontSize: 12, 
          fontWeight: 600, 
          color: 'var(--color-text-tertiary)', 
          textTransform: 'uppercase', 
          letterSpacing: '0.06em' 
        }}>
          {label}
        </span>
        <div style={{
          width: 36, 
          height: 36, 
          borderRadius: 10,
          background: accentColor + '15', 
          color: accentColor,
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'center',
        }}>
          {icon}
        </div>
      </div>
      <div>
        <div style={{ 
          fontSize: 28, 
          fontWeight: 700, 
          color: 'var(--color-text-primary)', 
          letterSpacing: '-0.02em' 
        }}>
          {value}
        </div>
        {trend && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginTop: 6 }}>
            <span style={{
              display: 'inline-flex', 
              alignItems: 'center', 
              gap: 2,
              fontSize: 12, 
              fontWeight: 600,
              color: trendUp ? 'var(--color-success)' : 'var(--color-danger)',
              background: trendUp ? 'var(--color-success-subtle)' : 'var(--color-danger-subtle)',
              padding: '2px 8px', 
              borderRadius: 6,
            }}>
              {trendUp ? ICONS.arrowUp(12) : ICONS.arrowDown(12)}
              {trend}
            </span>
            <span style={{ fontSize: 11, color: 'var(--color-text-tertiary)' }}>vs last month</span>
          </div>
        )}
      </div>
    </Card>
  )
}
