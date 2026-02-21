import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts';
import { formatCurrency } from '../utils/formatCurrency';

export default function SalesChart({ data, loading }) {
  return (
    <div className="card border-0 shadow-sm mb-4">
      <div className="card-body">
        <h5 className="card-title mb-3"><i className="bi bi-bar-chart-fill me-2 text-primary"/>Sales Record</h5>
        {loading && <div className="skeleton skeleton-chart w-100" style={{ height: 180 }} />}
        {!loading && data.length === 0 && <p className="text-muted mb-0">No sales data yet.</p>}
        {!loading && data.length > 0 && (
          <div style={{ width: '100%', height: 250 }}>
            <ResponsiveContainer>
              <BarChart data={data} margin={{ top: 10, right: 20, left: 0, bottom: 10 }}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" tickFormatter={(d) => d.slice(5)} />
                <YAxis tickFormatter={(v) => `$${v}` } width={50} />
                <Tooltip formatter={(v) => formatCurrency(v)} labelFormatter={(l) => `Date: ${l}`} />
                <Bar dataKey="total" fill="#D8A35D" radius={[4,4,0,0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>
    </div>
  );
}
