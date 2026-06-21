export default function StatCard({ label, value, accent = "indigo" }: { label: string; value: number | string; accent?: "indigo" | "cyan" | "orange" }) {
  const colors = { indigo: "bg-indigo-50 text-indigo-700", cyan: "bg-cyan-50 text-cyan-700", orange: "bg-orange-50 text-orange-700" };
  return (
    <div className="card">
      <span className={`inline-flex rounded-lg px-2 py-1 text-xs font-bold ${colors[accent]}`}>{label}</span>
      <div className="mt-4 text-3xl font-black text-slate-950">{value}</div>
    </div>
  );
}
