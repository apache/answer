/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

export interface FormValue<T = any> {
  value: T;
  isInvalid: boolean;
  errorMsg: string;
  [prop: string]: any;
}

export interface FormDataType {
  [prop: string]: FormValue;
}

export interface FieldError {
  error_field: string;
  error_msg: string;
}

export interface Paging {
  page: number;
  page_size?: number;
}

export type ReportType = 'question' | 'answer' | 'comment' | 'user';
export type ReportAction = 'close' | 'flag' | 'review';
export interface ReportParams {
  type: ReportType;
  action: ReportAction;
}

export interface TagBase {
  display_name: string;
  slug_name: string;
  original_text?: string;
  recommend?: boolean;
  reserved?: boolean;
}

export interface Tag extends TagBase {
  main_tag_slug_name?: string;
  parsed_text?: string;
  tag_id?: string;
}

export interface SynonymsTag extends Tag {
  tag_id: string;
  tag?: string;
}

export interface TagInfo extends TagBase {
  tag_id: string;
  original_text: string;
  parsed_text: string;
  follow_count: number;
  question_count: number;
  is_follower: boolean;
  member_actions;
  created_at?;
  updated_at?;
  main_tag_slug_name?: string;
  excerpt?;
  status: string;
}
export interface QuestionParams extends ImgCodeReq {
  title: string;
  url_title?: string;
  content: string;
  tags: Tag[];
}

export interface QuestionWithAnswer extends QuestionParams {
  answer_content: string;
}

export interface ListResult<T = any> {
  count: number;
  list: T[];
}

export interface AnswerParams extends ImgCodeReq {
  content: string;
  html: string;
  question_id: string;
  id: string;
  edit_summary?: string;
}

export interface LoginReqParams {
  e_mail: string;
  /** password */
  pass: string;
  captcha_id?: string;
  captcha_code?: string;
}

export interface RegisterReqParams extends LoginReqParams {
  name: string;
}

export interface ModifyPasswordReq {
  old_pass: string;
  pass: string;
}

/** User  */
export interface ModifyUserReq {
  display_name: string;
  username?: string;
  avatar: any;
  bio: string;
  bio_html?: string;
  location: string;
  website: string;
}

enum RoleId {
  User = 1,
  Admin = 2,
  Moderator = 3,
}

export interface User {
  username: string;
  rank: number;
  vote_count: number;
  display_name: string;
  avatar: string;
}

export interface UserInfoBase {
  id?: string;
  avatar: any;
  username: string;
  display_name: string;
  rank: number;
  website: string;
  location: string;
  ip_info?: string;
  status?: 'normal' | 'suspended' | 'deleted' | 'inactive';
  /** roles */
  role_id?: RoleId;
}

export interface UserInfoRes extends UserInfoBase {
  bio: string;
  bio_html: string;
  create_time?: string;
  /**
   * value = 1 active;
   * value = 2 inactivated
   */
  mail_status: number;
  language: string;
  e_mail?: string;
  have_password: boolean;
  [prop: string]: any;
}

export type UploadType = 'post' | 'avatar' | 'branding' | 'post_attachment';
export interface UploadReq {
  file: FormData;
}

export interface ImgCodeReq {
  captcha_id?: string;
  captcha_code?: string;
}

export interface ImgCodeRes {
  captcha_id: string;
  captcha_img: string;
  verify: boolean;
}

export interface PasswordResetReq extends ImgCodeReq {
  e_mail: string;
}

export interface PasswordReplaceReq extends ImgCodeReq {
  code: string;
  pass: string;
}

export interface CaptchaReq extends ImgCodeReq {
  verify: ImgCodeRes['verify'];
}

export type CaptchaKey =
  | 'email'
  | 'password'
  | 'edit_userinfo'
  | 'question'
  | 'answer'
  | 'comment'
  | 'edit'
  | 'invitation_answer'
  | 'search'
  | 'report'
  | 'delete'
  | 'vote';

export interface SetNoticeReq {
  notice_switch: boolean;
}

export interface NotificationBadgeAward {
  notification_id: string;
  badge_id: string;
  name: string;
  icon: string;
  level: number;
}
export interface NotificationStatus {
  inbox: number;
  achievement: number;
  revision: number;
  can_revision: boolean;
  badge_award: NotificationBadgeAward | null;
}

