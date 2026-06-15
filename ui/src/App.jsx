import React, { useState, useEffect, useCallback } from 'react'
import axios from 'axios'

const API = 'http://localhost:8090'

const EVENT_TYPES = [
  'SAMPLE_CREATED',
  'RESULT_SUBMITTED',
  'RESULT_VERIFIED',
  'REPORT_PUBLISHED',
  'DATA_MODIFIED',
  'INSTRUMENT_IMPORT',
]

const EMPTY_FORM = {
  lims_record_id: '',
  event_type: 'RESULT_VERIFIED',
  sample_id: '',
  result: '',
  user_id: '',
  status: 'VERIFIED',
  timestamp: '',
}

// ─── Дашборд ─────────────────────────────────────────────────────────────────

function Dashboard({ events, adapterInfo }) {
  const verified = events.filter(e => e.status === 'VERIFIED').length

  return (
    <div className="page">
      <h2 className="page-title">Дашборд</h2>
      <p className="page-sub">
        Обзор системы верификации целостности лабораторных данных на основе блокчейна
      </p>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-value">{events.length}</div>
          <div className="stat-label">Всего хэш-записей</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{verified}</div>
          <div className="stat-label">Верифицированных событий</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">SHA-256</div>
          <div className="stat-label">Модель целостности</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">Активен</div>
          <div className="stat-label">Статус адаптера</div>
        </div>
      </div>

      <div className="card">
        <h3 className="card-title">Архитектура системы</h3>
        <div className="arch-flow">
          <div className="arch-node">
            <div className="arch-icon">🔬</div>
            <div className="arch-name">ПК / Приборы</div>
            <div className="arch-desc">Лабораторные инструменты</div>
          </div>
          <div className="arch-arrow">→</div>
          <div className="arch-node arch-primary">
            <div className="arch-icon">🗄️</div>
            <div className="arch-name">SENAITE LIMS</div>
            <div className="arch-desc">Хранит лабораторные данные</div>
          </div>
          <div className="arch-arrow">→</div>
          <div className="arch-node arch-primary">
            <div className="arch-icon">⚙️</div>
            <div className="arch-name">Go Адаптер</div>
            <div className="arch-desc">Интеллектуальный слой управления</div>
          </div>
          <div className="arch-arrow">→</div>
          <div className="arch-node">
            <div className="arch-icon">📜</div>
            <div className="arch-name">Смарт-контракт</div>
            <div className="arch-desc">LimsHashRegistry</div>
          </div>
          <div className="arch-arrow">→</div>
          <div className="arch-node arch-primary">
            <div className="arch-icon">⛓️</div>
            <div className="arch-name">Блокчейн</div>
            <div className="arch-desc">Неизменяемый слой доверия</div>
          </div>
        </div>
      </div>

      <div className="card">
        <h3 className="card-title">Статус адаптера</h3>
        {adapterInfo ? (
          <div className="info-grid">
            <div className="info-row">
              <span className="info-key">Сервис</span>
              <span className="info-val">{adapterInfo.service}</span>
            </div>
            <div className="info-row">
              <span className="info-key">Роль</span>
              <span className="info-val">{adapterInfo.role}</span>
            </div>
            <div className="info-row">
              <span className="info-key">Модель целостности</span>
              <span className="info-val badge">{adapterInfo.integrity_model}</span>
            </div>
            <div className="info-row">
              <span className="info-key">Blockchain RPC</span>
              <span className="info-val mono">{adapterInfo.blockchain_rpc}</span>
            </div>
            <div className="info-row">
              <span className="info-key">Адрес контракта</span>
              <span className="info-val mono">
                {adapterInfo.contract_address || 'Не задеплоен (локальный режим)'}
              </span>
            </div>
          </div>
        ) : (
          <p className="muted">Подключение к адаптеру…</p>
        )}
      </div>
    </div>
  )
}

// ─── Регистрация события ───────────────────────────────────────────────────────

