export interface User {
  id: string;
  email: string;
  nickname: string;
  role: "user" | "admin";
  school: string;
  created_at?: string;
}

export interface Creator {
  id?: string;
  nickname: string;
  school: string;
}

export interface Activity {
  id: string;
  creator_id: string;
  title: string;
  type: string;
  description: string;
  required_count: number;
  joined_count: number;
  tags: string[];
  preferred_tags?: string[];
  time_text: string;
  location_text: string;
  status: string;
  creator?: Creator;
  created_at?: string;
}

export interface Application {
  id: string;
  activity_id: string;
  activity_title?: string;
  activity_status?: string;
  applicant_id?: string;
  applicant?: Creator;
  reason: string;
  match_score: number;
  status: string;
  created_at?: string;
}

export interface ProfileScores {
  interest_score: number;
  skill_score: number;
  time_score: number;
  goal_score: number;
  communication_score: number;
}

export interface Profile {
  user_id?: string;
  profile_type: string;
  tags: string[];
  scores: ProfileScores;
  summary: string;
}

export interface DetailScores {
  interest: number;
  skill: number;
  type: number;
  time: number;
  goal: number;
}

export interface Recommendation {
  id?: string;
  activity: Activity;
  score: number;
  detail_scores: DetailScores;
  reason: string;
  updated_at?: string;
}

export interface AdminStats {
  users_count: number;
  activities_count: number;
  applications_count: number;
  matches_count: number;
  questionnaires_count: number;
  feedbacks_count: number;
  recruiting_activities_count: number;
  pending_applications_count: number;
  approved_applications_count: number;
}

export interface AdminActivity {
  id: string;
  title: string;
  type: string;
  status: string;
  creator_id: string;
  creator: Creator;
  required_count: number;
  joined_count: number;
  created_at: string;
}

export interface Feedback {
  id: string;
  user_id: string;
  activity_id?: string;
  match_id?: string;
  rating: number;
  comment?: string;
  created_at: string;
}