export interface QuestionDetailRes {
  id: string;
  title: string;
  content: string;
  html: string;
  tags: any[];
  view_count: number;
  unique_view_count?: number;
  answer_count: number;
  favorites_count: number;
  follow_counts: 0;
  accepted_answer_id: string;
  last_answer_id: string;
  create_time: string;
  update_time: string;
  user_info: UserInfoBase;
  answered: boolean;
  collected: boolean;
  answer_ids: string[];

  [prop: string]: any;
}

export interface AnswersReq extends Paging {
  order?: 'default' | 'updated' | 'created';
  question_id: string;
}

export interface AnswerItem {
  id: string;
  question_id: string;
  content: string;
  html: string;
  create_time: string;
  update_time: string;
  user_info: UserInfoBase;
  [prop: string]: any;
}

export interface PostAnswerReq extends ImgCodeReq {
  content: string;
  html?: string;
  question_id: string;
}

export interface PageUser {
  id?;
  displayName;
  userName?;
  avatar_url?;
}

export interface LangsType {
  label: string;
  value: string;
}

/**
 * @description interface for Question
 */
export type QuestionOrderBy =
  | 'recommend'
  | 'newest'
  | 'active'
  | 'hot'
  | 'score'
  | 'unanswered'
  | 'frequent';

export interface QueryQuestionsReq extends Paging {
  order: QuestionOrderBy;
  tag?: string;
  in_days?: number;
}

export type AdminQuestionStatus =
  | 'available'
  | 'pending'
  | 'closed'
  | 'deleted';

export type AdminContentsFilterBy = 'normal' | 'pending' | 'closed' | 'deleted';

export interface AdminContentsReq extends Paging {
  status: AdminContentsFilterBy;
  query?: string;
}

/**
 * @description interface for Answer
 */
export type AdminAnswerStatus = 'available' | 'deleted';

/**
 * @description interface for Users
 */
export type UserFilterBy =
  | 'normal'
  | 'staff'
  | 'inactive'
  | 'suspended'
  | 'deleted';

export type BadgeFilterBy = 'all' | 'active' | 'inactive';

export type InstalledPluginsFilterBy =
  | 'all'
  | 'active'
  | 'inactive'
  | 'outdated';
/**
 * @description interface for Flags
 */
export type FlagStatus = 'pending' | 'completed';
export type FlagType = 'all' | 'question' | 'answer' | 'comment';
export interface AdminFlagsReq extends Paging {
  status: FlagStatus;
  object_type: FlagType;
}

/**
 * @description interface for Admin Settings
 */
export interface AdminSettingsGeneral {
  name: string;
  short_description: string;
  description: string;
  site_url: string;
  contact_email: string;
  permalink?: number;
}

export interface HelmetBase {
  pageTitle?: string;
  description?: string;
  keywords?: string;
}

export interface HelmetUpdate extends Omit<HelmetBase, 'pageTitle'> {
  title?: string;
  subtitle?: string;
}

export interface AdminSettingsInterface {
  language: string;
  time_zone?: string;
}

export interface AdminSettingsSmtp {
  encryption: string;
  from_email: string;
  from_name: string;
  smtp_authentication: boolean;
  smtp_host: string;
  smtp_password?: string;
  smtp_port: number;
  smtp_username?: string;
  test_email_recipient?: string;
}

export interface AdminSettingsUsers {
  allow_update_avatar: boolean;
  allow_update_bio: boolean;
  allow_update_display_name: boolean;
  allow_update_location: boolean;
  allow_update_username: boolean;
  allow_update_website: boolean;
  default_avatar: string;
  gravatar_base_url: string;
}

export interface AdminSettingsSecurity {
  external_content_display: string;
  check_update: boolean;
  login_required: boolean;
}

export interface SiteSettings {
  branding: AdminSettingBranding;
  general: AdminSettingsGeneral;
  interface: AdminSettingsInterface;
  login: AdminSettingsLogin;
  custom_css_html: AdminSettingsCustom;
  theme: AdminSettingsTheme;
  site_seo: AdminSettingsSeo;
  site_users: AdminSettingsUsers;
  site_advanced: AdminSettingsWrite;
  site_questions: AdminQuestionSetting;
  site_tags: AdminTagsSetting;
  version: string;
  revision: string;
  site_security: AdminSettingsSecurity;
}

