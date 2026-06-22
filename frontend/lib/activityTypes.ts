export const activityTypeOptions = [
  { value: "competition", label: "比赛组队" },
  { value: "project", label: "项目合作" },
  { value: "study", label: "学习搭子" },
  { value: "club", label: "社团活动" },
  { value: "volunteer", label: "志愿活动" },
  { value: "workshop", label: "讲座沙龙" },
  { value: "social", label: "兴趣活动" },
  { value: "startup", label: "创业招募" },
  { value: "parttime", label: "短期协作" },
] as const;

export type ActivityTypeValue = (typeof activityTypeOptions)[number]["value"];

export function activityTypeLabel(type?: string): string {
  return activityTypeOptions.find((item) => item.value === type)?.label || type || "校园协作";
}

export const activityPlaceholders: Record<
  string,
  { title: string; description: string; tags: string; preferredTags: string; roles: string }
> = {
  competition: {
    title: "智能硬件比赛组队",
    description: "准备参加电子设计竞赛，寻找会单片机、传感器和结构设计的队友。",
    tags: "电赛, STM32, 硬件",
    preferredTags: "嵌入式, 焊接, 控制",
    roles: "嵌入式、焊接、控制",
  },
  project: {
    title: "校园调研项目搭子",
    description: "围绕校园学习空间做一次小型调研，需要同学一起访谈、整理问卷和写报告。",
    tags: "调研, 报告, 校园项目",
    preferredTags: "沟通, 写作, 数据整理",
    roles: "访谈、写作、数据整理",
  },
  study: {
    title: "高数复习学习搭子",
    description: "准备期末复习，希望找同学一起整理题型、互相提醒进度。",
    tags: "学习, 高数, 期末",
    preferredTags: "自律, 稳定, 愿意讨论",
    roles: "一起复习、整理题型、互相提醒",
  },
  social: {
    title: "周末摄影散步搭子",
    description: "周末想在校园拍照散步，欢迎喜欢摄影或想练习构图的同学一起参加。",
    tags: "摄影, 校园, 兴趣活动",
    preferredTags: "拍照, 审美, 轻松交流",
    roles: "拍照、路线规划、轻松交流",
  },
  club: {
    title: "社团活动执行伙伴",
    description: "筹备一次社团开放活动，需要同学一起做现场执行、物料整理和报名沟通。",
    tags: "社团, 活动执行, 校园",
    preferredTags: "组织, 沟通, 执行",
    roles: "现场执行、物料整理、报名沟通",
  },
  volunteer: {
    title: "校园志愿服务招募",
    description: "计划组织一次图书整理或校园引导志愿服务，需要时间稳定、认真负责的同学参与。",
    tags: "志愿服务, 校园, 公益",
    preferredTags: "负责, 耐心, 稳定",
    roles: "现场服务、秩序维护、物资整理",
  },
  workshop: {
    title: "AI 工具分享沙龙",
    description: "准备做一次小型经验分享，邀请对 AI、效率工具或内容创作感兴趣的同学参与。",
    tags: "AI, 分享, 沙龙",
    preferredTags: "表达, 记录, 讨论",
    roles: "分享者、记录者、现场协助",
  },
  startup: {
    title: "校园创业想法招募",
    description: "围绕校园服务做一个早期想法验证，希望找到产品、设计和运营方向的伙伴。",
    tags: "创业, 产品, 校园服务",
    preferredTags: "产品, 设计, 运营",
    roles: "产品调研、设计、运营",
  },
  parttime: {
    title: "短期协作任务招募",
    description: "需要短期协助完成资料整理、活动拍摄或内容运营，适合时间灵活的同学。",
    tags: "短期协作, 资料整理, 运营",
    preferredTags: "细心, 执行, 时间稳定",
    roles: "资料整理、拍摄、运营协助",
  },
};

export function placeholdersFor(type: string) {
  return activityPlaceholders[type] || activityPlaceholders.project;
}
