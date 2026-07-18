import { useState, useEffect, useRef } from 'react'
import { useNavigate, useLocation, Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { requestOTP, verifyOTP, initQR, pollQR, verifyMagicLink } from '../api/endpoints'
import { QRCodeSVG } from 'qrcode.react'
import { getAdminHandle } from '../api/client'
import { IS_ANDROID, autofillSafeType } from '../lib/platform'

export default function Login() {
  const [tab, setTab] = useState<'otp' | 'qr'>('otp')
  const [magicLoading, setMagicLoading] = useState(false)
  const [magicError, setMagicError] = useState('')
  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const [qrDeepLink, setQrDeepLink] = useState('')
  const [, setQrSessionID] = useState('')
  const [qrStatus, setQrStatus] = useState<'idle' | 'pending' | 'approved' | 'error'>('idle')
  const [qrError, setQrError] = useState('')
  const pollingRef = useRef<any>(null)

  useEffect(() => {
    const params = new URLSearchParams(location.search)
    const token = params.get('token')
    if (token) handleMagicLogin(token)
  }, [location])

  useEffect(() => { return () => { if (pollingRef.current) clearInterval(pollingRef.current) } }, [])

  const handleMagicLogin = async (token: string) => {
    setMagicLoading(true); setMagicError('')
    try { const data = await verifyMagicLink(token); login(data.accessToken, (data as any).refreshToken); navigate('/') }
    catch (e: any) { setMagicError(e.message || 'Magic link login failed') }
    finally { setMagicLoading(false) }
  }

  const startQR = async () => {
    if (qrStatus === 'pending') return
    setQrError(''); setQrStatus('idle')
    try {
      const data = await initQR(); setQrDeepLink(data.deepLink); setQrSessionID(data.sessionID); setQrStatus('pending')
      startPolling(data.sessionID)
    } catch (e: any) {
      setQrStatus('error')
      setQrError(friendlyQRError(e.message))
    }
  }

  const startPolling = (sessionID: string) => {
    if (pollingRef.current) clearInterval(pollingRef.current)
    pollingRef.current = setInterval(async () => {
      try {
        const data = await pollQR(sessionID)
        if (data.status === 'approved' && data.accessToken) { clearInterval(pollingRef.current); login(data.accessToken, (data as any).refreshToken); setQrStatus('approved'); navigate('/') }
        else if (data.status === 'denied') { clearInterval(pollingRef.current); setQrStatus('error'); setQrError('Access Denied. You cancelled the login from Telegram.') }
        else if (data.status === 'expired') { clearInterval(pollingRef.current); setQrStatus('error'); setQrError('QR session expired. Please generate a new one.') }
      } catch (e: any) {
        clearInterval(pollingRef.current); setQrStatus('error')
        setQrError(friendlyQRError(e.message))
      }
    }, 3000)
  }

  if (magicLoading) {
    return (
      <div style={wrapperStyle}>
        <div style={{ width: 40, height: 40, border: '3px solid #DFE1E6', borderTopColor: '#0052CC', borderRadius: '50%', animation: 'spin 0.8s linear infinite' }}></div>
        <p style={{ color: '#6B778C', fontWeight: 700, marginTop: 16 }}>Authenticating...</p>
        <style>{`@keyframes spin { to { transform: rotate(360deg) } }`}</style>
      </div>
    )
  }

  return (
    <div style={wrapperStyle}>
      <style>{`@keyframes spin { to { transform: rotate(360deg) } } @keyframes pulse { 0%, 100% { opacity: 1 } 50% { opacity: 0.4 } }`}</style>

      <div style={cardStyle}>
        {/* Back to home */}
        <Link to="/" style={{ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 13, color: '#6B778C', textDecoration: 'none', fontWeight: 500, marginBottom: 24, transition: 'color 0.15s' }}>
          <svg width={16} height={16} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
          Back to home
        </Link>

        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          {/* Logo */}
          <img src="/logo-mark.svg" alt="Khorcha-Pati" style={{ width: 48, height: 48, margin: '0 auto 16px', borderRadius: 14 }} />
          <h1 style={{ fontSize: 28, fontWeight: 700, color: '#172B4D', margin: '0 0 6px', letterSpacing: '-0.02em', fontFamily: "var(--font-display)" }}>Welcome to Khorcha-Pati</h1>
          <p style={{ fontSize: 14, color: '#6B778C', margin: 0, fontWeight: 500, fontStyle: 'italic' }}>Keep your khorcha on track.</p>
        </div>

        {magicError && (
          <div style={{ marginBottom: 24, padding: 12, background: 'var(--color-danger-subtle)', border: '1px solid var(--color-danger)', color: 'var(--color-danger)', fontSize: 13, borderRadius: 12, textAlign: 'center', fontWeight: 700 }}>{magicError}</div>
        )}

        {/* Tab bar */}
        <div style={{ display: 'flex', background: '#F4F5F7', borderRadius: 12, padding: 4, marginBottom: 28, gap: 4 }}>
          <button style={tabStyle(tab === 'otp')} onClick={() => setTab('otp')}>OTP Code</button>
          <button style={tabStyle(tab === 'qr')} onClick={() => { setTab('qr'); if (qrStatus === 'idle') startQR() }}>QR Scan</button>
        </div>

        <div style={{ minHeight: 220 }}>
          {tab === 'otp' ? <OTPLogin /> : <QRLogin status={qrStatus} deepLink={qrDeepLink} error={qrError} onRetry={startQR} />}
        </div>

        <div style={{ marginTop: 28, paddingTop: 20, borderTop: '1px solid #DFE1E6', textAlign: 'center' }}>
          <p style={{ fontSize: 12, color: '#6B778C', fontWeight: 500, margin: '0 0 8px' }}>
            💡 Fastest way in: send <code style={{ background: '#F4F5F7', color: '#003D9B', padding: '1px 6px', borderRadius: 5, fontWeight: 600 }}>/dashboard</code> to the bot and tap the button.
          </p>
          <p style={{ fontSize: 12, color: '#6B778C', fontWeight: 500, margin: 0 }}>
            Need help? Contact the admin{getAdminHandle() && <> — <a href={`https://t.me/${getAdminHandle().slice(1)}`} target="_blank" rel="noreferrer" style={{ color: '#0052CC', fontWeight: 700, textDecoration: 'none' }}>{getAdminHandle()}</a></>}
          </p>
        </div>
      </div>
    </div>
  )
}

const IDENTITY_HISTORY_KEY = 'khp_login_ids'

function OTPLogin() {
  const [identifier, setIdentifier] = useState('')
  const [code, setCode] = useState('')
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  // Previously used identifiers, surfaced via a datalist so the field can suggest
  // past usernames without inviting the browser's card/credential autofill.
  const [history, setHistory] = useState<string[]>(() => {
    try { return JSON.parse(localStorage.getItem(IDENTITY_HISTORY_KEY) || '[]') } catch { return [] }
  })
  const { login } = useAuth()
  const navigate = useNavigate()

  const rememberIdentifier = (id: string) => {
    const next = [id, ...history.filter(h => h !== id)].slice(0, 5)
    setHistory(next)
    localStorage.setItem(IDENTITY_HISTORY_KEY, JSON.stringify(next))
  }

  const handleSend = async () => {
    setError(''); setLoading(true)
    try { await requestOTP(identifier); rememberIdentifier(identifier.trim()); setSent(true) }
    catch (e: any) { setError(e.message) } finally { setLoading(false) }
  }
  const handleVerify = async () => {
    setError(''); setLoading(true)
    try { const data = await verifyOTP(identifier, code); login(data.accessToken, (data as any).refreshToken); navigate('/') }
    catch (e: any) { setError(e.message) } finally { setLoading(false) }
  }


  return (
    <form style={{ display: 'flex', flexDirection: 'column', gap: 16 }} autoComplete="off" onSubmit={e => e.preventDefault()}>
      <div>
        <label style={labelStyle}>Identity</label>
        <input type={autofillSafeType('username')} style={inputStyle} placeholder="Username or phone" value={identifier} onChange={e => setIdentifier(e.target.value)} disabled={sent} onKeyDown={e => { if (e.key === 'Enter' && !sent && identifier && !loading) handleSend(); }} autoComplete="khp-identity" autoCorrect="off" autoCapitalize="none" spellCheck={false} name="khp-identity" id="khp-identity" list="khp-identity-history" data-lpignore="true" data-1p-ignore="true" data-form-type="other" />
        {history.length > 0 && (
          <datalist id="khp-identity-history">
            {history.map(h => <option key={h} value={h} />)}
          </datalist>
        )}
      </div>
      {sent && (
        <div>
          <label style={labelStyle}>Verification Code</label>
          <OtpBoxes value={code} onChange={setCode} onEnter={() => { if (code.length === 6 && !loading) handleVerify() }} />
        </div>
      )}
      {error && <LoginNotice text={error} />}
      <button type="button" style={btnStyle} onClick={sent ? handleVerify : handleSend} disabled={loading || (!sent && !identifier) || (sent && code.length < 6)}>
        {loading ? 'Please wait...' : sent ? 'Verify Code' : 'Send Code'}
      </button>
      {sent && (
        <button type="button" style={{ background: 'none', border: 'none', fontSize: 12, color: '#6B778C', cursor: 'pointer', padding: 8, fontWeight: 600, fontFamily: 'inherit' }} onClick={() => { setSent(false); setCode('') }}>Resend code</button>
      )}
    </form>
  )
}

// linkify turns bare URLs in server messages into clickable links.
function linkify(text: string) {
  return text.split(/(https?:\/\/\S+)/g).map((part, i) =>
    /^https?:\/\//.test(part)
      ? <a key={i} href={part} target="_blank" rel="noreferrer" style={{ color: '#0052CC', fontWeight: 600, wordBreak: 'break-all' }}>{part}</a>
      : part,
  )
}

/* Server messages under the login form. The restricted-instance redirect is
   informational (amber card, clickable links); real failures stay danger-red. */
function LoginNotice({ text }: { text: string }) {
  const info = text.includes('🔒')
  return (
    <div style={{
      padding: '10px 14px', borderRadius: 12, fontSize: 12.5, lineHeight: 1.6,
      textAlign: 'left', whiteSpace: 'pre-line',
      background: info ? '#FFF7E6' : '#FFEBE6',
      border: `1px solid ${info ? '#FFD591' : '#FFBDAD'}`,
      color: '#172B4D', fontWeight: 500,
    }}>
      {linkify(text)}
    </div>
  )
}

const OTP_LEN = 6

/* One logical OTP input rendered as 6 boxes. The string state is the single
   source of truth: typing appends at the active (first empty) cell and
   advances, Backspace always clears the last filled cell, and a paste or
   keyboard autofill anywhere replaces the whole code. Focus is forced onto
   the active cell so users can't type into arbitrary boxes. */
function OtpBoxes({ value, onChange, onEnter }: { value: string; onChange: (v: string) => void; onEnter: () => void }) {
  const refs = useRef<(HTMLInputElement | null)[]>([])
  const [focused, setFocused] = useState(false)
  const active = Math.min(value.length, OTP_LEN - 1)
  const complete = value.length === OTP_LEN

  // Keep focus on the active cell as the value grows/shrinks, but only when
  // focus is already inside the boxes — never steal it from another field.
  useEffect(() => {
    if (refs.current.some(el => el === document.activeElement)) refs.current[active]?.focus()
  }, [active])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace') {
      e.preventDefault()
      onChange(value.slice(0, -1))
    } else if (e.key === 'Enter') {
      onEnter()
    } else if (/^[0-9]$/.test(e.key)) {
      e.preventDefault()
      if (value.length < OTP_LEN) onChange(value + e.key)
    }
  }

  // Fallback for virtual keyboards that don't emit proper key events, and for
  // OTP autofill dumping the entire code into one box.
  const handleInput = (e: React.FormEvent<HTMLInputElement>) => {
    const digits = e.currentTarget.value.replace(/\D/g, '')
    if (digits.length > 1) onChange(digits.slice(0, OTP_LEN))
    else if (digits.length === 1 && value.length < OTP_LEN) onChange(value + digits)
    e.currentTarget.value = value[Number(e.currentTarget.dataset.idx)] ?? ''
  }

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault()
    const digits = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, OTP_LEN)
    if (digits) onChange(digits)
  }

  return (
    <div style={{ display: 'flex', gap: 6, justifyContent: 'center' }}>
      {Array.from({ length: OTP_LEN }, (_, i) => {
        const filled = value[i] !== undefined
        const isActive = focused && i === active && !complete
        // Cell states, in the project palette: idle gray → filled primary
        // tint → active primary ring → all-six success green.
        const border = complete ? '#36B37E' : isActive ? '#0052CC' : filled ? '#99C2FF' : '#DFE1E6'
        const background = complete ? '#E8F9F1' : isActive ? '#FFFFFF' : filled ? '#E9F2FF' : '#FAFBFC'
        const color = complete ? '#00875A' : filled ? '#0052CC' : '#172B4D'
        return (
          <input
            key={i}
            ref={el => { refs.current[i] = el }}
            data-idx={i}
            type={autofillSafeType('text')}
            value={value[i] ?? ''}
            placeholder="•"
            onChange={() => { /* handled via onInput to survive autofill */ }}
            onInput={handleInput}
            onKeyDown={handleKeyDown}
            onPaste={handlePaste}
            onFocus={() => { setFocused(true); if (i !== active) refs.current[active]?.focus() }}
            onBlur={() => setFocused(false)}
            autoFocus={i === 0}
            inputMode="numeric"
            pattern="[0-9]*"
            autoComplete={i === 0 && !IS_ANDROID ? 'one-time-code' : 'khp-otp'}
            name={`khp-otp-${i}`}
            data-lpignore="true"
            data-1p-ignore="true"
            data-form-type="other"
            style={{
              width: 34, height: 40, textAlign: 'center',
              fontSize: 17, fontWeight: 700, color, border: `1.5px solid ${border}`, borderRadius: 8,
              background,
              // Group the code as 3+3 — easier to read back against the message.
              marginRight: i === 2 ? 10 : 0,
              boxShadow: isActive ? '0 0 0 3px rgba(0, 82, 204, 0.16)' : complete ? '0 0 0 3px rgba(54, 179, 126, 0.14)' : 'none',
              transform: isActive ? 'scale(1.07)' : 'scale(1)',
              outline: 'none', caretColor: 'transparent',
              transition: 'border-color 0.15s, background 0.15s, box-shadow 0.15s, transform 0.15s',
            }}
          />
        )
      })}
    </div>
  )
}

