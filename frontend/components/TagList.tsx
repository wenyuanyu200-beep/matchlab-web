export default function TagList({ tags = [] }: { tags?: string[] }) {
  if (!tags.length) return null;
  return (
    <div className="flex flex-wrap gap-2" aria-label="标签">
      {tags.map((tag) => <span className="tag" key={tag}>{tag}</span>)}
    </div>
  );
}
