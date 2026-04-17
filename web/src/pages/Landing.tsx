import { Link } from 'react-router-dom'
import { Bot, BarChart3, ShieldCheck, MessageSquare, Zap, Globe, Layout } from 'lucide-react'

export default function Landing() {
  return (
    <div className="min-h-screen bg-slate-50">
      {/* Navigation */}
      <nav className="flex items-center justify-between px-6 py-6 max-w-7xl mx-auto">
        <div className="flex items-center space-x-2">
          <div className="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center text-white font-bold shadow-lg shadow-blue-200">E</div>
          <span className="text-2xl font-bold text-slate-900 tracking-tight">Expense<span className="text-blue-600"> Tracker</span></span>
        </div>
        <div className="hidden md:flex items-center space-x-8">
          <a href="#features" className="text-sm font-semibold text-slate-600 hover:text-blue-600 transition-colors">Features</a>
          <Link to="/login" className="text-sm font-bold text-slate-900 bg-white px-6 py-2.5 rounded-full border border-slate-200 hover:border-blue-600 hover:text-blue-600 transition-all shadow-sm">
            Sign In
          </Link>
          <a
            href="https://t.me/expense_tracker_bot"
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm font-bold text-white bg-blue-600 px-6 py-2.5 rounded-full hover:bg-blue-700 transition-all shadow-lg shadow-blue-200 flex items-center gap-2"
          >
            Launch on Telegram
          </a>
        </div>
      </nav>

      {/* Hero Section */}
      <header className="px-6 pt-20 pb-28 text-center max-w-5xl mx-auto">
        <div className="inline-flex items-center space-x-2 bg-blue-50 text-blue-700 px-4 py-2 rounded-full mb-10 border border-blue-100 shadow-sm">
          <Zap size={14} className="fill-blue-600" />
          <span className="text-[10px] font-bold uppercase tracking-[0.2em]">Next-Gen Wealth Management</span>
        </div>
        <h1 className="text-6xl md:text-8xl font-bold text-slate-900 mb-8 leading-[1.1] tracking-tight">
          Financial Clarity, <br />
          <span className="text-blue-600">One Chat Away.</span>
        </h1>
        <p className="text-lg md:text-2xl text-slate-500 mb-14 max-w-3xl mx-auto leading-relaxed font-medium">
          A powerful, Go-powered finance engine that lives in your Telegram. Track every penny with natural language commands and take full control of your wealth across all accounts.
        </p>
        <div className="flex flex-col sm:flex-row justify-center items-center gap-5">
          <Link
            to="/login"
            className="w-full sm:w-auto px-12 py-5 text-lg font-bold text-white bg-slate-900 rounded-[2rem] hover:bg-black shadow-2xl shadow-slate-200 transition-all transform hover:-translate-y-1"
          >
            Open Dashboard
          </Link>
          <a
            href="https://t.me/expense_tracker_bot"
            target="_blank"
            rel="noopener noreferrer"
            className="w-full sm:w-auto px-12 py-5 text-lg font-bold text-blue-600 bg-white border-2 border-blue-100 rounded-[2rem] hover:border-blue-600 transition-all flex items-center justify-center gap-3 transform hover:-translate-y-1"
          >
            <MessageSquare size={22} />
            Start Chatting
          </a>
        </div>
      </header>

      {/* Primary Features Grid */}
      <section id="features" className="px-6 py-24 bg-white">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-24">
            <h2 className="text-4xl md:text-5xl font-bold text-slate-900 mb-6">Built for Modern Finances.</h2>
            <p className="text-slate-500 text-lg font-medium max-w-2xl mx-auto">Sophisticated tracking capabilities simplified into a conversational interface.</p>
          </div>
          
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            <FeatureCard 
              icon={<Bot className="text-blue-600" />}
              title="Conversational AI"
              description="Just say 'Spent 50$ on Groceries'. Our engine parses your natural language to record expenses, income, and transfers instantly."
            />
            <FeatureCard 
              icon={<Layout className="text-emerald-600" />}
              title="Multi-Wallet Sync"
              description="Manage cash, bank accounts, and investments in one place. Real-time balance updates across your entire portfolio."
            />
            <FeatureCard 
              icon={<BarChart3 className="text-orange-600" />}
              title="Proactive Budgeting"
              description="Set monthly targets and get smart alerts before you overspend. Stay on top of your financial goals automatically."
            />
            <FeatureCard 
              icon={<ShieldCheck className="text-indigo-600" />}
              title="Self-Hosted Privacy"
              description="Your data, your rules. Built for secure self-hosting with an automation-friendly workflow that puts your privacy first."
            />
          </div>
        </div>
      </section>

      {/* Statement Preview Section */}
      <section className="px-6 py-24 bg-slate-50 border-y border-slate-100">
        <div className="max-w-7xl mx-auto grid lg:grid-cols-2 gap-16 items-center">
            <div className="space-y-8">
                <h2 className="text-4xl md:text-5xl font-bold text-slate-900 leading-tight">Professional Statements, <br /><span className="text-blue-600">Generated Instantly.</span></h2>
                <p className="text-lg text-slate-600 font-medium leading-relaxed">
                    Need a deep dive into your spending? Generate comprehensive PDF statements directly from the dashboard. See categorized summaries, wallet breakdowns, and full transaction histories in a single, clean document.
                </p>
                <ul className="space-y-4">
                    <li className="flex items-center gap-3 font-bold text-slate-700">
                        <div className="w-6 h-6 bg-blue-100 rounded-full flex items-center justify-center text-blue-600"><Zap size={12} /></div>
                        Detailed Transaction Histories
                    </li>
                    <li className="flex items-center gap-3 font-bold text-slate-700">
                        <div className="w-6 h-6 bg-emerald-100 rounded-full flex items-center justify-center text-emerald-600"><Zap size={12} /></div>
                        Wallet & Contact Net Balances
                    </li>
                    <li className="flex items-center gap-3 font-bold text-slate-700">
                        <div className="w-6 h-6 bg-orange-100 rounded-full flex items-center justify-center text-orange-600"><Zap size={12} /></div>
                        Categorized Monthly Summaries
                    </li>
                </ul>
            </div>
            <div className="bg-white p-4 rounded-[2.5rem] shadow-2xl shadow-slate-200 border border-slate-100 transform rotate-2">
                <div className="bg-slate-50 rounded-[2rem] p-10 aspect-[3/4] flex flex-col justify-center items-center text-center space-y-6">
                    <FileText size={64} className="text-slate-200" />
                    <p className="text-slate-400 font-bold uppercase tracking-widest text-xs">Sample Statement View</p>
                </div>
            </div>
        </div>
      </section>

      {/* Stats/Social Proof */}
      <section className="bg-slate-900 py-24 text-white overflow-hidden relative">
         <div className="absolute top-0 right-0 -translate-y-1/2 translate-x-1/2 w-[500px] h-[500px] bg-blue-600 rounded-full blur-[140px] opacity-10"></div>
         <div className="max-w-7xl mx-auto px-6 grid md:grid-cols-3 gap-16 text-center relative z-10">
            <StatItem value="Zero" label="Learning Curve" />
            <StatItem value="100%" label="Private & Self-Hosted" />
            <StatItem value="Instant" label="Telegram Sync" />
         </div>
      </section>

      {/* Footer */}
      <footer className="px-6 py-16 bg-white border-t border-slate-100">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-center gap-8">
          <div className="flex items-center space-x-2">
            <div className="w-9 h-9 bg-slate-900 rounded-xl flex items-center justify-center text-white font-bold">E</div>
            <span className="text-2xl font-bold text-slate-900 tracking-tight">Expense Tracker</span>
          </div>
          <div className="flex flex-col items-center md:items-end gap-2">
            <p className="text-slate-500 font-medium text-sm">Built with Go for speed and privacy.</p>
            <p className="text-slate-300 text-xs font-bold uppercase tracking-widest">© 2026 Expense Tracker</p>
          </div>
          <div className="flex space-x-8">
            <a href="#" className="text-slate-400 hover:text-blue-600 transition-colors"><Globe size={24} /></a>
            <a href="https://t.me/expense_tracker_bot" className="text-slate-400 hover:text-blue-600 transition-colors"><MessageSquare size={24} /></a>
          </div>
        </div>
      </footer>
    </div>
  )
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode, title: string, description: string }) {
  return (
    <div className="p-10 rounded-[2.5rem] border border-slate-50 bg-white hover:shadow-2xl hover:shadow-slate-200/50 transition-all group">
      <div className="w-16 h-16 bg-slate-50 rounded-2xl flex items-center justify-center mb-8 group-hover:scale-110 group-hover:bg-blue-50 transition-all duration-500">
        {icon}
      </div>
      <h3 className="text-xl font-bold text-slate-900 mb-4">{title}</h3>
      <p className="text-slate-500 leading-relaxed font-medium text-sm">
        {description}
      </p>
    </div>
  )
}

function StatItem({ value, label }: { value: string, label: string }) {
    return (
        <div className="space-y-2">
            <p className="text-5xl font-bold tracking-tighter text-blue-400">{value}</p>
            <p className="text-slate-400 font-bold uppercase tracking-[0.2em] text-[10px]">{label}</p>
        </div>
    )
}

function FileText({ size, className }: { size: number, className: string }) {
    return (
        <svg 
            xmlns="http://www.w3.org/2000/svg" 
            width={size} 
            height={size} 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2" 
            strokeLinecap="round" 
            strokeLinejoin="round" 
            className={className}
        >
            <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
            <line x1="10" y1="9" x2="8" y2="9"/>
        </svg>
    )
}