function friendlyQRError(msg: string): string {
  const lower = (msg || '').toLowerCase()
  if (lower.includes('expired')) return 'QR session expired. Please generate a new one.'
  if (lower.includes('denied') || lower.includes('cancelled')) return 'Access Denied. You cancelled the login from Telegram.'
  if (lower.includes('already used')) return 'This QR code has already been used. Please generate a new one.'
  if (lower.includes('network') || lower.includes('fetch')) return 'Connection lost. Please check your network and try again.'
  return msg || 'Something went wrong. Please try again.'
}

function qrErrorTitle(error: string): string {
  if (error.includes('Denied') || error.includes('cancelled')) return 'Login Denied'
  if (error.includes('expired')) return 'Session Expired'
  if (error.includes('Connection')) return 'Connection Error'
  return 'Login Failed'
}

function QRLogin({ status, deepLink, error, onRetry }: { status: string; deepLink: string; error: string; onRetry: () => void }) {
  if (status === 'idle') {
    return (
      <div style={{ textAlign: 'center', padding: '32px 0' }}>
        <div style={{ width: 40, height: 40, border: '3px solid #DFE1E6', borderTopColor: '#0052CC', borderRadius: '50%', animation: 'spin 0.8s linear infinite', margin: '0 auto 16px' }}></div>
        <p style={{ color: '#6B778C', fontSize: 13, fontWeight: 500 }}>Generating secure QR code...</p>
      </div>
    )
  }
  if (status === 'error') {
    return (
      <div style={{ textAlign: 'center', padding: '16px 0', display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ width: 64, height: 64, borderRadius: '50%', background: '#FFEBE6', color: '#DE350B', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto', fontSize: 24, border: '1px solid #DE350B' }}>
          {error.includes('Denied') ? '🚫' : '⚠️'}
        </div>
        <p style={{ fontSize: 14, fontWeight: 700, color: '#172B4D', margin: '8px 0 0' }}>{qrErrorTitle(error)}</p>
        <p style={{ fontSize: 12, color: '#6B778C', margin: 0, padding: '0 16px' }}>{error}</p>
        <button style={{ ...btnStyle, background: '#172B4D' }} onClick={onRetry}>Generate New QR</button>
      </div>
    )
  }
  return (
    <div style={{ textAlign: 'center', padding: '4px 0' }}>
      <div style={{ width: 180, height: 180, margin: '0 auto 20px', borderRadius: 16, border: '2px dashed #DEEBFF', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'white' }}>
        <QRCodeSVG value={deepLink} size={150} />
      </div>
      <p style={{ fontSize: 14, fontWeight: 700, color: '#505F79', margin: '0 0 8px' }}>Scan with phone camera or Google Lens</p>
      <p style={{ fontSize: 11, color: '#6B778C', margin: '0 0 4px' }}>
        Telegram's in-app camera only previews bot profile.
      </p>
      <p style={{ fontSize: 12, color: '#6B778C', margin: 0 }}>
        Or <a href={deepLink} target="_blank" rel="noreferrer" style={{ color: '#0052CC', fontWeight: 700, textDecoration: 'none' }}>click here</a> to open in Telegram
      </p>
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8, marginTop: 16, background: '#DEEBFF', padding: '8px 24px', borderRadius: 12 }}>
        <div style={{ height: 6, width: 6, background: '#0052CC', borderRadius: '50%', animation: 'pulse 1.5s infinite' }}></div>
        <span style={{ fontSize: 11, fontWeight: 700, color: '#0052CC', textTransform: 'uppercase', letterSpacing: '0.12em' }}>Waiting for scan</span>
      </div>
    </div>
  )
}

