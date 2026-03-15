import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
  LabelList,
} from 'recharts';
import { CategoryBreakdownItem } from '../lib/summary-api';

interface Props {
  data: CategoryBreakdownItem[];
}

const COLORS = [
  '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6',
  '#EC4899', '#14B8A6', '#F97316', '#6366F1', '#84CC16',
];

function formatCurrency(cents: number): string {
  return ((cents ?? 0) / 100).toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  });
}

function renderLabel(props: { x?: number; y?: number; width?: number; value?: number; index?: number }, total: number) {
  const { x = 0, y = 0, width = 0, value = 0, index = 0 } = props;
  const pct = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0';
  return (
    <text
      x={x + width / 2}
      y={y - 8}
      textAnchor="middle"
      fontSize={12}
      fill="#374151"
    >
      {formatCurrency(value)} ({pct}%)
    </text>
  );
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: { value: number; payload: { category_name: string } }[];
  total: number;
}

function CustomTooltip({ active, payload, total }: CustomTooltipProps) {
  if (!active || !payload?.length) return null;
  const item = payload[0];
  const pct = total > 0 ? ((item.value / total) * 100).toFixed(1) : '0.0';
  return (
    <div className="rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm shadow-lg">
      <p className="font-medium text-gray-900">{item.payload.category_name}</p>
      <p className="text-gray-600">
        {formatCurrency(item.value)} ({pct}%)
      </p>
    </div>
  );
}

export default function CategoryDistributionChart({ data }: Props) {
  if (!data || data.length === 0) return null;

  const total = data.reduce((sum, d) => sum + d.total_cents, 0);
  if (total === 0) return null;

  const chartHeight = Math.max(300, data.length * 50);

  return (
    <div className="mb-6 rounded-lg bg-white p-5 shadow">
      <h3 className="mb-4 text-lg font-semibold text-gray-800">
        Gastos por Categoria
      </h3>
      <ResponsiveContainer width="100%" height={chartHeight}>
        <BarChart
          data={data}
          margin={{ top: 24, right: 20, left: 20, bottom: 5 }}
        >
          <XAxis
            dataKey="category_name"
            tick={{ fontSize: 12, fill: '#6B7280' }}
            axisLine={{ stroke: '#E5E7EB' }}
            tickLine={false}
            interval={0}
            angle={data.length > 5 ? -30 : 0}
            textAnchor={data.length > 5 ? 'end' : 'middle'}
            height={data.length > 5 ? 80 : 40}
          />
          <YAxis
            tickFormatter={(v: number) => formatCurrency(v)}
            tick={{ fontSize: 11, fill: '#9CA3AF' }}
            axisLine={false}
            tickLine={false}
            width={100}
          />
          <Tooltip
            content={<CustomTooltip total={total} />}
            cursor={{ fill: 'rgba(59, 130, 246, 0.05)' }}
          />
          <Bar dataKey="total_cents" radius={[6, 6, 0, 0]} maxBarSize={60}>
            {data.map((_, index) => (
              <Cell key={index} fill={COLORS[index % COLORS.length]} />
            ))}
            <LabelList
              dataKey="total_cents"
              position="top"
              content={(props) => renderLabel(props as { x?: number; y?: number; width?: number; value?: number; index?: number }, total)}
            />
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
