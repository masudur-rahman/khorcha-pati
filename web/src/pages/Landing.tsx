import { useEffect, useRef } from 'react'
import { Link } from 'react-router-dom'
import './Landing.css'

const BOT_URL = 'https://t.me/XpenseTrackerBot'
const REPO_URL = 'https://github.com/masudur-rahman/expense-tracker-bot'

export default function Landing() {
  const heroLayerRefs = useRef<HTMLElement[]>([])

  useEffect(() => {
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches

    // Reveal-on-scroll with stagger via `data-stagger` (ms)
    const obs = new IntersectionObserver(
      entries => entries.forEach(e => {
        if (e.isIntersecting) {
          const target = e.target as HTMLElement
          const delay = Number(target.dataset.stagger || '0')
          target.style.transitionDelay = `${delay}ms`
          target.classList.add('visible')
          obs.unobserve(target)
        }
      }),
      { threshold: 0.18 }
    )
    document.querySelectorAll('.reveal').forEach(el => obs.observe(el))

    if (reduced) return () => obs.disconnect()

    // Parallax — hero stack + section banners
    const layers = heroLayerRefs.current
    let rafId = 0
    const apply = () => {
      const y = window.scrollY
      layers.forEach(el => {
        const speed = Number(el.dataset.speed || '0.1')
        const rotate = Number(el.dataset.rotate || '0')
        const baseRotate = el.dataset.baseRotate || '0deg'
        el.style.transform = `translate3d(0, ${-(y * speed)}px, 0) rotate(calc(${baseRotate} + ${y * rotate}deg))`
      })
      rafId = 0
    }
    const onScroll = () => {
      if (!rafId) rafId = requestAnimationFrame(apply)
    }
    window.addEventListener('scroll', onScroll, { passive: true })
    apply()

    return () => {
      obs.disconnect()
      window.removeEventListener('scroll', onScroll)
      if (rafId) cancelAnimationFrame(rafId)
    }
  }, [])

  const registerLayer = (el: HTMLElement | null) => {
    if (el && !heroLayerRefs.current.includes(el)) heroLayerRefs.current.push(el)
  }

  return (
    <div className="landing-body">
      <nav className="landing-nav">
        <Link to="/" className="logo">
          <img src="/logo-short.svg" alt="" className="logo-icon" />Khorcha-Pati
        </Link>
        <ul>
          <li><a href="#features">Features</a></li>
          <li><a href="#how">How it works</a></li>
          <li><a href="#commands">Commands</a></li>
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
              Keep your khorcha on track. Khorcha-Pati turns plain-English chat messages into structured ledger entries
              — categorized, balanced, and ready to review whenever you open the dashboard.
            </p>
            <p className="lead sub">
              Built for people who hate spreadsheets and love clarity. Track expenses, settle debts with friends,
              monitor budgets, and export beautifully-typeset PDF statements — all from your phone.
            </p>
            <div className="hero-cta">
              <a className="btn primary" href={BOT_URL} target="_blank" rel="noreferrer">Start on Telegram →</a>
              <Link className="btn ghost" to="/login">View Dashboard</Link>
            </div>
            <div className="trust">
              <span>Open source</span>
              <span>•</span>
              <span>Free, forever</span>
              <span>•</span>
              <span>Self-hostable</span>
            </div>
          </div>

          <div className="hero-stack" aria-hidden>
            <div ref={registerLayer} className="parallax bg-art" data-speed="0.08" />
            <div ref={registerLayer} className="parallax phone" data-speed="0.18">
              <div className="phone-screen">
                <div className="phone-head"><span className="dot" /> Khorcha-Pati Bot</div>
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
            <div ref={registerLayer} className="parallax card-floater bank" data-speed="0.26" data-base-rotate="-2deg" data-rotate="0.003">
              <span className="card-brand">Khorcha-Pati</span>
              <span className="card-num">•••• 4521</span>
              <span className="card-amt">৳ 37,630</span>
              <span className="card-name">BRAC BANK</span>
            </div>
            <div ref={registerLayer} className="parallax card-floater cash" data-speed="0.32" data-base-rotate="-6deg" data-rotate="0.005">
              <span className="card-brand">Khorcha-Pati</span>
              <span className="card-num">— — — —</span>
              <span className="card-amt">৳ 3,210</span>
              <span className="card-name">WALLET</span>
            </div>
          </div>
        </div>
      </section>

      <section className="band reveal" id="how">
        <span className="eyebrow">How it works</span>
        <h2>Designed for speed. Built for depth.</h2>
        <p>
          Stop wrestling with spreadsheets. Use the chat on the go, and the command center at home.
          Three steps from messy receipts to a clean, queryable ledger.
        </p>
        <div className="steps">
          <div className="step reveal" data-stagger="0">
            <span className="step-num">1</span>
            <h3>Type it like you'd say it</h3>
            <p>"Lunch 320", "Salary 52k", "Paid Karim 500 ricksha". The bot parses amount, date, wallet, and contact automatically.</p>
          </div>
          <div className="step reveal" data-stagger="120">
            <span className="step-num">2</span>
            <h3>AI picks the category</h3>
            <p>Gemini classifies your subcategory. The model caches frequent patterns so over time it gets faster and free-er.</p>
          </div>
          <div className="step reveal" data-stagger="240">
            <span className="step-num">3</span>
            <h3>Review &amp; reconcile</h3>
            <p>Open the dashboard whenever. Wallets, budgets, donut charts, debt circle, PDF statements — all in one place.</p>
          </div>
        </div>
      </section>

      <section className="features" id="features">
        <div className="feature reveal" data-stagger="0">
          <span className="emoji">💬</span>
          <h3>Write like you talk</h3>
          <p>
            Forget rigid forms. Khorcha-Pati understands free-text shorthand — "lunch 320", "paid karim 500", "salary 52k" — and parses
            amount, date, wallet, and contact into structured entries you can audit later.
          </p>
        </div>
        <div className="feature reveal" data-stagger="80">
          <span className="emoji">🤖</span>
          <h3>AI categorization</h3>
          <p>
            Gemini classifies subcategories with cached embeddings so the model gets faster and cheaper over time.
            Manual overrides train it without leaving the chat.
          </p>
        </div>
        <div className="feature reveal" data-stagger="160">
          <span className="emoji">🧾</span>
          <h3>PDF statements</h3>
          <p>
            Generate beautifully-typeset statements for any date range — weekly, monthly, custom — and share them as a single
            PDF. Server-rendered for print-perfect output.
          </p>
        </div>
        <div className="feature reveal" data-stagger="240">
          <span className="emoji">📊</span>
          <h3>Deep analytics</h3>
          <p>
            Wallets as credit-card visuals, donut charts by category, income-vs-expense bars, budget gauges with alerts,
            and a financial circle that tracks who owes whom.
          </p>
        </div>
        <div className="feature reveal" data-stagger="320">
          <span className="emoji">👥</span>
          <h3>Debt tracking with contacts</h3>
          <p>
            Log who paid for what, who owes whom, and what's settled. The contacts view shows net balances at a glance —
            then drill in to see the exact history of every shared transaction.
          </p>
        </div>
        <div className="feature reveal" data-stagger="400">
          <span className="emoji">🎯</span>
          <h3>Budgets &amp; alerts</h3>
          <p>
            Set monthly limits per category. Khorcha-Pati notifies you when you're approaching the cap and shows whether you're
            on track to save — or about to overshoot.
          </p>
        </div>
      </section>

      <section className="dashboard-preview reveal" id="dashboard">
        <div className="dash">
          <div className="dash-side">
            <img src="/logo-short.svg" alt="" />
            <span>Khorcha-Pati</span>
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
          <p>
            A web dashboard that gives the bot the big picture: balance trends, category spend, wallet flow, contact debts,
            budgets, and downloadable statements.
          </p>
          <ul className="bullets">
            <li>Credit-card-style wallets with multi-palette variants</li>
            <li>Donut + bar charts that read at a glance</li>
            <li>Dark mode, collapsible sidebar, mobile-first</li>
            <li>One-click PDF statement for any date range</li>
          </ul>
        </div>
      </section>

      <section className="commands reveal" id="commands">
        <span className="eyebrow">Bot Commands</span>
        <h2>The whole bot in nine commands.</h2>
        <div className="cmd-grid">
          <div className="cmd"><code>/new</code><span>Log an income, expense, or transfer in natural language.</span></div>
          <div className="cmd"><code>/list</code><span>Browse recent transactions with type and wallet filters.</span></div>
          <div className="cmd"><code>/wallets</code><span>See balances per wallet — cash, bank, mobile.</span></div>
          <div className="cmd"><code>/contacts</code><span>Track who owes whom and quickly settle debts.</span></div>
          <div className="cmd"><code>/budgets</code><span>Set monthly limits and get alerts before you overshoot.</span></div>
          <div className="cmd"><code>/summary</code><span>This month's income, expense, and category breakdown.</span></div>
          <div className="cmd"><code>/statement</code><span>Generate a PDF statement for any date range.</span></div>
          <div className="cmd"><code>/undo</code><span>Reverse the last transaction with a single tap.</span></div>
          <div className="cmd"><code>/help</code><span>Show command reference and natural-language tips.</span></div>
        </div>
      </section>

      <section className="cta">
        <h2>Ready to master your money?</h2>
        <p>Reconcile your finances in seconds, not hours. Free. Forever. Open source.</p>
        <div className="cta-actions">
          <a className="btn primary" href={BOT_URL} target="_blank" rel="noreferrer">Start tracking on Telegram</a>
          <Link className="btn ghost" to="/login">Explore Dashboard</Link>
        </div>
      </section>

      <footer>
        <div>
          <Link to="/" className="logo small">
            <img src="/logo-short.svg" alt="" /> Khorcha-Pati
          </Link>
          <p>Keep your khorcha on track.</p>
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
        <div className="copy">© {new Date().getFullYear()} Khorcha-Pati.</div>
      </footer>
    </div>
  )
}
