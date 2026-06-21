export function splitList(value: string): string[] {
  return value
    .split(/[,，]/)
    .map((item) => item.trim())
    .filter(Boolean);
}

export function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

export function friendlyStatus(status?: string): string {
  const labels: Record<string, string> = {
    recruiting: "招募中",
    full: "已满员",
    closed: "已结束",
    pending: "待审核",
    approved: "已通过",
    rejected: "未通过",
    cancelled: "已取消",
    recommended: "已推荐",
  };
  return labels[status || ""] || status || "未知";
}