function RegisterEvent({ onSuccess }) {
  const [form, setForm] = useState(EMPTY_FORM)
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)

  const handleChange = e =>
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }))

  const handleSubmit = async e => {
    e.preventDefault()
    setLoading(true)
    setResult(null)
    try {
      const payload = {
        ...form,
        timestamp: form.timestamp || new Date().toISOString(),
      }
      const res = await axios.post(`${API}/events/hash`, payload)
      setResult({ ok: true, data: res.data })
      onSuccess()
    } catch (err) {
      setResult({ ok: false, data: err.response?.data || { error: err.message } })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <h2 className="page-title">Регистрация события</h2>
      <p className="page-sub">
        Отправьте лабораторное событие для вычисления хэша и регистрации в блокчейне.
        Полные данные остаются в LIMS — в блокчейн записывается только SHA-256 хэш.
      </p>

      <div className="card">
        <form className="form" onSubmit={handleSubmit}>
          <div className="form-row">
            <label>ID записи LIMS *</label>
            <input
              name="lims_record_id"
              value={form.lims_record_id}
              onChange={handleChange}
              placeholder="SAMPLE-2026-001"
              required
            />
          </div>
          <div className="form-row">
            <label>Тип события *</label>
            <select name="event_type" value={form.event_type} onChange={handleChange}>
              {EVENT_TYPES.map(t => <option key={t}>{t}</option>)}
            </select>
          </div>
          <div className="form-row">
            <label>ID пробы</label>
            <input
              name="sample_id"
              value={form.sample_id}
              onChange={handleChange}
              placeholder="S-001"
            />
          </div>
          <div className="form-row">
            <label>Результат</label>
            <input
              name="result"
              value={form.result}
              onChange={handleChange}
              placeholder="pH=7.2"
            />
          </div>
          <div className="form-row">
            <label>ID пользователя</label>
            <input
              name="user_id"
              value={form.user_id}
              onChange={handleChange}
              placeholder="lab_user_01"
            />
          </div>
          <div className="form-row">
            <label>Статус</label>
            <input
              name="status"
              value={form.status}
              onChange={handleChange}
              placeholder="VERIFIED"
            />
          </div>
          <div className="form-row">
            <label>Временная метка (ISO 8601)</label>
            <input
              name="timestamp"
              value={form.timestamp}
              onChange={handleChange}
              placeholder="2026-06-15T12:00:00Z"
            />
            <span className="hint">Оставьте пустым — будет использовано текущее время</span>
          </div>
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Регистрация…' : 'Зарегистрировать хэш в блокчейне'}
          </button>
        </form>
      </div>

      {result && (
        <div className={`alert ${result.ok ? 'alert-success' : 'alert-error'}`}>
          {result.ok ? (
            <>
              <div className="alert-title">Хэш успешно зарегистрирован</div>
              <div className="mono small">{result.data.data_hash}</div>
              <div className="small muted">
                Запись LIMS: {result.data.lims_record_id} | Событие: {result.data.event_type}
              </div>
            </>
          ) : (
            <>
              <div className="alert-title">Ошибка регистрации</div>
              <div className="small">{JSON.stringify(result.data)}</div>
            </>
          )}
        </div>
      )}
    </div>
  )
}

// ─── Реестр хэшей ─────────────────────────────────────────────────────────────