const wrapperStyle: React.CSSProperties = {
  minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center',
  flexDirection: 'column', background: '#F4F5F7', padding: 24,
}
const cardStyle: React.CSSProperties = {
  width: '100%', maxWidth: 420, background: 'white', borderRadius: 24,
  padding: '40px 36px', boxShadow: '0 20px 60px rgba(0,0,0,0.06), 0 0 0 1px #DFE1E6',
}
const tabStyle = (active: boolean): React.CSSProperties => ({
  flex: 1, padding: '10px 0', textAlign: 'center', borderRadius: 10, fontSize: 13, fontWeight: 600,
  cursor: 'pointer', border: 'none', transition: 'all 0.2s', fontFamily: 'inherit',
  background: active ? 'white' : 'transparent',
  color: active ? '#0052CC' : '#6B778C',
  boxShadow: active ? '0 2px 8px rgba(0,0,0,0.05)' : 'none',
})
const inputStyle: React.CSSProperties = {
  width: '100%', padding: '14px 16px', borderRadius: 12, fontSize: 14,
  border: '1px solid #DFE1E6', background: '#F4F5F7', color: '#172B4D',
  outline: 'none', transition: 'border 0.2s', boxSizing: 'border-box', fontFamily: 'inherit',
  WebkitAppearance: 'none', appearance: 'none',
}
const labelStyle: React.CSSProperties = {
  fontSize: 11, fontWeight: 700, color: '#6B778C', textTransform: 'uppercase',
  letterSpacing: '0.06em', marginBottom: 6, display: 'block',
}
const btnStyle: React.CSSProperties = {
  width: '100%', padding: '14px 0', borderRadius: 12, fontSize: 14, fontWeight: 700,
  background: '#0052CC', color: 'white', border: 'none', cursor: 'pointer',
  boxShadow: '0 4px 16px rgba(0,82,204,0.2)', transition: 'all 0.15s', marginTop: 8,
  fontFamily: 'inherit',
}
