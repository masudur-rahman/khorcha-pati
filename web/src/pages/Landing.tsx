import { useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import './Landing.css'

const BOT_URL = 'https://t.me/XpenseTrackerBot'
const REPO_URL = 'https://github.com/masudur-rahman/expense-tracker-bot'

export default function Landing() {
  const heroLayerRefs = useRef<HTMLElement[]>([])

  useEffect(() => {
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    const obs = new IntersectionObserver(
      entries => entries.forEach(e => {
        if (e.isIntersecting) {
          e.target.classList.add('visible')
          obs.unobserve(e.target)
        }
      }),
      { threshold: 0.18 }
    )
    document.querySelectorAll('.reveal').forEach(el => obs.observe(el))

    if (reduced) return () => obs.disconnect()

    const layers = heroLayerRefs.current
    const onScroll = () => {
      const y = window.scrollY
      layers.forEach(el => {
        const speed = Number(el.dataset.speed || '0.1')
        el.style.transform = `translate3d(0, ${-(y * speed)}px, 0)`
      })
    }
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => { obs.disconnect(); window.removeEventListener('scroll', onScroll) }
  }, [])

  const registerLayer = (el: HTMLElement | null) => {
    if (el && !heroLayerRefs.current.includes(el)) heroLayerRefs.current.push(el)
  }

  return (
    <div className="landing-body">
      <nav className="landing-nav">
        <Link to="/" className="logo">
          <img src="/logo-short.svg" alt="" className="logo-icon" />Hisab
        </Link>
        <ul>
          <li><a href="#features">Features</a></li>
          <li><a href="#how">How it works</a></li>
          <li><a href="#dashboard">Dashboard</a></li>
          <li><Link to="/login">Sign In</Link></li>
        </ul>
        <a className="nav-cta" href={BOT_URL} target="_blank" rel="noreferrer">Open in Telegram</a>
      </nav>

      <section className="hero">
        <div className="mesh" aria-hidden />
        <div className="hero-inner">
          <div className="hero-copy reveal">
            <span className="eyebrow">Personal finance · Telegram-first</span>
            <h1>
              Your finances,
              <br />
              one chat &amp; <span className="hl">one dashboard.</span>
            </h1>
            <p className="lead">
              Every taka, accounted for. Log expenses in plain English, get AI-classified categories, and review them in a precision dashboard built for clarity.
            </p>
            <div className="hero-cta">
              <a className="btn primary" href={BOT_URL} target="_blank" rel="noreferrer">Start on Telegram →</a>
              <Link className="btn ghost" to="/login">View Dashboard</Link>
            </div>
            <div className="trust">
              <span>⭐ 4.8 · happy trackers</span>
              <span>•</span>
              <span>Free, forever</span>
            </div>
          </div>

          <div className="hero-stack" aria-hidden>
            <div ref={registerLayer} className="parallax bg-art" data-speed="0.08" />
            <div ref={registerLayer} className="parallax phone" data-speed="0.18">
              <div className="phone-screen">
                <div className="phone-head"><span className="dot" /> Hisab Bot</div>
                <div className="chat">
                  <div className="bubble user">Lunch 320</div>
                  <div className="bubble bot">Saved 🍱 Food · Dining · ৳320 from bKash</div>
                  <div className="bubble user">Paid Karim 500</div>
                  <div className="bubble bot">Tracked 💰 Loan · −৳500 to @karim</div>
                  <div className="bubble user">Salary 52000</div>
                  <div className="bubble bot">Logged 💼 Income · +৳52,000 to BRAC</div>
                </div>
              </div>
            </div>
            <div ref={registerLayer} className="parallax card-floater bank" data-speed="0.26">
              <span className="card-brand">HISAB</span>
              <span className="card-num">•••• 4521</span>
              <span className="card-amt">৳ 37,630</span>
              <span className="card-name">BRAC BANK</span>
            </div>
            <div ref={registerLayer} className="parallax card-floater cash" data-speed="0.32">
              <span className="card-brand">HISAB</span>
              <span className="card-num">— — — —</span>
              <span className="card-amt">৳ 3,210</span>
              <span className="card-name">WALLET</span>
            </div>
          </div>
        </div>
      </section>

      <section className="band reveal" id="how">
        <h2>Designed for speed. Built for depth.</h2>
        <p>Stop wrestling spreadsheets. Use the chat in the go, and the command center at home.</p>
      </section>

      <section className="features" id="features">
        <div className="feature reveal">
          <span className="emoji">💬</span>
          <h3>Write like you talk</h3>
          <p>"Lunch 320", "Paid Karim 500", "Salary 52k" — parsed into structured entries. No forms.</p>
        </div>
        <div className="feature reveal">
          <span className="emoji">🤖</span>
          <h3>AI categorization</h3>
          <p>Gemini classifies subcategories with caching so the model gets faster and cheaper over time.</p>
        </div>
        <div className="feature reveal">
          <span className="emoji">☁️</span>
          <h3>Cloud-synced SQLite</h3>
          <p>Local-first DB with Google Drive backup. Zero vendor lock-in, your data stays portable.</p>
        </div>
        <div className="feature reveal">
          <span className="emoji">📊</span>
          <h3>Deep analytics</h3>
          <p>Wallets, contacts, budgets, donut charts, PDF statements — review in seconds.</p>
        </div>
      </section>

      <section className="dashboard-preview reveal" id="dashboard">
        <div className="dash">
          <div className="dash-side">
            <img src="/logo-short.svg" alt="" />
            <span>Hisab</span>
          </div>
          <div className="dash-main">
            <div className="dash-hero">
              <span>Good morning</span>
              <strong>Current Balance is ৳37,630</strong>
            </div>
            <div className="dash-strip">
              <div><em>Income</em><strong>+৳52,000</strong></div>
              <div><em>Expense</em><strong>−৳18,450</strong></div>
              <div><em>Net</em><strong>৳33,550</strong></div>
            </div>
          </div>
        </div>
        <div className="dash-copy">
          <span className="eyebrow">Financial command center</span>
          <h3>Everything important in one glance.</h3>
          <p>A web dashboard that gives the bot the big picture: balance trends, category spend, wallet flow, and downloadable statements.</p>
          <ul className="bullets">
            <li>100% data privacy — self-hostable</li>
            <li>0.1s p95 sync via SQLite</li>
            <li>Dark mode, mobile-first</li>
          </ul>
        </div>
      </section>

      <section className="cta">
        <h2>Ready to master your money?</h2>
        <p>Join trackers who reconcile their finances in seconds, not hours. Free. Forever.</p>
        <div className="cta-actions">
          <a className="btn primary" href={BOT_URL} target="_blank" rel="noreferrer">Start tracking on Telegram</a>
          <Link className="btn ghost" to="/login">Explore Dashboard</Link>
        </div>
      </section>

      <footer>
        <div>
          <Link to="/" className="logo small">
            <img src="/logo-short.svg" alt="" /> Hisab
          </Link>
          <p>Every taka, accounted for.</p>
        </div>
        <div>
          <h4>Product</h4>
          <a href={BOT_URL} target="_blank" rel="noreferrer">Telegram Bot</a>
          <Link to="/login">Dashboard</Link>
        </div>
        <div>
          <h4>Project</h4>
          <a href={REPO_URL} target="_blank" rel="noreferrer">GitHub</a>
          <a href={`${REPO_URL}/issues`} target="_blank" rel="noreferrer">Issues</a>
        </div>
        <div className="copy">© {new Date().getFullYear()} Hisab.</div>
      </footer>
    </div>
  )
}