function HashRegistry({ events, loading, onRefresh }) {
  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2 className="page-title">Реестр хэшей</h2>
          <p className="page-sub">Все хэш-записи, сохранённые в блокчейне</p>
        </div>
        <button className="btn btn-secondary" onClick={onRefresh} disabled={loading}>
          {loading ? 'Загрузка…' : 'Обновить'}
        </button>
      </div>

      <div className="card">
        {events.length === 0 ? (
          <p className="muted">Записей пока нет. Используйте «Регистрация события» для добавления.</p>
        ) : (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>ID записи LIMS</th>
                  <th>Тип события</th>
                  <th>Хэш данных</th>
                  <th>Статус</th>
                  <th>Зарегистрировано</th>
                </tr>
              </thead>
              <tbody>
                {events.map(r => (
                  <tr key={r.lims_record_id}>
                    <td className="mono">{r.lims_record_id}</td>
                    <td><span className="badge">{r.event_type}</span></td>
                    <td className="mono small hash-cell" title={r.data_hash}>
                      {r.data_hash.slice(0, 16)}…
                    </td>
                    <td>{r.status}</td>
                    <td className="small">
                      {new Date(r.registered_at).toLocaleString('ru-RU')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

// ─── Верификация данных ────────────────────────────────────────────────────────

function VerifyData() {
  const [form, setForm] = useState(EMPTY_FORM)
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)

  const handleChange = e =>
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }))

  const handleSubmit = async e => {
    e.preventDefault()
    setLoading(true)
    setResult(null)
    try {
      const res = await axios.post(`${API}/events/verify`, form)
      setResult({ ok: true, data: res.data })
    } catch (err) {
      if (err.response?.status === 404) {
        setResult({ ok: true, data: err.response.data })
      } else {
        setResult({ ok: false, data: err.response?.data || { error: err.message } })
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <h2 className="page-title">Верификация целостности данных</h2>
      <p className="page-sub">
        Повторно вычисляет SHA-256 хэш из данных события и сравнивает с хэшем,
        сохранённым в блокчейне. Несовпадение означает, что данные в LIMS могли быть изменены.
      </p>

      <div className="card">
        <form className="form" onSubmit={handleSubmit}>
          <div className="form-row">
            <label>ID записи LIMS *</label>
            <input
              name="lims_record_id"
              value={form.lims_record_id}
              onChange={handleChange}
              placeholder="SAMPLE-2026-001"
              required
            />
          </div>
          <div className="form-row">
            <label>Тип события *</label>
            <select name="event_type" value={form.event_type} onChange={handleChange}>
              {EVENT_TYPES.map(t => <option key={t}>{t}</option>)}
            </select>
          </div>
          <div className="form-row">
            <label>ID пробы</label>
            <input name="sample_id" value={form.sample_id} onChange={handleChange} placeholder="S-001" />
          </div>
          <div className="form-row">
            <label>Результат</label>
            <input name="result" value={form.result} onChange={handleChange} placeholder="pH=7.2" />
          </div>
          <div className="form-row">
            <label>ID пользователя</label>
            <input name="user_id" value={form.user_id} onChange={handleChange} placeholder="lab_user_01" />
          </div>
          <div className="form-row">
            <label>Статус</label>
            <input name="status" value={form.status} onChange={handleChange} placeholder="VERIFIED" />
          </div>
          <div className="form-row">
            <label>Временная метка (должна совпадать с регистрацией)</label>
            <input name="timestamp" value={form.timestamp} onChange={handleChange} placeholder="2026-06-15T12:00:00Z" />
          </div>
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Верификация…' : 'Проверить целостность данных'}
          </button>
        </form>
      </div>

      {result && result.data && (
        <div className={`alert ${result.data.verified ? 'alert-success' : 'alert-error'}`}>
          <div className="alert-title">
            {result.data.verified ? '✓ Целостность данных подтверждена' : '✗ Проверка целостности не пройдена'}
          </div>
          <div className="small">{result.data.message}</div>
          {result.data.hash && (
            <div className="mono small mt-8">{result.data.hash}</div>
          )}
          {result.data.computed_hash && (
            <div className="small muted mt-8">
              Вычислен:&nbsp; <span className="mono">{result.data.computed_hash.slice(0, 32)}…</span>
              <br />
              Сохранён:&nbsp; <span className="mono">{result.data.stored_hash?.slice(0, 32)}…</span>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// ─── Архитектура ──────────────────────────────────────────────────────────────

function Architecture() {
  const layers = [
    {
      num: '1',
      title: 'SENAITE LIMS',
      subtitle: 'Лабораторная информационная система',
      color: 'layer-lims',
      icon: '🗄️',
      items: [
        'Хранит пробы, результаты, пользователей, приборы',
        'Управляет статусами рабочего процесса и записями аудита',
        'Единственный источник достоверных лабораторных данных',
        'Конфиденциальные данные не покидают LIMS',
      ],
    },
    {
      num: '2',
      title: 'Go Gin Адаптер',
      subtitle: 'Интеллектуальный слой управления',
      color: 'layer-adapter',
      icon: '⚙️',
      items: [
        'Получает лабораторные события от LIMS или UI',
        'Нормализует и валидирует метаданные события',
        'Вычисляет детерминированный SHA-256 хэш',
        'Определяет важность события для регистрации в блокчейне',
        'Отправляет хэш в смарт-контракт',
        'Верифицирует текущие данные по сохранённому хэшу',
      ],
    },
    {
      num: '3',
      title: 'LimsHashRegistry.sol',
      subtitle: 'Смарт-контракт Solidity',
      color: 'layer-contract',
      icon: '📜',
      items: [
        'registerHash(limsRecordId, eventType, dataHash, status)',
        'getHash(limsRecordId) — получить сохранённый хэш',
        'verifyHash(limsRecordId, dataHash) — проверка целостности',
        'Генерирует событие HashRegistered для аудита',
        'Хранит только хэш + метаданные, никогда не хранит лабораторные данные',
      ],
    },
    {
      num: '4',
      title: 'Блокчейн-сеть',
      subtitle: 'Доверенный неизменяемый слой верификации',
      color: 'layer-blockchain',
      icon: '⛓️',
      items: [
        'Неизменяемость — зарегистрированные хэши не могут быть изменены',
        'Временны́е метки — каждый хэш имеет метку блока',
        'Аудиторский след — все события логируются через Solidity events',
        'Децентрализованность — нет единой точки контроля',
        'Доказательство доверия — верифицируется любой стороной',
      ],
    },
  ]

  return (
    <div className="page">
      <h2 className="page-title">Архитектура системы</h2>
      <p className="page-sub">
        Четырёхуровневая архитектура, разделяющая хранение лабораторных данных и верификацию целостности.
        Блокчейн — это слой доверия и доказательств, а не замена LIMS.
      </p>

      <div className="card">
        <h3 className="card-title">Основной принцип</h3>
        <div className="principle-box">
          <div className="principle-item">
            <span className="principle-icon">🗄️</span>
            <span>Данные остаются в LIMS</span>
          </div>
          <div className="principle-sep">→</div>
          <div className="principle-item">
            <span className="principle-icon">⚙️</span>
            <span>Адаптер вычисляет хэш</span>
          </div>
          <div className="principle-sep">→</div>
          <div className="principle-item">
            <span className="principle-icon">⛓️</span>
            <span>Блокчейн хранит доказательство целостности</span>
          </div>
        </div>
      </div>

      <div className="layers">
        {layers.map(layer => (
          <div key={layer.num} className={`layer-card ${layer.color}`}>
            <div className="layer-header">
              <span className="layer-num">{layer.num}</span>
              <div>
                <div className="layer-title">
                  {layer.icon} {layer.title}
                </div>
                <div className="layer-sub">{layer.subtitle}</div>
              </div>
            </div>
            <ul className="layer-list">
              {layer.items.map((item, i) => (
                <li key={i}>{item}</li>
              ))}
            </ul>
          </div>
        ))}
      </div>

      <div className="card">
        <h3 className="card-title">Точки интеграции блокчейна</h3>
        <div className="event-types">
          {[
            'SAMPLE_CREATED',
            'RESULT_SUBMITTED',
            'RESULT_VERIFIED',
            'REPORT_PUBLISHED',
            'DATA_MODIFIED',
            'INSTRUMENT_IMPORT',
          ].map(t => (
            <span key={t} className="badge badge-lg">{t}</span>
          ))}
        </div>
        <p className="small muted mt-16">
          Это критические события рабочего процесса, в которых адаптер перехватывает данные LIMS,
          вычисляет хэш и регистрирует его в блокчейне для последующей верификации целостности.
        </p>
      </div>
    </div>
  )
}

// ─── App ───────────────────────────────────────────────────────────────────────

const PAGES = [
  { id: 'dashboard',    label: 'Дашборд' },
  { id: 'register',     label: 'Регистрация события' },
  { id: 'registry',     label: 'Реестр хэшей' },
  { id: 'verify',       label: 'Верификация данных' },
  { id: 'architecture', label: 'Архитектура' },
]

export default function App() {
  const [page, setPage] = useState('dashboard')
  const [events, setEvents] = useState([])
  const [loading, setLoading] = useState(false)
  const [adapterInfo, setAdapterInfo] = useState(null)

  const fetchEvents = useCallback(async () => {
    setLoading(true)
    try {
      const res = await axios.get(`${API}/events`)
      setEvents(res.data.records || [])
    } catch {
      // адаптер ещё не доступен
    } finally {
      setLoading(false)
    }
  }, [])

  const fetchAdapterInfo = useCallback(async () => {
    try {
      const res = await axios.get(`${API}/`)
      setAdapterInfo(res.data)
    } catch {
      // адаптер ещё не доступен
    }
  }, [])

  useEffect(() => {
    fetchEvents()
    fetchAdapterInfo()
  }, [fetchEvents, fetchAdapterInfo])

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <div className="brand-icon">⛓️</div>
          <div>
            <div className="brand-title">LIMS Blockchain</div>
            <div className="brand-sub">Реестр хэшей</div>
          </div>
        </div>
        <nav className="nav">
          {PAGES.map(p => (
            <button
              key={p.id}
              className={`nav-item ${page === p.id ? 'nav-item-active' : ''}`}
              onClick={() => setPage(p.id)}
            >
              {p.label}
            </button>
          ))}
        </nav>
        <div className="sidebar-footer">
          <div className="small muted">Адаптер: {API}</div>
          <div className="small muted">{events.length} записей</div>
        </div>
      </aside>

      <main className="main">
        {page === 'dashboard'    && <Dashboard events={events} adapterInfo={adapterInfo} />}
        {page === 'register'     && <RegisterEvent onSuccess={fetchEvents} />}
        {page === 'registry'     && <HashRegistry events={events} loading={loading} onRefresh={fetchEvents} />}
        {page === 'verify'       && <VerifyData />}
        {page === 'architecture' && <Architecture />}
      </main>
    </div>
  )
}
