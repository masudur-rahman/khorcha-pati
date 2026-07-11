interface BudgetGaugeProps {
  percent: number
  size?: number
  color?: string
  endColor?: string
  trackColor?: string
  textColor?: string
}

export default function BudgetGauge({
  percent,
  size = 100,
  color,
  endColor,
  trackColor,
  textColor,
}: BudgetGaugeProps) {
  const isOverflow = percent > 100
  const clamped = Math.max(0, Math.min(100, percent))
  const percentDeg = (clamped / 100) * 360

  let overflowPercent = 0
  if (isOverflow) {
    overflowPercent = percent % 100
    if (overflowPercent === 0) overflowPercent = 100
  }
  const overflowDeg = (overflowPercent / 100) * 360

  const resolvedColor = color ?? (
    percent > 90 ? 'var(--color-danger)'
      : percent > 70 ? '#f59e0b'
      : 'var(--color-primary)'
  )
  const resolvedEnd = endColor ?? resolvedColor
  const resolvedTrack = trackColor ?? 'var(--color-border)'
  const resolvedText = textColor ?? 'var(--color-text-primary)'

  // Fixed angular palette across the full 360° (clockwise from 12 o'clock):
  //   0°   → 90°  : pure accent
  //   90° → 270°  : accent → 70% red (gradient ramp 1)
  //   270° → 360° : 70% red → full red (gradient ramp 2 — steeper "more reddish" feel)
  // The arc fills 0 → percentDeg and stops, showing whatever color the palette
  // has at the cutoff angle.
  const PLATEAU = 90
  const MID_STOP = 270
  const MID_RATIO = 0.7
  const mix = (ratio: number) => `color-mix(in srgb, ${resolvedEnd} ${ratio * 100}%, ${resolvedColor})`
  const midColor = mix(MID_RATIO)

  let cutoff: string
  if (percentDeg <= PLATEAU) {
    cutoff = resolvedColor
  } else if (percentDeg <= MID_STOP) {
    const t = (percentDeg - PLATEAU) / (MID_STOP - PLATEAU)
    cutoff = mix(t * MID_RATIO)
  } else {
    const t = (percentDeg - MID_STOP) / (360 - MID_STOP)
    cutoff = mix(MID_RATIO + t * (1 - MID_RATIO))
  }

  let stops: string
  if (percentDeg <= 0) {
    stops = `${resolvedTrack} 0deg 360deg`
  } else if (percentDeg <= PLATEAU) {
    stops = `${resolvedColor} 0deg ${percentDeg}deg, ${resolvedTrack} ${percentDeg}deg 360deg`
  } else if (percentDeg <= MID_STOP) {
    stops = `${resolvedColor} 0deg ${PLATEAU}deg, ${cutoff} ${percentDeg}deg, ${resolvedTrack} ${percentDeg}deg 360deg`
  } else {
    stops = `${resolvedColor} 0deg ${PLATEAU}deg, ${midColor} ${MID_STOP}deg, ${cutoff} ${percentDeg}deg, ${resolvedTrack} ${percentDeg}deg 360deg`
  }

  const ringWidth = size * 0.1
  const innerR = (size - ringWidth * 2) / 2
  const mask = `radial-gradient(circle, transparent ${innerR}px, #000 ${innerR + 1}px)`

  // Second lap (overflow) stops
  const overflowBase = '#991b1b' // dark red
  const overflowTip = '#450a0a'  // darker red for 3D overlap effect
  let stops2 = ''
  if (isOverflow) {
    if (overflowDeg >= 360) {
      stops2 = `${overflowBase} 0deg 360deg`
    } else {
      const shadowStart = Math.max(0, overflowDeg - 5)
      stops2 = `${overflowBase} 0deg ${shadowStart}deg, ${overflowTip} ${overflowDeg}deg, transparent ${overflowDeg}deg 360deg`
    }
  }

  return (
    <div style={{ width: size, height: size, position: 'relative', display: 'inline-block' }}>
      {/* Base Lap */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          borderRadius: '50%',
          background: `conic-gradient(from 0deg, ${stops})`,
          WebkitMask: mask,
          mask: mask,
          transition: 'background 0.4s ease',
        }}
      />
      {/* Overflow Lap */}
      {isOverflow && (
        <div
          style={{
            position: 'absolute',
            inset: 0,
            borderRadius: '50%',
            background: `conic-gradient(from 0deg, ${stops2})`,
            WebkitMask: mask,
            mask: mask,
            transition: 'background 0.4s ease',
            zIndex: 2,
          }}
        />
      )}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: size * 0.2,
          fontWeight: 700,
          color: resolvedText,
          fontFamily: 'var(--font-display)',
          letterSpacing: '-0.02em',
          zIndex: 3,
        }}
      >
        {percent.toFixed(0)}%
      </div>
    </div>
  )
}