export interface AdminSettingBranding {
  logo: string;
  square_icon: string;
  mobile_logo?: string;
  favicon?: string;
}

export interface AdminSettingsLegal {
  privacy_policy_original_text?: string;
  privacy_policy_parsed_text?: string;
  terms_of_service_original_text?: string;
  terms_of_service_parsed_text?: string;
}

export interface AdminSettingsWrite {
  max_image_size?: number;
  max_attachment_size?: number;
  max_image_megapixel?: number;
  authorized_image_extensions?: string[];
  authorized_attachment_extensions?: string[];
}

export interface AdminSettingsSeo {
  robots: string;
  /**
   * 0: not set
   * 1：with title
   * 2: no title
   */
  permalink: number;
}

export type themeConfig = {
  navbar_style: string;
  primary_color: string;
  [k: string]: string | number;
};
export interface AdminSettingsTheme {
  theme: string;
  color_scheme: string;
  layout: string;
  theme_options?: { label: string; value: string }[];
  theme_config: Record<string, themeConfig>;
}

export interface AdminSettingsCustom {
  custom_css: string;
  custom_head: string;
  custom_header: string;
  custom_footer: string;
  custom_sidebar: string;
}

export interface AdminSettingsLogin {
  allow_new_registrations: boolean;
  allow_email_registrations: boolean;
  allow_email_domains: string[];
  allow_password_login: boolean;
}

/**
 * @description interface for Activity
 */
export interface FollowParams {
  is_cancel: boolean;
  object_id: string;
}

/**
 * @description search request params
 */
export interface SearchParams extends ImgCodeReq {
  q: string;
  order: string;
  page: number;
  size?: number;
}

/**
 * @description search response data
 */
export interface SearchResItem {
  object_type: string;
  object: {
    url_title?: string;
    id: string;
    question_id?: string;
    title: string;
    excerpt: string;
    created_at: number;
    user_info: UserInfoBase;
    vote_count: number;
    answer_count: number;
    accepted: boolean;
    tags: TagBase[];
    status?: string;
  };
}
export interface SearchRes extends ListResult<SearchResItem> {
  extra: any;
}

export interface AdminDashboard {
  info: {
    question_count: number;
    resolved_count: number;
    resolved_rate: string;
    unanswered_count: number;
    unanswered_rate: string;
    answer_count: number;
    comment_count: number;
    vote_count: number;
    user_count: number;
    report_count: number;
    uploading_files: boolean;
    smtp: 'enabled' | 'disabled' | 'not_configured';
    time_zone: string;
    occupying_storage_space: string;
    app_start_time: number;
    https: boolean;
    login_required: boolean;
    go_version: string;
    database_version: string;
    database_size: string;
    version_info: {
      remote_version: string;
      version: string;
    };
  };
}

export interface TimelineReq {
  show_vote: boolean;
  object_id: string;
}

export interface TimelineItem {
  activity_id: number;
  revision_id: number;
  created_at: number;
  activity_type: string;
  comment: string;
  object_id: string;
  object_type: string;
  cancelled: boolean;
  cancelled_at: any;
  user_info: UserInfoBase;
}

export interface TimelineObject {
  title: string;
  url_title?: string;
  object_type: string;
  question_id: string;
  answer_id: string;
  main_tag_slug_name?: string;
  display_name?: string;
}

export interface TimelineRes {
  object_info: TimelineObject;
  timeline: TimelineItem[];
}

export interface SuggestReviewItem {
  type: 'question' | 'answer' | 'tag';
  info: {
    url_title?: string;
    object_id: string;
    title: string;
    content: string;
    html: string;
    tags: Tag[];
  };
  unreviewed_info: {
    id: string;
    use_id: string;
    object_id: string;
    title: string;
    status: 0 | 1;
    create_at: number;
    user_info: UserInfoBase;
    reason: string;
    content: Tag | QuestionDetailRes | AnswerItem;
  };
}
export interface SuggestReviewResp {
  count: number;
  list: SuggestReviewItem[];
}

