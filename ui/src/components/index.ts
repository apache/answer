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

import Avatar from './Avatar';
import Editor, { EditorRef, htmlRender } from './Editor';
import Header from './Header';
import Footer from './Footer';
import Icon from './Icon';
import SvgIcon from './Icon/svg';
import Modal from './Modal';
import TagSelector from './TagSelector';
import Unactivate from './Unactivate';
import UploadImg from './UploadImg';
import Actions from './Actions';
import Tag from './Tag';
import Operate from './Operate';
import UserCard from './UserCard';
import Pagination from './Pagination';
import Comment from './Comment';
import TextArea from './TextArea';
import Mentions from './Mentions';
import FormatTime from './FormatTime';
import Toast from './Toast';
import AccordionNav from './AccordionNav';
import Empty from './Empty';
import BaseUserCard from './BaseUserCard';
import FollowingTags from './FollowingTags';
import QueryGroup from './QueryGroup';
import BrandUpload from './BrandUpload';
import SchemaForm, { JSONSchema, UISchema, initFormData } from './SchemaForm';
import DiffContent from './DiffContent';
import Customize from './Customize';
import CustomizeTheme from './CustomizeTheme';
import PageTags from './PageTags';
import QuestionListLoader from './QuestionListLoader';
import TagsLoader from './TagsLoader';
import WelcomeTitle from './WelcomeTitle';
import Counts from './Counts';
import QuestionList from './QuestionList';
import HotQuestions from './HotQuestions';
import HttpErrorContent from './HttpErrorContent';
import CustomSidebar from './CustomSidebar';
import ImgViewer from './ImgViewer';
import SideNav from './SideNav';
import PluginRender from './PluginRender';
import HighlightText from './HighlightText';

export {
  Avatar,
  Header,
  Footer,
  Icon,
  SvgIcon,
  Modal,
  Unactivate,
  UploadImg,
  Editor,
  Tag,
  TagSelector,
  Actions,
  Operate,
  UserCard,
  Pagination,
  Comment,
  TextArea,
  Mentions,
  FormatTime,
  Toast,
  AccordionNav,
  Empty,
  BaseUserCard,
  FollowingTags,
  htmlRender,
  QueryGroup,
  BrandUpload,
  SchemaForm,
  initFormData,
  DiffContent,
  Customize,
  CustomizeTheme,
  PageTags,
  QuestionListLoader,
  TagsLoader,
  WelcomeTitle,
  Counts,
  QuestionList,
  HotQuestions,
  HttpErrorContent,
  CustomSidebar,
  ImgViewer,
  SideNav,
  PluginRender,
  HighlightText,
};
export type { EditorRef, JSONSchema, UISchema };
