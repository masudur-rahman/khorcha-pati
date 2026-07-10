import { useState } from 'react'

interface DateRangePickerProps {
  startDate: string
  endDate: string
  onChange: (start: string, end: string) => void
}

export default function DateRangePicker({ startDate, endDate, onChange }: DateRangePickerProps) {
  const [viewDate, setViewDate] = useState(startDate ? new Date(startDate) : new Date())
  const [focusedInput, setFocusedInput] = useState<'start' | 'end'>('start')

  const currentMonth = viewDate.getMonth()
  const currentYear = viewDate.getFullYear()

  const daysInMonth = new Date(currentYear, currentMonth + 1, 0).getDate()
  const firstDayOfMonth = new Date(currentYear, currentMonth, 1).getDay()

  const handlePrevMonth = () => setViewDate(new Date(currentYear, currentMonth - 1, 1))
  const handleNextMonth = () => setViewDate(new Date(currentYear, currentMonth + 1, 1))

  const handleDateClick = (day: number) => {
    const clickedDate = new Date(currentYear, currentMonth, day)
    const offset = clickedDate.getTimezoneOffset() * 60000
    const localISOTime = new Date(clickedDate.getTime() - offset).toISOString().split('T')[0]

    if (focusedInput === 'start') {
      if (endDate && new Date(localISOTime) > new Date(endDate)) {
        onChange(localISOTime, '')
      } else {
        onChange(localISOTime, endDate)
      }
      setFocusedInput('end')
    } else {
      if (startDate && new Date(localISOTime) < new Date(startDate)) {
        onChange(localISOTime, '')
        setFocusedInput('end')
      } else {
        onChange(startDate, localISOTime)
        // Optionally switch back to start or blur
      }
    }
  }

  const isSelected = (day: number) => {
    const d = new Date(currentYear, currentMonth, day)
    const offset = d.getTimezoneOffset() * 60000
    const dateStr = new Date(d.getTime() - offset).toISOString().split('T')[0]
    return dateStr === startDate || dateStr === endDate
  }

  const isInRange = (day: number) => {
    if (!startDate || !endDate) return false
    const d = new Date(currentYear, currentMonth, day)
    const offset = d.getTimezoneOffset() * 60000
    const dateStr = new Date(d.getTime() - offset).toISOString().split('T')[0]
    return dateStr > startDate && dateStr < endDate
  }

  const today = new Date()
  const todayOffset = today.getTimezoneOffset() * 60000
  const todayStr = new Date(today.getTime() - todayOffset).toISOString().split('T')[0]

  const days = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']
  const months = Array.from({ length: 12 }, (_, i) => new Date(0, i).toLocaleString('default', { month: 'long' }))
  const years = Array.from({ length: 30 }, (_, i) => new Date().getFullYear() - 10 + i)

  const fmtDate = (d: string) => d ? new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' }) : ''

  return (
    <div>
      <div style={{ display: 'flex', gap: 12, marginBottom: 12 }}>
        <div 
          onClick={() => setFocusedInput('start')}
          style={{ 
            flex: 1, padding: '10px 14px', background: 'var(--color-bg)', 
            border: `1px solid ${focusedInput === 'start' ? 'var(--color-primary)' : 'var(--color-border)'}`, 
            borderRadius: 8, fontSize: 13, cursor: 'pointer',
            color: startDate ? 'var(--color-text-primary)' : 'var(--color-text-tertiary)',
            boxShadow: focusedInput === 'start' ? '0 0 0 2px var(--color-primary-subtle)' : 'none',
            transition: 'all 0.1s'
          }}>
          {startDate ? fmtDate(startDate) : 'Start Date'}
        </div>
        <div 
          onClick={() => setFocusedInput('end')}
          style={{ 
            flex: 1, padding: '10px 14px', background: 'var(--color-bg)', 
            border: `1px solid ${focusedInput === 'end' ? 'var(--color-primary)' : 'var(--color-border)'}`, 
            borderRadius: 8, fontSize: 13, cursor: 'pointer',
            color: endDate ? 'var(--color-text-primary)' : 'var(--color-text-tertiary)',
            boxShadow: focusedInput === 'end' ? '0 0 0 2px var(--color-primary-subtle)' : 'none',
            transition: 'all 0.1s'
          }}>
          {endDate ? fmtDate(endDate) : 'End Date'}
        </div>
      </div>

      <div style={{ padding: '16px 20px', background: 'var(--color-bg)', borderRadius: 'var(--radius-md)', border: '1px solid var(--color-border)', userSelect: 'none' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <button onClick={handlePrevMonth} style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: 'var(--color-text-secondary)', padding: 4 }}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="15 18 9 12 15 6"></polyline></svg>
          </button>
          
          <div style={{ display: 'flex', gap: 4 }}>
            <select 
              value={currentMonth} 
              onChange={e => setViewDate(new Date(currentYear, parseInt(e.target.value), 1))}
              style={{ fontWeight: 600, fontSize: 14, color: 'var(--color-text-primary)', background: 'transparent', border: 'none', cursor: 'pointer', outline: 'none' }}
            >
              {months.map((m, i) => <option key={m} value={i}>{m}</option>)}
            </select>
            <select 
              value={currentYear} 
              onChange={e => setViewDate(new Date(parseInt(e.target.value), currentMonth, 1))}
              style={{ fontWeight: 600, fontSize: 14, color: 'var(--color-text-primary)', background: 'transparent', border: 'none', cursor: 'pointer', outline: 'none' }}
            >
              {years.map(y => <option key={y} value={y}>{y}</option>)}
            </select>
          </div>

          <button onClick={handleNextMonth} style={{ background: 'transparent', border: 'none', cursor: 'pointer', color: 'var(--color-text-secondary)', padding: 4 }}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="9 18 15 12 9 6"></polyline></svg>
          </button>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 4, marginBottom: 8, textAlign: 'center' }}>
          {days.map(d => (
            <div key={d} style={{ fontSize: 12, fontWeight: 600, color: 'var(--color-text-tertiary)' }}>{d}</div>
          ))}
        </div>

        {/* Fixed height grid to prevent jumping (6 rows * 36px) */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: '4px 0', textAlign: 'center', minHeight: 236 }}>
          {Array.from({ length: 42 }).map((_, i) => {
            const day = i - firstDayOfMonth + 1
            const isCurrentMonth = day > 0 && day <= daysInMonth

            if (!isCurrentMonth) {
              return <div key={`empty-${i}`} />
            }

            const d = new Date(currentYear, currentMonth, day)
            const offset = d.getTimezoneOffset() * 60000
            const dateStr = new Date(d.getTime() - offset).toISOString().split('T')[0]
            
            const isStart = dateStr === startDate
            const isEnd = dateStr === endDate
            const isToday = dateStr === todayStr

            const selected = isSelected(day)
            const inRange = isInRange(day)
            
            let bg = 'transparent'
            let color = 'var(--color-text-primary)'
            let borderRadius = '50%'
            
            if (selected) {
              bg = 'var(--color-primary)'
              color = 'white'
            } else if (inRange) {
              bg = 'var(--color-primary-subtle)'
              borderRadius = '0'
            } else if (isToday) {
              bg = 'rgba(9, 30, 66, 0.08)' // Visible slight dark round background for today
              color = 'var(--color-primary)'
            }

            return (
              <div key={i} style={{ position: 'relative', display: 'flex', justifyContent: 'center', height: 36, alignItems: 'center' }}>
                {inRange && <div style={{ position: 'absolute', top: 2, bottom: 2, left: 0, right: 0, background: 'var(--color-primary-subtle)' }} />}
                {selected && startDate && endDate && startDate !== endDate && (
                  <div style={{ 
                    position: 'absolute', top: 2, bottom: 2, 
                    left: isStart ? '50%' : 0, 
                    right: isEnd ? '50%' : 0, 
                    background: 'var(--color-primary-subtle)' 
                  }} />
                )}
                <button
                  onClick={() => handleDateClick(day)}
                  style={{
                    position: 'relative',
                    width: 32, height: 32,
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    background: bg, color, borderRadius,
                    border: 'none', cursor: 'pointer', fontSize: 14, fontWeight: selected ? 600 : 400,
                    transition: 'background 0.1s', fontFamily: 'inherit'
                  }}
                  onMouseEnter={e => { if (!selected && !inRange) e.currentTarget.style.background = 'var(--color-surface-hover)' }}
                  onMouseLeave={e => { if (!selected && !inRange) e.currentTarget.style.background = bg }}
                >
                  {day}
                </button>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