export interface ReasonItem {
  content_type: string;
  description: string;
  name: string;
  placeholder: string;
  reason_type: number;
}

export interface BaseReviewItem {
  object_type: 'question' | 'answer' | 'comment' | 'user';
  object_id: string;
  object_show_status: number;
  object_status: number;
  tags: Tag[];
  title: string;
  original_text: string;
  author_user_info: UserInfoBase;
  created_at: number;
  submit_at: number;
  comment_id: string;
  question_id: string;
  answer_id: string;
  answer_count: number;
  answer_accepted?: boolean;
  flag_id: string;
  url_title: string;
  parsed_text: string;
}

export interface FlagReviewItem extends BaseReviewItem {
  reason: ReasonItem;
  reason_content: string;
  submitter_user: UserInfoBase;
}

export interface FlagReviewResp {
  count: number;
  list: FlagReviewItem[];
}

export interface QueuedReviewItem extends BaseReviewItem {
  review_id: number;
  reason: string;
  submitter_display_name: string;
}

export interface QueuedReviewResp {
  count: number;
  list: QueuedReviewItem[];
}

export interface UserRoleItem {
  id: number;
  name: string;
  description: string;
}
export interface MemberActionItem {
  action: string;
  name: string;
  type: string;
}

export interface QuestionOperationReq {
  id: string;
  operation: 'pin' | 'unpin' | 'hide' | 'show';
}

export interface OauthBindEmailReq {
  binding_key: string;
  email: string;
  must: boolean;
}

export interface UserOauthConnectorItem {
  icon: string;
  name: string;
  link: string;
  binding: boolean;
  external_id: string;
}

export interface NotificationConfigItem {
  enable: boolean;
  key: string;
}
export interface NotificationConfig {
  all_new_question: NotificationConfigItem;
  all_new_question_for_following_tags: NotificationConfigItem;
  inbox: NotificationConfigItem;
}

export interface ActivatedPlugin {
  slug_name: string;
  enabled: boolean;
}

export interface UserPluginsConfigRes {
  name: string;
  slug_name: string;
}

export interface ReviewTypeItem {
  label: string;
  name: string;
  todo_amount: number;
}

export interface PutFlagReviewParams {
  operation_type:
    | 'edit_post'
    | 'close_post'
    | 'delete_post'
    | 'unlist_post'
    | 'ignore_report';
  flag_id: string;
  close_msg?: string;
  close_type?: number;
  title?: string;
  content?: string;
  tags?: Tag[];
  // mention_username_list?: any;
  captcha_code?: any;
  captcha_id?: any;
}

/**
 * @description response for reaction
 */
export interface ReactionItems {
  reaction_summary: ReactionItem[];
}

export interface ReactionItem {
  emoji: string;
  count: number;
  tooltip: string;
  is_active: boolean;
}

export interface BadgeListItem {
  id: string;
  name: string;
  icon: string;
  award_count: number;
  earned: boolean;
  /** 1: bronze 2: silver 3:gold */
  level: number;
  earned_count?: number;
}

export interface BadgeListGroupItem {
  badges: BadgeListItem[];
  group_name: string;
}

export interface BadgeInfo extends BadgeListItem {
  description: string;
  earned_count: number;
  is_single: boolean;
}

export interface AdminBadgeListItem extends BadgeListItem {
  group_name: string;
  status: string;
  description: string;
}

export interface BadgeDetailListReq {
  page: number;
  page_size: number;
  badge_id: string;
  username?: string | null;
}
export interface BadgeDetailListItem {
  created_at: number;
  author_user_info: UserInfoBase;
  object_type: string;
  object_id: string;
  url_title: string;
  question_id: string;
  answer_id: string;
  comment_id: string;
}

export interface BadgeDetailListRes {
  count: number;
  list: BadgeDetailListItem[];
}

export interface AdminQuestionSetting {
  min_tags: number;
  min_content: number;
  restrict_answer: boolean;
}

export interface AdminTagsSetting {
  recommend_tags: Tag[];
  required_tag: boolean;
  reserved_tags: Tag[];
}
