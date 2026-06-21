export default function EmptyState({ title = "暂时没有内容", description }: { title?: string; description?: string }) {
  return (
    <div className="card col-span-full py-14 text-center">
      <div className="mx-auto mb-4 grid size-14 place-items-center rounded-2xl bg-indigo-50 text-2xl text-indigo-600">✦</div>
      <h3 className="text-lg font-bold text-slate-900">{title}</h3>
      {description && <p className="mt-2 text-sm text-slate-600">{description}</p>}
    </div>
  );
}
