import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import './Landing.css'

export default function Landing() {
  useEffect(() => {
    const obs = new IntersectionObserver(
      (entries) => {
        entries.forEach((e) => {
          if (e.isIntersecting) {
            e.target.classList.add('visible')
            obs.unobserve(e.target)
          }
        })
      },
      { threshold: 0.15 }
    )
    document.querySelectorAll('.fade-up').forEach((el) => obs.observe(el))

    return () => obs.disconnect()
  }, [])

  return (
    <div className="landing-body">
      {/* NAV */}
      <nav className="landing-nav">
        <Link to="/" className="logo">
          <img src="/logo-short.svg" alt="" className="logo-icon" />Hisab
        </Link>
        <ul>
          <li>
            <a href="#features">Features</a>
          </li>
          <li>
            <a href="#commands">Commands</a>
          </li>
          <li>
            <a href="#dashboard">Dashboard</a>
          </li>
          <li>
            <Link to="/login">Sign In</Link>
          </li>
        </ul>
        <button
          className="nav-cta"
          onClick={() => window.open('https://t.me/XpenseTrackerBot', '_blank')}
        >
          Open in Telegram
        </button>
      </nav>

      {/* HERO */}
      <section className="hero">
        <div className="hero-inner">
          <div>
            <h1>
              Your finances,
              <br />
              one chat &<br />
              <span className="hl">one dashboard.</span>
            </h1>
            <p>
              Track expenses with plain English in Telegram. View charts, download statements, and
              manage wallets from a full web dashboard.
            </p>
            <div className="hero-actions">
              <a
                href="https://t.me/XpenseTrackerBot"
                target="_blank"
                rel="noreferrer"
                className="btn-primary"
              >
                <svg
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M21.2 4.6 2.4 11.1c-.7.3-.7 1.3 0 1.5l4.3 1.4 1.6 5.1c.2.6 1 .8 1.4.3l2.3-2.6 4.5 3.3c.5.4 1.3.1 1.4-.5L21.8 5.5c.2-.8-.6-1.3-1.3-1z" />
                </svg>
                Start on Telegram
              </a>
              <Link to="/login" className="btn-secondary">
                <svg
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <rect x="3" y="3" width="7" height="7" rx="1.5" />
                  <rect x="14" y="3" width="7" height="7" rx="1.5" />
                  <rect x="3" y="14" width="7" height="7" rx="1.5" />
                  <rect x="14" y="14" width="7" height="7" rx="1.5" />
                </svg>
                View Dashboard
              </Link>
            </div>
          </div>

          {/* Dual mockup: phone + dashboard */}
          <div className="hero-mockup">
            <div className="phone-frame">
              <div className="phone-screen">
                <div className="chat-header">
                  <div className="chat-avatar">X</div>
                  <div>
                    <div className="chat-name">XpenseTrackerBot</div>
                    <div className="chat-sub">online</div>
                  </div>
                </div>
                <div className="msg msg-user">lunch 250</div>
                <div className="msg msg-bot">
                  ✅ Food → Restaurant
                  <br />৳250 from Cash
                </div>
                <div className="msg msg-user">got bonus 20k</div>
                <div className="msg msg-bot">
                  ✅ Income — ৳20,000
                  <br />
                  Added to BRAC Bank
                </div>
                <div className="msg msg-user">/balance</div>
                <div className="msg msg-bot">
                  💳 Account Balances
                  <br />
                  <br />
                  Cash — ৳4,500
                  <br />
                  BRAC — ৳23,800
                  <br />
                  City — ৳12,100
                </div>
                <div className="msg msg-user">/summary</div>
              </div>
            </div>
            <div className="dash-frame">
              <div className="dash-sidebar">
                <img src="/logo-short.svg" alt="" className="dot" />
                <span>Hisab</span>
              </div>
              <div className="dash-body">
                <div className="dash-stats">
                  <div className="dash-stat">
                    <div className="dash-stat-label">Balance</div>
                    <div className="dash-stat-val" style={{ color: 'oklch(0.45 0.18 260)' }}>
                      ৳40,400
                    </div>
                  </div>
                  <div className="dash-stat">
                    <div className="dash-stat-label">Income</div>
                    <div className="dash-stat-val" style={{ color: 'oklch(0.45 0.18 155)' }}>
                      ৳52,000
                    </div>
                  </div>
                  <div className="dash-stat">
                    <div className="dash-stat-label">Expense</div>
                    <div className="dash-stat-val" style={{ color: 'oklch(0.50 0.18 25)' }}>
                      ৳18,450
                    </div>
                  </div>
                  <div className="dash-stat">
                    <div className="dash-stat-label">Budget</div>
                    <div className="dash-stat-val">62%</div>
                  </div>
                </div>
                <div className="dash-chart">
                  <div className="bar bar-inc" style={{ height: '60%' }}></div>
                  <div className="bar bar-exp" style={{ height: '35%' }}></div>
                  <div className="bar bar-inc" style={{ height: '45%' }}></div>
                  <div className="bar bar-exp" style={{ height: '50%' }}></div>
                  <div className="bar bar-inc" style={{ height: '70%' }}></div>
                  <div className="bar bar-exp" style={{ height: '30%' }}></div>
                  <div className="bar bar-inc" style={{ height: '55%' }}></div>
                  <div className="bar bar-exp" style={{ height: '45%' }}></div>
                  <div className="bar bar-inc" style={{ height: '80%' }}></div>
                  <div className="bar bar-exp" style={{ height: '40%' }}></div>
                </div>
                <div className="dash-txn-list">
                  <div className="dash-txn">
                    <span className="dash-txn-cat">Team Lunch</span>
                    <span className="dash-txn-amt neg">-৳1,200</span>
                  </div>
                  <div className="dash-txn">
                    <span className="dash-txn-cat">Salary</span>
                    <span className="dash-txn-amt pos">+৳52,000</span>
                  </div>
                  <div className="dash-txn">
                    <span className="dash-txn-cat">Groceries</span>
                    <span className="dash-txn-amt neg">-৳1,500</span>
                  </div>
                  <div className="dash-txn">
                    <span className="dash-txn-cat">New Shirt</span>
                    <span className="dash-txn-amt neg">-৳2,500</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* NATURAL LANGUAGE INPUT */}
      <section className="section" id="features">
        <div className="section-label">Simple Input</div>
        <div className="section-title fade-up">
          Just type what you did.
          <br />
          No forms. No menus.
        </div>
        <div className="section-subtitle fade-up">
          Send a plain text message to the Telegram bot — it understands amounts, categories,
          wallets, contacts, and dates automatically.
        </div>

        <div className="input-showcase">
          <div className="input-examples fade-up">
            <div className="input-ex">
              <span className="arrow">→</span> lunch 250
            </div>
            <div className="input-ex">
              <span className="arrow">→</span> groceries 1.5k
            </div>
            <div className="input-ex">
              <span className="arrow">→</span> transfer 10k from brac to city
            </div>
            <div className="input-ex">
              <span className="arrow">→</span> lent 5000 to karim
            </div>
            <div className="input-ex">
              <span className="arrow">→</span> got bonus 20k
            </div>
          </div>
          <div className="input-desc fade-up">
            <h3>Write like you talk.</h3>
            <p>
              No commands to memorize. The bot parses shorthand amounts like <strong>1.5k</strong>,
              figures out the category from context, picks the right wallet, and handles lending and
              borrowing between contacts.
            </p>
            <p style={{ marginTop: '14px' }}>
              Expenses, income, transfers, loans — all from a single line of text.
            </p>
          </div>
        </div>
      </section>

      {/* FEATURES GRID */}
      <section className="section">
        <div className="section-label">Capabilities</div>
        <div className="section-title fade-up">
          Bot + Dashboard.
          <br />
          Everything covered.
        </div>
        <div className="features-grid">
          <div className="feature-card fade-up">
            <div className="feature-icon fi-blue">
              <svg
                width="22"
                height="22"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
              </svg>
            </div>
            <h3>Natural Language Tracking</h3>
            <p>
              Just describe your transaction in plain text. The bot parses amounts, categories,
              wallets, and dates instantly.
            </p>
          </div>
          <div className="feature-card fade-up">
            <div className="feature-icon fi-green">
              <svg
                width="22"
                height="22"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <path d="M3 9h18M9 21V9" />
              </svg>
            </div>
            <h3>Visual Dashboard</h3>
            <p>
              Charts for expense breakdown, income vs expense comparison, budget gauges, and
              categorized summaries at a glance.
            </p>
          </div>
          <div className="feature-card fade-up">
            <div className="feature-icon fi-blue">
              <svg
                width="22"
                height="22"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                <polyline points="14 2 14 8 20 8" />
              </svg>
            </div>
            <h3>PDF Statements</h3>
            <p>
              Generate and download professional financial statements from the dashboard for any
              time period.
            </p>
          </div>
          <div className="feature-card fade-up">
            <div className="feature-icon fi-green">
              <svg
                width="22"
                height="22"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                <circle cx="9" cy="7" r="4" />
                <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
              </svg>
            </div>
            <h3>Contacts & Lending</h3>
            <p>
              Track who owes you and who you owe. Manage contacts with net balance tracking and full
              history.
            </p>
          </div>
          <div className="feature-card fade-up">
            <div className="feature-icon fi-blue">
              <svg
                width="22"
                height="22"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
              </svg>
            </div>
            <h3>Multi-Wallet</h3>
            <p>
              Manage cash, bank accounts, and cards. Real-time balance tracking with seamless
              transfers between wallets.
            </p>
          </div>
          <div className="feature-card fade-up">
            <div className="feature-icon fi-green">
              <svg
                width="22"
                height="22"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              </svg>
            </div>
            <h3>Self-Hosted & Private</h3>
            <p>
              Open source under Apache 2.0. Deploy on your own infrastructure with full control over
              your data.
            </p>
          </div>
        </div>
      </section>

      {/* COMMANDS */}
      <section className="section" id="commands">
        <div className="section-label">Bot Commands</div>
        <div className="section-title fade-up">
          Quick commands for
          <br />
          everything else.
        </div>
        <div className="section-subtitle fade-up">
          Beyond natural language, these slash commands give you instant access to summaries,
          reports, and more.
        </div>
        <div className="commands-grid">
          <div className="cmd-row fade-up">
            <span className="cmd-code">/balance</span>
            <span className="cmd-desc">View balances across all your accounts</span>
          </div>
          <div className="cmd-row fade-up">
            <span className="cmd-code">/expense</span>
            <span className="cmd-desc">Fetch expenses for the current month</span>
          </div>
          <div className="cmd-row fade-up">
            <span className="cmd-code">/summary</span>
            <span className="cmd-desc">Transaction summary for the current month</span>
          </div>
          <div className="cmd-row fade-up">
            <span className="cmd-code">/report</span>
            <span className="cmd-desc">Generate a PDF report for any time period</span>
          </div>
          <div className="cmd-row fade-up">
            <span className="cmd-code">/newtxn</span>
            <span className="cmd-desc">Add a transaction interactively with guided prompts</span>
          </div>
          <div className="cmd-row fade-up">
            <span className="cmd-code">/new</span>
            <span className="cmd-desc">Add accounts, debtors, or creditors</span>
          </div>
          <div className="cmd-row fade-up">
            <span className="cmd-code">/users</span>
            <span className="cmd-desc">List contacts involved in lending or borrowing</span>
          </div>
          <div className="cmd-row fade-up">
            <span className="cmd-code">/cat</span>
            <span className="cmd-desc">Browse all transaction categories</span>
          </div>
        </div>
      </section>

      {/* LOGIN METHODS */}
      <section className="login-section" id="login">
        <div className="section-label">Authentication</div>
        <div className="section-title fade-up">
          Three ways to sign in.
          <br />
          All through Telegram.
        </div>
        <div className="section-subtitle fade-up">
          Your Telegram identity is your login. No passwords to remember, no accounts to create.
        </div>
        <div className="login-methods">
          <div className="login-card fade-up">
            <div className="login-card-icon">
              <svg width="26" height="26" fill="none" stroke="white" strokeWidth="2" viewBox="0 0 24 24">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                <circle cx="12" cy="7" r="4" />
              </svg>
            </div>
            <h3>Username / Phone</h3>
            <p>
              Enter your Telegram username or mobile number. Receive a one-time code via the bot to
              verify.
            </p>
          </div>
          <div className="login-card fade-up">
            <div className="login-card-icon">
              <svg width="26" height="26" fill="none" stroke="white" strokeWidth="2" viewBox="0 0 24 24">
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <path d="M7 7h.01M7 12h.01M7 17h.01M12 7h.01M12 12h.01M12 17h.01M17 7h.01M17 12h.01M17 17h.01" />
              </svg>
            </div>
            <h3>QR Code Scan</h3>
            <p>
              Scan a QR code with your Telegram app for instant passwordless login. Approved right
              from your phone.
            </p>
          </div>
          <div className="login-card fade-up">
            <div className="login-card-icon">
              <svg width="26" height="26" fill="none" stroke="white" strokeWidth="2" viewBox="0 0 24 24">
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
              </svg>
            </div>
            <h3>Magic Link</h3>
            <p>
              Click a secure link sent directly to your Telegram chat. One tap and you're in the
              dashboard.
            </p>
          </div>
        </div>
      </section>

      {/* DASHBOARD FEATURES */}
      <section className="section" id="dashboard">
        <div className="section-label">Web Dashboard</div>
        <div className="section-title fade-up">
          Full control from
          <br />
          your browser.
        </div>
        <div className="section-subtitle fade-up">
          Everything the bot can do, plus visual charts, filterable tables, and downloadable
          reports.
        </div>

        <div className="dash-features">
          <div className="dash-feat fade-up">
            <div className="dash-feat-icon fi-blue">
              <svg
                width="20"
                height="20"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <path d="M18 20V10M12 20V4M6 20v-6" />
              </svg>
            </div>
            <div>
              <h3>Charts & Analytics</h3>
              <p>
                Expense donut by category, income vs expense bar charts, and budget usage gauges —
                all updating in real time.
              </p>
            </div>
          </div>
          <div className="dash-feat fade-up">
            <div className="dash-feat-icon fi-green">
              <svg
                width="20"
                height="20"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                <polyline points="14 2 14 8 20 8" />
                <line x1="16" y1="13" x2="8" y2="13" />
                <line x1="16" y1="17" x2="8" y2="17" />
              </svg>
            </div>
            <div>
              <h3>Transaction History</h3>
              <p>
                Browse, filter by type, add, edit, or delete transactions. Full CRUD with inline
                wallet and contact selectors.
              </p>
            </div>
          </div>
          <div className="dash-feat fade-up">
            <div className="dash-feat-icon fi-blue">
              <svg
                width="20"
                height="20"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <rect x="2" y="5" width="20" height="14" rx="2" />
                <path d="M2 10h20" />
              </svg>
            </div>
            <div>
              <h3>Wallets & Contacts</h3>
              <p>
                Add bank accounts and cash wallets. Manage contacts with net balance tracking and
                transaction history.
              </p>
            </div>
          </div>
          <div className="dash-feat fade-up">
            <div className="dash-feat-icon fi-green">
              <svg
                width="20"
                height="20"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <circle cx="12" cy="12" r="3" />
                <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
              </svg>
            </div>
            <div>
              <h3>Profile & Settings</h3>
              <p>
                Update your mobile number, timezone, and preferences. Linked to your Telegram
                identity for seamless sync.
              </p>
            </div>
          </div>
          <div className="dash-feat fade-up">
            <div className="dash-feat-icon fi-blue">
              <svg
                width="20"
                height="20"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
              </svg>
            </div>
            <div>
              <h3>Download Statements</h3>
              <p>
                Generate PDF reports for this month, last 30 days, 6 months, or the full year. View
                and download directly.
              </p>
            </div>
          </div>
          <div className="dash-feat fade-up">
            <div className="dash-feat-icon fi-green">
              <svg
                width="20"
                height="20"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                viewBox="0 0 24 24"
              >
                <circle cx="12" cy="12" r="10" />
                <path d="M12 6v6l4 2" />
              </svg>
            </div>
            <div>
              <h3>Budgets & Targets</h3>
              <p>
                Set monthly spending limits and track budget usage with visual gauges. Stay on top
                of your financial goals.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="cta-section">
        <h2 className="fade-up">
          Start tracking today.
          <br />
          It's free, forever.
        </h2>
        <p className="fade-up">Open source. No ads. No data selling. Your finances, your control.</p>
        <div className="hero-actions fade-up">
          <a
            href="https://t.me/XpenseTrackerBot"
            target="_blank"
            rel="noreferrer"
            className="btn-primary"
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M21.2 4.6 2.4 11.1c-.7.3-.7 1.3 0 1.5l4.3 1.4 1.6 5.1c.2.6 1 .8 1.4.3l2.3-2.6 4.5 3.3c.5.4 1.3.1 1.4-.5L21.8 5.5c.2-.8-.6-1.3-1.3-1z" />
            </svg>
            Open in Telegram
          </a>
          <a
            href="https://github.com/masudur-rahman/expense-tracker-bot"
            target="_blank"
            rel="noreferrer"
            className="btn-secondary"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.2 11.39.6.11.82-.26.82-.58v-2.03c-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.34-1.76-1.34-1.76-1.09-.75.08-.73.08-.73 1.21.08 1.85 1.24 1.85 1.24 1.07 1.84 2.81 1.31 3.5 1 .11-.78.42-1.31.76-1.61-2.67-.3-5.47-1.33-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.13-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23a11.5 11.5 0 0 1 6.02 0c2.28-1.55 3.29-1.23 3.29-1.23.66 1.66.25 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.63-5.48 5.92.43.37.81 1.1.81 2.22v3.29c0 .32.22.7.82.58A12.01 12.01 0 0 0 24 12C24 5.37 18.63 0 12 0z" />
            </svg>
            View Source
          </a>
        </div>
      </section>

      <footer className="landing-footer">
        <p>
          Built by{' '}
          <a
            href="https://t.me/masudur_rahman"
            target="_blank"
            rel="noreferrer"
            style={{ color: 'var(--accent)', textDecoration: 'none' }}
          >
            masudur-rahman
          </a>{' '}
          · Apache 2.0 License ·{' '}
          <a
            href="https://github.com/masudur-rahman/expense-tracker-bot"
            target="_blank"
            rel="noreferrer"
            style={{ color: 'var(--fg-muted)', textDecoration: 'none' }}
          >
            GitHub
          </a>
        </p>
      </footer>
    </div>
  )
}
